package utils

import (
	"bytes"
	//"compress/gzip"
	"crypto/ecdsa"
	//"crypto/elliptic"
	"crypto/rand"
	//"crypto/x509"
	//"encoding/hex"
	"encoding/pem"
	//"errors"
	// "fmt"
	"io/ioutil"

	//"math/big"
	"os"
	//"strings"
)

/*
	生成私钥和公钥文件到本地
*/
func GenerateKey(prktype, puktype, prkpath, pukpath string) error {

	//生成私钥
	priv, err := ecdsa.GenerateKey(S256(), rand.Reader)
	if err != nil {
		return err
	}
	ecder, err := MarshalPrikey(priv)
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

	ecder, err = MarshalPubkey(publicKey)
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

/*
	创建一对密钥，私钥已经加密，公钥不需要加密
*/
func BuildKey(passwd, prktype, puktype, prkpath, pukpath string) (*ecdsa.PrivateKey, error) {
	//将密码hash运算，不暴露密码
	hash := Hash_SHA3_256([]byte(passwd))
	//	fmt.Println("hash1", hash)
	key := hash[len(hash)-16:]
	//	fmt.Println("hash2", key)

	//生成私钥
	priv, err := ecdsa.GenerateKey(S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	ecder, err := MarshalPrikey(priv)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)

	err = pem.Encode(buf, &pem.Block{Type: prktype, Bytes: ecder})
	if err != nil {
		return nil, err
	}
	// fmt.Println("加密前的", string(buf.Bytes()))
	bs := EncryptPrk(buf.Bytes(), key)
	// fmt.Println("加密后的", string(bs))
	keypem, err := os.OpenFile(prkpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	keypem.Write(bs)

	//
	publicKey := &priv.PublicKey

	ecder, err = MarshalPubkey(publicKey)
	if err != nil {
		// fmt.Println("111", err)
		return nil, err
	}
	keypem, err = os.OpenFile(pukpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	err = pem.Encode(keypem, &pem.Block{Type: puktype, Bytes: ecder})
	if err != nil {
		return nil, err
	}

	return priv, nil

}

/*
	生成私钥
*/
func GeneratePrkKey() (*ecdsa.PrivateKey, error) {
	//生成私钥
	//	priv, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	//	if err != nil {
	//		return err
	//	}
	return ecdsa.GenerateKey(S256(), rand.Reader)
}

/*
	加载本地私钥
*/
func LoadPrk(name string) (*ecdsa.PrivateKey, error) {
	bs, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	b, _ := pem.Decode(bs)
	//	fmt.Println(b, rest)

	prk, err := ParsePrikey(b.Bytes)
	if err != nil {
		return nil, err
	}
	return prk, nil
}

/*
	加载本地公钥
*/
func LoadPuk(name string) (*ecdsa.PublicKey, error) {
	bs, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	b, _ := pem.Decode(bs)
	//	fmt.Println(b, rest)

	puk, err := ParsePubkey(b.Bytes)
	if err != nil {
		return nil, err
	}
	return puk, nil
}

/*
	对内容生成签名
*/
func Sign(prk *ecdsa.PrivateKey, bs []byte) (*[]byte, error) {
	sign, err := SignCompact(prk, bs)
	return sign, err
}

/*func signOld(prk *ecdsa.PrivateKey, bs []byte) (string, error) {
	//	r, s, err := ecdsa.Sign(strings.NewReader(randSign), prk, []byte(text))
	r, s, err := ecdsa.Sign(rand.Reader, prk, bs)
	if err != nil {
		return "", err
	}
	rt, err := r.MarshalText()
	if err != nil {
		return "", err
	}
	st, err := s.MarshalText()
	if err != nil {
		return "", err
	}
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	defer w.Close()
	_, err = w.Write([]byte(string(rt) + "+" + string(st)))
	if err != nil {
		return "", err
	}
	w.Flush()
	return hex.EncodeToString(b.Bytes()), nil
}*/

//解密
/*func getSign(text, byterun []byte) (rint, sint big.Int, err error) {
	r, err := gzip.NewReader(bytes.NewBuffer(byterun))
	if err != nil {
		err = errors.New("decode error," + err.Error())
		return
	}
	defer r.Close()
	buf := make([]byte, 1024)
	count, err := r.Read(buf)
	if err != nil {
		fmt.Println("decode =", err)
		err = errors.New("decode read error," + err.Error())
		return
	}
	rs := strings.Split(string(buf[:count]), "+")
	if len(rs) != 2 {
		err = errors.New("decode fail")
		return
	}
	err = rint.UnmarshalText([]byte(rs[0]))
	if err != nil {
		err = errors.New("decrypt rint fail, " + err.Error())
		return
	}
	err = sint.UnmarshalText([]byte(rs[1]))
	if err != nil {
		err = errors.New("decrypt sint fail, " + err.Error())
		return
	}
	return
}*/

//Verify 对密文和明文进行匹配校验
func Verify(pukbyte, text []byte, sign []byte) (bool, error) {
	//	fmt.Println("对密文和明文进行匹配校验\n", len(pukbyte), pukbyte, "\n", string(text), "\n", len(sign), sign)
	/*pukbyte, err := MarshalPubkey(puk)
	if err != nil {
		return false, err
	}*/
	res := VerifyS256(text, pukbyte, sign)
	return res, nil
}

/*func verifyOld(puk *ecdsa.PublicKey, text []byte, passwd string) (bool, error) {
	byterun, err := hex.DecodeString(passwd)
	if err != nil {
		return false, err
	}
	rint, sint, err := getSign(text, byterun)
	if err != nil {
		return false, err
	}
	result := ecdsa.Verify(puk, text, &rint, &sint)
	return result, nil
}*/

///*
//	将公钥字符串解析为[]byte
//*/
//func ParsePukToBytes(puk string) []byte {
//	base64.StdEncoding.DecodeString(puk)
//}

/*
-----BEGIN EC PRIVATE KEY-----
MIHcAgEBBEIAJuVZ2ujd64R6iA3mJ2Iuaorwc3kUXihcLvF51nuF8MHE82ZobU4k
JBwOjjFX1jjZGDAyiIgXpD9HXtfIWCpIksigBwYFK4EEACOhgYkDgYYABAB7ZnsN
kzgg7JrVxohq2FOxRpA+bUsVsxHKHRur0ezttNRu7ldR36fGgyztoplyUltUndID
7EEi2o9z2E4q99ozOgHwbHnx5WROGYtwvdwDQtvHnCynzO7eixMFZ8pKLddz7YLs
hm8sApO1ZneSDMMBoKhtlX4+pywXZ91WxEfOY/GXlA==
-----END EC PRIVATE KEY-----

将这种格式的[]byte，去掉头和尾后拼接中间的字符
*/
func ParseKeyToByte(prkbs []byte) []byte {
	bss := bytes.Split(prkbs, []byte("\n"))
	//	bss = bss[:len(bss)-1]
	//	begin := string(bss[0])
	//	end := string(bss[len(bss)-1])
	return bytes.Join(bss[1:len(bss)-2], nil)
}

func BuildKeyToByte(t, key string) (prkbs string) {
	begin := "-----BEGIN " + t + "-----"
	end := "-----END " + t + "-----"
	bs := []byte(key)
	out := append([]byte(begin), []byte("\n")...)
	//	fmt.Println("---3", string(bss[len(bss)-1]))
	for len(bs) > 64 {
		//		fmt.Println("---4", string(end))
		out = append(out, bs[:64]...)
		out = append(out, []byte("\n")...)
		bs = bs[64:]
	}
	out = append(out, bs...)
	out = append(out, []byte("\n")...)
	out = append(out, end...)
	out = append(out, []byte("\n")...)
	return string(out)
}

//解析pem公钥
func DecodePubkey(pubkey []byte) *ecdsa.PublicKey {
	b, _ := pem.Decode(pubkey)
	prk, err := ParsePubkey(b.Bytes)
	if err != nil {
		// fmt.Println(err)
	}
	return prk

}
