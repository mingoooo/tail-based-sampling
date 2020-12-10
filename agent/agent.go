package agent

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	pb "github.com/mingoooo/tail-based-sampling/g"
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
}

type Span struct {
	Raw     string
	Wrong   bool
	TraceID string
}

type Cfg struct {
	HttpPort   string
	DataSuffix string
	ReadLimit  int64
	CacheLen   uint64
}

// New agent
func New(cfg *Cfg) (*Receiver, error) {
	r := &Receiver{
		HTTPPort:       cfg.HttpPort,
		DataSuffix:     cfg.DataSuffix,
		Cache:          newCache(),
		endIndex:       cfg.ReadLimit,
		ReadLimit:      cfg.ReadLimit,
		curContentLen:  -1,
		finishWg:       &sync.WaitGroup{},
		traceCh:        make(chan *pb.Trace, 512),
		errTidSubCh:    make(chan *pb.TraceID, 512),
		errTidPubCh:    make(chan *pb.TraceID, 512),
		errTidMark:     &sync.Map{},
		flushPosTidMap: map[uint64]string{},
		cacheLen:       cfg.CacheLen,
		markerExit:     make(chan bool),
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
				preTail = line
				break
			} else if err != nil {
				return err
			}

			// TODO: slop
			r.curPos++
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

func (r *Receiver) setCache(tid, line string) {
	r.Cache.Set(tid, line)
}
