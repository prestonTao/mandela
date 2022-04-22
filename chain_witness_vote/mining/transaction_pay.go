package mining

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"encoding/binary"
	"encoding/hex"
	// "mandela/core/engine"
	// "time"
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
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}

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
	// return this.MultipleExpenditures(txs...)
	return true
}

/*
	统计交易余额
*/
func (this *Tx_Pay) CountTxItems(height uint64) *TxItemCount {
	itemCount := TxItemCount{
		Additems: make([]*TxItem, 0),
		SubItems: make([]*TxSubItems, 0),
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
		itemCount.SubItems = append(itemCount.SubItems, &TxSubItems{
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
		txItem := TxItem{
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
		TxCache.AddTxInTxItem(*this.GetHash(), this)

	}
	return &itemCount
}

/*
	创建一个转款交易
*/
func CreateTxPay(srcAddress, address, change *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string) (*Tx_Pay, error) {
	// engine.Log.Info("start CreateTxPay")
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}
	// start := time.Now()

	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	//查找余额
	vins := make([]*Vin, 0)
	// total := uint64(0)
	total, items := chain.balance.BuildPayVin(*srcAddress, amount+gas)

	// engine.Log.Info("查找余额耗时 %s", time.Now().Sub(start))

	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}
	if len(items) > config.Mining_pay_vin_max {
		return nil, config.ERROR_pay_vin_too_much
	}

	for _, item := range items {
		// engine.Log.Info("use item:%s_%d %d %d", hex.EncodeToString(item.Txid), item.VoutIndex, item.LockupHeight, item.Value)
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		vin := Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
		}
		vins = append(vins, &vin)
	}

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
	}
	vouts = append(vouts, &vout)
	//找零
	changeAddr := change
	if changeAddr == nil || len(*changeAddr) <= 0 {
		changeAddr = &(*items[0].Addr)
	}
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
			Address: *changeAddr,          //找零地址
		}
		vouts = append(vouts, &vout)
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
			Payload:    commentbs,                                      //
		}
		pay = &Tx_Pay{
			TxBase: base,
		}

		// pay.MergeVout()

		// startTime := time.Now()

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

		// engine.Log.Info("给输出签名 耗时 %d %s", i, time.Now().Sub(startTime))

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
	engine.Log.Info("start CreateTxPay %s", hex.EncodeToString(*pay.GetHash()))
	chain.balance.Frozen(items, pay)
	return pay, nil
}

/*
	合并多个items的转款交易
*/
func MergeTxPay(items []*TxItem, address *crypto.AddressCoin, gas, frozenHeight uint64, pwd, comment string) (*Tx_Pay, error) {
	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}
	// start := time.Now()

	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	//查找余额
	vins := make([]*Vin, 0)
	// total := uint64(0)
	var total = uint64(0)

	for _, item := range items {
		total += item.Value
	}
	// engine.Log.Info("查找余额耗时 %s", time.Now().Sub(start))
	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough
	}
	amount := total - gas
	if len(items) > config.Mining_pay_vin_max {
		return nil, config.ERROR_pay_vin_too_much
	}

	for _, item := range items {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		vin := Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
		}
		vins = append(vins, &vin)
		// engine.Log.Info("item:%+v", item)
	}

	//构建交易输出
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *address,     //钱包地址
		FrozenHeight: frozenHeight, //
	}
	vouts = append(vouts, &vout)
	//找零
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas, //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr,       //找零地址
		}
		vouts = append(vouts, &vout)
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
			Payload:    commentbs,                                      //
		}
		pay = &Tx_Pay{
			TxBase: base,
		}

		// pay.MergeVout()

		// startTime := time.Now()

		//给输出签名，防篡改
		for i, one := range pay.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))
			sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			// engine.Log.Info("sign :%v", sign)
			pay.Vin[i].Sign = *sign
		}

		// engine.Log.Info("给输出签名 耗时 %d %s", i, time.Now().Sub(startTime))

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

	chain.balance.Frozen(items, pay)
	return pay, nil
}

/*
	创建多个转款交易
*/
func CreateTxsPay(srcAddr crypto.AddressCoin, address []PayNumber, gas uint64, pwd, comment string) (*Tx_Pay, error) {

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]*Vin, 0)

	total, items := chain.balance.BuildPayVin(srcAddr, amount+gas)
	if total < amount+gas {
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
		vin := Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)
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
			Value:   total - amount - gas,       //输出金额 = 实际金额 * 100000000
			Address: keystore.GetAddr()[0].Addr, //钱包地址
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
			Payload: commentbs, //
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
	chain.balance.Frozen(items, pay)
	return pay, nil
}

/*
	多人转账
*/
type PayNumber struct {
	Address      crypto.AddressCoin //转账地址
	Amount       uint64             //转账金额
	FrozenHeight uint64             //冻结高度
}

/*
	创建多个地址转账交易，带签名
*/
func CreateTxsPayByPayload(address []PayNumber, gas uint64, pwd string, cs *CommunitySign) (*Tx_Pay, error) {
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}

	//查找余额
	vins := make([]*Vin, 0)

	total, items := chain.balance.BuildPayVin(nil, amount+gas)
	if total < amount+gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	if len(items) > config.Mining_pay_vin_max {
		return nil, config.ERROR_pay_vin_too_much
	}
	for _, item := range items {
		// engine.Log.Info("打印地址 %s", item.Addr.B58String())
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			// engine.Log.Error("异常：未找到公钥")
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
			Value:   total - amount - gas,       //输出金额 = 实际金额 * 100000000
			Address: keystore.GetAddr()[0].Addr, //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	vouts = CleanZeroVouts(&vouts)
	vouts = MergeVouts(&vouts)

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
		// pay.CleanZeroVout()
		var txItr TxItr = &Tx_Pay{
			TxBase: base,
		}

		//给payload签名
		if cs != nil {
			addr := crypto.BuildAddr(config.AddrPre, cs.Puk)
			_, prk, _, err := keystore.GetKeyByAddr(addr, pwd)
			if err != nil {
				return nil, err
			}
			txItr = SignPayload(txItr, cs.Puk, prk, cs.StartHeight, cs.EndHeight)
		}

		pay = txItr.(*Tx_Pay)

		//给输出签名，防篡改
		for i, one := range *pay.GetVin() {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
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
	chain.balance.Frozen(items, pay)
	return pay, nil
}
