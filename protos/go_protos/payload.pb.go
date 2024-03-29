// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: payload.proto

//包名，通过protoc生成时go文件时

package go_protos

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type CommunitySign struct {
	Type                 uint64   `protobuf:"varint,1,opt,name=Type,proto3" json:"Type,omitempty"`
	StartHeight          uint64   `protobuf:"varint,2,opt,name=StartHeight,proto3" json:"StartHeight,omitempty"`
	EndHeight            uint64   `protobuf:"varint,3,opt,name=EndHeight,proto3" json:"EndHeight,omitempty"`
	Rand                 uint64   `protobuf:"varint,4,opt,name=Rand,proto3" json:"Rand,omitempty"`
	Puk                  []byte   `protobuf:"bytes,5,opt,name=Puk,proto3" json:"Puk,omitempty"`
	Sign                 []byte   `protobuf:"bytes,6,opt,name=Sign,proto3" json:"Sign,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CommunitySign) Reset()         { *m = CommunitySign{} }
func (m *CommunitySign) String() string { return proto.CompactTextString(m) }
func (*CommunitySign) ProtoMessage()    {}
func (*CommunitySign) Descriptor() ([]byte, []int) {
	return fileDescriptor_678c914f1bee6d56, []int{0}
}
func (m *CommunitySign) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *CommunitySign) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_CommunitySign.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *CommunitySign) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommunitySign.Merge(m, src)
}
func (m *CommunitySign) XXX_Size() int {
	return m.Size()
}
func (m *CommunitySign) XXX_DiscardUnknown() {
	xxx_messageInfo_CommunitySign.DiscardUnknown(m)
}

var xxx_messageInfo_CommunitySign proto.InternalMessageInfo

func (m *CommunitySign) GetType() uint64 {
	if m != nil {
		return m.Type
	}
	return 0
}

func (m *CommunitySign) GetStartHeight() uint64 {
	if m != nil {
		return m.StartHeight
	}
	return 0
}

func (m *CommunitySign) GetEndHeight() uint64 {
	if m != nil {
		return m.EndHeight
	}
	return 0
}

func (m *CommunitySign) GetRand() uint64 {
	if m != nil {
		return m.Rand
	}
	return 0
}

func (m *CommunitySign) GetPuk() []byte {
	if m != nil {
		return m.Puk
	}
	return nil
}

func (m *CommunitySign) GetSign() []byte {
	if m != nil {
		return m.Sign
	}
	return nil
}

func init() {
	proto.RegisterType((*CommunitySign)(nil), "go_protos.CommunitySign")
}

func init() { proto.RegisterFile("payload.proto", fileDescriptor_678c914f1bee6d56) }

var fileDescriptor_678c914f1bee6d56 = []byte{
	// 184 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2d, 0x48, 0xac, 0xcc,
	0xc9, 0x4f, 0x4c, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4c, 0xcf, 0x8f, 0x07, 0xb3,
	0x8a, 0x95, 0xe6, 0x32, 0x72, 0xf1, 0x3a, 0xe7, 0xe7, 0xe6, 0x96, 0xe6, 0x65, 0x96, 0x54, 0x06,
	0x67, 0xa6, 0xe7, 0x09, 0x09, 0x71, 0xb1, 0x84, 0x54, 0x16, 0xa4, 0x4a, 0x30, 0x2a, 0x30, 0x6a,
	0xb0, 0x04, 0x81, 0xd9, 0x42, 0x0a, 0x5c, 0xdc, 0xc1, 0x25, 0x89, 0x45, 0x25, 0x1e, 0xa9, 0x99,
	0xe9, 0x19, 0x25, 0x12, 0x4c, 0x60, 0x29, 0x64, 0x21, 0x21, 0x19, 0x2e, 0x4e, 0xd7, 0xbc, 0x14,
	0xa8, 0x3c, 0x33, 0x58, 0x1e, 0x21, 0x00, 0x32, 0x33, 0x28, 0x31, 0x2f, 0x45, 0x82, 0x05, 0x62,
	0x26, 0x88, 0x2d, 0x24, 0xc0, 0xc5, 0x1c, 0x50, 0x9a, 0x2d, 0xc1, 0xaa, 0xc0, 0xa8, 0xc1, 0x13,
	0x04, 0x62, 0x82, 0x54, 0x81, 0x5c, 0x20, 0xc1, 0x06, 0x16, 0x02, 0xb3, 0x9d, 0x64, 0x4f, 0x3c,
	0x92, 0x63, 0xbc, 0xf0, 0x48, 0x8e, 0xf1, 0xc1, 0x23, 0x39, 0xc6, 0x19, 0x8f, 0xe5, 0x18, 0xa2,
	0xb8, 0xf5, 0xf4, 0xe1, 0xce, 0x4f, 0x62, 0x03, 0xd3, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x32, 0x26, 0x84, 0x5a, 0xe1, 0x00, 0x00, 0x00,
}

func (m *CommunitySign) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *CommunitySign) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *CommunitySign) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Sign) > 0 {
		i -= len(m.Sign)
		copy(dAtA[i:], m.Sign)
		i = encodeVarintPayload(dAtA, i, uint64(len(m.Sign)))
		i--
		dAtA[i] = 0x32
	}
	if len(m.Puk) > 0 {
		i -= len(m.Puk)
		copy(dAtA[i:], m.Puk)
		i = encodeVarintPayload(dAtA, i, uint64(len(m.Puk)))
		i--
		dAtA[i] = 0x2a
	}
	if m.Rand != 0 {
		i = encodeVarintPayload(dAtA, i, uint64(m.Rand))
		i--
		dAtA[i] = 0x20
	}
	if m.EndHeight != 0 {
		i = encodeVarintPayload(dAtA, i, uint64(m.EndHeight))
		i--
		dAtA[i] = 0x18
	}
	if m.StartHeight != 0 {
		i = encodeVarintPayload(dAtA, i, uint64(m.StartHeight))
		i--
		dAtA[i] = 0x10
	}
	if m.Type != 0 {
		i = encodeVarintPayload(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintPayload(dAtA []byte, offset int, v uint64) int {
	offset -= sovPayload(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *CommunitySign) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovPayload(uint64(m.Type))
	}
	if m.StartHeight != 0 {
		n += 1 + sovPayload(uint64(m.StartHeight))
	}
	if m.EndHeight != 0 {
		n += 1 + sovPayload(uint64(m.EndHeight))
	}
	if m.Rand != 0 {
		n += 1 + sovPayload(uint64(m.Rand))
	}
	l = len(m.Puk)
	if l > 0 {
		n += 1 + l + sovPayload(uint64(l))
	}
	l = len(m.Sign)
	if l > 0 {
		n += 1 + l + sovPayload(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovPayload(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPayload(x uint64) (n int) {
	return sovPayload(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *CommunitySign) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPayload
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: CommunitySign: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: CommunitySign: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartHeight", wireType)
			}
			m.StartHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StartHeight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EndHeight", wireType)
			}
			m.EndHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EndHeight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Rand", wireType)
			}
			m.Rand = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Rand |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Puk", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPayload
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPayload
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Puk = append(m.Puk[:0], dAtA[iNdEx:postIndex]...)
			if m.Puk == nil {
				m.Puk = []byte{}
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sign", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthPayload
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthPayload
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sign = append(m.Sign[:0], dAtA[iNdEx:postIndex]...)
			if m.Sign == nil {
				m.Sign = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPayload(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthPayload
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthPayload
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipPayload(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPayload
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPayload
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthPayload
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPayload
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPayload
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPayload        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPayload          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPayload = fmt.Errorf("proto: unexpected end of group")
)
