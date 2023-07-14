package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type AckUpdate struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	QueueName string `protobuf:"bytes,1,opt,name=queueName,proto3" json:"queueName,omitempty"`
	Id        []byte `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *AckUpdate) Reset() {
	*x = AckUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AckUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AckUpdate) ProtoMessage() {}

func (x *AckUpdate) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AckUpdate.ProtoReflect.Descriptor instead.
func (*AckUpdate) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{6}
}

func (x *AckUpdate) GetQueueName() string {
	if x != nil {
		return x.QueueName
	}
	return ""
}

func (x *AckUpdate) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}
