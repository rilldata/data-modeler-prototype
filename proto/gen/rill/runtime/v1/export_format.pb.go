// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        (unknown)
// source: rill/runtime/v1/export_format.proto

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

type ExportFormat int32

const (
	ExportFormat_EXPORT_FORMAT_UNSPECIFIED ExportFormat = 0
	ExportFormat_EXPORT_FORMAT_CSV         ExportFormat = 1
	ExportFormat_EXPORT_FORMAT_XLSX        ExportFormat = 2
	ExportFormat_EXPORT_FORMAT_PARQUET     ExportFormat = 3
)

// Enum value maps for ExportFormat.
var (
	ExportFormat_name = map[int32]string{
		0: "EXPORT_FORMAT_UNSPECIFIED",
		1: "EXPORT_FORMAT_CSV",
		2: "EXPORT_FORMAT_XLSX",
		3: "EXPORT_FORMAT_PARQUET",
	}
	ExportFormat_value = map[string]int32{
		"EXPORT_FORMAT_UNSPECIFIED": 0,
		"EXPORT_FORMAT_CSV":         1,
		"EXPORT_FORMAT_XLSX":        2,
		"EXPORT_FORMAT_PARQUET":     3,
	}
)

func (x ExportFormat) Enum() *ExportFormat {
	p := new(ExportFormat)
	*p = x
	return p
}

func (x ExportFormat) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ExportFormat) Descriptor() protoreflect.EnumDescriptor {
	return file_rill_runtime_v1_export_format_proto_enumTypes[0].Descriptor()
}

func (ExportFormat) Type() protoreflect.EnumType {
	return &file_rill_runtime_v1_export_format_proto_enumTypes[0]
}

func (x ExportFormat) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ExportFormat.Descriptor instead.
func (ExportFormat) EnumDescriptor() ([]byte, []int) {
	return file_rill_runtime_v1_export_format_proto_rawDescGZIP(), []int{0}
}

var File_rill_runtime_v1_export_format_proto protoreflect.FileDescriptor

var file_rill_runtime_v1_export_format_proto_rawDesc = []byte{
	0x0a, 0x23, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76,
	0x31, 0x2f, 0x65, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x5f, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74,
	0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2a, 0x77, 0x0a, 0x0c, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74,
	0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x1d, 0x0a, 0x19, 0x45, 0x58, 0x50, 0x4f, 0x52, 0x54,
	0x5f, 0x46, 0x4f, 0x52, 0x4d, 0x41, 0x54, 0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46,
	0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x45, 0x58, 0x50, 0x4f, 0x52, 0x54, 0x5f,
	0x46, 0x4f, 0x52, 0x4d, 0x41, 0x54, 0x5f, 0x43, 0x53, 0x56, 0x10, 0x01, 0x12, 0x16, 0x0a, 0x12,
	0x45, 0x58, 0x50, 0x4f, 0x52, 0x54, 0x5f, 0x46, 0x4f, 0x52, 0x4d, 0x41, 0x54, 0x5f, 0x58, 0x4c,
	0x53, 0x58, 0x10, 0x02, 0x12, 0x19, 0x0a, 0x15, 0x45, 0x58, 0x50, 0x4f, 0x52, 0x54, 0x5f, 0x46,
	0x4f, 0x52, 0x4d, 0x41, 0x54, 0x5f, 0x50, 0x41, 0x52, 0x51, 0x55, 0x45, 0x54, 0x10, 0x03, 0x42,
	0xc4, 0x01, 0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e,
	0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x42, 0x11, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x46,
	0x6f, 0x72, 0x6d, 0x61, 0x74, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3c, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74,
	0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e,
	0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76, 0x31,
	0x3b, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x52, 0x58,
	0xaa, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e,
	0x56, 0x31, 0xca, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x1b, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x52, 0x75, 0x6e, 0x74,
	0x69, 0x6d, 0x65, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x11, 0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a, 0x52, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_runtime_v1_export_format_proto_rawDescOnce sync.Once
	file_rill_runtime_v1_export_format_proto_rawDescData = file_rill_runtime_v1_export_format_proto_rawDesc
)

func file_rill_runtime_v1_export_format_proto_rawDescGZIP() []byte {
	file_rill_runtime_v1_export_format_proto_rawDescOnce.Do(func() {
		file_rill_runtime_v1_export_format_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_runtime_v1_export_format_proto_rawDescData)
	})
	return file_rill_runtime_v1_export_format_proto_rawDescData
}

var file_rill_runtime_v1_export_format_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_rill_runtime_v1_export_format_proto_goTypes = []any{
	(ExportFormat)(0), // 0: rill.runtime.v1.ExportFormat
}
var file_rill_runtime_v1_export_format_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_rill_runtime_v1_export_format_proto_init() }
func file_rill_runtime_v1_export_format_proto_init() {
	if File_rill_runtime_v1_export_format_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_rill_runtime_v1_export_format_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rill_runtime_v1_export_format_proto_goTypes,
		DependencyIndexes: file_rill_runtime_v1_export_format_proto_depIdxs,
		EnumInfos:         file_rill_runtime_v1_export_format_proto_enumTypes,
	}.Build()
	File_rill_runtime_v1_export_format_proto = out.File
	file_rill_runtime_v1_export_format_proto_rawDesc = nil
	file_rill_runtime_v1_export_format_proto_goTypes = nil
	file_rill_runtime_v1_export_format_proto_depIdxs = nil
}
