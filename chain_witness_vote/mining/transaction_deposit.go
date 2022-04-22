/*
	参与挖矿投票
	参与挖矿的节点发出投票请求，
*/
package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	// "mandela/core/engine"
)

/*
	交押金，成为候选见证人
*/
type Tx_deposit_in struct {
	TxBase
	Puk  []byte `json:"puk"` //候选见证人公钥，这个公钥需要和交易index=0的输出地址保持一致
	Name []byte `json:"n"`   //TODO 给见证人绑定域名，现在只定义了字段，还未上链
}

/*
	交押金，成为候选见证人
*/
type Tx_deposit_in_VO struct {
	TxBaseVO
	Puk string `json:"puk"` //候选见证人公钥，这个公钥需要和交易index=0的输出地址保持一致
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_deposit_in) GetVOJSON() interface{} {
	return Tx_deposit_in_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		Puk:      hex.EncodeToString(this.Puk),
	}
}

/*
	构建hash值得到交易id
*/
func (this *Tx_deposit_in) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	buf := bytes.NewBuffer(*bs)
	buf.Write(this.Puk)
	*bs = buf.Bytes()

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_deposit_in)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	获取签名
*/
func (this *Tx_deposit_in) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {

	// txItr, err := FindTxBase(txid)
	txItr, err := LoadTxBase(txid)
	if err != nil {
		return nil
	}

	blockhash, err := db.GetTxToBlockHash(&txid)
	if err != nil || blockhash == nil {
		return nil
	}

	// if txItr.GetBlockHash() == nil {
	// 	txItr = GetRemoteTxAndSave(txid)
	// 	if txItr.GetBlockHash() == nil {
	// 		return nil
	// 	}
	// }

	voutBs := txItr.GetVoutSignSerialize(voutIndex)
	signBs := make([]byte, 0, len(*blockhash)+len(*txItr.GetHash())+len(*voutBs))
	signBs = append(signBs, *blockhash...)
	signBs = append(signBs, *txItr.GetHash()...)
	signBs = append(signBs, *voutBs...)

	// buf := bytes.NewBuffer(nil)
	// //上一个交易 所属的区块hash
	// buf.Write(*blockhash)
	// // buf.Write(*txItr.GetBlockHash())
	// //上一个交易的hash
	// buf.Write(*txItr.GetHash())
	// //上一个交易的指定输出序列化
	// buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	// //本交易类型输入输出数量等信息和所有输出
	// signBs := buf.Bytes()
	signDst := this.GetSignSerialize(&signBs, vinIndex)
	//本交易特有信息
	*signDst = append(*signDst, this.Puk...)

	// fmt.Println("签名前的字节", len(*signDst), hex.EncodeToString(*signDst), "\n")
	sign := keystore.Sign(*key, *signDst)

	return &sign

}

/*
	验证是否合法
*/
func (this *Tx_deposit_in) Check() error {
	// fmt.Println("开始验证交易合法性 Tx_deposit_in")

	//判断vin是否太多
	// if len(this.Vin) > config.Mining_pay_vin_max {
	// 	return config.ERROR_pay_vin_too_much
	// }

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费
	inTotal := uint64(0)
	for i, one := range this.Vin {

		txItr, err := LoadTxBase(one.Txid)
		// txItr, err := FindTxBase(one.Txid)

		// txbs, err := db.Find(one.Txid)
		// if err != nil {
		// 	return config.ERROR_tx_not_exist
		// }
		// txItr, err := ParseTxBase(ParseTxClass(one.Txid), txbs)
		if err != nil {
			return config.ERROR_tx_format_fail
		}

		blockhash, err := db.GetTxToBlockHash(&one.Txid)
		if err != nil || blockhash == nil {
			return config.ERROR_tx_format_fail
		}
		vout := (*txItr.GetVout())[one.Vout]
		//如果这个交易已经被使用，则验证不通过，否则会出现双花问题。
		// if vout.Txid != nil {
		// 	return config.ERROR_tx_is_use
		// }
		inTotal = inTotal + vout.Value

		//验证公钥是否和地址对应
		addr := crypto.BuildAddr(config.AddrPre, one.Puk)
		if !bytes.Equal(addr, (*txItr.GetVout())[one.Vout].Address) {
			return config.ERROR_public_and_addr_notMatch
		}

		voutBs := txItr.GetVoutSignSerialize(one.Vout)
		signBs := make([]byte, 0, len(*blockhash)+len(*txItr.GetHash())+len(*voutBs))
		signBs = append(signBs, *blockhash...)
		signBs = append(signBs, *txItr.GetHash()...)
		signBs = append(signBs, *voutBs...)

		// //验证签名
		// buf := bytes.NewBuffer(nil)
		// //上一个交易 所属的区块hash
		// buf.Write(*blockhash)
		// // buf.Write(*txItr.GetBlockHash())
		// //上一个交易的hash
		// buf.Write(*txItr.GetHash())
		// //上一个交易的指定输出序列化
		// buf.Write(*txItr.GetVoutSignSerialize(one.Vout))
		// //本交易类型输入输出数量等信息和所有输出
		// signBs := buf.Bytes()

		signDst := this.GetSignSerialize(&signBs, uint64(i))
		//本交易特有信息
		*signDst = append(*signDst, this.Puk...)
		// fmt.Println("验证签名前的字节3", len(*signDst), hex.EncodeToString(*signDst))
		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*signDst))
		}
		if !ed25519.Verify(puk, *signDst, one.Sign) {
			return config.ERROR_sign_fail
		}

	}
	//判断输入输出是否相等
	outTotal := uint64(0)
	for _, one := range this.Vout {
		outTotal = outTotal + one.Value
	}
	// fmt.Println("这里的手续费是否正确", outTotal, inTotal, this.Gas)
	if outTotal > inTotal {
		return config.ERROR_tx_fail
	}
	this.Gas = inTotal - outTotal

	return nil
}

