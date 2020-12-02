package agent

import (
	"context"
	"io"
	"log"
	"time"

	pb "github.com/mingoooo/tail-based-sampling/g"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Postman struct {
	collectorURL string
	conn         *grpc.ClientConn
	AgentName    string
}

func NewPostman(url, agentName string) (*Postman, error) {
	p := &Postman{
		collectorURL: url,
		AgentName:    agentName,
	}
	err := p.connect()
	return p, err
}

type WrongTrace struct {
}

func (p Postman) TracePublisher(ch <-chan *pb.Trace) error {
	client := pb.NewCollectorClient(p.conn)
	c := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("FromAgent", p.AgentName))
	stream, err := client.SendTrace(c)
	if err != nil {
		return err
	}
	defer stream.CloseSend()
	// m := map[string]bool{}

	for {
		trace, ok := <-ch
		if !ok {
			return nil
		}

		// TODO: find out double send and need to fix
		// if m[trace.TraceID] {
		// 	continue
		// }
		// m[trace.TraceID] = true
		// log.Println(trace.TraceID)
		if err := stream.Send(trace); err != nil {
			return err
		}
	}
}

func (p Postman) TraceIDSubscriber(ch chan<- *pb.TraceID) error {
	client := pb.NewCollectorClient(p.conn)
	c := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("FromAgent", p.AgentName))
	stream, err := client.SubscribeTraceID(c, &emptypb.Empty{})
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		// log.Printf("Receive trace id: %s", in.ID)
		ch <- in
	}
	return nil
}

func (p *Postman) ConfirmFinish(traceCh chan *pb.Trace, tidCh chan *pb.TraceID) error {
	// TODO: TEST
	// time.Sleep(5 * time.Second)
	client := pb.NewCollectorClient(p.conn)
	c := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("FromAgent", p.AgentName))
	stream, err := client.ConfirmFinish(c)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	// Wating other agent
	log.Printf("Send confirm")
	err = stream.Send(&pb.AgentStatus{Status: pb.AgentStatus_CONFIRM})
	if err != nil {
		return err
	}

	ok, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	for {
		// fmt.Println(len(tidCh))
		// fmt.Println(len(traceCh))
		if ok.Ok && len(tidCh) < 1 && len(traceCh) < 1 {
			close(tidCh)
			close(traceCh)
			break
		}
	}

	// finish
	log.Printf("Send closed")
	err = stream.Send(&pb.AgentStatus{Status: pb.AgentStatus_CLOSED})
	if err != nil {
		return err
	}

	ok, err = stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *Postman) connect() (err error) {
	kacp := keepalive.ClientParameters{
		Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
		Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}
	p.conn, err = grpc.Dial(p.collectorURL, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	return
}
