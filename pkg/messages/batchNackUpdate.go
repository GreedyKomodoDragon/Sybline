package messages

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type BatchNackUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	QueueName string   `protobuf:"bytes,1,opt,name=queueName,proto3" json:"queueName,omitempty"`
	Ids       [][]byte `protobuf:"bytes,2,rep,name=ids,proto3" json:"ids,omitempty"`
}

func (x *BatchNackUpdate) Reset() {
	*x = BatchNackUpdate{}
	if protoimpl.UnsafeEnabled {
		mi := &File_mq_proto_msgTypes[17]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchNackUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchNackUpdate) ProtoMessage() {}

func (x *BatchNackUpdate) ProtoReflect() protoreflect.Message {
	mi := &File_mq_proto_msgTypes[17]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchNackUpdate.ProtoReflect.Descriptor instead.
func (*BatchNackUpdate) Descriptor() ([]byte, []int) {
	return File_mq_proto_rawDescGZIP(), []int{17}
}

func (x *BatchNackUpdate) GetQueueName() string {
	if x != nil {
		return x.QueueName
	}
	return ""
}

func (x *BatchNackUpdate) GetIds() [][]byte {
	if x != nil {
		return x.Ids
	}
	return nil
}
