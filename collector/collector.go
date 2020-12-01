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
	Result   map[string]string
	ResultMu *sync.Mutex
	pb.UnsafeCollectorServer
	HTTPPort       string
	RPCPort        string
	DataPort       string
	AgentTaskChMap map[string]chan *pb.Trace
	AgentList      []string
	AgentFinishWg  *sync.WaitGroup
	AgentConfirmWg *sync.WaitGroup
	TraceCache     *sync.Map
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
		Result:         map[string]string{},
		ResultMu:       &sync.Mutex{},
		TraceCache:     &sync.Map{},
		HTTPPort:       httpPort,
		RPCPort:        rpcPort,
		AgentList:      agents,
		AgentTaskChMap: map[string]chan *pb.Trace{},
		AgentFinishWg:  &sync.WaitGroup{},
		AgentConfirmWg: &sync.WaitGroup{},
	}
	for _, a := range c.AgentList {
		c.AgentTaskChMap[a] = make(chan *pb.Trace, 128)
		c.AgentFinishWg.Add(1)
		c.AgentConfirmWg.Add(1)
	}
	return c
}

func (c Collector) SendFinish() {
	c.AgentFinishWg.Wait()

	c.ResultMu.Lock()
	result, err := json.Marshal(c.Result)
	if err != nil {
		log.Fatalln(err)
		return
	}
	c.ResultMu.Unlock()
	data := make(url.Values)
	data.Add("result", utils.ByteSliceToString(result))

	resp, err := http.PostForm(fmt.Sprintf("http://127.0.0.1:%s/api/finished", c.DataPort), data)
	if err != nil {
		log.Fatalln(err)
		return
	}
	if resp.StatusCode == http.StatusOK {
		log.Printf("Done")
		os.Exit(0)
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

func (c *Collector) StoreResult(tid string) {
	spans, ok := c.TraceCache.Load(tid)
	if !ok {
		return
	}
	m := c.GetMd5BySpans(spans.([]*pb.Span))
	// log.Printf("trace id: %s, md5: %s", tid, m)
	c.ResultMu.Lock()
	c.Result[tid] = m
	c.ResultMu.Unlock()
	// c.TraceCache.Delete(tid)
}
