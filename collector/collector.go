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
	"os"
	"sort"
	"strings"
	"sync"

	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/mingoooo/tail-based-sampling/utils"
)

// Collector struct
type Collector struct {
	pb.UnsafeCollectorServer
	Result           map[string]string
	HTTPPort         string
	RPCPort          string
	DataPort         string
	AgentTaskChMap   map[string]chan *pb.Trace
	AgentList        []string
	AgentFinishWg    *sync.WaitGroup
	AgentConfirmWg   *sync.WaitGroup
	TraceCache       map[string][]*pb.Span
	TraceCacheLocker *sync.RWMutex
}

// Run is entrypoint
func (c *Collector) Run(ctx context.Context, cancel context.CancelFunc) error {
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
		TraceCache:       map[string][]*pb.Span{},
		TraceCacheLocker: &sync.RWMutex{},
		HTTPPort:         httpPort,
		RPCPort:          rpcPort,
		AgentList:        agents,
		AgentTaskChMap:   map[string]chan *pb.Trace{},
		AgentFinishWg:    &sync.WaitGroup{},
		AgentConfirmWg:   &sync.WaitGroup{},
	}
	for _, a := range c.AgentList {
		c.AgentTaskChMap[a] = make(chan *pb.Trace, 128)
		c.AgentFinishWg.Add(1)
		c.AgentConfirmWg.Add(1)
	}
	return c
}

func (c Collector) SendFinish() {
	defer os.Exit(0)
	// defer trace.Stop()
	c.AgentFinishWg.Wait()
	c.FlushResult()

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

func (c *Collector) FlushResult() {
	for tid, spans := range c.TraceCache {
		m := c.GetMd5BySpans(spans)
		c.Result[tid] = m
	}
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
