// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: rill/admin/v1/internal.proto

package adminv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StringPageToken struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Val string `protobuf:"bytes,1,opt,name=val,proto3" json:"val,omitempty"`
}

func (x *StringPageToken) Reset() {
	*x = StringPageToken{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_admin_v1_internal_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringPageToken) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringPageToken) ProtoMessage() {}

func (x *StringPageToken) ProtoReflect() protoreflect.Message {
	mi := &file_rill_admin_v1_internal_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringPageToken.ProtoReflect.Descriptor instead.
func (*StringPageToken) Descriptor() ([]byte, []int) {
	return file_rill_admin_v1_internal_proto_rawDescGZIP(), []int{0}
}

func (x *StringPageToken) GetVal() string {
	if x != nil {
		return x.Val
	}
	return ""
}

type StringTimestampPageToken struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Str string                 `protobuf:"bytes,1,opt,name=str,proto3" json:"str,omitempty"`
	Ts  *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=ts,proto3" json:"ts,omitempty"`
}

func (x *StringTimestampPageToken) Reset() {
	*x = StringTimestampPageToken{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_admin_v1_internal_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StringTimestampPageToken) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StringTimestampPageToken) ProtoMessage() {}

func (x *StringTimestampPageToken) ProtoReflect() protoreflect.Message {
	mi := &file_rill_admin_v1_internal_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StringTimestampPageToken.ProtoReflect.Descriptor instead.
func (*StringTimestampPageToken) Descriptor() ([]byte, []int) {
	return file_rill_admin_v1_internal_proto_rawDescGZIP(), []int{1}
}

func (x *StringTimestampPageToken) GetStr() string {
	if x != nil {
		return x.Str
	}
	return ""
}

func (x *StringTimestampPageToken) GetTs() *timestamppb.Timestamp {
	if x != nil {
		return x.Ts
	}
	return nil
}

var File_rill_admin_v1_internal_proto protoreflect.FileDescriptor

var file_rill_admin_v1_internal_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x2f,
	0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d,
	0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x23,
	0x0a, 0x0f, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65,
	0x6e, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x76, 0x61, 0x6c, 0x22, 0x58, 0x0a, 0x18, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x50, 0x61, 0x67, 0x65, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12,
	0x10, 0x0a, 0x03, 0x73, 0x74, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x74,
	0x72, 0x12, 0x2a, 0x0a, 0x02, 0x74, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x02, 0x74, 0x73, 0x42, 0xb2, 0x01,
	0x0a, 0x11, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x76, 0x31, 0x42, 0x0d, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x61, 0x64,
	0x6d, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x3b, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x76, 0x31, 0xa2, 0x02,
	0x03, 0x52, 0x41, 0x58, 0xaa, 0x02, 0x0d, 0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0d, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x19, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x41, 0x64, 0x6d, 0x69,
	0x6e, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x3a, 0x3a,
	0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_admin_v1_internal_proto_rawDescOnce sync.Once
	file_rill_admin_v1_internal_proto_rawDescData = file_rill_admin_v1_internal_proto_rawDesc
)

func file_rill_admin_v1_internal_proto_rawDescGZIP() []byte {
	file_rill_admin_v1_internal_proto_rawDescOnce.Do(func() {
		file_rill_admin_v1_internal_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_admin_v1_internal_proto_rawDescData)
	})
	return file_rill_admin_v1_internal_proto_rawDescData
}

var file_rill_admin_v1_internal_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_rill_admin_v1_internal_proto_goTypes = []interface{}{
	(*StringPageToken)(nil),          // 0: rill.admin.v1.StringPageToken
	(*StringTimestampPageToken)(nil), // 1: rill.admin.v1.StringTimestampPageToken
	(*timestamppb.Timestamp)(nil),    // 2: google.protobuf.Timestamp
}
var file_rill_admin_v1_internal_proto_depIdxs = []int32{
	2, // 0: rill.admin.v1.StringTimestampPageToken.ts:type_name -> google.protobuf.Timestamp
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_rill_admin_v1_internal_proto_init() }
func file_rill_admin_v1_internal_proto_init() {
	if File_rill_admin_v1_internal_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rill_admin_v1_internal_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringPageToken); i {
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
		file_rill_admin_v1_internal_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StringTimestampPageToken); i {
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
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rill_admin_v1_internal_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rill_admin_v1_internal_proto_goTypes,
		DependencyIndexes: file_rill_admin_v1_internal_proto_depIdxs,
		MessageInfos:      file_rill_admin_v1_internal_proto_msgTypes,
	}.Build()
	File_rill_admin_v1_internal_proto = out.File
	file_rill_admin_v1_internal_proto_rawDesc = nil
	file_rill_admin_v1_internal_proto_goTypes = nil
	file_rill_admin_v1_internal_proto_depIdxs = nil
}
