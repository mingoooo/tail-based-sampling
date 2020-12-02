package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/mingoooo/tail-based-sampling/utils"
)

// Receiver struct
type Receiver struct {
	HTTPPort       string
	DataURL        string
	DataPort       string
	DataSuffix     string
	startIndex     int64
	endIndex       int64
	ReadLimit      int64
	curContentLen  int64
	finishWg       *sync.WaitGroup
	Cache          *Cache
	Postman        *Postman
	traceCh        chan *pb.Trace
	tidCh          chan *pb.TraceID
	errPosIDMap    *sync.Map
	errTidMark     *sync.Map
	curPos         int64
	searchLen      int64
	flushPosTidMap map[int64]string
}

type Span struct {
	Raw     string
	Wrong   bool
	TraceID string
}

// New agent
func New(httpPort string, dataSuffix string) (*Receiver, error) {
	r := &Receiver{
		HTTPPort:       httpPort,
		DataSuffix:     dataSuffix,
		Cache:          newCache(128 * 1024 * 1024),
		endIndex:       8 * 1024 * 1024,
		ReadLimit:      8 * 1024 * 1024,
		curContentLen:  -1,
		finishWg:       &sync.WaitGroup{},
		traceCh:        make(chan *pb.Trace, 128),
		tidCh:          make(chan *pb.TraceID, 128),
		errPosIDMap:    &sync.Map{},
		errTidMark:     &sync.Map{},
		flushPosTidMap: map[int64]string{},
		searchLen:      40000,
	}
	var err error
	r.Postman, err = NewPostman(fmt.Sprintf("127.0.0.1:%s", "8003"), r.HTTPPort)
	if err != nil {
		return r, err
	}

	return r, err
}

// Run is entrypoint
func (r Receiver) Run(ctx context.Context, cancel context.CancelFunc) (err error) {
	r.RunHTTPSvr()
	return
}

