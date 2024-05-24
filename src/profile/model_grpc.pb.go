// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.0
// source: model.proto

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
	ClientToNameService_PutFile_FullMethodName           = "/ClientToNameService/PutFile"
	ClientToNameService_ConfirmFile_FullMethodName       = "/ClientToNameService/ConfirmFile"
	ClientToNameService_GetFile_FullMethodName           = "/ClientToNameService/GetFile"
	ClientToNameService_GetFileLocation_FullMethodName   = "/ClientToNameService/GetFileLocation"
	ClientToNameService_GetFileStoreChain_FullMethodName = "/ClientToNameService/GetFileStoreChain"
	ClientToNameService_DeleteFile_FullMethodName        = "/ClientToNameService/DeleteFile"
	ClientToNameService_ListDir_FullMethodName           = "/ClientToNameService/ListDir"
	ClientToNameService_ReName_FullMethodName            = "/ClientToNameService/ReName"
)

// ClientToNameServiceClient is the client API for ClientToNameService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientToNameServiceClient interface {
	PutFile(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DataNodeChain, error)
	ConfirmFile(ctx context.Context, in *ConfirmFileArg, opts ...grpc.CallOption) (*ConfirmFileReply, error)
	GetFile(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*FileLocationInfo, error)
	GetFileLocation(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*FileLocationInfo, error)
	GetFileStoreChain(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DataNodeChain, error)
	DeleteFile(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DeleteFileReply, error)
	ListDir(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DirMetaList, error)
	ReName(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*ReNameReply, error)
}

type clientToNameServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientToNameServiceClient(cc grpc.ClientConnInterface) ClientToNameServiceClient {
	return &clientToNameServiceClient{cc}
}

