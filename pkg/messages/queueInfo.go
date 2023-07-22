package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type QueueInfo struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	RoutingKey string
	Name       string
	Size       uint32
	RetryLimit uint32
	HasDLQueue bool
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
