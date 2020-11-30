// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: collector.proto

package g

import (
	proto "github.com/golang/protobuf/proto"
	empty "github.com/golang/protobuf/ptypes/empty"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type AgentStatus_Status int32

const (
	AgentStatus_RUNNING AgentStatus_Status = 0
	AgentStatus_CONFIRM AgentStatus_Status = 1
	AgentStatus_CLOSED  AgentStatus_Status = 2
)

// Enum value maps for AgentStatus_Status.
var (
	AgentStatus_Status_name = map[int32]string{
		0: "RUNNING",
		1: "CONFIRM",
		2: "CLOSED",
	}
	AgentStatus_Status_value = map[string]int32{
		"RUNNING": 0,
		"CONFIRM": 1,
		"CLOSED":  2,
	}
)

func (x AgentStatus_Status) Enum() *AgentStatus_Status {
	p := new(AgentStatus_Status)
	*p = x
	return p
}

func (x AgentStatus_Status) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AgentStatus_Status) Descriptor() protoreflect.EnumDescriptor {
	return file_collector_proto_enumTypes[0].Descriptor()
}

func (AgentStatus_Status) Type() protoreflect.EnumType {
	return &file_collector_proto_enumTypes[0]
}

func (x AgentStatus_Status) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AgentStatus_Status.Descriptor instead.
func (AgentStatus_Status) EnumDescriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{2, 0}
}

type OK struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (x *OK) Reset() {
	*x = OK{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OK) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OK) ProtoMessage() {}

func (x *OK) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OK.ProtoReflect.Descriptor instead.
func (*OK) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{0}
}

func (x *OK) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

