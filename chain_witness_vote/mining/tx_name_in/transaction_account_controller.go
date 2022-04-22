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
	"mandela/protos/go_protos"
	"bytes"
	"sync"

	"github.com/gogo/protobuf/proto"
)

func init() {
	ac := new(AccountController)
	mining.RegisterTransaction(config.Wallet_tx_type_account, ac)
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
	txProto := new(go_protos.TxNameIn)
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

	addrNet := make([]nodeStore.AddressNet, 0)
	for i, _ := range txProto.NetIds {
		one := txProto.NetIds[i]
		addrNet = append(addrNet, one)
	}
	addrCoins := make([]crypto.AddressCoin, 0)
	for i, _ := range txProto.AddrCoins {
		one := txProto.AddrCoins[i]
		addrCoins = append(addrCoins, one)
	}
	tx := &Tx_account{
		TxBase:              txBase,
		Account:             txProto.Account,
		NetIds:              addrNet,
		NetIdsMerkleHash:    txProto.NetIdsMerkleHash,
		AddrCoins:           addrCoins,
		AddrCoinsMerkleHash: txProto.AddrCoinsMerkleHash,
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

		if txItr.Class() != config.Wallet_tx_type_account {
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
		//生成新的UTXO收益，保存到列表中
		for voutIndex, vout := range *txItr.GetVout() {
			//找出需要统计余额的地址
			// validate := keystore.ValidateByAddress(vout.Address.B58String())

			//下标为0的收益为押金，不要急于使用，使用了代表域名已经被注销
			if voutIndex == 0 {
				txItem := mining.TxItem{
					Addr: &vout.Address, //
					// AddrStr:  vout.GetAddrStr(), //
					Value:     vout.Value,        //余额
					Txid:      *txItr.GetHash(),  //交易id
					VoutIndex: uint64(voutIndex), //交易输出index，从0开始
				}

				txAcc := txItr.(*Tx_account)

				nameObj := name.Nameinfo{
					Name:      string(txAcc.Account), //域名
					Txid:      *txItr.GetHash(),      //交易id
					NetIds:    txAcc.NetIds,          //节点地址列表
					AddrCoins: txAcc.AddrCoins,       //钱包收款地址
					Height:    bhvo.BH.Height,        //注册区块高度
					Deposit:   vout.Value,            //
				}

				nameinfoBS, _ := nameObj.Proto()
				// nameinfoBS, _ := json.Marshal(nameObj)

				// txNameIn.Type

				//保存域名对应的交易id
				//有过期的域名，先删除再保存
				db.LevelTempDB.Remove(append([]byte(config.Name), txAcc.Account...))
				db.LevelTempDB.Save(append([]byte(config.Name), txAcc.Account...), &nameinfoBS)

				//判断是自己相关的地址
				_, ok := keystore.FindAddress(vout.Address)
				if !ok {
					//不是自己的地址，有可能是转移给其它地址了，转移后自己要删除
					depositIn.Delete(string(txAcc.Account))
					name.DelName(txAcc.Account)
					continue
				}
				//保存押金到内存
				depositIn.Store(string(txAcc.Account), &txItem)
				//保存自己相关的域名到内存

				name.AddName(nameObj)
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

	if amount < config.Mining_name_deposit_min {
		return nil, config.ERROR_name_deposit
	}

	var depositIn *sync.Map
	v, ok := deposit.Load(config.Wallet_tx_type_account)
	if ok {
		depositIn = v.(*sync.Map)
	} else {
		depositIn = new(sync.Map)
		deposit.Store(config.Wallet_tx_type_account, depositIn)
	}

	if len(params) < 3 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}
	nameStr := params[0].(string)
	netidsMHash := params[1].([]nodeStore.AddressNet)
	addrCoins := params[2].([]crypto.AddressCoin)

	var commentBs []byte
	if comment != "" {
		commentBs = []byte(comment)
	}

	isReg := true

	// netids := params[2].([][]byte)
	netids := make([][]byte, 0)
	for _, one := range netidsMHash {
		netids = append(netids, one)
	}

	addrCoinBs := make([][]byte, 0)
	for _, one := range addrCoins {
		addrCoinBs = append(addrCoinBs, one)
	}

	isHave := false     //记录域名是否存在
	isOvertime := false //若存在，记录是否过期
	{
		//判断域名是否已经注册
		nameinfo := name.FindNameToNet(nameStr)
		if nameinfo != nil {
			isHave = true
			isOvertime = nameinfo.CheckIsOvertime(mining.GetHighestBlock())
		}
	}
	//域名不存在，可以注册
	chain := mining.GetLongChain()

	//查找域名是否属于自己
	txid := name.FindName(nameStr)

	// fmt.Println("注册域名的参数", isReg, isHave, isOvertime, txid)

	//查找余额
	vins := make([]*mining.Vin, 0)
	var item *mining.TxItem
	total := uint64(0)
	if (isReg && isHave && isOvertime && txid == nil) || (isReg && !isHave) {
		//注册

	} else if isReg && isHave && txid != nil {

		//续费
		itemItr, ok := depositIn.Load(nameStr)
		if !ok {
			//未找到对应押金
			return nil, config.ERROR_deposit_not_exist // errors.New("未找到对应押金")
		}
		item = itemItr.(*mining.TxItem)
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
		// prk, err := keystore.GetPriKeyByAddress(vout.Address.B58String(), pwd)
		// if err != nil {
		// 	return nil, err
		// }

		// pukBs, err := utils.MarshalPubkey(&prk.PublicKey)
		// if err != nil {
		// 	return nil, err
		// }
		vin := mining.Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  pukBs,          //公钥
			//			Sign: *sign,         //签名
		}
		vins = append(vins, &vin)

		total = total + item.Value
		// if total >= amount+gas {
		// 	return nil, err
		// }

	} else if !isReg && txid != nil {
		//注销

	} else {
		//参数错误
		if isHave && !isOvertime {
			return nil, config.ERROR_name_exist
		}
		//参数错误
		return nil, config.ERROR_params_fail // errors.New("参数错误")
	}
	var items []*mining.TxItem
	//资金不够
	if total < amount+gas {
		//余额不够给手续费，需要从其他账户余额作为输入给手续费
		// totalTemp := uint64(0)

		total, items = chain.GetBalance().BuildPayVin(nil, amount+gas-total)
		if total < amount+gas {
			// engine.Log.Error("11111111余额不足 %d %d %d", total, amount, gas)
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

	}
	//余额不够给手续费
	if total < (amount + gas) {
		// engine.Log.Error("222222222余额不足 %d %d %d", total, amount, gas)
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
	//找零
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		// newAddrCoin, err := keystore.GetNewAddr(pwd)
		// if err != nil {
		// 	//密码错误
		// 	return nil, config.ERROR_password_fail // errors.New("密码错误")
		// }
		vout := mining.Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			// Address: newAddrCoin,          //钱包地址
		}
		if item != nil {
			vout.Address = *item.Addr
		} else {
			vout.Address = *items[0].Addr
		}
		vouts = append(vouts, &vout)
	}
	var class uint64
	if isReg {
		// fmt.Println("类型为 注册账号")
		class = config.Wallet_tx_type_account
	} else {
		// fmt.Println("类型为 注销账号")
		class = config.Wallet_tx_type_account_cancel
	}
	_, block := chain.GetLastBlock()
	var txin *Tx_account
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
		txin = &Tx_account{
			TxBase:              base,
			Account:             []byte(nameStr),                   //账户名称
			NetIds:              netidsMHash,                       //网络地址列表
			NetIdsMerkleHash:    utils.BuildMerkleRoot(netids),     //
			AddrCoins:           addrCoins,                         //网络地址列表
			AddrCoinsMerkleHash: utils.BuildMerkleRoot(addrCoinBs), //网络地址默克尔树hash
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
