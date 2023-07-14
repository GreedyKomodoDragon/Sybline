package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type UserInformation struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
}

func (x *UserInformation) Reset() {
	*x = UserInformation{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[15]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserInformation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserInformation) ProtoMessage() {}

func (x *UserInformation) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[15]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserInformation.ProtoReflect.Descriptor instead.
func (*UserInformation) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{15}
}

func (x *UserInformation) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}
