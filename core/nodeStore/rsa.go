package nodeStore

//import (
//	"crypto/rand"
//	"crypto/rsa"
//	"crypto/x509"
//	"encoding/pem"
//	"errors"
//	//	"fmt"
//	"io/ioutil"
//	"os"
//	"path/filepath"
//)

//var PrivateKey, PublicKey []byte

//func GenRsaKey(bits int) error {
//	// 生成私钥文件
//	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
//	if err != nil {
//		return err
//	}
//	derStream := x509.MarshalPKCS1PrivateKey(privateKey)
//	block := &pem.Block{
//		Type:  "私钥",
//		Bytes: derStream,
//	}
//	file, err := os.Create(filepath.Join("conf", "private.pem"))
//	if err != nil {
//		return err
//	}
//	err = pem.Encode(file, block)
//	if err != nil {
//		file.Close()
//		return err
//	}
//	file.Close()
//	// 生成公钥文件
//	publicKey := &privateKey.PublicKey
//	derPkix, err := x509.MarshalPKIXPublicKey(publicKey)
//	if err != nil {
//		return err
//	}
//	block = &pem.Block{
//		Type:  "公钥",
//		Bytes: derPkix,
//	}
//	file, err = os.Create(filepath.Join("conf", "public.pem"))
//	if err != nil {
//		return err
//	}
//	err = pem.Encode(file, block)
//	if err != nil {
//		file.Close()
//		return err
//	}
//	file.Close()
//	return nil
//}

///*
//	加载公钥私钥文件
//*/
//func LoadRSAFile() error {
//	var err error
//	PublicKey, err = ioutil.ReadFile(filepath.Join("conf", "public.pem"))
//	if err != nil {
//		if err == os.ErrNotExist {
//			return nil
//		}
//		return err
//	}
//	PrivateKey, err = ioutil.ReadFile(filepath.Join("conf", "private.pem"))
//	if err != nil {
//		if err == os.ErrNotExist {
//			return nil
//		}
//		return err
//	}

//	return nil
//}

//// 加密
//func RsaEncrypt(publicKey, origData []byte) ([]byte, error) {
//	block, _ := pem.Decode(publicKey)
//	if block == nil {
//		return nil, errors.New("public key error")
//	}
//	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
//	if err != nil {
//		return nil, err
//	}
//	pub := pubInterface.(*rsa.PublicKey)
//	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
//}

//// 解密
//func RsaDecrypt(privateKey, ciphertext []byte) ([]byte, error) {
//	block, _ := pem.Decode(privateKey)
//	if block == nil {
//		return nil, errors.New("private key error!")
//	}
//	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
//	if err != nil {
//		return nil, err
//	}
//	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
//}
