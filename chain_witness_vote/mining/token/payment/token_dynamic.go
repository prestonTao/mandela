package payment

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/token/publish"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"strconv"

	// jsoniter "github.com/json-iterator/go"
	"golang.org/x/crypto/ed25519"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
	交押金，注册一个域名
	域名为抢注式，只需要一点点手续费就能注册。注册时间为一年到期，快要到期时续费。
	注册域名需要押金，押金不管多少。
	域名可以转让，可以修改解析的网络地址和收款账号。不能注销，到期自动失效。
*/
type TxToken struct {
	mining.TxBase
	Token_Vin_total  uint64         `json:"token_vin_total"`  //输入交易数量
	Token_Vin        []*mining.Vin  `json:"token_vin"`        //交易输入
	Token_Vout_total uint64         `json:"token_vout_total"` //输出交易数量
	Token_Vout       []*mining.Vout `json:"token_vout"`       //交易输出
	// Token_publish_txid     []byte         `json:"token_publish_txid"` //token的合约地址
	// Token_publish_txid_str string `json:"-"` //
}

// func (this *TxToken) GetPublishTxidStr() string {
// 	if this.Token_publish_txid_str == "" {
// 		this.Token_publish_txid_str = hex.EncodeToString(this.Token_publish_txid)
// 	}
// 	return this.Token_publish_txid_str
// }

/*

 */
type TxToken_VO struct {
	mining.TxBaseVO
	Token_name         string           `json:"token_name"`         //名称
	Token_symbol       string           `json:"token_symbol"`       //单位
	Token_supply       uint64           `json:"token_supply"`       //发行总量
	Token_Vin_total    uint64           `json:"token_vin_total"`    //输入交易数量
	Token_Vin          []*mining.VinVO  `json:"token_vin"`          //交易输入
	Token_Vout_total   uint64           `json:"token_vout_total"`   //输出交易数量
	Token_Vout         []*mining.VoutVO `json:"token_vout"`         //交易输出
	Token_publish_txid string           `json:"token_publish_txid"` //token的合约地址

}

/*
	用于地址和txid格式化显示
*/
func (this *TxToken) GetVOJSON() interface{} {

	vins := make([]*mining.VinVO, 0)
	for _, one := range this.Token_Vin {
		vins = append(vins, one.ConversionVO())
	}
	vouts := make([]*mining.VoutVO, 0)
	for _, one := range this.Token_Vout {
		vouts = append(vouts, one.ConversionVO())
	}

	return TxToken_VO{
		TxBaseVO:         this.TxBase.ConversionVO(),
		Token_Vin_total:  this.Token_Vin_total,  //输入交易数量
		Token_Vin:        vins,                  //交易输入
		Token_Vout_total: this.Token_Vout_total, //输出交易数量
		Token_Vout:       vouts,                 //交易输出
		// Token_publish_txid: hex.EncodeToString(this.Token_publish_txid), //token的合约地址
	}
}

