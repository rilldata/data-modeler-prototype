// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        (unknown)
// source: rill/admin/v1/telemetry.proto

package adminv1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RecordEventsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Events []*structpb.Struct `protobuf:"bytes,1,rep,name=events,proto3" json:"events,omitempty"`
}

func (x *RecordEventsRequest) Reset() {
	*x = RecordEventsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_admin_v1_telemetry_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecordEventsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordEventsRequest) ProtoMessage() {}

func (x *RecordEventsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rill_admin_v1_telemetry_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecordEventsRequest.ProtoReflect.Descriptor instead.
func (*RecordEventsRequest) Descriptor() ([]byte, []int) {
	return file_rill_admin_v1_telemetry_proto_rawDescGZIP(), []int{0}
}

func (x *RecordEventsRequest) GetEvents() []*structpb.Struct {
	if x != nil {
		return x.Events
	}
	return nil
}

type RecordEventsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RecordEventsResponse) Reset() {
	*x = RecordEventsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_admin_v1_telemetry_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecordEventsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordEventsResponse) ProtoMessage() {}

func (x *RecordEventsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rill_admin_v1_telemetry_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecordEventsResponse.ProtoReflect.Descriptor instead.
func (*RecordEventsResponse) Descriptor() ([]byte, []int) {
	return file_rill_admin_v1_telemetry_proto_rawDescGZIP(), []int{1}
}

var File_rill_admin_v1_telemetry_proto protoreflect.FileDescriptor

var file_rill_admin_v1_telemetry_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2f, 0x76, 0x31, 0x2f,
	0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x0d, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x1a, 0x1c,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74,
	0x72, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x46, 0x0a, 0x13, 0x52, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x2f, 0x0a, 0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x06, 0x65, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x22, 0x16, 0x0a, 0x14, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x8c, 0x01, 0x0a, 0x10, 0x54,
	0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x78, 0x0a, 0x0c, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x12,
	0x22, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x2e,
	0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2e, 0x76, 0x31, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1f, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x19,
	0x3a, 0x01, 0x2a, 0x22, 0x14, 0x2f, 0x76, 0x31, 0x2f, 0x74, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74,
	0x72, 0x79, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x73, 0x42, 0xb3, 0x01, 0x0a, 0x11, 0x63, 0x6f,
	0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x76, 0x31, 0x42,
	0x0e, 0x54, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x74, 0x72, 0x79, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50,
	0x01, 0x5a, 0x38, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x69,
	0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x61, 0x64, 0x6d, 0x69, 0x6e,
	0x2f, 0x76, 0x31, 0x3b, 0x61, 0x64, 0x6d, 0x69, 0x6e, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x41,
	0x58, 0xaa, 0x02, 0x0d, 0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x2e, 0x56,
	0x31, 0xca, 0x02, 0x0d, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x5c, 0x56,
	0x31, 0xe2, 0x02, 0x19, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x5c, 0x56,
	0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0f,
	0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a, 0x41, 0x64, 0x6d, 0x69, 0x6e, 0x3a, 0x3a, 0x56, 0x31, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_admin_v1_telemetry_proto_rawDescOnce sync.Once
	file_rill_admin_v1_telemetry_proto_rawDescData = file_rill_admin_v1_telemetry_proto_rawDesc
)

func file_rill_admin_v1_telemetry_proto_rawDescGZIP() []byte {
	file_rill_admin_v1_telemetry_proto_rawDescOnce.Do(func() {
		file_rill_admin_v1_telemetry_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_admin_v1_telemetry_proto_rawDescData)
	})
	return file_rill_admin_v1_telemetry_proto_rawDescData
}

var file_rill_admin_v1_telemetry_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_rill_admin_v1_telemetry_proto_goTypes = []interface{}{
	(*RecordEventsRequest)(nil),  // 0: rill.admin.v1.RecordEventsRequest
	(*RecordEventsResponse)(nil), // 1: rill.admin.v1.RecordEventsResponse
	(*structpb.Struct)(nil),      // 2: google.protobuf.Struct
}
var file_rill_admin_v1_telemetry_proto_depIdxs = []int32{
	2, // 0: rill.admin.v1.RecordEventsRequest.events:type_name -> google.protobuf.Struct
	0, // 1: rill.admin.v1.TelemetryService.RecordEvents:input_type -> rill.admin.v1.RecordEventsRequest
	1, // 2: rill.admin.v1.TelemetryService.RecordEvents:output_type -> rill.admin.v1.RecordEventsResponse
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_rill_admin_v1_telemetry_proto_init() }
func file_rill_admin_v1_telemetry_proto_init() {
	if File_rill_admin_v1_telemetry_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rill_admin_v1_telemetry_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecordEventsRequest); i {
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
		file_rill_admin_v1_telemetry_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecordEventsResponse); i {
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
			RawDescriptor: file_rill_admin_v1_telemetry_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rill_admin_v1_telemetry_proto_goTypes,
		DependencyIndexes: file_rill_admin_v1_telemetry_proto_depIdxs,
		MessageInfos:      file_rill_admin_v1_telemetry_proto_msgTypes,
	}.Build()
	File_rill_admin_v1_telemetry_proto = out.File
	file_rill_admin_v1_telemetry_proto_rawDesc = nil
	file_rill_admin_v1_telemetry_proto_goTypes = nil
	file_rill_admin_v1_telemetry_proto_depIdxs = nil
}