func (c *clientToNameServiceClient) PutFile(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DataNodeChain, error) {
	out := new(DataNodeChain)
	err := c.cc.Invoke(ctx, ClientToNameService_PutFile_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientToNameServiceClient) ConfirmFile(ctx context.Context, in *ConfirmFileArg, opts ...grpc.CallOption) (*ConfirmFileReply, error) {
	out := new(ConfirmFileReply)
	err := c.cc.Invoke(ctx, ClientToNameService_ConfirmFile_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientToNameServiceClient) GetFile(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*FileLocationInfo, error) {
	out := new(FileLocationInfo)
	err := c.cc.Invoke(ctx, ClientToNameService_GetFile_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientToNameServiceClient) GetFileLocation(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*FileLocationInfo, error) {
	out := new(FileLocationInfo)
	err := c.cc.Invoke(ctx, ClientToNameService_GetFileLocation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientToNameServiceClient) GetFileStoreChain(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DataNodeChain, error) {
	out := new(DataNodeChain)
	err := c.cc.Invoke(ctx, ClientToNameService_GetFileStoreChain_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientToNameServiceClient) DeleteFile(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DeleteFileReply, error) {
	out := new(DeleteFileReply)
	err := c.cc.Invoke(ctx, ClientToNameService_DeleteFile_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientToNameServiceClient) ListDir(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*DirMetaList, error) {
	out := new(DirMetaList)
	err := c.cc.Invoke(ctx, ClientToNameService_ListDir_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientToNameServiceClient) ReName(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*ReNameReply, error) {
	out := new(ReNameReply)
	err := c.cc.Invoke(ctx, ClientToNameService_ReName_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientToNameServiceServer is the server API for ClientToNameService service.
// All implementations must embed UnimplementedClientToNameServiceServer
// for forward compatibility
type ClientToNameServiceServer interface {
	PutFile(context.Context, *FileOperationArg) (*DataNodeChain, error)
	ConfirmFile(context.Context, *ConfirmFileArg) (*ConfirmFileReply, error)
	GetFile(context.Context, *FileOperationArg) (*FileLocationInfo, error)
	GetFileLocation(context.Context, *FileOperationArg) (*FileLocationInfo, error)
	GetFileStoreChain(context.Context, *FileOperationArg) (*DataNodeChain, error)
	DeleteFile(context.Context, *FileOperationArg) (*DeleteFileReply, error)
	ListDir(context.Context, *FileOperationArg) (*DirMetaList, error)
	ReName(context.Context, *FileOperationArg) (*ReNameReply, error)
	mustEmbedUnimplementedClientToNameServiceServer()
}

// UnimplementedClientToNameServiceServer must be embedded to have forward compatible implementations.
type UnimplementedClientToNameServiceServer struct {
}

func (UnimplementedClientToNameServiceServer) PutFile(context.Context, *FileOperationArg) (*DataNodeChain, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PutFile not implemented")
}
func (UnimplementedClientToNameServiceServer) ConfirmFile(context.Context, *ConfirmFileArg) (*ConfirmFileReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmFile not implemented")
}
func (UnimplementedClientToNameServiceServer) GetFile(context.Context, *FileOperationArg) (*FileLocationInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFile not implemented")
}
func (UnimplementedClientToNameServiceServer) GetFileLocation(context.Context, *FileOperationArg) (*FileLocationInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFileLocation not implemented")
}
func (UnimplementedClientToNameServiceServer) GetFileStoreChain(context.Context, *FileOperationArg) (*DataNodeChain, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFileStoreChain not implemented")
}
func (UnimplementedClientToNameServiceServer) DeleteFile(context.Context, *FileOperationArg) (*DeleteFileReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteFile not implemented")
}
func (UnimplementedClientToNameServiceServer) ListDir(context.Context, *FileOperationArg) (*DirMetaList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDir not implemented")
}
func (UnimplementedClientToNameServiceServer) ReName(context.Context, *FileOperationArg) (*ReNameReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReName not implemented")
}
func (UnimplementedClientToNameServiceServer) mustEmbedUnimplementedClientToNameServiceServer() {}

// UnsafeClientToNameServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientToNameServiceServer will
// result in compilation errors.
type UnsafeClientToNameServiceServer interface {
	mustEmbedUnimplementedClientToNameServiceServer()
}

func RegisterClientToNameServiceServer(s grpc.ServiceRegistrar, srv ClientToNameServiceServer) {
	s.RegisterService(&ClientToNameService_ServiceDesc, srv)
}

func _ClientToNameService_PutFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).PutFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_PutFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).PutFile(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientToNameService_ConfirmFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ConfirmFileArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).ConfirmFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_ConfirmFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).ConfirmFile(ctx, req.(*ConfirmFileArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientToNameService_GetFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).GetFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_GetFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).GetFile(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientToNameService_GetFileLocation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).GetFileLocation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_GetFileLocation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).GetFileLocation(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientToNameService_GetFileStoreChain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).GetFileStoreChain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_GetFileStoreChain_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).GetFileStoreChain(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientToNameService_DeleteFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).DeleteFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_DeleteFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).DeleteFile(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientToNameService_ListDir_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).ListDir(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_ListDir_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).ListDir(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientToNameService_ReName_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientToNameServiceServer).ReName(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientToNameService_ReName_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientToNameServiceServer).ReName(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

// ClientToNameService_ServiceDesc is the grpc.ServiceDesc for ClientToNameService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientToNameService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ClientToNameService",
	HandlerType: (*ClientToNameServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PutFile",
			Handler:    _ClientToNameService_PutFile_Handler,
		},
		{
			MethodName: "ConfirmFile",
			Handler:    _ClientToNameService_ConfirmFile_Handler,
		},
		{
			MethodName: "GetFile",
			Handler:    _ClientToNameService_GetFile_Handler,
		},
		{
			MethodName: "GetFileLocation",
			Handler:    _ClientToNameService_GetFileLocation_Handler,
		},
		{
			MethodName: "GetFileStoreChain",
			Handler:    _ClientToNameService_GetFileStoreChain_Handler,
		},
		{
			MethodName: "DeleteFile",
			Handler:    _ClientToNameService_DeleteFile_Handler,
		},
		{
			MethodName: "ListDir",
			Handler:    _ClientToNameService_ListDir_Handler,
		},
		{
			MethodName: "ReName",
			Handler:    _ClientToNameService_ReName_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "model.proto",
}

const (
	ClientToDataService_GetChunk_FullMethodName = "/ClientToDataService/GetChunk"
	ClientToDataService_PutChunk_FullMethodName = "/ClientToDataService/PutChunk"
)

// ClientToDataServiceClient is the client API for ClientToDataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ClientToDataServiceClient interface {
	GetChunk(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (ClientToDataService_GetChunkClient, error)
	PutChunk(ctx context.Context, opts ...grpc.CallOption) (ClientToDataService_PutChunkClient, error)
}

type clientToDataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientToDataServiceClient(cc grpc.ClientConnInterface) ClientToDataServiceClient {
	return &clientToDataServiceClient{cc}
}

func (c *clientToDataServiceClient) GetChunk(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (ClientToDataService_GetChunkClient, error) {
	stream, err := c.cc.NewStream(ctx, &ClientToDataService_ServiceDesc.Streams[0], ClientToDataService_GetChunk_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &clientToDataServiceGetChunkClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ClientToDataService_GetChunkClient interface {
	Recv() (*FileDataStream, error)
	grpc.ClientStream
}

type clientToDataServiceGetChunkClient struct {
	grpc.ClientStream
}

func (x *clientToDataServiceGetChunkClient) Recv() (*FileDataStream, error) {
	m := new(FileDataStream)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *clientToDataServiceClient) PutChunk(ctx context.Context, opts ...grpc.CallOption) (ClientToDataService_PutChunkClient, error) {
	stream, err := c.cc.NewStream(ctx, &ClientToDataService_ServiceDesc.Streams[1], ClientToDataService_PutChunk_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &clientToDataServicePutChunkClient{stream}
	return x, nil
}

type ClientToDataService_PutChunkClient interface {
	Send(*FileDataStream) error
	CloseAndRecv() (*FileLocationInfo, error)
	grpc.ClientStream
}

type clientToDataServicePutChunkClient struct {
	grpc.ClientStream
}

func (x *clientToDataServicePutChunkClient) Send(m *FileDataStream) error {
	return x.ClientStream.SendMsg(m)
}

func (x *clientToDataServicePutChunkClient) CloseAndRecv() (*FileLocationInfo, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(FileLocationInfo)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ClientToDataServiceServer is the server API for ClientToDataService service.
// All implementations must embed UnimplementedClientToDataServiceServer
// for forward compatibility
type ClientToDataServiceServer interface {
	GetChunk(*FileOperationArg, ClientToDataService_GetChunkServer) error
	PutChunk(ClientToDataService_PutChunkServer) error
	mustEmbedUnimplementedClientToDataServiceServer()
}

// UnimplementedClientToDataServiceServer must be embedded to have forward compatible implementations.
type UnimplementedClientToDataServiceServer struct {
}

func (UnimplementedClientToDataServiceServer) GetChunk(*FileOperationArg, ClientToDataService_GetChunkServer) error {
	return status.Errorf(codes.Unimplemented, "method GetChunk not implemented")
}
func (UnimplementedClientToDataServiceServer) PutChunk(ClientToDataService_PutChunkServer) error {
	return status.Errorf(codes.Unimplemented, "method PutChunk not implemented")
}
func (UnimplementedClientToDataServiceServer) mustEmbedUnimplementedClientToDataServiceServer() {}

// UnsafeClientToDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientToDataServiceServer will
// result in compilation errors.
type UnsafeClientToDataServiceServer interface {
	mustEmbedUnimplementedClientToDataServiceServer()
}

func RegisterClientToDataServiceServer(s grpc.ServiceRegistrar, srv ClientToDataServiceServer) {
	s.RegisterService(&ClientToDataService_ServiceDesc, srv)
}

func _ClientToDataService_GetChunk_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(FileOperationArg)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ClientToDataServiceServer).GetChunk(m, &clientToDataServiceGetChunkServer{stream})
}

type ClientToDataService_GetChunkServer interface {
	Send(*FileDataStream) error
	grpc.ServerStream
}

type clientToDataServiceGetChunkServer struct {
	grpc.ServerStream
}

func (x *clientToDataServiceGetChunkServer) Send(m *FileDataStream) error {
	return x.ServerStream.SendMsg(m)
}

func _ClientToDataService_PutChunk_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ClientToDataServiceServer).PutChunk(&clientToDataServicePutChunkServer{stream})
}

type ClientToDataService_PutChunkServer interface {
	SendAndClose(*FileLocationInfo) error
	Recv() (*FileDataStream, error)
	grpc.ServerStream
}

type clientToDataServicePutChunkServer struct {
	grpc.ServerStream
}

func (x *clientToDataServicePutChunkServer) SendAndClose(m *FileLocationInfo) error {
	return x.ServerStream.SendMsg(m)
}

func (x *clientToDataServicePutChunkServer) Recv() (*FileDataStream, error) {
	m := new(FileDataStream)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ClientToDataService_ServiceDesc is the grpc.ServiceDesc for ClientToDataService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientToDataService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ClientToDataService",
	HandlerType: (*ClientToDataServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetChunk",
			Handler:       _ClientToDataService_GetChunk_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "PutChunk",
			Handler:       _ClientToDataService_PutChunk_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "model.proto",
}

const (
	NameToDataService_GetDataNodeInfo_FullMethodName = "/NameToDataService/GetDataNodeInfo"
)

// NameToDataServiceClient is the client API for NameToDataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NameToDataServiceClient interface {
	GetDataNodeInfo(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*FileLocationInfo, error)
}

type nameToDataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewNameToDataServiceClient(cc grpc.ClientConnInterface) NameToDataServiceClient {
	return &nameToDataServiceClient{cc}
}

func (c *nameToDataServiceClient) GetDataNodeInfo(ctx context.Context, in *FileOperationArg, opts ...grpc.CallOption) (*FileLocationInfo, error) {
	out := new(FileLocationInfo)
	err := c.cc.Invoke(ctx, NameToDataService_GetDataNodeInfo_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NameToDataServiceServer is the server API for NameToDataService service.
// All implementations must embed UnimplementedNameToDataServiceServer
// for forward compatibility
type NameToDataServiceServer interface {
	GetDataNodeInfo(context.Context, *FileOperationArg) (*FileLocationInfo, error)
	mustEmbedUnimplementedNameToDataServiceServer()
}

// UnimplementedNameToDataServiceServer must be embedded to have forward compatible implementations.
type UnimplementedNameToDataServiceServer struct {
}

func (UnimplementedNameToDataServiceServer) GetDataNodeInfo(context.Context, *FileOperationArg) (*FileLocationInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDataNodeInfo not implemented")
}
func (UnimplementedNameToDataServiceServer) mustEmbedUnimplementedNameToDataServiceServer() {}

// UnsafeNameToDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NameToDataServiceServer will
// result in compilation errors.
type UnsafeNameToDataServiceServer interface {
	mustEmbedUnimplementedNameToDataServiceServer()
}

func RegisterNameToDataServiceServer(s grpc.ServiceRegistrar, srv NameToDataServiceServer) {
	s.RegisterService(&NameToDataService_ServiceDesc, srv)
}

func _NameToDataService_GetDataNodeInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileOperationArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NameToDataServiceServer).GetDataNodeInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NameToDataService_GetDataNodeInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NameToDataServiceServer).GetDataNodeInfo(ctx, req.(*FileOperationArg))
	}
	return interceptor(ctx, in, info, handler)
}

// NameToDataService_ServiceDesc is the grpc.ServiceDesc for NameToDataService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NameToDataService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "NameToDataService",
	HandlerType: (*NameToDataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDataNodeInfo",
			Handler:    _NameToDataService_GetDataNodeInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "model.proto",
}

const (
	DataToNameService_RegisterDataNode_FullMethodName = "/DataToNameService/RegisterDataNode"
	DataToNameService_HeartBeat_FullMethodName        = "/DataToNameService/HeartBeat"
	DataToNameService_ChunkReport_FullMethodName      = "/DataToNameService/ChunkReport"
	DataToNameService_CommitChunk_FullMethodName      = "/DataToNameService/CommitChunk"
)

// DataToNameServiceClient is the client API for DataToNameService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataToNameServiceClient interface {
	RegisterDataNode(ctx context.Context, in *DataNodeRegisterArg, opts ...grpc.CallOption) (*DataNodeRegisterReply, error)
	HeartBeat(ctx context.Context, in *HeartBeatArg, opts ...grpc.CallOption) (*HeartBeatReply, error)
	ChunkReport(ctx context.Context, in *FileLocationInfo, opts ...grpc.CallOption) (*ChunkReportReply, error)
	CommitChunk(ctx context.Context, in *CommitChunkArg, opts ...grpc.CallOption) (*CommitChunkReply, error)
}

type dataToNameServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDataToNameServiceClient(cc grpc.ClientConnInterface) DataToNameServiceClient {
	return &dataToNameServiceClient{cc}
}

func (c *dataToNameServiceClient) RegisterDataNode(ctx context.Context, in *DataNodeRegisterArg, opts ...grpc.CallOption) (*DataNodeRegisterReply, error) {
	out := new(DataNodeRegisterReply)
	err := c.cc.Invoke(ctx, DataToNameService_RegisterDataNode_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataToNameServiceClient) HeartBeat(ctx context.Context, in *HeartBeatArg, opts ...grpc.CallOption) (*HeartBeatReply, error) {
	out := new(HeartBeatReply)
	err := c.cc.Invoke(ctx, DataToNameService_HeartBeat_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataToNameServiceClient) ChunkReport(ctx context.Context, in *FileLocationInfo, opts ...grpc.CallOption) (*ChunkReportReply, error) {
	out := new(ChunkReportReply)
	err := c.cc.Invoke(ctx, DataToNameService_ChunkReport_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataToNameServiceClient) CommitChunk(ctx context.Context, in *CommitChunkArg, opts ...grpc.CallOption) (*CommitChunkReply, error) {
	out := new(CommitChunkReply)
	err := c.cc.Invoke(ctx, DataToNameService_CommitChunk_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataToNameServiceServer is the server API for DataToNameService service.
// All implementations must embed UnimplementedDataToNameServiceServer
// for forward compatibility
type DataToNameServiceServer interface {
	RegisterDataNode(context.Context, *DataNodeRegisterArg) (*DataNodeRegisterReply, error)
	HeartBeat(context.Context, *HeartBeatArg) (*HeartBeatReply, error)
	ChunkReport(context.Context, *FileLocationInfo) (*ChunkReportReply, error)
	CommitChunk(context.Context, *CommitChunkArg) (*CommitChunkReply, error)
	mustEmbedUnimplementedDataToNameServiceServer()
}

// UnimplementedDataToNameServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDataToNameServiceServer struct {
}

func (UnimplementedDataToNameServiceServer) RegisterDataNode(context.Context, *DataNodeRegisterArg) (*DataNodeRegisterReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterDataNode not implemented")
}
func (UnimplementedDataToNameServiceServer) HeartBeat(context.Context, *HeartBeatArg) (*HeartBeatReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HeartBeat not implemented")
}
func (UnimplementedDataToNameServiceServer) ChunkReport(context.Context, *FileLocationInfo) (*ChunkReportReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ChunkReport not implemented")
}
func (UnimplementedDataToNameServiceServer) CommitChunk(context.Context, *CommitChunkArg) (*CommitChunkReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitChunk not implemented")
}
func (UnimplementedDataToNameServiceServer) mustEmbedUnimplementedDataToNameServiceServer() {}

// UnsafeDataToNameServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataToNameServiceServer will
// result in compilation errors.
type UnsafeDataToNameServiceServer interface {
	mustEmbedUnimplementedDataToNameServiceServer()
}

func RegisterDataToNameServiceServer(s grpc.ServiceRegistrar, srv DataToNameServiceServer) {
	s.RegisterService(&DataToNameService_ServiceDesc, srv)
}

func _DataToNameService_RegisterDataNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DataNodeRegisterArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataToNameServiceServer).RegisterDataNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataToNameService_RegisterDataNode_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataToNameServiceServer).RegisterDataNode(ctx, req.(*DataNodeRegisterArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataToNameService_HeartBeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HeartBeatArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataToNameServiceServer).HeartBeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataToNameService_HeartBeat_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataToNameServiceServer).HeartBeat(ctx, req.(*HeartBeatArg))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataToNameService_ChunkReport_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileLocationInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataToNameServiceServer).ChunkReport(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataToNameService_ChunkReport_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataToNameServiceServer).ChunkReport(ctx, req.(*FileLocationInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataToNameService_CommitChunk_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CommitChunkArg)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataToNameServiceServer).CommitChunk(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataToNameService_CommitChunk_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataToNameServiceServer).CommitChunk(ctx, req.(*CommitChunkArg))
	}
	return interceptor(ctx, in, info, handler)
}

// DataToNameService_ServiceDesc is the grpc.ServiceDesc for DataToNameService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataToNameService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DataToNameService",
	HandlerType: (*DataToNameServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterDataNode",
			Handler:    _DataToNameService_RegisterDataNode_Handler,
		},
		{
			MethodName: "HeartBeat",
			Handler:    _DataToNameService_HeartBeat_Handler,
		},
		{
			MethodName: "ChunkReport",
			Handler:    _DataToNameService_ChunkReport_Handler,
		},
		{
			MethodName: "CommitChunk",
			Handler:    _DataToNameService_CommitChunk_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "model.proto",
}

const (
	DataToDataService_PutChunk_FullMethodName = "/DataToDataService/PutChunk"
)

// DataToDataServiceClient is the client API for DataToDataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataToDataServiceClient interface {
	PutChunk(ctx context.Context, opts ...grpc.CallOption) (DataToDataService_PutChunkClient, error)
}

type dataToDataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDataToDataServiceClient(cc grpc.ClientConnInterface) DataToDataServiceClient {
	return &dataToDataServiceClient{cc}
}

func (c *dataToDataServiceClient) PutChunk(ctx context.Context, opts ...grpc.CallOption) (DataToDataService_PutChunkClient, error) {
	stream, err := c.cc.NewStream(ctx, &DataToDataService_ServiceDesc.Streams[0], DataToDataService_PutChunk_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &dataToDataServicePutChunkClient{stream}
	return x, nil
}

type DataToDataService_PutChunkClient interface {
	Send(*FileDataStream) error
	CloseAndRecv() (*FileLocationInfo, error)
	grpc.ClientStream
}

type dataToDataServicePutChunkClient struct {
	grpc.ClientStream
}

func (x *dataToDataServicePutChunkClient) Send(m *FileDataStream) error {
	return x.ClientStream.SendMsg(m)
}

func (x *dataToDataServicePutChunkClient) CloseAndRecv() (*FileLocationInfo, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(FileLocationInfo)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DataToDataServiceServer is the server API for DataToDataService service.
// All implementations must embed UnimplementedDataToDataServiceServer
// for forward compatibility
type DataToDataServiceServer interface {
	PutChunk(DataToDataService_PutChunkServer) error
	mustEmbedUnimplementedDataToDataServiceServer()
}

// UnimplementedDataToDataServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDataToDataServiceServer struct {
}

func (UnimplementedDataToDataServiceServer) PutChunk(DataToDataService_PutChunkServer) error {
	return status.Errorf(codes.Unimplemented, "method PutChunk not implemented")
}
func (UnimplementedDataToDataServiceServer) mustEmbedUnimplementedDataToDataServiceServer() {}

// UnsafeDataToDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataToDataServiceServer will
// result in compilation errors.
type UnsafeDataToDataServiceServer interface {
	mustEmbedUnimplementedDataToDataServiceServer()
}

func RegisterDataToDataServiceServer(s grpc.ServiceRegistrar, srv DataToDataServiceServer) {
	s.RegisterService(&DataToDataService_ServiceDesc, srv)
}

func _DataToDataService_PutChunk_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DataToDataServiceServer).PutChunk(&dataToDataServicePutChunkServer{stream})
}

type DataToDataService_PutChunkServer interface {
	SendAndClose(*FileLocationInfo) error
	Recv() (*FileDataStream, error)
	grpc.ServerStream
}

type dataToDataServicePutChunkServer struct {
	grpc.ServerStream
}

func (x *dataToDataServicePutChunkServer) SendAndClose(m *FileLocationInfo) error {
	return x.ServerStream.SendMsg(m)
}

func (x *dataToDataServicePutChunkServer) Recv() (*FileDataStream, error) {
	m := new(FileDataStream)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// DataToDataService_ServiceDesc is the grpc.ServiceDesc for DataToDataService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataToDataService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DataToDataService",
	HandlerType: (*DataToDataServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "PutChunk",
			Handler:       _DataToDataService_PutChunk_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "model.proto",
}
