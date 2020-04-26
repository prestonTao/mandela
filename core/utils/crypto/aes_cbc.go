package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"strconv"
)

/*
	加密
*/
func EncryptCBC(plantText, key, iv []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		//"VI长度错误(" + strconv.Itoa(len(iv)) + ")，aes cbc IV长度应该是" + strconv.Itoa(aes.BlockSize)
		return nil, errors.New("VI length error(" + strconv.Itoa(len(iv)) + ")，aes cbc IV length should be " + strconv.Itoa(aes.BlockSize))
	}
	block, err := aes.NewCipher(key) //选择加密算法
	if err != nil {
		return nil, err
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())

	blockModel := cipher.NewCBCEncrypter(block, iv)

	ciphertext := make([]byte, len(plantText))

	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext, nil
}

/*
	PKCS #7 填充字符串由一个字节序列组成，每个字节填充该字节序列的长度。
	下面的示例演示这些模式的工作原理。假定块长度为 8，数据长度为 9，则填充用八位字节数等于 7，数据等于 FF FF FF FF FF FF FF FF FF：
	数据： FF FF FF FF FF FF FF FF FF
	PKCS7 填充： FF FF FF FF FF FF FF FF FF 07 07 07 07 07 07 07
*/
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	if padding == 0 {
		return ciphertext
	}
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

/*
	解密
*/
func DecryptCBC(ciphertext, key, iv []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize {
		//"VI长度错误(" + strconv.Itoa(len(iv)) + ")，aes cbc IV长度应该是" + strconv.Itoa(aes.BlockSize)
		return nil, errors.New("VI length error(" + strconv.Itoa(len(iv)) + ")，aes cbc IV length should be" + strconv.Itoa(aes.BlockSize))
	}
	keyBytes := key
	block, err := aes.NewCipher(keyBytes) //选择加密算法
	if err != nil {
		return nil, err
	}
	blockModel := cipher.NewCBCDecrypter(block, iv)
	plantText := make([]byte, len(ciphertext))
	blockModel.CryptBlocks(plantText, ciphertext)
	return PKCS7UnPadding(plantText, block.BlockSize())
	// return plantText, nil
}

func PKCS7UnPadding(plantText []byte, blockSize int) ([]byte, error) {
	length := len(plantText)
	unpadding := int(plantText[length-1])
	if length < unpadding {
		return nil, errors.New("unpadding too long")
	}
	return plantText[:(length - unpadding)], nil
}
