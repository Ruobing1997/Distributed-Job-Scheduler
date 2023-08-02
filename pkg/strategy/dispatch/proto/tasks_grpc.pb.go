// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.3
// source: proto/tasks_grpc.proto

package task_grpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
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

type Payload struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Format int32  `protobuf:"varint,1,opt,name=format,proto3" json:"format,omitempty"`
	Script string `protobuf:"bytes,2,opt,name=script,proto3" json:"script,omitempty"`
}

func (x *Payload) Reset() {
	*x = Payload{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tasks_grpc_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Payload) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Payload) ProtoMessage() {}

func (x *Payload) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tasks_grpc_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Payload.ProtoReflect.Descriptor instead.
func (*Payload) Descriptor() ([]byte, []int) {
	return file_proto_tasks_grpc_proto_rawDescGZIP(), []int{0}
}

func (x *Payload) GetFormat() int32 {
	if x != nil {
		return x.Format
	}
	return 0
}

func (x *Payload) GetScript() string {
	if x != nil {
		return x.Script
	}
	return ""
}

type TaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Payload       *Payload               `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	ExecutionTime *timestamppb.Timestamp `protobuf:"bytes,3,opt,name=execution_time,json=executionTime,proto3" json:"execution_time,omitempty"`
	MaxRetryCount int32                  `protobuf:"varint,4,opt,name=max_retry_count,json=maxRetryCount,proto3" json:"max_retry_count,omitempty"`
}

func (x *TaskRequest) Reset() {
	*x = TaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tasks_grpc_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskRequest) ProtoMessage() {}

func (x *TaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tasks_grpc_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskRequest.ProtoReflect.Descriptor instead.
func (*TaskRequest) Descriptor() ([]byte, []int) {
	return file_proto_tasks_grpc_proto_rawDescGZIP(), []int{1}
}

func (x *TaskRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TaskRequest) GetPayload() *Payload {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *TaskRequest) GetExecutionTime() *timestamppb.Timestamp {
	if x != nil {
		return x.ExecutionTime
	}
	return nil
}

func (x *TaskRequest) GetMaxRetryCount() int32 {
	if x != nil {
		return x.MaxRetryCount
	}
	return 0
}

type TaskResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id     string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Status int32  `protobuf:"varint,2,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *TaskResponse) Reset() {
	*x = TaskResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tasks_grpc_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TaskResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TaskResponse) ProtoMessage() {}

func (x *TaskResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tasks_grpc_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TaskResponse.ProtoReflect.Descriptor instead.
func (*TaskResponse) Descriptor() ([]byte, []int) {
	return file_proto_tasks_grpc_proto_rawDescGZIP(), []int{2}
}

func (x *TaskResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TaskResponse) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

type NotifyMessageRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WorkerId string `protobuf:"bytes,1,opt,name=workerId,proto3" json:"workerId,omitempty"`
	TaskId   string `protobuf:"bytes,2,opt,name=taskId,proto3" json:"taskId,omitempty"`
	Status   int32  `protobuf:"varint,3,opt,name=status,proto3" json:"status,omitempty"`
}

func (x *NotifyMessageRequest) Reset() {
	*x = NotifyMessageRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tasks_grpc_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyMessageRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyMessageRequest) ProtoMessage() {}

func (x *NotifyMessageRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tasks_grpc_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyMessageRequest.ProtoReflect.Descriptor instead.
func (*NotifyMessageRequest) Descriptor() ([]byte, []int) {
	return file_proto_tasks_grpc_proto_rawDescGZIP(), []int{3}
}

func (x *NotifyMessageRequest) GetWorkerId() string {
	if x != nil {
		return x.WorkerId
	}
	return ""
}

func (x *NotifyMessageRequest) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *NotifyMessageRequest) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

type NotifyMessageResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *NotifyMessageResponse) Reset() {
	*x = NotifyMessageResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tasks_grpc_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyMessageResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyMessageResponse) ProtoMessage() {}

func (x *NotifyMessageResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tasks_grpc_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyMessageResponse.ProtoReflect.Descriptor instead.
func (*NotifyMessageResponse) Descriptor() ([]byte, []int) {
	return file_proto_tasks_grpc_proto_rawDescGZIP(), []int{4}
}

func (x *NotifyMessageResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type RenewLeaseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id            string               `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	LeaseDuration *durationpb.Duration `protobuf:"bytes,2,opt,name=lease_duration,json=leaseDuration,proto3" json:"lease_duration,omitempty"`
}

func (x *RenewLeaseRequest) Reset() {
	*x = RenewLeaseRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tasks_grpc_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RenewLeaseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RenewLeaseRequest) ProtoMessage() {}

func (x *RenewLeaseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tasks_grpc_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RenewLeaseRequest.ProtoReflect.Descriptor instead.
func (*RenewLeaseRequest) Descriptor() ([]byte, []int) {
	return file_proto_tasks_grpc_proto_rawDescGZIP(), []int{5}
}

func (x *RenewLeaseRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RenewLeaseRequest) GetLeaseDuration() *durationpb.Duration {
	if x != nil {
		return x.LeaseDuration
	}
	return nil
}

type RenewLeaseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *RenewLeaseResponse) Reset() {
	*x = RenewLeaseResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_tasks_grpc_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RenewLeaseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RenewLeaseResponse) ProtoMessage() {}

func (x *RenewLeaseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_tasks_grpc_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RenewLeaseResponse.ProtoReflect.Descriptor instead.
func (*RenewLeaseResponse) Descriptor() ([]byte, []int) {
	return file_proto_tasks_grpc_proto_rawDescGZIP(), []int{6}
}

func (x *RenewLeaseResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_proto_tasks_grpc_proto protoreflect.FileDescriptor

var file_proto_tasks_grpc_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x73, 0x5f, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x39, 0x0a, 0x07, 0x50, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x22, 0xac, 0x01, 0x0a, 0x0b, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x64, 0x12, 0x22, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x52,
	0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x41, 0x0a, 0x0e, 0x65, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0d, 0x65, 0x78,
	0x65, 0x63, 0x75, 0x74, 0x69, 0x6f, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x26, 0x0a, 0x0f, 0x6d,
	0x61, 0x78, 0x5f, 0x72, 0x65, 0x74, 0x72, 0x79, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x6d, 0x61, 0x78, 0x52, 0x65, 0x74, 0x72, 0x79, 0x43, 0x6f,
	0x75, 0x6e, 0x74, 0x22, 0x36, 0x0a, 0x0c, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x62, 0x0a, 0x14, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x77, 0x6f, 0x72, 0x6b, 0x65, 0x72, 0x49, 0x64, 0x12,
	0x16, 0x0a, 0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x74, 0x61, 0x73, 0x6b, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22,
	0x31, 0x0a, 0x15, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x22, 0x65, 0x0a, 0x11, 0x52, 0x65, 0x6e, 0x65, 0x77, 0x4c, 0x65, 0x61, 0x73, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x40, 0x0a, 0x0e, 0x6c, 0x65, 0x61, 0x73, 0x65,
	0x5f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0d, 0x6c, 0x65, 0x61, 0x73,
	0x65, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x2e, 0x0a, 0x12, 0x52, 0x65, 0x6e,
	0x65, 0x77, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x32, 0x7c, 0x0a, 0x0b, 0x54, 0x61, 0x73,
	0x6b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x2a, 0x0a, 0x0b, 0x45, 0x78, 0x65, 0x63,
	0x75, 0x74, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x12, 0x0c, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x41, 0x0a, 0x10, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x61,
	0x73, 0x6b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x15, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x16, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x45, 0x0a, 0x0c, 0x4c, 0x65, 0x61, 0x73, 0x65,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x35, 0x0a, 0x0a, 0x52, 0x65, 0x6e, 0x65, 0x77,
	0x4c, 0x65, 0x61, 0x73, 0x65, 0x12, 0x12, 0x2e, 0x52, 0x65, 0x6e, 0x65, 0x77, 0x4c, 0x65, 0x61,
	0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x13, 0x2e, 0x52, 0x65, 0x6e, 0x65,
	0x77, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x0d,
	0x5a, 0x0b, 0x2e, 0x2f, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_tasks_grpc_proto_rawDescOnce sync.Once
	file_proto_tasks_grpc_proto_rawDescData = file_proto_tasks_grpc_proto_rawDesc
)

func file_proto_tasks_grpc_proto_rawDescGZIP() []byte {
	file_proto_tasks_grpc_proto_rawDescOnce.Do(func() {
		file_proto_tasks_grpc_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_tasks_grpc_proto_rawDescData)
	})
	return file_proto_tasks_grpc_proto_rawDescData
}

var file_proto_tasks_grpc_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_proto_tasks_grpc_proto_goTypes = []interface{}{
	(*Payload)(nil),               // 0: Payload
	(*TaskRequest)(nil),           // 1: TaskRequest
	(*TaskResponse)(nil),          // 2: TaskResponse
	(*NotifyMessageRequest)(nil),  // 3: NotifyMessageRequest
	(*NotifyMessageResponse)(nil), // 4: NotifyMessageResponse
	(*RenewLeaseRequest)(nil),     // 5: RenewLeaseRequest
	(*RenewLeaseResponse)(nil),    // 6: RenewLeaseResponse
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
	(*durationpb.Duration)(nil),   // 8: google.protobuf.Duration
}
var file_proto_tasks_grpc_proto_depIdxs = []int32{
	0, // 0: TaskRequest.payload:type_name -> Payload
	7, // 1: TaskRequest.execution_time:type_name -> google.protobuf.Timestamp
	8, // 2: RenewLeaseRequest.lease_duration:type_name -> google.protobuf.Duration
	1, // 3: TaskService.ExecuteTask:input_type -> TaskRequest
	3, // 4: TaskService.NotifyTaskStatus:input_type -> NotifyMessageRequest
	5, // 5: LeaseService.RenewLease:input_type -> RenewLeaseRequest
	2, // 6: TaskService.ExecuteTask:output_type -> TaskResponse
	4, // 7: TaskService.NotifyTaskStatus:output_type -> NotifyMessageResponse
	6, // 8: LeaseService.RenewLease:output_type -> RenewLeaseResponse
	6, // [6:9] is the sub-list for method output_type
	3, // [3:6] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_proto_tasks_grpc_proto_init() }
func file_proto_tasks_grpc_proto_init() {
	if File_proto_tasks_grpc_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_tasks_grpc_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Payload); i {
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
		file_proto_tasks_grpc_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskRequest); i {
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
		file_proto_tasks_grpc_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TaskResponse); i {
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
		file_proto_tasks_grpc_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyMessageRequest); i {
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
		file_proto_tasks_grpc_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyMessageResponse); i {
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
		file_proto_tasks_grpc_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RenewLeaseRequest); i {
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
		file_proto_tasks_grpc_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RenewLeaseResponse); i {
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
			RawDescriptor: file_proto_tasks_grpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   2,
		},
		GoTypes:           file_proto_tasks_grpc_proto_goTypes,
		DependencyIndexes: file_proto_tasks_grpc_proto_depIdxs,
		MessageInfos:      file_proto_tasks_grpc_proto_msgTypes,
	}.Build()
	File_proto_tasks_grpc_proto = out.File
	file_proto_tasks_grpc_proto_rawDesc = nil
	file_proto_tasks_grpc_proto_goTypes = nil
	file_proto_tasks_grpc_proto_depIdxs = nil
}