/*
	构建hash值得到交易id
*/
func (this *TxToken) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()

	// *bs = append(*bs, this.Token_name...)
	// *bs = append(*bs, this.Token_symbol...)
	// *bs = append(*bs, utils.Uint64ToBytes(this.Token_supply)...)

	*bs = append(*bs, utils.Uint64ToBytes(this.Token_Vin_total)...)
	if this.Token_Vin != nil {
		for _, one := range this.Token_Vin {
			*bs = append(*bs, *one.SerializeVin()...)
		}
	}
	*bs = append(*bs, utils.Uint64ToBytes(this.Token_Vout_total)...)
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			*bs = append(*bs, *one.Serialize()...)
		}
	}

	id := make([]byte, 8)
	binary.PutUvarint(id, Wallet_tx_class)

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
// func (this *TxToken) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	格式化成[]byte
*/
func (this *TxToken) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vins = append(vins, &go_protos.Vin{
			Txid: one.Txid,
			Vout: one.Vout,
			Puk:  one.Puk,
			Sign: one.Sign,
		})
	}
	vouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	txBase := go_protos.TxBase{
		Hash:       this.Hash,
		Type:       this.Type,
		VinTotal:   this.Vin_total,
		Vin:        vins,
		VoutTotal:  this.Vout_total,
		Vout:       vouts,
		Gas:        this.Gas,
		LockHeight: this.LockHeight,
		Payload:    this.Payload,
		BlockHash:  this.BlockHash,
	}
	tokenVins := make([]*go_protos.Vin, 0)
	for _, one := range this.Token_Vin {
		tokenVins = append(tokenVins, &go_protos.Vin{
			Txid: one.Txid,
			Vout: one.Vout,
			Puk:  one.Puk,
			Sign: one.Sign,
		})
	}
	tokenVouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Token_Vout {
		tokenVouts = append(tokenVouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}

	txPay := go_protos.TxTokenPay{
		TxBase:          &txBase,
		Token_VinTotal:  this.Token_Vin_total,
		Token_Vin:       tokenVins,
		Token_VoutTotal: this.Token_Vout_total,
		Token_Vout:      tokenVouts,
	}
	// txPay.Marshal()
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	格式化成json字符串
*/
func (this *TxToken) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)

	// buf.Write([]byte(this.Token_name))
	// buf.Write([]byte(this.Token_symbol))
	// buf.Write(utils.Uint64ToBytes(this.Token_supply))

	buf.Write(utils.Uint64ToBytes(this.Token_Vin_total))
	if this.Token_Vin != nil {
		for _, one := range this.Token_Vin {
			bs := one.SerializeVin()
			buf.Write(*bs)
		}
	}
	buf.Write(utils.Uint64ToBytes(this.Token_Vout_total))
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			bs := one.Serialize()
			buf.Write(*bs)
		}
	}

	*bs = buf.Bytes()
	return bs
}

/*
	获取签名
*/
func (this *TxToken) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {

	txItr, err := mining.LoadTxBase(txid)
	// txItr, err := mining.FindTxBase(txid)
	if err != nil {
		return nil
	}

	blockhash, err := db.GetTxToBlockHash(&txid)
	if err != nil || blockhash == nil {
		return nil
	}

	// if txItr.GetBlockHash() == nil {
	// 	txItr = mining.GetRemoteTxAndSave(txid)
	// 	if txItr.GetBlockHash() == nil {
	// 		return nil
	// 	}
	// }
	// txItr.SetBlockHash([]byte{})
	// fmt.Println("000000000000000000", len(*txItr.GetBlockHash()))

	buf := bytes.NewBuffer(nil)
	//上一个交易 所属的区块hash
	buf.Write(*blockhash)
	// buf.Write(*txItr.GetBlockHash())
	//上一个交易的hash
	buf.Write(*txItr.GetHash())
	//上一个交易的指定输出序列化
	buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	//本交易类型输入输出数量等信息和所有输出
	signBs := buf.Bytes()
	signDst := this.GetSignSerialize(&signBs, vinIndex)

	// engine.Log.Info("222222 %s", hex.EncodeToString(*signDst))

	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vin_total)...)
	if this.Token_Vin != nil {
		for _, one := range this.Token_Vin {
			*signDst = append(*signDst, *one.SerializeVin()...)
		}
	}
	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vout_total)...)
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			*signDst = append(*signDst, *one.Serialize()...)
		}
	}

	// engine.Log.Info("签名之前序列化字符串 %d %s", len(*signDst), hex.EncodeToString(*signDst))

	sign := keystore.Sign(*key, *signDst)

	return &sign

}

