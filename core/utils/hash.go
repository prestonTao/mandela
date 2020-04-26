package utils

import (
	"io"
	"os"

	"golang.org/x/crypto/sha3"
)

func Hash_SHA3_256(bs []byte) []byte {
	hash_sha3 := sha3.New256()
	hash_sha3.Write(bs)
	return hash_sha3.Sum(nil)
}

func Hash_SHA3_512(bs []byte) []byte {
	hash_sha3 := sha3.New512()
	hash_sha3.Write(bs)
	return hash_sha3.Sum(nil)
}

/*
	计算文件的hash值
*/
func FileSHA3_256(path string) ([]byte, error) {
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	hash_sha3 := sha3.New256()
	_, err = io.Copy(hash_sha3, file)
	if err != nil {
		return nil, err
	}
	return hash_sha3.Sum(nil), nil
}
