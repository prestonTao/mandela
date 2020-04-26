/*
	参与挖矿投票
	参与挖矿的节点发出投票请求，
*/
package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
)

/*
	交押金，成为候选见证人
*/
type Tx_deposit_in struct {
	TxBase
	Puk []byte `json:"puk"` //候选见证人公钥，这个公钥需要和交易index=0的输出地址保持一致
}

/*
	交押金，成为候选见证人
*/
type Tx_deposit_in_VO struct {
	TxBaseVO
	Puk string `json:"puk"` //候选见证人公钥，这个公钥需要和交易index=0的输出地址保持一致
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_deposit_in) GetVOJSON() interface{} {
	return Tx_deposit_in_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		Puk:      hex.EncodeToString(this.Puk),
	}
}

/*
	构建hash值得到交易id
*/
func (this *Tx_deposit_in) BuildHash() {
	bs := this.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(this.Puk)
	*bs = buf.Bytes()

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_deposit_in)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	获取签名
*/
func (this *Tx_deposit_in) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {
	bs, err := db.Find(txid)
	if err != nil {
		return nil
	}
	txItr, err := ParseTxBase(bs)
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
	//本交易特有信息
	*signDst = append(*signDst, this.Puk...)

	// fmt.Println("签名前的字节", len(*signDst), hex.EncodeToString(*signDst), "\n")
	sign := keystore.Sign(*key, *signDst)

	return &sign

}

/*
	验证是否合法
*/
func (this *Tx_deposit_in) Check() error {
	// fmt.Println("开始验证交易合法性 Tx_deposit_in")

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费
	inTotal := uint64(0)
	for i, one := range this.Vin {
		txbs, err := db.Find(one.Txid)
		if err != nil {
			return config.ERROR_tx_not_exist
		}
		txItr, err := ParseTxBase(txbs)
		if err != nil {
			return config.ERROR_tx_format_fail
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
		signBs := buf.Bytes()

		signDst := this.GetSignSerialize(&signBs, uint64(i))
		//本交易特有信息
		*signDst = append(*signDst, this.Puk...)
		// fmt.Println("验证签名前的字节3", len(*signDst), hex.EncodeToString(*signDst))
		puk := ed25519.PublicKey(one.Puk)
		if !ed25519.Verify(puk, *signDst, one.Sign) {
			return config.ERROR_sign_fail
		}

	}
	//判断输入输出是否相等
	outTotal := uint64(0)
	for _, one := range this.Vout {
		outTotal = outTotal + one.Value
	}
	// fmt.Println("这里的手续费是否正确", outTotal, inTotal, this.Gas)
	if outTotal > inTotal {
		return config.ERROR_tx_fail
	}
	this.Gas = inTotal - outTotal

	return nil
}

/*
	格式化成json字符串
*/
func (this *Tx_deposit_in) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	验证是否合法
*/
func (this *Tx_deposit_in) GetWitness() *crypto.AddressCoin {
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
func (this *Tx_deposit_in) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	if !this.MultipleExpenditures(txs...) {
		return false
	}
	addrSelf := this.Vout[0].Address
	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_deposit_in {
			continue
		}
		addr := (*one.GetVout())[0].Address
		if bytes.Equal(addrSelf, addr) {
			return false
		}
	}
	return true
}

/*
	见证人出块成功，退还押金
*/
type Tx_deposit_out struct {
	TxBase
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_deposit_out) GetVOJSON() interface{} {
	return this.TxBase.ConversionVO()
}

/*
	构建hash值得到交易id
*/
func (this *Tx_deposit_out) BuildHash() {
	bs := this.Serialize()

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_deposit_out)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	对整个交易签名
*/
//func (this *Tx_deposit_out) Sign(key *keystore.Address, pwd string) (*[]byte, error) {
//	bs := this.SignSerialize()
//	return key.Sign(*bs, pwd)
//}

/*
	格式化成json字符串
*/
func (this *Tx_deposit_out) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	验证是否合法
*/
func (this *Tx_deposit_out) Check() error {
	return this.TxBase.CheckBase()
}

/*
	检查重复的交易
	@return    bool    是否验证通过
*/
func (this *Tx_deposit_out) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	if !this.MultipleExpenditures(txs...) {
		return false
	}
	return true
}

