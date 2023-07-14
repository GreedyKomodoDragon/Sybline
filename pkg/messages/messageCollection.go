package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type MessageCollection struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	Messages []*MessageData `protobuf:"bytes,1,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *MessageCollection) Reset() {
	*x = MessageCollection{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageCollection) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageCollection) ProtoMessage() {}

func (x *MessageCollection) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageCollection.ProtoReflect.Descriptor instead.
func (*MessageCollection) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{2}
}

func (x *MessageCollection) GetMessages() []*MessageData {
	if x != nil {
		return x.Messages
	}
	return nil
}
