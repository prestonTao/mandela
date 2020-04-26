package mining

import (
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/binary"
)

/*
	转账交易
*/
type Tx_Pay struct {
	TxBase
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_Pay) GetVOJSON() interface{} {
	return this.TxBase.ConversionVO()
}

/*
	构建hash值得到交易id
*/
func (this *Tx_Pay) BuildHash() {
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_pay)
	// jsonBs, _ := this.Json()

	// fmt.Println("序列化输出 111", string(*jsonBs))
	// fmt.Println("序列化输出 222", len(*bs), hex.EncodeToString(*bs))
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	验证是否合法
*/
func (this *Tx_Pay) Check() error {
	// fmt.Println("开始验证交易合法性 Tx_Pay")
	if err := this.TxBase.CheckBase(); err != nil {
		return err
	}

	return nil
}

/*
	是否验证通过
*/
func (this *Tx_Pay) CheckRepeatedTx(txs ...TxItr) bool {
	//判断是否出现双花
	return this.MultipleExpenditures(txs...)
}

/*
	创建一个转款交易
*/
func CreateTxPay(address *crypto.AddressCoin, amount, gas uint64, pwd, comment string) (*Tx_Pay, error) {
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	//查找余额
	vins := make([]Vin, 0)
	// total := uint64(0)

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

	//构建交易输出
	vouts := make([]Vout, 0)
	vout := Vout{
		Value:   amount,   //输出金额 = 实际金额 * 100000000
		Address: *address, //钱包地址
	}
	vouts = append(vouts, vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas,  //输出金额 = 实际金额 * 100000000
			Address: keystore.GetAddr()[0], //钱包地址
		}
		vouts = append(vouts, vout)
	}
	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay, //交易类型
			Vin_total:  uint64(len(vins)),         //输入交易数量
			Vin:        vins,                      //交易输入
			Vout_total: uint64(len(vouts)),        //输出交易数量
			Vout:       vouts,                     //交易输出
			Gas:        gas,                       //交易手续费
			LockHeight: block.Height + 100 + i,    //锁定高度
			// CreateTime: time.Now().Unix(),         //创建时间
		}
		pay = &Tx_Pay{
			TxBase: base,
		}

		pay.MergeVout()

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			for _, key := range keystore.GetAddrAll() {

				puk, ok := keystore.GetPukByAddr(key)
				if !ok {
					return nil, config.ERROR_public_key_not_exist // errors.New("未找到公钥")
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key, pwd)
					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						return nil, err
					}
					sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
					//				sign := pay.GetVoutsSign(prk, uint64(i))
					pay.Vin[i].Sign = *sign
				}
			}
		}
		pay.BuildHash()
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	return pay, nil
}

/*
	创建多个转款交易
*/
func CreateTxsPay(address []PayNumber, gas uint64, pwd, comment string) (*Tx_Pay, error) {
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]Vin, 0)
	total := uint64(0)
	keys := keystore.GetAddrAll()
	for _, one := range keys {

		bas := chain.balance.FindBalance(&one)

		for _, two := range bas {
			two.Txs.Range(func(k, v interface{}) bool {
				item := v.(*TxItem)

				puk, ok := keystore.GetPukByAddr(one)
				if !ok {
					return false
				}
				// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

				vin := Vin{
					Txid: item.Txid,     //UTXO 前一个交易的id
					Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
					Puk:  puk,           //公钥
					//					Sign: *sign,           //签名
				}
				vins = append(vins, vin)

				total = total + item.Value
				if total >= amount+gas {
					return false
				}

				return true
			})
			if total >= amount+gas {
				break
			}
		}
	}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}

	//构建交易输出
	vouts := make([]Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:   one.Amount,  //输出金额 = 实际金额 * 100000000
			Address: one.Address, //钱包地址
		}
		vouts = append(vouts, vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas,  //输出金额 = 实际金额 * 100000000
			Address: keystore.GetAddr()[0], //钱包地址
		}
		vouts = append(vouts, vout)
	}

	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay, //交易类型
			Vin_total:  uint64(len(vins)),         //输入交易数量
			Vin:        vins,                      //交易输入
			Vout_total: uint64(len(vouts)),        //输出交易数量
			Vout:       vouts,                     //交易输出
			Gas:        gas,                       //交易手续费
			LockHeight: block.Height + 10 + i,     //锁定高度
			// CreateTime: time.Now().Unix(),         //创建时间
		}
		pay = &Tx_Pay{
			TxBase: base,
		}
		pay.MergeVout()

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			for _, key := range keystore.GetAddrAll() {

				puk, ok := keystore.GetPukByAddr(key)
				if !ok {
					//未找到公钥
					return nil, config.ERROR_public_key_not_exist // errors.New("未找到公钥")
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key, pwd)
					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						return nil, err
					}
					sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
					//				sign := pay.GetVoutsSign(prk, uint64(i))
					pay.Vin[i].Sign = *sign
				}
			}
		}
		pay.BuildHash()
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	return pay, nil
}

