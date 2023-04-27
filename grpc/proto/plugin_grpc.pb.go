// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: plugin.proto

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

// WrapperPluginClient is the client API for WrapperPlugin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WrapperPluginClient interface {
	EstablishMessageStream(ctx context.Context, in *EstablishMessageStreamRequest, opts ...grpc.CallOption) (WrapperPlugin_EstablishMessageStreamClient, error)
	GetSchema(ctx context.Context, in *GetSchemaRequest, opts ...grpc.CallOption) (*GetSchemaResponse, error)
	Execute(ctx context.Context, in *ExecuteRequest, opts ...grpc.CallOption) (WrapperPlugin_ExecuteClient, error)
	SetConnectionConfig(ctx context.Context, in *SetConnectionConfigRequest, opts ...grpc.CallOption) (*SetConnectionConfigResponse, error)
	SetAllConnectionConfigs(ctx context.Context, in *SetAllConnectionConfigsRequest, opts ...grpc.CallOption) (*SetConnectionConfigResponse, error)
	UpdateConnectionConfigs(ctx context.Context, in *UpdateConnectionConfigsRequest, opts ...grpc.CallOption) (*UpdateConnectionConfigsResponse, error)
	GetSupportedOperations(ctx context.Context, in *GetSupportedOperationsRequest, opts ...grpc.CallOption) (*GetSupportedOperationsResponse, error)
	SetCacheOptions(ctx context.Context, in *SetCacheOptionsRequest, opts ...grpc.CallOption) (*SetCacheOptionsResponse, error)
}

type wrapperPluginClient struct {
	cc grpc.ClientConnInterface
}

func NewWrapperPluginClient(cc grpc.ClientConnInterface) WrapperPluginClient {
	return &wrapperPluginClient{cc}
}

func (c *wrapperPluginClient) EstablishMessageStream(ctx context.Context, in *EstablishMessageStreamRequest, opts ...grpc.CallOption) (WrapperPlugin_EstablishMessageStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &WrapperPlugin_ServiceDesc.Streams[0], "/proto.WrapperPlugin/EstablishMessageStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &wrapperPluginEstablishMessageStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WrapperPlugin_EstablishMessageStreamClient interface {
	Recv() (*PluginMessage, error)
	grpc.ClientStream
}

type wrapperPluginEstablishMessageStreamClient struct {
	grpc.ClientStream
}

