// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: rill/runtime/v1/schema.proto

package runtimev1

import (
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
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

// Code enumerates all the types that can be represented in a schema
type Type_Code int32

const (
	Type_CODE_UNSPECIFIED Type_Code = 0
	Type_CODE_BOOL        Type_Code = 1
	Type_CODE_INT8        Type_Code = 2
	Type_CODE_INT16       Type_Code = 3
	Type_CODE_INT32       Type_Code = 4
	Type_CODE_INT64       Type_Code = 5
	Type_CODE_INT128      Type_Code = 6
	Type_CODE_UINT8       Type_Code = 7
	Type_CODE_UINT16      Type_Code = 8
	Type_CODE_UINT32      Type_Code = 9
	Type_CODE_UINT64      Type_Code = 10
	Type_CODE_UINT128     Type_Code = 11
	Type_CODE_FLOAT32     Type_Code = 12
	Type_CODE_FLOAT64     Type_Code = 13
	Type_CODE_TIMESTAMP   Type_Code = 14
	Type_CODE_DATE        Type_Code = 15
	Type_CODE_TIME        Type_Code = 16
	Type_CODE_STRING      Type_Code = 17
	Type_CODE_BYTES       Type_Code = 18
	Type_CODE_ARRAY       Type_Code = 19
	Type_CODE_STRUCT      Type_Code = 20
	Type_CODE_MAP         Type_Code = 21
	Type_CODE_DECIMAL     Type_Code = 22
	Type_CODE_JSON        Type_Code = 23
	Type_CODE_UUID        Type_Code = 24
)

// Enum value maps for Type_Code.
var (
	Type_Code_name = map[int32]string{
		0:  "CODE_UNSPECIFIED",
		1:  "CODE_BOOL",
		2:  "CODE_INT8",
		3:  "CODE_INT16",
		4:  "CODE_INT32",
		5:  "CODE_INT64",
		6:  "CODE_INT128",
		7:  "CODE_UINT8",
		8:  "CODE_UINT16",
		9:  "CODE_UINT32",
		10: "CODE_UINT64",
		11: "CODE_UINT128",
		12: "CODE_FLOAT32",
		13: "CODE_FLOAT64",
		14: "CODE_TIMESTAMP",
		15: "CODE_DATE",
		16: "CODE_TIME",
		17: "CODE_STRING",
		18: "CODE_BYTES",
		19: "CODE_ARRAY",
		20: "CODE_STRUCT",
		21: "CODE_MAP",
		22: "CODE_DECIMAL",
		23: "CODE_JSON",
		24: "CODE_UUID",
	}
	Type_Code_value = map[string]int32{
		"CODE_UNSPECIFIED": 0,
		"CODE_BOOL":        1,
		"CODE_INT8":        2,
		"CODE_INT16":       3,
		"CODE_INT32":       4,
		"CODE_INT64":       5,
		"CODE_INT128":      6,
		"CODE_UINT8":       7,
		"CODE_UINT16":      8,
		"CODE_UINT32":      9,
		"CODE_UINT64":      10,
		"CODE_UINT128":     11,
		"CODE_FLOAT32":     12,
		"CODE_FLOAT64":     13,
		"CODE_TIMESTAMP":   14,
		"CODE_DATE":        15,
		"CODE_TIME":        16,
		"CODE_STRING":      17,
		"CODE_BYTES":       18,
		"CODE_ARRAY":       19,
		"CODE_STRUCT":      20,
		"CODE_MAP":         21,
		"CODE_DECIMAL":     22,
		"CODE_JSON":        23,
		"CODE_UUID":        24,
	}
)

func (x Type_Code) Enum() *Type_Code {
	p := new(Type_Code)
	*p = x
	return p
}

func (x Type_Code) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Type_Code) Descriptor() protoreflect.EnumDescriptor {
	return file_rill_runtime_v1_schema_proto_enumTypes[0].Descriptor()
}

func (Type_Code) Type() protoreflect.EnumType {
	return &file_rill_runtime_v1_schema_proto_enumTypes[0]
}

func (x Type_Code) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Type_Code.Descriptor instead.
func (Type_Code) EnumDescriptor() ([]byte, []int) {
	return file_rill_runtime_v1_schema_proto_rawDescGZIP(), []int{0, 0}
}

