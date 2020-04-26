package crypto

import (
	"fmt"
	"testing"
)

func TestAES(t *testing.T) {
	cpt, err := NewCrypter("aes", []byte("1234567891234567"))
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

func TestCrypt(t *testing.T) {
	cpt1, _ := NewCrypter("aes", []byte("1234567890123456"))
	cpt2, _ := NewCrypter("aes", []byte("1234567890123456"))

	plaintext := "1234567890123456"
	ciphertext, _ := cpt1.Encrypt([]byte(plaintext))
	plain, _ := cpt2.Decrypt(ciphertext)
	fmt.Printf("plaintext = %+v\n", string(plain))

	plaintext = "1234567890123456sadfasdfasdf"
	ciphertext, _ = cpt1.Encrypt([]byte(plaintext))
	plain, _ = cpt2.Decrypt(ciphertext)
	fmt.Printf("plaintext = %+v\n", string(plain))

	plaintext = "1234567890123456sadfasdfasdasdfajsdfasfweff"
	ciphertext, _ = cpt2.Encrypt([]byte(plaintext))
	plain, _ = cpt1.Decrypt(ciphertext)
	fmt.Printf("plaintext = %+v\n", string(plain))

}

func TestCrypt2(t *testing.T) {
	cpt1, _ := NewCrypter("aes", []byte("1234567890123456"))
	cpt2, _ := NewCrypter("aes", []byte("1234567890123456"))

	plaintext := "1234567890123456"
	ciphertext, _ := cpt1.Encrypt([]byte(plaintext))
	plain, _ := cpt2.Decrypt(ciphertext)
	fmt.Printf("ciphertext = %+v\n", ciphertext)
	fmt.Printf("plaintext = %+v\n", string(plain))

	plaintext = "1234567890123456"
	ciphertext, _ = cpt1.Encrypt([]byte(plaintext))
	plain, _ = cpt2.Decrypt(ciphertext)
	fmt.Printf("ciphertext = %+v\n", ciphertext)
	fmt.Printf("plaintext = %+v\n", string(plain))

}

func TestDES(t *testing.T) {
	cpt, err := NewCrypter("des", []byte("1234567891234567"))
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
