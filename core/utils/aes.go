package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

/*
	对私钥加密
*/
func EncryptPrk(prkbs, key []byte) []byte {
	bss := bytes.Split(prkbs, []byte("\n"))
	bss = bss[:len(bss)-1]
	begin := string(bss[0])
	end := string(bss[len(bss)-1])
	//	fmt.Println("---1", string(bss[len(bss)-1]))
	//	for _, one := range bss {
	//		fmt.Println(len(one), string(one))
	//	}
	bs := bytes.Join(bss[1:len(bss)-1], nil)
	//	fmt.Println("---2", string(bss[len(bss)-1]))
	//	fmt.Println("加密前长度", len(bs), string(bs))

	bs, _ = Encrypt(bs, key)
	//	fmt.Println("加密后长度", len(bs), base64.StdEncoding.EncodeToString(bs))
	//	fmt.Println(base64.StdEncoding.EncodeToString(bs))
	newbs := []byte(base64.StdEncoding.EncodeToString(bs))
	out := append([]byte(begin), []byte("\n")...)
	//	fmt.Println("---3", string(bss[len(bss)-1]))
	for len(newbs) > 64 {
		//		fmt.Println("---4", string(end))
		out = append(out, newbs[:64]...)
		out = append(out, []byte("\n")...)
		newbs = newbs[64:]
	}
	out = append(out, newbs...)
	out = append(out, []byte("\n")...)
	out = append(out, end...)
	//	fmt.Println("---", string(bss[len(bss)-1]))

	//	bs, _ = Decrypt(bs, key)
	//	fmt.Println("解密后长度", len(bs))
	//	fmt.Println(string(bs))

	return out
}

func Encrypt(plantText, key []byte) ([]byte, error) {
	key = Key16Padding(key)
	//	fmt.Println("key", key)
	block, err := aes.NewCipher(key) //选择加密算法
	if err != nil {
		return nil, err
	}
	plantText = PKCS7Padding(plantText, block.BlockSize())

	blockModel := cipher.NewCBCEncrypter(block, key)

	ciphertext := make([]byte, len(plantText))

	blockModel.CryptBlocks(ciphertext, plantText)
	return ciphertext, nil
}

/*
	将加密秘钥填补成16字节长度
*/
func Key16Padding(key []byte) []byte {
	//	bs := []byte(key)
	if len(key) >= 16 {
		return key
	}
	newkey := make([]byte, 16)
	copy(newkey, key)
	return newkey
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
