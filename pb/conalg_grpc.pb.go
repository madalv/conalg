// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.12.4
// source: conalg.proto

package pb

import (
	context "context"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ConalgClient is the client API for Conalg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ConalgClient interface {
	FastProposeStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_FastProposeStreamClient, error)
	SlowProposeStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_SlowProposeStreamClient, error)
	RetryStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_RetryStreamClient, error)
	StableStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_StableStreamClient, error)
}

type conalgClient struct {
	cc grpc.ClientConnInterface
}

func NewConalgClient(cc grpc.ClientConnInterface) ConalgClient {
	return &conalgClient{cc}
}

func (c *conalgClient) FastProposeStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_FastProposeStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Conalg_ServiceDesc.Streams[0], "/proto.conalg.Conalg/FastProposeStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &conalgFastProposeStreamClient{stream}
	return x, nil
}

type Conalg_FastProposeStreamClient interface {
	Send(*Propose) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type conalgFastProposeStreamClient struct {
	grpc.ClientStream
}

func (x *conalgFastProposeStreamClient) Send(m *Propose) error {
	return x.ClientStream.SendMsg(m)
}

func (x *conalgFastProposeStreamClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *conalgClient) SlowProposeStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_SlowProposeStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Conalg_ServiceDesc.Streams[1], "/proto.conalg.Conalg/SlowProposeStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &conalgSlowProposeStreamClient{stream}
	return x, nil
}

type Conalg_SlowProposeStreamClient interface {
	Send(*Propose) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type conalgSlowProposeStreamClient struct {
	grpc.ClientStream
}

func (x *conalgSlowProposeStreamClient) Send(m *Propose) error {
	return x.ClientStream.SendMsg(m)
}

func (x *conalgSlowProposeStreamClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *conalgClient) RetryStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_RetryStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Conalg_ServiceDesc.Streams[2], "/proto.conalg.Conalg/RetryStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &conalgRetryStreamClient{stream}
	return x, nil
}

type Conalg_RetryStreamClient interface {
	Send(*Propose) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type conalgRetryStreamClient struct {
	grpc.ClientStream
}

func (x *conalgRetryStreamClient) Send(m *Propose) error {
	return x.ClientStream.SendMsg(m)
}

func (x *conalgRetryStreamClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *conalgClient) StableStream(ctx context.Context, opts ...grpc.CallOption) (Conalg_StableStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &Conalg_ServiceDesc.Streams[3], "/proto.conalg.Conalg/StableStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &conalgStableStreamClient{stream}
	return x, nil
}

type Conalg_StableStreamClient interface {
	Send(*Propose) error
	CloseAndRecv() (*empty.Empty, error)
	grpc.ClientStream
}

type conalgStableStreamClient struct {
	grpc.ClientStream
}

func (x *conalgStableStreamClient) Send(m *Propose) error {
	return x.ClientStream.SendMsg(m)
}

func (x *conalgStableStreamClient) CloseAndRecv() (*empty.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(empty.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ConalgServer is the server API for Conalg service.
// All implementations must embed UnimplementedConalgServer
// for forward compatibility
type ConalgServer interface {
	FastProposeStream(Conalg_FastProposeStreamServer) error
	SlowProposeStream(Conalg_SlowProposeStreamServer) error
	RetryStream(Conalg_RetryStreamServer) error
	StableStream(Conalg_StableStreamServer) error
	mustEmbedUnimplementedConalgServer()
}

// UnimplementedConalgServer must be embedded to have forward compatible implementations.
type UnimplementedConalgServer struct {
}

func (UnimplementedConalgServer) FastProposeStream(Conalg_FastProposeStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method FastProposeStream not implemented")
}
func (UnimplementedConalgServer) SlowProposeStream(Conalg_SlowProposeStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method SlowProposeStream not implemented")
}
func (UnimplementedConalgServer) RetryStream(Conalg_RetryStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method RetryStream not implemented")
}
func (UnimplementedConalgServer) StableStream(Conalg_StableStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method StableStream not implemented")
}
func (UnimplementedConalgServer) mustEmbedUnimplementedConalgServer() {}

// UnsafeConalgServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ConalgServer will
// result in compilation errors.
type UnsafeConalgServer interface {
	mustEmbedUnimplementedConalgServer()
}

func RegisterConalgServer(s grpc.ServiceRegistrar, srv ConalgServer) {
	s.RegisterService(&Conalg_ServiceDesc, srv)
}

func _Conalg_FastProposeStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ConalgServer).FastProposeStream(&conalgFastProposeStreamServer{stream})
}

type Conalg_FastProposeStreamServer interface {
	Send(*Response) error
	Recv() (*Propose, error)
	grpc.ServerStream
}

type conalgFastProposeStreamServer struct {
	grpc.ServerStream
}

func (x *conalgFastProposeStreamServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *conalgFastProposeStreamServer) Recv() (*Propose, error) {
	m := new(Propose)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Conalg_SlowProposeStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ConalgServer).SlowProposeStream(&conalgSlowProposeStreamServer{stream})
}

type Conalg_SlowProposeStreamServer interface {
	Send(*Response) error
	Recv() (*Propose, error)
	grpc.ServerStream
}

type conalgSlowProposeStreamServer struct {
	grpc.ServerStream
}

func (x *conalgSlowProposeStreamServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *conalgSlowProposeStreamServer) Recv() (*Propose, error) {
	m := new(Propose)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Conalg_RetryStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ConalgServer).RetryStream(&conalgRetryStreamServer{stream})
}

type Conalg_RetryStreamServer interface {
	Send(*Response) error
	Recv() (*Propose, error)
	grpc.ServerStream
}

type conalgRetryStreamServer struct {
	grpc.ServerStream
}

func (x *conalgRetryStreamServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *conalgRetryStreamServer) Recv() (*Propose, error) {
	m := new(Propose)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Conalg_StableStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ConalgServer).StableStream(&conalgStableStreamServer{stream})
}

type Conalg_StableStreamServer interface {
	SendAndClose(*empty.Empty) error
	Recv() (*Propose, error)
	grpc.ServerStream
}

type conalgStableStreamServer struct {
	grpc.ServerStream
}

func (x *conalgStableStreamServer) SendAndClose(m *empty.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *conalgStableStreamServer) Recv() (*Propose, error) {
	m := new(Propose)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Conalg_ServiceDesc is the grpc.ServiceDesc for Conalg service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Conalg_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.conalg.Conalg",
	HandlerType: (*ConalgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "FastProposeStream",
			Handler:       _Conalg_FastProposeStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "SlowProposeStream",
			Handler:       _Conalg_SlowProposeStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "RetryStream",
			Handler:       _Conalg_RetryStream_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "StableStream",
			Handler:       _Conalg_StableStream_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "conalg.proto",
}
