package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type QueueInfo struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	RoutingKey string `protobuf:"bytes,1,opt,name=routingKey,proto3" json:"routingKey,omitempty"`
	Name       string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Size       uint32 `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
}

func (x *QueueInfo) Reset() {
	*x = QueueInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueueInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueueInfo) ProtoMessage() {}

func (x *QueueInfo) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueueInfo.ProtoReflect.Descriptor instead.
func (*QueueInfo) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{5}
}

func (x *QueueInfo) GetRoutingKey() string {
	if x != nil {
		return x.RoutingKey
	}
	return ""
}

func (x *QueueInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *QueueInfo) GetSize() uint32 {
	if x != nil {
		return x.Size
	}
	return 0
}
