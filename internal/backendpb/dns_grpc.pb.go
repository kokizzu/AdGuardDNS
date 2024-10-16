// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.27.1
// source: dns.proto

package backendpb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	DNSService_GetDNSProfiles_FullMethodName         = "/DNSService/getDNSProfiles"
	DNSService_SaveDevicesBillingStat_FullMethodName = "/DNSService/saveDevicesBillingStat"
	DNSService_CreateDeviceByHumanId_FullMethodName  = "/DNSService/createDeviceByHumanId"
)

// DNSServiceClient is the client API for DNSService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DNSServiceClient interface {
	// Gets DNS profiles.
	//
	// Field "sync_time" in DNSProfilesRequest - pass to return the latest updates after this time moment.
	//
	// The trailers headers will include a "sync_time", given in milliseconds,
	// that should be used for subsequent incremental DNS profile synchronization requests.
	//
	// This method may return the following errors:
	// - RateLimitedError: If too many "full sync" concurrent requests are made.
	// - AuthenticationFailedError: If the authentication failed.
	GetDNSProfiles(ctx context.Context, in *DNSProfilesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[DNSProfile], error)
	// Stores devices activity.
	//
	// This method may return the following errors:
	// - AuthenticationFailedError: If the authentication failed.
	SaveDevicesBillingStat(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[DeviceBillingStat, emptypb.Empty], error)
	// Create device by "human_id".
	//
	// This method may return the following errors:
	// - RateLimitedError: If the request was made too frequently and the client must wait before retrying.
	// - DeviceQuotaExceededError: If the client has exceeded its quota for creating devices.
	// - BadRequestError: If the request is invalid: DNS server does not exist, creation of auto-devices is disabled or human_id validation failed.
	// - AuthenticationFailedError: If the authentication failed.
	CreateDeviceByHumanId(ctx context.Context, in *CreateDeviceRequest, opts ...grpc.CallOption) (*CreateDeviceResponse, error)
}

type dNSServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDNSServiceClient(cc grpc.ClientConnInterface) DNSServiceClient {
	return &dNSServiceClient{cc}
}

