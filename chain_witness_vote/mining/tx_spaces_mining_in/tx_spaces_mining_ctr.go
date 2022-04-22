package tx_spaces_mining_in

import (
	"mandela/cache/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/tx_spaces_mining"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"sync"
)

func init() {
	ac := new(SpacesMiningInController)
	mining.RegisterTransaction(config.Wallet_tx_type_spaces_mining_in, ac)
}

type SpacesMiningInController struct {
}

func (this *SpacesMiningInController) Factory() interface{} {
	return new(Tx_SpacesMining)
}

/*
	统计余额
	将已经注册的域名保存到数据库
	将自己注册的域名保存到内存
*/
func (this *SpacesMiningInController) CountBalance(balance *mining.TxItemManager, deposit *sync.Map, bhvo *mining.BlockHeadVO) {
	for _, txItr := range bhvo.Txs {

		if txItr.Class() != config.Wallet_tx_type_spaces_mining_in {
			continue
		}

		var depositIn *sync.Map
		v, ok := deposit.Load(config.Wallet_tx_type_spaces_mining_in)
		if ok {
			depositIn = v.(*sync.Map)
		} else {
			depositIn = new(sync.Map)
			deposit.Store(config.Wallet_tx_type_spaces_mining_in, depositIn)
		}

		// txItr := bhvo.Txs[txIndex]
		//生成新的UTXO收益，保存到列表中
		for voutIndex, vout := range *txItr.GetVout() {
			//找出需要统计余额的地址
			// validate := keystore.ValidateByAddress(vout.Address.B58String())

			//下标为0的收益为押金，不要急于使用，使用了代表存储空间已经被注销
			if voutIndex == 0 {
				txItem := mining.TxItem{
					Addr: &vout.Address, //
					// AddrStr:  vout.GetAddrStr(), //
					Value:    vout.Value,        //余额
					Txid:     *txItr.GetHash(),  //交易id
					OutIndex: uint64(voutIndex), //交易输出index，从0开始
				}

				tsm := txItr.(*Tx_SpacesMining)

				spacesMining := tx_spaces_mining.SpacesMining{
					Deposit: vout.Value, //
				}

				spacesMiningBS, _ := json.Marshal(spacesMining)

				//保存域名对应的交易id
				//有过期的域名，先删除再保存
				db.Remove(append([]byte(config.DB_spaces_mining_addr), tsm.NetId...))
				db.Save(append([]byte(config.DB_spaces_mining_addr), tsm.NetId...), spacesMiningBS)

				//判断是自己相关的地址
				_, ok := keystore.FindAddress(vout.Address)
				if !ok {
					continue
				}
				//保存押金到内存
				depositIn.Store(utils.Bytes2string(tsm.NetId), &txItem)
				//保存自己相关的域名到内存

				// name.AddName(nameObj)
				continue
			}
			// ok := vout.CheckIsSelf()
			// if !ok {
			// 	continue
			// }

			//计入余额列表
			// balance.AddTxItem(txItem)

		}

		// depositIn.Range(func(k, v interface{}) bool {
		// 	fmt.Println("查看其他的押金 3333", k.(string))
		// 	return true
		// })
	}
}

func (this *SpacesMiningInController) CheckMultiplePayments(txItr mining.TxItr) error {
	return nil
}

func (this *SpacesMiningInController) SyncCount() {

}

func (this *SpacesMiningInController) RollbackBalance() {
	// return new(Tx_account)
}