/*
	格式化成json字符串
*/
// func (this *Tx_deposit_in) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	格式化成[]byte
*/
func (this *Tx_deposit_in) Proto() (*[]byte, error) {
	vins := make([]*go_protos.Vin, 0)
	for _, one := range this.Vin {
		vins = append(vins, &go_protos.Vin{
			Txid: one.Txid,
			Vout: one.Vout,
			Puk:  one.Puk,
			Sign: one.Sign,
		})
	}
	vouts := make([]*go_protos.Vout, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, &go_protos.Vout{
			Value:        one.Value,
			Address:      one.Address,
			FrozenHeight: one.FrozenHeight,
		})
	}
	txBase := go_protos.TxBase{
		Hash:       this.Hash,
		Type:       this.Type,
		VinTotal:   this.Vin_total,
		Vin:        vins,
		VoutTotal:  this.Vout_total,
		Vout:       vouts,
		Gas:        this.Gas,
		LockHeight: this.LockHeight,
		Payload:    this.Payload,
		BlockHash:  this.BlockHash,
	}

	txPay := go_protos.TxDepositIn{
		TxBase: &txBase,
		Puk:    this.Puk,
	}
	// txPay.Marshal()
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	验证是否合法
*/
func (this *Tx_deposit_in) GetWitness() *crypto.AddressCoin {
	witness := crypto.BuildAddr(config.AddrPre, this.Vin[0].Puk)
	// witness, err := keystore.ParseHashByPubkey(this.Vin[0].Puk)
	// if err != nil {
	// 	return nil
	// }
	return &witness
}

/*
	检查重复的交易
	@return    bool    是否验证通过
*/
func (this *Tx_deposit_in) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }
	addrSelf := this.Vout[0].Address
	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_deposit_in {
			continue
		}
		addr := (*one.GetVout())[0].Address
		if bytes.Equal(addrSelf, addr) {
			return false
		}
	}
	return true
}

/*
	统计交易余额
*/
func (this *Tx_deposit_in) CountTxItems(height uint64) *TxItemCount {
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
		if voutIndex == 0 {
			continue
		}
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

func (this *Tx_deposit_in) CountTxHistory(height uint64) {
	//转出历史记录
	hiOut := HistoryItem{
		IsIn:    false,                          //资金转入转出方向，true=转入;false=转出;
		Type:    this.Class(),                   //交易类型
		InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
		OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
		// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
		// Value:  amount,          //交易金额
		Txid:   *this.GetHash(), //交易id
		Height: height,          //
		// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
	}
	//转入历史记录
	// hiIn := HistoryItem{
	// 	IsIn:    true,                           //资金转入转出方向，true=转入;false=转出;
	// 	Type:    this.Class(),                   //交易类型
	// 	InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
	// 	OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
	// 	// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
	// 	// Value:  amount,          //交易金额
	// 	Txid:   *this.GetHash(), //交易id
	// 	Height: height,          //
	// 	// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
	// }
	//
	addrCoin := make(map[string]bool)
	for _, vin := range this.Vin {
		addrInfo, isSelf := keystore.FindPuk(vin.Puk)
		// hiIn.InAddr = append(hiIn.InAddr, &addrInfo.Addr)
		if !isSelf {
			continue
		}
		if _, ok := addrCoin[utils.Bytes2string(addrInfo.Addr)]; ok {
			continue
		} else {
			addrCoin[utils.Bytes2string(addrInfo.Addr)] = false
		}
		hiOut.InAddr = append(hiOut.InAddr, &addrInfo.Addr)
	}

	//生成新的UTXO收益，保存到列表中
	addrCoin = make(map[string]bool)
	for voutIndex, vout := range this.Vout {
		if voutIndex != 0 {
			continue
		}
		hiOut.OutAddr = append(hiOut.OutAddr, &vout.Address)
		hiOut.Value += vout.Value
		_, ok := keystore.FindAddress(vout.Address)
		if !ok {
			continue
		}
		// hiIn.Value += vout.Value
		if _, ok := addrCoin[utils.Bytes2string(vout.Address)]; ok {
			continue
		} else {
			addrCoin[utils.Bytes2string(vout.Address)] = false
		}
		// hiIn.OutAddr = append(hiIn.OutAddr, &vout.Address)
	}
	if len(hiOut.InAddr) > 0 {
		balanceHistoryManager.Add(hiOut)
	}
	// if len(hiIn.OutAddr) > 0 {
	// 	balanceHistoryManager.Add(hiIn)
	// }
}

/*
	见证人出块成功，退还押金
*/
type Tx_deposit_out struct {
	TxBase
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_deposit_out) GetVOJSON() interface{} {
	return this.TxBase.ConversionVO()
}

/*
	构建hash值得到交易id
*/
func (this *Tx_deposit_out) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_deposit_out)
	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	对整个交易签名