/*
	创建一个见证人押金交易
	@amount    uint64    押金额度
*/
func CreateTxDepositIn(amount, gas uint64, pwd, payload string) (*Tx_deposit_in, error) {
	if amount < config.Mining_deposit {
		// fmt.Println("交押金数量最少", config.Mining_deposit)
		//押金太少
		return nil, config.ERROR_deposit_witness
	}
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	// b := chain.balance.FindBalanceOne(key)
	// if b == nil {
	// 	// fmt.Println("++++押金不够")
	// 	return nil
	// }
	key := keystore.GetCoinbase()
	puk, ok := keystore.GetPukByAddr(key)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}

	//获取解密后的私钥
	//	prk, err := key.GetPriKey(pwd)
	//	if err != nil {
	//		return nil
	//	}
	//查找余额
	vins := make([]Vin, 0)
	// total := uint64(0)

	// keys := keystore.GetAddrAll()
	// for _, one := range keys {

	// 	bas := chain.balance.FindBalance(&one)
	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			item := v.(*TxItem)

	// 			puk, ok := keystore.GetPukByAddr(one)
	// 			if !ok {
	// 				return false
	// 			}
	// 			// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

	// 			vin := Vin{
	// 				Txid: item.Txid,     //UTXO 前一个交易的id
	// 				Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
	// 				Puk:  puk,           //公钥
	// 				//					Sign: *sign,           //签名
	// 			}
	// 			vins = append(vins, vin)

	// 			total = total + item.Value
	// 			if total >= amount+gas {
	// 				return false
	// 			}

	// 			return true
	// 		})
	// 		if total >= amount+gas {
	// 			break
	// 		}
	// 	}
	// }

	total, items := chain.balance.BuildPayVin(amount + gas)

	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}

	for _, item := range items {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
		vin := Vin{
			Txid: item.Txid,     //UTXO 前一个交易的id
			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,           //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, vin)
	}

	//	for _, one := range b.Txs {

	//	}
	// if total < (amount + gas) {
	// 	//押金不够
	// 	// fmt.Println("++++押金不够222")
	// 	return nil, config.ERROR_not_enough
	// }

	//构建交易输出
	vouts := make([]Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   amount, //输出金额 = 实际金额 * 100000000
		Address: key,    //钱包地址
	}
	vouts = append(vouts, vout)
	//检查押金是否刚刚好，多了的转账给自己
	if total > amount {
		vout := Vout{
			Value:   total - (amount + gas), //输出金额 = 实际金额 * 100000000
			Address: key,                    //钱包地址
		}
		vouts = append(vouts, vout)
	}

	var txin *Tx_deposit_in
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_deposit_in, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vouts)),               //输入交易数量
			Vin:        vins,                             //交易输入
			Vout_total: uint64(len(vouts)),               //
			Vout:       vouts,                            //
			Gas:        gas,                              //交易手续费
			LockHeight: block.Height + 100 + i,           //锁定高度
			// CreateTime: time.Now().Unix(),                //创建时间
			Payload: []byte(payload),
		}
		txin = &Tx_deposit_in{
			TxBase: base,
			Puk:    puk,
		}
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			for _, key := range keystore.GetAddr() {
				puk, ok := keystore.GetPukByAddr(key)
				if !ok {
					return nil, config.ERROR_public_key_not_exist
				}
				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key, pwd)
					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						return nil, err
					}
					sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
					//				sign := txin.GetVoutsSign(prk, uint64(i))
					txin.Vin[i].Sign = *sign
				}
			}
		}
		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	return txin, nil
}

