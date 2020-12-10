package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/mingoooo/tail-based-sampling/utils"
)

var (
	TracePool   = sync.Pool{New: func() interface{} { return &pb.Trace{} }}
	TraceIDPool = sync.Pool{New: func() interface{} { return &pb.TraceID{} }}
	ReadPos     = int64(10 * 1024 * 1024)
)

// Receiver struct
type Receiver struct {
	HTTPPort       string
	DataURL        string
	DataPort       string
	DataSuffix     string
	startIndex     int
	endIndex       int
	ReadLimit      int
	curContentLen  int
	finishWg       *sync.WaitGroup
	Cache          *Cache
	Postman        *Postman
	traceCh        chan *pb.Trace
	errTidSubCh    chan *pb.TraceID
	errTidPubCh    chan *pb.TraceID
	errTidMark     *sync.Map
	curPos         uint
	cacheLen       uint
	flushPosTidMap map[uint]string
	markerExit     chan bool
	ringBuf        []byte
	bufferSize     int
	respChan       chan *respBlock
	ReadLineChan   chan *Span
}

type respBlock struct {
	Reader     io.Reader
	StartIdx   int
	EndIdx     int
	ExitSignal chan struct{}
	Done       sync.WaitGroup
	readIdx    int
	writeIdx   int
}

type Span struct {
	startIdx int
	endIdx   int
	buf      []byte
	Wrong    bool
	TraceID  string
}