*/
//func (this *Tx_deposit_out) Sign(key *keystore.Address, pwd string) (*[]byte, error) {
//	bs := this.SignSerialize()
//	return key.Sign(*bs, pwd)
//}

/*
	格式化成json字符串
*/
// func (this *Tx_deposit_out) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	验证是否合法
*/
func (this *Tx_deposit_out) Check() error {
	return this.TxBase.CheckBase()
}

/*
	检查重复的交易
	@return    bool    是否验证通过
*/
func (this *Tx_deposit_out) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }
	return true
}

/*
	统计交易余额
*/
func (this *Tx_deposit_out) CountTxItems(height uint64) *TxItemCount {
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

func (this *Tx_deposit_out) CountTxHistory(height uint64) {
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
		IsIn:    true,                           //资金转入转出方向，true=转入;false=转出;
		Type:    this.Class(),                   //交易类型
		InAddr:  make([]*crypto.AddressCoin, 0), //输入地址
		OutAddr: make([]*crypto.AddressCoin, 0), //输出地址
		// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
		// Value:  amount,          //交易金额
		Txid:   *this.GetHash(), //交易id
		Height: height,          //
		// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
	}
	//
	addrCoin := make(map[string]bool)
	for _, vin := range this.Vin {
		addrInfo, isSelf := keystore.FindPuk(vin.Puk)
		hiIn.InAddr = append(hiIn.InAddr, &addrInfo.Addr)
		if !isSelf {
			continue
		}
		if _, ok := addrCoin[utils.Bytes2string(addrInfo.Addr)]; ok {
			continue
		} else {
			addrCoin[utils.Bytes2string(addrInfo.Addr)] = false
		}
		// hiOut.InAddr = append(hiOut.InAddr, &addrInfo.Addr)
	}

	//生成新的UTXO收益，保存到列表中
	addrCoin = make(map[string]bool)
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
	// if len(hiOut.InAddr) > 0 {
	// 	balanceHistoryManager.Add(hiOut)
	// }
	if len(hiIn.OutAddr) > 0 {
		balanceHistoryManager.Add(hiIn)
	}
}

/*
	创建一个见证人押金交易
	@amount    uint64    押金额度
*/
func CreateTxDepositIn(amount, gas uint64, pwd, payload string) (*Tx_deposit_in, error) {
	// engine.Log.Debug("创建见证人押金交易 111")
	if amount < config.Mining_deposit {
		// fmt.Println("交押金数量最少", config.Mining_deposit)
		//押金太少
		return nil, config.ERROR_deposit_witness
	}
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()

	key := keystore.GetCoinbase()

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
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   amount,   //输出金额 = 实际金额 * 100000000
		Address: key.Addr, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	if total > amount {
		vout := Vout{
			Value:   total - (amount + gas), //输出金额 = 实际金额 * 100000000
			Address: key.Addr,               //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	var txin *Tx_deposit_in
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_deposit_in,               //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vouts)),                             //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			// CreateTime: time.Now().Unix(),                //创建时间
			Payload: []byte(payload),
		}
		txin = &Tx_deposit_in{
			TxBase: base,
			Puk:    key.Puk,
		}
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
			txin.Vin[i].Sign = *sign
			// for _, key := range keystore.GetAddr() {
			// 	puk, ok := keystore.GetPukByAddr(key.Addr)
			// 	if !ok {
			// 		return nil, config.ERROR_public_key_not_exist
			// 	}
			// 	if bytes.Equal(puk, one.Puk) {
			// 		_, prk, _, err := keystore.GetKeyByAddr(key.Addr, pwd)
			// 		// prk, err := key.GetPriKey(pwd)
			// 		if err != nil {
			// 			return nil, err
			// 		}
			// 		sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			// 		//				sign := txin.GetVoutsSign(prk, uint64(i))
			// 		txin.Vin[i].Sign = *sign
			// 	}
			// }
		}
		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	chain.balance.Frozen(items, txin)
	return txin, nil
}