func (c *dNSServiceClient) GetDNSProfiles(ctx context.Context, in *DNSProfilesRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[DNSProfile], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &DNSService_ServiceDesc.Streams[0], DNSService_GetDNSProfiles_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[DNSProfilesRequest, DNSProfile]{ClientStream: stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type DNSService_GetDNSProfilesClient = grpc.ServerStreamingClient[DNSProfile]

func (c *dNSServiceClient) SaveDevicesBillingStat(ctx context.Context, opts ...grpc.CallOption) (grpc.ClientStreamingClient[DeviceBillingStat, emptypb.Empty], error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	stream, err := c.cc.NewStream(ctx, &DNSService_ServiceDesc.Streams[1], DNSService_SaveDevicesBillingStat_FullMethodName, cOpts...)
	if err != nil {
		return nil, err
	}
	x := &grpc.GenericClientStream[DeviceBillingStat, emptypb.Empty]{ClientStream: stream}
	return x, nil
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type DNSService_SaveDevicesBillingStatClient = grpc.ClientStreamingClient[DeviceBillingStat, emptypb.Empty]

func (c *dNSServiceClient) CreateDeviceByHumanId(ctx context.Context, in *CreateDeviceRequest, opts ...grpc.CallOption) (*CreateDeviceResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CreateDeviceResponse)
	err := c.cc.Invoke(ctx, DNSService_CreateDeviceByHumanId_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DNSServiceServer is the server API for DNSService service.
// All implementations must embed UnimplementedDNSServiceServer
// for forward compatibility.
type DNSServiceServer interface {
	// Gets DNS profiles.
	//
	// Field "sync_time" in DNSProfilesRequest - pass to return the latest updates after this time moment.
	//
	// The trailers headers will include a "sync_time", given in milliseconds,
	// that should be used for subsequent incremental DNS profile synchronization requests.
	//
	// This method may return the following errors:
	// - RateLimitedError: If too many "full sync" concurrent requests are made.
	// - AuthenticationFailedError: If the authentication failed.
	GetDNSProfiles(*DNSProfilesRequest, grpc.ServerStreamingServer[DNSProfile]) error
	// Stores devices activity.
	//
	// This method may return the following errors:
	// - AuthenticationFailedError: If the authentication failed.
	SaveDevicesBillingStat(grpc.ClientStreamingServer[DeviceBillingStat, emptypb.Empty]) error
	// Create device by "human_id".
	//
	// This method may return the following errors:
	// - RateLimitedError: If the request was made too frequently and the client must wait before retrying.
	// - DeviceQuotaExceededError: If the client has exceeded its quota for creating devices.
	// - BadRequestError: If the request is invalid: DNS server does not exist, creation of auto-devices is disabled or human_id validation failed.
	// - AuthenticationFailedError: If the authentication failed.
	CreateDeviceByHumanId(context.Context, *CreateDeviceRequest) (*CreateDeviceResponse, error)
	mustEmbedUnimplementedDNSServiceServer()
}

// UnimplementedDNSServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedDNSServiceServer struct{}

func (UnimplementedDNSServiceServer) GetDNSProfiles(*DNSProfilesRequest, grpc.ServerStreamingServer[DNSProfile]) error {
	return status.Errorf(codes.Unimplemented, "method GetDNSProfiles not implemented")
}
func (UnimplementedDNSServiceServer) SaveDevicesBillingStat(grpc.ClientStreamingServer[DeviceBillingStat, emptypb.Empty]) error {
	return status.Errorf(codes.Unimplemented, "method SaveDevicesBillingStat not implemented")
}
func (UnimplementedDNSServiceServer) CreateDeviceByHumanId(context.Context, *CreateDeviceRequest) (*CreateDeviceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDeviceByHumanId not implemented")
}
func (UnimplementedDNSServiceServer) mustEmbedUnimplementedDNSServiceServer() {}
func (UnimplementedDNSServiceServer) testEmbeddedByValue()                    {}

// UnsafeDNSServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DNSServiceServer will
// result in compilation errors.
type UnsafeDNSServiceServer interface {
	mustEmbedUnimplementedDNSServiceServer()
}

func RegisterDNSServiceServer(s grpc.ServiceRegistrar, srv DNSServiceServer) {
	// If the following call pancis, it indicates UnimplementedDNSServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&DNSService_ServiceDesc, srv)
}

func _DNSService_GetDNSProfiles_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DNSProfilesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DNSServiceServer).GetDNSProfiles(m, &grpc.GenericServerStream[DNSProfilesRequest, DNSProfile]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type DNSService_GetDNSProfilesServer = grpc.ServerStreamingServer[DNSProfile]

func _DNSService_SaveDevicesBillingStat_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DNSServiceServer).SaveDevicesBillingStat(&grpc.GenericServerStream[DeviceBillingStat, emptypb.Empty]{ServerStream: stream})
}

// This type alias is provided for backwards compatibility with existing code that references the prior non-generic stream type by name.
type DNSService_SaveDevicesBillingStatServer = grpc.ClientStreamingServer[DeviceBillingStat, emptypb.Empty]

func _DNSService_CreateDeviceByHumanId_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateDeviceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DNSServiceServer).CreateDeviceByHumanId(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DNSService_CreateDeviceByHumanId_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DNSServiceServer).CreateDeviceByHumanId(ctx, req.(*CreateDeviceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DNSService_ServiceDesc is the grpc.ServiceDesc for DNSService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DNSService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DNSService",
	HandlerType: (*DNSServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "createDeviceByHumanId",
			Handler:    _DNSService_CreateDeviceByHumanId_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "getDNSProfiles",
			Handler:       _DNSService_GetDNSProfiles_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "saveDevicesBillingStat",
			Handler:       _DNSService_SaveDevicesBillingStat_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "dns.proto",
}
