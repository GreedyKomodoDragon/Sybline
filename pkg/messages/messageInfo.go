package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type MessageInfo struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	RoutingKey string `protobuf:"bytes,1,opt,name=routingKey,proto3" json:"routingKey,omitempty"`
	Data       []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *MessageInfo) Reset() {
	*x = MessageInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageInfo) ProtoMessage() {}

func (x *MessageInfo) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageInfo.ProtoReflect.Descriptor instead.
func (*MessageInfo) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{0}
}

func (x *MessageInfo) GetRoutingKey() string {
	if x != nil {
		return x.RoutingKey
	}
	return ""
}

func (x *MessageInfo) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}
