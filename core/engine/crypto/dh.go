// Package crypto provides  迪菲－赫尔曼密钥交换
// http://zh.wikipedia.org/wiki/%E8%BF%AA%E8%8F%B2%EF%BC%8D%E8%B5%AB%E5%B0%94%E6%9B%BC%E5%AF%86%E9%92%A5%E4%BA%A4%E6%8D%A2

package crypto

import (
	"errors"
	"math/big"

	"github.com/monnand/dhkx"
)

type DH struct {
	dhGroup *dhkx.DHGroup
	dhPriv  *dhkx.DHKey
	ch      chan string
}

func NewDH() (*DH, error) {
	g, err := dhkx.GetGroup(1)

	if err != nil {
		return nil, err
	}

	priv, err := g.GeneratePrivateKey(nil)
	if err != nil {
		return nil, err
	}

	dh := &DH{
		dhGroup: g,
		dhPriv:  priv,
		ch:      make(chan string, 10),
	}

	return dh, nil
}

func (dh *DH) ComputeKey(public string) ([]byte, error) {
	publicInt, ok := big.NewInt(0).SetString(public, 10)
	if !ok {
		//大数构成错误
		return nil, errors.New("Large numbers constitute errors")
	}

	pub := dhkx.NewPublicKey(publicInt.Bytes())
	k, err := dh.dhGroup.ComputeKey(pub, dh.dhPriv)
	if err != nil {
		return nil, err
	}

	return k.Bytes(), nil
}

func (dh *DH) PrivKey() string {
	return dh.dhPriv.String()
}
