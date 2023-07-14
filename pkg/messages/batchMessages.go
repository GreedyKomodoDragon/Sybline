package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type BatchMessages struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	Messages []*MessageInfo `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *BatchMessages) Reset() {
	*x = BatchMessages{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[13]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchMessages) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchMessages) ProtoMessage() {}

func (x *BatchMessages) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[13]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchMessages.ProtoReflect.Descriptor instead.
func (*BatchMessages) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{13}
}

func (x *BatchMessages) GetMessages() []*MessageInfo {
	if x != nil {
		return x.Messages
	}
	return nil
}
