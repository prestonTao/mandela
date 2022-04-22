package engine

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

// var defaultGetPacket GetPacket = RecvPackage
const (
	max_size = 1024 * 1024 * 1024 //一个包内容最大容量
)

var Netid uint32 = 2 //网络版本id，避免冲突

type GetPacket func(cache *[]byte, index *uint32) (packet *Packet, n int)
type GetPacketBytes func(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) *[]byte

type Packet struct {
	//	From byte //数据包来源分别为0（终端）、1（内网服务器）、2（master server）、4(chunk server)，占7位
	//	E     byte   //是否加密串，占一位，和cmd_from共用一字节
	MsgID uint64 //
	//	Errorcode byte   //数据包处理结果
	Size uint64 //数据包长度，包含头部4字节
	//	RealSize  uint16 //数据包不包含头部的实际长度，并非简单的packet len – 24
	//	Crypt_key []byte //共16字节的key
	// temp     []byte
	Data     []byte
	Dataplus []byte //未加密部分数据分开
	Session  Session
	// WaitChan chan bool
	// IsWait   bool
}

// func (this *Packet) Wait(second time.Duration) bool {
// 	select {
// 	case <-this.WaitChan:
// 		return true
// 	case <-time.After(second):
// 		return false
// 	}
// }

// func (this *Packet) FinishWait() {
// 	this.WaitChan <- true
// }

/*
	系统默认的消息接收并转化为Packet的方法
	一个packet包括包头和包体，保证在接收到包头后两秒钟内接收到包体，否则线程会一直阻塞
	因此，引入了超时机制
*/
/*
/*  0               1               2               3
/*  0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7 0 1 2 3 4 5 6 7
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/* |                        packet_len(8字节)
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/*                                                                 |
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/* |                        data_len(8字节)
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/*                                                                 |
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/* |                        net_id(4字节)                           |
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/* |                        msg_id(8字节)                           |
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/*                                                                 |
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/* |					    data(data_len - 12)                    |
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
/* |					    data_plus...					       |
/* +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
*/
func RecvPackage(conn net.Conn) (*Packet, error) {
	defer PrintPanicStack()
	packet := new(Packet)
	//先读包长度
	cache := make([]byte, 8)
	index := uint64(0)
	for index < 8 {
		// cache := make([]byte, 8)
		n, err := conn.Read(cache[index:8])
		if err != nil {
			return nil, err
		}
		index += uint64(n)
		// packet.temp = append(packet.temp, cache[:n]...)
		// Log.Info("read 111 %s", hex.EncodeToString(cache[:n]))

	}
	// Log.Info("read 111 %s", hex.EncodeToString(cache[:]))
	// for len(packet.temp) < 16 {
	// 	cache := make([]byte, 16)
	// 	n, err := conn.Read(cache)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	packet.temp = append(packet.temp, cache[:n]...)
	// 	Log.Info("read 111 %s", hex.EncodeToString(cache[:n]))
	// }

	//解析包长度
	packet.Size = binary.LittleEndian.Uint64(cache[:8])
	// Log.Info("包大小 %s %d", hex.EncodeToString(cache[:8]), packet.Size)

	if packet.Size > max_size {
		//包头错误 包长度大于最大值
		return nil, errors.New("Packet header error packet length greater than maximum")
	}
	if packet.Size < 8+8+4+8 {
		//包头错误 包长度太小
		return nil, errors.New("Packet header error packet length too small")
	}

	//读取指定长度的包大小
	cache = make([]byte, packet.Size-8)
	index = uint64(0)
	for index < packet.Size-8 {
		n, err := conn.Read(cache[index : packet.Size-8])
		if err != nil {
			return nil, err
		}
		index += uint64(n)
		// packet.temp = append(packet.temp, cache[:n]...)
		// Log.Info("read 222 %d %d %s", uint64(len(packet.temp)), packet.Size, hex.EncodeToString(cache[:n]))
	}

	// Log.Info("read 222 %d %s", packet.Size, hex.EncodeToString(cache[:]))

	// for uint64(len(packet.temp)) < packet.Size {
	// 	cache := make([]byte, packet.Size-(16))
	// 	n, err := conn.Read(cache)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	packet.temp = append(packet.temp, cache[:n]...)
	// 	Log.Info("read 222 %d %d %s", uint64(len(packet.temp)), packet.Size, hex.EncodeToString(cache[:n]))
	// }

	dataSize := binary.LittleEndian.Uint64(cache[:8])
	// Log.Info("dataSize %s %d", hex.EncodeToString(cache[:8]), dataSize)
	data := cache[8 : dataSize+8]
	// Log.Info("data %s %d", hex.EncodeToString(data), dataSize)
	//TODO 这里执行解密

	netid := binary.LittleEndian.Uint32(data[:4])
	if netid != uint32(Netid) {
		//网络号错误
		return nil, errors.New("netid error")
	}
	// Log.Info("netid %s %d", hex.EncodeToString(data[:4]), netid)
	packet.MsgID = binary.LittleEndian.Uint64(data[4 : 4+8])
	// Log.Info("MsgID %s %d", hex.EncodeToString(data[4:4+8]), packet.MsgID)

	packet.Data = data[4+8:]
	// Log.Info("Data %s", hex.EncodeToString(packet.Data))
	packet.Dataplus = cache[dataSize+8 : packet.Size-8]
	// Log.Info("Dataplus %s", hex.EncodeToString(cache[dataSize+8:packet.Size-8]))

	//	NLog.Info(LOG_file, "RecvPackage size:%d;data长度:%d;dataplus长度:%d", packet.Size, dataSize, len(packet.Dataplus))

	//--------以下代码避免内存溢出----------
	// oldtemp := packet.temp
	// packet.temp = make([]byte, uint64(len(oldtemp))-packet.Size)
	// //	fmt.Println("packet  ", uint64(len(oldtemp)), packet.Size)
	// if uint64(len(oldtemp))-packet.Size > 0 {
	// 	copy(packet.temp, oldtemp[packet.Size:])
	// }
	return packet, nil
}

