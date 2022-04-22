package chain_witness_vote

import (
	"mandela/config"
	"mandela/core/utils"
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var keyLock = new(sync.RWMutex)
var key *Key

/*
	检查是否有钱包key
*/
func CheckKey() bool {
	ok := false
	keyLock.RLock()
	if key != nil {
		ok = true
	}
	keyLock.RUnlock()
	return ok
}

/*
	先从本地加载密钥，没有再生成新的
*/
func LoadKeyStore() (*KeyStore, error) {
	key, err := LoadKeyStoreToLocal()
	if err != nil {
		//TODO 判断是什么错误，如果是读文件错误另外处理
		return nil, err
	}
	return key, err
}

type Keys struct {
	Seeds []SeedKey
}

/*
	获得钱包收款地址
*/
func (this *Keys) GetAddrs() []string {
	addr := make([]string, 0)
	for _, one := range this.Seeds {
		addr = append(addr, one.GetAddrs()...)
	}
	return addr
}

type SeedKey struct {
	Seed []byte //种子文件
	Keys []Key  //
}

/*
	获得钱包收款地址
*/
func (this *SeedKey) GetAddrs() []string {
	addr := make([]string, 0)
	for _, one := range this.Keys {
		addr = append(addr, one.Addr)
	}
	return addr
}

type Key struct {
	Index int64  //种子下标
	Puk   string //公钥
	Addr  string //收款地址
}

// /*
// 	生成一个种子文件
// */
// func BuildSeedKey() error {
// 	prk, err := utils.GeneratePrkKey()
// 	if err != nil {
// 		return err
// 	}
// 	puk := prk.PublicKey
// 	ecckey := nodeStore.NewECCKey(prk, &puk)
// 	bs, err := ecckey.GetPukBytes()
// 	bs = utils.Hash_SHA3_512(bs)
// 	// fmt.Println(len(bs))
// 	return nil
// }

/*
	保存多个种子密钥文件和多个单独密钥
*/
type KeyStore struct {
	Seed  []SeedKeyStore
	Alone []AloneKeyStore
}

/*
	保存密钥种子文件
*/
type SeedKeyStore struct {
	Key    []byte  //种子文件
	Indexs []int64 //已使用的密钥文件索引号
}

/*
	保存单个密钥文件
*/
type AloneKeyStore struct {
	Prk string
}

/*
	保存种子文件到本地
*/
func (this *KeyStore) Save() error {
	bs, err := json.Marshal(this)
	if err != nil {
		return err
	}
	return utils.SaveFile(filepath.Join(config.Wallet_path, config.Wallet_seed), &bs)
}

/*
	从本地文件加载钱包私钥
*/
func LoadKeyStoreToLocal() (*KeyStore, error) {
	//判断本地是否有私钥
	ok, err := utils.PathExists(filepath.Join(config.Wallet_path, config.Wallet_seed))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("There is no key file locally")
	}

	bs, err := ioutil.ReadFile(filepath.Join(config.Wallet_path, config.Wallet_seed))
	if err != nil {
		return nil, err
	}
	keyStore := new(KeyStore)

	// err = json.Unmarshal(bs, keyStore)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(keyStore)
	return keyStore, err

}

//import (
//	"bytes"
//	"crypto/aes"
//	"crypto/cipher"
//	"crypto/ecdsa"
//	"crypto/elliptic"
//	"crypto/rand"
//	"crypto/x509"
//	"encoding/base64"
//	"encoding/hex"
//	"encoding/pem"
//	"fmt"
//	"strings"
//	"time"
//	//	"io/ioutil"
//	"os"
//	"path/filepath"
//	gconfig "mandela/config"
//	"mandela/core/utils"
//	//	"github.com/ethereum/go-ethereum/crypto"
//)

//////Encrypt 对Text进行加密，返回加密后的字节流
////func Sign(text string) (string, error) {
////	r, s, err := ecdsa.Sign(strings.NewReader(randSign), prk, []byte(text))
////	if err != nil {
////		return "", err
////	}
////	rt, err := r.MarshalText()
////	if err != nil {
////		return "", err
////	}
////	st, err := s.MarshalText()
////	if err != nil {
////		return "", err
////	}
////	var b bytes.Buffer
////	w := gzip.NewWriter(&b)
////	defer w.Close()
////	_, err = w.Write([]byte(string(rt) + "+" + string(st)))
////	if err != nil {
////		return "", err
////	}
////	w.Flush()
////	return hex.EncodeToString(b.Bytes()), nil
////}

//func BuildPrkSead() error {
//	//产生私钥
//	bs := BuildPrk()

