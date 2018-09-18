// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/iam/v1/iam_policy.proto

package iam

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	math "math"
)

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Request message for `SetIamPolicy` method.
type SetIamPolicyRequest struct {
	// REQUIRED: The resource for which the policy is being specified.
	// `resource` is usually specified as a path. For example, a Project
	// resource is specified as `projects/{project}`.
	Resource string `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	// REQUIRED: The complete policy to be applied to the `resource`. The size of
	// the policy is limited to a few 10s of KB. An empty policy is a
	// valid policy but certain Cloud Platform services (such as Projects)
	// might reject them.
	Policy               *Policy  `protobuf:"bytes,2,opt,name=policy,proto3" json:"policy,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SetIamPolicyRequest) Reset()         { *m = SetIamPolicyRequest{} }
func (m *SetIamPolicyRequest) String() string { return proto.CompactTextString(m) }
func (*SetIamPolicyRequest) ProtoMessage()    {}
func (*SetIamPolicyRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2728eb97d748a32, []int{0}
}
func (m *SetIamPolicyRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SetIamPolicyRequest.Unmarshal(m, b)
}
func (m *SetIamPolicyRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SetIamPolicyRequest.Marshal(b, m, deterministic)
}
func (m *SetIamPolicyRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SetIamPolicyRequest.Merge(m, src)
}
func (m *SetIamPolicyRequest) XXX_Size() int {
	return xxx_messageInfo_SetIamPolicyRequest.Size(m)
}
func (m *SetIamPolicyRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_SetIamPolicyRequest.DiscardUnknown(m)
}

var xxx_messageInfo_SetIamPolicyRequest proto.InternalMessageInfo

func (m *SetIamPolicyRequest) GetResource() string {
	if m != nil {
		return m.Resource
	}
	return ""
}

func (m *SetIamPolicyRequest) GetPolicy() *Policy {
	if m != nil {
		return m.Policy
	}
	return nil
}

// Request message for `GetIamPolicy` method.
type GetIamPolicyRequest struct {
	// REQUIRED: The resource for which the policy is being requested.
	// `resource` is usually specified as a path. For example, a Project
	// resource is specified as `projects/{project}`.
	Resource             string   `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetIamPolicyRequest) Reset()         { *m = GetIamPolicyRequest{} }
func (m *GetIamPolicyRequest) String() string { return proto.CompactTextString(m) }
func (*GetIamPolicyRequest) ProtoMessage()    {}
func (*GetIamPolicyRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2728eb97d748a32, []int{1}
}
func (m *GetIamPolicyRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetIamPolicyRequest.Unmarshal(m, b)
}
func (m *GetIamPolicyRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetIamPolicyRequest.Marshal(b, m, deterministic)
}
func (m *GetIamPolicyRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetIamPolicyRequest.Merge(m, src)
}
func (m *GetIamPolicyRequest) XXX_Size() int {
	return xxx_messageInfo_GetIamPolicyRequest.Size(m)
}
func (m *GetIamPolicyRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetIamPolicyRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetIamPolicyRequest proto.InternalMessageInfo

func (m *GetIamPolicyRequest) GetResource() string {
	if m != nil {
		return m.Resource
	}
	return ""
}

// Request message for `TestIamPermissions` method.
type TestIamPermissionsRequest struct {
	// REQUIRED: The resource for which the policy detail is being requested.
	// `resource` is usually specified as a path. For example, a Project
	// resource is specified as `projects/{project}`.
	Resource string `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	// The set of permissions to check for the `resource`. Permissions with
	// wildcards (such as '*' or 'storage.*') are not allowed. For more
	// information see
	// [IAM Overview](https://cloud.google.com/iam/docs/overview#permissions).
	Permissions          []string `protobuf:"bytes,2,rep,name=permissions,proto3" json:"permissions,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TestIamPermissionsRequest) Reset()         { *m = TestIamPermissionsRequest{} }
func (m *TestIamPermissionsRequest) String() string { return proto.CompactTextString(m) }
func (*TestIamPermissionsRequest) ProtoMessage()    {}
func (*TestIamPermissionsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2728eb97d748a32, []int{2}
}
func (m *TestIamPermissionsRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TestIamPermissionsRequest.Unmarshal(m, b)
}
func (m *TestIamPermissionsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TestIamPermissionsRequest.Marshal(b, m, deterministic)
}
func (m *TestIamPermissionsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestIamPermissionsRequest.Merge(m, src)
}
func (m *TestIamPermissionsRequest) XXX_Size() int {
	return xxx_messageInfo_TestIamPermissionsRequest.Size(m)
}
func (m *TestIamPermissionsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TestIamPermissionsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TestIamPermissionsRequest proto.InternalMessageInfo

func (m *TestIamPermissionsRequest) GetResource() string {
	if m != nil {
		return m.Resource
	}
	return ""
}

func (m *TestIamPermissionsRequest) GetPermissions() []string {
	if m != nil {
		return m.Permissions
	}
	return nil
}

// Response message for `TestIamPermissions` method.
type TestIamPermissionsResponse struct {
	// A subset of `TestPermissionsRequest.permissions` that the caller is
	// allowed.
	Permissions          []string `protobuf:"bytes,1,rep,name=permissions,proto3" json:"permissions,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TestIamPermissionsResponse) Reset()         { *m = TestIamPermissionsResponse{} }
func (m *TestIamPermissionsResponse) String() string { return proto.CompactTextString(m) }
func (*TestIamPermissionsResponse) ProtoMessage()    {}
func (*TestIamPermissionsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_d2728eb97d748a32, []int{3}
}
func (m *TestIamPermissionsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TestIamPermissionsResponse.Unmarshal(m, b)
}
func (m *TestIamPermissionsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TestIamPermissionsResponse.Marshal(b, m, deterministic)
}
func (m *TestIamPermissionsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestIamPermissionsResponse.Merge(m, src)
}
func (m *TestIamPermissionsResponse) XXX_Size() int {
	return xxx_messageInfo_TestIamPermissionsResponse.Size(m)
}
func (m *TestIamPermissionsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TestIamPermissionsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TestIamPermissionsResponse proto.InternalMessageInfo

func (m *TestIamPermissionsResponse) GetPermissions() []string {
	if m != nil {
		return m.Permissions
	}
	return nil
}

func init() {
	proto.RegisterType((*SetIamPolicyRequest)(nil), "google.iam.v1.SetIamPolicyRequest")
	proto.RegisterType((*GetIamPolicyRequest)(nil), "google.iam.v1.GetIamPolicyRequest")
	proto.RegisterType((*TestIamPermissionsRequest)(nil), "google.iam.v1.TestIamPermissionsRequest")
	proto.RegisterType((*TestIamPermissionsResponse)(nil), "google.iam.v1.TestIamPermissionsResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// IAMPolicyClient is the client API for IAMPolicy service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type IAMPolicyClient interface {
	// Sets the access control policy on the specified resource. Replaces any
	// existing policy.
	SetIamPolicy(ctx context.Context, in *SetIamPolicyRequest, opts ...grpc.CallOption) (*Policy, error)
	// Gets the access control policy for a resource.
	// Returns an empty policy if the resource exists and does not have a policy
	// set.
	GetIamPolicy(ctx context.Context, in *GetIamPolicyRequest, opts ...grpc.CallOption) (*Policy, error)
	// Returns permissions that a caller has on the specified resource.
	// If the resource does not exist, this will return an empty set of
	// permissions, not a NOT_FOUND error.
	TestIamPermissions(ctx context.Context, in *TestIamPermissionsRequest, opts ...grpc.CallOption) (*TestIamPermissionsResponse, error)
}

type iAMPolicyClient struct {
	cc *grpc.ClientConn
}

func NewIAMPolicyClient(cc *grpc.ClientConn) IAMPolicyClient {
	return &iAMPolicyClient{cc}
}

func (c *iAMPolicyClient) SetIamPolicy(ctx context.Context, in *SetIamPolicyRequest, opts ...grpc.CallOption) (*Policy, error) {
	out := new(Policy)
	err := c.cc.Invoke(ctx, "/google.iam.v1.IAMPolicy/SetIamPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMPolicyClient) GetIamPolicy(ctx context.Context, in *GetIamPolicyRequest, opts ...grpc.CallOption) (*Policy, error) {
	out := new(Policy)
	err := c.cc.Invoke(ctx, "/google.iam.v1.IAMPolicy/GetIamPolicy", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iAMPolicyClient) TestIamPermissions(ctx context.Context, in *TestIamPermissionsRequest, opts ...grpc.CallOption) (*TestIamPermissionsResponse, error) {
	out := new(TestIamPermissionsResponse)
	err := c.cc.Invoke(ctx, "/google.iam.v1.IAMPolicy/TestIamPermissions", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IAMPolicyServer is the server API for IAMPolicy service.
type IAMPolicyServer interface {
	// Sets the access control policy on the specified resource. Replaces any
	// existing policy.
	SetIamPolicy(context.Context, *SetIamPolicyRequest) (*Policy, error)
	// Gets the access control policy for a resource.
	// Returns an empty policy if the resource exists and does not have a policy
	// set.
	GetIamPolicy(context.Context, *GetIamPolicyRequest) (*Policy, error)
	// Returns permissions that a caller has on the specified resource.
	// If the resource does not exist, this will return an empty set of
	// permissions, not a NOT_FOUND error.
	TestIamPermissions(context.Context, *TestIamPermissionsRequest) (*TestIamPermissionsResponse, error)
}

func RegisterIAMPolicyServer(s *grpc.Server, srv IAMPolicyServer) {
	s.RegisterService(&_IAMPolicy_serviceDesc, srv)
}

func _IAMPolicy_SetIamPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetIamPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMPolicyServer).SetIamPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.iam.v1.IAMPolicy/SetIamPolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMPolicyServer).SetIamPolicy(ctx, req.(*SetIamPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAMPolicy_GetIamPolicy_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIamPolicyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMPolicyServer).GetIamPolicy(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.iam.v1.IAMPolicy/GetIamPolicy",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMPolicyServer).GetIamPolicy(ctx, req.(*GetIamPolicyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IAMPolicy_TestIamPermissions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestIamPermissionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IAMPolicyServer).TestIamPermissions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.iam.v1.IAMPolicy/TestIamPermissions",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IAMPolicyServer).TestIamPermissions(ctx, req.(*TestIamPermissionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _IAMPolicy_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.iam.v1.IAMPolicy",
	HandlerType: (*IAMPolicyServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetIamPolicy",
			Handler:    _IAMPolicy_SetIamPolicy_Handler,
		},
		{
			MethodName: "GetIamPolicy",
			Handler:    _IAMPolicy_GetIamPolicy_Handler,
		},
		{
			MethodName: "TestIamPermissions",
			Handler:    _IAMPolicy_TestIamPermissions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/iam/v1/iam_policy.proto",
}

func init() { proto.RegisterFile("google/iam/v1/iam_policy.proto", fileDescriptor_d2728eb97d748a32) }

var fileDescriptor_d2728eb97d748a32 = []byte{
	// 411 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4b, 0xcf, 0xcf, 0x4f,
	0xcf, 0x49, 0xd5, 0xcf, 0x4c, 0xcc, 0xd5, 0x2f, 0x33, 0x04, 0x51, 0xf1, 0x05, 0xf9, 0x39, 0x99,
	0xc9, 0x95, 0x7a, 0x05, 0x45, 0xf9, 0x25, 0xf9, 0x42, 0xbc, 0x10, 0x79, 0xbd, 0xcc, 0xc4, 0x5c,
	0xbd, 0x32, 0x43, 0x29, 0x19, 0xa8, 0xf2, 0xc4, 0x82, 0x4c, 0xfd, 0xc4, 0xbc, 0xbc, 0xfc, 0x92,
	0xc4, 0x92, 0xcc, 0xfc, 0xbc, 0x62, 0x88, 0x62, 0x29, 0x29, 0x54, 0xc3, 0x90, 0x0d, 0x52, 0x4a,
	0xe0, 0x12, 0x0e, 0x4e, 0x2d, 0xf1, 0x4c, 0xcc, 0x0d, 0x00, 0x8b, 0x06, 0xa5, 0x16, 0x96, 0xa6,
	0x16, 0x97, 0x08, 0x49, 0x71, 0x71, 0x14, 0xa5, 0x16, 0xe7, 0x97, 0x16, 0x25, 0xa7, 0x4a, 0x30,
	0x2a, 0x30, 0x6a, 0x70, 0x06, 0xc1, 0xf9, 0x42, 0xba, 0x5c, 0x6c, 0x10, 0x23, 0x24, 0x98, 0x14,
	0x18, 0x35, 0xb8, 0x8d, 0x44, 0xf5, 0x50, 0x1c, 0xa3, 0x07, 0x35, 0x09, 0xaa, 0x48, 0xc9, 0x90,
	0x4b, 0xd8, 0x9d, 0x34, 0x1b, 0x94, 0x22, 0xb9, 0x24, 0x43, 0x52, 0x8b, 0xc1, 0x7a, 0x52, 0x8b,
	0x72, 0x33, 0x8b, 0x8b, 0x41, 0x9e, 0x21, 0xc6, 0x69, 0x0a, 0x5c, 0xdc, 0x05, 0x08, 0x1d, 0x12,
	0x4c, 0x0a, 0xcc, 0x1a, 0x9c, 0x41, 0xc8, 0x42, 0x4a, 0x76, 0x5c, 0x52, 0xd8, 0x8c, 0x2e, 0x2e,
	0xc8, 0xcf, 0x2b, 0xc6, 0xd0, 0xcf, 0x88, 0xa1, 0xdf, 0x68, 0x0a, 0x33, 0x17, 0xa7, 0xa7, 0xa3,
	0x2f, 0xc4, 0x2f, 0x42, 0x25, 0x5c, 0x3c, 0xc8, 0xa1, 0x27, 0xa4, 0x84, 0x16, 0x14, 0x58, 0x82,
	0x56, 0x0a, 0x7b, 0x70, 0x29, 0x69, 0x36, 0x5d, 0x7e, 0x32, 0x99, 0x49, 0x59, 0x49, 0x0e, 0x14,
	0x45, 0xd5, 0x30, 0x1f, 0xd9, 0x6a, 0x69, 0xd5, 0x5a, 0x15, 0x23, 0x99, 0x62, 0xc5, 0xa8, 0x05,
	0xb2, 0xd5, 0x1d, 0x9f, 0xad, 0xee, 0x54, 0xb1, 0x35, 0x1d, 0xcd, 0xd6, 0x59, 0x8c, 0x5c, 0x42,
	0x98, 0x41, 0x27, 0xa4, 0x81, 0x66, 0x30, 0xce, 0x88, 0x93, 0xd2, 0x24, 0x42, 0x25, 0x24, 0x1e,
	0x94, 0xf4, 0xc1, 0xce, 0xd2, 0x54, 0x52, 0xc1, 0x74, 0x56, 0x09, 0x86, 0x2e, 0x2b, 0x46, 0x2d,
	0xa7, 0x36, 0x46, 0x2e, 0xc1, 0xe4, 0xfc, 0x5c, 0x54, 0x1b, 0x9c, 0xf8, 0xe0, 0x1e, 0x08, 0x00,
	0x25, 0xf6, 0x00, 0xc6, 0x28, 0x03, 0xa8, 0x82, 0xf4, 0xfc, 0x9c, 0xc4, 0xbc, 0x74, 0xbd, 0xfc,
	0xa2, 0x74, 0xfd, 0xf4, 0xd4, 0x3c, 0x70, 0x56, 0xd0, 0x87, 0x48, 0x25, 0x16, 0x64, 0x16, 0x43,
	0x73, 0x8a, 0x75, 0x66, 0x62, 0xee, 0x0f, 0x46, 0xc6, 0x55, 0x4c, 0xc2, 0xee, 0x10, 0x5d, 0xce,
	0x39, 0xf9, 0xa5, 0x29, 0x7a, 0x9e, 0x89, 0xb9, 0x7a, 0x61, 0x86, 0xa7, 0x60, 0xa2, 0x31, 0x60,
	0xd1, 0x18, 0xcf, 0xc4, 0xdc, 0x98, 0x30, 0xc3, 0x24, 0x36, 0xb0, 0x59, 0xc6, 0x80, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xea, 0x62, 0x8f, 0x22, 0xc1, 0x03, 0x00, 0x00,
}