func MarshalPacket(msgID uint64, data, dataplus *[]byte) *[]byte {
	//	newCryKey := RandKey128()
	//	dataSize := 0
	//	if data != nil {
	//		dataSize = len(*data)
	//	}
	dataplusSize := 0
	if dataplus != nil {
		dataplusSize = len(*dataplus)
	}
	dataBuf := bytes.NewBuffer([]byte{})

	binary.Write(dataBuf, binary.LittleEndian, uint32(Netid))
	binary.Write(dataBuf, binary.LittleEndian, uint64(msgID))
	if data != nil {
		dataBuf.Write(*data)
	}

	//TODO 对dataBuf加密
	bs := dataBuf.Bytes()
	buf := bytes.NewBuffer([]byte{})
	totalSize := uint64(8 + 8 + len(bs) + dataplusSize)
	binary.Write(buf, binary.LittleEndian, totalSize)
	// Log.Info("打包头大小 %d 字节 %s", totalSize, hex.EncodeToString(buf.Bytes()))
	binary.Write(buf, binary.LittleEndian, uint64(len(bs)))
	buf.Write(bs)
	if dataplus != nil {
		buf.Write(*dataplus)
	}
	bs = buf.Bytes()
	//	NLog.Info(LOG_file, "MarshalPacket size:%d;data长度:%d;dataplus长度:%d", len(bs), len(dataBuf.Bytes()), dataplusSize)
	return &bs
}

///*
//	系统默认的消息接收并转化为Packet的方法
//	一个packet包括包头和包体，保证在接收到包头后两秒钟内接收到包体，否则线程会一直阻塞
//	因此，引入了超时机制
//*/
//func RecvPackage(conn net.Conn, packet *Packet) error {
//	// fmt.Println("packet   11111", *index, (*cache))
//	defer PrintPanicStack()
//	if len(packet.temp) < 16 {
//		cache := make([]byte, 16)
//		n, err := conn.Read(cache)
//		//	fmt.Println(n, err != nil)
//		if err != nil {
//			return err
//		}
//		packet.temp = append(packet.temp, cache[:n]...)
//	}

//	packet.Size = binary.LittleEndian.Uint64(packet.temp[:8])
//	packet.MsgID = binary.LittleEndian.Uint64(packet.temp[8:16])

//	if packet.Size < 16 {
//		return errors.New("包头错误")
//	}

//	for uint64(len(packet.temp)) < packet.Size {
//		cache := make([]byte, packet.Size-16)
//		n, err := conn.Read(cache)
//		if err != nil {
//			Log.Debug("err %v %d %d", err, n, uint64(n))
//			return err
//		}
//		packet.temp = append(packet.temp, cache[:n]...)
//	}
//	packet.Data = packet.temp[16:packet.Size]

//	//	packet.temp = make([]byte, uint64(len(packet.temp))-packet.Size)
//	//	if uint64(len(packet.temp))-packet.Size != 0 {
//	//		copy(packet.temp, packet.temp[packet.Size:])
//	//	}

//	oldtemp := packet.temp
//	packet.temp = make([]byte, uint64(len(oldtemp))-packet.Size)
//	//	fmt.Println("packet  ", uint64(len(oldtemp)), packet.Size)
//	if uint64(len(oldtemp))-packet.Size > 0 {
//		copy(packet.temp, oldtemp[packet.Size:])
//	}

//	return nil
//}

//func MarshalPacket(msgID uint64, data, dataplus *[]byte) *[]byte {
//	//	newCryKey := RandKey128()
//	if data == nil || len(*data) <= 0 {
//		buf := bytes.NewBuffer([]byte{})
//		binary.Write(buf, binary.LittleEndian, uint64(16))
//		binary.Write(buf, binary.LittleEndian, msgID)
//		bs := buf.Bytes()
//		return &bs
//	}

//	bodyBytes := *data
//	buf := bytes.NewBuffer([]byte{})
//	binary.Write(buf, binary.LittleEndian, uint64(len(bodyBytes)+16))
//	binary.Write(buf, binary.LittleEndian, msgID)
//	buf.Write(bodyBytes)
//	bs := buf.Bytes()
//	return &bs
//}

func cry(in []byte) []byte {
	i := 0
	tmpBuf := make([]byte, 128)
	for i < len(in) {
		if i+1 < len(in) {
			tmpBuf[i] = in[i+1]
			tmpBuf[i+1] = in[i]
		} else {
			tmpBuf[i] = in[i]
		}
		i += 2
	}
	out := make([]byte, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = tmpBuf[i] & 0x01
		out[i] = tmpBuf[i] & 0x0f
		out[i] <<= 4
		out[i] |= ((tmpBuf[i] & 0xf0) >> 4)
	}
	return out
}
