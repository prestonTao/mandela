package doubleratchet

import (
	"mandela/core/utils/crypto/dh"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

// DefaultCrypto is an implementation of Crypto with cryptographic primitives recommended
// by the Double Ratchet Algorithm specification. However, some details are different,
// see function comments for details.
//是双棘轮算法规范推荐使用加密原语的加密实现。但是，有些细节不同，有关详细信息，请参见功能注释。
type DefaultCrypto struct{}

//创建一个新的dh密钥对
func (c DefaultCrypto) GenerateDH() (DHPair, error) {
	var privKey [32]byte
	if _, err := io.ReadFull(rand.Reader, privKey[:]); err != nil {
		return dh.DHPair{}, fmt.Errorf("couldn't generate privKey: %s", err)
	}
	privKey[0] &= 248
	privKey[31] &= 127
	privKey[31] |= 64

	var pubKey [32]byte
	curve25519.ScalarBaseMult(&pubKey, &privKey)
	// return dh.DHPair{
	// 	privateKey: privKey,
	// 	publicKey:  pubKey,
	// } , nil

	return dh.NewDHPair(privKey, pubKey), nil
}

//计算公共密钥
func (c DefaultCrypto) DH(dhPair DHPair, dhPub dh.Key) dh.Key {
	var (
		dhOut   [32]byte
		privKey [32]byte = dhPair.GetPrivateKey()
		pubKey  [32]byte = dhPub
	)
	curve25519.ScalarMult(&dhOut, &privKey, &pubKey)
	return dhOut
}

// KdfRK returns a pair (32-byte root key, 32-byte chain key) as the output of applying
// a KDF keyed by a 32-byte root key rk to a Diffie-Hellman output dhOut.
//
func (c DefaultCrypto) KdfRK(rk, dhOut dh.Key) (rootKey, chainKey, headerKey dh.Key) {
	var (
		r   = hkdf.New(sha256.New, dhOut[:], rk[:], []byte("3Dsjd9qaor3bS8NTwsDGbTdGPA3tXiXGKdGJHWDhe6M8"))
		buf = make([]byte, 96)
	)

	// The only error here is an entropy limit which won't be reached for such a short buffer.
	_, _ = io.ReadFull(r, buf)

	copy(rootKey[:], buf[:32])
	copy(chainKey[:], buf[32:64])
	copy(headerKey[:], buf[64:96])
	return
}

// KdfCK returns a pair (32-byte chain key, 32-byte message key) as the output of applying
// a KDF keyed by a 32-byte chain key ck to some constant.
func (c DefaultCrypto) KdfCK(ck dh.Key) (chainKey dh.Key, msgKey dh.Key) {
	const (
		ckInput = 15
		mkInput = 16
	)

	h := hmac.New(sha256.New, ck[:])

	_, _ = h.Write([]byte{ckInput})
	copy(chainKey[:], h.Sum(nil))
	h.Reset()

	_, _ = h.Write([]byte{mkInput})
	copy(msgKey[:], h.Sum(nil))

	return chainKey, msgKey
}

// Encrypt使用与算法规范稍有不同的方法：它使用aes-256-ctr而不是aes-256-cbc来考虑安全性、密文长度和实现复杂性。
func (this DefaultCrypto) Encrypt(mk dh.Key, plaintext, ad []byte) []byte {
	encKey, authKey, iv := this.deriveEncKeys(mk)

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	copy(ciphertext, iv[:])

	var (
		block, _ = aes.NewCipher(encKey[:]) // No error will occur here as encKey is guaranteed to be 32 bytes.
		stream   = cipher.NewCTR(block, iv[:])
	)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return append(ciphertext, this.computeSignature(authKey[:], ciphertext, ad)...)
}

// Decrypt returns the AEAD decryption of ciphertext with message key mk.
func (c DefaultCrypto) Decrypt(mk dh.Key, authCiphertext, ad []byte) ([]byte, error) {
	var (
		l          = len(authCiphertext)
		ciphertext = authCiphertext[:l-sha256.Size]
		signature  = authCiphertext[l-sha256.Size:]
	)

	// Check the signature.
	encKey, authKey, _ := c.deriveEncKeys(mk)

	if s := c.computeSignature(authKey[:], ciphertext, ad); !bytes.Equal(s, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decrypt.
	var (
		block, _  = aes.NewCipher(encKey[:]) // No error will occur here as encKey is guaranteed to be 32 bytes.
		stream    = cipher.NewCTR(block, ciphertext[:aes.BlockSize])
		plaintext = make([]byte, len(ciphertext[aes.BlockSize:]))
	)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	return plaintext, nil
}

// deriveEncKeys派生用于消息加密和解密的密钥。返回（enckey、authkey、iv）。
func (c DefaultCrypto) deriveEncKeys(mk dh.Key) (encKey dh.Key, authKey dh.Key, iv [16]byte) {
	// 首先，从mk导出加密和身份验证密钥。
	salt := make([]byte, 32)
	var (
		r   = hkdf.New(sha256.New, mk[:], salt, []byte("CrPJtHJJp2PeuLVTzyMV8cFVcX5uv8NmtB17W8S9obYG"))
		buf = make([]byte, 80)
	)

	// 这里唯一的错误是熵限，对于如此短的缓冲区是无法达到的。
	_, _ = io.ReadFull(r, buf)

	copy(encKey[:], buf[0:32])
	copy(authKey[:], buf[32:64])
	copy(iv[:], buf[64:80])
	return
}

func (c DefaultCrypto) computeSignature(authKey, ciphertext, associatedData []byte) []byte {
	h := hmac.New(sha256.New, authKey)
	_, _ = h.Write(associatedData)
	_, _ = h.Write(ciphertext)
	return h.Sum(nil)
}

// type dhPair struct {
// 	privateKey dh.Key
// 	publicKey  dh.Key
// }

// func (p dhPair) PrivateKey() dh.Key {
// 	return p.privateKey
// }

// func (p dhPair) PublicKey() dh.Key {
// 	return p.publicKey
// }

// func (p dhPair) String() string {
// 	return fmt.Sprintf("{privateKey: %s publicKey: %s}", p.privateKey, p.publicKey)
// }
