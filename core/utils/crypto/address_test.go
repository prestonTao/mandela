package crypto

import (
	"mandela/core/utils/base58"
	"crypto/rand"
	"fmt"
	"testing"

	"golang.org/x/crypto/ed25519"
)

func TestAddr(t *testing.T) {
	puk, _, _ := ed25519.GenerateKey(rand.Reader)
	version := []byte{0, 0}
	addr, _ := BuildAddr(version, puk)
	addrStr := base58.Encode(addr)
	fmt.Println(string(addrStr))

	ok := ValidAddr(version, addr)
	fmt.Println("验证是否通过", ok)
}
