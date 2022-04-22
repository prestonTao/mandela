package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"golang.org/x/crypto/ed25519"
)

const (
	type_addr = uint16(1)
	type_name = uint16(2)

	VOTE_TYPE_community = 1 //社区节点押金
	VOTE_TYPE_vote      = 2 //轻节点投票
	VOTE_TYPE_light     = 3 //轻节点押金

)

/*
	投票类型
*/
type VoteAddress []byte

/*
	获取投票者地址
*/
func (this *VoteAddress) GetAddress() crypto.AddressCoin {
	if len(*this) <= 2 {
		return nil
	}
	addrType := utils.BytesToUint16((*this)[:2])
	switch addrType {
	case type_addr:
		return crypto.AddressCoin((*this)[2:])
	case type_name:
		index := utils.BytesToUint16((*this)[2:4])
		nameStr := string((*this)[4:])
		nameinfo := name.FindNameToNet(nameStr)
		if nameinfo != nil && len(nameinfo.AddrCoins) >= int(index) {
			addrCoin := nameinfo.AddrCoins[index]
			return addrCoin
		}
	}
	return nil
}

/*
	获取投票者地址
*/
func (this *VoteAddress) B58String() string {
	addr := this.GetAddress()
	if addr == nil {
		return ""
	}
	return addr.B58String()
}

/*
	通过域名创建一个投票地址
	[2 byte:类型][2 byte:下标][n byte:域名]
*/
func NewVoteAddressByName(name string, index uint16) VoteAddress {
	if name == "" {
		return nil
	}
	//类型
	buf := bytes.NewBuffer(utils.Uint16ToBytes(type_name))
	//解析的下标
	buf.Write(utils.Uint16ToBytes(index))
	//域名
	buf.WriteString(name)

	return VoteAddress(buf.Bytes())
}

/*
	通过地址创建一个投票地址
	[2 byte:类型][n byte:地址]
*/
func NewVoteAddressByAddr(addr crypto.AddressCoin) VoteAddress {
	if addr == nil {
		return nil
	}
	//类型
	buf := bytes.NewBuffer(utils.Uint16ToBytes(type_addr))
	//解析的下标
	buf.Write(addr)

	return VoteAddress(buf.Bytes())
}

/*
	交押金，成为备用见证人
*/
type Tx_vote_in struct {
	TxBase
	Vote     VoteAddress        `json:"v"`  //见证人地址
	VoteType uint16             `json:"vt"` //投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
	VoteAddr crypto.AddressCoin `json:"va"` //本字段不用上链。投票地址，根据域名解析的地址。
}

type Tx_vote_in_VO struct {
	TxBaseVO
	Vote     string `json:"vote"`     //见证人地址
	VoteType uint16 `json:"votetype"` //投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
	VoteAddr string `json:"voteaddr"` //本字段不用上链。投票地址，根据域名解析的地址。
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_vote_in) GetVOJSON() interface{} {
	return Tx_vote_in_VO{
		TxBaseVO: this.TxBase.ConversionVO(),
		Vote:     this.Vote.B58String(),     //见证人地址
		VoteType: this.VoteType,             //投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；
		VoteAddr: this.VoteAddr.B58String(), //本字段不用上链。投票地址，根据域名解析的地址。
	}
}

