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
	"time"

	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/mingoooo/tail-based-sampling/utils"
)

// Receiver struct
type Receiver struct {
	HttpPort      string
	DataURL       string
	DataPort      string
	DataSuffix    string
	startIndex    int64
	endIndex      int64
	ReadLimit     int64
	curContentLen int64
	finishWg      *sync.WaitGroup
	Cache         *Cache
	Postman       *Postman
	traceCh       chan *pb.Trace
	tidCh         chan *pb.TraceID
}

type Span struct {
	Raw     string
	Wrong   bool
	TraceID string
}

// New agent
func New(httpPort string, dataSuffix string) (*Receiver, error) {
	r := &Receiver{
		HttpPort:      httpPort,
		DataSuffix:    dataSuffix,
		Cache:         newCache(40000),
		endIndex:      512 * 1024 * 1024,
		ReadLimit:     512 * 1024 * 1024,
		curContentLen: -1,
		finishWg:      &sync.WaitGroup{},
		traceCh:       make(chan *pb.Trace, 128),
		tidCh:         make(chan *pb.TraceID, 128),
	}
	var err error
	r.Postman, err = NewPostman(fmt.Sprintf("127.0.0.1:%s", "8003"), r.HttpPort)
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
			return nil
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

			s := NewSpan(line)
			r.setCache(s.TraceID, s.Raw)
			if s.Wrong {
				go r.SendTraceBySpan(s)
			}
		}
	}
}

func (r *Receiver) TraceSearcher() {
	for {
		t, ok := <-r.tidCh
		if !ok {
			return
		}
		if trace, ok := r.GetTraceByID(t.ID); ok {
			r.traceCh <- trace
		}
	}
}

func (r *Receiver) GetTraceByID(tid string) (*pb.Trace, bool) {
	spans := r.Cache.Get(tid)
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
	time.Sleep(5 * time.Second)
	spans := r.Cache.Get(s.TraceID)
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