/*
	获取Token签名
*/
func (this *TxToken) GetTokenSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {
	txItr, err := mining.LoadTxBase(txid)
	// txItr, err := mining.FindTxBase(txid)
	// bs, err := db.Find(txid)
	// if err != nil {
	// 	return nil
	// }
	// txItr, err := mining.ParseTxBase(mining.ParseTxClass(txid), bs)
	if err != nil {
		return nil
	}

	blockhash, err := db.GetTxToBlockHash(&txid)
	if err != nil || blockhash == nil {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	//上一个交易 所属的区块hash
	buf.Write(*blockhash)
	// buf.Write(*txItr.GetBlockHash())
	//上一个交易的hash
	buf.Write(*txItr.GetHash())

	if txItr.Class() == config.Wallet_tx_type_token_publish {
		txtoken := txItr.(*publish.TxToken)
		//上一个交易的指定输出序列化
		buf.Write(*txtoken.GetTokenVoutSignSerialize(voutIndex))
	} else if txItr.Class() == config.Wallet_tx_type_token_payment {
		txtoken := txItr.(*TxToken)
		//上一个交易的指定输出序列化
		buf.Write(*txtoken.GetTokenVoutSignSerialize(voutIndex))
	}

	//本交易类型输入输出数量等信息和所有输出
	signBs := buf.Bytes()
	signDst := this.GetSignSerialize(&signBs, vinIndex)

	// *signDst = append(*signDst, this.Token_name...)
	// *signDst = append(*signDst, this.Token_symbol...)
	// *signDst = append(*signDst, utils.Uint64ToBytes(this.Token_supply)...)

	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vin_total)...)
	if this.Token_Vin != nil {
		for _, one := range this.Token_Vin {
			*signDst = append(*signDst, *one.SerializeVin()...)
		}
	}
	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vout_total)...)
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			*signDst = append(*signDst, *one.Serialize()...)
		}
	}

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
	获取待签名数据
*/
func (this *TxToken) GetWaitTokenSign(txid []byte, voutIndex, vinIndex uint64) *[]byte {
	txItr, err := mining.LoadTxBase(txid)
	// txItr, err := mining.FindTxBase(txid)
	// bs, err := db.Find(txid)
	// if err != nil {
	// 	return nil
	// }
	// txItr, err := mining.ParseTxBase(mining.ParseTxClass(txid), bs)
	if err != nil {
		return nil
	}
	blockhash, err := db.GetTxToBlockHash(&txid)
	if err != nil || blockhash == nil {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	//上一个交易 所属的区块hash
	buf.Write(*blockhash)
	// buf.Write(*txItr.GetBlockHash())
	//上一个交易的hash
	buf.Write(*txItr.GetHash())

	//上一个交易的指定输出序列化
	buf.Write(*this.GetTokenVoutSignSerialize(voutIndex))

	//本交易类型输入输出数量等信息和所有输出
	signBs := buf.Bytes()
	signDst := this.GetSignSerialize(&signBs, vinIndex)

	// *signDst = append(*signDst, this.Token_name...)
	// *signDst = append(*signDst, this.Token_symbol...)
	// *signDst = append(*signDst, utils.Uint64ToBytes(this.Token_supply)...)

	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vin_total)...)
	if this.Token_Vin != nil {
		for i, one := range this.Token_Vin {
			this.Token_Vin[i].Sign = nil
			*signDst = append(*signDst, *one.SerializeVin()...)
		}
	}
	*signDst = append(*signDst, utils.Uint64ToBytes(this.Token_Vout_total)...)
	if this.Token_Vout != nil {
		for _, one := range this.Token_Vout {
			*signDst = append(*signDst, *one.Serialize()...)
		}
	}

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
	//sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	//return &sign
	return signDst
}
func (this *TxToken) GetTokenVoutSignSerialize(voutIndex uint64) *[]byte {
	bufVout := bytes.NewBuffer(nil)
	//上一个交易的指定输出序列化
	bufVout.Write(utils.Uint64ToBytes(voutIndex))
	vout := this.Token_Vout[voutIndex]
	bs := vout.Serialize()
	bufVout.Write(*vout.Serialize())
	*bs = bufVout.Bytes()
	return bs
}

/*
	把相同地址的交易输出合并在一起
*/
func (this *TxToken) MergeTokenVout() {
	this.Token_Vout = mining.MergeVouts(&this.Token_Vout)
	this.Token_Vout_total = uint64(len(this.Token_Vout))
}