/*
	构建hash值得到交易id
*/
func (this *Tx_vote_in) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()
	// *bs = append(*bs, this.Vote...)

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_vote_in)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	格式化成json字符串
*/
// func (this *Tx_vote_in) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	格式化成[]byte
*/
func (this *Tx_vote_in) Proto() (*[]byte, error) {
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

	txPay := go_protos.TxVoteIn{
		TxBase:   &txBase,
		Vote:     this.Vote,
		VoteType: uint32(this.VoteType),
		VoteAddr: this.VoteAddr,
	}
	// txPay.Marshal()
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	格式化成json字符串
*/
func (this *Tx_vote_in) Serialize() *[]byte {
	bs := this.TxBase.Serialize()
	buf := bytes.NewBuffer(*bs)
	//	voteinfo := this.Vote.SignSerialize()
	buf.Write(this.Vote)
	buf.Write(utils.Uint16ToBytes(this.VoteType))
	*bs = buf.Bytes()
	return bs
}

/*
	获取签名
*/
func (this *Tx_vote_in) GetWaitSign(txid []byte, voutIndex, vinIndex uint64) *[]byte {

	txItr, err := LoadTxBase(txid)
	// txItr, err := FindTxBase(txid)

	// bs, err := db.Find(txid)
	// if err != nil {
	// 	return nil
	// }
	// txItr, err := ParseTxBase(ParseTxClass(txid), bs)
	if err != nil {
		return nil
	}

	blockhash, err := db.GetTxToBlockHash(&txid)
	if err != nil || blockhash == nil {
		return nil
	}

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

	*signDst = append(*signDst, this.Vote...)
	*signDst = append(*signDst, utils.Uint16ToBytes(this.VoteType)...)

	//sign := keystore.Sign(*key, *signDst)
	return signDst

	// bs = txItr.GetVoutSignSerialize(voutIndex)
	// bs = this.GetSignSerialize(bs, vinIndex)
	// *bs = append(*bs, this.Vote...)

	// *bs = keystore.Sign(*key, *bs)
	// return bs
}

/*
	获取签名
*/
func (this *Tx_vote_in) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {
	txItr, err := LoadTxBase(txid)
	// txItr, err := FindTxBase(txid)

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

	*signDst = append(*signDst, this.Vote...)
	*signDst = append(*signDst, utils.Uint16ToBytes(this.VoteType)...)

	sign := keystore.Sign(*key, *signDst)

	return &sign

	// bs = txItr.GetVoutSignSerialize(voutIndex)
	// bs = this.GetSignSerialize(bs, vinIndex)
	// *bs = append(*bs, this.Vote...)

	// *bs = keystore.Sign(*key, *bs)
	// return bs
}

/*
	检查交易是否合法
*/
func (this *Tx_vote_in) Check() error {

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
		// if vout.Tx != nil {
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
		*signDst = append(*signDst, this.Vote...)
		*signDst = append(*signDst, utils.Uint16ToBytes(this.VoteType)...)
		// fmt.Println("验证签名前的字节3", len(*signDst), hex.EncodeToString(*signDst))
		puk := ed25519.PublicKey(one.Puk)
		// fmt.Printf("sign后:puk:%x signdst:%x sign:%x", md5.Sum(puk), md5.Sum(*signDst), md5.Sum(one.Sign))
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
	验证是否合法
*/
func (this *Tx_vote_in) GetWitness() *crypto.AddressCoin {
	witness := crypto.BuildAddr(config.AddrPre, this.Vin[0].Puk)
	// witness, err := keystore.ParseHashByPubkey(this.Vin[0].Puk)
	// if err != nil {
	// 	return nil
	// }
	return &witness
}

/*
	设置投票地址
*/
func (this *Tx_vote_in) SetVoteAddr(addr crypto.AddressCoin) {
	this.VoteAddr = addr
	// bs, err := this.Json()
	bs, err := this.Proto()
	if err != nil {
		return
	}
	// TxCache.FlashTxInCache(hex.EncodeToString(*this.GetHash()), this)
	// TxCache.FlashTxInCache(this.GetHashStr(), this)
	db.LevelDB.Save(*this.GetHash(), bs)
}

/*
	检查重复的交易
*/
func (this *Tx_vote_in) CheckRepeatedTx(txs ...TxItr) bool {

	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }

	addrSelf := this.Vout[0].Address

	//判断历史区块中，这个交易的角色
	oldRule := GetAddrState(addrSelf)
	// engine.Log.Info("判断历史区块中，这个交易的角色 %s %d", "", oldRule)
	switch this.VoteType {
	case VOTE_TYPE_community:
		if oldRule == VOTE_TYPE_community || oldRule == VOTE_TYPE_vote || oldRule == VOTE_TYPE_light {
			return false
		}
	case VOTE_TYPE_vote:
		//社区节点不能投票
		if oldRule == VOTE_TYPE_community {
			return false
		}
		//不能给多人投票
		vs, ok := forks.LongChain.witnessBackup.haveVoteList(&addrSelf)
		if ok {
			if !bytes.Equal(*vs.Witness, this.Vote.GetAddress()) {
				return false
			}
		}
	case VOTE_TYPE_light:
		if oldRule == VOTE_TYPE_community || oldRule == VOTE_TYPE_light {
			return false
		}
	}

	// voteAddr := this.Vote.GetAddress()
	for _, one := range txs {
		if one.Class() != config.Wallet_tx_type_vote_in {
			continue
		}
		addr := (*one.GetVout())[0].Address

		//判断本区块中，这个交易的角色
		votein := one.(*Tx_vote_in)
		rule := votein.VoteType

		isSelf := bytes.Equal(addrSelf, addr)
		if !isSelf {
			continue
		}

		switch this.VoteType {
		case VOTE_TYPE_community:
			return false
		case VOTE_TYPE_vote:
			if rule == VOTE_TYPE_community {
				return false
			}
			if rule == VOTE_TYPE_vote {
				if !bytes.Equal(this.Vote.GetAddress(), votein.Vote.GetAddress()) {
					return false
				}
			}
		case VOTE_TYPE_light:
			return false
		}
	}
	return true
}

/*
	统计交易余额
*/
func (this *Tx_vote_in) CountTxItems(height uint64) *TxItemCount {
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

func (this *Tx_vote_in) CountTxHistory(height uint64) {
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
	创建一个见证人投票交易
	@amount    uint64    押金额度
*/
func CreateTxVoteIn(voteType uint16, witnessAddr crypto.AddressCoin, addr crypto.AddressCoin, amount, gas uint64, pwd, payload string) (*Tx_vote_in, error) {
	// if amount < config.Mining_vote {
	// 	// fmt.Println("投票交押金数量最少", config.Mining_vote)
	// 	return nil, errors.New("投票交押金数量最少" + strconv.Itoa(config.Mining_vote))
	// }

	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	//查找余额
	vins := make([]*Vin, 0)
	// total := uint64(0)
	// keys := keystore.GetAddr()
	// for _, one := range keys {
	// 	bas := chain.balance.FindBalance(&one)
	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			puk, ok := keystore.GetPukByAddr(one)
	// 			if !ok {
	// 				return false
	// 			}

	// 			item := v.(*TxItem)
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

	// if total < amount+gas {
	// 	//押金不够
	// 	// fmt.Println("++++余额不足222")
	// 	return nil, errors.New("余额不足")
	// }

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

	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin = addr

	if addr == nil {
		// fmt.Println("自己地址数量", len(keystore.GetAddr()))
		//为空则转给自己
		dstAddr = keystore.GetAddr()[0].Addr
	}

	//构建交易输出
	vouts := make([]*Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   amount,  //输出金额 = 实际金额 * 100000000
		Address: dstAddr, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value:   total - amount - gas,       //输出金额 = 实际金额 * 100000000
			Address: keystore.GetAddr()[0].Addr, //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	var txin *Tx_vote_in
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_vote_in,                  //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    []byte(payload),                                //
			// CreateTime: time.Now().Unix(),             //创建时间
		}

		voteAddr := NewVoteAddressByAddr(witnessAddr)
		txin = &Tx_vote_in{
			TxBase:   base,
			Vote:     voteAddr,
			VoteType: voteType,
		}
		//给输出签名，防篡改
		for i, one := range txin.Vin {
			for _, key := range keystore.GetAddr() {

				puk, ok := keystore.GetPukByAddr(key.Addr)
				if !ok {
					// fmt.Println("签名出错 1111111111")
					//签名出错
					return nil, config.ERROR_get_sign_data_fail // errors.New("签名出错")
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key.Addr, pwd)
					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						// fmt.Println("签名出错 2222222222222", err.Error())
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
	return txin, nil
}

/*
	退还押金，赎回押金，见证人因此可能会降低排名
*/
type Tx_vote_out struct {
	TxBase
}

/*
	用于地址和txid格式化显示
*/
func (this *Tx_vote_out) GetVOJSON() interface{} {
	return this.TxBase.ConversionVO()
}

/*
	构建hash值得到交易id
*/
func (this *Tx_vote_out) BuildHash() {
	if this.Hash != nil && len(this.Hash) > 0 {
		return
	}
	bs := this.Serialize()

	id := make([]byte, 8)
	binary.PutUvarint(id, config.Wallet_tx_type_vote_out)

	this.Hash = append(id, utils.Hash_SHA3_256(*bs)...)
}

/*
	格式化成json字符串
*/
// func (this *Tx_vote_out) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	验证是否合法
*/
func (this *Tx_vote_out) Check() error {
	//判断vin是否太多
	// if len(this.Vin) > config.Mining_pay_vin_max {
	// 	return config.ERROR_pay_vin_too_much
	// }
	// if err := this.TxBase.CheckBase(lockHeight); err != nil {
	// 	return err
	// }
	//退回轻节点押金，需要取消所有投票
	for _, oneVin := range this.Vin {
		//
		if oneVin.Vout != 0 {
			continue
		}
		txItr, err := LoadTxBase(oneVin.Txid)
		// txItr, err := FindTxBase(oneVin.Txid)
		// bs, err := db.Find(oneVin.Txid)
		// if err != nil {
		// 	//不能找到上一个交易，程序出错退出
		// 	return config.ERROR_tx_not_exist
		// }
		// txItr, err := ParseTxBase(ParseTxClass(oneVin.Txid), bs)
		if err != nil {
			//不能解析上一个交易，程序出错退出
			return config.ERROR_tx_format_fail
		}

		//因为有可能退回金额不够手续费，所以输入中加入了其他类型交易
		if txItr.Class() != config.Wallet_tx_type_vote_in {
			continue
		}
		votein := txItr.(*Tx_vote_in)
		vout := (*txItr.GetVout())[oneVin.Vout]

		if votein.VoteType == 3 {
			//退回轻节点押金前需要取消所有投票
			vs, ok := forks.GetLongChain().witnessBackup.haveVoteList(&vout.Address)
			if ok {
				engine.Log.Error("这个交易验证不通过 %s %+v\n%s", vout.Address.B58String(), vs, hex.EncodeToString(this.Hash))
				return config.ERROR_vote_exist
			}
		}
		return nil
	}
	return config.ERROR_tx_fail
}

/*
	验证是否合法
*/
func (this *Tx_vote_out) GetWitness() *crypto.AddressCoin {
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
func (this *Tx_vote_out) CheckRepeatedTx(txs ...TxItr) bool {
	//判断双花
	// if !this.MultipleExpenditures(txs...) {
	// 	return false
	// }
	return true
}

/*
	统计交易余额
*/
func (this *Tx_vote_out) CountTxItems(height uint64) *TxItemCount {
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

func (this *Tx_vote_out) CountTxHistory(height uint64) {
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
	创建一个投票押金退还交易
	退还按交易为单位，交易的押金全退
*/
func CreateTxVoteOut(txid []byte, addr crypto.AddressCoin, amount, gas uint64, pwd string) (*Tx_vote_out, error) {
	// fmt.Println("==============1")
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	//查找余额
	vins := make([]*Vin, 0)
	total := uint64(0)

	var item = chain.balance.GetVoteInByTxid(txid)
	// var item *TxItem
	// b.Txs.Range(func(txidItr, v interface{}) bool {
	// 	dstTxid := txidItr.(string)
	// 	//0600000000000000b027d84883693a16de4df892c4d856cbf103ed0e28a2d5d98277199ea2d79345_0
	// 	if utils.Bytes2string(txid) == strings.SplitN(dstTxid, "_", 2)[0] {
	// 		item = v.(*TxItem)
	// 		return false
	// 	}
	// 	return true
	// })

	if item == nil {
		//未找到这个交易
		return nil, config.ERROR_tx_not_exist // errors.New("未找到这个交易")
	}
	txItr, err := LoadTxBase(item.Txid)
	// txItr, err := FindTxBase(item.Txid)
	// bs, err := db.Find(item.Txid)
	// if err != nil {
	// 	return nil, err
	// }
	// txItr, err := ParseTxBase(ParseTxClass(item.Txid), bs)
	if err != nil {
		return nil, err
	}
	vout := (*txItr.GetVout())[item.VoutIndex]

	puk, ok := keystore.GetPukByAddr(vout.Address)
	if !ok {
		return nil, config.ERROR_public_key_not_exist
	}

	vin := Vin{
		Txid: item.Txid,      //UTXO 前一个交易的id
		Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
		Puk:  puk,            //公钥
		//			Sign: *sign,         //签名
	}
	vins = append(vins, &vin)
	total = total + item.Value

	// fmt.Println("==============3")
	//资金不够
	if total < amount+gas {
		//余额不够给手续费，需要从其他账户余额作为输入给手续费
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

	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin = addr
	if addr == nil {
		//为空则转给自己
		dstAddr = keystore.GetAddr()[0].Addr
	}
	// fmt.Println("==============6")

	//构建交易输出
	vouts := make([]*Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout2 := Vout{
		Value:   total - gas, //输出金额 = 实际金额 * 100000000
		Address: dstAddr,     //钱包地址
	}
	vouts = append(vouts, &vout2)

	//	crateTime := time.Now().Unix()

	var txout *Tx_vote_out
	for i := uint64(0); i < 10000; i++ {
		//
		base := TxBase{
			Type:       config.Wallet_tx_type_vote_out,                 //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			// CreateTime: time.Now().Unix(),              //创建时间
		}
		txout = &Tx_vote_out{
			TxBase: base,
		}
		// fmt.Println("==============7")

		//给输出签名，防篡改
		for i, one := range txout.Vin {
			for _, key := range keystore.GetAddr() {
				puk, ok := keystore.GetPukByAddr(key.Addr)
				if !ok {
					//未找到公钥
					return nil, config.ERROR_public_key_not_exist // errors.New("未找到公钥")
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key.Addr, pwd)

					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						// fmt.Println("获取key错误")
						return nil, err
					}
					sign := txout.GetSign(&prk, one.Txid, one.Vout, uint64(i))
					//				sign := txout.GetVoutsSign(prk, uint64(i))
					txout.Vin[i].Sign = *sign
				}
			}
		}
		// fmt.Println("==============8")
		txout.BuildHash()
		if txout.CheckHashExist() {
			txout = nil
			continue
		} else {
			break
		}
	}
	return txout, nil
}

// /*
// 	退还一笔指定交易的投票
// */
// func CreateTxVoteOutOne(item *TxItem, addr string, amount, gas uint64, pwd string) (*Tx_vote_out, error) {
// 	if item == nil {
// 		return nil, errors.New("退押金交易未找到")
// 	}
// 	// fmt.Println("==============1")
// 	chain := forks.GetLongChain()
// 	_, block := chain.GetLastBlock()
// 	// b := chain.balance.GetVoteInByTxid(witness)
// 	// if b == nil {
// 	// 	// fmt.Println("++++押金不够")
// 	// 	return nil
// 	// }
// 	// fmt.Println("==============2")
// 	//查找余额
// 	vins := make([]Vin, 0)
// 	total := uint64(0)
// 	bs, err := db.Find(item.Txid)
// 	if err != nil {
// 		return nil, errors.New("退押金交易未找到")
// 	}
// 	txItr, err := ParseTxBase(bs)
// 	if err != nil {
// 		return nil, errors.New("退押金交易未找到")
// 	}
// 	vout := (*txItr.GetVout())[item.OutIndex]

// 	puk, ok := keystore.GetPukByAddr(vout.Address)
// 	if !ok {
// 		return nil, errors.New("退押金交易未找到")
// 	}
// 	vin := Vin{
// 		Txid: item.Txid,     //UTXO 前一个交易的id
// 		Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 		Puk:  puk,           //公钥
// 		//			Sign: *sign,         //签名
// 	}
// 	vins = append(vins, vin)

// 	total = total + item.Value
// 	// if total > amount+gas {
// 	// 	return nil, errors.New("余额不足")
// 	// }
// 	// fmt.Println("==============3")
// 	//资金不够
// 	if total < amount+gas {
// 		//余额不够给手续费，需要从其他账户余额作为输入给手续费
// 		for _, one := range keystore.GetAddr() {
// 			bas, err := chain.balance.FindBalance(&one)
// 			if err != nil {
// 				// fmt.Println("==============3.1")
// 				return nil, errors.New("余额不足")
// 			}
// 			// fmt.Println("==============3.2")
// 			for _, two := range bas {
// 				two.Txs.Range(func(k, v interface{}) bool {
// 					puk, ok := keystore.GetPukByAddr(one)
// 					if !ok {
// 						return false
// 					}

// 					item := v.(*TxItem)
// 					vin := Vin{
// 						Txid: item.Txid,     //UTXO 前一个交易的id
// 						Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 						Puk:  puk,           //公钥
// 						//						Sign: *sign,           //签名
// 					}
// 					vins = append(vins, vin)

// 					total = total + item.Value
// 					if total >= amount+gas {
// 						return false
// 					}

// 					return true
// 				})

// 				// fmt.Println("==============3.3")
// 				if total >= amount+gas {

// 					// fmt.Println("==============3.4")
// 					break
// 				}
// 			}
// 		}

// 		// fmt.Println("==============3.5")
// 		//		return nil
// 	}
// 	// fmt.Println("==============4")
// 	//余额不够给手续费
// 	if total < (amount + gas) {
// 		// fmt.Println("押金不够")
// 		//押金不够
// 		return nil, errors.New("余额不足")
// 	}
// 	// fmt.Println("==============5")

// 	//解析转账目标账户地址
// 	var dstAddr crypto.AddressCoin
// 	if addr == "" {
// 		//为空则转给自己
// 		dstAddr = keystore.GetAddr()[0]
// 	} else {
// 		// var err error
// 		// *dstAddr, err = utils.FromB58String(addr)
// 		// if err != nil {
// 		// 	// fmt.Println("解析地址失败")
// 		// 	return nil
// 		// }
// 		dstAddr = crypto.AddressFromB58String(addr)
// 	}
// 	// fmt.Println("==============6")

// 	//构建交易输出
// 	vouts := make([]Vout, 0)
// 	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
// 	vout = Vout{
// 		Value:   total - gas, //输出金额 = 实际金额 * 100000000
// 		Address: dstAddr,     //钱包地址
// 	}
// 	vouts = append(vouts, vout)

// 	//	crateTime := time.Now().Unix()

// 	var txout *Tx_vote_out
// 	for i := uint64(0); i < 10000; i++ {
// 		//
// 		base := TxBase{
// 			Type:       config.Wallet_tx_type_vote_out, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
// 			Vin_total:  uint64(len(vins)),              //输入交易数量
// 			Vin:        vins,                           //交易输入
// 			Vout_total: uint64(len(vouts)),             //输出交易数量
// 			Vout:       vouts,                          //
// 			Gas:        gas,                            //交易手续费
// 			LockHeight: block.Height + 100 + i,         //锁定高度
// 			//		CreateTime: crateTime,                      //创建时间
// 		}
// 		txout = &Tx_vote_out{
// 			TxBase: base,
// 		}
// 		// fmt.Println("==============7")

// 		//给输出签名，防篡改
// 		for i, one := range txout.Vin {
// 			for _, key := range keystore.GetAddr() {
// 				puk, ok := keystore.GetPukByAddr(key)
// 				if !ok {
// 					return nil, errors.New("签名查找公钥出错")
// 				}

// 				if bytes.Equal(puk, one.Puk) {
// 					_, prk, _, err := keystore.GetKeyByAddr(key, pwd)

// 					// prk, err := key.GetPriKey(pwd)
// 					if err != nil {
// 						// fmt.Println("获取key错误")
// 						return nil, errors.New("签名获取key错误")
// 					}
// 					sign := txout.GetSign(&prk, one.Txid, one.Vout, uint64(i))
// 					//				sign := txout.GetVoutsSign(prk, uint64(i))
// 					txout.Vin[i].Sign = *sign
// 				}
// 			}
// 		}
// 		// fmt.Println("==============8")
// 		txout.BuildHash()
// 		if txout.CheckHashExist() {
// 			txout = nil
// 			continue
// 		} else {
// 			break
// 		}
// 	}
// 	return txout, nil
// }