//	fmt.Println("解析前", len(bs), string(bs))

//	key := ParseKeyToByte(bs)

//	fmt.Println(len(key), string(key))

//	bs, _ = base64.StdEncoding.DecodeString(string(key))

//	fmt.Println(len(bs), string(bs))
//	return nil
//}

///*
//	创建一对密钥，私钥已经加密，公钥不需要加密
//*/
//func BuildKey(passwd string) (*ecdsa.PrivateKey, error) {
//	//将密码hash运算，不暴露密码
//	hash := utils.Hash_SHA3_256([]byte(passwd))
//	//	fmt.Println("hash1", hash)
//	key := hash[len(hash)-16:]
//	//	fmt.Println("hash2", key)

//	//生成私钥
//	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
//	if err != nil {
//		return nil, err
//	}
//	ecder, err := x509.MarshalECPrivateKey(priv)
//	if err != nil {
//		return nil, err
//	}

//	buf := bytes.NewBuffer(nil)

//	err = pem.Encode(buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecder})
//	if err != nil {
//		return nil, err
//	}
//	fmt.Println("加密前的", string(buf.Bytes()))
//	bs := EncryptPrk(buf.Bytes(), key)
//	fmt.Println("加密后的", string(bs))
//	keypem, err := os.OpenFile(filepath.Join(gconfig.Wallet_path, gconfig.Wallet_path_prkName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
//	if err != nil {
//		return nil, err
//	}
//	keypem.Write(bs)

//	//
//	publicKey := &priv.PublicKey

//	ecder, err = x509.MarshalPKIXPublicKey(publicKey)
//	if err != nil {
//		fmt.Println("111", err)
//		return nil, err
//	}
//	keypem, err = os.OpenFile(filepath.Join(gconfig.Wallet_path, gconfig.Wallet_path_pukName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
//	if err != nil {
//		return nil, err
//	}
//	err = pem.Encode(keypem, &pem.Block{Type: "EC PUBLIC KEY", Bytes: ecder})
//	if err != nil {
//		return nil, err
//	}

//	return priv, nil

//}

///*
//	本地磁盘加载公钥
//*/
//func LoadPuk() error {
//	//	bs, err := ioutil.ReadFile(filepath.Join(gconfig.Wallet_path, gconfig.Wallet_path_pukName))

//	//		ecdsa.
//	return nil
//}

///*
//	对私钥加密
//*/
//func EncryptPrk(prkbs, key []byte) []byte {
//	bss := bytes.Split(prkbs, []byte("\n"))
//	bss = bss[:len(bss)-1]
//	begin := string(bss[0])
//	end := string(bss[len(bss)-1])
//	//	fmt.Println("---1", string(bss[len(bss)-1]))
//	//	for _, one := range bss {
//	//		fmt.Println(len(one), string(one))
//	//	}
//	bs := bytes.Join(bss[1:len(bss)-1], nil)
//	//	fmt.Println("---2", string(bss[len(bss)-1]))
//	//	fmt.Println("加密前长度", len(bs), string(bs))

//	bs, _ = Encrypt(bs, key)
//	//	fmt.Println("加密后长度", len(bs), base64.StdEncoding.EncodeToString(bs))
//	//	fmt.Println(base64.StdEncoding.EncodeToString(bs))
//	newbs := []byte(base64.StdEncoding.EncodeToString(bs))
//	out := append([]byte(begin), []byte("\n")...)
//	//	fmt.Println("---3", string(bss[len(bss)-1]))
//	for len(newbs) > 64 {
//		//		fmt.Println("---4", string(end))
//		out = append(out, newbs[:64]...)
//		out = append(out, []byte("\n")...)
//		newbs = newbs[64:]
//	}
//	out = append(out, newbs...)
//	out = append(out, []byte("\n")...)
//	out = append(out, end...)
//	//	fmt.Println("---", string(bss[len(bss)-1]))

//	//	bs, _ = Decrypt(bs, key)
//	//	fmt.Println("解密后长度", len(bs))
//	//	fmt.Println(string(bs))

//	return out
//}

///*
//	对私钥解密
//*/
//func DecryptPrk(prkbs, key []byte) []byte {
//	bss := bytes.Split(prkbs, []byte("\n"))
//	bs := bytes.Join(bss[1:len(bss)-2], nil)
//	fmt.Println(string(bs))

//	//	fmt.Println("加密前长度", len(bs))
//	//	bs, _ = Encrypt(bs, key)
//	//	fmt.Println("加密后长度", len(bs))
//	//	fmt.Println(base64.StdEncoding.EncodeToString(bs))

