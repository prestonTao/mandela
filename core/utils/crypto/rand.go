package crypto

import (
	"crypto/rand"
	"io"
)

/*
	生成32字节（256位）的随机数
*/
func Rand32Byte() ([32]byte, error) {
	k := [32]byte{}
	key := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return k, err
	}
	for i, _ := range k {
		k[i] = key[i]
	}
	return k, nil
}

/*
	生成16字节（128位）的随机数
*/
func Rand16Byte() ([16]byte, error) {
	k := [16]byte{}
	key := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return k, err
	}
	for i, _ := range k {
		k[i] = key[i]
	}
	return k, nil
}
