/*
	负责生成私钥，再通过私钥生成网络地址
	私钥 -> 公钥 -> sha3 -> sha3 -> 地址（sha3）
*/
package nodeStore

// import (
// 	"bytes"
// 	"crypto/ecdsa"
// 	"encoding/base64"
// 	"encoding/pem"
// 	"io/ioutil"
// 	"path/filepath"
// 	"mandela/config"
// 	"mandela/core/utils"
// )

// /*
// 	获取网络地址
// 	加载本地私钥，然后生成网络地址。没有私钥生成一个
// */
// func GetKey() (*ECCKey, error) {
// 	//判断本地是否有私钥
// 	ok, err := utils.PathExists(filepath.Join(config.Path_configDir, config.Core_addr_prk))
// 	if err != nil {
// 		return nil, err
// 	}
// 	if ok {
// 		//		return LoadECCPrk(filepath.Join(config.Path_configDir, config.Core_addr_prk))

// 		prk, err := utils.LoadPrk(filepath.Join(config.Path_configDir, config.Core_addr_prk))
// 		if err != nil {
// 			return nil, err
// 		}

// 		return NewECCKey(prk, nil), nil
// 	}
// 	//没有私钥则生成一个私钥
// 	err = utils.GenerateKey(config.Core_addr_prk_type, config.Core_addr_puk_type,
// 		filepath.Join(config.Path_configDir, config.Core_addr_prk),
// 		filepath.Join(config.Path_configDir, config.Core_addr_puk))
// 	if err != nil {
// 		return nil, err
// 	}
// 	//加载生成的私钥
// 	prk, err := utils.LoadPrk(filepath.Join(config.Path_configDir, config.Core_addr_prk))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewECCKey(prk, nil), nil
// }

// /*
// 	加载私钥
// */
// //func LoadPrk() error {
// //	bs, err := ioutil.ReadFile(filepath.Join(gconfig.Wallet_path, gconfig.Wallet_path_pukName))
// //	if err != nil {
// //		return err
// //	}
// //	b, rest := pem.Decode(bs)
// //	fmt.Println(b, rest)
// //	return nil
// //}

// type ECCKey struct {
// 	prk *ecdsa.PrivateKey
// 	puk *ecdsa.PublicKey
// }

// func (this *ECCKey) GetPrk() *ecdsa.PrivateKey {
// 	return this.prk
// }
// func (this *ECCKey) SetPuk(puk *ecdsa.PublicKey) {
// 	this.puk = puk
// }

// func (this *ECCKey) GetPukBytes() ([]byte, error) {
// 	return base64.StdEncoding.DecodeString(this.GetPukStrings())
// }
// func (this *ECCKey) GetPukStrings() string {
// 	if this.prk == nil && this.puk == nil {
// 		return ""
// 	}
// 	if this.puk == nil {
// 		this.puk = &this.prk.PublicKey
// 	}

// 	ecder, err := utils.MarshalPubkey(this.puk)
// 	if err != nil {
// 		return ""
// 	}
// 	buf := bytes.NewBuffer(nil)

// 	err = pem.Encode(buf, &pem.Block{Type: config.Core_addr_puk_type, Bytes: ecder})
// 	if err != nil {
// 		return ""
// 	}

// 	bss := bytes.Split(buf.Bytes(), []byte("\n"))
// 	bs := bytes.Join(bss[1:len(bss)-2], nil)
// 	return string(bs)

// }

// //func (this *ECCKey) GetPrkString() string {
// //	if this.prk == nil && this.puk == nil {
// //		return ""
// //	}
// //	if this.puk == nil {
// //		this.puk = this.prk.PublicKey
// //	}

// //	ecder, err := x509.MarshalECPrivateKey(priv)
// //	if err != nil {
// //		return nil, err
// //	}

// //	buf := bytes.NewBuffer(nil)

// //	err = pem.Encode(buf, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecder})
// //	if err != nil {
// //		return nil, err
// //	}

// //	bss := bytes.Split(prkbs, []byte("\n"))
// //	bss = bss[:len(bss)-1]
// //	begin := string(bss[0])
// //	end := string(bss[len(bss)-1])
// //	bs := bytes.Join(bss[1:len(bss)-1], nil)
// //	return string(bs)

// //}

// /*
// 	创建一个key
// */
// func NewECCKey(prk *ecdsa.PrivateKey, puk *ecdsa.PublicKey) *ECCKey {
// 	key := &ECCKey{
// 		prk: prk,
// 		puk: puk,
// 	}
// 	if prk != nil && puk == nil {
// 		key.puk = &prk.PublicKey
// 	}
// 	return key
// }

// /*
// 	从本地加载私钥
// */
// func LoadECCPrk(namepath string) (*ECCKey, error) {
// 	bs, err := ioutil.ReadFile(namepath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	b, _ := pem.Decode(bs)
// 	//	fmt.Println(b, rest)

// 	prk, err := utils.ParsePrikey(b.Bytes)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &ECCKey{
// 		prk: prk,
// 		puk: &prk.PublicKey,
// 	}, nil
// }
