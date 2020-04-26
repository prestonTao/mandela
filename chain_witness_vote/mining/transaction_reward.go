/*
	矿工费交易
*/
package mining

import (
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/json"
)

/*
	矿工费交易
	没有输入，只有输出
*/
type Tx_reward struct {
	TxBase
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_reward) GetVOJSON() interface{} {
	return this.TxBase.ConversionVO()
}

/*
	格式化成json字符串
*/
func (this *Tx_reward) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	获取签名
*/
func (this *Tx_reward) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {

	signDst := this.GetSignSerialize(nil, vinIndex)

	// fmt.Println("签名前的字节", len(*bs), hex.EncodeToString(*bs), "\n")
	sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return &sign
}

/*
	检查交易是否合法
*/
func (this *Tx_reward) Check() error {
	// fmt.Println("开始验证交易合法性 Tx_reward")
	//检查输入输出是否对等，还有手续费

	if this.Vin == nil || len(this.Vin) != 1 {
		return config.ERROR_tx_fail
	}

	one := this.Vin[0]
	signDst := this.GetSignSerialize(nil, uint64(0))

	puk := ed25519.PublicKey(one.Puk)
	if !ed25519.Verify(puk, *signDst, one.Sign) {
		return config.ERROR_sign_fail
	}

	outTotal := uint64(0)
	for _, one := range this.Vout {
		outTotal = outTotal + one.Value
	}

	return nil
}

/*
	构建hash值得到交易id
*/
func (this *Tx_reward) BuildHash() {
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_mining)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	是否验证通过
*/
func (this *Tx_reward) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}
