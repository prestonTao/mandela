/*
	矿工费交易
*/
package mining

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	// "mandela/core/engine"
	// "time"
)

/*
	矿工奖励交易
	没有输入，只有输出
*/
type Tx_reward struct {
	TxBase
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_reward) GetVOJSON() interface{} {
	return this.TxBase.ConversionVO()
}

/*
	格式化成json字符串
*/
// func (this *Tx_reward) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	获取签名
*/
func (this *Tx_reward) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {

	signDst := this.GetSignSerialize(nil, vinIndex)

	// fmt.Println("签名前的字节", len(*bs), hex.EncodeToString(*bs), "\n")
	sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return &sign
}

/*
	检查交易是否合法
*/
func (this *Tx_reward) Check() error {
	// start := time.Now()
	// engine.Log.Info("开始验证交易合法性 Tx_reward")
	//检查输入输出是否对等，还有手续费

	if this.Vin == nil || len(this.Vin) != 1 {
		return config.ERROR_tx_fail
	}

	one := this.Vin[0]
	signDst := this.GetSignSerialize(nil, uint64(0))
	// engine.Log.Info("开始验证交易合法性 Tx_reward 2222222222222222 %s", time.Now().Sub(start))
	puk := ed25519.PublicKey(one.Puk)
	// engine.Log.Info("开始验证交易合法性 Tx_reward 3333333333333333 %s", time.Now().Sub(start))
	if config.Wallet_print_serialize_hex {
		engine.Log.Info("sign serialize:%s", hex.EncodeToString(*signDst))
	}
	if !ed25519.Verify(puk, *signDst, one.Sign) {
		return config.ERROR_sign_fail
	}

	// engine.Log.Info("开始验证交易合法性 Tx_reward 4444444444444444 %s", time.Now().Sub(start))
	outTotal := uint64(0)
	for _, one := range this.Vout {
		outTotal = outTotal + one.Value
	}

	return nil
}

/*
	构建hash值得到交易id
*/
func (this *Tx_reward) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_mining)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	是否验证通过
*/
func (this *Tx_reward) CheckRepeatedTx(txs ...TxItr) bool {
	return true
}

/*
	统计交易余额
*/
func (this *Tx_reward) CountTxItems(height uint64) *TxItemCount {
	itemCount := TxItemCount{
		Additems: make([]*TxItem, 0),
		SubItems: make([]*TxSubItems, 0),
	}
	//将之前的UTXO标记为已经使用，余额中减去。
	// for _, vin := range this.Vin {
	// 	// engine.Log.Info("查看vin中的状态 %d", vin.PukIsSelf)
	// ok := vin.CheckIsSelf()
	// if !ok {
	// 	continue
	// }
	// 	// engine.Log.Info("统单易1耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	// 	//查找这个地址的余额列表，没有则创建一个
	// 	itemCount.SubItems = append(itemCount.SubItems, &TxSubItems{
	// 		Txid:      vin.Txid, //utils.Bytes2string(vin.Txid), //  vin.GetTxidStr(),
	// 		VoutIndex: vin.Vout,
	// 		Addr:      *vin.GetPukToAddr(), // utils.Bytes2string(*vin.GetPukToAddr()), // vin.GetPukToAddrStr(),
	// 	})
	// }

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

func (this *Tx_reward) CountTxHistory(height uint64) {
	// engine.Log.Info("CountTxHistory reward")
	//转出历史记录
	// hiOut := HistoryItem{
	// 	IsIn:    false,                          //资金转入转出方向，true=转入;false=转出;
	// 	Type:    this.Class(),                   //交易类型
	// 	InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
	// 	OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
	// 	// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
	// 	// Value:  amount,          //交易金额
	// 	Txid:   *this.GetHash(), //交易id
	// 	Height: height,          //
	// 	// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
	// }
	//转入历史记录
	hiIn := HistoryItem{
		IsIn: true,         //资金转入转出方向，true=转入;false=转出;
		Type: this.Class(), //交易类型
		// InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
		OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
		// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
		// Value:  amount,          //交易金额
		Txid:   *this.GetHash(), //交易id
		Height: height,          //
		// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
	}
	//
	// for _, vin := range this.Vin {
	// 	addrInfo, isSelf := keystore.FindPuk(vin.Puk)
	// 	hiIn.InAddr = append(hiIn.InAddr, &addrInfo.Addr)
	// 	if !isSelf {
	// 		continue
	// 	}
	// 	hiOut.InAddr = append(hiOut.InAddr, &addrInfo.Addr)
	// }

	addrCoin := make(map[string]bool)
	//生成新的UTXO收益，保存到列表中
	for _, vout := range this.Vout {
		// hiOut.OutAddr = append(hiOut.OutAddr, &vout.Address)
		// hiOut.Value += vout.Value
		_, ok := keystore.FindAddress(vout.Address)
		if !ok {
			continue
		}
		hiIn.Value += vout.Value
		if _, ok := addrCoin[utils.Bytes2string(vout.Address)]; ok {
			continue
		} else {
			addrCoin[utils.Bytes2string(vout.Address)] = false
		}
		hiIn.OutAddr = append(hiIn.OutAddr, &vout.Address)
	}

	if len(hiIn.OutAddr) > 0 && hiIn.Value > 0 {
		balanceHistoryManager.Add(hiIn)
	}
}
