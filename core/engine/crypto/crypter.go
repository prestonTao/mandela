// Package crypto provides 生成加解密算法
package crypto

import (
	"errors"
)

type Crypter interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)
}

func NewCrypter(kind string, key []byte) (Crypter, error) {

	switch kind {
	case "des":
		cpt, err := newDES(key)
		if err != nil {
			return nil, err
		}

		return cpt, nil

	case "aes":
		cpt, err := newAES(key)
		if err != nil {
			return nil, err
		}

		return cpt, nil

	default:
		//未找到对应种类的加密器
		return nil, errors.New("No corresponding kind of encryptor found")
	}

	//
	return nil, errors.New("No corresponding kind of encryptor found")
}