func (x *wrapperPluginEstablishMessageStreamClient) Recv() (*PluginMessage, error) {
	m := new(PluginMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *wrapperPluginClient) GetSchema(ctx context.Context, in *GetSchemaRequest, opts ...grpc.CallOption) (*GetSchemaResponse, error) {
	out := new(GetSchemaResponse)
	err := c.cc.Invoke(ctx, "/proto.WrapperPlugin/GetSchema", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wrapperPluginClient) Execute(ctx context.Context, in *ExecuteRequest, opts ...grpc.CallOption) (WrapperPlugin_ExecuteClient, error) {
	stream, err := c.cc.NewStream(ctx, &WrapperPlugin_ServiceDesc.Streams[1], "/proto.WrapperPlugin/Execute", opts...)
	if err != nil {
		return nil, err
	}
	x := &wrapperPluginExecuteClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WrapperPlugin_ExecuteClient interface {
	Recv() (*ExecuteResponse, error)
	grpc.ClientStream
}

type wrapperPluginExecuteClient struct {
	grpc.ClientStream
}

func (x *wrapperPluginExecuteClient) Recv() (*ExecuteResponse, error) {
	m := new(ExecuteResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *wrapperPluginClient) SetConnectionConfig(ctx context.Context, in *SetConnectionConfigRequest, opts ...grpc.CallOption) (*SetConnectionConfigResponse, error) {
	out := new(SetConnectionConfigResponse)
	err := c.cc.Invoke(ctx, "/proto.WrapperPlugin/SetConnectionConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wrapperPluginClient) SetAllConnectionConfigs(ctx context.Context, in *SetAllConnectionConfigsRequest, opts ...grpc.CallOption) (*SetConnectionConfigResponse, error) {
	out := new(SetConnectionConfigResponse)
	err := c.cc.Invoke(ctx, "/proto.WrapperPlugin/SetAllConnectionConfigs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wrapperPluginClient) UpdateConnectionConfigs(ctx context.Context, in *UpdateConnectionConfigsRequest, opts ...grpc.CallOption) (*UpdateConnectionConfigsResponse, error) {
	out := new(UpdateConnectionConfigsResponse)
	err := c.cc.Invoke(ctx, "/proto.WrapperPlugin/UpdateConnectionConfigs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wrapperPluginClient) GetSupportedOperations(ctx context.Context, in *GetSupportedOperationsRequest, opts ...grpc.CallOption) (*GetSupportedOperationsResponse, error) {
	out := new(GetSupportedOperationsResponse)
	err := c.cc.Invoke(ctx, "/proto.WrapperPlugin/GetSupportedOperations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *wrapperPluginClient) SetCacheOptions(ctx context.Context, in *SetCacheOptionsRequest, opts ...grpc.CallOption) (*SetCacheOptionsResponse, error) {
	out := new(SetCacheOptionsResponse)
	err := c.cc.Invoke(ctx, "/proto.WrapperPlugin/SetCacheOptions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WrapperPluginServer is the server API for WrapperPlugin service.
// All implementations must embed UnimplementedWrapperPluginServer
// for forward compatibility
type WrapperPluginServer interface {
	EstablishMessageStream(*EstablishMessageStreamRequest, WrapperPlugin_EstablishMessageStreamServer) error
	GetSchema(context.Context, *GetSchemaRequest) (*GetSchemaResponse, error)
	Execute(*ExecuteRequest, WrapperPlugin_ExecuteServer) error
	SetConnectionConfig(context.Context, *SetConnectionConfigRequest) (*SetConnectionConfigResponse, error)
	SetAllConnectionConfigs(context.Context, *SetAllConnectionConfigsRequest) (*SetConnectionConfigResponse, error)
	UpdateConnectionConfigs(context.Context, *UpdateConnectionConfigsRequest) (*UpdateConnectionConfigsResponse, error)
	GetSupportedOperations(context.Context, *GetSupportedOperationsRequest) (*GetSupportedOperationsResponse, error)
	SetCacheOptions(context.Context, *SetCacheOptionsRequest) (*SetCacheOptionsResponse, error)
	mustEmbedUnimplementedWrapperPluginServer()
}

// UnimplementedWrapperPluginServer must be embedded to have forward compatible implementations.
type UnimplementedWrapperPluginServer struct {
}

func (UnimplementedWrapperPluginServer) EstablishMessageStream(*EstablishMessageStreamRequest, WrapperPlugin_EstablishMessageStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method EstablishMessageStream not implemented")
}
func (UnimplementedWrapperPluginServer) GetSchema(context.Context, *GetSchemaRequest) (*GetSchemaResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSchema not implemented")
}
func (UnimplementedWrapperPluginServer) Execute(*ExecuteRequest, WrapperPlugin_ExecuteServer) error {
	return status.Errorf(codes.Unimplemented, "method Execute not implemented")
}
func (UnimplementedWrapperPluginServer) SetConnectionConfig(context.Context, *SetConnectionConfigRequest) (*SetConnectionConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetConnectionConfig not implemented")
}
func (UnimplementedWrapperPluginServer) SetAllConnectionConfigs(context.Context, *SetAllConnectionConfigsRequest) (*SetConnectionConfigResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetAllConnectionConfigs not implemented")
}
func (UnimplementedWrapperPluginServer) UpdateConnectionConfigs(context.Context, *UpdateConnectionConfigsRequest) (*UpdateConnectionConfigsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateConnectionConfigs not implemented")
}
func (UnimplementedWrapperPluginServer) GetSupportedOperations(context.Context, *GetSupportedOperationsRequest) (*GetSupportedOperationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSupportedOperations not implemented")
}
func (UnimplementedWrapperPluginServer) SetCacheOptions(context.Context, *SetCacheOptionsRequest) (*SetCacheOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetCacheOptions not implemented")
}
func (UnimplementedWrapperPluginServer) mustEmbedUnimplementedWrapperPluginServer() {}

// UnsafeWrapperPluginServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WrapperPluginServer will
// result in compilation errors.
type UnsafeWrapperPluginServer interface {
	mustEmbedUnimplementedWrapperPluginServer()
}

func RegisterWrapperPluginServer(s grpc.ServiceRegistrar, srv WrapperPluginServer) {
	s.RegisterService(&WrapperPlugin_ServiceDesc, srv)
}

func _WrapperPlugin_EstablishMessageStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(EstablishMessageStreamRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WrapperPluginServer).EstablishMessageStream(m, &wrapperPluginEstablishMessageStreamServer{stream})
}

type WrapperPlugin_EstablishMessageStreamServer interface {
	Send(*PluginMessage) error
	grpc.ServerStream
}

type wrapperPluginEstablishMessageStreamServer struct {
	grpc.ServerStream
}

func (x *wrapperPluginEstablishMessageStreamServer) Send(m *PluginMessage) error {
	return x.ServerStream.SendMsg(m)
}

func _WrapperPlugin_GetSchema_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSchemaRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WrapperPluginServer).GetSchema(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WrapperPlugin/GetSchema",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WrapperPluginServer).GetSchema(ctx, req.(*GetSchemaRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WrapperPlugin_Execute_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ExecuteRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WrapperPluginServer).Execute(m, &wrapperPluginExecuteServer{stream})
}

type WrapperPlugin_ExecuteServer interface {
	Send(*ExecuteResponse) error
	grpc.ServerStream
}

type wrapperPluginExecuteServer struct {
	grpc.ServerStream
}

func (x *wrapperPluginExecuteServer) Send(m *ExecuteResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _WrapperPlugin_SetConnectionConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetConnectionConfigRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WrapperPluginServer).SetConnectionConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WrapperPlugin/SetConnectionConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WrapperPluginServer).SetConnectionConfig(ctx, req.(*SetConnectionConfigRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WrapperPlugin_SetAllConnectionConfigs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetAllConnectionConfigsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WrapperPluginServer).SetAllConnectionConfigs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WrapperPlugin/SetAllConnectionConfigs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WrapperPluginServer).SetAllConnectionConfigs(ctx, req.(*SetAllConnectionConfigsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WrapperPlugin_UpdateConnectionConfigs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateConnectionConfigsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WrapperPluginServer).UpdateConnectionConfigs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WrapperPlugin/UpdateConnectionConfigs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WrapperPluginServer).UpdateConnectionConfigs(ctx, req.(*UpdateConnectionConfigsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WrapperPlugin_GetSupportedOperations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetSupportedOperationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WrapperPluginServer).GetSupportedOperations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WrapperPlugin/GetSupportedOperations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WrapperPluginServer).GetSupportedOperations(ctx, req.(*GetSupportedOperationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WrapperPlugin_SetCacheOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetCacheOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WrapperPluginServer).SetCacheOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.WrapperPlugin/SetCacheOptions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WrapperPluginServer).SetCacheOptions(ctx, req.(*SetCacheOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WrapperPlugin_ServiceDesc is the grpc.ServiceDesc for WrapperPlugin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WrapperPlugin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.WrapperPlugin",
	HandlerType: (*WrapperPluginServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSchema",
			Handler:    _WrapperPlugin_GetSchema_Handler,
		},
		{
			MethodName: "SetConnectionConfig",
			Handler:    _WrapperPlugin_SetConnectionConfig_Handler,
		},
		{
			MethodName: "SetAllConnectionConfigs",
			Handler:    _WrapperPlugin_SetAllConnectionConfigs_Handler,
		},
		{
			MethodName: "UpdateConnectionConfigs",
			Handler:    _WrapperPlugin_UpdateConnectionConfigs_Handler,
		},
		{
			MethodName: "GetSupportedOperations",
			Handler:    _WrapperPlugin_GetSupportedOperations_Handler,
		},
		{
			MethodName: "SetCacheOptions",
			Handler:    _WrapperPlugin_SetCacheOptions_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EstablishMessageStream",
			Handler:       _WrapperPlugin_EstablishMessageStream_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Execute",
			Handler:       _WrapperPlugin_Execute_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "plugin.proto",
}
