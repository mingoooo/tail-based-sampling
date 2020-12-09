package collector

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"sort"
	"strings"
	"sync"

	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/mingoooo/tail-based-sampling/utils"
)

type result struct {
	tid   string
	spans []*pb.Span
}

// Collector struct
type Collector struct {
	pb.UnsafeCollectorServer
	Result           map[string]string
	resultCh         chan result
	resultLocker     *sync.RWMutex
	resultHandlerWg  *sync.WaitGroup
	HTTPPort         string
	RPCPort          string
	DataPort         string
	AgentTaskChMap   map[string]chan *pb.TraceID
	AgentList        []string
	AgentFinishWg    *sync.WaitGroup
	AgentConfirmWg   *sync.WaitGroup
	TraceCache       map[string][]*pb.Span
	TraceCacheLocker *sync.RWMutex
	closeAgentCh     chan bool
}

// Run is entrypoint
func (c *Collector) Run(ctx context.Context, cancel context.CancelFunc) error {
	for i := 0; i <= runtime.NumCPU(); i++ {
		go c.ResultHandler()
	}
	go c.closeAllAgentCh()
	go func() {
		err := c.RunRPCSvr()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	c.RunHTTPSvr()
	return nil
}

// New collector
func New(httpPort, rpcPort string, agents []string) *Collector {
	c := &Collector{
		Result:           map[string]string{},
		resultCh:         make(chan result, 4796),
		resultLocker:     &sync.RWMutex{},
		resultHandlerWg:  &sync.WaitGroup{},
		TraceCache:       map[string][]*pb.Span{},
		TraceCacheLocker: &sync.RWMutex{},
		HTTPPort:         httpPort,
		RPCPort:          rpcPort,
		AgentList:        agents,
		AgentTaskChMap:   map[string]chan *pb.TraceID{},
		AgentFinishWg:    &sync.WaitGroup{},
		AgentConfirmWg:   &sync.WaitGroup{},
		closeAgentCh:     make(chan bool, 8),
	}
	for _, a := range c.AgentList {
		c.AgentTaskChMap[a] = make(chan *pb.TraceID, 128)
		c.AgentFinishWg.Add(1)
		c.AgentConfirmWg.Add(1)
	}
	return c
}

func (c *Collector) closeAllAgentCh() {
	<-c.closeAgentCh
	for {
		m := 0
		for _, ch := range c.AgentTaskChMap {
			if len(ch) < 1 {
				m++
			}
		}
		if m == len(c.AgentTaskChMap) {
			for a, ch := range c.AgentTaskChMap {
				log.Printf("close agent channel: %s", a)
				close(ch)
			}
			return
		}
	}

}

func (c Collector) SendFinish() {
	// defer os.Exit(0)
	// defer trace.Stop()
	c.AgentFinishWg.Wait()
	log.Printf("Flush result")
	// time.Sleep(5 * time.Second)
	c.TraceCacheLocker.Lock()
	c.FlushResult()
	c.TraceCacheLocker.Unlock()

	// TODO: Manually check md5 for testing
	/**
	checkSum := map[string]string{}
	if dir, err := os.Getwd(); err == nil {
		if checkFile, err := ioutil.ReadFile(fmt.Sprintf("%s/scoring/checkSum.data", dir)); err == nil {
			json.Unmarshal(checkFile, &checkSum)
			for tid, mdf := range checkSum {
				if m, ok := c.Result[tid]; ok {
					if m != mdf {
						log.Printf("Wrong md5 of trace: %s", tid)
						log.Println(c.TraceCache[tid])
					}
				} else {
					log.Printf("Missing trace id: %s", tid)
				}
			}
		} else {
			log.Println(err)
		}
	} else {
		log.Println(err)
	}
	**/

	result, err := json.Marshal(c.Result)
	if err != nil {
		log.Fatalln(err)
		return
	}
	data := make(url.Values)
	data.Add("result", utils.ByteSliceToString(result))

	resp, err := http.PostForm(fmt.Sprintf("http://127.0.0.1:%s/api/finished", c.DataPort), data)
	if err != nil {
		log.Fatalln(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		log.Printf("Done")
		return
	}
}

func (c *Collector) GetMd5BySpans(spans []*pb.Span) string {
	sort.Slice(spans, func(i, j int) bool {
		return spans[i].StartTime < spans[j].StartTime
	})
	m := md5.New()
	// log.Println(len(spans))
	for _, s := range spans {
		// log.Printf(s.StartTime)
		// log.Printf(s.Raw)
		m.Write([]byte(s.Raw))
	}
	return strings.ToUpper(hex.EncodeToString(m.Sum(nil)))
}

func (c *Collector) ResultHandler() {
	c.resultHandlerWg.Add(1)
	defer c.resultHandlerWg.Done()
	for {
		res, ok := <-c.resultCh
		if !ok {
			return
		}
		m := c.GetMd5BySpans(res.spans)
		c.resultLocker.Lock()
		c.Result[res.tid] = m
		c.resultLocker.Unlock()
	}

}

func (c *Collector) FlushResult() {
	for tid, spans := range c.TraceCache {
		c.resultCh <- result{
			tid:   tid,
			spans: spans,
		}
	}
	close(c.resultCh)
	c.resultHandlerWg.Wait()
}

func (c *Collector) FlushTrace(tid string) {
	if spans, ok := c.TraceCache[tid]; ok {
		m := c.GetMd5BySpans(spans)
		c.Result[tid] = m
	}
}

func (c *Collector) StoreResult(tid string) {
	spans, ok := c.TraceCache[tid]
	if !ok {
		return
	}
	m := c.GetMd5BySpans(spans)
	// log.Printf("trace id: %s, md5: %s", tid, m)
	c.Result[tid] = m
	// c.TraceCache.Delete(tid)
}