/*
	多人转账
*/
type PayNumber struct {
	Address crypto.AddressCoin //转账地址
	Amount  uint64             //转账金额
}

/*
	创建多个地址转账交易，带签名
*/
func CreateTxsPayByPayload(address []PayNumber, gas uint64, pwd string, cs CommunitySign) (*Tx_Pay, error) {
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]Vin, 0)
	total := uint64(0)
	keys := keystore.GetAddrAll()
	for _, one := range keys {

		bas := chain.balance.FindBalance(&one)

		for _, two := range bas {
			two.Txs.Range(func(k, v interface{}) bool {
				item := v.(*TxItem)

				puk, ok := keystore.GetPukByAddr(one)
				if !ok {
					return false
				}
				// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

				vin := Vin{
					Txid: item.Txid,     //UTXO 前一个交易的id
					Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
					Puk:  puk,           //公钥
					//					Sign: *sign,           //签名
				}
				vins = append(vins, vin)

				total = total + item.Value
				if total >= amount+gas {
					return false
				}

				return true
			})
			if total >= amount+gas {
				break
			}
		}
	}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}

	//构建交易输出
	vouts := make([]Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:   one.Amount,  //输出金额 = 实际金额 * 100000000
			Address: one.Address, //钱包地址
		}
		vouts = append(vouts, vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas,  //输出金额 = 实际金额 * 100000000
			Address: keystore.GetAddr()[0], //钱包地址
		}
		vouts = append(vouts, vout)
	}

	var pay *Tx_Pay
	for i := uint64(0); i < 10000; i++ {
		//没有输出
		base := TxBase{
			Type:       config.Wallet_tx_type_pay, //交易类型
			Vin_total:  uint64(len(vins)),         //输入交易数量
			Vin:        vins,                      //交易输入
			Vout_total: uint64(len(vouts)),        //输出交易数量
			Vout:       vouts,                     //交易输出
			Gas:        gas,                       //交易手续费
			LockHeight: block.Height + 10 + i,     //锁定高度
			// CreateTime: time.Now().Unix(),         //创建时间
		}
		base.MergeVout()
		var pay TxItr = &Tx_Pay{
			TxBase: base,
		}

		//给payload签名
		addr := crypto.BuildAddr(config.AddrPre, cs.Puk)
		_, prk, _, err := keystore.GetKeyByAddr(addr, pwd)
		if err != nil {
			return nil, err
		}
		pay = SignPayload(pay, cs.Puk, prk, cs.StartHeight, cs.EndHeight)

		//给输出签名，防篡改
		for i, one := range *pay.GetVin() {
			for j, key := range keystore.GetAddrAll() {

				puk, ok := keystore.GetPukByAddr(key)
				if !ok {
					//未找到公钥
					return nil, config.ERROR_not_enough // errors.New("未找到公钥")
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key, pwd)
					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						return nil, err
					}
					sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
					//				sign := pay.GetVoutsSign(prk, uint64(i))
					// pay.Vin[i].Sign = *sign
					pay.SetSign(uint64(j), *sign)
				}
			}
		}
		pay.BuildHash()
		if pay.CheckHashExist() {
			pay = nil
			continue
		} else {
			break
		}
	}
	return pay, nil
}
