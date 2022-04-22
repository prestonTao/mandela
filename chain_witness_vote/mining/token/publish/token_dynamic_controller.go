package publish

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/token"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"strconv"
	"sync"

	"github.com/gogo/protobuf/proto"
)

func init() {
	tpc := new(TokenPublishController)
	mining.RegisterTransaction(Wallet_tx_class, tpc)
}

type TokenPublishController struct {
}

func (this *TokenPublishController) Factory() interface{} {
	return new(TxToken)
}

func (this *TokenPublishController) ParseProto(bs *[]byte) (interface{}, error) {
	if bs == nil {
		return nil, nil
	}
	txProto := new(go_protos.TxTokenPublish)
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

	tokenVouts := make([]mining.Vout, 0)
	for _, one := range txProto.Token_Vout {
		tokenVouts = append(tokenVouts, mining.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	tx := &TxToken{
		TxBase:           txBase,
		Token_name:       txProto.TokenName,
		Token_symbol:     txProto.TokenSymbol,
		Token_supply:     txProto.TokenSupply,
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

		//交易地址
		// txidStr := txToken.GetHashStr()

		//添加一个token信息
		token.SaveTokenInfo(*txToken.GetHash(), txToken.Token_name, txToken.Token_symbol, txToken.Token_supply)

		// bs, err := txToken.Json()
		// if err != nil {
		// 	continue
		// }
		// db.Save(token.BuildKeyForPublishToken(txidStr), bs)

		for voutIndex, vout := range txToken.Token_Vout {

			//保存未使用的vout索引
			keyStr := utils.Bytes2string(*txToken.GetHash()) + "_" + utils.Bytes2string(*txToken.GetHash()) + "_" + strconv.Itoa(voutIndex)
			bs := utils.Uint64ToBytes(vout.FrozenHeight)
			db.LevelTempDB.Save([]byte(keyStr), &bs)

			//保存token交易输出的token发布交易txid
			keyStr = config.TokenPublishTxid + utils.Bytes2string(*txToken.GetHash()) + "_" + strconv.Itoa(voutIndex)
			// bs = []byte(txToken.GetHashStr())
			db.LevelTempDB.Save([]byte(keyStr), txToken.GetHash())

			_, ok := keystore.FindAddress(vout.Address)
			if !ok {
				continue
			}
			//生成新的UTXO收益，保存到列表中
			txItem := mining.TxItem{
				Addr:      &vout.Address,     //
				Value:     vout.Value,        //余额
				Txid:      *txItr.GetHash(),  //交易id
				VoutIndex: uint64(voutIndex), //交易输出index，从0开始
			}
			// token.AddToken(txidStr, txItem)

			txItemCounts = append(txItemCounts, token.TokenTxItemCount{
				PublishTxidStr: utils.Bytes2string(*txToken.GetHash()), // txToken.GetHashStr(), //token发布的交易id
				Name:           txToken.Token_name,                     //名称（全称）
				Symbol:         txToken.Token_symbol,                   //单位
				Supply:         txToken.Token_supply,                   //发行总量
				Additems:       &txItem,                                //交易
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
		// 	// txHashStr := utils.Bytes2string(*txItr.GetHash()) //txItr.GetHashStr()

		// 	// engine.Log.Info("统单易5耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
		// 	txItem := mining.TxItem{
		// 		Addr: &(*txItr.GetVout())[voutIndex].Address, //  &vout.Address,
		// 		// AddrStr:  vout.GetAddrStr(),                      //
		// 		Value: vout.Value,       //余额
		// 		Txid:  *txItr.GetHash(), //交易id
		// 		// TxidStr:  txHashStr,         //
		// 		VoutIndex: uint64(voutIndex), //交易输出index，从0开始
		// 		Height:    bhvo.BH.Height,    //
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
}

func (this *TokenPublishController) CheckMultiplePayments(txItr mining.TxItr) error {
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
func (this *TokenPublishController) BuildTx(balance *mining.TxItemManager, deposit *sync.Map, srcAddr,
	addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (mining.TxItr, error) {

	if len(params) < 4 {
		//参数不够
		return nil, config.ERROR_params_not_enough // errors.New("参数不够")
	}

	// @name    string    Token名称全称
	nameStr := params[0].(string)
	// @symbol    string    Token单位，符号
	symbolStr := params[1].(string)
	// @supply    uint64    发行总量
	supply := params[2].(uint64)
	// @owner    crypto.AddressCoin    所有者
	owner := params[3].(crypto.AddressCoin)

	var commentbs []byte
	if comment != "" {
		commentbs = []byte(comment)
	}

	//验证发行总量最少
	if supply < config.Witness_token_supply_min {
		return nil, config.ERROR_token_min_fail //
	}

	//所有人为空，则设置所有人为本钱包coinbase
	if owner == nil {
		owner = keystore.GetCoinbase().Addr
	}

	tokenVout := make([]mining.Vout, 0)
	voutOne := mining.Vout{
		Value:   supply, //输出金额 = 实际金额 * 100000000
		Address: owner,  //钱包地址
	}
	tokenVout = append(tokenVout, voutOne)

	//
	chain := mining.GetLongChain()

	//查找余额
	vins := make([]*mining.Vin, 0)

	total, items := chain.GetBalance().BuildPayVin(*srcAddr, amount+gas)
	if total <= 0 || total < amount+gas {
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
		vin := mining.Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
		}
		vins = append(vins, &vin)
	}

	//余额不够给手续费
	if total < (amount + gas) {
		//余额不够
		// _, e := model.Errcode(model.NotEnough)
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
	//找回零钱
	if total > amount+gas {
		vout := mining.Vout{
			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr,       // keystore.GetAddr()[0].Addr, //钱包地址
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
			Token_name:       nameStr,                //名称
			Token_symbol:     symbolStr,              //单位
			Token_supply:     supply,                 //发行总量
			Token_Vout_total: uint64(len(tokenVout)), //输出交易数量
			Token_Vout:       tokenVout,              //交易输出
		}

		//这个交易存在手续费为0的情况，所以不合并vout
		// txin.MergeVout()
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
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
	chain.GetBalance().Frozen(items, txin)
	return txin, nil
}

// func (this *TokenPublishController) Check(txItr mining.TxItr, lockHeight uint64) error {
// 	txAcc := txItr.(*TxToken)
// 	return txAcc.Check(lockHeight)
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
