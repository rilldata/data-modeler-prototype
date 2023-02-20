// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: rill/ui/v1/dashboard.proto

package uiv1

import (
	v1 "github.com/rilldata/rill/rill/runtime/v1"
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

type DashboardState struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TimeStart       *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=time_start,json=timeStart,proto3" json:"time_start,omitempty"`
	TimeEnd         *timestamppb.Timestamp `protobuf:"bytes,2,opt,name=time_end,json=timeEnd,proto3" json:"time_end,omitempty"`
	TimeGranularity v1.TimeGrain           `protobuf:"varint,3,opt,name=time_granularity,json=timeGranularity,proto3,enum=rill.runtime.v1.TimeGrain" json:"time_granularity,omitempty"`
	Filters         *v1.MetricsViewFilter  `protobuf:"bytes,4,opt,name=filters,proto3" json:"filters,omitempty"`
}

func (x *DashboardState) Reset() {
	*x = DashboardState{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_ui_v1_dashboard_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DashboardState) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DashboardState) ProtoMessage() {}

func (x *DashboardState) ProtoReflect() protoreflect.Message {
	mi := &file_rill_ui_v1_dashboard_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DashboardState.ProtoReflect.Descriptor instead.
func (*DashboardState) Descriptor() ([]byte, []int) {
	return file_rill_ui_v1_dashboard_proto_rawDescGZIP(), []int{0}
}

func (x *DashboardState) GetTimeStart() *timestamppb.Timestamp {
	if x != nil {
		return x.TimeStart
	}
	return nil
}

func (x *DashboardState) GetTimeEnd() *timestamppb.Timestamp {
	if x != nil {
		return x.TimeEnd
	}
	return nil
}

func (x *DashboardState) GetTimeGranularity() v1.TimeGrain {
	if x != nil {
		return x.TimeGranularity
	}
	return v1.TimeGrain(0)
}

func (x *DashboardState) GetFilters() *v1.MetricsViewFilter {
	if x != nil {
		return x.Filters
	}
	return nil
}

var File_rill_ui_v1_dashboard_proto protoreflect.FileDescriptor

var file_rill_ui_v1_dashboard_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x75, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x64, 0x61, 0x73,
	0x68, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x72, 0x69,
	0x6c, 0x6c, 0x2e, 0x75, 0x69, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x72, 0x69, 0x6c, 0x6c, 0x2f,
	0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x61, 0x70, 0x69, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x61, 0x74, 0x61, 0x6c, 0x6f, 0x67, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x87, 0x02, 0x0a, 0x0e, 0x44, 0x61, 0x73, 0x68, 0x62, 0x6f, 0x61, 0x72,
	0x64, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x39, 0x0a, 0x0a, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x73,
	0x74, 0x61, 0x72, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x74, 0x69, 0x6d, 0x65, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x12, 0x35, 0x0a, 0x08, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x65, 0x6e, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52,
	0x07, 0x74, 0x69, 0x6d, 0x65, 0x45, 0x6e, 0x64, 0x12, 0x45, 0x0a, 0x10, 0x74, 0x69, 0x6d, 0x65,
	0x5f, 0x67, 0x72, 0x61, 0x6e, 0x75, 0x6c, 0x61, 0x72, 0x69, 0x74, 0x79, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x47, 0x72, 0x61, 0x69, 0x6e, 0x52, 0x0f,
	0x74, 0x69, 0x6d, 0x65, 0x47, 0x72, 0x61, 0x6e, 0x75, 0x6c, 0x61, 0x72, 0x69, 0x74, 0x79, 0x12,
	0x3c, 0x0a, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x22, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e,
	0x76, 0x31, 0x2e, 0x4d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x56, 0x69, 0x65, 0x77, 0x46, 0x69,
	0x6c, 0x74, 0x65, 0x72, 0x52, 0x07, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x42, 0x94, 0x01,
	0x0a, 0x0e, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x75, 0x69, 0x2e, 0x76, 0x31,
	0x42, 0x0e, 0x44, 0x61, 0x73, 0x68, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x50, 0x01, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72,
	0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x69, 0x6c,
	0x6c, 0x2f, 0x75, 0x69, 0x2f, 0x76, 0x31, 0x3b, 0x75, 0x69, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52,
	0x55, 0x58, 0xaa, 0x02, 0x0a, 0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x55, 0x69, 0x2e, 0x56, 0x31, 0xca,
	0x02, 0x0a, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x55, 0x69, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x16, 0x52,
	0x69, 0x6c, 0x6c, 0x5c, 0x55, 0x69, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0c, 0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a, 0x55, 0x69,
	0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_ui_v1_dashboard_proto_rawDescOnce sync.Once
	file_rill_ui_v1_dashboard_proto_rawDescData = file_rill_ui_v1_dashboard_proto_rawDesc
)

func file_rill_ui_v1_dashboard_proto_rawDescGZIP() []byte {
	file_rill_ui_v1_dashboard_proto_rawDescOnce.Do(func() {
		file_rill_ui_v1_dashboard_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_ui_v1_dashboard_proto_rawDescData)
	})
	return file_rill_ui_v1_dashboard_proto_rawDescData
}

var file_rill_ui_v1_dashboard_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_rill_ui_v1_dashboard_proto_goTypes = []interface{}{
	(*DashboardState)(nil),        // 0: rill.ui.v1.DashboardState
	(*timestamppb.Timestamp)(nil), // 1: google.protobuf.Timestamp
	(v1.TimeGrain)(0),             // 2: rill.runtime.v1.TimeGrain
	(*v1.MetricsViewFilter)(nil),  // 3: rill.runtime.v1.MetricsViewFilter
}
var file_rill_ui_v1_dashboard_proto_depIdxs = []int32{
	1, // 0: rill.ui.v1.DashboardState.time_start:type_name -> google.protobuf.Timestamp
	1, // 1: rill.ui.v1.DashboardState.time_end:type_name -> google.protobuf.Timestamp
	2, // 2: rill.ui.v1.DashboardState.time_granularity:type_name -> rill.runtime.v1.TimeGrain
	3, // 3: rill.ui.v1.DashboardState.filters:type_name -> rill.runtime.v1.MetricsViewFilter
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_rill_ui_v1_dashboard_proto_init() }
func file_rill_ui_v1_dashboard_proto_init() {
	if File_rill_ui_v1_dashboard_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rill_ui_v1_dashboard_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DashboardState); i {
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
			RawDescriptor: file_rill_ui_v1_dashboard_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rill_ui_v1_dashboard_proto_goTypes,
		DependencyIndexes: file_rill_ui_v1_dashboard_proto_depIdxs,
		MessageInfos:      file_rill_ui_v1_dashboard_proto_msgTypes,
	}.Build()
	File_rill_ui_v1_dashboard_proto = out.File
	file_rill_ui_v1_dashboard_proto_rawDesc = nil
	file_rill_ui_v1_dashboard_proto_goTypes = nil
	file_rill_ui_v1_dashboard_proto_depIdxs = nil
}
