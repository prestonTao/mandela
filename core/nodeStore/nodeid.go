package nodeStore

import (
	"mandela/core/utils"
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"github.com/satori/go.uuid"
)

var (
	Zero = 8
)

func GetBinary(key []byte) string {
	bstring := fmt.Sprintf("%b", key)
	b := bstring[1 : len(bstring)-1]
	bs := strings.Split(b, " ")
	bstr := ""
	for _, v := range bs {
		if len(v) < 8 {
			bstr += strings.Repeat("0", 8-len(v)) + v
		} else {
			bstr += v
		}
	}
	return bstr
}
func IntToBytes(n int64) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.LittleEndian, n)
	return bytesBuffer.Bytes()
}

// //生成私/公钥对
// func GetEccKey(rand []byte) (*ECCKey, error) {
// 	//curve := elliptic.P256()
// 	curve := utils.S256()
// 	//private, err := ecdsa.GenerateKey(curve, rand.Reader)
// 	private, err := ecdsa.GenerateKey(curve, bytes.NewReader(rand))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return NewECCKey(private, nil), nil
// }

func CreateUUID() string {
	uuid := uuid.NewV4()
	return uuid.String()
}
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	return publicSHA256[:]
}

func CheckZn(text []byte) bool {
	bs := GetBinary(text)
	pre := strings.Repeat("0", Zero)
	if strings.HasPrefix(bs, pre) {
		return true
	}
	return false
}
func CheckSafeAddr(pubkey []byte) bool {
	ts := HashPubKey(pubkey)
	return CheckZn(ts)
}

// func CreateAddr(uuid string, index int64) (*ECCKey, []byte) {
// 	//1
// 	hash := sha256.Sum256([]byte(uuid))
// 	//2
// 	rand := IntToBytes(index)
// 	rand = append(rand, hash[:]...)
// 	//3
// 	ecc, err := GetEccKey(rand)
// 	if err != nil {
// 		return nil, nil
// 	}
// 	//log.Println(prikey, pubkey)
// 	pb, err := ecc.GetPukBytes()
// 	if err != nil {
// 		return nil, nil
// 	}
// 	str := HashPubKey(pb)
// 	return ecc, str
// }
// func GetAddrInfo() *ECCKey {
// 	uuid := CreateUUID()
// 	uuid = uuid + time.Now().Format("2006-01-02 15:04:05.999999999")
// 	//不检查前导0
// 	ecc, _ := CreateAddr(uuid, 0)
// 	return ecc
// 	//	for i := 0; ; i++ {
// 	//		ecc, chks := CreateAddr(uuid, int64(i))
// 	//		if CheckZn(chks) {
// 	//			//fmt.Printf("ok: %s %v\n", GetBinary(chks), pubs)
// 	//			return ecc
// 	//		}
// 	//	}
// }
func SaveKey(priv *ecdsa.PrivateKey, prktype, puktype, prkpath, pukpath string) error {

	//生成私钥
	/*priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	if err != nil {
		return err
	}*/
	ecder, err := utils.MarshalPrikey(priv)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)

	err = pem.Encode(buf, &pem.Block{Type: prktype, Bytes: ecder})
	if err != nil {
		return err
	}
	//	fmt.Println("加密前的", string(buf.Bytes()))
	//	bs := EncryptPrk(buf.Bytes(), key)
	//	fmt.Println("加密后的", string(bs))
	keypem, err := os.OpenFile(prkpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		keypem.Close()
		return err
	}
	keypem.Write(buf.Bytes())
	keypem.Close()

	//
	publicKey := &priv.PublicKey

	ecder, err = utils.MarshalPubkey(publicKey)
	if err != nil {
		//		fmt.Println("111", err)
		return err
	}
	keypem, err = os.OpenFile(pukpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		keypem.Close()
		return err
	}
	err = pem.Encode(keypem, &pem.Block{Type: puktype, Bytes: ecder})
	if err != nil {
		return err
	}
	keypem.Close()

	return nil
}

// func GetKeyPair() (*ECCKey, error) {
// 	// //判断本地是否有私钥
// 	// ok, err := utils.PathExists(filepath.Join(config.Path_configDir, config.Core_addr_prk))
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// if ok {
// 	// 	//		return LoadECCPrk(filepath.Join(config.Path_configDir, config.Core_addr_prk))
// 	// 	prk, err := utils.LoadPrk(filepath.Join(config.Path_configDir, config.Core_addr_prk))
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}
// 	// 	//pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
// 	// 	return NewECCKey(prk, nil), nil
// 	// }
// 	// //没有私钥则生成一个私钥
// 	// ecc := GetAddrInfo()
// 	// err = SaveKey(ecc.GetPrk(), config.Core_addr_prk_type, config.Core_addr_puk_type,
// 	// 	filepath.Join(config.Path_configDir, config.Core_addr_prk),
// 	// 	filepath.Join(config.Path_configDir, config.Core_addr_puk))
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// //加载生成的私钥
// 	// prk, err := utils.LoadPrk(filepath.Join(config.Path_configDir, config.Core_addr_prk))
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// return NewECCKey(prk, nil), nil
// 	return nil, nil
// }
