// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: proto/service.proto

package proto

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

const (
	CreatePlaylistService_CreatePlaylist_FullMethodName = "/CreatePlaylistService/CreatePlaylist"
)

// CreatePlaylistServiceClient is the client API for CreatePlaylistService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CreatePlaylistServiceClient interface {
	CreatePlaylist(ctx context.Context, in *CreatePlaylistRequest, opts ...grpc.CallOption) (*CreatePlaylistResponse, error)
}

type createPlaylistServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCreatePlaylistServiceClient(cc grpc.ClientConnInterface) CreatePlaylistServiceClient {
	return &createPlaylistServiceClient{cc}
}

func (c *createPlaylistServiceClient) CreatePlaylist(ctx context.Context, in *CreatePlaylistRequest, opts ...grpc.CallOption) (*CreatePlaylistResponse, error) {
	out := new(CreatePlaylistResponse)
	err := c.cc.Invoke(ctx, CreatePlaylistService_CreatePlaylist_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CreatePlaylistServiceServer is the server API for CreatePlaylistService service.
// All implementations must embed UnimplementedCreatePlaylistServiceServer
// for forward compatibility
type CreatePlaylistServiceServer interface {
	CreatePlaylist(context.Context, *CreatePlaylistRequest) (*CreatePlaylistResponse, error)
	mustEmbedUnimplementedCreatePlaylistServiceServer()
}

// UnimplementedCreatePlaylistServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCreatePlaylistServiceServer struct {
}

func (UnimplementedCreatePlaylistServiceServer) CreatePlaylist(context.Context, *CreatePlaylistRequest) (*CreatePlaylistResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreatePlaylist not implemented")
}
func (UnimplementedCreatePlaylistServiceServer) mustEmbedUnimplementedCreatePlaylistServiceServer() {}

// UnsafeCreatePlaylistServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CreatePlaylistServiceServer will
// result in compilation errors.
type UnsafeCreatePlaylistServiceServer interface {
	mustEmbedUnimplementedCreatePlaylistServiceServer()
}

func RegisterCreatePlaylistServiceServer(s grpc.ServiceRegistrar, srv CreatePlaylistServiceServer) {
	s.RegisterService(&CreatePlaylistService_ServiceDesc, srv)
}

func _CreatePlaylistService_CreatePlaylist_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePlaylistRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CreatePlaylistServiceServer).CreatePlaylist(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CreatePlaylistService_CreatePlaylist_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CreatePlaylistServiceServer).CreatePlaylist(ctx, req.(*CreatePlaylistRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CreatePlaylistService_ServiceDesc is the grpc.ServiceDesc for CreatePlaylistService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CreatePlaylistService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "CreatePlaylistService",
	HandlerType: (*CreatePlaylistServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreatePlaylist",
			Handler:    _CreatePlaylistService_CreatePlaylist_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/service.proto",
}