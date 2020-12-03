// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package g

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// CollectorClient is the client API for Collector service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CollectorClient interface {
	SendTrace(ctx context.Context, opts ...grpc.CallOption) (Collector_SendTraceClient, error)
	SetErrTraceID(ctx context.Context, opts ...grpc.CallOption) (Collector_SetErrTraceIDClient, error)
	SubscribeTraceID(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (Collector_SubscribeTraceIDClient, error)
	ConfirmFinish(ctx context.Context, opts ...grpc.CallOption) (Collector_ConfirmFinishClient, error)
}

type collectorClient struct {
	cc grpc.ClientConnInterface
}

func NewCollectorClient(cc grpc.ClientConnInterface) CollectorClient {
	return &collectorClient{cc}
}

func (c *collectorClient) SendTrace(ctx context.Context, opts ...grpc.CallOption) (Collector_SendTraceClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Collector_serviceDesc.Streams[0], "/g.Collector/SendTrace", opts...)
	if err != nil {
		return nil, err
	}
	x := &collectorSendTraceClient{stream}
	return x, nil
}

type Collector_SendTraceClient interface {
	Send(*Trace) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type collectorSendTraceClient struct {
	grpc.ClientStream
}

func (x *collectorSendTraceClient) Send(m *Trace) error {
	return x.ClientStream.SendMsg(m)
}

func (x *collectorSendTraceClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *collectorClient) SetErrTraceID(ctx context.Context, opts ...grpc.CallOption) (Collector_SetErrTraceIDClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Collector_serviceDesc.Streams[1], "/g.Collector/SetErrTraceID", opts...)
	if err != nil {
		return nil, err
	}
	x := &collectorSetErrTraceIDClient{stream}
	return x, nil
}

type Collector_SetErrTraceIDClient interface {
	Send(*TraceID) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type collectorSetErrTraceIDClient struct {
	grpc.ClientStream
}

func (x *collectorSetErrTraceIDClient) Send(m *TraceID) error {
	return x.ClientStream.SendMsg(m)
}

func (x *collectorSetErrTraceIDClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *collectorClient) SubscribeTraceID(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (Collector_SubscribeTraceIDClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Collector_serviceDesc.Streams[2], "/g.Collector/SubscribeTraceID", opts...)
	if err != nil {
		return nil, err
	}
	x := &collectorSubscribeTraceIDClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Collector_SubscribeTraceIDClient interface {
	Recv() (*TraceID, error)
	grpc.ClientStream
}

type collectorSubscribeTraceIDClient struct {
	grpc.ClientStream
}

func (x *collectorSubscribeTraceIDClient) Recv() (*TraceID, error) {
	m := new(TraceID)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *collectorClient) ConfirmFinish(ctx context.Context, opts ...grpc.CallOption) (Collector_ConfirmFinishClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Collector_serviceDesc.Streams[3], "/g.Collector/ConfirmFinish", opts...)
	if err != nil {
		return nil, err
	}
	x := &collectorConfirmFinishClient{stream}
	return x, nil
}

type Collector_ConfirmFinishClient interface {
	Send(*AgentStatus) error
	Recv() (*OK, error)
	grpc.ClientStream
}

type collectorConfirmFinishClient struct {
	grpc.ClientStream
}

func (x *collectorConfirmFinishClient) Send(m *AgentStatus) error {
	return x.ClientStream.SendMsg(m)
}

func (x *collectorConfirmFinishClient) Recv() (*OK, error) {
	m := new(OK)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// CollectorServer is the server API for Collector service.
// All implementations must embed UnimplementedCollectorServer
// for forward compatibility
type CollectorServer interface {
	SendTrace(Collector_SendTraceServer) error
	SetErrTraceID(Collector_SetErrTraceIDServer) error
	SubscribeTraceID(*empty.Empty, Collector_SubscribeTraceIDServer) error
	ConfirmFinish(Collector_ConfirmFinishServer) error
	mustEmbedUnimplementedCollectorServer()
}

// UnimplementedCollectorServer must be embedded to have forward compatible implementations.
type UnimplementedCollectorServer struct {
}

func (UnimplementedCollectorServer) SendTrace(Collector_SendTraceServer) error {
	return status.Errorf(codes.Unimplemented, "method SendTrace not implemented")
}
func (UnimplementedCollectorServer) SetErrTraceID(Collector_SetErrTraceIDServer) error {
	return status.Errorf(codes.Unimplemented, "method SetErrTraceID not implemented")
}
func (UnimplementedCollectorServer) SubscribeTraceID(*empty.Empty, Collector_SubscribeTraceIDServer) error {
	return status.Errorf(codes.Unimplemented, "method SubscribeTraceID not implemented")
}
func (UnimplementedCollectorServer) ConfirmFinish(Collector_ConfirmFinishServer) error {
	return status.Errorf(codes.Unimplemented, "method ConfirmFinish not implemented")
}
func (UnimplementedCollectorServer) mustEmbedUnimplementedCollectorServer() {}

// UnsafeCollectorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CollectorServer will
// result in compilation errors.
type UnsafeCollectorServer interface {
	mustEmbedUnimplementedCollectorServer()
}

func RegisterCollectorServer(s grpc.ServiceRegistrar, srv CollectorServer) {
	s.RegisterService(&_Collector_serviceDesc, srv)
}

func _Collector_SendTrace_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CollectorServer).SendTrace(&collectorSendTraceServer{stream})
}

