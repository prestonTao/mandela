package dh

import (
	"bytes"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/curve25519"
)

/*
	私钥和公钥
*/
type Key [32]byte

// 桁条接口符合性。
func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

// type PublicKey [32]byte

/*
	DH秘钥交换结构体
*/
type DHPair struct {
	privateKey Key
	publicKey  Key
}

func (this DHPair) GetPrivateKey() Key {
	return this.privateKey
}
func (this DHPair) GetPublicKey() Key {
	return this.publicKey
}

/*
	创建一个DH密钥对
*/
func NewDHPair(prk, puk Key) DHPair {
	return DHPair{
		privateKey: prk,
		publicKey:  puk,
	}
}

/*
	创建一个密钥交换对
*/
// func NewDHPair(prik Key, pubk Key) DHPair {
// 	dhp := DHPair{
// 		privateKey: prik,
// 		publicKey:  pubk,
// 	}
// 	return dhp
// }

/*
	公钥私钥对
*/
type KeyPair struct {
	PublicKey  Key
	PrivateKey Key
}

func (this KeyPair) GetPrivateKey() Key {
	return this.PrivateKey
}
func (this KeyPair) GetPublicKey() Key {
	return this.PublicKey
}

/*
	生成公钥私钥对
*/
func GenerateKeyPair(rand []byte) (KeyPair, error) {
	size := 32
	if len(rand) != size {
		//私钥长度不够
		return KeyPair{}, errors.New("Insufficient length of private key")
	}
	var priv [32]byte

	//用随机数填满私钥
	buf := bytes.NewBuffer(rand)
	n, err := buf.Read(priv[:])
	if n != size {
		//读取私钥长度不够
		return KeyPair{}, errors.New("Insufficient length to read private key")
	}
	if err != nil {
		return KeyPair{}, err
	}

	// _, err := rand.Reader.Read(priv[:])
	// if err != nil {
	// 	return KeyPair{}, err
	// }

	// priv[0] &= 248
	// priv[31] &= 127
	// priv[31] |= 64

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &priv)

	return KeyPair{
		PrivateKey: priv,
		PublicKey:  pubKey,
	}, nil

}

/*
	协商密钥
*/
func KeyExchange(dh DHPair) [32]byte {
	var (
		sharedSecret [32]byte
		priv         [32]byte = dh.GetPrivateKey()
		pub          [32]byte = dh.GetPublicKey()
	)
	curve25519.ScalarMult(&sharedSecret, &priv, &pub)
	return sharedSecret
}
