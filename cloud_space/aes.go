/**
私有文件加解密
*/
package cloud_space

import (
	"mandela/config"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	blocksize = 16
	size      = 32 //AES-256
)

var (
	fileprik []byte
)

func init() {
	prikname := filepath.Join(config.Path_configDir, fileprikname)
	//判断密钥是否存在，如果存在则读取，不存在则创建
	ok, err := utils.PathExists(prikname)
	if err != nil {
		fmt.Println(err)
		return
	}
	if ok {
		file, err := ioutil.ReadFile(prikname)
		if err != nil {
			fmt.Println(err)
		}
		fileprik = file
		return
	}
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		fmt.Println(err)
	}
	prk, err := utils.MarshalPrikey(priv)
	if err != nil {
		fmt.Println(err)
		return
	}
	fi, err := os.OpenFile(prikname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	fi.Write(prk)
	fi.Close()
	fileprik = prk
}

//获取私钥
func getPrikey() ([]byte, error) {
	return fileprik, nil
}

//加密
func Encrypt(text []byte) ([]byte, error) {
	prk, err := getPrikey()
	if err != nil {
		return nil, err
	}
	prks := sha256.Sum256(prk)
	retext, err := crypto.EncryptCBC(text, prks[0:size], prk[0:blocksize])
	return retext, err
}

//解密
func Decrypt(text []byte) ([]byte, error) {
	prk, err := getPrikey()
	if err != nil {
		return nil, err
	}
	prks := sha256.Sum256(prk)
	retext, err := crypto.DecryptCBC(text, prks[0:size], prk[0:blocksize])
	if err != nil {
		err = config.ERROR_password_fail
	}
	return retext, err
}
