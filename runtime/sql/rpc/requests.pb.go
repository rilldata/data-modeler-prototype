// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: sql/src/main/proto/requests.proto

package rpc

import (
	ast "github.com/rilldata/rill/runtime/sql/ast"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Dialect int32

const (
	Dialect_DRUID  Dialect = 0
	Dialect_DUCKDB Dialect = 1
)

// Enum value maps for Dialect.
var (
	Dialect_name = map[int32]string{
		0: "DRUID",
		1: "DUCKDB",
	}
	Dialect_value = map[string]int32{
		"DRUID":  0,
		"DUCKDB": 1,
	}
)

func (x Dialect) Enum() *Dialect {
	p := new(Dialect)
	*p = x
	return p
}

func (x Dialect) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Dialect) Descriptor() protoreflect.EnumDescriptor {
	return file_sql_src_main_proto_requests_proto_enumTypes[0].Descriptor()
}

func (Dialect) Type() protoreflect.EnumType {
	return &file_sql_src_main_proto_requests_proto_enumTypes[0]
}

func (x Dialect) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Dialect.Descriptor instead.
func (Dialect) EnumDescriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{0}
}

type Request struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Request:
	//
	//	*Request_ParseRequest
	//	*Request_TranspileRequest
	Request isRequest_Request `protobuf_oneof:"request"`
}

func (x *Request) Reset() {
	*x = Request{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sql_src_main_proto_requests_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Request) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Request) ProtoMessage() {}

func (x *Request) ProtoReflect() protoreflect.Message {
	mi := &file_sql_src_main_proto_requests_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Request.ProtoReflect.Descriptor instead.
func (*Request) Descriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{0}
}

func (m *Request) GetRequest() isRequest_Request {
	if m != nil {
		return m.Request
	}
	return nil
}

func (x *Request) GetParseRequest() *ParseRequest {
	if x, ok := x.GetRequest().(*Request_ParseRequest); ok {
		return x.ParseRequest
	}
	return nil
}

func (x *Request) GetTranspileRequest() *TranspileRequest {
	if x, ok := x.GetRequest().(*Request_TranspileRequest); ok {
		return x.TranspileRequest
	}
	return nil
}

type isRequest_Request interface {
	isRequest_Request()
}

type Request_ParseRequest struct {
	ParseRequest *ParseRequest `protobuf:"bytes,1,opt,name=parse_request,json=parseRequest,proto3,oneof"`
}

type Request_TranspileRequest struct {
	// UnparseRequest unparse_request = 2;
	TranspileRequest *TranspileRequest `protobuf:"bytes,3,opt,name=transpile_request,json=transpileRequest,proto3,oneof"` // ApplyRequest apply_request = 4;
}

func (*Request_ParseRequest) isRequest_Request() {}

