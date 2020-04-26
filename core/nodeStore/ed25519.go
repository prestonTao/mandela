package nodeStore

import (
	"golang.org/x/crypto/ed25519"
)

type KeyPairED struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
}

/*
	生成一对密钥
*/
func GenerateKey_ed() *KeyPairED {
	return nil
}
