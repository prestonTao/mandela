package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/token/payment"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore/pubstore"
	"mandela/core/utils/crypto"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	// "time"
)

const (
	wallet_db_path    = "D:/temp/peer/wallet/data"
	tx                = `{"hash":"0400000000000000f824c50fce25b94332f6ac2c96fe996ec61eabccc42a83d760867263ad78bc21","type":4,"vin_total":1,"vin":[{"txid":"0100000000000000eb95bf79cc5eb0345fdf14b38e0458c4412962d7c9c5d3dbd335eba5ee40441d","vout":0,"puk":"20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9","sign":"ea69cc78c60768fc3f5cd4c02d530dc43a8d87604abe2e46944f4d38da4a9060a0aafb2666deac014f7dd6f75cbd9cde11e30fd4ceb0727c0bb6cf7f3970d703"}],"vout_total":2,"vout":[{"value":1000000000000000,"address":"ZHCQ3hzd66mYRGFU8HXLD2QguNnUpVh9wVVA4","frozen_height":0},{"value":298999999990000000,"address":"ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4","frozen_height":0}],"gas":10000000,"lock_height":370901,"payload":"test","blockhash":"bb6298d4f0c9e8c0cdf406225f917305da0de020f8ff2ef2500ae8e1bd837b81"}`
	txIsVisualization = false
	seed              = "bExBxUCsLVSk5czubhZYm5mqrRE2tcCCZjWCQKPk9k2UoA5LJKCWbFsLKAsyiDNixD28Y9G2C6GLjTdFW2eDc9xwTd7pwdTWz3Xqm7Bt9z2UjtfdRyLt1dRxKtEvPa5pZmfjpEAKfAuyBnmB25HNQ5qcavHKRZYfozgLafSh8wHudDkgBdsfsbuy8dUDhHYdsJF8iTEsyyduWdEnApo6tyxyMrXUz7XHy3UvTLCQibkp8mr63ZTE7hM9GtCapkWhmEMmrEaJHyKRyeyeDNXJfKiyQHWf27P43ZTWCTdbN6Ja36Sp8XTZ5u4aL6Ux86MLXEUugERWt"
	pwd               = "123456789"
	lockHeight        = 1127226
)

// var txItr mining.TxItr

func init() {
	tpc := new(payment.TokenPublishController)
	tpc.ActiveVoutIndex = new(sync.Map)
	mining.RegisterTransaction(config.Wallet_tx_type_account, tpc)

	err := db.InitDB(config.DB_path)
	if err != nil {
		fmt.Println("init db error:", err.Error())
		panic(err)
	}

}

func main() {
	example2()
}

func example1() {
	txItr := parseTx()
	if txItr.Class() == config.Wallet_tx_type_pay {
		createSign(txItr)
	} else if txItr.Class() == config.Wallet_tx_type_token_payment {
		txItr := parseTxTokenPay(tx)
		reSign(txItr, seed, pwd)
	} else {
		panic("未识别的交易类型")
	}
}
func example2() {
	txItr := parseTx()
	if txItr.Class() == config.Wallet_tx_type_pay {
		// createSign(txItr)
		printVO(txItr)
	} else if txItr.Class() == config.Wallet_tx_type_token_payment {
		txItr := parseTxTokenPay(tx)
		// reSign(txItr, seed, pwd)
		printVO(txItr)
	} else {
		panic("未识别的交易类型")
	}

}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}

func parseTx() *mining.Tx_Pay {
	txbaseVO := new(mining.TxBaseVO)

	err := json.Unmarshal([]byte(tx), txbaseVO)
	panicError(err)

	txhash, err := hex.DecodeString(txbaseVO.Hash)
	panicError(err)

	vins := make([]*mining.Vin, 0)
	for _, one := range txbaseVO.Vin {
		txid, err := hex.DecodeString(one.Txid)
		panicError(err)
		puk, err := hex.DecodeString(one.Puk)
		panicError(err)
		vin := mining.Vin{
			Txid: txid,     //UTXO 前一个交易的id
			Vout: one.Vout, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
			Puk:  puk,      //公钥
		}
		vins = append(vins, &vin)
	}

	vouts := make([]*mining.Vout, 0)
	for _, one := range txbaseVO.Vout {

		vout := mining.Vout{
			Value:        one.Value,                                //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressFromB58String(one.Address), //钱包地址
			FrozenHeight: one.FrozenHeight,                         //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
		}
		vouts = append(vouts, &vout)
	}

	blockHash, err := hex.DecodeString(txbaseVO.BlockHash)
	panicError(err)

	txbase := mining.TxBase{
		Hash:       txhash,              //本交易hash，不参与区块hash，只用来保存
		Type:       txbaseVO.Type,       //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  txbaseVO.Vin_total,  //输入交易数量
		Vin:        vins,                //交易输入
		Vout_total: txbaseVO.Vout_total, //输出交易数量
		Vout:       vouts,               //交易输出
		Gas:        txbaseVO.Gas,        //交易手续费，此字段不参与交易hash
		LockHeight: txbaseVO.LockHeight, //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		// LockHeight: lockHeight,               //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		Payload:   []byte(txbaseVO.Payload), //备注信息
		BlockHash: blockHash,                //本交易属于的区块hash，不参与区块hash，只用来保存
	}
	txPay := mining.Tx_Pay{
		TxBase: txbase,
	}
	return &txPay
}

