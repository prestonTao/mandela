package engine

import (
	"crypto/aes"
	"crypto/cipher"
	"math/big"
	"math/rand"
	"time"
)

//var IsCrypt bool = false //是否通道加密

//加密字符串
func Encrypt(key, srcByte []byte) ([]byte, error) {
	//	srcTempByte := srcByte
	fillByteLength := ((len(srcByte) + 15) / 16 * 16)
	fillByte := make([]byte, fillByteLength)
	copy(fillByte, srcByte)
	//	fillByte := append(srcByte, t...)

	//	fmt.Println(aes.BlockSize)
	//	var iv = key[:aes.BlockSize]
	encrypted := make([]byte, 0, len(fillByte))
	aesBlockEncrypter, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	//	aesEncrypter.XORKeyStream(encrypted, srcByte)
	oneEncrypted := make([]byte, 16)
	for i := 0; i < fillByteLength/16; i++ {
		aesBlockEncrypter.Encrypt(oneEncrypted, fillByte[i*16:(i+1)*16])
		encrypted = append(encrypted, oneEncrypted...)
	}
	//	aesBlockEncrypter.Decrypt()
	return encrypted, nil
}

//解密字符串
func Decrypt(key, src []byte, length int) (strDesc []byte, err error) {
	decrypted := make([]byte, 0, len(src))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	//	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	//	aesDecrypter.XORKeyStream(decrypted, src)
	oneDecrypted := make([]byte, 16)
	for i := 0; i < len(src)/16; i++ {
		aesBlockDecrypter.Decrypt(oneDecrypted, src[i*16:(i+1)*16])
		decrypted = append(decrypted, oneDecrypted...)
	}
	return decrypted[:length], nil
}

/*
	随机获取一个128位的key
*/
func RandKey128() []byte {
	min := rand.New(rand.NewSource(99))
	min.Seed(int64(time.Now().Nanosecond()))
	maxId := new(big.Int).Lsh(big.NewInt(1), 128)
	randInt := new(big.Int).Rand(min, maxId)
	bs := randInt.Bytes()
	if len(bs) < 16 {
		bs = append(randInt.Bytes(), bs...)
	}
	return bs[:16]
}
