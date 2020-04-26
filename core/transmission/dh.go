package transmission

import (
	"mandela/config"
	"mandela/core/engine/crypto"
	"mandela/core/utils"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/crypto/curve25519"
)

const (
	dhkey = "dhkey.key"
)

var (
	PPJSON PPJson
)

type PPJson struct {
	Pubk string
	Prik string
}

//初始化dhkey
func InitDh() error {
	utils.CheckCreateDir(config.Path_configDir)
	fileAbsPath := filepath.Join(config.Path_configDir, dhkey)
	exist, err := utils.PathExists(fileAbsPath)
	if err != nil {
		return err
	}
	if exist {
		bs, err := ioutil.ReadFile(fileAbsPath)
		if err != nil {
			return err
		}
		PPJSON = Parse(string(bs))
	} else {
		ppj := PPJson{}
		prik, pubk := CreatePrkPuk()
		ppj.Prik = prik
		ppj.Pubk = pubk
		utils.SaveJsonFile(fileAbsPath, ppj)
		PPJSON = ppj
	}
	return nil
}

//解析ppjson
func Parse(jsonstr string) PPJson {
	ppj := PPJson{}
	// json.Unmarshal([]byte(jsonstr), &ppj)
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(jsonstr)))
	decoder.UseNumber()
	err := decoder.Decode(&ppj)
	if err != nil {
		fmt.Println(err)
	}
	return ppj
}

//生成公私钥
func CreatePrkPuk() (prikStr, pubkStr string) {
	prik, pubk := genKeyPair()
	prikStr = hex.EncodeToString(prik[:])
	fmt.Println("prkA", len(prik), prikStr)

	pubkStr = hex.EncodeToString(pubk[:])
	fmt.Println("pukA", len(pubk), pubkStr)
	return
}

//获取密钥
func GetKey(prika, pubkb string) (string, error) {
	prikA, err := hex.DecodeString(prika)
	if err != nil {
		return "", err
	}
	pubkB, err := hex.DecodeString(pubkb)
	if err != nil {
		return "", err
	}
	var dhOutA, priA, pubB [32]byte
	copy(priA[:], prikA)
	copy(pubB[:], pubkB)
	curve25519.ScalarMult(&dhOutA, &priA, &pubB)
	//fmt.Println("协商密钥A", hex.EncodeToString(dhOutA[:]))
	return hex.EncodeToString(dhOutA[:]), nil
}

//byte to string
func Btos(bs []byte) string {
	return hex.EncodeToString(bs)
}

//string to byte
func Stob(bs string) []byte {
	rs, err := hex.DecodeString(bs)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return rs
}

//加密
func Encrypt(key, bs []byte) ([]byte, error) {
	crypter, err := crypto.NewCrypter("aes", key)
	if err != nil {
		return nil, err
	}
	rs, err := crypter.Encrypt(bs)
	if err != nil {
		return nil, err
	}
	return rs, nil
}

//解密
func Decrypt(key, rs []byte) ([]byte, error) {
	crypter, err := crypto.NewCrypter("aes", key)
	if err != nil {
		return nil, err
	}
	bs, err := crypter.Decrypt(rs)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
