package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type DeleteRoute struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	QueueName string `protobuf:"bytes,1,opt,name=queueName,proto3" json:"queueName,omitempty"`
	RouteName string `protobuf:"bytes,2,opt,name=routeName,proto3" json:"routeName,omitempty"`
}

func (x *DeleteRoute) Reset() {
	*x = DeleteRoute{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteRoute) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteRoute) ProtoMessage() {}

func (x *DeleteRoute) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteRoute.ProtoReflect.Descriptor instead.
func (*DeleteRoute) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{11}
}

func (x *DeleteRoute) GetQueueName() string {
	if x != nil {
		return x.QueueName
	}
	return ""
}

func (x *DeleteRoute) GetRouteName() string {
	if x != nil {
		return x.RouteName
	}
	return ""
}