func createSign(txItr mining.TxItr) {

	txPay := txItr.(*mining.Tx_Pay)
	// fmt.Println("11111111111111")

	// fmt.Println("222222222222222222222")
	// txItr = parseTx()
	// fmt.Println("333333333333333333")
	// txBs := []byte(tx)
	// txItr, err = mining.ParseTxBase(0, &txBs)
	// if err != nil {
	// 	fmt.Println("parse txbase error:", err.Error())
	// 	return
	// }

	// keystore, err := pubstore.GetPubStore(pwd, seed)
	// if err != nil {
	// 	fmt.Println("初始化密钥失败:", err.Error())
	// 	return
	// }

	//构建txitem
	items := make([]*mining.TxItem, 0)

	for _, vin := range *txPay.GetVin() {

		bs, err := db.Find(vin.Txid)
		if err != nil {
			fmt.Println("查询交易错误:", err.Error())
			return
		}

		txItrOne, err := mining.ParseTxBase(mining.ParseTxClass(vin.Txid), bs)
		if err != nil {
			fmt.Println("解析交易错误:", err.Error())
			return
		}

		// txItrOne, err := mining.FindTxBase(vin.Txid, "")
		// if err != nil {
		// 	fmt.Println("查询交易错误:", err.Error())
		// 	return
		// }

		vouts := txItrOne.GetVout()
		voutOne := (*vouts)[vin.Vout]

		txItem := mining.TxItem{
			Addr:     &voutOne.Address, //收款地址
			Value:    voutOne.Value,    //余额
			Txid:     vin.Txid,         //交易id
			OutIndex: vin.Vout,         //交易输出index，从0开始
		}

		items = append(items, &txItem)

	}

	voutOne := (*txPay.GetVout())[0]

	// txBase := txItr.(*mining.Tx_Pay)

	txPay, err := CreateTxPayPub(txItr, pwd, seed, items, &voutOne.Address, voutOne.Value, 0, 0, "")
	if err != nil {
		fmt.Println("创建交易失败:", err.Error())
		return
	}
	fmt.Println("4444444444444444444444")
	txPayBs, err := json.Marshal(txPay.GetVOJSON())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(txPayBs))

	fmt.Println("打印可以上链的交易:")

	txBs, _ := txPay.Json()
	fmt.Println(strconv.Quote(string(*txBs)))

	fmt.Println("end!!!!!!")
}

/*
	创建一个转款交易（公用）
*/
func CreateTxPayPub(txItr mining.TxItr, pwd, seed string, items []*mining.TxItem, address *crypto.AddressCoin,
	amount, gas, frozenHeight uint64, comment string) (*mining.Tx_Pay, error) {
	txBase := txItr.(*mining.Tx_Pay)

	keystore, err := pubstore.GetPubStore(pwd, seed)
	if err != nil {
		return nil, err
	}
	// start := time.Now()
	if len(items) == 0 {
		return nil, errors.New("items empty")
	}
	// chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	//查找余额
	vins := make([]*mining.Vin, 0)
	// total := uint64(0)
	// fmt.Println("构建vins 222222222222222222")
	//total, items := chain.balance.BuildPayVin(srcAddress, amount+gas)

	// engine.Log.Info("查找余额耗时 %s", time.Now().Sub(start))

	// fmt.Printf("构建vins 333333333333333 %d\n%+v\n", total, items)

	var total uint64
	for _, item := range items {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
		vin := &mining.Vin{
			Txid: item.Txid,     //UTXO 前一个交易的id
			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,           //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, vin)
		total = total + item.Value
	}
	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	vout := &mining.Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
	}
	vouts = append(vouts, vout)
	//找零
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := &mining.Vout{
			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr,       //  keystore.GetAddr()[0], //找零地址
		}
		vouts = append(vouts, vout)
	}

	// engine.Log.Info("构建输入输出 耗时 %s", time.Now().Sub(start))

	var pay *mining.Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := mining.TxBase{
			Type:       config.Wallet_tx_type_pay, //交易类型
			Vin_total:  uint64(len(vins)),         //输入交易数量
			Vin:        vins,                      //交易输入
			Vout_total: uint64(len(vouts)),        //输出交易数量
			Vout:       vouts,                     //交易输出
			Gas:        txBase.Gas,                //交易手续费
			// LockHeight: txBase.LockHeight,         //锁定高度
			LockHeight: lockHeight, //锁定高度
			// CreateTime: time.Now().Unix(),         //创建时间
			Payload: txBase.Payload,
		}
		pay = &mining.Tx_Pay{
			TxBase: base,
		}

		// pay.MergeVout()
		// pay.CleanZeroVout()

		// startTime := time.Now()

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			pay.Vin[i].Sign = *sign

		}

		// engine.Log.Info("pub给输出签名 耗时 %d %s", i, time.Now().Sub(startTime))

		pay.BuildHash()
		// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	// engine.Log.Info("交易签名 耗时 %s", time.Now().Sub(start))
	//chain.balance.Frozen(items, pay)
	return pay, nil
}

