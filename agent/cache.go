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
	sync.RWMutex
	tidToSpans  map[string][]string
}

func newCache() *Cache {
	t := &Cache{
		tidToSpans: map[string][]string{},
	}
	return t
}

func (c *Cache) Delete(tid string) {
	// log.Printf("Delete trace id: %s", tid)
	c.Lock()
	delete(c.tidToSpans, tid)
	c.Unlock()
}

func (c *Cache) UnsafeDelete(tid string) {
	// log.Printf("Delete trace id: %s", tid)
	delete(c.tidToSpans, tid)
}

func (c *Cache) Set(tid, span string) {
	c.Lock()
	spans := c.tidToSpans[tid]
	c.tidToSpans[tid] = append(spans, span)
	c.Unlock()
}

func (c *Cache) Get(key string) ([]string, bool) {
	c.RLock()
	val, ok := c.tidToSpans[key]
	c.RUnlock()
	return val, ok
}

func (c *Cache) Drop(key string) []string {
	// log.Printf("Drop trace id: %s", key)
	c.Lock()
	val := c.tidToSpans[key]
	delete(c.tidToSpans, key)
	c.Unlock()
	return val
}

func (c *Cache) UnsafeDrop(key string) []string {
	val := c.tidToSpans[key]
	delete(c.tidToSpans, key)
	return val
}
