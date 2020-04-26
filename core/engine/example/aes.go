package main

import (
	"fmt"
	"github.com/prestonTao/engine/crypto"
)

func main() {
	cpt, err := crypto.NewCrypter("aes", []byte("1234567891234567"))
	if err != nil {
		fmt.Printf("err = %+v\n", err)
	}

	dst, err := cpt.Encrypt([]byte("1111111sdasdfjiwejfasdjfif"))
	if err != nil {
		fmt.Printf("err = %+v\n", err)
	}

	fmt.Println(dst)

	src, err := cpt.Decrypt(dst)
	if err != nil {
		fmt.Printf("err = %+v\n", err)
	}
	fmt.Println(string(src))
}
