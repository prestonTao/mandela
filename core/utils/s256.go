package utils

import (
	"crypto/ecdsa"
	//	"encoding/hex"
	"errors"
	// "fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func S256() *btcec.KoblitzCurve {
	return btcec.S256()
}
func MarshalPubkey(puk *ecdsa.PublicKey) ([]byte, error) {
	btcpuk := (*btcec.PublicKey)(puk)
	ecder := btcpuk.SerializeCompressed()
	if len(ecder) == 0 {
		return nil, errors.New("pubkey parse error")
	}
	return ecder, nil
}
func ParsePubkey(pubk []byte) (*ecdsa.PublicKey, error) {
	btcpub, err := btcec.ParsePubKey(pubk, btcec.S256())
	if err != nil {
		return nil, err
	}
	return (*ecdsa.PublicKey)(btcpub), nil
}
func MarshalPrikey(pri *ecdsa.PrivateKey) ([]byte, error) {
	btcpri := (*btcec.PrivateKey)(pri)
	bs := btcpri.Serialize()
	if len(bs) == 0 {
		return nil, errors.New("prikey parse error")
	}
	return bs, nil
}
func ParsePrikey(pk []byte) (*ecdsa.PrivateKey, error) {
	pri, _ := btcec.PrivKeyFromBytes(btcec.S256(), pk)
	if pri == nil {
		return nil, errors.New("prikey marshal error")
	}
	return (*ecdsa.PrivateKey)(pri), nil
}
func DoubleHashB(text []byte) []byte {
	return chainhash.DoubleHashB(text)
}

//s256签名
func SignCompact(pri *ecdsa.PrivateKey, hash []byte) (*[]byte, error) {
	//sign, err := btcec.SignCompact(S256(), (*btcec.PrivateKey)(pri), hash, true)
	/*priByte, err := MarshalPrikey(pri)
	if err != nil {
		return "", err
	}
	privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), priByte)*/
	hashs := DoubleHashB(hash)
	signature, err := (*btcec.PrivateKey)(pri).Sign(hashs)
	bs := signature.Serialize()
	return &bs, err
}

//验证s256签名
func VerifyS256(message, pubKeyBytes []byte, sign []byte) bool {
	pubKey, err := btcec.ParsePubKey(pubKeyBytes, btcec.S256())
	if err != nil {
		// fmt.Println("pub:", err)
		return false
	}

	// Decode hex-encoded serialized signature.
	//	sigBytes, err := hex.DecodeString(sign)

	//	if err != nil {
	//		fmt.Println("sign:", err)
	//		return false
	//	}
	signature, err := btcec.ParseSignature(sign, btcec.S256())
	if err != nil {
		// fmt.Println("parse:", err)
		return false
	}

	// Verify the signature for the message using the public key.
	messageHash := DoubleHashB(message)
	verified := signature.Verify(messageHash, pubKey)
	// fmt.Println("Signature Verified?", verified)

	return verified
}