//	bs, _ = Decrypt(bs, key)
//	fmt.Println("解密后长度", len(bs))
//	fmt.Println(string(bs))

//	return nil
//}

////func encode(passwd string) []byte {
////	//	srcByte := []byte("nihao")
////	iv := []byte(passwd)
////	key := []byte("#QvLABlM/?1z3!b#")
////	c, err := aes.NewCipher(key)
////	if err != nil {
////		panic(err)
////	}
////	encrypter := cipher.NewCBCEncrypter(c, iv)
////	data := make([]byte, len(commonInput))
////	copy(data, commonInput)
////	encrypter.CryptBlocks(data, data)
////	fmt.Println(hex.EncodeToString(data))
////	return data
////}

//func Encrypt(plantText, key []byte) ([]byte, error) {
//	key = Key16Padding(key)
//	fmt.Println("key", key)
//	block, err := aes.NewCipher(key) //选择加密算法
//	if err != nil {
//		return nil, err
//	}
//	plantText = PKCS7Padding(plantText, block.BlockSize())

//	blockModel := cipher.NewCBCEncrypter(block, key)

//	ciphertext := make([]byte, len(plantText))

//	blockModel.CryptBlocks(ciphertext, plantText)
//	return ciphertext, nil
//}

//func Decrypt(ciphertext, key []byte) ([]byte, error) {
//	key = Key16Padding(key)
//	fmt.Println("key", key)

//	//	keyBytes := []byte(key)
//	block, err := aes.NewCipher(key) //选择加密算法
//	if err != nil {
//		return nil, err
//	}
//	blockModel := cipher.NewCBCDecrypter(block, key)
//	plantText := make([]byte, len(ciphertext))
//	blockModel.CryptBlocks(plantText, ciphertext)
//	plantText = PKCS7UnPadding(plantText, block.BlockSize())
//	return plantText, nil
//}

//func PKCS7UnPadding(plantText []byte, blockSize int) []byte {
//	length := len(plantText)
//	unpadding := int(plantText[length-1])
//	//	checkbs := plantText[length-unpadding:]
//	//	for _, one := range checkbs {
//	//		if int(one) != unpadding {
//	//			return plantText
//	//		}
//	//	}
//	return plantText[:(length - unpadding)]
//}

///*
//	将加密秘钥填补成16字节长度
//*/
//func Key16Padding(key []byte) []byte {
//	//	bs := []byte(key)
//	if len(key) >= 16 {
//		return key
//	}
//	newkey := make([]byte, 16)
//	copy(newkey, key)
//	return newkey
//}

//func BuildPrk() []byte {
//	//生成私钥
//	//	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
//	priv, err := ecdsa.GenerateKey(elliptic.P521(), strings.NewReader(hex.EncodeToString(utils.Hash_SHA3_256(time.Now().String()))))
//	if err != nil {
//		fmt.Println("-1---", err)
//		return nil
//	}
//	ecder, err := x509.MarshalECPrivateKey(priv)
//	if err != nil {
//		return nil
//	}

//	buf := bytes.NewBuffer(nil)

//	err = pem.Encode(buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecder})
//	if err != nil {
//		return nil
//	}
//	return buf.Bytes()
//}

///*
//-----BEGIN EC PRIVATE KEY-----
//MIHcAgEBBEIAJuVZ2ujd64R6iA3mJ2Iuaorwc3kUXihcLvF51nuF8MHE82ZobU4k
//JBwOjjFX1jjZGDAyiIgXpD9HXtfIWCpIksigBwYFK4EEACOhgYkDgYYABAB7ZnsN
//kzgg7JrVxohq2FOxRpA+bUsVsxHKHRur0ezttNRu7ldR36fGgyztoplyUltUndID
//7EEi2o9z2E4q99ozOgHwbHnx5WROGYtwvdwDQtvHnCynzO7eixMFZ8pKLddz7YLs
//hm8sApO1ZneSDMMBoKhtlX4+pywXZ91WxEfOY/GXlA==
//-----END EC PRIVATE KEY-----

//将这种格式的[]byte，去掉头和尾后拼接中间的字符
//*/
//func ParseKeyToByte(prkbs []byte) []byte {
//	bss := bytes.Split(prkbs, []byte("\n"))
//	//	bss = bss[:len(bss)-1]
//	//	begin := string(bss[0])
//	//	end := string(bss[len(bss)-1])
//	return bytes.Join(bss[1:len(bss)-2], nil)
//}

//func BuildKeyToByte(bs []byte) (prkbs []byte) {
//	return nil
//}
