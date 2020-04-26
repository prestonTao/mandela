package doubleratchet

import (
	"mandela/core/utils/crypto/dh"
	"encoding/binary"
	"fmt"
)

// message包含密文和加密的头。
type MessageHE struct {
	Header     []byte `json:"header"`
	Ciphertext []byte `json:"ciphertext"`
}

// 消息是双方交换的单个消息。
type Message struct {
	Header     MessageHeader `json:"header"`
	Ciphertext []byte        `json:"ciphertext"`
}

// 每个消息前面都有消息头。
type MessageHeader struct {
	// DHR是发送方当前的棘轮公钥。
	DH dh.Key `json:"dh"`

	// n是发送链中消息的编号。
	N uint32 `json:"n"`

	// pn是上一个发送链的长度。
	PN uint32 `json:"pn"`
}

// 以二进制格式对头文件进行编码。
func (mh MessageHeader) Encode() MessageEncHeader {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], mh.N)
	binary.LittleEndian.PutUint32(buf[4:8], mh.PN)
	return append(buf, mh.DH[:]...)
}

// messageencheader是消息头的二进制编码表示。
type MessageEncHeader []byte

// 从二进制编码表示中解码消息头。
func (mh MessageEncHeader) Decode() (MessageHeader, error) {
	// n (4 bytes) + pn (4 bytes) + dh (32 bytes)
	if len(mh) != 40 {
		return MessageHeader{}, fmt.Errorf("encoded message header must be 40 bytes, %d given", len(mh))
	}
	var dh dh.Key
	copy(dh[:], mh[8:40])
	return MessageHeader{
		DH: dh,
		N:  binary.LittleEndian.Uint32(mh[0:4]),
		PN: binary.LittleEndian.Uint32(mh[4:8]),
	}, nil
}
