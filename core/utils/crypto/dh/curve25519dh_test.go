package dh

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"golang.org/x/crypto/curve25519"

	"golang.org/x/crypto/ed25519"
)

func TestDH(t *testing.T) {
	//生成A的密钥对
	keyA, _ := GenerateKeyPair()
	fmt.Println("打印A随机生成的私钥", hex.EncodeToString(keyA.PrivateKey[:]))
	//生成B的密钥对
	keyB, _ := GenerateKeyPair()
	fmt.Println("打印B随机生成的私钥", hex.EncodeToString(keyB.PrivateKey[:]))

	//用A的私钥和B的公钥计算A得到的协商密钥
	dbpA := NewDHPair(keyA.PrivateKey, keyB.PublicKey)
	AKey := KeyExchange(dbpA)
	aKeyStr := hex.EncodeToString(AKey[:])
	fmt.Println("A计算出来的协商密钥", aKeyStr)

	//用B的私钥和A的公钥计算B得到的协商密钥
	dbpB := NewDHPair(keyB.PrivateKey, keyA.PublicKey)
	BKey := KeyExchange(dbpB)
	bKeyStr := hex.EncodeToString(BKey[:])
	fmt.Println("B计算出来的协商密钥", bKeyStr)

	//-------------------------------
	//尝试使用ed25519生成的公钥私钥对，应用在DH中
	puk, prk, _ := ed25519.GenerateKey(rand.Reader)
	prkBs := prk.Public().([]byte)
	fmt.Println("ed25519生成的密钥对", hex.EncodeToString(prkBs), hex.EncodeToString(puk[:]))

	hash := sha256.New()
	hash.Write(prkBs)
	temp := hash.Sum(nil)

	var priKey [32]byte = [32]byte{temp[0], temp[1], temp[2], temp[3], temp[4],
		temp[5], temp[6], temp[7], temp[8], temp[9], temp[10], temp[11], temp[12],
		temp[13], temp[14], temp[15], temp[16], temp[17], temp[18], temp[19], temp[20],
		temp[21], temp[22], temp[23], temp[24], temp[25], temp[26],
		temp[27], temp[28], temp[29], temp[30], temp[31]}

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &priKey)
	fmt.Println("ed25519生成的密钥对", hex.EncodeToString(prkBs), hex.EncodeToString(pubKey[:]))

}
