// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: rill/runtime/v1/time_grain.proto

package runtimev1

import (
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

type TimeGrain int32

const (
	TimeGrain_TIME_GRAIN_UNSPECIFIED TimeGrain = 0
	TimeGrain_TIME_GRAIN_MILLISECOND TimeGrain = 1
	TimeGrain_TIME_GRAIN_SECOND      TimeGrain = 2
	TimeGrain_TIME_GRAIN_MINUTE      TimeGrain = 3
	TimeGrain_TIME_GRAIN_HOUR        TimeGrain = 4
	TimeGrain_TIME_GRAIN_DAY         TimeGrain = 5
	TimeGrain_TIME_GRAIN_WEEK        TimeGrain = 6
	TimeGrain_TIME_GRAIN_MONTH       TimeGrain = 7
	TimeGrain_TIME_GRAIN_QUARTER     TimeGrain = 8
	TimeGrain_TIME_GRAIN_YEAR        TimeGrain = 9
)

// Enum value maps for TimeGrain.
var (
	TimeGrain_name = map[int32]string{
		0: "TIME_GRAIN_UNSPECIFIED",
		1: "TIME_GRAIN_MILLISECOND",
		2: "TIME_GRAIN_SECOND",
		3: "TIME_GRAIN_MINUTE",
		4: "TIME_GRAIN_HOUR",
		5: "TIME_GRAIN_DAY",
		6: "TIME_GRAIN_WEEK",
		7: "TIME_GRAIN_MONTH",
		8: "TIME_GRAIN_QUARTER",
		9: "TIME_GRAIN_YEAR",
	}
	TimeGrain_value = map[string]int32{
		"TIME_GRAIN_UNSPECIFIED": 0,
		"TIME_GRAIN_MILLISECOND": 1,
		"TIME_GRAIN_SECOND":      2,
		"TIME_GRAIN_MINUTE":      3,
		"TIME_GRAIN_HOUR":        4,
		"TIME_GRAIN_DAY":         5,
		"TIME_GRAIN_WEEK":        6,
		"TIME_GRAIN_MONTH":       7,
		"TIME_GRAIN_QUARTER":     8,
		"TIME_GRAIN_YEAR":        9,
	}
)

func (x TimeGrain) Enum() *TimeGrain {
	p := new(TimeGrain)
	*p = x
	return p
}

func (x TimeGrain) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TimeGrain) Descriptor() protoreflect.EnumDescriptor {
	return file_rill_runtime_v1_time_grain_proto_enumTypes[0].Descriptor()
}

func (TimeGrain) Type() protoreflect.EnumType {
	return &file_rill_runtime_v1_time_grain_proto_enumTypes[0]
}

func (x TimeGrain) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TimeGrain.Descriptor instead.
func (TimeGrain) EnumDescriptor() ([]byte, []int) {
	return file_rill_runtime_v1_time_grain_proto_rawDescGZIP(), []int{0}
}

var File_rill_runtime_v1_time_grain_proto protoreflect.FileDescriptor

