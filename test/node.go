package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	//"crypto/elliptic"
	//"crypto/rand"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	//"encoding/binary"
	//"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"mandela/config"
	"mandela/core/nodeStore"
	//"strings"
	"encoding/hex"
	"time"
	"mandela/core/utils"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	//"github.com/syndtr/goleveldb/leveldb"
)

type A struct {
	Id   string `"json:"id"`
	Name []byte `json:"name"`
}

type X struct {
	E []byte
}

func (x *X) Ts(b *[]byte, c string) {
	x.Tss(b)
}
func (x *X) Tss(e *[]byte) {
	x.E = []byte("aaa")
}
func EncodePubkey(pubkey *ecdsa.PublicKey) ([]byte, error) {
	pubKey, err := x509.MarshalPKIXPublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	fmt.Println(len(pubKey))
	buf := bytes.NewBuffer(nil)
	err = pem.Encode(buf, &pem.Block{Type: config.Core_addr_puk_type, Bytes: pubKey})
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func DecodePubkey(pubkey []byte) (*ecdsa.PublicKey, error) {
	b, _ := pem.Decode(pubkey)
	fmt.Println(b)
	prk, err := x509.ParsePKIXPublicKey(b.Bytes)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return prk.(*ecdsa.PublicKey), nil

}
func Verify() error {
	curve := btcec.S256() //elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	/*str := strings.Repeat("b", 36)
	rand := []byte(str)
	fmt.Println(str, len(rand))
	private, err := ecdsa.GenerateKey(curve, bytes.NewReader(rand))
	/*
	if err != nil {
		return err
	}
	//pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	/*bss := bytes.Split(buf.Bytes(), []byte("\n"))
	pubkey := bytes.Join(bss[1:len(bss)-2], nil)
	*/
	pub, _ := EncodePubkey(&private.PublicKey)
	fmt.Printf("2%s\n", pub)

	x, err := DecodePubkey(pub)
	fmt.Printf("%+v %v\n", x, err)
	sign, err := utils.Sign(private, []byte("aaa"))
	if err != nil {
		fmt.Println(x, err)
	}
	fmt.Printf("sign:%v\n", sign)
	ok, _ := utils.Verify(pub, []byte("aaa"), hex.EncodeToString(*sign))
	if ok {
		fmt.Println("verify success")
	} else {
		fmt.Println("verify fail")
	}
	return nil
}

//github.com/btcsuite/btcd/btcec/exmple_test.go
func VerifyS256(text string, pub []byte, sign string) bool {
	// Decode hex-encoded serialized public key.
	//pubKeyBytes, err := base64.StdEncoding.DecodeString(pub)
	pubKeyBytes := pub
	fmt.Printf("%x\n", pubKeyBytes)
	/*pubKeyBytes, err := hex.DecodeString(pub)
	if err != nil {
		fmt.Println(err)
		return false
	}*/
	pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		fmt.Println("pub", err)
		return false
	}

	// Decode hex-encoded serialized signature.
	sigBytes, err := hex.DecodeString(sign)

	if err != nil {
		fmt.Println("sign", err)
		return false
	}
	//sigBytes, err := base64.StdEncoding.DecodeString(sign)
	/*if err != nil {
		fmt.Println(err)
		return false
	}*/
	signature, err := btcec.ParseSignature(sigBytes, btcec.S256())
	if err != nil {
		fmt.Println("parse", err)
		return false
	}

	// Verify the signature for the message using the public key.
	message := text
	messageHash := chainhash.DoubleHashB([]byte(message))
	verified := signature.Verify(messageHash, pubKey)
	fmt.Println("Signature Verified?", verified)

	return verified
}
func main() {
	/*db, _ := leveldb.OpenFile("../example/peer_root/wallet/data/", nil)
	fmt.Printf("%+v\n", db)
	ls, _ := db.GetProperty("leveldb.stats")
	fmt.Println(ls)
	os.Exit(0)
	sss, _ := base64.StdEncoding.DecodeString("FiD2VFX7X+a8jlSEoNmqGslf26Kuy8wlnLooq1Er5YgzaQ==")
	mulitihash := utils.Multihash(sss)
	fmt.Println(mulitihash.B58String())*/
	/*buf := utils.Hash_SHA3_256(utils.Hash_SHA3_256([]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")))
	fmt.Printf("1 %+v %d\n", buf, len(buf))
	start := make([]byte, 2*binary.MaxVarintLen64)
	spot := start
	fmt.Printf("2 %v %v\n", start, spot)
	n := binary.PutUvarint(spot, uint64(0x16))
	fmt.Printf("3 %+v %+v\n", spot, start)
	spot = start[n:]
	fmt.Printf("3 %v %v\n", n, spot)
	n += binary.PutUvarint(spot, uint64(len(buf)))
	fmt.Printf("4 %v %v %+v\n", n, spot, start)
	res := append(start[:n], buf...)
	fmt.Printf("5%+v %+v %d\n", start[:n], res, len(res))
	//os.Exit(0)
	pubtexts, puberrs := base64.StdEncoding.DecodeString("FiAI8d9O/ksvIBkljS2+Xv96i+3wmfWrZXzcJT7cs8U8xA==")
	hash, hasherr := utils.Decode(pubtexts)
	fmt.Printf("a %+v %v\n", hash, hasherr)
	fmt.Printf("b %+v %v %d %x\n", pubtexts, puberrs, len(pubtexts), pubtexts)
	b1, err1 := utils.FromB58String("W1a3uQWUt56vbEm4s9UNZ6CnJiJLvj434Go2KLLWp8akzP")
	fmt.Printf("c %+v err: %v %d \n", b1, err1, len(b1))
	fmt.Println(b1.B58String())
	os.Exit(0)
	fmt.Println(hex.DecodeString("01D1"))*/
	js := `{
			"idinfo": {
				"id": "CPHfTv5LLyAZJY0tvl7/eovt8Jn1q2V83CU+3LPFPMQ=",
				"sign": "3045022100b869bb326cbc418cd6714f1247852a40ca5940a8304ffb0d9a8e9264d4defc4c02205e0082a9f2278e829f3e1231c6cf2df3dd72b8b8416e5abacdef0f197df061f0",
				"ctype": "s256",
				"puk": "BN8YWizuPl9mgeLX+D3kpMaV8YhmilRxjI4qZES2jZEwzoNRWuCr331gNhe63n8xXx+GoA9q0CeXqv3hH7ybfN4="
			},
			"issuper": false,
			"tcpport": 49537
		}`
	js = `{"idinfo":{"id":"QjRRZVN1S0NvekJvZ2tDRkxqb1NrTjI3RTdZSjFGYlJqTkc2a005dGNDOU4=","puk":"AhKZo9QM3ewdv8f87dKHBct+0gphM3Ys5yleZywLC5Df","sign":"3044022050be7307f5a266893b52f5d75dfe7387280ca844fec21067535f8d0a72092dd4022032fe51d57156d2f899ca259a55ebc746e396b1f1d5e8d4b30a47c8fe07e8ab5c"},"issuper":false,"tcpport":19981,"addr":"192.168.2.194"}`
	ss, ess := nodeStore.ParseNode([]byte(js))
	fmt.Printf("%+v %v\n", ss, ess)
	fmt.Println(ss.IdInfo.Id.B58String())
	//os.Exit(0)
	fmt.Println(nodeStore.CheckSafeAddr(ss.IdInfo.Puk))
	fmt.Println(ss.IdInfo.Id.B58String(), "bvB8b8GMc2i5ybEpV44SbEeuA8BZHCmUZXVH68sZgsu")
	fmt.Println(VerifyS256(ss.IdInfo.Id.B58String(), ss.IdInfo.Puk, ss.IdInfo.Sign))
	fmt.Println(utils.VerifyS256(*ss.IdInfo.Id, ss.IdInfo.Puk, ss.IdInfo.Sign))
	os.Exit(0)
	pubtext, puberr := base64.StdEncoding.DecodeString("BN8YWizuPl9mgeLX+D3kpMaV8YhmilRxjI4qZES2jZEwzoNRWuCr331gNhe63n8xXx+GoA9q0CeXqv3hH7ybfN4=")
	fmt.Println(nodeStore.CheckSafeAddr(pubtext), puberr)
	os.Exit(0)
	text, err := base64.StdEncoding.DecodeString("YnZCOGI4R01jMmk1eWJFcFY0NFNiRWV1QThCWkhDbVVaWFZINjhzWmdzdQ==")
	fmt.Println(string(text))
	//b := VerifyS256(string(text), "BN8YWizuPl9mgeLX+D3kpMaV8YhmilRxjI4qZES2jZEwzoNRWuCr331gNhe63n8xXx+GoA9q0CeXqv3hH7ybfN4=", "3045022100b869bb326cbc418cd6714f1247852a40ca5940a8304ffb0d9a8e9264d4defc4c02205e0082a9f2278e829f3e1231c6cf2df3dd72b8b8416e5abacdef0f197df061f0")
	//fmt.Println(b)
	//os.Exit(0)
	ecc1, err1 := nodeStore.GetKeyPair()
	pub1, _ := ecc1.GetPukBytes()
	sha := sha256.Sum256(pub1)
	fmt.Println(sha, err1)
	os.Exit(0)

	//os.Exit(0)
	st := time.Now()
	ecc, err := nodeStore.GetKeyPair()
	fmt.Println(ecc, err)
	et := time.Now().Sub(st)
	fmt.Println(et)
	//os.Exit(0)
	//Verify()
	//os.Exit(0)
	/*str := `
	-----BEGIN EC PUBLIC KEY-----
		MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE2DIr6uFSZdmokn7CYCFUWsK5eZCB
		aNxy00iD4BdxNo0+1Cr/fXh1xT4B8WsYwlKITlCEaoCyVUrec4lvwrdkaQ==
	-----END EC PUBLIC KEY-----`*/

	//str := "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE1cxEBi7YBwjNW611y1OYJFG5WPhpYOv7CJClsvKZWU3VzEQGLtgHCM1brXXLU5gkUblY+Glg6/sIkKWy8plZTQ=="
	//str := "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDAqjJs08oHvNdhlWC+kGBd90PD7CVjClhRTk3nn+2NNaP4Bi5N/A18rdrV6clNAGUz4i/5q/VQXeLiGYYqgmAkKCJegReMsfcnoOSWu+Tvxih/48pu1hwBrmMLFZPOOUWQ9YjQEo7SYBe0HKoEl6XMqNwzHV7sk9x6BKz9QeLi5QIDAQAB"
	//p1, e1 := base64.StdEncoding.DecodeString(str)
	//fmt.Println(e1, p1, len(p1))
	//ss, errs := DecodePubkey(pub)
	//fmt.Printf("2%+v %v", ss, errs)
	//os.Exit(0)
	//xx := utils.CheckIdInfo()
	//fmt.Println(xx)
	//os.Exit(0)
	/*str := `{
	  "idinfo" : {
	    "id" : "42703155594a425934314e614b786272416e6f5341613553584d457a50795a6935326f65525275323348544b",
	    "puk" : "YWFh",
	    "sign" : "bf8099cbd7af26a43eb5ccdee99423a6c5f8036d5a1959c4c08cc387ac2d196962f1f89e8874a8f4090f08394ed035fa57b322580597b11f9df254c4c7c366a8"
	  },
	  "issuper" : false,
	  "tcpport" : 52989
	}`
		node := new(nodeStore.Node)
		err := json.Unmarshal([]byte(str), node)
		fmt.Printf("dddd%+v %v\n", node, err)
		as := A{Id: "aaa", Name: []byte("aaa")}
		s, _ := json.Marshal(as)
		fmt.Printf("%+v %s\n", s, s)
		a, b := binary.Uvarint([]byte{'a'})
		fmt.Printf("%+v %+v\n", a, b)
		c, d := binary.Varint([]byte("0"))
		fmt.Printf("%+v %+v", c, d)
	*/
	Verify()
	os.Exit(0)
}
