package crypto

import (
	"crypto/sha256"
	"errors"
	"io"

	"golang.org/x/crypto/hkdf"
)

/*
	获取hkdf链编码
	@master    []byte    随机数
	@salt      []byte    盐
	@index     uint64    索引，棘轮数
*/
func GetHkdfChainCode(master, salt []byte, index uint64) (key, chainCode []byte, err error) {
	key, chainCode = master, salt
	for i := 0; i <= int(index); i++ {
		key, chainCode, err = hkdfChainCode(key, chainCode)
		if err != nil {
			return nil, nil, err
		}
		// fmt.Println("----", len(key), len(chainCode))
	}
	return
}

func hkdfChainCode(master, salt []byte) (key, chainCode []byte, err error) {
	hkdf := hkdf.New(sha256.New, master, salt, nil)

	keys := make([][]byte, 2)
	for i := 0; i < len(keys); i++ {
		keys[i] = make([]byte, 32)
		n, err := io.ReadFull(hkdf, keys[i])
		if n != len(keys[i]) {
			return nil, nil, errors.New("hkdf chain read hash fail")
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return keys[0], keys[1], nil
}
