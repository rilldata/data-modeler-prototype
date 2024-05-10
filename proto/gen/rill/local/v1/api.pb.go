// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        (unknown)
// source: rill/local/v1/api.proto

package localv1

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

type PingRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *PingRequest) Reset() {
	*x = PingRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_local_v1_api_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingRequest) ProtoMessage() {}

func (x *PingRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rill_local_v1_api_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingRequest.ProtoReflect.Descriptor instead.
func (*PingRequest) Descriptor() ([]byte, []int) {
	return file_rill_local_v1_api_proto_rawDescGZIP(), []int{0}
}

type PingResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Time *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *PingResponse) Reset() {
	*x = PingResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_local_v1_api_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PingResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingResponse) ProtoMessage() {}

func (x *PingResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rill_local_v1_api_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingResponse.ProtoReflect.Descriptor instead.
func (*PingResponse) Descriptor() ([]byte, []int) {
	return file_rill_local_v1_api_proto_rawDescGZIP(), []int{1}
}

func (x *PingResponse) GetTime() *timestamppb.Timestamp {
	if x != nil {
		return x.Time
	}
	return nil
}

type GetMetadataRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetMetadataRequest) Reset() {
	*x = GetMetadataRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_local_v1_api_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMetadataRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMetadataRequest) ProtoMessage() {}

func (x *GetMetadataRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rill_local_v1_api_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMetadataRequest.ProtoReflect.Descriptor instead.
func (*GetMetadataRequest) Descriptor() ([]byte, []int) {
	return file_rill_local_v1_api_proto_rawDescGZIP(), []int{2}
}

type GetMetadataResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InstanceId       string `protobuf:"bytes,1,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	ProjectPath      string `protobuf:"bytes,4,opt,name=project_path,json=projectPath,proto3" json:"project_path,omitempty"`
	InstallId        string `protobuf:"bytes,2,opt,name=install_id,json=installId,proto3" json:"install_id,omitempty"`
	UserId           string `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Version          string `protobuf:"bytes,5,opt,name=version,proto3" json:"version,omitempty"`
	BuildCommit      string `protobuf:"bytes,6,opt,name=build_commit,json=buildCommit,proto3" json:"build_commit,omitempty"`
	BuildTime        string `protobuf:"bytes,7,opt,name=build_time,json=buildTime,proto3" json:"build_time,omitempty"`
	IsDev            bool   `protobuf:"varint,8,opt,name=is_dev,json=isDev,proto3" json:"is_dev,omitempty"`
	AnalyticsEnabled bool   `protobuf:"varint,9,opt,name=analytics_enabled,json=analyticsEnabled,proto3" json:"analytics_enabled,omitempty"`
	Readonly         bool   `protobuf:"varint,10,opt,name=readonly,proto3" json:"readonly,omitempty"`
	GrpcPort         int32  `protobuf:"varint,11,opt,name=grpc_port,json=grpcPort,proto3" json:"grpc_port,omitempty"`
}

func (x *GetMetadataResponse) Reset() {
	*x = GetMetadataResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_local_v1_api_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetMetadataResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetMetadataResponse) ProtoMessage() {}

func (x *GetMetadataResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rill_local_v1_api_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetMetadataResponse.ProtoReflect.Descriptor instead.
func (*GetMetadataResponse) Descriptor() ([]byte, []int) {
	return file_rill_local_v1_api_proto_rawDescGZIP(), []int{3}
}

func (x *GetMetadataResponse) GetInstanceId() string {
	if x != nil {
		return x.InstanceId
	}
	return ""
}

func (x *GetMetadataResponse) GetProjectPath() string {
	if x != nil {
		return x.ProjectPath
	}
	return ""
}

func (x *GetMetadataResponse) GetInstallId() string {
	if x != nil {
		return x.InstallId
	}
	return ""
}

func (x *GetMetadataResponse) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

func (x *GetMetadataResponse) GetVersion() string {
	if x != nil {
		return x.Version
	}
	return ""
}

func (x *GetMetadataResponse) GetBuildCommit() string {
	if x != nil {
		return x.BuildCommit
	}
	return ""
}

func (x *GetMetadataResponse) GetBuildTime() string {
	if x != nil {
		return x.BuildTime
	}
	return ""
}

func (x *GetMetadataResponse) GetIsDev() bool {
	if x != nil {
		return x.IsDev
	}
	return false
}

func (x *GetMetadataResponse) GetAnalyticsEnabled() bool {
	if x != nil {
		return x.AnalyticsEnabled
	}
	return false
}

func (x *GetMetadataResponse) GetReadonly() bool {
	if x != nil {
		return x.Readonly
	}
	return false
}

func (x *GetMetadataResponse) GetGrpcPort() int32 {
	if x != nil {
		return x.GrpcPort
	}
	return 0
}

type GetVersionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetVersionRequest) Reset() {
	*x = GetVersionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_local_v1_api_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetVersionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetVersionRequest) ProtoMessage() {}

