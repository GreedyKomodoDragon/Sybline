package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type BatchAckUpdate struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	QueueName string   `protobuf:"bytes,1,opt,name=queueName,proto3" json:"queueName,omitempty"`
	Id        [][]byte `protobuf:"bytes,2,rep,name=id,proto3" json:"id,omitempty"`
}

func (x *BatchAckUpdate) Reset() {
	*x = BatchAckUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[16]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchAckUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchAckUpdate) ProtoMessage() {}

func (x *BatchAckUpdate) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[16]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchAckUpdate.ProtoReflect.Descriptor instead.
func (*BatchAckUpdate) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{16}
}

func (x *BatchAckUpdate) GetQueueName() string {
	if x != nil {
		return x.QueueName
	}
	return ""
}

func (x *BatchAckUpdate) GetId() [][]byte {
	if x != nil {
		return x.Id
	}
	return nil
}