/*
	检查交易是否合法
*/
func (this *TxToken) Check() error {
	// start := time.Now()
	//判断vin是否太多
	// if len(this.Token_Vin)+len(this.Vin) > config.Mining_pay_vin_max {
	// 	return config.ERROR_pay_vin_too_much
	// }

	for i, _ := range this.Token_Vin {
		this.Token_Vin[i].Sign = nil
	}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费;3.输入不能重复。
	vinMap := make(map[string]int)
	inTotal := uint64(0)
	for i, one := range this.Vin {

		//不能有重复的vin
		key := string(mining.BuildKeyForUnspentTransaction(one.Txid, one.Vout))
		if _, ok := vinMap[key]; ok {
			return config.ERROR_tx_Repetitive_vin
		}
		vinMap[key] = 0

		txItr, err := mining.LoadTxBase(one.Txid)
		// txItr, err := mining.FindTxBase(one.Txid)
		if err != nil {
			return config.ERROR_tx_format_fail
		}

		blockhash, err := db.GetTxToBlockHash(&one.Txid)
		if err != nil || blockhash == nil {
			return config.ERROR_tx_format_fail
		}

		vout := (*txItr.GetVout())[one.Vout]
		//如果这个交易已经被使用，则验证不通过，否则会出现双花问题。

		inTotal = inTotal + vout.Value

		//验证公钥是否和地址对应
		addr := crypto.BuildAddr(config.AddrPre, one.Puk)
		if !bytes.Equal(addr, (*txItr.GetVout())[one.Vout].Address) {
			engine.Log.Error("ERROR_sign_fail")
			return config.ERROR_sign_fail
		}

		//验证签名
		buf := bytes.NewBuffer(nil)
		//上一个交易 所属的区块hash
		buf.Write(*blockhash)
		// buf.Write(*txItr.GetBlockHash())
		//上一个交易的hash
		buf.Write(*txItr.GetHash())
		//上一个交易的指定输出序列化
		buf.Write(*txItr.GetVoutSignSerialize(one.Vout))
		//本交易类型输入输出数量等信息和所有输出
		bs := buf.Bytes()
		sign := this.GetSignSerialize(&bs, uint64(i))

		*sign = append(*sign, utils.Uint64ToBytes(this.Token_Vin_total)...)
		if this.Token_Vin != nil {
			for _, one := range this.Token_Vin {
				*sign = append(*sign, *one.SerializeVin()...)
			}
		}
		*sign = append(*sign, utils.Uint64ToBytes(this.Token_Vout_total)...)
		if this.Token_Vout != nil {
			for _, one := range this.Token_Vout {
				*sign = append(*sign, *one.Serialize()...)
			}
		}

		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		if !ed25519.Verify(puk, *sign, one.Sign) {
			return config.ERROR_sign_fail
		}
	}

	// engine.Log.Info("验证token转账耗时 111111111111111 %s", time.Now().Sub(start))

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

	//token交易vin和vout额度相等
	//同时判断vin中所有token种类相同
	var publishTxid *[]byte
	tokenVoutTotal := uint64(0)
	for _, one := range this.Token_Vout {
		tokenVoutTotal = tokenVoutTotal + one.Value
	}

	//输入不能重复。
	vinMap = make(map[string]int)
	tokenVinTotal := uint64(0)
	for vinIndex, vin := range this.Token_Vin {

		keyStr := config.TokenPublishTxid + utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))

		//不能有重复的vin
		if _, ok := vinMap[keyStr]; ok {
			return config.ERROR_tx_Repetitive_vin
		}
		vinMap[keyStr] = 0

		publishTxidBs, err := db.LevelTempDB.Find([]byte(keyStr))
		if err != nil {
			return err
		}
		//先判断种类是否相同
		if vinIndex == 0 {
			publishTxid = publishTxidBs
		} else {
			if !bytes.Equal(*publishTxid, *publishTxidBs) {
				return config.ERROR_tx_fail
			}
		}

		// engine.Log.Info("验证token转账耗时 222222222222222 %s", time.Now().Sub(start))

		//查询余额
		txItr, err := mining.LoadTxBase(vin.Txid)
		// txItr, err := mining.FindTxBase(vin.Txid)
		if err != nil {
			return config.ERROR_tx_format_fail
		}

		// bs, err := json.Marshal(txItr)
		// if err != nil {
		// 	return err
		// }
		// txToken := new(TxToken)
		// decoder := json.NewDecoder(bytes.NewBuffer(bs))
		// decoder.UseNumber()
		// err = decoder.Decode(txToken)
		// if err != nil {
		// 	return err
		// }

		if mining.ParseTxClass(vin.Txid) == config.Wallet_tx_type_token_publish {
			txToken := txItr.(*publish.TxToken)
			vout := (txToken.Token_Vout)[vin.Vout]
			tokenVinTotal = tokenVinTotal + vout.Value
		} else if mining.ParseTxClass(vin.Txid) == config.Wallet_tx_type_token_payment {
			txToken := txItr.(*TxToken)
			vout := (txToken.Token_Vout)[vin.Vout]
			tokenVinTotal = tokenVinTotal + vout.Value
		} else {
			return config.ERROR_tx_fail
		}

	}
	if tokenVoutTotal != tokenVinTotal {
		return config.ERROR_tx_fail
	}

	// engine.Log.Info("验证token转账耗时 3333333333333333 %s", time.Now().Sub(start))

	return nil
}

