package x3dh

import (
	"mandela/core/utils/crypto/dh"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"testing"

	"golang.org/x/crypto/hkdf"
)

func TestX3DH(t *testing.T) {
	IPKA, _ := dh.GenerateKeyPair()
	SPKA, _ := dh.GenerateKeyPair()
	OPKA, _ := dh.GenerateKeyPair()

	IPKB, _ := dh.GenerateKeyPair()
	SPKB, _ := dh.GenerateKeyPair()
	OPKB, _ := dh.GenerateKeyPair()

	DHA1 := dh.KeyExchange(dh.NewDHPair(IPKA.PrivateKey, SPKB.PublicKey))
	DHA2 := dh.KeyExchange(dh.NewDHPair(OPKA.PrivateKey, IPKB.PublicKey))
	DHA3 := dh.KeyExchange(dh.NewDHPair(OPKA.PrivateKey, SPKB.PublicKey))
	DHA4 := dh.KeyExchange(dh.NewDHPair(OPKA.PrivateKey, OPKB.PublicKey))

	DHA := bytes.NewBuffer(DHA1[:])
	DHA.Write(DHA2[:])
	DHA.Write(DHA3[:])
	DHA.Write(DHA4[:])

	fmt.Println("A的长版密钥", hex.EncodeToString(DHA.Bytes()))

	kdf := hkdf.New(sha256.New, DHA.Bytes(), nil, nil)
	out := make([]byte, 32)
	io.ReadFull(kdf, out)

	fmt.Println("A的协商密钥", hex.EncodeToString(out[:]))

	DHB1 := dh.KeyExchange(dh.NewDHPair(IPKB.PrivateKey, SPKA.PublicKey))
	DHB2 := dh.KeyExchange(dh.NewDHPair(OPKB.PrivateKey, IPKA.PublicKey))
	DHB3 := dh.KeyExchange(dh.NewDHPair(OPKB.PrivateKey, SPKA.PublicKey))
	DHB4 := dh.KeyExchange(dh.NewDHPair(OPKB.PrivateKey, OPKA.PublicKey))

	DHB := bytes.NewBuffer(DHB1[:])
	DHB.Write(DHB2[:])
	DHB.Write(DHB3[:])
	DHB.Write(DHB4[:])

	fmt.Println("B的长版密钥", hex.EncodeToString(DHB.Bytes()))

	kdf = hkdf.New(sha256.New, DHB.Bytes(), nil, nil)
	out = make([]byte, 32)
	io.ReadFull(kdf, out)
	fmt.Println("B的协商密钥", hex.EncodeToString(out[:]))

	// //用A的私钥和B的公钥计算A得到的协商密钥
	// dbpA := NewDHPair(keyA.PrivateKey, keyB.PublicKey)
	// AKey := KeyExchange(dbpA)
	// aKeyStr := hex.EncodeToString(AKey[:])
	// fmt.Println("A计算出来的协商密钥", aKeyStr)

	// //用B的私钥和A的公钥计算B得到的协商密钥
	// dbpB := NewDHPair(keyB.PrivateKey, keyA.PublicKey)
	// BKey := KeyExchange(dbpB)
	// bKeyStr := hex.EncodeToString(BKey[:])
	// fmt.Println("B计算出来的协商密钥", bKeyStr)
}