type Collector_SendTraceServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*Trace, error)
	grpc.ServerStream
}

type collectorSendTraceServer struct {
	grpc.ServerStream
}

func (x *collectorSendTraceServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *collectorSendTraceServer) Recv() (*Trace, error) {
	m := new(Trace)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Collector_SetErrTraceID_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CollectorServer).SetErrTraceID(&collectorSetErrTraceIDServer{stream})
}

type Collector_SetErrTraceIDServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*TraceID, error)
	grpc.ServerStream
}

type collectorSetErrTraceIDServer struct {
	grpc.ServerStream
}

func (x *collectorSetErrTraceIDServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *collectorSetErrTraceIDServer) Recv() (*TraceID, error) {
	m := new(TraceID)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Collector_SubscribeTraceID_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(empty.Empty)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CollectorServer).SubscribeTraceID(m, &collectorSubscribeTraceIDServer{stream})
}

type Collector_SubscribeTraceIDServer interface {
	Send(*TraceID) error
	grpc.ServerStream
}

type collectorSubscribeTraceIDServer struct {
	grpc.ServerStream
}

func (x *collectorSubscribeTraceIDServer) Send(m *TraceID) error {
	return x.ServerStream.SendMsg(m)
}

func _Collector_ConfirmFinish_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CollectorServer).ConfirmFinish(&collectorConfirmFinishServer{stream})
}

type Collector_ConfirmFinishServer interface {
	Send(*OK) error
	Recv() (*AgentStatus, error)
	grpc.ServerStream
}

type collectorConfirmFinishServer struct {
	grpc.ServerStream
}

func (x *collectorConfirmFinishServer) Send(m *OK) error {
	return x.ServerStream.SendMsg(m)
}

func (x *collectorConfirmFinishServer) Recv() (*AgentStatus, error) {
	m := new(AgentStatus)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Collector_serviceDesc = grpc.ServiceDesc{
	ServiceName: "g.Collector",
	HandlerType: (*CollectorServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SendTrace",
			Handler:       _Collector_SendTrace_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "SetErrTraceID",
			Handler:       _Collector_SetErrTraceID_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "SubscribeTraceID",
			Handler:       _Collector_SubscribeTraceID_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "ConfirmFinish",
			Handler:       _Collector_ConfirmFinish_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "collector.proto",
}
