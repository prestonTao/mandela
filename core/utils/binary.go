package utils

import (
	"bytes"
	"encoding/binary"
)

//uint64转byte
func Uint64ToBytes(n uint64) []byte {
	bytesBuffer := bytes.NewBuffer(nil)
	binary.Write(bytesBuffer, binary.LittleEndian, n)
	return bytesBuffer.Bytes()
}

//byte转uint64
func BytesToUint64(b []byte) uint64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint64
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return tmp
}

//int64转byte
func Int64ToBytes(n int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, n)
	return bytesBuffer.Bytes()
}

//byte转int64
func BytesToInt64(b []byte) int64 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp int64
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return tmp
}

//uint64转byte
func Uint32ToBytes(n uint32) []byte {
	bytesBuffer := bytes.NewBuffer(nil)
	binary.Write(bytesBuffer, binary.LittleEndian, n)
	return bytesBuffer.Bytes()
}

//byte转uint64
func BytesToUint32(b []byte) uint32 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint32
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return tmp
}

/*
	uint16转byte
*/
func Uint16ToBytes(n uint16) []byte {
	bytesBuffer := bytes.NewBuffer(nil)
	binary.Write(bytesBuffer, binary.LittleEndian, n)
	return bytesBuffer.Bytes()
}

/*
	byte转uint16
*/
func BytesToUint16(b []byte) uint16 {
	bytesBuffer := bytes.NewBuffer(b)
	var tmp uint16
	binary.Read(bytesBuffer, binary.LittleEndian, &tmp)
	return tmp
}
