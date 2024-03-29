// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.23.4
// source: model.proto

package sharding_protobuf_test

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ShardingData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TotalNum int32  `protobuf:"varint,1,opt,name=total_num,json=totalNum,proto3" json:"total_num,omitempty"`
	ShardNum int32  `protobuf:"varint,2,opt,name=shard_num,json=shardNum,proto3" json:"shard_num,omitempty"`
	Digest   []byte `protobuf:"bytes,3,opt,name=digest,proto3" json:"digest,omitempty"`
	Data     []byte `protobuf:"bytes,4,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *ShardingData) Reset() {
	*x = ShardingData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_model_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ShardingData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ShardingData) ProtoMessage() {}

func (x *ShardingData) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ShardingData.ProtoReflect.Descriptor instead.
func (*ShardingData) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{0}
}

func (x *ShardingData) GetTotalNum() int32 {
	if x != nil {
		return x.TotalNum
	}
	return 0
}

func (x *ShardingData) GetShardNum() int32 {
	if x != nil {
		return x.ShardNum
	}
	return 0
}

func (x *ShardingData) GetDigest() []byte {
	if x != nil {
		return x.Digest
	}
	return nil
}

func (x *ShardingData) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type TestShardingModel struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id        int64         `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	BizData   []byte        `protobuf:"bytes,2,opt,name=biz_data,json=bizData,proto3" json:"biz_data,omitempty"`
	ShardData *ShardingData `protobuf:"bytes,255,opt,name=shard_data,json=shardData,proto3,oneof" json:"shard_data,omitempty"`
}

func (x *TestShardingModel) Reset() {
	*x = TestShardingModel{}
	if protoimpl.UnsafeEnabled {
		mi := &file_model_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TestShardingModel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TestShardingModel) ProtoMessage() {}

func (x *TestShardingModel) ProtoReflect() protoreflect.Message {
	mi := &file_model_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TestShardingModel.ProtoReflect.Descriptor instead.
func (*TestShardingModel) Descriptor() ([]byte, []int) {
	return file_model_proto_rawDescGZIP(), []int{1}
}

func (x *TestShardingModel) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *TestShardingModel) GetBizData() []byte {
	if x != nil {
		return x.BizData
	}
	return nil
}

func (x *TestShardingModel) GetShardData() *ShardingData {
	if x != nil {
		return x.ShardData
	}
	return nil
}

var File_model_proto protoreflect.FileDescriptor

var file_model_proto_rawDesc = []byte{
	0x0a, 0x0b, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x74, 0x0a,
	0x0c, 0x53, 0x68, 0x61, 0x72, 0x64, 0x69, 0x6e, 0x67, 0x44, 0x61, 0x74, 0x61, 0x12, 0x1b, 0x0a,
	0x09, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x6e, 0x75, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x08, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x4e, 0x75, 0x6d, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x68,
	0x61, 0x72, 0x64, 0x5f, 0x6e, 0x75, 0x6d, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x73,
	0x68, 0x61, 0x72, 0x64, 0x4e, 0x75, 0x6d, 0x12, 0x16, 0x0a, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73,
	0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x64, 0x69, 0x67, 0x65, 0x73, 0x74, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x22, 0x81, 0x01, 0x0a, 0x11, 0x54, 0x65, 0x73, 0x74, 0x53, 0x68, 0x61, 0x72,
	0x64, 0x69, 0x6e, 0x67, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x62, 0x69, 0x7a,
	0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x62, 0x69, 0x7a,
	0x44, 0x61, 0x74, 0x61, 0x12, 0x32, 0x0a, 0x0a, 0x73, 0x68, 0x61, 0x72, 0x64, 0x5f, 0x64, 0x61,
	0x74, 0x61, 0x18, 0xff, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x53, 0x68, 0x61, 0x72,
	0x64, 0x69, 0x6e, 0x67, 0x44, 0x61, 0x74, 0x61, 0x48, 0x00, 0x52, 0x09, 0x73, 0x68, 0x61, 0x72,
	0x64, 0x44, 0x61, 0x74, 0x61, 0x88, 0x01, 0x01, 0x42, 0x0d, 0x0a, 0x0b, 0x5f, 0x73, 0x68, 0x61,
	0x72, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x42, 0x1a, 0x5a, 0x18, 0x2e, 0x3b, 0x73, 0x68, 0x61,
	0x72, 0x64, 0x69, 0x6e, 0x67, 0x5f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x5f, 0x74,
	0x65, 0x73, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_model_proto_rawDescOnce sync.Once
	file_model_proto_rawDescData = file_model_proto_rawDesc
)

func file_model_proto_rawDescGZIP() []byte {
	file_model_proto_rawDescOnce.Do(func() {
		file_model_proto_rawDescData = protoimpl.X.CompressGZIP(file_model_proto_rawDescData)
	})
	return file_model_proto_rawDescData
}

var file_model_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_model_proto_goTypes = []interface{}{
	(*ShardingData)(nil),      // 0: ShardingData
	(*TestShardingModel)(nil), // 1: TestShardingModel
}
var file_model_proto_depIdxs = []int32{
	0, // 0: TestShardingModel.shard_data:type_name -> ShardingData
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_model_proto_init() }
func file_model_proto_init() {
	if File_model_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_model_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ShardingData); i {
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
		file_model_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TestShardingModel); i {
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
	file_model_proto_msgTypes[1].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_model_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_model_proto_goTypes,
		DependencyIndexes: file_model_proto_depIdxs,
		MessageInfos:      file_model_proto_msgTypes,
	}.Build()
	File_model_proto = out.File
	file_model_proto_rawDesc = nil
	file_model_proto_goTypes = nil
	file_model_proto_depIdxs = nil
}