type TraceID struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ID string `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
}

func (x *TraceID) Reset() {
	*x = TraceID{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TraceID) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TraceID) ProtoMessage() {}

func (x *TraceID) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TraceID.ProtoReflect.Descriptor instead.
func (*TraceID) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{1}
}

func (x *TraceID) GetID() string {
	if x != nil {
		return x.ID
	}
	return ""
}

type AgentStatus struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status AgentStatus_Status `protobuf:"varint,1,opt,name=status,proto3,enum=g.AgentStatus_Status" json:"status,omitempty"`
}

func (x *AgentStatus) Reset() {
	*x = AgentStatus{}
	if protoimpl.UnsafeEnabled {
		mi := &file_collector_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AgentStatus) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AgentStatus) ProtoMessage() {}

func (x *AgentStatus) ProtoReflect() protoreflect.Message {
	mi := &file_collector_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AgentStatus.ProtoReflect.Descriptor instead.
func (*AgentStatus) Descriptor() ([]byte, []int) {
	return file_collector_proto_rawDescGZIP(), []int{2}
}

func (x *AgentStatus) GetStatus() AgentStatus_Status {
	if x != nil {
		return x.Status
	}
	return AgentStatus_RUNNING
}

var File_collector_proto protoreflect.FileDescriptor

var file_collector_proto_rawDesc = []byte{
	0x0a, 0x0f, 0x63, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x01, 0x67, 0x1a, 0x0b, 0x74, 0x72, 0x61, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x14,
	0x0a, 0x02, 0x4f, 0x4b, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x02, 0x6f, 0x6b, 0x22, 0x19, 0x0a, 0x07, 0x54, 0x72, 0x61, 0x63, 0x65, 0x49, 0x44, 0x12,
	0x0e, 0x0a, 0x02, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x49, 0x44, 0x22,
	0x6c, 0x0a, 0x0b, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x2d,
	0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15,
	0x2e, 0x67, 0x2e, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x2e, 0x0a,
	0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0b, 0x0a, 0x07, 0x52, 0x55, 0x4e, 0x4e, 0x49,
	0x4e, 0x47, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x52, 0x4d, 0x10,
	0x01, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x4c, 0x4f, 0x53, 0x45, 0x44, 0x10, 0x02, 0x32, 0xa8, 0x01,
	0x0a, 0x09, 0x43, 0x6f, 0x6c, 0x6c, 0x65, 0x63, 0x74, 0x6f, 0x72, 0x12, 0x31, 0x0a, 0x09, 0x53,
	0x65, 0x6e, 0x64, 0x54, 0x72, 0x61, 0x63, 0x65, 0x12, 0x08, 0x2e, 0x67, 0x2e, 0x54, 0x72, 0x61,
	0x63, 0x65, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x28, 0x01, 0x12, 0x3a,
	0x0a, 0x10, 0x53, 0x75, 0x62, 0x73, 0x63, 0x72, 0x69, 0x62, 0x65, 0x54, 0x72, 0x61, 0x63, 0x65,
	0x49, 0x44, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0a, 0x2e, 0x67, 0x2e, 0x54,
	0x72, 0x61, 0x63, 0x65, 0x49, 0x44, 0x22, 0x00, 0x30, 0x01, 0x12, 0x2c, 0x0a, 0x0d, 0x43, 0x6f,
	0x6e, 0x66, 0x69, 0x72, 0x6d, 0x46, 0x69, 0x6e, 0x69, 0x73, 0x68, 0x12, 0x0e, 0x2e, 0x67, 0x2e,
	0x41, 0x67, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x1a, 0x05, 0x2e, 0x67, 0x2e,
	0x4f, 0x4b, 0x22, 0x00, 0x28, 0x01, 0x30, 0x01, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6d, 0x69, 0x6e, 0x67, 0x6f, 0x6f, 0x6f, 0x6f, 0x2f,
	0x74, 0x61, 0x69, 0x6c, 0x2d, 0x62, 0x61, 0x73, 0x65, 0x64, 0x2d, 0x73, 0x61, 0x6d, 0x70, 0x6c,
	0x69, 0x6e, 0x67, 0x2f, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_collector_proto_rawDescOnce sync.Once
	file_collector_proto_rawDescData = file_collector_proto_rawDesc
)

func file_collector_proto_rawDescGZIP() []byte {
	file_collector_proto_rawDescOnce.Do(func() {
		file_collector_proto_rawDescData = protoimpl.X.CompressGZIP(file_collector_proto_rawDescData)
	})
	return file_collector_proto_rawDescData
}

var file_collector_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_collector_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_collector_proto_goTypes = []interface{}{
	(AgentStatus_Status)(0), // 0: g.AgentStatus.Status
	(*OK)(nil),              // 1: g.OK
	(*TraceID)(nil),         // 2: g.TraceID
	(*AgentStatus)(nil),     // 3: g.AgentStatus
	(*Trace)(nil),           // 4: g.Trace
	(*empty.Empty)(nil),     // 5: google.protobuf.Empty
}
var file_collector_proto_depIdxs = []int32{
	0, // 0: g.AgentStatus.status:type_name -> g.AgentStatus.Status
	4, // 1: g.Collector.SendTrace:input_type -> g.Trace
	5, // 2: g.Collector.SubscribeTraceID:input_type -> google.protobuf.Empty
	3, // 3: g.Collector.ConfirmFinish:input_type -> g.AgentStatus
	5, // 4: g.Collector.SendTrace:output_type -> google.protobuf.Empty
	2, // 5: g.Collector.SubscribeTraceID:output_type -> g.TraceID
	1, // 6: g.Collector.ConfirmFinish:output_type -> g.OK
	4, // [4:7] is the sub-list for method output_type
	1, // [1:4] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_collector_proto_init() }
func file_collector_proto_init() {
	if File_collector_proto != nil {
		return
	}
	file_trace_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_collector_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OK); i {
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
		file_collector_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TraceID); i {
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
		file_collector_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AgentStatus); i {
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
			RawDescriptor: file_collector_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_collector_proto_goTypes,
		DependencyIndexes: file_collector_proto_depIdxs,
		EnumInfos:         file_collector_proto_enumTypes,
		MessageInfos:      file_collector_proto_msgTypes,
	}.Build()
	File_collector_proto = out.File
	file_collector_proto_rawDesc = nil
	file_collector_proto_goTypes = nil
	file_collector_proto_depIdxs = nil
}
