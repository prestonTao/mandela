package tx_name_in

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/binary"
	"encoding/hex"
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
	Account             []byte                 `json:"account"`               //账户名称
	NetIds              []nodeStore.AddressNet `json:"netids"`                //网络地址列表
	NetIdsMerkleHash    []byte                 `json:"netids_merkle_hash"`    //网络地址默克尔树hash
	AddrCoins           []crypto.AddressCoin   `json:"addrcoins"`             //网络地址列表
	AddrCoinsMerkleHash []byte                 `json:"addrcoins_merkle_hash"` //网络地址默克尔树hash
}

/*

 */
type Tx_account_VO struct {
	mining.TxBaseVO
	Account             string   `json:"account"`               //账户名称
	NetIds              []string `json:"netids"`                //网络地址列表
	NetIdsMerkleHash    string   `json:"netids_merkle_hash"`    //网络地址默克尔树hash
	AddrCoins           []string `json:"addrcoins"`             //网络地址列表
	AddrCoinsMerkleHash string   `json:"addrcoins_merkle_hash"` //网络地址默克尔树hash
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_account) GetVOJSON() interface{} {
	netids := make([]string, 0)
	for _, one := range this.NetIds {
		netids = append(netids, one.B58String())
	}
	addrs := make([]string, 0)
	for _, one := range this.AddrCoins {
		addrs = append(addrs, one.B58String())
	}

	return Tx_account_VO{
		TxBaseVO:            this.TxBase.ConversionVO(),
		Account:             string(this.Account),                         //账户名称
		NetIds:              netids,                                       //网络地址列表
		NetIdsMerkleHash:    hex.EncodeToString(this.NetIdsMerkleHash),    //网络地址默克尔树hash
		AddrCoins:           addrs,                                        //网络地址列表
		AddrCoinsMerkleHash: hex.EncodeToString(this.AddrCoinsMerkleHash), //网络地址默克尔树hash
	}
}

/*
	构建hash值得到交易id
*/
func (this *Tx_account) BuildHash() {
	bs := this.Serialize()

	*bs = append(*bs, this.Account...)
	for _, one := range this.NetIds {
		*bs = append(*bs, one...)
	}
	*bs = append(*bs, this.NetIdsMerkleHash...)
	for _, one := range this.AddrCoins {
		*bs = append(*bs, one...)
	}
	*bs = append(*bs, this.AddrCoinsMerkleHash...)

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_account)

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
	//	voteinfo := this.Vote.SignSerialize()
	buf.Write(this.Account)
	for _, one := range this.NetIds {
		buf.Write(one)
	}
	buf.Write(this.NetIdsMerkleHash)

	for _, one := range this.AddrCoins {
		buf.Write(one)
	}
	buf.Write(this.AddrCoinsMerkleHash)
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

	*signDst = append(*signDst, this.Account...)
	for _, one := range this.NetIds {
		*signDst = append(*signDst, one...)
	}
	*signDst = append(*signDst, this.NetIdsMerkleHash...)
	for _, one := range this.AddrCoins {
		*signDst = append(*signDst, one...)
	}
	*signDst = append(*signDst, this.AddrCoinsMerkleHash...)

	// fmt.Println("签名前的字节", len(*bs), hex.EncodeToString(*bs), "\n")
	sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return &sign

}

/*
	检查交易是否合法
*/
func (this *Tx_account) Check() error {
	// fmt.Println("开始验证交易合法性")

	isRenew := false //是否续费

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
		vout := (*txItr.GetVout())[one.Vout]
		//如果这个交易已经被使用，则验证不通过，否则会出现双花问题。
		if vout.Tx != nil {
			return config.ERROR_tx_is_use
		}
		inTotal = inTotal + vout.Value

		if txItr.Class() == this.Type {
			isRenew = true
		}

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

		*sign = append(*sign, this.Account...)
		for _, one := range this.NetIds {
			*sign = append(*sign, one...)
		}
		*sign = append(*sign, this.NetIdsMerkleHash...)

		for _, one := range this.AddrCoins {
			*sign = append(*sign, one...)
		}
		*sign = append(*sign, this.AddrCoinsMerkleHash...)

		puk := ed25519.PublicKey(one.Puk)
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
	nameinfo := name.FindNameToNet(string(this.Account))
	if nameinfo == nil {
		//域名不存在，可以注册
		return nil
	}
	// engine.Log.Info("域名已经注册 %v", *nameinfo)
	//域名已经存在，检查之前的域名是否过期，检查是否是续签
	if nameinfo.CheckIsOvertime(mining.GetHighestBlock()) {
		//已经过期
		return nil
	}

	//判断续费
	if isRenew {
		return nil
	}
	// engine.Log.Info("域名续费 %v", isRenew)
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
*/
func (this *Tx_account) CheckRepeatedTx(txs ...mining.TxItr) bool {
	//判断双花
	if !this.MultipleExpenditures(txs...) {
		return false
	}

	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_account {
			continue
		}
		ta := one.(*Tx_account)
		if bytes.Equal(ta.Account, this.Account) {
			return false
		}
	}
	return true
}