func (r *Receiver) getRange() (int64, error) {
	req, err := http.NewRequest("HEAD", r.DataURL, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	return resp.ContentLength, nil
}

func (r *Receiver) httpGet() (io.ReadCloser, error) {
	// log.Printf("Pulling data... URL:%s \nstart index: %d\nend index: %d", r.DataURL, r.startIndex, r.endIndex)
	req, err := http.NewRequest("GET", r.DataURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", r.startIndex, r.endIndex))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusPartialContent, http.StatusOK:
		r.startIndex, r.endIndex = r.endIndex+1, r.endIndex+r.ReadLimit
		r.curContentLen = resp.ContentLength
	}
	return resp.Body, err
}

func (r *Receiver) PullData(preTail []byte) error {
	defer r.finishWg.Done()

	// for {
	// 	n, err := r.getRange()
	// 	if err != nil {
	// 		log.Println(err)
	// 		continue
	// 	}
	// 	if n > 0 {
	// 		break
	// 	}
	// }

	for {
		if r.curContentLen != -1 && r.curContentLen < r.ReadLimit {
			break
		}
		resp, err := r.httpGet()
		if err != nil {
			return err
		}
		defer resp.Close()

		reader := bufio.NewReader(resp)
		for {
			line, err := reader.ReadBytes('\n')
			// log.Printf(string(line))
			if len(preTail) > 0 {
				line = append(preTail, line...)
				preTail = []byte{}
			}
			if err == io.EOF {
				// r.pullData(line, true)
				// return nil
				preTail = line
				break
			} else if err != nil {
				return err
			}

			// TODO: slop
			atomic.AddInt64(&r.curPos, 1)
			s := NewSpan(line)

			if _, ok := r.Cache.Get(s.TraceID); !ok {
				r.flushPosTidMap[r.curPos+r.searchLen] = s.TraceID
			}
			r.setCache(s.TraceID, s.Raw)

			// send wrong trace id to backend
			if s.Wrong {
				// log.Printf("Set trace ID: %s", s.TraceID)
				if _, ok := r.errTidMark.LoadOrStore(s.TraceID, true); !ok {
					if tids, ok := r.errPosIDMap.LoadOrStore(r.curPos+r.searchLen, []string{s.TraceID}); ok {
						r.errPosIDMap.Store(r.curPos+r.searchLen, append(tids.([]string), s.TraceID))
					}
					r.SetErrTraceID(s.TraceID)
				}
			}

			if tids, ok := r.errPosIDMap.Load(r.curPos); ok {
				for _, tid := range tids.([]string) {
					i := tid
					// log.Printf("Send trace: %s", i)
					r.SendTraceByID(i)
					r.errTidMark.Delete(i)
				}
				r.errPosIDMap.Delete(r.curPos)
				// cleanLen := r.curPos - r.searchLen*2
				// if cleanLen > 0 {
				// 	log.Printf("Clean cache")
				// 	r.Cache.Clean(cleanLen)
				// }
			}
			// go r.SendTraceBySpan(s)

			// flush trace
			if r.curPos > r.searchLen {
				if tid, ok := r.flushPosTidMap[r.curPos]; ok {
					if _, ok := r.errTidMark.Load(tid); ok {
						r.SendTraceByID(tid)
					} else {
						r.Cache.UnsafeDelete(tid)
						delete(r.flushPosTidMap, r.curPos)
						// log.Printf("Delete trace id: %s", tid)
					}
				}
			}

		}
	}

	log.Printf("Flush all traces")
	r.errTidMark.Range(func(key, _ interface{}) bool {
		r.SendTraceByID(key.(string))
		return true
	})
	return nil
}

func (r *Receiver) SendTraceByID(tid string) {
	if trace, ok := r.DropTraceByID(tid); ok {
		r.traceCh <- trace
	}
}

func (r *Receiver) TraceSearcher() {
	for {
		t, ok := <-r.tidCh
		if !ok {
			return
		}
		r.errTidMark.LoadOrStore(t.ID, true)
		// if _, ok := r.errTidMark.LoadOrStore(t.ID, true); !ok {
		// if _, o := r.errPosIDMap.Load(atomic.LoadInt64(&r.curPos) + r.searchLen); o {
		// 	log.Printf("Set pos from searcher: exist")
		// }
		// pos := atomic.LoadInt64(&r.curPos) + r.searchLen
		// if tids, ok := r.errPosIDMap.LoadOrStore(pos, []string{t.ID}); ok {
		// 	r.errPosIDMap.Store(pos, append(tids.([]string), t.ID))
		// }
		// }
		// if trace, ok := r.GetTraceByID(t.ID); ok {
		// 	r.traceCh <- trace
		// }
	}
}

func (r *Receiver) SetErrTraceID(tid string) {
	t := &pb.Trace{
		TraceID:  tid,
		EndSpan:  false,
		SpanList: []*pb.Span{},
	}
	r.traceCh <- t
}

func (r *Receiver) DropTraceByID(tid string) (*pb.Trace, bool) {
	spans := r.Cache.Drop(tid)
	if len(spans) < 1 {
		return nil, false
	}

	t := &pb.Trace{
		TraceID:  tid,
		EndSpan:  false,
		SpanList: []*pb.Span{},
	}
	for _, s := range spans {
		t.SpanList = append(t.SpanList, ParseSpan(s))
	}
	return t, true
}

func (r Receiver) SendTraceBySpan(s *Span) {
	trace := &pb.Trace{
		TraceID:  s.TraceID,
		EndSpan:  true,
		SpanList: []*pb.Span{},
	}
	// TODO: TEST
	// time.Sleep(5 * time.Second)
	spans := r.Cache.Drop(s.TraceID)
	if len(spans) > 0 {
		for _, s := range spans {
			trace.SpanList = append(trace.SpanList, ParseSpan(s))
		}
	}
	// r.Cache.Delete(s.TraceID)
	r.traceCh <- trace
}

func (r *Receiver) setCache(tid, line string) {
	r.Cache.Set(tid, line)
}

func NewSpan(line []byte) *Span {
	s := &Span{}
	s.Raw = utils.ByteSliceToString(line)
	tidEndIndex := strings.IndexByte(s.Raw, '|')
	s.TraceID = s.Raw[:tidEndIndex]
	s.filterWrongSpan()
	return s
}

func (s *Span) filterWrongSpan() {
	tagFirstIndex := strings.LastIndexByte(s.Raw, '|')
	tags := s.Raw[tagFirstIndex:]
	if strings.Index(tags, "error=1") != -1 {
		s.Wrong = true
	}
	if strings.Index(tags, "http.status_code=") != -1 && strings.Index(tags, "http.status_code=200") == -1 {
		s.Wrong = true
	}
}

func ParseSpan(s string) *pb.Span {
	firstIndex := strings.IndexByte(s, '|')
	secondIndex := strings.IndexByte(s[firstIndex+1:], '|')
	return &pb.Span{
		Raw:       s,
		StartTime: s[firstIndex+1 : firstIndex+1+secondIndex],
	}
}
