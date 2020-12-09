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

func (p Postman) ErrTraceIdPublisher(ch <-chan *pb.TraceID) error {
	client := pb.NewCollectorClient(p.conn)
	c := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("FromAgent", p.AgentName))
	stream, err := client.SetErrTraceID(c)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	for {
		tid, ok := <-ch
		if !ok {
			return nil
		}

		err := stream.Send(tid)
		if err == io.EOF {
			log.Printf("%s received EOF in TracePublisher", p.AgentName)
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func (p Postman) TracePublisher(ch <-chan *pb.Trace) error {
	client := pb.NewCollectorClient(p.conn)
	c := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("FromAgent", p.AgentName))
	stream, err := client.SendTrace(c)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	for {
		trace, ok := <-ch
		if !ok {
			return nil
		}

		log.Println(trace.TraceID)
		err := stream.Send(trace)
		if err == io.EOF {
			log.Printf("%s received EOF in TracePublisher", p.AgentName)
			return nil
		}
		if err != nil {
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
			log.Printf("%s received EOF in TraceIDSubscriber", p.AgentName)
			return nil
		}
		if err != nil {
			return err
		}
		// log.Printf("Receive trace id: %s", in.ID)
		ch <- in
	}
}

func (p *Postman) ConfirmFinish(traceCh chan *pb.Trace, tidSubCh chan *pb.TraceID, tidPubCh chan *pb.TraceID) error {
	client := pb.NewCollectorClient(p.conn)
	c := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("FromAgent", p.AgentName))
	stream, err := client.ConfirmFinish(c)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	// wait for set remaning trace ids
	for {
		if len(tidPubCh) < 1 {
			close(tidPubCh)
			break
		}
	}

	log.Printf("Send confirm")
	err = stream.Send(&pb.AgentStatus{Status: pb.AgentStatus_CONFIRM})
	if err != nil {
		return err
	}

	// Wait all the agents confirm and transfer trace ids
	ok, err := stream.Recv()
	if err == io.EOF {
		log.Printf("%s received EOF in ConfirmFinish", p.AgentName)
		return nil
	}
	if err != nil {
		return err
	}

	// Wait for receive and send trace from other agents
	for {
		if ok.Ok && len(tidSubCh) < 1 && len(traceCh) < 1 {
			close(tidSubCh)
			close(traceCh)
			return nil
		}
		// log.Printf("confirm tid chan len: %d", len(tidCh))
		// log.Printf("confirm trace chan len: %d", len(traceCh))
	}
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