var file_rill_runtime_v1_time_grain_proto_rawDesc = []byte{
	0x0a, 0x20, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76,
	0x31, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x67, 0x72, 0x61, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x0f, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x2e, 0x76, 0x31, 0x2a, 0xf2, 0x01, 0x0a, 0x09, 0x54, 0x69, 0x6d, 0x65, 0x47, 0x72, 0x61, 0x69,
	0x6e, 0x12, 0x1a, 0x0a, 0x16, 0x54, 0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f,
	0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x1a, 0x0a,
	0x16, 0x54, 0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f, 0x4d, 0x49, 0x4c, 0x4c,
	0x49, 0x53, 0x45, 0x43, 0x4f, 0x4e, 0x44, 0x10, 0x01, 0x12, 0x15, 0x0a, 0x11, 0x54, 0x49, 0x4d,
	0x45, 0x5f, 0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f, 0x53, 0x45, 0x43, 0x4f, 0x4e, 0x44, 0x10, 0x02,
	0x12, 0x15, 0x0a, 0x11, 0x54, 0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f, 0x4d,
	0x49, 0x4e, 0x55, 0x54, 0x45, 0x10, 0x03, 0x12, 0x13, 0x0a, 0x0f, 0x54, 0x49, 0x4d, 0x45, 0x5f,
	0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f, 0x48, 0x4f, 0x55, 0x52, 0x10, 0x04, 0x12, 0x12, 0x0a, 0x0e,
	0x54, 0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f, 0x44, 0x41, 0x59, 0x10, 0x05,
	0x12, 0x13, 0x0a, 0x0f, 0x54, 0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f, 0x57,
	0x45, 0x45, 0x4b, 0x10, 0x06, 0x12, 0x14, 0x0a, 0x10, 0x54, 0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52,
	0x41, 0x49, 0x4e, 0x5f, 0x4d, 0x4f, 0x4e, 0x54, 0x48, 0x10, 0x07, 0x12, 0x16, 0x0a, 0x12, 0x54,
	0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52, 0x41, 0x49, 0x4e, 0x5f, 0x51, 0x55, 0x41, 0x52, 0x54, 0x45,
	0x52, 0x10, 0x08, 0x12, 0x13, 0x0a, 0x0f, 0x54, 0x49, 0x4d, 0x45, 0x5f, 0x47, 0x52, 0x41, 0x49,
	0x4e, 0x5f, 0x59, 0x45, 0x41, 0x52, 0x10, 0x09, 0x42, 0xc1, 0x01, 0x0a, 0x13, 0x63, 0x6f, 0x6d,
	0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31,
	0x42, 0x0e, 0x54, 0x69, 0x6d, 0x65, 0x47, 0x72, 0x61, 0x69, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f,
	0x50, 0x01, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72,
	0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74,
	0x69, 0x6d, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x76, 0x31,
	0xa2, 0x02, 0x03, 0x52, 0x52, 0x58, 0xaa, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x52, 0x75,
	0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x5c,
	0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x1b, 0x52, 0x69, 0x6c,
	0x6c, 0x5c, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42,
	0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x11, 0x52, 0x69, 0x6c, 0x6c, 0x3a,
	0x3a, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_runtime_v1_time_grain_proto_rawDescOnce sync.Once
	file_rill_runtime_v1_time_grain_proto_rawDescData = file_rill_runtime_v1_time_grain_proto_rawDesc
)

func file_rill_runtime_v1_time_grain_proto_rawDescGZIP() []byte {
	file_rill_runtime_v1_time_grain_proto_rawDescOnce.Do(func() {
		file_rill_runtime_v1_time_grain_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_runtime_v1_time_grain_proto_rawDescData)
	})
	return file_rill_runtime_v1_time_grain_proto_rawDescData
}

var file_rill_runtime_v1_time_grain_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_rill_runtime_v1_time_grain_proto_goTypes = []interface{}{
	(TimeGrain)(0), // 0: rill.runtime.v1.TimeGrain
}
var file_rill_runtime_v1_time_grain_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_rill_runtime_v1_time_grain_proto_init() }
func file_rill_runtime_v1_time_grain_proto_init() {
	if File_rill_runtime_v1_time_grain_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rill_runtime_v1_time_grain_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rill_runtime_v1_time_grain_proto_goTypes,
		DependencyIndexes: file_rill_runtime_v1_time_grain_proto_depIdxs,
		EnumInfos:         file_rill_runtime_v1_time_grain_proto_enumTypes,
	}.Build()
	File_rill_runtime_v1_time_grain_proto = out.File
	file_rill_runtime_v1_time_grain_proto_rawDesc = nil
	file_rill_runtime_v1_time_grain_proto_goTypes = nil
	file_rill_runtime_v1_time_grain_proto_depIdxs = nil
}