func (*Request_TranspileRequest) isRequest_Request() {}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Response:
	//
	//	*Response_ParseResponse
	//	*Response_TranspileResponse
	Response isResponse_Response `protobuf_oneof:"response"`
	Error    *Error              `protobuf:"bytes,5,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sql_src_main_proto_requests_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_sql_src_main_proto_requests_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{1}
}

func (m *Response) GetResponse() isResponse_Response {
	if m != nil {
		return m.Response
	}
	return nil
}

func (x *Response) GetParseResponse() *ParseResponse {
	if x, ok := x.GetResponse().(*Response_ParseResponse); ok {
		return x.ParseResponse
	}
	return nil
}

func (x *Response) GetTranspileResponse() *TranspileResponse {
	if x, ok := x.GetResponse().(*Response_TranspileResponse); ok {
		return x.TranspileResponse
	}
	return nil
}

func (x *Response) GetError() *Error {
	if x != nil {
		return x.Error
	}
	return nil
}

type isResponse_Response interface {
	isResponse_Response()
}

type Response_ParseResponse struct {
	ParseResponse *ParseResponse `protobuf:"bytes,1,opt,name=parse_response,json=parseResponse,proto3,oneof"`
}

type Response_TranspileResponse struct {
	// UnparseResponse unparse_response = 2;
	TranspileResponse *TranspileResponse `protobuf:"bytes,3,opt,name=transpile_response,json=transpileResponse,proto3,oneof"` // ApplyResponse apply_response = 4;
}

func (*Response_ParseResponse) isResponse_Response() {}

func (*Response_TranspileResponse) isResponse_Response() {}

type ParseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sql         string `protobuf:"bytes,1,opt,name=sql,proto3" json:"sql,omitempty"`
	Catalog     string `protobuf:"bytes,2,opt,name=catalog,proto3" json:"catalog,omitempty"`
	AddTypeInfo bool   `protobuf:"varint,3,opt,name=addTypeInfo,proto3" json:"addTypeInfo,omitempty"`
}

func (x *ParseRequest) Reset() {
	*x = ParseRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sql_src_main_proto_requests_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseRequest) ProtoMessage() {}

func (x *ParseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_sql_src_main_proto_requests_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseRequest.ProtoReflect.Descriptor instead.
func (*ParseRequest) Descriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{2}
}

func (x *ParseRequest) GetSql() string {
	if x != nil {
		return x.Sql
	}
	return ""
}

func (x *ParseRequest) GetCatalog() string {
	if x != nil {
		return x.Catalog
	}
	return ""
}

func (x *ParseRequest) GetAddTypeInfo() bool {
	if x != nil {
		return x.AddTypeInfo
	}
	return false
}

type ParseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ast *ast.SqlNodeProto `protobuf:"bytes,1,opt,name=ast,proto3" json:"ast,omitempty"`
}

func (x *ParseResponse) Reset() {
	*x = ParseResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sql_src_main_proto_requests_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ParseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ParseResponse) ProtoMessage() {}

func (x *ParseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_sql_src_main_proto_requests_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ParseResponse.ProtoReflect.Descriptor instead.
func (*ParseResponse) Descriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{3}
}

func (x *ParseResponse) GetAst() *ast.SqlNodeProto {
	if x != nil {
		return x.Ast
	}
	return nil
}

type TranspileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sql     string  `protobuf:"bytes,1,opt,name=sql,proto3" json:"sql,omitempty"`
	Dialect Dialect `protobuf:"varint,2,opt,name=dialect,proto3,enum=rill.sql.v1.Dialect" json:"dialect,omitempty"`
	Catalog string  `protobuf:"bytes,3,opt,name=catalog,proto3" json:"catalog,omitempty"`
}

func (x *TranspileRequest) Reset() {
	*x = TranspileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sql_src_main_proto_requests_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TranspileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TranspileRequest) ProtoMessage() {}

func (x *TranspileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_sql_src_main_proto_requests_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TranspileRequest.ProtoReflect.Descriptor instead.
func (*TranspileRequest) Descriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{4}
}

func (x *TranspileRequest) GetSql() string {
	if x != nil {
		return x.Sql
	}
	return ""
}

func (x *TranspileRequest) GetDialect() Dialect {
	if x != nil {
		return x.Dialect
	}
	return Dialect_DRUID
}

func (x *TranspileRequest) GetCatalog() string {
	if x != nil {
		return x.Catalog
	}
	return ""
}

type TranspileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Sql string `protobuf:"bytes,1,opt,name=sql,proto3" json:"sql,omitempty"`
}

func (x *TranspileResponse) Reset() {
	*x = TranspileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sql_src_main_proto_requests_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TranspileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TranspileResponse) ProtoMessage() {}

func (x *TranspileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_sql_src_main_proto_requests_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TranspileResponse.ProtoReflect.Descriptor instead.
func (*TranspileResponse) Descriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{5}
}

func (x *TranspileResponse) GetSql() string {
	if x != nil {
		return x.Sql
	}
	return ""
}

type Error struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Message    string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	StackTrace string `protobuf:"bytes,2,opt,name=stack_trace,json=stackTrace,proto3" json:"stack_trace,omitempty"`
}

func (x *Error) Reset() {
	*x = Error{}
	if protoimpl.UnsafeEnabled {
		mi := &file_sql_src_main_proto_requests_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Error) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Error) ProtoMessage() {}

func (x *Error) ProtoReflect() protoreflect.Message {
	mi := &file_sql_src_main_proto_requests_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Error.ProtoReflect.Descriptor instead.
func (*Error) Descriptor() ([]byte, []int) {
	return file_sql_src_main_proto_requests_proto_rawDescGZIP(), []int{6}
}

func (x *Error) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Error) GetStackTrace() string {
	if x != nil {
		return x.StackTrace
	}
	return ""
}

var File_sql_src_main_proto_requests_proto protoreflect.FileDescriptor

var file_sql_src_main_proto_requests_proto_rawDesc = []byte{
	0x0a, 0x21, 0x73, 0x71, 0x6c, 0x2f, 0x73, 0x72, 0x63, 0x2f, 0x6d, 0x61, 0x69, 0x6e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x0b, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x73, 0x71, 0x6c, 0x2e, 0x76, 0x31,
	0x1a, 0x1c, 0x73, 0x71, 0x6c, 0x2f, 0x73, 0x72, 0x63, 0x2f, 0x6d, 0x61, 0x69, 0x6e, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x73, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa4,
	0x01, 0x0a, 0x07, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x40, 0x0a, 0x0d, 0x70, 0x61,
	0x72, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x73, 0x71, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x50, 0x61, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x0c,
	0x70, 0x61, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x4c, 0x0a, 0x11,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x69, 0x6c, 0x65, 0x5f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x73,
	0x71, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x69, 0x6c, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x48, 0x00, 0x52, 0x10, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70,
	0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x42, 0x09, 0x0a, 0x07, 0x72, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0xd6, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x43, 0x0a, 0x0e, 0x70, 0x61, 0x72, 0x73, 0x65, 0x5f, 0x72, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x72, 0x69, 0x6c,
	0x6c, 0x2e, 0x73, 0x71, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x61, 0x72, 0x73, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x0d, 0x70, 0x61, 0x72, 0x73, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x4f, 0x0a, 0x12, 0x74, 0x72, 0x61, 0x6e, 0x73,
	0x70, 0x69, 0x6c, 0x65, 0x5f, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x73, 0x71, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x48, 0x00, 0x52, 0x11, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x70, 0x69, 0x6c, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x28, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f,
	0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x73,
	0x71, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x52, 0x05, 0x65, 0x72, 0x72,
	0x6f, 0x72, 0x42, 0x0a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x5c,
	0x0a, 0x0c, 0x50, 0x61, 0x72, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x73, 0x71, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x71, 0x6c,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x12, 0x20, 0x0a, 0x0b, 0x61, 0x64,
	0x64, 0x54, 0x79, 0x70, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x0b, 0x61, 0x64, 0x64, 0x54, 0x79, 0x70, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x3c, 0x0a, 0x0d,
	0x50, 0x61, 0x72, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2b, 0x0a,
	0x03, 0x61, 0x73, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x72, 0x69, 0x6c,
	0x6c, 0x2e, 0x73, 0x71, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x71, 0x6c, 0x4e, 0x6f, 0x64, 0x65,
	0x50, 0x72, 0x6f, 0x74, 0x6f, 0x52, 0x03, 0x61, 0x73, 0x74, 0x22, 0x6e, 0x0a, 0x10, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x10,
	0x0a, 0x03, 0x73, 0x71, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x71, 0x6c,
	0x12, 0x2e, 0x0a, 0x07, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x14, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x73, 0x71, 0x6c, 0x2e, 0x76, 0x31, 0x2e,
	0x44, 0x69, 0x61, 0x6c, 0x65, 0x63, 0x74, 0x52, 0x07, 0x64, 0x69, 0x61, 0x6c, 0x65, 0x63, 0x74,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x22, 0x25, 0x0a, 0x11, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x70, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x10, 0x0a, 0x03, 0x73, 0x71, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x71,
	0x6c, 0x22, 0x42, 0x0a, 0x05, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x5f, 0x74, 0x72,
	0x61, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x74, 0x61, 0x63, 0x6b,
	0x54, 0x72, 0x61, 0x63, 0x65, 0x2a, 0x20, 0x0a, 0x07, 0x44, 0x69, 0x61, 0x6c, 0x65, 0x63, 0x74,
	0x12, 0x09, 0x0a, 0x05, 0x44, 0x52, 0x55, 0x49, 0x44, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x44,
	0x55, 0x43, 0x4b, 0x44, 0x42, 0x10, 0x01, 0x42, 0x21, 0x0a, 0x1f, 0x63, 0x6f, 0x6d, 0x2e, 0x72,
	0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x67, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x64, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_sql_src_main_proto_requests_proto_rawDescOnce sync.Once
	file_sql_src_main_proto_requests_proto_rawDescData = file_sql_src_main_proto_requests_proto_rawDesc
)

func file_sql_src_main_proto_requests_proto_rawDescGZIP() []byte {
	file_sql_src_main_proto_requests_proto_rawDescOnce.Do(func() {
		file_sql_src_main_proto_requests_proto_rawDescData = protoimpl.X.CompressGZIP(file_sql_src_main_proto_requests_proto_rawDescData)
	})
	return file_sql_src_main_proto_requests_proto_rawDescData
}

var file_sql_src_main_proto_requests_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_sql_src_main_proto_requests_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_sql_src_main_proto_requests_proto_goTypes = []interface{}{
	(Dialect)(0),              // 0: rill.sql.v1.Dialect
	(*Request)(nil),           // 1: rill.sql.v1.Request
	(*Response)(nil),          // 2: rill.sql.v1.Response
	(*ParseRequest)(nil),      // 3: rill.sql.v1.ParseRequest
	(*ParseResponse)(nil),     // 4: rill.sql.v1.ParseResponse
	(*TranspileRequest)(nil),  // 5: rill.sql.v1.TranspileRequest
	(*TranspileResponse)(nil), // 6: rill.sql.v1.TranspileResponse
	(*Error)(nil),             // 7: rill.sql.v1.Error
	(*ast.SqlNodeProto)(nil),  // 8: rill.sql.v1.SqlNodeProto
}
var file_sql_src_main_proto_requests_proto_depIdxs = []int32{
	3, // 0: rill.sql.v1.Request.parse_request:type_name -> rill.sql.v1.ParseRequest
	5, // 1: rill.sql.v1.Request.transpile_request:type_name -> rill.sql.v1.TranspileRequest
	4, // 2: rill.sql.v1.Response.parse_response:type_name -> rill.sql.v1.ParseResponse
	6, // 3: rill.sql.v1.Response.transpile_response:type_name -> rill.sql.v1.TranspileResponse
	7, // 4: rill.sql.v1.Response.error:type_name -> rill.sql.v1.Error
	8, // 5: rill.sql.v1.ParseResponse.ast:type_name -> rill.sql.v1.SqlNodeProto
	0, // 6: rill.sql.v1.TranspileRequest.dialect:type_name -> rill.sql.v1.Dialect
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_sql_src_main_proto_requests_proto_init() }
func file_sql_src_main_proto_requests_proto_init() {
	if File_sql_src_main_proto_requests_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_sql_src_main_proto_requests_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Request); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sql_src_main_proto_requests_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sql_src_main_proto_requests_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParseRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sql_src_main_proto_requests_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ParseResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sql_src_main_proto_requests_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TranspileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sql_src_main_proto_requests_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TranspileResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_sql_src_main_proto_requests_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Error); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_sql_src_main_proto_requests_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*Request_ParseRequest)(nil),
		(*Request_TranspileRequest)(nil),
	}
	file_sql_src_main_proto_requests_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*Response_ParseResponse)(nil),
		(*Response_TranspileResponse)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_sql_src_main_proto_requests_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_sql_src_main_proto_requests_proto_goTypes,
		DependencyIndexes: file_sql_src_main_proto_requests_proto_depIdxs,
		EnumInfos:         file_sql_src_main_proto_requests_proto_enumTypes,
		MessageInfos:      file_sql_src_main_proto_requests_proto_msgTypes,
	}.Build()
	File_sql_src_main_proto_requests_proto = out.File
	file_sql_src_main_proto_requests_proto_rawDesc = nil
	file_sql_src_main_proto_requests_proto_goTypes = nil
	file_sql_src_main_proto_requests_proto_depIdxs = nil
}
