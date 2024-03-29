// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package is2fs

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// FileRefSyncClient is the client API for FileRefSync service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type FileRefSyncClient interface {
	SyncUser(ctx context.Context, in *SyncUserRequest, opts ...grpc.CallOption) (FileRefSync_SyncUserClient, error)
}

type fileRefSyncClient struct {
	cc grpc.ClientConnInterface
}

func NewFileRefSyncClient(cc grpc.ClientConnInterface) FileRefSyncClient {
	return &fileRefSyncClient{cc}
}

func (c *fileRefSyncClient) SyncUser(ctx context.Context, in *SyncUserRequest, opts ...grpc.CallOption) (FileRefSync_SyncUserClient, error) {
	stream, err := c.cc.NewStream(ctx, &FileRefSync_ServiceDesc.Streams[0], "/FileRefSync/SyncUser", opts...)
	if err != nil {
		return nil, err
	}
	x := &fileRefSyncSyncUserClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type FileRefSync_SyncUserClient interface {
	Recv() (*Update, error)
	grpc.ClientStream
}

type fileRefSyncSyncUserClient struct {
	grpc.ClientStream
}

func (x *fileRefSyncSyncUserClient) Recv() (*Update, error) {
	m := new(Update)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// FileRefSyncServer is the server API for FileRefSync service.
// All implementations must embed UnimplementedFileRefSyncServer
// for forward compatibility
type FileRefSyncServer interface {
	SyncUser(*SyncUserRequest, FileRefSync_SyncUserServer) error
	mustEmbedUnimplementedFileRefSyncServer()
}

// UnimplementedFileRefSyncServer must be embedded to have forward compatible implementations.
type UnimplementedFileRefSyncServer struct {
}

func (UnimplementedFileRefSyncServer) SyncUser(*SyncUserRequest, FileRefSync_SyncUserServer) error {
	return status.Errorf(codes.Unimplemented, "method SyncUser not implemented")
}
func (UnimplementedFileRefSyncServer) mustEmbedUnimplementedFileRefSyncServer() {}

// UnsafeFileRefSyncServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to FileRefSyncServer will
// result in compilation errors.
type UnsafeFileRefSyncServer interface {
	mustEmbedUnimplementedFileRefSyncServer()
}

func RegisterFileRefSyncServer(s grpc.ServiceRegistrar, srv FileRefSyncServer) {
	s.RegisterService(&FileRefSync_ServiceDesc, srv)
}

func _FileRefSync_SyncUser_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(SyncUserRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(FileRefSyncServer).SyncUser(m, &fileRefSyncSyncUserServer{stream})
}

type FileRefSync_SyncUserServer interface {
	Send(*Update) error
	grpc.ServerStream
}

type fileRefSyncSyncUserServer struct {
	grpc.ServerStream
}

func (x *fileRefSyncSyncUserServer) Send(m *Update) error {
	return x.ServerStream.SendMsg(m)
}

// FileRefSync_ServiceDesc is the grpc.ServiceDesc for FileRefSync service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var FileRefSync_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "FileRefSync",
	HandlerType: (*FileRefSyncServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "SyncUser",
			Handler:       _FileRefSync_SyncUser_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "changes.proto",
}
