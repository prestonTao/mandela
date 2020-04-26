package tx_name_out

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/binary"
	"encoding/json"

	"golang.org/x/crypto/ed25519"
)

/*
	交押金，注册一个域名
	域名为抢注式，只需要一点点手续费就能注册。注册时间为一年到期，快要到期时续费。
	注册域名需要押金，押金不管多少。
	域名可以转让，可以修改解析的网络地址和收款账号。不能注销，到期自动失效。

*/
type Tx_account struct {
	mining.TxBase
	// Account []byte `json:"account"` //账户名称
}

/*

 */
type Tx_account_VO struct {
	mining.TxBaseVO
	// Account string `json:"account"` //账户名称
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_account) GetVOJSON() interface{} {
	return Tx_account_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		// Account:  string(this.Account), //账户名称
	}
}

/*
	构建hash值得到交易id
*/
func (this *Tx_account) BuildHash() {
	bs := this.Serialize()

	// *bs = append(*bs, this.Account...)

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_account_cancel)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	格式化成json字符串
*/
func (this *Tx_account) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	格式化成json字符串
*/
func (this *Tx_account) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	// buf.Write(this.Account)
	*bs = buf.Bytes()
	return bs
}

/*
	获取签名
*/
func (this *Tx_account) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {
	bs, err := db.Find(txid)
	if err != nil {
		return nil
	}
	txItr, err := mining.ParseTxBase(bs)
	if err != nil {
		return nil
	}
	buf := bytes.NewBuffer(nil)
	//上一个交易 所属的区块hash
	buf.Write(*txItr.GetBlockHash())
	//上一个交易的hash
	buf.Write(*txItr.GetHash())
	//上一个交易的指定输出序列化
	buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	//本交易类型输入输出数量等信息和所有输出
	signBs := buf.Bytes()
	signDst := this.GetSignSerialize(&signBs, vinIndex)

	// *signDst = append(*signDst, this.Account...)

	// fmt.Println("签名前的字节", len(*bs), hex.EncodeToString(*bs), "\n")
	sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return &sign
}

/*
	检查交易是否合法
*/
func (this *Tx_account) Check() error {
	// this.TxBase.CheckBase()

	oldTxClass := uint64(config.Wallet_tx_type_mining)
	//
	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费
	inTotal := uint64(0)
	for i, one := range this.Vin {
		txbs, err := db.Find(one.Txid)
		if err != nil {
			return config.ERROR_tx_not_exist
		}
		txItr, err := mining.ParseTxBase(txbs)
		if err != nil {
			return config.ERROR_tx_format_fail
		}

		//记录上一笔交易的交易类型
		if i == 0 {
			oldTxClass = txItr.Class()
		}

		vout := (*txItr.GetVout())[one.Vout]
		//如果这个交易已经被使用，则验证不通过，否则会出现双花问题。
		if vout.Tx != nil {
			return config.ERROR_tx_is_use
		}
		inTotal = inTotal + vout.Value

		//验证公钥是否和地址对应
		addr := crypto.BuildAddr(config.AddrPre, one.Puk)
		if !bytes.Equal(addr, (*txItr.GetVout())[one.Vout].Address) {
			return config.ERROR_sign_fail
		}

		//验证签名
		buf := bytes.NewBuffer(nil)
		//上一个交易 所属的区块hash
		buf.Write(*txItr.GetBlockHash())
		//上一个交易的hash
		buf.Write(*txItr.GetHash())
		//上一个交易的指定输出序列化
		buf.Write(*txItr.GetVoutSignSerialize(one.Vout))
		//本交易类型输入输出数量等信息和所有输出
		bs := buf.Bytes()
		sign := this.GetSignSerialize(&bs, uint64(i))

		// *sign = append(*sign, this.Account...)

		puk := ed25519.PublicKey(one.Puk)
		if !ed25519.Verify(puk, *sign, one.Sign) {
			return config.ERROR_sign_fail
		}

	}
	if this.Type == config.Wallet_tx_type_account_cancel && oldTxClass == config.Wallet_tx_type_account {
		return nil
	}

	// //确认是本人取消

	//还要检查域名是否属于那个人

	return config.ERROR_tx_fail
}

/*
	验证是否合法
*/
func (this *Tx_account) GetWitness() *crypto.AddressCoin {
	witness := crypto.BuildAddr(config.AddrPre, this.Vin[0].Puk)
	// witness, err := keystore.ParseHashByPubkey(this.Vin[0].Puk)
	// if err != nil {
	// 	return nil
	// }
	return &witness
}

/*
	检查重复的交易
	@return    bool    是否验证通过
*/
func (this *Tx_account) CheckRepeatedTx(txs ...mining.TxItr) bool {
	//判断双花
	if !this.MultipleExpenditures(txs...) {
		return false
	}
	return true
}
