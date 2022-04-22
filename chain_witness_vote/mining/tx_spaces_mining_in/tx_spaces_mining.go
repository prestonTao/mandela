package tx_spaces_mining_in

import (
	"mandela/chain_witness_vote/mining"
	// "mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/binary"

	// jsoniter "github.com/json-iterator/go"
	"golang.org/x/crypto/ed25519"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*

 */
type Tx_SpacesMining struct {
	mining.TxBase
	NetId nodeStore.AddressNet `json:"netid"` //网络地址
}

/*

 */
type Tx_SpacesMining_VO struct {
	mining.TxBaseVO
	NetId string `json:"netid"` //网络地址
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_SpacesMining) GetVOJSON() interface{} {
	return Tx_SpacesMining_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		NetId:    this.NetId.B58String(), //网络地址
	}
}

/*
	构建hash值得到交易id
*/
func (this *Tx_SpacesMining) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()

	*bs = append(*bs, this.NetId...)

	// *bs = append(*bs, this.Account...)
	// for _, one := range this.NetIds {
	// 	*bs = append(*bs, one...)
	// }
	// *bs = append(*bs, this.NetIdsMerkleHash...)
	// for _, one := range this.AddrCoins {
	// 	*bs = append(*bs, one...)
	// }
	// *bs = append(*bs, this.AddrCoinsMerkleHash...)

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_spaces_mining_in)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	对整个交易签名
*/
//func (this *Tx_vote_in) Sign(key *keystore.Address, pwd string) (*[]byte, error) {
//	bs := this.SignSerialize()
//	return key.Sign(*bs, pwd)
//}

/*
	格式化成json字符串
*/
func (this *Tx_SpacesMining) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	格式化成json字符串
*/
func (this *Tx_SpacesMining) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)

	buf.Write(this.NetId)

	// buf.Write(this.Account)
	// for _, one := range this.NetIds {
	// 	buf.Write(one)
	// }
	// buf.Write(this.NetIdsMerkleHash)

	// for _, one := range this.AddrCoins {
	// 	buf.Write(one)
	// }
	// buf.Write(this.AddrCoinsMerkleHash)
	*bs = buf.Bytes()
	return bs
}

/*
	获取签名
*/
func (this *Tx_SpacesMining) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {
	txItr, err := mining.FindTxBase(txid)
	if err != nil {
		return nil
	}

	if txItr.GetBlockHash() == nil {
		txItr = mining.GetRemoteTxAndSave(txid)
		if txItr.GetBlockHash() == nil {
			return nil
		}
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

	*signDst = append(*signDst, this.NetId...)

	// *signDst = append(*signDst, this.Account...)
	// for _, one := range this.NetIds {
	// 	*signDst = append(*signDst, one...)
	// }
	// *signDst = append(*signDst, this.NetIdsMerkleHash...)
	// for _, one := range this.AddrCoins {
	// 	*signDst = append(*signDst, one...)
	// }
	// *signDst = append(*signDst, this.AddrCoinsMerkleHash...)

	// fmt.Println("签名前的字节", len(*bs), hex.EncodeToString(*bs), "\n")
	sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return &sign

}

/*
	检查交易是否合法
*/
func (this *Tx_SpacesMining) Check() error {
	// fmt.Println("开始验证交易合法性")

	//判断vin是否太多
	// if len(this.Vin) > config.Mining_pay_vin_max {
	// 	return config.ERROR_pay_vin_too_much
	// }

	// isRenew := false //是否续费

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费
	inTotal := uint64(0)
	for i, one := range this.Vin {

		txItr, err := mining.FindTxBase(one.Txid)
		if err != nil {
			return config.ERROR_tx_format_fail

		}
		vout := (*txItr.GetVout())[one.Vout]
		//如果这个交易已经被使用，则验证不通过，否则会出现双花问题。
		inTotal = inTotal + vout.Value

		// if txItr.Class() == this.Type {
		// 	isRenew = true
		// }

		//验证公钥是否和地址对应
		addr := crypto.BuildAddr(config.AddrPre, one.Puk)
		if !bytes.Equal(addr, (*txItr.GetVout())[one.Vout].Address) {
			engine.Log.Error("ERROR_sign_fail")
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

		*sign = append(*sign, this.NetId...)

		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		if !ed25519.Verify(puk, *sign, one.Sign) {
			return config.ERROR_sign_fail
		}
	}
	//判断输入输出是否相等
	outTotal := uint64(0)
	for _, one := range this.Vout {
		outTotal = outTotal + one.Value
	}
	// engine.Log.Info("这里的手续费是否正确 %d %d %d", outTotal, inTotal, this.Gas)
	if outTotal > inTotal {
		return config.ERROR_tx_fail
	}
	this.Gas = inTotal - outTotal

	//检查这个域名是否已经被注册了。如果已经被注册了，可以续费。
	// nameinfo := name.FindNameToNet(string(this.Account))
	// if nameinfo == nil {
	// 	//域名不存在，可以注册
	// 	return nil
	// }
	// engine.Log.Info("域名已经注册 %v", *nameinfo)
	//域名已经存在，检查之前的域名是否过期，检查是否是续签
	// if nameinfo.CheckIsOvertime(mining.GetHighestBlock()) {
	// 	//已经过期
	// 	return nil
	// }

	//判断续费
	// if isRenew {
	// 	return nil
	// }
	// engine.Log.Info("域名续费 %v", isRenew)
	return config.ERROR_tx_fail
}

/*
	验证是否合法
*/
func (this *Tx_SpacesMining) GetWitness() *crypto.AddressCoin {
	witness := crypto.BuildAddr(config.AddrPre, this.Vin[0].Puk)
	// witness, err := keystore.ParseHashByPubkey(this.Vin[0].Puk)
	// if err != nil {
	// 	return nil
	// }
	return &witness
}

/*
	检查重复的交易
*/
func (this *Tx_SpacesMining) CheckRepeatedTx(txs ...mining.TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }

	// for _, one := range txs {
	// 	if one.Class() != config.Wallet_tx_type_spaces_mining_in {
	// 		continue
	// 	}
	// 	ta := one.(*Tx_SpacesMining)
	// 	if bytes.Equal(ta.Account, this.Account) {
	// 		return false
	// 	}
	// }
	// return true

	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_spaces_mining_in {
			continue
		}
		tsm := one.(*Tx_SpacesMining)
		if bytes.Equal(this.NetId, tsm.NetId) {
			return false
		}
	}
	return true
}
