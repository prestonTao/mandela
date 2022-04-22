package payment

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/token"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"encoding/hex"
	"strconv"
	"sync"

	"github.com/gogo/protobuf/proto"
)

func init() {
	tpc := new(TokenPublishController)
	tpc.ActiveVoutIndex = new(sync.Map)
	mining.RegisterTransaction(Wallet_tx_class, tpc)
}

type TokenPublishController struct {
	ActiveVoutIndex *sync.Map //活动的交易输出，key:string=[txid]_[vout index];value:=;

}

func (this *TokenPublishController) Factory() interface{} {
	return new(TxToken)
}

func (this *TokenPublishController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxTokenPay)
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

	tokenVins := make([]*mining.Vin, 0)
	for _, one := range txProto.Token_Vin {
		tokenVins = append(tokenVins, &mining.Vin{
			Txid: one.Txid,
			Vout: one.Vout,
			Puk:  one.Puk,
			Sign: one.Sign,
		})
	}

	tokenVouts := make([]*mining.Vout, 0)
	for _, one := range txProto.Token_Vout {
		tokenVouts = append(tokenVouts, &mining.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	tx := &TxToken{
		TxBase:           txBase,
		Token_Vin_total:  txProto.Token_VinTotal,
		Token_Vin:        tokenVins,
		Token_Vout_total: txProto.Token_VoutTotal,
		Token_Vout:       tokenVouts,
	}
	return tx, nil
}

/*
	统计余额
	将已经注册的域名保存到数据库
	将自己注册的域名保存到内存
*/
func (this *TokenPublishController) CountBalance(balance *mining.TxItemManager, deposit *sync.Map, bhvo *mining.BlockHeadVO) {

	// engine.Log.Info("开始统计token余额 1111111")

	txItemCounts := make([]token.TokenTxItemCount, 0)
	subItems := make([]token.TokenTxSubItems, 0)

	// itemCount := mining.TxItemCount{
	// 	Additems: make([]*mining.TxItem, 0),
	// 	SubItems: make([]*mining.TxSubItems, 0),
	// }

	for _, txItr := range bhvo.Txs {
		if txItr.Class() != Wallet_tx_class {
			continue
		}

		txToken := txItr.(*TxToken)

		var publishId *[]byte
		publishIdStr := "" //本次交易的token合约地址
		for vinIndex, vin := range txToken.Token_Vin {
			if vinIndex == 0 {
				var err error
				publishTxidKeyStr := config.TokenPublishTxid + utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
				publishId, err = db.LevelTempDB.Find([]byte(publishTxidKeyStr))
				if err != nil {
					engine.Log.Warn("查询token_id错误 %s", err.Error())
				}
				publishIdStr = string(*publishId)
			}
			_, ok := vin.ValidateAddr()
			if ok {
				// token.DelFrozenToken(vin.Txid, vin.Vout)
				subItems = append(subItems, token.TokenTxSubItems{
					PublishTxid: *publishId,          // txToken.GetPublishTxidStr(), //token发布的交易id
					Addr:        *vin.GetPukToAddr(), //   vin.GetPukToAddrStr(), //
					Txid:        vin.Txid,            //   vin.GetTxidStr(),      //
					VoutIndex:   vin.Vout,            //
				})
			}

			//删除一个已经使用了的交易输出
			keyStr := publishIdStr + "_" + utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
			db.LevelTempDB.Remove([]byte(keyStr))

			//删除活动的索引
			// engine.Log.Info("删除活动的交易 %s", keyStr)
			this.ActiveVoutIndex.Delete(keyStr)
		}

		for voutIndex, vout := range txToken.Token_Vout {
			//保存未使用的vout索引
			keyStr := publishIdStr + "_" + utils.Bytes2string(*txToken.GetHash()) + "_" + strconv.Itoa(voutIndex)
			bs := utils.Uint64ToBytes(vout.FrozenHeight)
			db.LevelTempDB.Save([]byte(keyStr), &bs)

			// engine.Log.Debug("未使用的token vout索引 %s", keyStr)

			//保存token交易输出的token发布交易txid
			keyStr = config.TokenPublishTxid + utils.Bytes2string(*txToken.GetHash()) + "_" + strconv.Itoa(voutIndex)
			// engine.Log.Info("保存类型为7的key %s", keyStr)
			db.LevelTempDB.Save([]byte(keyStr), publishId)

			// engine.Log.Info("开始统计token余额 2222222222")
			_, ok := keystore.FindAddress(vout.Address)
			if !ok {
				continue
			}
			// engine.Log.Info("开始统计token余额 33333333333")
			//生成新的UTXO收益，保存到列表中
			txItem := mining.TxItem{
				Addr:         &vout.Address,     //
				Value:        vout.Value,        //余额
				Txid:         *txItr.GetHash(),  //交易id
				VoutIndex:    uint64(voutIndex), //交易输出index，从0开始
				LockupHeight: vout.FrozenHeight, //锁仓高度
			}
			// token.AddToken(txToken.GetPublishTxidStr(), txItem)
			txItemCounts = append(txItemCounts, token.TokenTxItemCount{
				PublishTxidStr: publishIdStr, //token发布的交易id
				Additems:       &txItem,      //交易
			})
		}

		//生成新的UTXO收益，保存到列表中
		// for voutIndex, vout := range *txItr.GetVout() {
		// 	//找出需要统计余额的地址
		// 	//和自己无关的地址
		// 	ok := vout.CheckIsSelf()
		// 	if !ok {
		// 		continue
		// 	}
		// 	// txHashStr := txItr.GetHashStr()

		// 	// engine.Log.Info("统单易5耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
		// 	txItem := mining.TxItem{
		// 		Addr: &(*txItr.GetVout())[voutIndex].Address, //  &vout.Address,
		// 		// AddrStr: vout.GetAddrStr(),                      //
		// 		Value: vout.Value,       //余额
		// 		Txid:  *txItr.GetHash(), //交易id
		// 		// TxidStr:      txHashStr,                              //
		// 		VoutIndex:    uint64(voutIndex), //交易输出index，从0开始
		// 		Height:       bhvo.BH.Height,    //
		// 		LockupHeight: vout.FrozenHeight, //
		// 	}

		// 	//计入余额列表
		// 	itemCount.Additems = append(itemCount.Additems, &txItem)

		// 	//保存到缓存
		// 	// engine.Log.Info("开始统计交易余额 区块高度 %d 保存到缓存", bhvo.BH.Height)
		// 	mining.TxCache.AddTxInTxItem(*txItr.GetHash(), txItr)
		// }
	}
	//统计主链上的余额
	// balance.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)

	//统计token的余额
	token.CountToken(txItemCounts, subItems)

	//回滚冻结的token交易
	token.UnfrozenToken(bhvo.BH.Height - 1)
	// this.Unfrozen(bhvo.BH.Height - 1)
}

func (this *TokenPublishController) CheckMultiplePayments(txItr mining.TxItr) error {
	txToken := txItr.(*TxToken)
	activiVoutIndexs := make(map[string]interface{})
	for _, vin := range txToken.Token_Vin {
		//查询前token交易的txid
		keyStr := config.TokenPublishTxid + utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
		publishTxidBs, err := db.LevelTempDB.Find([]byte(keyStr))
		if err != nil {
			engine.Log.Warn("token 查找不到这个key %s %d", hex.EncodeToString(vin.Txid), vin.Vout)
			return err
		}

		// keyStr = string(*publishTxidBs) + "_" + vin.GetTxidStr() + "_" + strconv.Itoa(int(vin.Vout))
		keyStr = utils.Bytes2string(*publishTxidBs) + "_" + utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
		//先验证数据库里的交易是否有双花
		if !db.LevelTempDB.CheckHashExist([]byte(keyStr)) {
			//已经使用过了
			engine.Log.Warn("token 这个交易已经使用过了 %s", hex.EncodeToString(*txItr.GetHash()))
			// engine.Log.Warn("使用过的vout %s",hex.EncodeToString(vin.Txid)+"_"+strconv.Itoa(int(vin.Vout))
			return config.ERROR_tx_is_use
		}
		//再验证未上链的交易中是否有双花
		_, ok := this.ActiveVoutIndex.Load(keyStr)
		if ok {
			//已经存在
			engine.Log.Warn("未上链的交易有双花 %s %s %s %d", hex.EncodeToString(*txItr.GetHash()), hex.EncodeToString(*publishTxidBs),
				hex.EncodeToString(vin.Txid), vin.Vout)

			// this.ActiveVoutIndex.Range(func(k, v interface{}) bool {
			// 	engine.Log.Debug("打印缓存中的token交易 %s", k)
			// 	return true
			// })

			return config.ERROR_tx_is_use
		}
		activiVoutIndexs[keyStr] = nil
	}

	//保存活动的交易输出
	for k, v := range activiVoutIndexs {
		// engine.Log.Info("添加活动的交易 %s", k)
		this.ActiveVoutIndex.Store(k, v)
	}
	return nil
}

func (this *TokenPublishController) SyncCount() {

}

func (this *TokenPublishController) RollbackBalance() {
	// return new(Tx_account)
}

/*
	注册域名交易，域名续费交易，修改域名的网络地址交易
	@isReg    bool    是否注册。true=注册和续费或者修改域名地址；false=注销域名；
*/
func (this *TokenPublishController) BuildTx(balance *mining.TxItemManager, deposit *sync.Map,
	srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (mining.TxItr, error) {

	if len(params) < 1 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}

	//---------------------开始构建token的交易----------------------
	//发布token的交易id
	txid := params[0].(string)
	txidBs, err := hex.DecodeString(txid)
	if err != nil {
		return nil, config.ERROR_params_fail
	}

	// txidBs, err := hex.DecodeString(txid)
	// if err != nil {
	// 	return nil, config.ERROR_params_fail
	// }

	var commentbs []byte
	if comment != "" {
		commentbs = []byte(comment)
	}

	// srcAddrStr := ""
	// if srcAddr != nil {
	// 	srcAddrStr = srcAddr.B58String()
	// }
	tokenTotal, tokenTxItems := token.GetReadyPayToken(txidBs, *srcAddr, amount)
	if tokenTotal <= 0 || tokenTotal < amount {
		return nil, config.ERROR_token_not_enough
	}
	tokenVins := make([]*mining.Vin, 0)
	for _, item := range tokenTxItems {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		vin := mining.Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
		}
		tokenVins = append(tokenVins, &vin)
	}

	//构建交易输出
	tokenVouts := make([]*mining.Vout, 0)
	//转账token给目标地址
	tokenVout := mining.Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *addr,        //钱包地址
		FrozenHeight: frozenHeight, //
	}
	tokenVouts = append(tokenVouts, &tokenVout)
	//找零
	if tokenTotal > amount {
		tokenVout := mining.Vout{
			Value:   tokenTotal - amount,   //输出金额 = 实际金额 * 100000000
			Address: *tokenTxItems[0].Addr, // keystore.GetAddr()[0].Addr, //钱包地址
		}
		tokenVouts = append(tokenVouts, &tokenVout)
	}

	//---------------------开始构建主链上的交易----------------------
	//查找余额
	vins := make([]*mining.Vin, 0)
	chain := mining.GetLongChain() // forks.GetLongChain()
	total, items := chain.GetBalance().BuildPayVin(nil, gas)

	if total < gas {
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

	//构建交易输出
	vouts := make([]*mining.Vout, 0)

	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > gas {
		vout := mining.Vout{
			Value:   total - gas,    //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr, // keystore.GetAddr()[0].Addr, //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	_, block := chain.GetLastBlock()
	var txin *TxToken
	for i := uint64(0); i < 10000; i++ {
		//
		base := mining.TxBase{
			Type:       Wallet_tx_class,                                //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                      //
		}
		txin = &TxToken{
			TxBase:           base,
			Token_Vin_total:  uint64(len(tokenVins)),  //输入交易数量
			Token_Vin:        tokenVins,               //交易输入
			Token_Vout_total: uint64(len(tokenVouts)), //输出交易数量
			Token_Vout:       tokenVouts,              //交易输出
		}

		// txin.MergeVout()
		// txin.MergeTokenVout()

		//给token交易签名 给输出签名，防篡改
		// for i, one := range txin.Token_Vin {
		// 	_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	sign := txin.GetTokenSign(&prk, one.Txid, one.Vout, uint64(i))
		// 	txin.Token_Vin[i].Sign = *sign
		// }

		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			txin.Vin[i].Sign = *sign
		}

		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}

	//把txitem冻结起来
	token.FrozenToken(txidBs, tokenTxItems, txin)

	chain.GetBalance().Frozen(items, txin)
	return txin, nil
}
