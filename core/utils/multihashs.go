package utils

import (
	"encoding/hex"

	b58 "github.com/mr-tron/base58/base58"
)

type Multihashs struct {
	Multihash `json:"mutihash"`
	B58string string `json:"b58string"`
}

func NewMultihashs(data []byte) Multihashs {
	multihashs := Multihashs{}
	multihashs.Multihash = Multihash(data)
	multihashs.B58string = b58.Encode(data)
	return multihashs
}
func (m *Multihashs) Byte() []byte {
	return m.Multihash
}
func (m *Multihashs) B58String() string {
	if m.B58string != "" {
		return m.B58string
	} else {
		return b58.Encode([]byte(m.B58string))
	}
}
func FromHexStrings(s string) (Multihashs, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return Multihashs{}, err
	}

	return Casts(b)
}
func FromB58Strings(s string) (m Multihashs, err error) {
	b, err := b58.Decode(s)
	if err != nil {
		return Multihashs{}, ErrInvalidMultihash
	}

	return Casts(b)
}
func Casts(buf []byte) (Multihashs, error) {
	dm, err := Decode(buf)
	if err != nil {
		return Multihashs{}, err
	}

	if !ValidCode(dm.Code) {
		return Multihashs{}, ErrUnknownCode
	}

	return NewMultihashs(buf), nil
}
