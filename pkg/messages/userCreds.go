package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type UserCreds struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password string `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
}

func (x *UserCreds) Reset() {
	*x = UserCreds{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserCreds) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserCreds) ProtoMessage() {}

func (x *UserCreds) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserCreds.ProtoReflect.Descriptor instead.
func (*UserCreds) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{12}
}

func (x *UserCreds) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *UserCreds) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}