// New agent
func New(httpPort string, dataSuffix string) (*Receiver, error) {
	r := &Receiver{
		HTTPPort:       httpPort,
		DataSuffix:     dataSuffix,
		Cache:          newCache(128 * 1024 * 1024),
		endIndex:       256 * 1024 * 1024,
		ReadLimit:      256 * 1024 * 1024,
		curContentLen:  -1,
		finishWg:       &sync.WaitGroup{},
		traceCh:        make(chan *pb.Trace, 128),
		errTidSubCh:    make(chan *pb.TraceID, 128),
		errTidPubCh:    make(chan *pb.TraceID, 128),
		errTidMark:     &sync.Map{},
		flushPosTidMap: map[uint]string{},
		cacheLen:       3.3 * 1000 * 1000,
		markerExit:     make(chan bool),
		ringBuf:        make([]byte, 2*1024*1024*1024),
		bufferSize:     2 * 1024 * 1024 * 1024,
		ReadLineChan:   make(chan *Span, 2*1024),
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

func (r *Receiver) httpGet() (*respBlock, error) {
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

	rs := &respBlock{
		Reader:     resp.Body,
		ExitSignal: make(chan struct{}),
		Done:       sync.WaitGroup{},
		StartIdx:   r.startIndex,
		EndIdx:     r.endIndex,
		readIdx:    int(ReadPos),
		writeIdx:   int(ReadPos),
	}
	r.startIndex, r.endIndex = r.endIndex+1, r.endIndex+r.ReadLimit
	r.curContentLen = int(resp.ContentLength)
	ReadPos += int64(resp.ContentLength)
	if int(ReadPos)+r.ReadLimit > r.bufferSize {
		ReadPos = 10 * 1024 * 1024
	}

	rs.Done.Add(1)
	return rs, err

}

func (r *Receiver) GetDataReaders() (rs []*respBlock, err error) {
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

func (r *respBlock) fill(buf []byte) (err error) {
	log.Printf("Start fill")
	defer r.Done.Done()
	defer log.Printf("Exit fill")
	// readSize := 8 * 1024 * 1024

	n := 0
	for {
		select {
		case <-r.ExitSignal:
			return nil
		default:
			n, err = r.Reader.Read(buf[r.writeIdx:])
			r.writeIdx += n

			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			if n > 0 {
				continue
			}
		}
	}
}

func (r *Receiver) StartDownload(tasks []*respBlock) error {
	log.Printf("Start download")
	for _, resp := range tasks {
		if err := resp.fill(r.ringBuf); err != nil {
			return err
		}
	}
	log.Printf("Exit download")
	return nil
}

func (r *Receiver) readLines(resp []*respBlock) (err error) {
	defer close(r.ReadLineChan)
	readSize := 64 * 1024 * 1024

	offset := 0
	for _, reader := range resp {
		close(reader.ExitSignal)
		reader.Done.Wait()
		log.Printf("Read lines")

		n := 0
		si := reader.StartIdx
		ri := reader.readIdx
		wi := reader.writeIdx
		ei := reader.EndIdx

		if ri > si {
			offset = r.splitLine(ri, wi, offset)
		}

		for {

			if ri+readSize <= ei {
				n, err = io.ReadAtLeast(reader.Reader, r.ringBuf[wi:], readSize)
			} else {
				n, err = reader.Reader.Read(r.ringBuf[wi:])
			}

			if n > 0 {
				offset = r.splitLine(wi, wi+n, offset)
				ri += n
			}

			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
		}

		log.Printf("Exit read lines")
	}
	return nil
}

func (r *Receiver) splitLine(si, ei, offset int) int {
	for si != ei {
		if offset != 0 {
			si -= offset
			if si < 10*1024*1024 {
				copy(r.ringBuf[si-offset:], r.ringBuf[r.bufferSize-offset:])
			}
			offset = 0
		}
		n := bytes.IndexByte(r.ringBuf[si:ei], '\n')

		if n < 0 {
			return ei - si
		} else if n == 0 {
			return 0
		}

		r.ReadLineChan <- NewSpan(r.ringBuf, si, si+n+1)
		si += n + 1
	}
	return 0
}

func (r *Receiver) Filter(rs []*respBlock) error {
	defer r.finishWg.Done()
	log.Printf("Filtering...")

	for {
		s, ok := <-r.ReadLineChan
		if !ok {
			break
		}

		// TODO: slop
		r.curPos++

		// mark flush pos if not exist
		if _, ok := r.Cache.Get(s.TraceID); !ok {
			r.flushPosTidMap[r.curPos+r.cacheLen] = s.TraceID
		}

		// stroe into cache
		r.setCache(s)

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
	// r.dataPool.Put(reader)

	log.Printf("All the data has been pulled")
	r.markerExit <- true
	go r.TraceFlusher()

	log.Printf("Flush all the traces in errTidMark")
	r.errTidMark.Range(func(key, _ interface{}) bool {
		r.SendTraceByID(key.(string))
		return true
	})

	close(r.errTidPubCh)
	log.Printf("Exit filter")
	return nil
}

func (r *Receiver) flushTrace(tid string) {

	if _, ok := r.errTidMark.Load(tid); ok {
		// send error trace
		r.SendTraceByID(tid)
		r.errTidMark.Delete(tid)
	} else {
		r.Cache.UnsafeDelete(tid)
	}
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
	spans := r.Cache.Drop(s.TraceID)
	if len(spans) > 0 {
		for _, s := range spans {
			trace.SpanList = append(trace.SpanList, ParseSpan(s))
		}
	}
	// r.Cache.Delete(s.TraceID)
	r.traceCh <- trace
}

func (r *Receiver) setCache(s *Span) {
	r.Cache.Set(s.TraceID, s)
}

func NewSpan(buf []byte, startIdx, endIdx int) *Span {
	s := &Span{
		startIdx: startIdx,
		endIdx:   endIdx,
		buf:      buf,
	}
	line := s.Get()
	tidEndIndex := bytes.IndexByte(line, '|')
	if tidEndIndex == -1 {
		log.Fatalf("Parse trace id error: %s", string(line))
	}
	s.TraceID = utils.ByteSliceToString(line[:tidEndIndex])

	// filter tag
	tagFirstIndex := bytes.LastIndexByte(line, '|')
	if tagFirstIndex == -1 {
		log.Fatalf("Parse trace tag error: %s", string(line))
	}

	tags := line[tagFirstIndex:]
	if bytes.Index(tags, utils.StringToBytes("error=1")) != -1 {
		s.Wrong = true
		return s
	}
	if bytes.Index(tags, utils.StringToBytes("http.status_code=")) != -1 && bytes.Index(tags, utils.StringToBytes("http.status_code=200")) == -1 {
		s.Wrong = true
	}
	return s
}
func (s Span) Get() []byte {
	return s.buf[s.startIdx:s.endIdx]
}

func ParseSpan(s *Span) *pb.Span {
	line := s.Get()
	firstIndex := bytes.IndexByte(line, '|')
	secondIndex := bytes.IndexByte(line[firstIndex+1:], '|')
	return &pb.Span{
		Raw:       line,
		StartTime: utils.ByteSliceToString((line[firstIndex+1 : firstIndex+1+secondIndex])),
	}
}
