package node_shadow

import (
	"mandela/core/utils/base58"
	"bytes"
	"crypto/sha256"
)

//影子节点地址
type AddressNetShadow []byte

func (this *AddressNetShadow) B58String() string {
	return string(base58.Encode(*this))
}

func AddressFromB58String(str string) AddressNetShadow {
	return AddressNetShadow(base58.Decode(str))
}
