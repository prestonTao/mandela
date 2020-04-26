// Package crypto provides des 加解密
package crypto

import (
	"crypto/cipher"
	"crypto/des"
	"errors"
)

type DESCrypter struct {
	encryptBlock cipher.BlockMode
	decryptBlock cipher.BlockMode
}

func newDES(key []byte) (Crypter, error) {

	block, err := des.NewCipher(key[0:des.BlockSize])
	if err != nil {
		return nil, err
	}

	cpt := &AESCrypter{
		encryptBlock: cipher.NewCBCEncrypter(block, []byte(key)[0:des.BlockSize]),
		decryptBlock: cipher.NewCBCDecrypter(block, []byte(key)[0:des.BlockSize]),
	}

	return cpt, nil

}

func (cpt *DESCrypter) Encrypt(src []byte) ([]byte, error) {
	src = PKCS5Padding(src, des.BlockSize)

	if len(src)%des.BlockSize != 0 {
		//填充错误
		return nil, errors.New("Fill error")
	}

	dst := make([]byte, len(src))

	cpt.encryptBlock.CryptBlocks(dst, src)

	return dst, nil
}

func (cpt *DESCrypter) Decrypt(src []byte) ([]byte, error) {

	dst := make([]byte, len(src))

	cpt.decryptBlock.CryptBlocks(dst, src)

	dst = PKCS5UnPadding(dst)

	return dst, nil
}
