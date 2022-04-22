package utils

import (
	"encoding/base64"
	"unsafe"
)

func EnBase64(bs []byte) string {
	return base64.StdEncoding.EncodeToString(bs)
}
func DenBase64(str string) (bs []byte, err error) {
	bs, err = base64.StdEncoding.DecodeString(str)
	return
}

/*
	把[]byte转换为string，只要性能，不在乎可读性
*/
func Bytes2string(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
