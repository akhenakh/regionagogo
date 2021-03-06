// Code generated by protoc-gen-go. DO NOT EDIT.
// source: geostore.proto

/*
Package geostore is a generated protocol buffer package.

It is generated from these files:
	geostore.proto

It has these top-level messages:
	FenceStorage
	CPoint
	FenceCover
*/
package geostore

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// FenceStorage is used to represent a Fence in storage
type FenceStorage struct {
	Points []*CPoint         `protobuf:"bytes,1,rep,name=points" json:"points,omitempty"`
	Data   map[string]string `protobuf:"bytes,2,rep,name=data" json:"data,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *FenceStorage) Reset()                    { *m = FenceStorage{} }
func (m *FenceStorage) String() string            { return proto.CompactTextString(m) }
func (*FenceStorage) ProtoMessage()               {}
func (*FenceStorage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *FenceStorage) GetPoints() []*CPoint {
	if m != nil {
		return m.Points
	}
	return nil
}

func (m *FenceStorage) GetData() map[string]string {
	if m != nil {
		return m.Data
	}
	return nil
}

// CPoint represent a coordinates lat & lng
type CPoint struct {
	Lat float32 `protobuf:"fixed32,1,opt,name=lat" json:"lat,omitempty"`
	Lng float32 `protobuf:"fixed32,2,opt,name=lng" json:"lng,omitempty"`
}

func (m *CPoint) Reset()                    { *m = CPoint{} }
func (m *CPoint) String() string            { return proto.CompactTextString(m) }
func (*CPoint) ProtoMessage()               {}
func (*CPoint) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CPoint) GetLat() float32 {
	if m != nil {
		return m.Lat
	}
	return 0
}

func (m *CPoint) GetLng() float32 {
	if m != nil {
		return m.Lng
	}
	return 0
}

// FenceCover is used to store an s2 coverage of a fence
type FenceCover struct {
	Cellunion []uint64 `protobuf:"varint,1,rep,packed,name=cellunion" json:"cellunion,omitempty"`
}

func (m *FenceCover) Reset()                    { *m = FenceCover{} }
func (m *FenceCover) String() string            { return proto.CompactTextString(m) }
func (*FenceCover) ProtoMessage()               {}
func (*FenceCover) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *FenceCover) GetCellunion() []uint64 {
	if m != nil {
		return m.Cellunion
	}
	return nil
}

func init() {
	proto.RegisterType((*FenceStorage)(nil), "geostore.FenceStorage")
	proto.RegisterType((*CPoint)(nil), "geostore.CPoint")
	proto.RegisterType((*FenceCover)(nil), "geostore.FenceCover")
}

func init() { proto.RegisterFile("geostore.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 220 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x90, 0xc1, 0x4a, 0x03, 0x31,
	0x10, 0x86, 0x49, 0x5a, 0x17, 0x77, 0x14, 0x29, 0xc1, 0x43, 0x10, 0x0f, 0xcb, 0x9e, 0x16, 0x91,
	0x3d, 0xa8, 0xa0, 0x78, 0xad, 0x7a, 0x96, 0xf8, 0x04, 0xb1, 0x0e, 0xa1, 0x18, 0x32, 0x25, 0x9d,
	0x16, 0xfa, 0x44, 0xbe, 0xa6, 0x64, 0x56, 0x77, 0x7b, 0x9b, 0xf9, 0xe6, 0xff, 0xc2, 0x64, 0xe0,
	0x22, 0x20, 0x6d, 0x99, 0x32, 0xf6, 0x9b, 0x4c, 0x4c, 0xe6, 0xf4, 0xbf, 0x6f, 0x7f, 0x14, 0x9c,
	0xbf, 0x61, 0x5a, 0xe1, 0x07, 0x53, 0xf6, 0x01, 0x4d, 0x07, 0xd5, 0x86, 0xd6, 0x89, 0xb7, 0x56,
	0x35, 0xb3, 0xee, 0xec, 0x6e, 0xd1, 0x8f, 0xee, 0xf2, 0xbd, 0x0c, 0xdc, 0xdf, 0xdc, 0x3c, 0xc0,
	0xfc, 0xcb, 0xb3, 0xb7, 0x5a, 0x72, 0xcd, 0x94, 0x3b, 0x7e, 0xaf, 0x7f, 0xf1, 0xec, 0x5f, 0x13,
	0xe7, 0x83, 0x93, 0xf4, 0xd5, 0x23, 0xd4, 0x23, 0x32, 0x0b, 0x98, 0x7d, 0xe3, 0xc1, 0xaa, 0x46,
	0x75, 0xb5, 0x2b, 0xa5, 0xb9, 0x84, 0x93, 0xbd, 0x8f, 0x3b, 0xb4, 0x5a, 0xd8, 0xd0, 0x3c, 0xeb,
	0x27, 0xd5, 0xde, 0x42, 0x35, 0x2c, 0x50, 0xac, 0xe8, 0x59, 0x2c, 0xed, 0x4a, 0x29, 0x24, 0x05,
	0x71, 0x0a, 0x49, 0xa1, 0xbd, 0x01, 0x90, 0x35, 0x96, 0xb4, 0xc7, 0x6c, 0xae, 0xa1, 0x5e, 0x61,
	0x8c, 0xbb, 0xb4, 0xa6, 0x24, 0xff, 0x9a, 0xbb, 0x09, 0x7c, 0x56, 0x72, 0x94, 0xfb, 0xdf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0xa4, 0x4b, 0x2e, 0x0c, 0x26, 0x01, 0x00, 0x00,
}
