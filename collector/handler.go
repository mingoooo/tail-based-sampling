package collector

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	pb "github.com/mingoooo/tail-based-sampling/g"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"

	// "google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

func (c *Collector) ReadyHTTPHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (c *Collector) SetParamHandler(ctx *fasthttp.RequestCtx) {
	c.DataPort = string(ctx.QueryArgs().Peek("port"))
	go c.SendFinish()
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// RunHTTPSvr Run HTTP server
func (c *Collector) RunHTTPSvr() {
	//TODO handel context
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/ready":
			c.ReadyHTTPHandler(ctx)
		case "/setParameter":
			c.SetParamHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	fasthttp.ListenAndServe(fmt.Sprintf(":%s", c.HTTPPort), m)
}

func (c Collector) getAgentNameFromMetadata(ctx context.Context) (string, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ok
	}
	a := md.Get("FromAgent")
	if len(a) < 1 {
		return "", false
	}
	return a[0], ok
}

func (c *Collector) SendTrace(stream pb.Collector_SendTraceServer) error {
	// get agent name
	fromAgent, ok := c.getAgentNameFromMetadata(stream.Context())
	if !ok {
		return errors.New("Missing metadata")
	}

	for {
		// receive
		trace, err := stream.Recv()
		if err == io.EOF {
			// return stream.SendAndClose(&empty.Empty{})
			return nil
		}
		if err != nil {
			return err
		}

		// store trace
		if len(trace.SpanList) > 0 {
			// log.Printf("Store trace id of spans: %s", trace.TraceID)
			c.TraceCacheLocker.Lock()
			// flush trace to result
			if spans, ok := c.TraceCache[trace.TraceID]; ok {
				c.TraceCache[trace.TraceID] = append(spans, trace.SpanList...)
				// c.FlushTrace(trace.TraceID)
				// delete(c.TraceCache, trace.TraceID)
			} else {
				c.TraceCache[trace.TraceID] = trace.SpanList
			}
			c.TraceCacheLocker.Unlock()
		} else {
			// send to other agent
			for agent, ch := range c.AgentTaskChMap {
				if agent != fromAgent {
					ch <- trace
				}
			}
		}
	}

}
func (c *Collector) SubscribeTraceID(_ *empty.Empty, stream pb.Collector_SubscribeTraceIDServer) error {
	// get agent name
	fromAgent, ok := c.getAgentNameFromMetadata(stream.Context())
	if !ok {
		return errors.New("Missing metadata")
	}
	for {
		// send
		t := <-c.AgentTaskChMap[fromAgent]
		// log.Printf("Send trace id: %s", t.TraceID)
		go func() {
			stream.Send(&pb.TraceID{ID: t.TraceID})
		}()
	}
}

func (c *Collector) ConfirmFinish(stream pb.Collector_ConfirmFinishServer) error {
	for {
		inStatus, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		switch inStatus.Status {
		case pb.AgentStatus_CLOSED:
			c.AgentFinishWg.Done()
			stream.Send(&pb.OK{Ok: true})
		case pb.AgentStatus_CONFIRM:
			c.AgentConfirmWg.Done()
			c.AgentConfirmWg.Wait()
			stream.Send(&pb.OK{Ok: true})
		}
	}
}

func (c *Collector) RunRPCSvr() error {
	//TODO passing context
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%s", c.RPCPort))
	if err != nil {
		return err
	}
	kaep := keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,            // Allow pings even when there are no active streams
	}
	kasp := keepalive.ServerParameters{
		MaxConnectionIdle: 30 * time.Minute, // If a client is idle for 15 seconds, send a GOAWAY
		// MaxConnectionAge:      30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		// MaxConnectionAgeGrace: 60 * time.Second, // Allow 60 seconds for pending RPCs to complete before forcibly closing connections
		// Time:                  5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout: 1 * time.Second, // Wait 1 second for the ping ack before assuming the connection is dead
	}
	grpcServer := grpc.NewServer(grpc.KeepaliveEnforcementPolicy(kaep), grpc.KeepaliveParams(kasp))
	// grpcServer := grpc.NewServer(grpc.EmptyServerOption{})
	pb.RegisterCollectorServer(grpcServer, c)
	return grpcServer.Serve(lis)
}
