package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type DeleteQueueInfo struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	QueueName string `protobuf:"bytes,1,opt,name=queueName,proto3" json:"queueName,omitempty"`
}

func (x *DeleteQueueInfo) Reset() {
	*x = DeleteQueueInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteQueueInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteQueueInfo) ProtoMessage() {}

func (x *DeleteQueueInfo) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteQueueInfo.ProtoReflect.Descriptor instead.
func (*DeleteQueueInfo) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{9}
}

func (x *DeleteQueueInfo) GetQueueName() string {
	if x != nil {
		return x.QueueName
	}
	return ""
}