/*
	验证是否合法
*/
func (this *TxToken) GetWitness() *crypto.AddressCoin {
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
func (this *TxToken) CheckRepeatedTx(txs ...mining.TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }

	for _, one := range txs {
		if one.Class() != Wallet_tx_class {
			continue
		}
		// ta := one.(*TxToken)
		// if bytes.Equal(ta.Account, this.Account) {
		// 	return false
		// }
	}
	return true
}

/*
	统计交易余额
*/
func (this *TxToken) CountTxItems(height uint64) *mining.TxItemCount {
	itemCount := mining.TxItemCount{
		Additems: make([]*mining.TxItem, 0),
		SubItems: make([]*mining.TxSubItems, 0),
	}
	//将之前的UTXO标记为已经使用，余额中减去。
	for _, vin := range this.Vin {
		// engine.Log.Info("查看vin中的状态 %d", vin.PukIsSelf)
		ok := vin.CheckIsSelf()
		if !ok {
			continue
		}
		// engine.Log.Info("统单易1耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
		//查找这个地址的余额列表，没有则创建一个
		itemCount.SubItems = append(itemCount.SubItems, &mining.TxSubItems{
			Txid:      vin.Txid, //utils.Bytes2string(vin.Txid), //  vin.GetTxidStr(),
			VoutIndex: vin.Vout,
			Addr:      *vin.GetPukToAddr(), // utils.Bytes2string(*vin.GetPukToAddr()), // vin.GetPukToAddrStr(),
		})
	}

	//生成新的UTXO收益，保存到列表中
	for voutIndex, vout := range this.Vout {
		// if voutIndex == 0 {
		// 	continue
		// }
		//找出需要统计余额的地址
		//和自己无关的地址
		ok := vout.CheckIsSelf()
		if !ok {
			continue
		}

		// engine.Log.Info("统单易5耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
		txItem := mining.TxItem{
			Addr: &(this.Vout)[voutIndex].Address, //  &vout.Address,
			// AddrStr: vout.GetAddrStr(),                      //
			Value: vout.Value,      //余额
			Txid:  *this.GetHash(), //交易id
			// TxidStr:      txHashStr,                              //
			VoutIndex:    uint64(voutIndex), //交易输出index，从0开始
			Height:       height,            //
			LockupHeight: vout.FrozenHeight, //锁仓高度
		}

		//计入余额列表
		// this.notspentBalance.AddTxItem(txItem)
		itemCount.Additems = append(itemCount.Additems, &txItem)

		//保存到缓存
		// engine.Log.Info("开始统计交易余额 区块高度 %d 保存到缓存", bhvo.BH.Height)
		// TxCache.AddTxInTxItem(txHashStr, txItr)
		mining.TxCache.AddTxInTxItem(*this.GetHash(), this)

	}
	return &itemCount
}