/*
	创建一个退还押金交易
	额度超过了押金额度，那么会从自己账户余额转账到目标账户（因为考虑到押金太少还不够给手续费的情况）
	@addr      *utils.Multihash    退回到的目标账户地址
	@amount    uint64              押金额度
*/
func CreateTxDepositOut(addr string, amount, gas uint64, pwd string) (*Tx_deposit_out, error) {

	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	item := chain.balance.GetDepositIn()
	if item == nil {
		// fmt.Println("没有押金")
		//没有缴纳押金
		return nil, config.ERROR_deposit_not_exist
	}

	vins := make([]Vin, 0)
	total := uint64(item.Value)
	//查看余额够不够
	if total < (amount + gas) {
		//余额不够给手续费，需要从其他账户余额作为输入给手续费
		// var items []*TxItem
		totalAll, items := chain.balance.BuildPayVin(amount + gas - total)
		total = total + totalAll
		if total < amount+gas {
			//资金不够
			return nil, config.ERROR_not_enough // errors.New("余额不足")
		}

		for _, item := range items {
			puk, ok := keystore.GetPukByAddr(*item.Addr)
			if !ok {
				return nil, config.ERROR_public_key_not_exist
			}
			// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
			vin := Vin{
				Txid: item.Txid,     //UTXO 前一个交易的id
				Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
				Puk:  puk,           //公钥
				//					Sign: *sign,           //签名
			}
			vins = append(vins, vin)
		}
		// keys := keystore.GetAddr()
		// for i, one := range keys {
		// 	bas := chain.balance.FindBalance(&keys[i])

		// 	for _, two := range bas {
		// 		two.Txs.Range(func(k, v interface{}) bool {
		// 			item := v.(*TxItem)

		// 			puk, ok := keystore.GetPukByAddr(one)
		// 			if !ok {
		// 				return false
		// 			}

		// 			vin := Vin{
		// 				Txid: item.Txid,     //UTXO 前一个交易的id
		// 				Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		// 				Puk:  puk,           //公钥
		// 				//						Sign: *sign,           //签名
		// 			}
		// 			vins = append(vins, vin)

		// 			total = total + item.Value
		// 			if total >= amount+gas {
		// 				return false
		// 			}

		// 			return true
		// 		})
		// 		if total >= amount+gas {
		// 			break
		// 		}
		// 	}
		// }

	}

	// if total < (amount + gas) {
	// 	//押金不够
	// 	return nil, errors.New("余额不够")
	// }

	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin
	if addr == "" {
		//为空则转给自己
		dstAddr = keystore.GetAddr()[0]
	} else {
		// var err error
		// *dstAddr, err = utils.FromB58String(addr)
		dstAddr = crypto.AddressFromB58String(addr)
		// if err != nil {
		// 	// fmt.Println("解析地址失败")
		// 	return nil
		// }
	}

	bs, err := db.Find(item.Txid)
	if err != nil {
		//余额交易未找到
		return nil, config.ERROR_tx_not_exist
	}
	txItr, err := ParseTxBase(bs)
	if err != nil {
		//解析交易出错
		return nil, config.ERROR_tx_format_fail
	}

	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	//地址字符串查私钥
	puk, ok := keystore.GetPukByAddr((*txItr.GetVout())[0].Address)
	if !ok {
		//未找到地址对应的公钥
		return nil, config.ERROR_public_key_not_exist
	}

	// prvKey, err := keystore.GetPriKeyByAddress((*txItr.GetVout())[0].Address.B58String(), pwd)
	// if err != nil {
	// 	return nil
	// }
	// //给输出签名，用于下一个输入
	// //	sign := txItr.GetSign(prvKey, item.OutIndex)

	// //公钥格式化
	// pub, err := utils.MarshalPubkey(&prvKey.PublicKey)
	// if err != nil {
	// 	return nil
	// }
	vin := Vin{
		Txid: item.Txid,     //UTXO 前一个交易的id
		Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk:  puk,           //公钥
		//		Sign: *sign,         //签名
	}
	vins = append(vins, vin) // []Vin{vin} // append(vins, vin)

	//构建交易输出
	vouts := make([]Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   amount,  //输出金额 = 实际金额 * 100000000
		Address: dstAddr, //钱包地址
	}
	vouts = append(vouts, vout)

	//退还剩余的钱
	if total-amount-gas > 0 {
		resultVout := Vout{
			Value:   total - amount - gas,
			Address: keystore.GetAddr()[0],
		}
		vouts = append(vouts, resultVout)
	}

	var txin *Tx_deposit_out
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_deposit_out, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                 //输入交易数量
			Vin:        vins,                              //交易输入
			Vout_total: uint64(len(vouts)),                //
			Vout:       vouts,                             //
			Gas:        gas,                               //交易手续费
			LockHeight: block.Height + 100 + i,            //锁定高度
			// CreateTime: time.Now().Unix(),                 //创建时间
		}
		txin = &Tx_deposit_out{
			TxBase: base,
		}
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			for _, key := range keystore.GetAddr() {

				puk, ok := keystore.GetPukByAddr(key)
				if !ok {
					//未找到地址对应的公钥
					return nil, config.ERROR_public_key_not_exist
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key, pwd)

					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						return nil, err
					}
					sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
					//				sign := txin.GetVoutsSign(prk, uint64(i))
					txin.Vin[i].Sign = *sign
				}
			}
		}
		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	return txin, nil
}