/*
	注册域名交易，域名续费交易，修改域名的网络地址交易
	@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *SpacesMiningInController) BuildTx(balance *mining.TxItemManager, deposit *sync.Map, srcAddr,
	addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (mining.TxItr, error) {

	if amount < config.Mining_name_deposit_min {
		return nil, config.ERROR_name_deposit
	}

	var depositIn *sync.Map
	v, ok := deposit.Load(config.Wallet_tx_type_spaces_mining_in)
	if ok {
		depositIn = v.(*sync.Map)
	} else {
		depositIn = new(sync.Map)
		deposit.Store(config.Wallet_tx_type_spaces_mining_in, depositIn)
	}

	if len(params) < 1 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}
	netidsMHash := params[1].(nodeStore.AddressNet)

	var commentBs []byte
	if comment != "" {
		commentBs = []byte(comment)
	}

	isReg := true

	// netids := params[2].([][]byte)
	// netids := make([][]byte, 0)
	// for _, one := range netidsMHash {
	// 	netids = append(netids, one)
	// }

	// addrCoinBs := make([][]byte, 0)
	// for _, one := range addrCoins {
	// 	addrCoinBs = append(addrCoinBs, one)
	// }

	// isHave := false     //记录域名是否存在
	// isOvertime := false //若存在，记录是否过期
	// {
	// 	//判断域名是否已经注册
	// 	nameinfo := name.FindNameToNet(nameStr)
	// 	if nameinfo != nil {
	// 		isHave = true
	// 		isOvertime = nameinfo.CheckIsOvertime(mining.GetHighestBlock())
	// 	}

	// }
	//域名不存在，可以注册
	chain := mining.GetLongChain()

	//查找域名是否属于自己
	// txid := name.FindName(nameStr)

	// fmt.Println("注册域名的参数", isReg, isHave, isOvertime, txid)

	//查找余额
	vins := make([]*mining.Vin, 0)
	total := uint64(0)
	// if (isReg && isHave && isOvertime && txid == nil) || (isReg && !isHave) {
	// 	//注册

	// } else if isReg && isHave && txid != nil {

	// 	//续费
	// 	itemItr, ok := depositIn.Load(nameStr)
	// 	if !ok {
	// 		//未找到对应押金
	// 		return nil, config.ERROR_deposit_not_exist // errors.New("未找到对应押金")
	// 	}
	// 	item := itemItr.(*mining.TxItem)

	// 	txItr, err := mining.FindTxBase(item.Txid)
	// 	// bs, err := db.Find(item.Txid)
	// 	// if err != nil {
	// 	// 	return nil, err
	// 	// }
	// 	// txItr, err := mining.ParseTxBase(mining.ParseTxClass(item.Txid), bs)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	vout := (*txItr.GetVout())[item.OutIndex]

	// 	pukBs, ok := keystore.GetPukByAddr(vout.Address)
	// 	if !ok {
	// 		return nil, err
	// 	}
	// 	// prk, err := keystore.GetPriKeyByAddress(vout.Address.B58String(), pwd)
	// 	// if err != nil {
	// 	// 	return nil, err
	// 	// }

	// 	// pukBs, err := utils.MarshalPubkey(&prk.PublicKey)
	// 	// if err != nil {
	// 	// 	return nil, err
	// 	// }
	// 	vin := mining.Vin{
	// 		Txid: item.Txid,     //UTXO 前一个交易的id
	// 		Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
	// 		Puk:  pukBs,         //公钥
	// 		//			Sign: *sign,         //签名
	// 	}
	// 	vins = append(vins, &vin)

	// 	total = total + item.Value
	// 	// if total >= amount+gas {
	// 	// 	return nil, err
	// 	// }

	// } else if !isReg && txid != nil {
	// 	//注销

	// } else {
	// 	//参数错误
	// 	if isHave && !isOvertime {
	// 		return nil, config.ERROR_name_exist
	// 	}
	// 	//参数错误
	// 	return nil, config.ERROR_params_fail // errors.New("参数错误")
	// }
	var items []*mining.TxItem
	//资金不够
	if total < amount+gas {
		//余额不够给手续费，需要从其他账户余额作为输入给手续费
		// totalTemp := uint64(0)

		total, items = chain.GetBalance().BuildPayVin(nil, amount+gas-total)
		if total < amount+gas {
			engine.Log.Error("11111111余额不足 %d %d %d", total, amount, gas)
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
				Txid: item.Txid,     //UTXO 前一个交易的id
				Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
				Puk:  puk,           //公钥
				//					Sign: *sign,           //签名
			}
			vins = append(vins, &vin)
		}

	}
	//余额不够给手续费
	if total < (amount + gas) {
		engine.Log.Error("222222222余额不足 %d %d %d", total, amount, gas)
		//余额不够
		return nil, config.ERROR_not_enough
	}

	//押金冻结存放地址
	var dstAddr crypto.AddressCoin
	if addr == nil {
		dstAddr = keystore.GetCoinbase().Addr
	} else {
		dstAddr = *addr
	}
	//构建交易输出
	vouts := make([]*mining.Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := mining.Vout{
		Value:   amount,  //输出金额 = 实际金额 * 100000000
		Address: dstAddr, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		newAddrCoin, err := keystore.GetNewAddr(pwd)
		if err != nil {
			//密码错误
			return nil, config.ERROR_password_fail // errors.New("密码错误")
		}
		vout := mining.Vout{
			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
			Address: newAddrCoin,          //钱包地址
		}
		vouts = append(vouts, &vout)
	}
	var class uint64
	if isReg {
		// fmt.Println("类型为 注册账号")
		class = config.Wallet_tx_type_spaces_mining_in
	} else {
		// fmt.Println("类型为 注销账号")
		class = config.Wallet_tx_type_account_cancel
	}
	_, block := chain.GetLastBlock()
	var txin *Tx_SpacesMining
	for i := uint64(0); i < 10000; i++ {
		//
		base := mining.TxBase{
			Type:       class,                                          //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentBs,                                      //
			// CreateTime: time.Now().Unix(),      //创建时间
		}
		txin = &Tx_SpacesMining{
			TxBase: base,
			NetId:  netidsMHash, //网络地址列表
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

// /*
// 	检查域名是否过期
// 	@return    bool    域名是否存在
// 	@return    bool    域名是否过期
// */
// func CheckName(nameStr string) (bool, bool, error) {
// 	//判断域名是否已经注册
// 	txid, err := db.Find(append([]byte(config.Name), []byte(nameStr)...))
// 	if err != nil {
// 		if err == leveldb.ErrNotFound {
// 			return false, true, errors.New("域名账号不存在")
// 		}
// 		return false, true, err
// 	}

// 	bs, err := db.Find(*txid)
// 	if err != nil {
// 		return false, true, err
// 	}

// 	//域名已经存在，检查之前的域名是否过期，检查是否是续签
// 	existTx, err := mining.ParseTxBase(bs)
// 	if err != nil {
// 		return false, true, errors.New("checkname 解析域名注册交易出错")
// 	}
// 	//检查区块高度，查看是否过期
// 	blockBs, err := db.Find(*existTx.GetBlockHash())
// 	if err != nil {
// 		//TODO 可能是数据库损坏或数据被篡改出错
// 		return false, true, errors.New("查找域名注册交易对应的区块出错")
// 	}
// 	bh, err := mining.ParseBlockHead(blockBs)
// 	if err != nil {
// 		return false, true, errors.New("解析域名注册交易对应的区块出错")
// 	}
// 	//检查是否过期
// 	if mining.GetHighestBlock() > (bh.Height + name.NameOfValidity) {
// 		//域名已经存在
// 		return true, true, nil
// 	} else {
// 		return true, false, nil
// 	}

// }
