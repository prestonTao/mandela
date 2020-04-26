package nodeStore

// import (
// 	"golang.org/x/crypto/curve25519"
// )

type PrivateKeyCurve [32]byte //dh私钥
type PublicKeyCurve [32]byte  //dh公钥

type KeyPairCurve struct {
	PrivateKey PrivateKeyCurve
	PublicKey  PublicKeyCurve
}

/*
	生成一对密钥
*/
func GenerateKey_curve() *KeyPairCurve {
	return nil
}