/*
	创建一个退还押金交易
	额度超过了押金额度，那么会从自己账户余额转账到目标账户（因为考虑到押金太少还不够给手续费的情况）
	@addr      *utils.Multihash    退回到的目标账户地址
	@amount    uint64              押金额度
*/
func CreateTxDepositOut(addr string, amount, gas uint64, pwd string) (*Tx_deposit_out, error) {

	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	item := chain.balance.GetDepositIn()
	if item == nil {
		// fmt.Println("没有押金")
		//没有缴纳押金
		return nil, config.ERROR_deposit_not_exist
	}

	vins := make([]*Vin, 0)
	total := uint64(item.Value)
	//查看余额够不够
	if total < (amount + gas) {
		//余额不够给手续费，需要从其他账户余额作为输入给手续费
		// var items []*TxItem
		totalAll, items := chain.balance.BuildPayVin(nil, amount+gas-total)
		total = total + totalAll
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
	}

	// if total < (amount + gas) {
	// 	//押金不够
	// 	return nil, errors.New("余额不够")
	// }

	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin
	if addr == "" {
		//为空则转给自己
		dstAddr = keystore.GetAddr()[0].Addr
	} else {
		// var err error
		// *dstAddr, err = utils.FromB58String(addr)
		dstAddr = crypto.AddressFromB58String(addr)
		// if err != nil {
		// 	// fmt.Println("解析地址失败")
		// 	return nil
		// }
	}

	txItr, err := LoadTxBase(item.Txid)
	// txItr, err := FindTxBase(item.Txid)
	// bs, err := db.Find(item.Txid)
	// if err != nil {
	// 	//余额交易未找到
	// 	return nil, config.ERROR_tx_not_exist
	// }
	// txItr, err := ParseTxBase(ParseTxClass(item.Txid), bs)
	if err != nil {
		//解析交易出错
		return nil, config.ERROR_tx_format_fail
	}

	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	//地址字符串查私钥
	puk, ok := keystore.GetPukByAddr((*txItr.GetVout())[0].Address)
	if !ok {
		//未找到地址对应的公钥
		return nil, config.ERROR_public_key_not_exist
	}

	// prvKey, err := keystore.GetPriKeyByAddress((*txItr.GetVout())[0].Address.B58String(), pwd)
	// if err != nil {
	// 	return nil
	// }
	// //给输出签名，用于下一个输入
	// //	sign := txItr.GetSign(prvKey, item.OutIndex)

	// //公钥格式化
	// pub, err := utils.MarshalPubkey(&prvKey.PublicKey)
	// if err != nil {
	// 	return nil
	// }
	vin := Vin{
		Txid: item.Txid,      //UTXO 前一个交易的id
		Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk:  puk,            //公钥
		//		Sign: *sign,         //签名
	}
	vins = append(vins, &vin) // []Vin{vin} // append(vins, vin)

	//构建交易输出
	vouts := make([]*Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   amount,  //输出金额 = 实际金额 * 100000000
		Address: dstAddr, //钱包地址
	}
	vouts = append(vouts, &vout)

	//退还剩余的钱
	if total-amount-gas > 0 {
		resultVout := Vout{
			Value:   total - amount - gas,
			Address: keystore.GetAddr()[0].Addr,
		}
		vouts = append(vouts, &resultVout)
	}

	var txin *Tx_deposit_out
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_deposit_out,              //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			// CreateTime: time.Now().Unix(),                 //创建时间
		}
		txin = &Tx_deposit_out{
			TxBase: base,
		}
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			//				sign := pay.GetVoutsSign(prk, uint64(i))
			txin.Vin[i].Sign = *sign
			// for _, key := range keystore.GetAddr() {

			// 	puk, ok := keystore.GetPukByAddr(key.Addr)
			// 	if !ok {
			// 		//未找到地址对应的公钥
			// 		return nil, config.ERROR_public_key_not_exist
			// 	}

			// 	if bytes.Equal(puk, one.Puk) {
			// 		_, prk, _, err := keystore.GetKeyByAddr(key.Addr, pwd)

			// 		// prk, err := key.GetPriKey(pwd)
			// 		if err != nil {
			// 			return nil, err
			// 		}
			// 		sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			// 		//				sign := txin.GetVoutsSign(prk, uint64(i))
			// 		txin.Vin[i].Sign = *sign
			// 	}
			// }
		}
		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	return txin, nil
}