// Type represents a data type in a schema
type Type struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Code designates the type
	Code Type_Code `protobuf:"varint,1,opt,name=code,proto3,enum=rill.runtime.v1.Type_Code" json:"code,omitempty"`
	// Nullable indicates whether null values are possible
	Nullable bool `protobuf:"varint,2,opt,name=nullable,proto3" json:"nullable,omitempty"`
	// If code is CODE_ARRAY, array_element_type specifies the type of the array elements
	ArrayElementType *Type `protobuf:"bytes,3,opt,name=array_element_type,json=arrayElementType,proto3" json:"array_element_type,omitempty"`
	// If code is CODE_STRUCT, struct_type specifies the type of the struct's fields
	StructType *StructType `protobuf:"bytes,4,opt,name=struct_type,json=structType,proto3" json:"struct_type,omitempty"`
	// If code is CODE_MAP, map_type specifies the map's key and value types
	MapType *MapType `protobuf:"bytes,5,opt,name=map_type,json=mapType,proto3" json:"map_type,omitempty"`
}

func (x *Type) Reset() {
	*x = Type{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_schema_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Type) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Type) ProtoMessage() {}

func (x *Type) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_schema_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Type.ProtoReflect.Descriptor instead.
func (*Type) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_schema_proto_rawDescGZIP(), []int{0}
}

func (x *Type) GetCode() Type_Code {
	if x != nil {
		return x.Code
	}
	return Type_CODE_UNSPECIFIED
}

func (x *Type) GetNullable() bool {
	if x != nil {
		return x.Nullable
	}
	return false
}

func (x *Type) GetArrayElementType() *Type {
	if x != nil {
		return x.ArrayElementType
	}
	return nil
}

func (x *Type) GetStructType() *StructType {
	if x != nil {
		return x.StructType
	}
	return nil
}

func (x *Type) GetMapType() *MapType {
	if x != nil {
		return x.MapType
	}
	return nil
}

// StructType is a type composed of ordered, named and typed sub-fields
type StructType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Fields []*StructType_Field `protobuf:"bytes,1,rep,name=fields,proto3" json:"fields,omitempty"`
}

func (x *StructType) Reset() {
	*x = StructType{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_schema_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StructType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StructType) ProtoMessage() {}

func (x *StructType) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_schema_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StructType.ProtoReflect.Descriptor instead.
func (*StructType) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_schema_proto_rawDescGZIP(), []int{1}
}

func (x *StructType) GetFields() []*StructType_Field {
	if x != nil {
		return x.Fields
	}
	return nil
}

// MapType is a complex type for mapping keys to values
type MapType struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KeyType   *Type `protobuf:"bytes,1,opt,name=key_type,json=keyType,proto3" json:"key_type,omitempty"`
	ValueType *Type `protobuf:"bytes,2,opt,name=value_type,json=valueType,proto3" json:"value_type,omitempty"`
}

func (x *MapType) Reset() {
	*x = MapType{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_schema_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MapType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MapType) ProtoMessage() {}

func (x *MapType) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_schema_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MapType.ProtoReflect.Descriptor instead.
func (*MapType) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_schema_proto_rawDescGZIP(), []int{2}
}

func (x *MapType) GetKeyType() *Type {
	if x != nil {
		return x.KeyType
	}
	return nil
}

func (x *MapType) GetValueType() *Type {
	if x != nil {
		return x.ValueType
	}
	return nil
}

