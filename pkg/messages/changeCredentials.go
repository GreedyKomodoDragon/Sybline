package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type ChangeCredentials struct {
	State         protoimpl.MessageState
	SizeCache     protoimpl.SizeCache
	UnknownFields protoimpl.UnknownFields

	Username    string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	OldPassword string `protobuf:"bytes,2,opt,name=oldPassword,proto3" json:"oldPassword,omitempty"`
	NewPassword string `protobuf:"bytes,3,opt,name=newPassword,proto3" json:"newPassword,omitempty"`
}

func (x *ChangeCredentials) Reset() {
	*x = ChangeCredentials{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ChangeCredentials) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChangeCredentials) ProtoMessage() {}

func (x *ChangeCredentials) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChangeCredentials.ProtoReflect.Descriptor instead.
func (*ChangeCredentials) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{8}
}

func (x *ChangeCredentials) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *ChangeCredentials) GetOldPassword() string {
	if x != nil {
		return x.OldPassword
	}
	return ""
}

func (x *ChangeCredentials) GetNewPassword() string {
	if x != nil {
		return x.NewPassword
	}
	return ""
}
