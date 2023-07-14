package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type LeaderNodeRequest struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields
}

func (x *LeaderNodeRequest) Reset() {
	*x = LeaderNodeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[14]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaderNodeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaderNodeRequest) ProtoMessage() {}

func (x *LeaderNodeRequest) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[14]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaderNodeRequest.ProtoReflect.Descriptor instead.
func (*LeaderNodeRequest) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{14}
}
