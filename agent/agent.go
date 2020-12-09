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

	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/mingoooo/tail-based-sampling/utils"
)

var (
	TraceIDPool = &sync.Pool{}
	TracePool   = &sync.Pool{}
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
	errTidSubCh    chan *pb.TraceID
	errTidPubCh    chan *pb.TraceID
	errTidMark     *sync.Map
	curPos         uint64
	cacheLen       uint64
	flushPosTidMap map[uint64]string
	markerExit     chan bool
	readChan       chan []byte
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
		endIndex:       128 * 1024 * 1024,
		ReadLimit:      128 * 1024 * 1024,
		curContentLen:  -1,
		finishWg:       &sync.WaitGroup{},
		traceCh:        make(chan *pb.Trace, 128),
		errTidSubCh:    make(chan *pb.TraceID, 128),
		errTidPubCh:    make(chan *pb.TraceID, 128),
		errTidMark:     &sync.Map{},
		flushPosTidMap: map[uint64]string{},
		// cacheLen:       3.5 * 10000,
		cacheLen:   3.5 * 1000 * 1000,
		markerExit: make(chan bool),
		readChan:   make(chan []byte, 4896),
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

func (r *Receiver) httpGet() (io.Reader, error) {
	log.Printf("Pulling data... URL:%s \nstart index: %d\nend index: %d", r.DataURL, r.startIndex, r.endIndex)

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

func (r *Receiver) GetDataReaders() (rs []io.Reader, err error) {
	for {
		if r.curContentLen != -1 && r.curContentLen < r.ReadLimit {
			return rs, err
		}
		resp, err := r.httpGet()
		if err != nil {
			return rs, err
		}

		rs = append(rs, resp)
	}
}

func (r *Receiver) Download(resps []io.Reader) error {
	defer close(r.readChan)
	preTail := []byte{}

	for _, resp := range resps {
		for {
			reader := bufio.NewReader(resp)
			line, err := reader.ReadBytes('\n')
			if len(preTail) > 0 {
				line = append(preTail, line...)
				preTail = []byte{}
			}
			if err == io.EOF {
				preTail = line
				break
			} else if err != nil {
				return err
			}
			r.readChan <- line
		}
	}
	return nil
}

func (r *Receiver) Filter() error {
	defer r.finishWg.Done()

	for {
		line, ok := <-r.readChan
		if !ok {
			break
		}
		// log.Printf(string(line))
		// TODO: slop
		s := NewSpan(line)

		// mark flush pos if not exist
		if _, ok := r.Cache.Get(s.TraceID); !ok {
			r.flushPosTidMap[r.curPos+r.cacheLen] = s.TraceID
		}

		// stroe into cache
		r.setCache(s.TraceID, s.Raw)

		if s.Wrong {
			// log.Printf("Set trace ID: %s", s.TraceID)
			if _, ok := r.errTidMark.LoadOrStore(s.TraceID, true); !ok {
				// send wrong trace id to backend
				r.SetErrTraceID(s.TraceID)
			}
		}
		r.curPos++

		if r.curPos <= r.cacheLen {
			continue
		}

		// flush trace
		if tid, ok := r.flushPosTidMap[r.curPos]; ok {
			if _, ok := r.errTidMark.Load(tid); ok {
				// send error trace
				r.SendTraceByID(tid)
				r.errTidMark.Delete(tid)
			} else {
				r.Cache.UnsafeDelete(tid)
			}
			delete(r.flushPosTidMap, r.curPos)
		}

	}

	log.Printf("All the data has been pulled")
	r.markerExit <- true
	go r.TraceFlusher()

	log.Printf("Flush all the traces in errTidMark")
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
func (r *Receiver) TraceFlusher() {
	for {
		t, ok := <-r.errTidSubCh
		if !ok {
			return
		}
		r.SendTraceByID(t.ID)
	}
}

func (r *Receiver) TraceMarker() {
	for {
		select {
		case <-r.markerExit:
			return
		case t, ok := <-r.errTidSubCh:
			if !ok {
				return
			}
			r.errTidMark.Store(t.ID, true)
		}
	}
}

func (r *Receiver) SetErrTraceID(tid string) {
	r.errTidPubCh <- &pb.TraceID{
		ID: tid,
	}
}

func (r *Receiver) DropTraceByID(tid string) (*pb.Trace, bool) {
	spans := r.Cache.UnsafeDrop(tid)
	if len(spans) < 1 {
		return nil, false
	}

	t := &pb.Trace{
		TraceID:  tid,
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
