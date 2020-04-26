package doubleratchet

import (
	"mandela/core/utils/crypto/dh"
)

// Crypto是对加密库的补充。
type Crypto interface {
	//创建新的dh密钥对
	GenerateDH() (DHPair, error)

	//计算公共密钥
	DH(dhPair DHPair, dhPub dh.Key) dh.Key

	//使用消息密钥 mk加密内容
	Encrypt(mk dh.Key, plaintext, ad []byte) (authCiphertext []byte)

	// Decrypt returns the AEAD decryption of ciphertext with message key mk.
	//通过消息密钥mk解密内容
	Decrypt(mk dh.Key, ciphertext, ad []byte) (plaintext []byte, err error)

	KDFer
}

//dh密钥对
type DHPair interface {
	GetPrivateKey() dh.Key
	GetPublicKey() dh.Key
}

// 键是任何32字节的键。它是为漂亮的十六进制输出而创建的。
// type Key [32]byte

// 桁条接口符合性。
// func (k Key) String() string {
// 	return hex.EncodeToString(k[:])
// }