// func

/*
	重新签名
*/
func reSign(txItr mining.TxItr, seed, pwd string) {
	keystore, err := pubstore.GetPubStore(pwd, seed)
	if err != nil {
		return
	}
	pay := txItr.(*payment.TxToken)
	pay.BlockHash = nil
	pay.Hash = nil
	// pay.HashStr = ""
	//给输出签名，防篡改
	for i, one := range *pay.GetVin() {
		_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
		panicError(err)
		// one.Puk

		engine.Log.Info("签名key %s", hex.EncodeToString(prk))
		sign := txItr.GetSign(&prk, one.Txid, one.Vout, uint64(i))
		txItr.SetSign(uint64(i), *sign)
		// pay.Vin[i].Sign = *sign
	}

	txItr.BuildHash()
	// engine.Log.Info("交易id是否有重复 %s", hex.EncodeToString(*pay.GetHash()))
	if txItr.CheckHashExist() {
		panic("生成的区块hash相同")
	}

	txPayBs, err := json.Marshal(txItr.GetVOJSON())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(txPayBs))

	fmt.Println("打印可以上链的交易:")

	txBs, _ := txItr.Json()
	fmt.Println(strconv.Quote(string(*txBs)))

	fmt.Println("end!!!!!!")

}

/*
	解析token交易
*/
func parseTxTokenPay(txStr string) *payment.TxToken {
	txbaseVO := new(payment.TxToken_VO)

	err := json.Unmarshal([]byte(txStr), txbaseVO)
	panicError(err)

	txhash, err := hex.DecodeString(txbaseVO.Hash)
	panicError(err)

	vins := make([]*mining.Vin, 0)
	for _, one := range txbaseVO.Vin {
		txid, err := hex.DecodeString(one.Txid)
		panicError(err)
		puk, err := hex.DecodeString(one.Puk)
		panicError(err)
		vin := mining.Vin{
			Txid: txid,     //UTXO 前一个交易的id
			Vout: one.Vout, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
			Puk:  puk,      //公钥
		}
		vins = append(vins, &vin)
	}

	vouts := make([]*mining.Vout, 0)
	for _, one := range txbaseVO.Vout {

		vout := mining.Vout{
			Value:        one.Value,                                //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressFromB58String(one.Address), //钱包地址
			FrozenHeight: one.FrozenHeight,                         //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
		}
		vouts = append(vouts, &vout)
	}

	blockHash, err := hex.DecodeString(txbaseVO.BlockHash)
	panicError(err)

	txbase := mining.TxBase{
		Hash:       txhash,              //本交易hash，不参与区块hash，只用来保存
		Type:       txbaseVO.Type,       //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  txbaseVO.Vin_total,  //输入交易数量
		Vin:        vins,                //交易输入
		Vout_total: txbaseVO.Vout_total, //输出交易数量
		Vout:       vouts,               //交易输出
		Gas:        txbaseVO.Gas,        //交易手续费，此字段不参与交易hash
		LockHeight: txbaseVO.LockHeight, //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		// LockHeight: lockHeight,               //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		Payload:   []byte(txbaseVO.Payload), //备注信息
		BlockHash: blockHash,                //本交易属于的区块hash，不参与区块hash，只用来保存
	}

	tokenVin := make([]*mining.Vin, 0)
	for _, one := range txbaseVO.Token_Vin {
		txid, err := hex.DecodeString(one.Txid)
		panicError(err)
		puk, err := hex.DecodeString(one.Puk)
		panicError(err)
		vin := mining.Vin{
			Txid: txid,     //UTXO 前一个交易的id
			Vout: one.Vout, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
			Puk:  puk,      //公钥
		}
		tokenVin = append(tokenVin, &vin)
	}

	tokenVouts := make([]*mining.Vout, 0)
	for _, one := range txbaseVO.Token_Vout {

		vout := mining.Vout{
			Value:        one.Value,                                //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressFromB58String(one.Address), //钱包地址
			FrozenHeight: one.FrozenHeight,                         //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
		}
		tokenVouts = append(tokenVouts, &vout)
	}

	txPay := payment.TxToken{
		TxBase:           txbase,
		Token_Vin_total:  txbaseVO.Token_Vin_total,  //输入交易数量
		Token_Vin:        tokenVin,                  //交易输入
		Token_Vout_total: txbaseVO.Token_Vout_total, //输出交易数量
		Token_Vout:       tokenVouts,                //交易输出
	}
	return &txPay
}

func printVO(txItr mining.TxItr) {
	txPayBs, err := json.Marshal(txItr.GetVOJSON())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(txPayBs))

	fmt.Println("打印可以上链的交易:")

	txBs, _ := txItr.Json()
	fmt.Println(strconv.Quote(string(*txBs)))

	fmt.Println("end!!!!!!")
}
