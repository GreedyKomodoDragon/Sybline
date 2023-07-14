package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type MessageData struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	Id   []byte `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *MessageData) Reset() {
	*x = MessageData{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageData) ProtoMessage() {}

func (x *MessageData) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageData.ProtoReflect.Descriptor instead.
func (*MessageData) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{1}
}

func (x *MessageData) GetId() []byte {
	if x != nil {
		return x.Id
	}
	return nil
}

func (x *MessageData) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}
