package messages

import (
	"sync"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

var File_mq_proto protoreflect.FileDescriptor

var File_mq_proto_rawDesc = []byte{
	0x0a, 0x08, 0x6d, 0x71, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x4d, 0x51, 0x22, 0x41,
	0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x1e, 0x0a,
	0x0a, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x67, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x67, 0x4b, 0x65, 0x79, 0x12, 0x12, 0x0a,
	0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x22, 0x31, 0x0a, 0x0b, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x44, 0x61, 0x74, 0x61,
	0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x22, 0x40, 0x0a, 0x11, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x43,
	0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x2b, 0x0a, 0x08, 0x6d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x4d, 0x51,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x44, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22, 0x4a, 0x0a, 0x12, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x09,
	0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x6d,
	0x6f, 0x75, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x61, 0x6d, 0x6f, 0x75,
	0x6e, 0x74, 0x22, 0x20, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x16, 0x0a, 0x06,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x22, 0x53, 0x0a, 0x09, 0x51, 0x75, 0x65, 0x75, 0x65, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x1e, 0x0a, 0x0a, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x67, 0x4b, 0x65, 0x79, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x72, 0x6f, 0x75, 0x74, 0x69, 0x6e, 0x67, 0x4b, 0x65,
	0x79, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x22, 0x39, 0x0a, 0x09, 0x41, 0x63, 0x6b,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x02, 0x69, 0x64, 0x22, 0x45, 0x0a, 0x0b, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69,
	0x61, 0x6c, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x22, 0x73, 0x0a, 0x11, 0x43,
	0x68, 0x61, 0x6e, 0x67, 0x65, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73,
	0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b,
	0x6f, 0x6c, 0x64, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x6f, 0x6c, 0x64, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x20,
	0x0a, 0x0b, 0x6e, 0x65, 0x77, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0b, 0x6e, 0x65, 0x77, 0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64,
	0x22, 0x2f, 0x0a, 0x0f, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x49,
	0x6e, 0x66, 0x6f, 0x12, 0x1c, 0x0a, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d,
	0x65, 0x22, 0x46, 0x0a, 0x08, 0x41, 0x64, 0x64, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x72,
	0x6f, 0x75, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x72, 0x6f, 0x75, 0x74, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x22, 0x49, 0x0a, 0x0b, 0x44, 0x65, 0x6c,
	0x65, 0x74, 0x65, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x71, 0x75, 0x65, 0x75,
	0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x71, 0x75, 0x65,
	0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x6f, 0x75, 0x74, 0x65,
	0x4e, 0x61, 0x6d, 0x65, 0x22, 0x43, 0x0a, 0x09, 0x55, 0x73, 0x65, 0x72, 0x43, 0x72, 0x65, 0x64,
	0x73, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x22, 0x3c, 0x0a, 0x0d, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x2b, 0x0a, 0x08, 0x6d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x4d,
	0x51, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08, 0x6d,
	0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22, 0x13, 0x0a, 0x11, 0x4c, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x2d, 0x0a, 0x0f,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x3e, 0x0a, 0x0e, 0x42,
	0x61, 0x74, 0x63, 0x68, 0x41, 0x63, 0x6b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1c, 0x0a,
	0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x02, 0x69, 0x64, 0x22, 0x41, 0x0a, 0x0f, 0x42,
	0x61, 0x74, 0x63, 0x68, 0x4e, 0x61, 0x63, 0x6b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x1c,
	0x0a, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x71, 0x75, 0x65, 0x75, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x69, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x03, 0x69, 0x64, 0x73, 0x32, 0x8c,
	0x06, 0x0a, 0x0b, 0x4d, 0x51, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x73, 0x12, 0x3e,
	0x0a, 0x0b, 0x47, 0x65, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x16, 0x2e,
	0x4d, 0x51, 0x2e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x15, 0x2e, 0x4d, 0x51, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x00, 0x12, 0x2e,
	0x0a, 0x0d, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12,
	0x0f, 0x2e, 0x4d, 0x51, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x66, 0x6f,
	0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x2a,
	0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75, 0x65, 0x12, 0x0d, 0x2e,
	0x4d, 0x51, 0x2e, 0x51, 0x75, 0x65, 0x75, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x0a, 0x2e, 0x4d,
	0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x22, 0x0a, 0x03, 0x41, 0x63,
	0x6b, 0x12, 0x0d, 0x2e, 0x4d, 0x51, 0x2e, 0x41, 0x63, 0x6b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x26,
	0x0a, 0x05, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x12, 0x0f, 0x2e, 0x4d, 0x51, 0x2e, 0x43, 0x72, 0x65,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x35, 0x0a, 0x0e, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65,
	0x50, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x15, 0x2e, 0x4d, 0x51, 0x2e, 0x43, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x43, 0x72, 0x65, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x61, 0x6c, 0x73, 0x1a,
	0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x23, 0x0a,
	0x04, 0x4e, 0x61, 0x63, 0x6b, 0x12, 0x0d, 0x2e, 0x4d, 0x51, 0x2e, 0x41, 0x63, 0x6b, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x22, 0x00, 0x12, 0x30, 0x0a, 0x0b, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x51, 0x75, 0x65, 0x75,
	0x65, 0x12, 0x13, 0x2e, 0x4d, 0x51, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x51, 0x75, 0x65,
	0x75, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x22, 0x00, 0x12, 0x2b, 0x0a, 0x0d, 0x41, 0x64, 0x64, 0x52, 0x6f, 0x75, 0x74, 0x69,
	0x6e, 0x67, 0x4b, 0x65, 0x79, 0x12, 0x0c, 0x2e, 0x4d, 0x51, 0x2e, 0x41, 0x64, 0x64, 0x52, 0x6f,
	0x75, 0x74, 0x65, 0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22,
	0x00, 0x12, 0x31, 0x0a, 0x10, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x6f, 0x75, 0x74, 0x69,
	0x6e, 0x67, 0x4b, 0x65, 0x79, 0x12, 0x0f, 0x2e, 0x4d, 0x51, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74,
	0x65, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x22, 0x00, 0x12, 0x29, 0x0a, 0x0a, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x55, 0x73,
	0x65, 0x72, 0x12, 0x0d, 0x2e, 0x4d, 0x51, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x43, 0x72, 0x65, 0x64,
	0x73, 0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12,
	0x38, 0x0a, 0x15, 0x53, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x42, 0x61, 0x74, 0x63, 0x68, 0x65, 0x64,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x12, 0x11, 0x2e, 0x4d, 0x51, 0x2e, 0x42, 0x61,
	0x74, 0x63, 0x68, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x1a, 0x0a, 0x2e, 0x4d, 0x51,
	0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x33, 0x0a, 0x0c, 0x49, 0x73, 0x4c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x4e, 0x6f, 0x64, 0x65, 0x12, 0x15, 0x2e, 0x4d, 0x51, 0x2e, 0x4c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x4e, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x2f,
	0x0a, 0x0a, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x55, 0x73, 0x65, 0x72, 0x12, 0x13, 0x2e, 0x4d,
	0x51, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x49, 0x6e, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x1a, 0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12,
	0x2c, 0x0a, 0x08, 0x42, 0x61, 0x74, 0x63, 0x68, 0x41, 0x63, 0x6b, 0x12, 0x12, 0x2e, 0x4d, 0x51,
	0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x41, 0x63, 0x6b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x1a,
	0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x12, 0x2e, 0x0a,
	0x09, 0x42, 0x61, 0x74, 0x63, 0x68, 0x4e, 0x61, 0x63, 0x6b, 0x12, 0x13, 0x2e, 0x4d, 0x51, 0x2e,
	0x42, 0x61, 0x74, 0x63, 0x68, 0x4e, 0x61, 0x63, 0x6b, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x1a,
	0x0a, 0x2e, 0x4d, 0x51, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x00, 0x42, 0x08, 0x5a,
	0x06, 0x2e, 0x2f, 0x63, 0x6f, 0x72, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	File_mq_proto_rawDescOnce sync.Once
	File_mq_proto_rawDescData = File_mq_proto_rawDesc
)

func File_mq_proto_rawDescGZIP() []byte {
	File_mq_proto_rawDescOnce.Do(func() {
		File_mq_proto_rawDescData = protoimpl.X.CompressGZIP(File_mq_proto_rawDescData)
	})
	return File_mq_proto_rawDescData
}

var File_mq_proto_msgTypes = make([]protoimpl.MessageInfo, 18)

var File_mq_proto_depIdxs = []int32{
	1,  // 0: MQ.MessageCollection.messages:type_name -> MQ.MessageData
	0,  // 1: MQ.BatchMessages.messages:type_name -> MQ.MessageInfo
	3,  // 2: MQ.MQEndpoints.GetMessages:input_type -> MQ.RequestMessageData
	0,  // 3: MQ.MQEndpoints.SubmitMessage:input_type -> MQ.MessageInfo
	5,  // 4: MQ.MQEndpoints.CreateQueue:input_type -> MQ.QueueInfo
	6,  // 5: MQ.MQEndpoints.Ack:input_type -> MQ.AckUpdate
	7,  // 6: MQ.MQEndpoints.Login:input_type -> MQ.Credentials
	8,  // 7: MQ.MQEndpoints.ChangePassword:input_type -> MQ.ChangeCredentials
	6,  // 8: MQ.MQEndpoints.Nack:input_type -> MQ.AckUpdate
	9,  // 9: MQ.MQEndpoints.DeleteQueue:input_type -> MQ.DeleteQueueInfo
	10, // 10: MQ.MQEndpoints.AddRoutingKey:input_type -> MQ.AddRoute
	11, // 11: MQ.MQEndpoints.DeleteRoutingKey:input_type -> MQ.DeleteRoute
	12, // 12: MQ.MQEndpoints.CreateUser:input_type -> MQ.UserCreds
	13, // 13: MQ.MQEndpoints.SubmitBatchedMessages:input_type -> MQ.BatchMessages
	14, // 14: MQ.MQEndpoints.IsLeaderNode:input_type -> MQ.LeaderNodeRequest
	15, // 15: MQ.MQEndpoints.DeleteUser:input_type -> MQ.UserInformation
	16, // 16: MQ.MQEndpoints.BatchAck:input_type -> MQ.BatchAckUpdate
	17, // 17: MQ.MQEndpoints.BatchNack:input_type -> MQ.BatchNackUpdate
	2,  // 18: MQ.MQEndpoints.GetMessages:output_type -> MQ.MessageCollection
	4,  // 19: MQ.MQEndpoints.SubmitMessage:output_type -> MQ.Status
	4,  // 20: MQ.MQEndpoints.CreateQueue:output_type -> MQ.Status
	4,  // 21: MQ.MQEndpoints.Ack:output_type -> MQ.Status
	4,  // 22: MQ.MQEndpoints.Login:output_type -> MQ.Status
	4,  // 23: MQ.MQEndpoints.ChangePassword:output_type -> MQ.Status
	4,  // 24: MQ.MQEndpoints.Nack:output_type -> MQ.Status
	4,  // 25: MQ.MQEndpoints.DeleteQueue:output_type -> MQ.Status
	4,  // 26: MQ.MQEndpoints.AddRoutingKey:output_type -> MQ.Status
	4,  // 27: MQ.MQEndpoints.DeleteRoutingKey:output_type -> MQ.Status
	4,  // 28: MQ.MQEndpoints.CreateUser:output_type -> MQ.Status
	4,  // 29: MQ.MQEndpoints.SubmitBatchedMessages:output_type -> MQ.Status
	4,  // 30: MQ.MQEndpoints.IsLeaderNode:output_type -> MQ.Status
	4,  // 31: MQ.MQEndpoints.DeleteUser:output_type -> MQ.Status
	4,  // 32: MQ.MQEndpoints.BatchAck:output_type -> MQ.Status
	4,  // 33: MQ.MQEndpoints.BatchNack:output_type -> MQ.Status
	18, // [18:34] is the sub-list for method output_type
	2,  // [2:18] is the sub-list for method input_type
	2,  // [2:2] is the sub-list for extension type_name
	2,  // [2:2] is the sub-list for extension extendee
	0,  // [0:2] is the sub-list for field type_name
}
