package tx_name_out

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/name"
	"mandela/chain_witness_vote/mining/tx_name_in"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"bytes"
	"sync"

	"github.com/gogo/protobuf/proto"
)

func init() {
	ac := new(AccountController)
	mining.RegisterTransaction(config.Wallet_tx_type_account_cancel, ac)
}

type AccountController struct {
}

func (this *AccountController) Factory() interface{} {
	return new(Tx_account)
}

func (this *AccountController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxNameOut)
	err := proto.Unmarshal(*bs, txProto)
	if err != nil {
		return nil, err
	}
	vins := make([]*mining.Vin, 0)
	for _, one := range txProto.TxBase.Vin {
		vins = append(vins, &mining.Vin{
			Txid: one.Txid,
			Vout: one.Vout,
			Puk:  one.Puk,
			Sign: one.Sign,
		})
	}
	vouts := make([]*mining.Vout, 0)
	for _, one := range txProto.TxBase.Vout {
		vouts = append(vouts, &mining.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	txBase := mining.TxBase{}
	txBase.Hash = txProto.TxBase.Hash
	txBase.Type = txProto.TxBase.Type
	txBase.Vin_total = txProto.TxBase.VinTotal
	txBase.Vin = vins
	txBase.Vout_total = txProto.TxBase.VoutTotal
	txBase.Vout = vouts
	txBase.Gas = txProto.TxBase.Gas
	txBase.LockHeight = txProto.TxBase.LockHeight
	txBase.Payload = txProto.TxBase.Payload
	txBase.BlockHash = txProto.TxBase.BlockHash
	tx := &Tx_account{
		TxBase: txBase,
	}
	return tx, nil
}

/*
	统计余额
	将已经注册的域名保存到数据库
	将自己注册的域名保存到内存
*/
func (this *AccountController) CountBalance(balance *mining.TxItemManager, deposit *sync.Map, bhvo *mining.BlockHeadVO) {

	for _, txItr := range bhvo.Txs {

		if txItr.Class() != config.Wallet_tx_type_account_cancel {
			continue
		}

		var depositIn *sync.Map
		v, ok := deposit.Load(config.Wallet_tx_type_account)
		if ok {
			depositIn = v.(*sync.Map)
		} else {
			depositIn = new(sync.Map)
			deposit.Store(config.Wallet_tx_type_account, depositIn)
		}

		// txItr := bhvo.Txs[txIndex]

		for _, vin := range *txItr.GetVin() {
			preTxItr, err := mining.LoadTxBase(vin.Txid)
			// preTxItr, err := mining.FindTxBase(vin.Txid)
			if err != nil {
				//TODO 不能解析上一个交易，程序出错退出
				continue
			}
			if preTxItr.Class() != config.Wallet_tx_type_account {
				continue
			}
			if vin.Vout != 0 {
				continue
			}

			namein := preTxItr.(*tx_name_in.Tx_account)

			//删除域名对应的交易id
			// db.Save(append([]byte(config.Name), txAcc.Account...), txItr.GetHash())
			db.LevelTempDB.Remove(append([]byte(config.Name), namein.Account...))

			//判断是自己相关的地址
			// if !keystore.FindAddress(namein.Vout[0].Address) {
			// 	continue
			// }
			_, ok := keystore.FindAddress(namein.Vout[0].Address)
			if !ok {
				continue
			}

			//从内存中删除域名押金交易
			depositIn.Delete(string(namein.Account))
			// depositIn.Store(string(txAcc.Account), &txItem)
			//保存自己相关的域名到内存
			name.DelName(namein.Account)

		}
		//生成新的UTXO收益，保存到列表中
		// for voutIndex, vout := range *txItr.GetVout() {
		// 	//找出需要统计余额的地址
		// 	// validate := keystore.ValidateByAddress(vout.Address.B58String())

		// 	txItem := mining.TxItem{
		// 		Addr: &vout.Address, //
		// 		// AddrStr:  vout.GetAddrStr(), //
		// 		Value:     vout.Value,        //余额
		// 		Txid:      *txItr.GetHash(),  //交易id
		// 		VoutIndex: uint64(voutIndex), //交易输出index，从0开始
		// 	}
		// 	// //下标为0的收益为押金，不要急于使用，使用了代表域名已经被注销
		// 	// if voutIndex == 0 {
		// 	// 	txAcc := txItr.(*Tx_account)

		// 	// 	//删除域名对应的交易id
		// 	// 	// db.Save(append([]byte(config.Name), txAcc.Account...), txItr.GetHash())
		// 	// 	db.Remove(append([]byte(config.Name), txAcc.Account...))

		// 	// 	//判断是自己相关的地址
		// 	// 	if !keystore.FindAddress(vout.Address) {

		// 	// 		continue
		// 	// 	}
		// 	// 	//从内存中删除域名押金交易
		// 	// 	depositIn.Delete(string(txAcc.Account))
		// 	// 	//保存自己相关的域名到内存
		// 	// 	name.DelName(txAcc.Account)
		// 	// }

		// 	// if !keystore.FindAddress(vout.Address) {
		// 	// 	continue
		// 	// }
		// 	ok := vout.CheckIsSelf()
		// 	// _, ok := keystore.FindAddress(vout.Address)
		// 	if !ok {
		// 		continue
		// 	}
		// 	// if !validate.IsVerify || !validate.IsMine {
		// 	// 	continue
		// 	// }
		// 	// fmt.Println("域名注销  333333333")
		// 	//计入余额列表
		// 	balance.AddTxItem(txItem)

		// 	// v, ok := balance.Load(vout.Address.B58String())
		// 	// var ba *mining.Balance
		// 	// if ok {
		// 	// 	ba = v.(*mining.Balance)
		// 	// } else {
		// 	// 	ba = new(mining.Balance)
		// 	// 	ba.Txs = new(sync.Map)
		// 	// }
		// 	// ba.Txs.Store(txItr.GetHashStr()+"_"+strconv.Itoa(voutIndex), &txItem)
		// 	// balance.Store(vout.Address.B58String(), ba)

		// }
	}
}

func (this *AccountController) CheckMultiplePayments(txItr mining.TxItr) error {
	return nil
}

func (this *AccountController) SyncCount() {

}

func (this *AccountController) RollbackBalance() {
	// return new(Tx_account)
}

/*
	注册域名交易，域名续费交易，修改域名的网络地址交易
	@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *AccountController) BuildTx(balance *mining.TxItemManager, deposit *sync.Map, srcAddr,
	addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (mining.TxItr, error) {
	var depositIn *sync.Map
	v, ok := deposit.Load(config.Wallet_tx_type_account)
	if ok {
		depositIn = v.(*sync.Map)
	} else {
		depositIn = new(sync.Map)
		deposit.Store(config.Wallet_tx_type_account, depositIn)
	}

	if len(params) < 1 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}
	nameStr := params[0].(string)

	var commentBs []byte
	if comment != "" {
		commentBs = []byte(comment)
	}

	//检查域名是否属于自己
	nameInTxid := name.FindName(nameStr)
	if nameInTxid == nil {
		return nil, config.ERROR_name_not_exist
	}

	// depositIn.Range(func(k, v interface{}) bool {
	// 	fmt.Println("遍历域名押金", k.(string))
	// 	return true
	// })

	itemItr, ok := depositIn.Load(nameStr)
	if !ok {
		//未找到对应押金
		return nil, config.ERROR_deposit_not_exist // errors.New("未找到对应押金")
	}

	item := itemItr.(*mining.TxItem)
	txItr, err := mining.LoadTxBase(item.Txid)
	// txItr, err := mining.FindTxBase(item.Txid)
	// bs, err := db.Find(item.Txid)
	// if err != nil {
	// 	return nil, err
	// }
	// txItr, err := mining.ParseTxBase(mining.ParseTxClass(item.Txid), bs)
	if err != nil {
		return nil, err
	}
	vout := (*txItr.GetVout())[item.VoutIndex]

	pukBs, ok := keystore.GetPukByAddr(vout.Address)
	if !ok {
		return nil, err
	}

	vin := mining.Vin{
		Txid: item.Txid,      //UTXO 前一个交易的id
		Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk:  pukBs,          //公钥
		//			Sign: *sign,         //签名
	}
	vins := make([]*mining.Vin, 0)
	total := uint64(0)
	vins = append(vins, &vin)
	total = item.Value

	var items []*mining.TxItem
	//资金不够
	chain := mining.GetLongChain()
	if total < amount+gas {
		//余额不够给手续费，需要从其他账户余额作为输入给手续费
		totalTemp := uint64(0)
		total, items = chain.GetBalance().BuildPayVin(nil, amount+gas-total)
		if totalTemp < amount+gas {
			//资金不够
			return nil, config.ERROR_not_enough // errors.New("余额不足")
		}

		if len(items) > config.Mining_pay_vin_max {
			return nil, config.ERROR_pay_vin_too_much
		}

		for _, item := range items {
			puk, ok := keystore.GetPukByAddr(*item.Addr)
			if !ok {
				return nil, config.ERROR_public_key_not_exist
			}
			// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
			vin := mining.Vin{
				Txid: item.Txid,      //UTXO 前一个交易的id
				Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
				Puk:  puk,            //公钥
				//					Sign: *sign,           //签名
			}
			vins = append(vins, &vin)
		}

		// for _, one := range keystore.GetAddr() {
		// 	bas, _ := chain.GetBalance().FindBalance(&one)
		// 	puk, ok := keystore.GetPukByAddr(one)
		// 	if !ok {
		// 		//未找到地址对应的公钥
		// 		return nil, config.ERROR_public_key_not_exist // errors.New("未找到地址对应的公钥")
		// 	}

		// 	for _, two := range bas {
		// 		two.Txs.Range(func(k, v interface{}) bool {
		// 			item := v.(*mining.TxItem)

		// 			vin := mining.Vin{
		// 				Txid: item.Txid,     //UTXO 前一个交易的id
		// 				Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		// 				Puk:  puk,           //公钥
		// 				//						Sign: *sign,           //签名
		// 			}
		// 			vins = append(vins, vin)

		// 			total = total + item.Value

		// 			// fmt.Println("---", total >= amount+gas, total, amount+gas)

		// 			if total >= amount+gas {
		// 				// fmt.Println("凑够 退出")
		// 				return false
		// 			}

		// 			return true
		// 		})

		// 		if total >= amount+gas {
		// 			// fmt.Println("凑够 完成")
		// 			break
		// 		}
		// 	}
		// }
	}
	//余额不够给手续费
	if total < (amount + gas) {
		//余额不够
		// fmt.Println("余额不够")
		return nil, config.ERROR_not_enough
	}

	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin
	if addr == nil {
		//押金退还地址为空
		dstAddr = keystore.GetCoinbase().Addr
	} else {
		// fmt.Println("有押金退还地址", addr.B58String())
		dstAddr = *addr
	}

	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout = &mining.Vout{
		Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
		Address: dstAddr,              //钱包地址
	}
	vouts = append(vouts, vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	// if total > amount+gas {
	// 	vout := mining.Vout{
	// 		Value:   total - amount - gas,   //输出金额 = 实际金额 * 100000000
	// 		Address: keystore.GetCoinbase(), //钱包地址
	// 	}
	// 	vouts = append(vouts, vout)
	// }

	// var class uint64
	// if isReg {
	// 	// fmt.Println("类型为 注册账号")
	// 	class = config.Wallet_tx_type_account
	// } else {
	// 	// fmt.Println("类型为 注销账号")
	// 	class = config.Wallet_tx_type_account_cancel
	// }
	//

	_, block := chain.GetLastBlock()

	var txin *Tx_account
	for i := uint64(0); i < 10000; i++ {
		base := mining.TxBase{
			Type:       config.Wallet_tx_type_account_cancel,           //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentBs,                                      //
			// CreateTime: time.Now().Unix(),                    //创建时间
		}
		txin = &Tx_account{
			TxBase: base,
			// Account: []byte(nameStr), //账户名称
			// NetIds:           netidsMHash,                   //网络地址列表
			// NetIdsMerkleHash: utils.BuildMerkleRoot(netids), //
		}

		// txin.MergeVout()
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			for _, key := range keystore.GetAddr() {
				puk, ok := keystore.GetPukByAddr(key.Addr)
				if !ok {
					//未找到地址对应的公钥
					return nil, config.ERROR_public_key_not_exist // errors.New("未找到地址对应的公钥")
				}

				if bytes.Equal(puk, one.Puk) {

					_, prk, _, err := keystore.GetKeyByAddr(key.Addr, pwd)

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
	chain.GetBalance().Frozen(items, txin)
	return txin, nil
}

// func (this *AccountController) Check(txItr mining.TxItr) error {
// 	txAcc := txItr.(*Tx_account)
// 	return txAcc.Check()
// }
