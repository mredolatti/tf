// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.19.2
// source: changes.proto

package is2fs

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

type ChangeType int32

const (
	ChangeType_FileChangeAdd    ChangeType = 0
	ChangeType_FileChangeDelete ChangeType = 1
	ChangeType_FileChangeUpdate ChangeType = 2
)

// Enum value maps for ChangeType.
var (
	ChangeType_name = map[int32]string{
		0: "FileChangeAdd",
		1: "FileChangeDelete",
		2: "FileChangeUpdate",
	}
	ChangeType_value = map[string]int32{
		"FileChangeAdd":    0,
		"FileChangeDelete": 1,
		"FileChangeUpdate": 2,
	}
)

func (x ChangeType) Enum() *ChangeType {
	p := new(ChangeType)
	*p = x
	return p
}

func (x ChangeType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ChangeType) Descriptor() protoreflect.EnumDescriptor {
	return file_changes_proto_enumTypes[0].Descriptor()
}

func (ChangeType) Type() protoreflect.EnumType {
	return &file_changes_proto_enumTypes[0]
}

func (x ChangeType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ChangeType.Descriptor instead.
func (ChangeType) EnumDescriptor() ([]byte, []int) {
	return file_changes_proto_rawDescGZIP(), []int{0}
}

type Update struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FileReference string     `protobuf:"bytes,1,opt,name=fileReference,proto3" json:"fileReference,omitempty"`
	ChangeType    ChangeType `protobuf:"varint,2,opt,name=changeType,proto3,enum=ChangeType" json:"changeType,omitempty"`
	Checkpoint    int64      `protobuf:"varint,3,opt,name=checkpoint,proto3" json:"checkpoint,omitempty"`
}

func (x *Update) Reset() {
	*x = Update{}
	if protoimpl.UnsafeEnabled {
		mi := &file_changes_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Update) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Update) ProtoMessage() {}

func (x *Update) ProtoReflect() protoreflect.Message {
	mi := &file_changes_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Update.ProtoReflect.Descriptor instead.
func (*Update) Descriptor() ([]byte, []int) {
	return file_changes_proto_rawDescGZIP(), []int{0}
}

func (x *Update) GetFileReference() string {
	if x != nil {
		return x.FileReference
	}
	return ""
}

func (x *Update) GetChangeType() ChangeType {
	if x != nil {
		return x.ChangeType
	}
	return ChangeType_FileChangeAdd
}

func (x *Update) GetCheckpoint() int64 {
	if x != nil {
		return x.Checkpoint
	}
	return 0
}

type SyncUserRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID     string `protobuf:"bytes,1,opt,name=userID,proto3" json:"userID,omitempty"`
	Checkpoint int64  `protobuf:"varint,2,opt,name=checkpoint,proto3" json:"checkpoint,omitempty"`
	KeepAlive  bool   `protobuf:"varint,3,opt,name=keepAlive,proto3" json:"keepAlive,omitempty"`
}

func (x *SyncUserRequest) Reset() {
	*x = SyncUserRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_changes_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncUserRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncUserRequest) ProtoMessage() {}

func (x *SyncUserRequest) ProtoReflect() protoreflect.Message {
	mi := &file_changes_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncUserRequest.ProtoReflect.Descriptor instead.
func (*SyncUserRequest) Descriptor() ([]byte, []int) {
	return file_changes_proto_rawDescGZIP(), []int{1}
}

func (x *SyncUserRequest) GetUserID() string {
	if x != nil {
		return x.UserID
	}
	return ""
}

func (x *SyncUserRequest) GetCheckpoint() int64 {
	if x != nil {
		return x.Checkpoint
	}
	return 0
}

func (x *SyncUserRequest) GetKeepAlive() bool {
	if x != nil {
		return x.KeepAlive
	}
	return false
}

var File_changes_proto protoreflect.FileDescriptor

var file_changes_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x7b, 0x0a, 0x06, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x66, 0x69, 0x6c,
	0x65, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x66, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x12,
	0x2b, 0x0a, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x0b, 0x2e, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x0a, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1e, 0x0a, 0x0a,
	0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03,
	0x52, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x22, 0x67, 0x0a, 0x0f,
	0x53, 0x79, 0x6e, 0x63, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x16, 0x0a, 0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x75, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x1e, 0x0a, 0x0a, 0x63, 0x68, 0x65, 0x63, 0x6b,
	0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0a, 0x63, 0x68, 0x65,
	0x63, 0x6b, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6b, 0x65, 0x65, 0x70, 0x41,
	0x6c, 0x69, 0x76, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x6b, 0x65, 0x65, 0x70,
	0x41, 0x6c, 0x69, 0x76, 0x65, 0x2a, 0x4b, 0x0a, 0x0a, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x12, 0x11, 0x0a, 0x0d, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67,
	0x65, 0x41, 0x64, 0x64, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x46, 0x69, 0x6c, 0x65, 0x43, 0x68,
	0x61, 0x6e, 0x67, 0x65, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x10, 0x01, 0x12, 0x14, 0x0a, 0x10,
	0x46, 0x69, 0x6c, 0x65, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x10, 0x02, 0x32, 0x38, 0x0a, 0x0b, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x66, 0x53, 0x79, 0x6e,
	0x63, 0x12, 0x29, 0x0a, 0x08, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x73, 0x65, 0x72, 0x12, 0x10, 0x2e,
	0x53, 0x79, 0x6e, 0x63, 0x55, 0x73, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x07, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x22, 0x00, 0x30, 0x01, 0x42, 0x2e, 0x5a, 0x2c,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x72, 0x65, 0x64, 0x6f,
	0x6c, 0x61, 0x74, 0x74, 0x69, 0x2f, 0x74, 0x66, 0x2f, 0x63, 0x6f, 0x64, 0x69, 0x67, 0x6f, 0x2f,
	0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x69, 0x73, 0x32, 0x66, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_changes_proto_rawDescOnce sync.Once
	file_changes_proto_rawDescData = file_changes_proto_rawDesc
)

func file_changes_proto_rawDescGZIP() []byte {
	file_changes_proto_rawDescOnce.Do(func() {
		file_changes_proto_rawDescData = protoimpl.X.CompressGZIP(file_changes_proto_rawDescData)
	})
	return file_changes_proto_rawDescData
}

var file_changes_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_changes_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_changes_proto_goTypes = []interface{}{
	(ChangeType)(0),         // 0: ChangeType
	(*Update)(nil),          // 1: Update
	(*SyncUserRequest)(nil), // 2: SyncUserRequest
}
var file_changes_proto_depIdxs = []int32{
	0, // 0: Update.changeType:type_name -> ChangeType
	2, // 1: FileRefSync.SyncUser:input_type -> SyncUserRequest
	1, // 2: FileRefSync.SyncUser:output_type -> Update
	2, // [2:3] is the sub-list for method output_type
	1, // [1:2] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_changes_proto_init() }
func file_changes_proto_init() {
	if File_changes_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_changes_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Update); i {
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
		file_changes_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncUserRequest); i {
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
			RawDescriptor: file_changes_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_changes_proto_goTypes,
		DependencyIndexes: file_changes_proto_depIdxs,
		EnumInfos:         file_changes_proto_enumTypes,
		MessageInfos:      file_changes_proto_msgTypes,
	}.Build()
	File_changes_proto = out.File
	file_changes_proto_rawDesc = nil
	file_changes_proto_goTypes = nil
	file_changes_proto_depIdxs = nil
}