func (x *GetVersionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_rill_local_v1_api_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetVersionRequest.ProtoReflect.Descriptor instead.
func (*GetVersionRequest) Descriptor() ([]byte, []int) {
	return file_rill_local_v1_api_proto_rawDescGZIP(), []int{4}
}

type GetVersionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Current string `protobuf:"bytes,1,opt,name=current,proto3" json:"current,omitempty"`
	Latest  string `protobuf:"bytes,2,opt,name=latest,proto3" json:"latest,omitempty"`
}

func (x *GetVersionResponse) Reset() {
	*x = GetVersionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_rill_local_v1_api_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetVersionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetVersionResponse) ProtoMessage() {}

func (x *GetVersionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_rill_local_v1_api_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetVersionResponse.ProtoReflect.Descriptor instead.
func (*GetVersionResponse) Descriptor() ([]byte, []int) {
	return file_rill_local_v1_api_proto_rawDescGZIP(), []int{5}
}

func (x *GetVersionResponse) GetCurrent() string {
	if x != nil {
		return x.Current
	}
	return ""
}

func (x *GetVersionResponse) GetLatest() string {
	if x != nil {
		return x.Latest
	}
	return ""
}

var File_rill_local_v1_api_proto protoreflect.FileDescriptor

var file_rill_local_v1_api_proto_rawDesc = []byte{
	0x0a, 0x17, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x2f,
	0x61, 0x70, 0x69, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x72, 0x69, 0x6c, 0x6c, 0x2e,
	0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x0d, 0x0a, 0x0b, 0x50, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x3e, 0x0a, 0x0c, 0x50, 0x69, 0x6e, 0x67,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61,
	0x6d, 0x70, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x22, 0x14, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0xea,
	0x02, 0x0a, 0x13, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x49, 0x64, 0x12, 0x21, 0x0a, 0x0c, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x5f, 0x70, 0x61, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x70,
	0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74, 0x50, 0x61, 0x74, 0x68, 0x12, 0x1d, 0x0a, 0x0a, 0x69, 0x6e,
	0x73, 0x74, 0x61, 0x6c, 0x6c, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x69, 0x6e, 0x73, 0x74, 0x61, 0x6c, 0x6c, 0x49, 0x64, 0x12, 0x17, 0x0a, 0x07, 0x75, 0x73, 0x65,
	0x72, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x75, 0x73, 0x65, 0x72,
	0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x76, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x0c,
	0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x12,
	0x1d, 0x0a, 0x0a, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x62, 0x75, 0x69, 0x6c, 0x64, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x15,
	0x0a, 0x06, 0x69, 0x73, 0x5f, 0x64, 0x65, 0x76, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05,
	0x69, 0x73, 0x44, 0x65, 0x76, 0x12, 0x2b, 0x0a, 0x11, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69,
	0x63, 0x73, 0x5f, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x10, 0x61, 0x6e, 0x61, 0x6c, 0x79, 0x74, 0x69, 0x63, 0x73, 0x45, 0x6e, 0x61, 0x62, 0x6c,
	0x65, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x61, 0x64, 0x6f, 0x6e, 0x6c, 0x79, 0x18, 0x0a,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x72, 0x65, 0x61, 0x64, 0x6f, 0x6e, 0x6c, 0x79, 0x12, 0x1b,
	0x0a, 0x09, 0x67, 0x72, 0x70, 0x63, 0x5f, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x08, 0x67, 0x72, 0x70, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x22, 0x13, 0x0a, 0x11, 0x47,
	0x65, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0x46, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e,
	0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x6c, 0x61, 0x74, 0x65, 0x73, 0x74, 0x32, 0xfe, 0x01, 0x0a, 0x0c, 0x4c, 0x6f, 0x63,
	0x61, 0x6c, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41, 0x0a, 0x04, 0x50, 0x69, 0x6e,
	0x67, 0x12, 0x1a, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x76,
	0x31, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e,
	0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x50, 0x69,
	0x6e, 0x67, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x56, 0x0a, 0x0b,
	0x47, 0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x21, 0x2e, 0x72, 0x69,
	0x6c, 0x6c, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x22,
	0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47,
	0x65, 0x74, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x53, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x20, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e,
	0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x21, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x6c, 0x6f, 0x63, 0x61,
	0x6c, 0x2e, 0x76, 0x31, 0x2e, 0x47, 0x65, 0x74, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0xad, 0x01, 0x0a, 0x11, 0x63, 0x6f,
	0x6d, 0x2e, 0x72, 0x69, 0x6c, 0x6c, 0x2e, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x76, 0x31, 0x42,
	0x08, 0x41, 0x70, 0x69, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x38, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x64, 0x61, 0x74, 0x61,
	0x2f, 0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x65, 0x6e, 0x2f,
	0x72, 0x69, 0x6c, 0x6c, 0x2f, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x2f, 0x76, 0x31, 0x3b, 0x6c, 0x6f,
	0x63, 0x61, 0x6c, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x52, 0x4c, 0x58, 0xaa, 0x02, 0x0d, 0x52, 0x69,
	0x6c, 0x6c, 0x2e, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x0d, 0x52, 0x69,
	0x6c, 0x6c, 0x5c, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x19, 0x52, 0x69,
	0x6c, 0x6c, 0x5c, 0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0f, 0x52, 0x69, 0x6c, 0x6c, 0x3a, 0x3a,
	0x4c, 0x6f, 0x63, 0x61, 0x6c, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_rill_local_v1_api_proto_rawDescOnce sync.Once
	file_rill_local_v1_api_proto_rawDescData = file_rill_local_v1_api_proto_rawDesc
)

func file_rill_local_v1_api_proto_rawDescGZIP() []byte {
	file_rill_local_v1_api_proto_rawDescOnce.Do(func() {
		file_rill_local_v1_api_proto_rawDescData = protoimpl.X.CompressGZIP(file_rill_local_v1_api_proto_rawDescData)
	})
	return file_rill_local_v1_api_proto_rawDescData
}

var file_rill_local_v1_api_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_rill_local_v1_api_proto_goTypes = []interface{}{
	(*PingRequest)(nil),           // 0: rill.local.v1.PingRequest
	(*PingResponse)(nil),          // 1: rill.local.v1.PingResponse
	(*GetMetadataRequest)(nil),    // 2: rill.local.v1.GetMetadataRequest
	(*GetMetadataResponse)(nil),   // 3: rill.local.v1.GetMetadataResponse
	(*GetVersionRequest)(nil),     // 4: rill.local.v1.GetVersionRequest
	(*GetVersionResponse)(nil),    // 5: rill.local.v1.GetVersionResponse
	(*timestamppb.Timestamp)(nil), // 6: google.protobuf.Timestamp
}
var file_rill_local_v1_api_proto_depIdxs = []int32{
	6, // 0: rill.local.v1.PingResponse.time:type_name -> google.protobuf.Timestamp
	0, // 1: rill.local.v1.LocalService.Ping:input_type -> rill.local.v1.PingRequest
	2, // 2: rill.local.v1.LocalService.GetMetadata:input_type -> rill.local.v1.GetMetadataRequest
	4, // 3: rill.local.v1.LocalService.GetVersion:input_type -> rill.local.v1.GetVersionRequest
	1, // 4: rill.local.v1.LocalService.Ping:output_type -> rill.local.v1.PingResponse
	3, // 5: rill.local.v1.LocalService.GetMetadata:output_type -> rill.local.v1.GetMetadataResponse
	5, // 6: rill.local.v1.LocalService.GetVersion:output_type -> rill.local.v1.GetVersionResponse
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_rill_local_v1_api_proto_init() }
func file_rill_local_v1_api_proto_init() {
	if File_rill_local_v1_api_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_rill_local_v1_api_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingRequest); i {
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
		file_rill_local_v1_api_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PingResponse); i {
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
		file_rill_local_v1_api_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMetadataRequest); i {
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
		file_rill_local_v1_api_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetMetadataResponse); i {
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
		file_rill_local_v1_api_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetVersionRequest); i {
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
		file_rill_local_v1_api_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetVersionResponse); i {
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
			RawDescriptor: file_rill_local_v1_api_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_rill_local_v1_api_proto_goTypes,
		DependencyIndexes: file_rill_local_v1_api_proto_depIdxs,
		MessageInfos:      file_rill_local_v1_api_proto_msgTypes,
	}.Build()
	File_rill_local_v1_api_proto = out.File
	file_rill_local_v1_api_proto_rawDesc = nil
	file_rill_local_v1_api_proto_goTypes = nil
	file_rill_local_v1_api_proto_depIdxs = nil
}
