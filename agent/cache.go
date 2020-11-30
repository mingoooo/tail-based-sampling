package agent

import (
	"sync"
)

// traceId：全局唯一的Id，用作整个链路的唯一标识与组装
// startTime：调用的开始时间
// spanId: 调用链中某条数据(span)的id
// parentSpanId: 调用链中某条数据(span)的父亲id，头节点的span的parantSpanId为0
// duration：调用耗时
// serviceName：调用的服务名
// spanName：调用的埋点名
// host：机器标识，比如ip，机器名
// tags: 链路信息中tag信息，存在多个tag的key和value信息。格式为key1=val1&key2=val2&key3=val3 比如 http.status_code=200&error=1
// type Span struct {
// 	TraceID      string
// 	StartTime    string
// 	SpanID       string
// 	ParentSpanID string
// 	StatusCode   string
// 	Error        string
// }

type Cache struct {
	// tidToSids map[string][]string
	// sidToSpan map[string]string
	sync.RWMutex
	tidToSpans map[string][]string
	tidChan    chan string
	size       int
}

func newCache(size int) *Cache {
	t := &Cache{
		tidToSpans: map[string][]string{},
		tidChan:    make(chan string, size),
		size:       size,
	}
	return t
}

func (c *Cache) cleanWorker() {
	for i := 0; i <= c.size/2; i++ {
		tid := <-c.tidChan
		c.Delete(tid)
	}
}

func (c *Cache) Delete(tid string) {
	c.Lock()
	delete(c.tidToSpans, tid)
	c.Unlock()
}

func (c *Cache) Set(key, val string) {
	spans, exist := c.tidToSpans[key]
	if !exist {
		c.tidChan <- key
		if len(c.tidChan) >= c.size {
			c.cleanWorker()
		}
	}
	c.Lock()
	c.tidToSpans[key] = append(spans, val)
	c.Unlock()
}

func (c *Cache) Get(key string) []string {
	c.RLock()
	val := c.tidToSpans[key]
	c.RUnlock()
	return val
}
