package messages

import (
	"reflect"

	"google.golang.org/protobuf/runtime/protoimpl"
)

func Initialise() {
	if File_mq_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		File_mq_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageInfo); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageData); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageCollection); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestMessageData); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueueInfo); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AckUpdate); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Credentials); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ChangeCredentials); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteQueueInfo); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddRoute); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteRoute); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserCreds); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[13].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchMessages); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[14].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaderNodeRequest); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[15].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserInformation); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[16].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchAckUpdate); i {
			case 0:
				return &v.State
			case 1:
				return &v.SizeCache
			case 2:
				return &v.UnknownFields
			default:
				return nil
			}
		}
		File_mq_proto_msgTypes[17].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchNackUpdate); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: File_mq_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   18,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           File_mq_proto_goTypes,
		DependencyIndexes: File_mq_proto_depIdxs,
		MessageInfos:      File_mq_proto_msgTypes,
	}.Build()
	File_mq_proto = out.File
	File_mq_proto_rawDesc = nil
	File_mq_proto_goTypes = nil
	File_mq_proto_depIdxs = nil
}
