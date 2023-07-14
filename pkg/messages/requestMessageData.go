package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type RequestMessageData struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	QueueName string `protobuf:"bytes,1,opt,name=queueName,proto3" json:"queueName,omitempty"`
	Amount    uint32 `protobuf:"varint,2,opt,name=amount,proto3" json:"amount,omitempty"`
}

func (x *RequestMessageData) Reset() {
	*x = RequestMessageData{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestMessageData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestMessageData) ProtoMessage() {}

func (x *RequestMessageData) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestMessageData.ProtoReflect.Descriptor instead.
func (*RequestMessageData) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{3}
}

func (x *RequestMessageData) GetQueueName() string {
	if x != nil {
		return x.QueueName
	}
	return ""
}

func (x *RequestMessageData) GetAmount() uint32 {
	if x != nil {
		return x.Amount
	}
	return 0
}
