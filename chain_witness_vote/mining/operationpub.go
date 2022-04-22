package mining

import (
	"mandela/config"
	"mandela/core/keystore/pubstore"
	"mandela/core/utils/crypto"
	"errors"
	//"mandela/core/utils"
	//"encoding/binary"
)

/*
	创建一个转款交易（公用）
*/
func CreateTxPayPub(pwd, seed string, items []*TxItem, address *crypto.AddressCoin, amount, gas, frozenHeight uint64, comment string) (*Tx_Pay, error) {
	keystore, err := pubstore.GetPubStore(pwd, seed)
	if err != nil {
		return nil, err
	}
	// start := time.Now()
	if len(items) == 0 {
		return nil, errors.New("items empty")
	}
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	//查找余额
	vins := make([]*Vin, 0)
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
		vin := &Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
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
	vouts := make([]*Vout, 0)
	vout := &Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
	}
	vouts = append(vouts, vout)
	//找零
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := &Vout{
			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr,       //  keystore.GetAddr()[0], //找零地址
		}
		vouts = append(vouts, vout)
	}

	// engine.Log.Info("构建输入输出 耗时 %s", time.Now().Sub(start))

	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay,                      //交易类型
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //交易输出
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			// CreateTime: time.Now().Unix(),         //创建时间
			Payload: []byte(comment),
		}
		pay = &Tx_Pay{
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

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
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

/*
	创建多个转款交易
*/
func CreateTxsPayPub(pwd, seed string, items []*TxItem, address []PayNumber, gas uint64, comment string) (*Tx_Pay, error) {
	keystore, err := pubstore.GetPubStore(pwd, seed)
	if err != nil {
		return nil, err
	}
	// start := time.Now()
	if len(items) == 0 {
		return nil, errors.New("items empty")
	}
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]*Vin, 0)

	// total, items := chain.balance.BuildPayVin(srcAddr, amount+gas)
	var total uint64
	for _, item := range items {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
		vin := Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)
		total = total + item.Value
	}
	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	//构建交易输出
	vouts := make([]*Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:        one.Amount,       //输出金额 = 实际金额 * 100000000
			Address:      one.Address,      //钱包地址
			FrozenHeight: one.FrozenHeight, //
		}
		vouts = append(vouts, &vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr,       //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay,                      //交易类型
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //交易输出
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			// CreateTime: time.Now().Unix(),         //创建时间
			Payload: []byte(comment),
		}
		pay = &Tx_Pay{
			TxBase: base,
		}
		// pay.CleanZeroVout()

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			pay.Vin[i].Sign = *sign

		}
		pay.BuildHash()
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	// engine.Log.Info("冻结交易 %+v", items)
	// engine.Log.Info("创建多人转账交易 %d %+v", block.Height, pay)
	//chain.balance.Frozen(items, pay)
	return pay, nil
}