type StructType_Field struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Type *Type  `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
}

func (x *StructType_Field) Reset() {
	*x = StructType_Field{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_runtime_v1_schema_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StructType_Field) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StructType_Field) ProtoMessage() {}

func (x *StructType_Field) ProtoReflect() protoreflect.Message {
	mi := &file_rill_runtime_v1_schema_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StructType_Field.ProtoReflect.Descriptor instead.
func (*StructType_Field) Descriptor() ([]byte, []int) {
	return file_rill_runtime_v1_schema_proto_rawDescGZIP(), []int{1, 0}
}

func (x *StructType_Field) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StructType_Field) GetType() *Type {
	if x != nil {
		return x.Type
	}
	return nil
}

var File_rill_runtime_v1_schema_proto protoreflect.FileDescriptor

var file_rill_runtime_v1_schema_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76,
	0x31, 0x2f, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f,
	0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x1a,
	0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xbd, 0x05, 0x0a, 0x04, 0x54, 0x79, 0x70,
	0x65, 0x12, 0x38, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x1a, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76,
	0x31, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x2e, 0x43, 0x6f, 0x64, 0x65, 0x42, 0x08, 0xfa, 0x42, 0x05,
	0x82, 0x01, 0x02, 0x10, 0x01, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6e,
	0x75, 0x6c, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x6e,
	0x75, 0x6c, 0x6c, 0x61, 0x62, 0x6c, 0x65, 0x12, 0x43, 0x0a, 0x12, 0x61, 0x72, 0x72, 0x61, 0x79,
	0x5f, 0x65, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x10, 0x61, 0x72, 0x72, 0x61,
	0x79, 0x45, 0x6c, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x3c, 0x0a, 0x0b,
	0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x1b, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a,
	0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x33, 0x0a, 0x08, 0x6d, 0x61,
	0x70, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x18, 0x2e, 0x72,
	0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d,
	0x61, 0x70, 0x54, 0x79, 0x70, 0x65, 0x52, 0x07, 0x6d, 0x61, 0x70, 0x54, 0x79, 0x70, 0x65, 0x22,
	0xa6, 0x03, 0x0a, 0x04, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x43, 0x4f, 0x44, 0x45,
	0x5f, 0x55, 0x4e, 0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0d,
	0x0a, 0x09, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x42, 0x4f, 0x4f, 0x4c, 0x10, 0x01, 0x12, 0x0d, 0x0a,
	0x09, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x49, 0x4e, 0x54, 0x38, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a,
	0x43, 0x4f, 0x44, 0x45, 0x5f, 0x49, 0x4e, 0x54, 0x31, 0x36, 0x10, 0x03, 0x12, 0x0e, 0x0a, 0x0a,
	0x43, 0x4f, 0x44, 0x45, 0x5f, 0x49, 0x4e, 0x54, 0x33, 0x32, 0x10, 0x04, 0x12, 0x0e, 0x0a, 0x0a,
	0x43, 0x4f, 0x44, 0x45, 0x5f, 0x49, 0x4e, 0x54, 0x36, 0x34, 0x10, 0x05, 0x12, 0x0f, 0x0a, 0x0b,
	0x43, 0x4f, 0x44, 0x45, 0x5f, 0x49, 0x4e, 0x54, 0x31, 0x32, 0x38, 0x10, 0x06, 0x12, 0x0e, 0x0a,
	0x0a, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x49, 0x4e, 0x54, 0x38, 0x10, 0x07, 0x12, 0x0f, 0x0a,
	0x0b, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x49, 0x4e, 0x54, 0x31, 0x36, 0x10, 0x08, 0x12, 0x0f,
	0x0a, 0x0b, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x49, 0x4e, 0x54, 0x33, 0x32, 0x10, 0x09, 0x12,
	0x0f, 0x0a, 0x0b, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x49, 0x4e, 0x54, 0x36, 0x34, 0x10, 0x0a,
	0x12, 0x10, 0x0a, 0x0c, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x55, 0x49, 0x4e, 0x54, 0x31, 0x32, 0x38,
	0x10, 0x0b, 0x12, 0x10, 0x0a, 0x0c, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x46, 0x4c, 0x4f, 0x41, 0x54,
	0x33, 0x32, 0x10, 0x0c, 0x12, 0x10, 0x0a, 0x0c, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x46, 0x4c, 0x4f,
	0x41, 0x54, 0x36, 0x34, 0x10, 0x0d, 0x12, 0x12, 0x0a, 0x0e, 0x43, 0x4f, 0x44, 0x45, 0x5f, 0x54,
	0x49, 0x4d, 0x45, 0x53, 0x54, 0x41, 0x4d, 0x50, 0x10, 0x0e, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f,
	0x44, 0x45, 0x5f, 0x44, 0x41, 0x54, 0x45, 0x10, 0x0f, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x44,
	0x45, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x10, 0x10, 0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x4f, 0x44, 0x45,
	0x5f, 0x53, 0x54, 0x52, 0x49, 0x4e, 0x47, 0x10, 0x11, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x4f, 0x44,
	0x45, 0x5f, 0x42, 0x59, 0x54, 0x45, 0x53, 0x10, 0x12, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x4f, 0x44,
	0x45, 0x5f, 0x41, 0x52, 0x52, 0x41, 0x59, 0x10, 0x13, 0x12, 0x0f, 0x0a, 0x0b, 0x43, 0x4f, 0x44,
	0x45, 0x5f, 0x53, 0x54, 0x52, 0x55, 0x43, 0x54, 0x10, 0x14, 0x12, 0x0c, 0x0a, 0x08, 0x43, 0x4f,
	0x44, 0x45, 0x5f, 0x4d, 0x41, 0x50, 0x10, 0x15, 0x12, 0x10, 0x0a, 0x0c, 0x43, 0x4f, 0x44, 0x45,
	0x5f, 0x44, 0x45, 0x43, 0x49, 0x4d, 0x41, 0x4c, 0x10, 0x16, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f,
	0x44, 0x45, 0x5f, 0x4a, 0x53, 0x4f, 0x4e, 0x10, 0x17, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x44,
	0x45, 0x5f, 0x55, 0x55, 0x49, 0x44, 0x10, 0x18, 0x22, 0x8f, 0x01, 0x0a, 0x0a, 0x53, 0x74, 0x72,
	0x75, 0x63, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x39, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x21, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72,
	0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c,
	0x64, 0x73, 0x1a, 0x46, 0x0a, 0x05, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x29, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e,
	0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22, 0x71, 0x0a, 0x07, 0x4d, 0x61,
	0x70, 0x54, 0x79, 0x70, 0x65, 0x12, 0x30, 0x0a, 0x08, 0x6b, 0x65, 0x79, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72,
	0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x79, 0x70, 0x65, 0x52, 0x07,
	0x6b, 0x65, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x34, 0x0a, 0x0a, 0x76, 0x61, 0x6c, 0x75, 0x65,
	0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x72, 0x69,
	0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x09, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x54, 0x79, 0x70, 0x65, 0x42, 0xb4, 0x01,
	0x0a, 0x13, 0x63, 0x6f, 0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x72, 0x75, 0x6e, 0x74, 0x69,
	0x6d, 0x65, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x50, 0x72, 0x6f,
	0x74, 0x6f, 0x50, 0x01, 0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x72,
	0x69, 0x6c, 0x6c, 0x2f, 0x72, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2f, 0x76, 0x31, 0x3b, 0x72,
	0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x52, 0x58, 0xaa, 0x02,
	0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x2e, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x2e, 0x56, 0x31,
	0xca, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65, 0x5c,
	0x56, 0x31, 0xe2, 0x02, 0x1b, 0x52, 0x69, 0x6c, 0x6c, 0x5c, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d,
	0x65, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0xea, 0x02, 0x11, 0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a, 0x52, 0x75, 0x6e, 0x74, 0x69, 0x6d, 0x65,
	0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_rill_runtime_v1_schema_proto_rawDescOnce sync.Once
	file_rill_runtime_v1_schema_proto_rawDescData = file_rill_runtime_v1_schema_proto_rawDesc
)

func file_rill_runtime_v1_schema_proto_rawDescGZIP() []byte {
	file_rill_runtime_v1_schema_proto_rawDescOnce.Do(func() {
		file_rill_runtime_v1_schema_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_runtime_v1_schema_proto_rawDescData)
	})
	return file_rill_runtime_v1_schema_proto_rawDescData
}

var file_rill_runtime_v1_schema_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_rill_runtime_v1_schema_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_rill_runtime_v1_schema_proto_goTypes = []interface{}{
	(Type_Code)(0),           // 0: rill.runtime.v1.Type.Code
	(*Type)(nil),             // 1: rill.runtime.v1.Type
	(*StructType)(nil),       // 2: rill.runtime.v1.StructType
	(*MapType)(nil),          // 3: rill.runtime.v1.MapType
	(*StructType_Field)(nil), // 4: rill.runtime.v1.StructType.Field
}
var file_rill_runtime_v1_schema_proto_depIdxs = []int32{
	0, // 0: rill.runtime.v1.Type.code:type_name -> rill.runtime.v1.Type.Code
	1, // 1: rill.runtime.v1.Type.array_element_type:type_name -> rill.runtime.v1.Type
	2, // 2: rill.runtime.v1.Type.struct_type:type_name -> rill.runtime.v1.StructType
	3, // 3: rill.runtime.v1.Type.map_type:type_name -> rill.runtime.v1.MapType
	4, // 4: rill.runtime.v1.StructType.fields:type_name -> rill.runtime.v1.StructType.Field
	1, // 5: rill.runtime.v1.MapType.key_type:type_name -> rill.runtime.v1.Type
	1, // 6: rill.runtime.v1.MapType.value_type:type_name -> rill.runtime.v1.Type
	1, // 7: rill.runtime.v1.StructType.Field.type:type_name -> rill.runtime.v1.Type
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_rill_runtime_v1_schema_proto_init() }
func file_rill_runtime_v1_schema_proto_init() {
	if File_rill_runtime_v1_schema_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rill_runtime_v1_schema_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Type); i {
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
		file_rill_runtime_v1_schema_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StructType); i {
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
		file_rill_runtime_v1_schema_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MapType); i {
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
		file_rill_runtime_v1_schema_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StructType_Field); i {
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
			RawDescriptor: file_rill_runtime_v1_schema_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_rill_runtime_v1_schema_proto_goTypes,
		DependencyIndexes: file_rill_runtime_v1_schema_proto_depIdxs,
		EnumInfos:         file_rill_runtime_v1_schema_proto_enumTypes,
		MessageInfos:      file_rill_runtime_v1_schema_proto_msgTypes,
	}.Build()
	File_rill_runtime_v1_schema_proto = out.File
	file_rill_runtime_v1_schema_proto_rawDesc = nil
	file_rill_runtime_v1_schema_proto_goTypes = nil
	file_rill_runtime_v1_schema_proto_depIdxs = nil
}
