package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	mc "mandela/core/message_center"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"bytes"
	"encoding/hex"
	"errors"
	"runtime"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"golang.org/x/crypto/ed25519"
)

const (
// BlockTx_Gas       = "gas"
// BlockTx_Hash      = "hash"
// BlockTx_Vout      = "vout"
// BlockTx_Vout_Tx   = "tx"
// BlockTx_Blockhash = "blockhash"
)

type TxItr interface {
	Class() uint64    //交易类型
	BuildHash()       //构建交易hash
	GetHash() *[]byte //获得交易hash
	// GetHashStr() string                                                                               //获取交易hash字符串
	CheckLockHeight(lockHeight uint64) error                       //检查锁定高度是否合法
	CheckFrozenHeight(frozenHeight uint64, frozenTime int64) error //检查余额冻结高度
	Check() error                                                  //检查交易是否合法
	CheckHashExist() bool                                          //检查hash在数据库中是否已经存在
	CheckRepeatedTx(txs ...TxItr) bool                             //验证同一区块（或者同一组区块）中是否有相同的交易，比如重复押金。
	// Json() (*[]byte, error)                                                           //将交易格式化成json字符串
	Proto() (*[]byte, error)                                                          //将交易格式化成proto字节
	Serialize() *[]byte                                                               //将需要签名的字段序列化
	GetVin() *[]*Vin                                                                  //
	GetVout() *[]*Vout                                                                //
	GetGas() uint64                                                                   //
	GetVoutSignSerialize(voutIndex uint64) *[]byte                                    //获取交易输出序列化
	GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte //获取签名
	// SetBlockHash(bs []byte)                                                           //设置本交易所属的区块hash
	// GetBlockHash() *[]byte                                                            //
	GetLockHeight() uint64                //获取锁定区块高度
	SetSign(index uint64, bs []byte) bool //修改签名
	SetPayload(bs []byte)                 //修改备注
	GetPayload() []byte                   //获取备注内容
	GetVOJSON() interface{}               //用于地址和txid格式化显示
	// CheckSign() bool                         //检查签名是否正确，异步验证
	// CheckRepeated(tx TxItr) bool                                                       //验证是否双重支付，押金是否重复，使用同步验证
	// Balance() *sync.Map                                                               //查询交易输出，统计输出地址余额key:utils.Multihash=收款地址;value:TxItem=地址余额;
	// SetTxid(index uint64, txid *[]byte) error //这个交易输出被使用之后，需要把UTXO输出标记下
	// UnSetTxid(bs *[]byte, index uint64) error                                         //区块回滚，把之前标记为已经使用过的交易的标记去掉
	CountTxItems(height uint64) *TxItemCount //统计可用余额
	CountTxHistory(height uint64)            //统计交易记录
}

/*
	交易
*/
type TxBase struct {
	Hash       []byte  `json:"h"`     //本交易hash，不参与区块hash，只用来保存
	Type       uint64  `json:"t"`     //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
	Vin_total  uint64  `json:"vin_t"` //输入交易数量
	Vin        []*Vin  `json:"vin"`   //交易输入
	Vout_total uint64  `json:"vot_t"` //输出交易数量
	Vout       []*Vout `json:"vot"`   //交易输出
	Gas        uint64  `json:"g"`     //交易手续费，此字段不参与交易hash
	LockHeight uint64  `json:"l_h"`   //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
	Payload    []byte  `json:"p"`     //备注信息
	BlockHash  []byte  `json:"bh"`    //本交易属于的区块hash，不参与区块hash，只用来保存
}

/*
	交易
*/
type TxBaseVO struct {
	Hash       string    `json:"hash"`        //本交易hash，不参与区块hash，只用来保存
	Type       uint64    `json:"type"`        //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
	Vin_total  uint64    `json:"vin_total"`   //输入交易数量
	Vin        []*VinVO  `json:"vin"`         //交易输入
	Vout_total uint64    `json:"vout_total"`  //输出交易数量
	Vout       []*VoutVO `json:"vout"`        //交易输出
	Gas        uint64    `json:"gas"`         //交易手续费，此字段不参与交易hash
	LockHeight uint64    `json:"lock_height"` //本交易锁定在低于这个高度的块中，超过这个高度，块将不被打包到区块中。
	Payload    string    `json:"payload"`     //备注信息
	BlockHash  string    `json:"blockhash"`   //本交易属于的区块hash，不参与区块hash，只用来保存
}

/*
	转化为VO对象
*/
func (this *TxBase) ConversionVO() TxBaseVO {
	vins := make([]*VinVO, 0)
	for _, one := range this.Vin {
		vins = append(vins, one.ConversionVO())
	}

	vouts := make([]*VoutVO, 0)
	for _, one := range this.Vout {
		vouts = append(vouts, one.ConversionVO())
	}

	return TxBaseVO{
		Hash:       hex.EncodeToString(this.Hash),      //本交易hash，不参与区块hash，只用来保存
		Type:       this.Type,                          //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  this.Vin_total,                     //输入交易数量
		Vin:        vins,                               //交易输入
		Vout_total: this.Vout_total,                    //输出交易数量
		Vout:       vouts,                              //交易输出
		Gas:        this.Gas,                           //交易手续费，此字段不参与交易hash
		LockHeight: this.LockHeight,                    //本交易锁定在低于这个高度的块中，超过这个高度，块将不被打包到区块中。
		Payload:    string(this.Payload),               //备注信息
		BlockHash:  hex.EncodeToString(this.BlockHash), //本交易属于的区块hash，不参与区块hash，只用来保存
	}
}

//修改签名
func (this *TxBase) SetSign(index uint64, bs []byte) bool {
	this.Vin[index].Sign = bs
	return true
}

/*
	设置本交易所属的区块hash
*/
func (this *TxBase) SetBlockHash(bs []byte) {
	this.BlockHash = bs
}

/*
	设置本交易所属的区块hash
*/
func (this *TxBase) GetBlockHash() *[]byte {
	return &this.BlockHash
}

/*
	获取锁定区块高度
*/
func (this *TxBase) GetLockHeight() uint64 {
	return this.LockHeight
}

/*
	将需要hash的字段序列化
*/
func (this *TxBase) Serialize() *[]byte {
	length := 0
	var vinSs []*[]byte
	if this.Vin != nil {
		vinSs = make([]*[]byte, 0, len(this.Vin))
		for _, one := range this.Vin {
			bsOne := one.SerializeVin()
			vinSs = append(vinSs, bsOne)
			length += len(*bsOne)
		}
	}
	var voutSs []*[]byte
	if this.Vout != nil {
		voutSs = make([]*[]byte, 0, len(this.Vout))
		for _, one := range this.Vout {
			bsOne := one.Serialize()
			voutSs = append(voutSs, bsOne)
			length += len(*bsOne)
		}
	}
	length += 8 + 8 + 8 + 8 + len(this.Payload)
	bs := make([]byte, 0, length)

	bs = append(bs, utils.Uint64ToBytes(this.Type)...)
	bs = append(bs, utils.Uint64ToBytes(this.Vin_total)...)
	if vinSs != nil {
		for _, one := range vinSs {
			bs = append(bs, *one...)
		}
	}
	bs = append(bs, utils.Uint64ToBytes(this.Vout_total)...)
	if voutSs != nil {
		for _, one := range voutSs {
			bs = append(bs, *one...)
		}
	}
	bs = append(bs, utils.Uint64ToBytes(this.LockHeight)...)
	bs = append(bs, this.Payload...)
	return &bs

	//---------------------
	// buf := bytes.NewBuffer(nil)
	// buf.Write(utils.Uint64ToBytes(this.Type))
	// buf.Write(utils.Uint64ToBytes(this.Vin_total))
	// if this.Vin != nil {
	// 	for _, one := range this.Vin {
	// 		bs := one.SerializeVin()
	// 		buf.Write(*bs)
	// 	}
	// }
	// buf.Write(utils.Uint64ToBytes(this.Vout_total))
	// if this.Vout != nil {
	// 	for _, one := range this.Vout {
	// 		bs := one.Serialize()
	// 		buf.Write(*bs)
	// 	}
	// }
	// //	buf.Write(utils.Int64ToBytes(this.CreateTime))
	// // buf.Write(utils.Uint64ToBytes(this.Gas))
	// buf.Write(utils.Uint64ToBytes(this.LockHeight))
	// buf.Write(this.Payload)
	// // buf.Write(utils.Uint64ToBytes(this.FrozenHeight))

	// // buf.Write(utils.Int64ToBytes(this.CreateTime))
	// bs := buf.Bytes()
	// return &bs
}

func (this *TxBase) GetVin() *[]*Vin {
	return &this.Vin
}
func (this *TxBase) GetVout() *[]*Vout {
	return &this.Vout
}

func (this *TxBase) GetGas() uint64 {
	return this.Gas
}

func (this *TxBase) GetHash() *[]byte {
	return &this.Hash
}

// func (this *TxBase) GetHashStr() string {
// 	if this.HashStr == "" {
// 		this.HashStr = hex.EncodeToString(this.Hash)
// 	}
// 	return this.HashStr
// }

func (this *TxBase) Class() uint64 {
	return this.Type
}

/*
	修改备注
*/
func (this *TxBase) SetPayload(bs []byte) {
	this.Payload = bs
}

/*
	获取备注内容
*/
func (this *TxBase) GetPayload() []byte {
	return this.Payload
}

/*
	获取输出序列化
	[UTXO输入引用的块hash]+[UTXO输入引用的块交易hash]+[UTXO输入引用的输出index(uint64)]+
	[UTXO输入引用的输出序列化]
*/
func (this *TxBase) GetVoutSignSerialize(voutIndex uint64) *[]byte {
	if voutIndex > uint64(len(this.Vout)) {
		return nil
	}
	vout := this.Vout[voutIndex]
	voutBs := vout.Serialize()
	bs := make([]byte, 0, len(*voutBs)+8)
	bs = append(bs, utils.Uint64ToBytes(voutIndex)...)
	bs = append(bs, *voutBs...)
	return &bs

	//-----------------
	// // //上一个交易 所属的区块hash
	// // buf := bytes.NewBuffer(this.BlockHash)
	// // //上一个交易的hash
	// // buf.Write(this.Hash)
	// buf := bytes.NewBuffer(nil)

	// //上一个交易的指定输出序列化
	// buf.Write(utils.Uint64ToBytes(voutIndex))
	// vout := this.Vout[voutIndex]
	// bs := vout.Serialize()
	// buf.Write(*bs)
	// *bs = buf.Bytes()
	// return bs
}

/*
	获取本交易用作签名的序列化
	[上一个交易GetVoutSignSerialize()返回]+[本交易类型]+[本交易输入总数]+[本交易输入index]+
	[本交易输出总数]+[vouts序列化]+[锁定区块高度]
	@voutBs    *[]byte    上一个交易GetVoutSignSerialize()返回
*/
func (this *TxBase) GetSignSerialize(voutBs *[]byte, vinIndex uint64) *[]byte {
	if vinIndex > uint64(len(this.Vin)) {
		return nil
	}

	voutBssLenght := 0
	voutBss := make([]*[]byte, 0, len(this.Vout))
	for _, one := range this.Vout {
		voutBsOne := one.Serialize()
		voutBss = append(voutBss, voutBsOne)
		voutBssLenght += len(*voutBsOne)
	}

	var bs []byte
	if voutBs == nil {
		bs = make([]byte, 0, 8+8+8+8+voutBssLenght+8+len(this.Payload))
	} else {
		bs = make([]byte, 0, len(*voutBs)+8+8+8+8+voutBssLenght+8+len(this.Payload))
		bs = append(bs, *voutBs...)
	}

	bs = append(bs, utils.Uint64ToBytes(this.Type)...)
	bs = append(bs, utils.Uint64ToBytes(this.Vin_total)...)
	bs = append(bs, utils.Uint64ToBytes(vinIndex)...)
	bs = append(bs, utils.Uint64ToBytes(this.Vout_total)...)
	for _, one := range voutBss {
		bs = append(bs, *one...)
	}
	bs = append(bs, utils.Uint64ToBytes(this.LockHeight)...)
	bs = append(bs, this.Payload...)
	return &bs

	//-----------------------
	// buf := bytes.NewBuffer(nil)
	// if voutBs != nil {
	// 	buf.Write(*voutBs)
	// }
	// // buf := bytes.NewBuffer(*voutBs)
	// // engine.Log.Info("1111111 %d", hex.EncodeToString(utils.Uint64ToBytes(this.Type)))
	// buf.Write(utils.Uint64ToBytes(this.Type))
	// // engine.Log.Info("1111111 %d", hex.EncodeToString(utils.Uint64ToBytes(this.Type)))
	// buf.Write(utils.Uint64ToBytes(this.Vin_total))
	// buf.Write(utils.Uint64ToBytes(vinIndex))
	// buf.Write(utils.Uint64ToBytes(this.Vout_total))
	// bs := make([]byte, 0)
	// for _, one := range this.Vout {
	// 	bs = append(bs, *one.Serialize()...)
	// }
	// // engine.Log.Info("1111111 %s", hex.EncodeToString(bs))
	// buf.Write(bs)
	// buf.Write(utils.Uint64ToBytes(this.LockHeight))
	// buf.Write(this.Payload)
	// // buf.Write(utils.Uint64ToBytes(this.FrozenHeight))
	// // buf.Write(utils.Int64ToBytes(this.CreateTime))
	// bs = buf.Bytes()
	// return &bs
}

/*
	获取签名
*/
func (this *TxBase) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {
	// start := time.Now()

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

	//-------------------------
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

	// engine.Log.Info("签名字符序列化 耗时 %s", time.Now().Sub(start))
	// fmt.Println("签名前的字节", len(*signDst), hex.EncodeToString(*signDst), "\n")
	// fmt.Printf("签名前的字节 len=%d signDst=%s key=%s \n", len(*signDst), hex.EncodeToString(*signDst), hex.EncodeToString(*key))
	sign := keystore.Sign(*key, *signDst)

	// engine.Log.Info("签名 耗时 %s", time.Now().Sub(start))
	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return &sign
}

/*
	获取待签名数据
*/
func (this *TxBase) GetWaitSign(txid []byte, voutIndex, vinIndex uint64) *[]byte {
	// txItr, err := FindTxBase(txid)
	txItr, err := LoadTxBase(txid)
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

	//---------------------------
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

	//fmt.Println("预签名字节", len(*bs), hex.EncodeToString(*bs), "\n")
	//sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return signDst
}

/*
	检查锁定高度是否合法
	判断交易中锁定高度值是否大于等于参数值
	@localHeight    uint64    交易锁定高度
*/
func (this *TxBase) CheckLockHeight(lockHeight uint64) error {
	// engine.Log.Info("对比锁定区块高度 %d %d", this.GetLockHeight(), lockHeight)
	if lockHeight < config.Mining_block_start_height+config.Mining_block_start_height_jump {
		return nil
	}

	if this.GetLockHeight() < lockHeight {
		// engine.Log.Warn("对比锁定区块高度 失败 %d %d %s", this.GetLockHeight(), lockHeight, hex.EncodeToString(*this.GetHash()))
		engine.Log.Warn("Failed to compare lock block height: LockHeight=%d %d %s", this.GetLockHeight(), lockHeight, hex.EncodeToString(*this.GetHash()))
		return config.ERROR_tx_lockheight
	}
	return nil
}

/*
	检查输入中对应的上一个交易的冻结高度
	判断交易中冻结高度值是否大于参数值
	@frozenHeight    uint64    交易冻结高度
*/
func (this *TxBase) CheckFrozenHeight(frozenHeight uint64, frozenTime int64) error {
	// engine.Log.Info("对比锁定区块高度 %d %d", this.GetLockHeight(), frozenHeight)

	// if this.FrozenHeight <= frozenHeight {
	// 	// engine.Log.Warn("对比锁定区块高度 失败")
	// 	return config.ERROR_tx_frozenheight
	// }

	if GetLongChain().GetCurrentBlock() < config.Mining_block_start_height+config.Mining_block_start_height_jump {
		return nil
	}

	if this.Class() == config.Wallet_tx_type_mining {
		return nil
	}

	for index, one := range this.Vin {
		key := BuildKeyForUnspentTransaction(one.Txid, one.Vout)
		// engine.Log.Info("查询")
		bs, err := db.LevelTempDB.Find(key)
		if err != nil {
			// if err == leveldb.ErrNotFound {
			// 	//认为这是一个空数据库
			// 	return nil
			// }
			return errors.New("vin index " + strconv.Itoa(index) + " " + err.Error()) //err
		}
		fh := utils.BytesToUint64(*bs)

		if !CheckFrozenHeightFree(fh, frozenHeight, frozenTime) {
			engine.Log.Error("Failed to compare frozen height %d %d %d %d", fh, frozenHeight, GetHighestBlock(), config.Mining_block_start_height+config.Mining_block_start_height_jump)
			return config.ERROR_tx_frozenheight
		}

		// if fh > config.Wallet_frozen_time_min {
		// 	//使用时间锁仓
		// 	if int64(fh) > frozenTime {
		// 		engine.Log.Error("Failed to compare freezing time %d %d", fh, frozenTime)
		// 		return config.ERROR_tx_frozenheight
		// 	}
		// } else {
		// 	//使用区块高度锁仓
		// 	if fh >= frozenHeight {
		// 		engine.Log.Error("Failed to compare frozen height %d %d %d %d", fh, frozenHeight, GetHighestBlock(), config.Mining_block_start_height+config.Mining_block_start_height_jump)
		// 		return config.ERROR_tx_frozenheight
		// 	}
		// }

	}

	return nil
}

/*
	检查交易是否合法
	@localHeight    uint64    交易锁定高度
*/
func (this *TxBase) CheckBase() error {
	// fmt.Println("开始验证交易合法性 Tx_deposit_in")
	//判断vin是否太多
	// if len(this.Vin) > config.Mining_pay_vin_max {
	// 	return config.ERROR_pay_vin_too_much
	// }

	//不能出现余额为0的转账
	for i, one := range this.Vout {
		if i != 0 && one.Value <= 0 {
			return config.ERROR_amount_zero
		}
	}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费;3.输入不能重复。
	vinMap := make(map[string]int)
	inTotal := uint64(0)
	for i, one := range this.Vin {

		//不能有重复的vin
		key := string(BuildKeyForUnspentTransaction(one.Txid, one.Vout))
		if _, ok := vinMap[key]; ok {
			return config.ERROR_tx_Repetitive_vin
		}
		vinMap[key] = 0

		// start := time.Now()
		// engine.Log.Debug("CheckBase 00000000000000 %s", time.Now().Sub(start))
		// txItr, err := FindTxBase(one.Txid)
		txItr, err := LoadTxBase(one.Txid)

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

		// engine.Log.Debug("CheckBase 111111111111111 %s", time.Now().Sub(start))

		vout := (*txItr.GetVout())[one.Vout]
		//如果这个交易已经被使用，则验证不通过，否则会出现双花问题。
		// if vout.Txid != nil {
		// 	engine.Log.Debug("There were two spending mistakes in the deal %s %s %d", hex.EncodeToString(vout.Tx), hex.EncodeToString(one.Txid), one.Vout)
		// 	return config.ERROR_tx_is_use
		// }
		inTotal = inTotal + vout.Value

		//验证公钥是否和地址对应
		addr := crypto.BuildAddr(config.AddrPre, one.Puk)

		if !bytes.Equal(addr, (*txItr.GetVout())[one.Vout].Address) {
			// engine.Log.Debug("ERROR_sign_fail 111 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			engine.Log.Debug("ERROR_sign_fail 111 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			return config.ERROR_public_and_addr_notMatch
		}
		// engine.Log.Debug("CheckBase 22222222222222222 %s", time.Now().Sub(start))

		//验证签名
		voutSS := txItr.GetVoutSignSerialize(one.Vout)
		bs := make([]byte, 0, len(*blockhash)+len(*txItr.GetHash())+len(*voutSS))
		// buf := bytes.NewBuffer(nil)
		//上一个交易 所属的区块hash
		// buf.Write(*blockhash)
		bs = append(bs, *blockhash...)
		// buf.Write(*txItr.GetBlockHash())
		//上一个交易的hash
		// buf.Write(*txItr.GetHash())
		bs = append(bs, *txItr.GetHash()...)
		//上一个交易的指定输出序列化
		// buf.Write(*txItr.GetVoutSignSerialize(one.Vout))
		bs = append(bs, *voutSS...)
		//本交易类型输入输出数量等信息和所有输出
		// bs := buf.Bytes()
		// engine.Log.Debug("CheckBase 333333333333333333 %s", time.Now().Sub(start))
		sign := this.GetSignSerialize(&bs, uint64(i))

		// engine.Log.Debug("CheckBase 44444444444444 %s", time.Now().Sub(start))

		puk := ed25519.PublicKey(one.Puk)
		if config.Wallet_print_serialize_hex {
			engine.Log.Info("sign serialize:%s", hex.EncodeToString(*sign))
		}
		// fmt.Printf("txid:%x puk:%x sign:%x", md5.Sum(one.Txid), md5.Sum(one.Puk), md5.Sum(one.Sign))
		if !ed25519.Verify(puk, *sign, one.Sign) {
			// engine.Log.Debug("ERROR_sign_fail 222 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			// engine.Log.Debug("ed25519.Verify: puk: %x; waitSignData: %x; sign: %x\n", puk, *sign, one.Sign)
			engine.Log.Debug("ERROR_sign_fail 222 %s %d", hex.EncodeToString(one.Txid), one.Vout)
			return config.ERROR_sign_fail
		}
		// engine.Log.Debug("CheckBase 5555555555555555 %s", time.Now().Sub(start))

	}

	//判断输入输出是否相等
	outTotal := uint64(0)
	for _, one := range this.Vout {
		outTotal = outTotal + one.Value
	}
	// engine.Log.Info("这里的手续费是否正确 %d %d %d", outTotal, inTotal, this.Gas)
	if outTotal > inTotal {
		return config.ERROR_tx_fail
	}
	this.Gas = inTotal - outTotal
	return nil
}

// func (this *TxBase) Balance() *sync.Map {
// 	result := new(sync.Map)
// 	for i, one := range this.Vout {
// 		if one.Txid != nil {
// 			continue
// 		}
// 		item := TxItem{
// 			Addr:     &one.Address,
// 			Value:    one.Value, //余额
// 			Txid:     this.Hash, //交易id
// 			OutIndex: uint64(i), //交易输出index，从0开始
// 		}
// 		result.Store(one.Address.B58String(), item)
// 	}
// 	return result
// }

/*
	这个交易输出被使用之后，需要把UTXO输出标记下
*/
// func (this *TxBase) SetTxid(index uint64, txid *[]byte) error {
// 	this.Vout[index].Txid = *txid
// 	return nil

// 	// txMap := make(map[string]interface{})
// 	// var jso = jsoniter.ConfigCompatibleWithStandardLibrary
// 	// err := jso.Unmarshal(*bs, &txMap)
// 	// // decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	// // decoder.UseNumber()
// 	// // err := decoder.Decode(&txMap)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// v := txMap["vout"]
// 	// if v == nil {
// 	// 	//解析失败
// 	// 	return errors.New("Resolution failure")
// 	// }
// 	// vs := v.([]interface{})
// 	// vouts := make([]Vout, 0)
// 	// for _, one := range vs {
// 	// 	voutBs, err := json.Marshal(one)
// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	vout := new(Vout)
// 	// 	var jso = jsoniter.ConfigCompatibleWithStandardLibrary
// 	// 	err = jso.Unmarshal(voutBs, vout)
// 	// 	// decoder := json.NewDecoder(bytes.NewBuffer(voutBs))
// 	// 	// decoder.UseNumber()
// 	// 	// err = decoder.Decode(vout)

// 	// 	if err != nil {
// 	// 		return err
// 	// 	}
// 	// 	vouts = append(vouts, *vout)
// 	// }

// 	// vouts[index].Tx = *txid
// 	// txMap["vout"] = vouts

// 	// txbs, err := json.Marshal(txMap)
// 	// if err != nil {
// 	// 	return err
// 	// }
// 	// err = db.Save(this.Hash, &txbs)
// 	// return err
// }

/*
	区块回滚，把之前标记为已经使用过的交易的标记去掉
*/
// func (this *TxBase) UnSetTxid(bs *[]byte, index uint64) error {
// 	txMap := make(map[string]interface{})
// 	// var jso = jsoniter.ConfigCompatibleWithStandardLibrary
// 	// err := json.Unmarshal(*bs, &txMap)
// 	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	decoder.UseNumber()
// 	err := decoder.Decode(&txMap)
// 	if err != nil {
// 		return err
// 	}
// 	v := txMap["vout"]
// 	if v == nil {
// 		//解析失败
// 		return errors.New("Resolution failure")
// 	}
// 	vs := v.([]interface{})
// 	vouts := make([]Vout, 0)
// 	for _, one := range vs {
// 		voutBs, err := json.Marshal(one)
// 		if err != nil {
// 			return err
// 		}
// 		vout := new(Vout)
// 		// var jso = jsoniter.ConfigCompatibleWithStandardLibrary
// 		// err = json.Unmarshal(voutBs, vout)

// 		decoder := json.NewDecoder(bytes.NewBuffer(voutBs))
// 		decoder.UseNumber()
// 		err = decoder.Decode(vout)

// 		if err != nil {
// 			return err
// 		}
// 		vouts = append(vouts, *vout)
// 	}

// 	vouts[index].Tx = nil
// 	txMap["vout"] = vouts

// 	txbs, err := json.Marshal(txMap)
// 	if err != nil {
// 		return err
// 	}
// 	err = db.Save(this.Hash, &txbs)
// 	return err
// }

/*
	检查这个交易hash是否在数据库中已经存在
*/
func (this *TxBase) CheckHashExist() bool {
	return db.LevelDB.CheckHashExist(this.Hash)
}

/*
	格式化成json字符串
*/
// func (this *TxBase) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	格式化成[]byte
*/
func (this *TxBase) Proto() (*[]byte, error) {
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

	txPay := go_protos.TxPay{
		TxBase: &txBase,
	}
	// txPay.Marshal()
	bs, err := txPay.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func (this *TxBase) CountTxHistory(height uint64) {
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
		// Payload:
	}
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
	//只有转账才保存备注信息
	if this.Class() == config.Wallet_tx_type_pay {
		hiOut.Payload = this.Payload
		hiIn.Payload = this.Payload
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
		hiOut.InAddr = append(hiOut.InAddr, &addrInfo.Addr)
	}

	//生成新的UTXO收益，保存到列表中
	addrCoin = make(map[string]bool)
	for _, vout := range this.Vout {
		hiOut.OutAddr = append(hiOut.OutAddr, &vout.Address)
		hiOut.Value += vout.Value
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
	if len(hiOut.InAddr) > 0 {
		balanceHistoryManager.Add(hiOut)
	}
	if len(hiIn.OutAddr) > 0 {
		balanceHistoryManager.Add(hiIn)
	}
}

/*
	将输出为0的vout删除
*/
func CleanZeroVouts(vs *[]*Vout) []*Vout {
	vouts := make([]*Vout, 0)
	for _, one := range *vs {
		if one.Value == 0 {
			continue
		}
		vouts = append(vouts, one)
	}
	return vouts
}

/*
	将输出为0的vout删除，并且合并相同地址的余额
*/
func MergeVouts(vs *[]*Vout) []*Vout {
	voutMap := make(map[string]*Vout)
	for i, one := range *vs {
		if one.Value == 0 {
			continue
		}
		v, ok := voutMap[utils.Bytes2string(one.Address)+strconv.Itoa(int(one.FrozenHeight))]
		if ok {
			v.Value = v.Value + one.Value
			continue
		}
		voutMap[utils.Bytes2string(one.Address)+strconv.Itoa(int(one.FrozenHeight))] = (*vs)[i]
	}
	vouts := make([]*Vout, 0)
	for _, v := range voutMap {
		vouts = append(vouts, v)
	}
	return vouts
}

/*
	把相同地址的交易输出合并在一起，删除余额为0的输出
*/
func (this *TxBase) MergeVout() {
	this.Vout = MergeVouts(this.GetVout())
	this.Vout_total = uint64(len(this.Vout))
}

/*
	删除余额为0的输出
*/
func (this *TxBase) CleanZeroVout() {
	this.Vout = CleanZeroVouts(this.GetVout())
	this.Vout_total = uint64(len(this.Vout))
}

/*
	检查是否有多次消费
	@return    bool    true=验证通过
*/
func (this *TxBase) MultipleExpenditures(txs ...TxItr) bool {
	m := make(map[string]interface{})
	for _, one := range txs {
		for _, two := range *one.GetVin() {
			// m[hex.EncodeToString(two.Txid)+"_"+strconv.Itoa(int(two.Vout))] = nil
			m[utils.Bytes2string(two.Txid)+"_"+strconv.Itoa(int(two.Vout))] = nil
		}
	}
	for _, one := range this.Vin {
		// key := hex.EncodeToString(one.Txid) + "_" + strconv.Itoa(int(one.Vout))
		key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.Vout))
		_, ok := m[key]
		if ok {
			return false
		} else {
			m[key] = nil
		}
	}
	return true
}

/*
	是否验证通过
*/
// func (this *TxBase) CheckRepeatedTx(txs ...TxItr) bool {
// 	return true
// }

/*
	解析交易
*/
// func ParseTxBases(txtype uint64, bs *[]byte) (TxItr, error) {

// 	// engine.Log.Info("解析交易 %d", txtype)

// 	// timeNow := time.Now()

// 	if txtype == 0 {
// 		bh := new(TxBase)

// 		// var jso = jsoniter.ConfigCompatibleWithStandardLibrary
// 		// err := json.Unmarshal(*bs, bh)

// 		decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 		decoder.UseNumber()
// 		err := decoder.Decode(bh)

// 		if err != nil {
// 			// fmt.Println("1111 这里错误", err, string(*bs))
// 			return nil, err
// 		}
// 		txtype = bh.Type

// 	}

// 	// engine.Log.Info("耗时 111 %s", time.Now().Sub(timeNow))

// 	var tx interface{}

// 	switch txtype {
// 	case config.Wallet_tx_type_mining: //挖矿所得
// 		tx = new(Tx_reward)
// 	case config.Wallet_tx_type_deposit_in: //投票参与挖矿输入，余额锁定
// 		tx = new(Tx_deposit_in)
// 	case config.Wallet_tx_type_deposit_out: //投票参与挖矿输出，余额解锁
// 		tx = new(Tx_deposit_out)
// 	case config.Wallet_tx_type_pay: //普通支付
// 		tx = new(Tx_Pay)
// 	case config.Wallet_tx_type_vote_in: //
// 		tx = new(Tx_vote_in)
// 	case config.Wallet_tx_type_vote_out: //
// 		tx = new(Tx_vote_out)
// 	default:
// 		tx = GetNewTransaction(txtype)
// 		if tx == nil {
// 			//未知交易类型
// 			return nil, errors.New("Unknown transaction type")
// 		}
// 	}

// 	// engine.Log.Info("耗时 222 %s", time.Now().Sub(timeNow))

// 	// err := json.Unmarshal(*bs, tx)

// 	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	decoder.UseNumber()
// 	err := decoder.Decode(tx)
// 	if err != nil {
// 		// fmt.Println("2222 这里错误", err)
// 		return nil, err
// 	}
// 	// engine.Log.Info("耗时 333 %s", time.Now().Sub(timeNow))
// 	return tx.(TxItr), nil
// }

/*
	解析交易
*/
func ParseTxBaseProto(txtype uint64, bs *[]byte) (TxItr, error) {
	if bs == nil {
		return nil, nil
	}
	// engine.Log.Info("解析交易 %d", txtype)

	// timeNow := time.Now()

	if txtype == 0 {
		// bh := new(TxBase)

		// var jso = jsoniter.ConfigCompatibleWithStandardLibrary
		// err := json.Unmarshal(*bs, bh)

		txPay := go_protos.TxPay{}
		err := proto.Unmarshal(*bs, &txPay)
		if err != nil {
			return nil, err
		}
		txtype = txPay.TxBase.Type

		// decoder := json.NewDecoder(bytes.NewBuffer(*bs))
		// decoder.UseNumber()
		// err := decoder.Decode(bh)

		// if err != nil {
		// 	// fmt.Println("1111 这里错误", err, string(*bs))
		// 	return nil, err
		// }
		// txtype = bh.Type

	}

	// engine.Log.Info("耗时 111 %s", time.Now().Sub(timeNow))

	var tx interface{}

	switch txtype {
	case config.Wallet_tx_type_mining: //挖矿所得

		txProto := new(go_protos.TxReward)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}

		if txProto.TxBase.Type != config.Wallet_tx_type_mining {
			return nil, errors.New("tx type error")
		}

		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
		for _, one := range txProto.TxBase.Vin {
			vins = append(vins, &Vin{
				Txid: one.Txid,
				Vout: one.Vout,
				Puk:  one.Puk,
				Sign: one.Sign,
			})
		}
		vouts := make([]*Vout, 0, len(txProto.TxBase.Vout))
		for _, one := range txProto.TxBase.Vout {
			vouts = append(vouts, &Vout{
				Value:        one.Value,
				Address:      one.Address,
				FrozenHeight: one.FrozenHeight,
			})
		}
		txBase := TxBase{}
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
		tx = &Tx_reward{
			TxBase: txBase,
		}

	case config.Wallet_tx_type_deposit_in: //投票参与挖矿输入，余额锁定

		txProto := new(go_protos.TxDepositIn)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
		for _, one := range txProto.TxBase.Vin {
			vins = append(vins, &Vin{
				Txid: one.Txid,
				Vout: one.Vout,
				Puk:  one.Puk,
				Sign: one.Sign,
			})
		}
		vouts := make([]*Vout, 0, len(txProto.TxBase.Vout))
		for _, one := range txProto.TxBase.Vout {
			vouts = append(vouts, &Vout{
				Value:        one.Value,
				Address:      one.Address,
				FrozenHeight: one.FrozenHeight,
			})
		}
		txBase := TxBase{}
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
		tx = &Tx_deposit_in{
			TxBase: txBase,
			Puk:    txProto.Puk,
		}

	case config.Wallet_tx_type_deposit_out: //投票参与挖矿输出，余额解锁
		// tx = new(Tx_deposit_out)
		txProto := new(go_protos.TxDepositOut)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
		for _, one := range txProto.TxBase.Vin {
			vins = append(vins, &Vin{
				Txid: one.Txid,
				Vout: one.Vout,
				Puk:  one.Puk,
				Sign: one.Sign,
			})
		}
		vouts := make([]*Vout, 0, len(txProto.TxBase.Vout))
		for _, one := range txProto.TxBase.Vout {
			vouts = append(vouts, &Vout{
				Value:        one.Value,
				Address:      one.Address,
				FrozenHeight: one.FrozenHeight,
			})
		}
		txBase := TxBase{}
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
		tx = &Tx_deposit_out{
			TxBase: txBase,
		}
	case config.Wallet_tx_type_pay: //普通支付
		txProto := new(go_protos.TxPay)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
		for _, one := range txProto.TxBase.Vin {
			vins = append(vins, &Vin{
				Txid: one.Txid,
				Vout: one.Vout,
				Puk:  one.Puk,
				Sign: one.Sign,
			})
		}
		vouts := make([]*Vout, 0, len(txProto.TxBase.Vout))
		for _, one := range txProto.TxBase.Vout {
			vouts = append(vouts, &Vout{
				Value:        one.Value,
				Address:      one.Address,
				FrozenHeight: one.FrozenHeight,
			})
		}
		txBase := TxBase{}
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
		tx = &Tx_Pay{
			TxBase: txBase,
		}
	case config.Wallet_tx_type_vote_in: //
		txProto := new(go_protos.TxVoteIn)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
		for _, one := range txProto.TxBase.Vin {
			vins = append(vins, &Vin{
				Txid: one.Txid,
				Vout: one.Vout,
				Puk:  one.Puk,
				Sign: one.Sign,
			})
		}
		vouts := make([]*Vout, 0, len(txProto.TxBase.Vout))
		for _, one := range txProto.TxBase.Vout {
			vouts = append(vouts, &Vout{
				Value:        one.Value,
				Address:      one.Address,
				FrozenHeight: one.FrozenHeight,
			})
		}
		txBase := TxBase{}
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
		tx = &Tx_vote_in{
			TxBase:   txBase,
			Vote:     txProto.Vote,
			VoteType: uint16(txProto.VoteType),
			VoteAddr: txProto.VoteAddr,
		}

	case config.Wallet_tx_type_vote_out: //
		txProto := new(go_protos.TxVoteOut)
		err := proto.Unmarshal(*bs, txProto)
		if err != nil {
			return nil, err
		}
		vins := make([]*Vin, 0, len(txProto.TxBase.Vin))
		for _, one := range txProto.TxBase.Vin {
			vins = append(vins, &Vin{
				Txid: one.Txid,
				Vout: one.Vout,
				Puk:  one.Puk,
				Sign: one.Sign,
			})
		}
		vouts := make([]*Vout, 0, len(txProto.TxBase.Vout))
		for _, one := range txProto.TxBase.Vout {
			vouts = append(vouts, &Vout{
				Value:        one.Value,
				Address:      one.Address,
				FrozenHeight: one.FrozenHeight,
			})
		}
		txBase := TxBase{}
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
		tx = &Tx_vote_out{
			TxBase: txBase,
		}
	default:
		tx = GetNewTransaction(txtype, bs)
		if tx == nil {
			//未知交易类型
			return nil, errors.New("Unknown transaction type")
		}
	}

	// engine.Log.Info("耗时 222 %s", time.Now().Sub(timeNow))

	// err := json.Unmarshal(*bs, tx)

	// decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	// decoder.UseNumber()
	// err := decoder.Decode(tx)
	// if err != nil {
	// 	// fmt.Println("2222 这里错误", err)
	// 	return nil, err
	// }
	// engine.Log.Info("耗时 333 %s", time.Now().Sub(timeNow))
	return tx.(TxItr), nil
}

/*
	挖矿所得收益输入
*/
//type Coinbase struct{
//	Coinbase :"",
//}

/*
	UTXO输入
*/
type Vin struct {
	Txid []byte `json:"txid"` //UTXO 前一个交易的id
	// TxidStr string `json:"-"`    //
	Vout uint64 `json:"vout"` //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	Puk  []byte `json:"puk"`  //公钥
	// PukStr       string             `json:"-"`    //缓存公钥字符串
	PukIsSelf int                `json:"-"` //缓存此公钥是否属于自己钱包。0=未检查;1=不属于自己;2=是属于自己;
	PukToAddr crypto.AddressCoin `json:"-"` //缓存此公钥的地址
	// PukToAddrStr string             `json:"-"`    //缓存此公钥的地址字符串
	Sign []byte `json:"sign"` //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	//	VoutSign []byte `json:"voutsign"` //对本交易的输出签名
}

type VinVO struct {
	Txid string `json:"txid"` //UTXO 前一个交易的id
	Vout uint64 `json:"vout"` //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	Puk  string `json:"puk"`  //公钥
	Sign string `json:"sign"` //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	//	VoutSign []byte `json:"voutsign"` //对本交易的输出签名
}

/*

 */
func (this *Vin) ConversionVO() *VinVO {
	return &VinVO{
		Txid: hex.EncodeToString(this.Txid), // this.GetTxidStr(),             //UTXO 前一个交易的id
		Vout: this.Vout,                     //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
		Puk:  hex.EncodeToString(this.Puk),  // this.GetPukStr(),              //公钥
		Sign: hex.EncodeToString(this.Sign), //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
}

/*
	获取交易id字符串
*/
// func (this *Vin) GetTxidStr() string {
// 	if this.TxidStr == "" {
// 		this.TxidStr = hex.EncodeToString(this.Txid)
// 	}
// 	return this.TxidStr
// }

/*
	获取公钥字符串
*/
// func (this *Vin) GetPukStr() string {
// 	if this.PukStr == "" {
// 		this.PukStr = hex.EncodeToString(this.Puk)
// 	}
// 	return this.PukStr
// }

/*
	检查公钥是否属于自己
*/
func (this *Vin) CheckIsSelf() bool {
	// return true
	// engine.Log.Info("判断地址 111 %d", this.PukIsSelf)
	if this.PukIsSelf == 0 {
		// engine.Log.Info("从新判断vin中地址是否属于自己 %d", this.PukIsSelf)
		_, ok := keystore.FindPuk(this.Puk)
		if ok {
			this.PukIsSelf = 2
			// this.PukToAddrStr = addrInfo.AddrStr
		} else {
			this.PukIsSelf = 1
		}
	}
	// engine.Log.Info("判断地址 222 %d", this.PukIsSelf)
	if this.PukIsSelf == 1 {
		return false
	} else {
		return true
	}
	// return this.PukIsSelf
}

/*
	获取公钥对应的地址字符串
*/
// func (this *Vin) GetPukToAddrStr() string {
// 	if this.PukToAddrStr == "" {
// 		addr := this.GetPukToAddr() // crypto.BuildAddr(config.AddrPre, this.Puk)
// 		this.PukToAddrStr = addr.B58String()
// 	}
// 	return this.PukToAddrStr
// }

/*
	获取公钥对应的地址
*/
func (this *Vin) GetPukToAddr() *crypto.AddressCoin {
	if this.PukToAddr == nil {
		addr := crypto.BuildAddr(config.AddrPre, this.Puk)
		this.PukToAddr = addr
	}
	return &this.PukToAddr
}

/*
	将需要签名的字段序列化
*/
func (this *Vin) SerializeVin() *[]byte {
	if this.Txid == nil {
		bs := make([]byte, 0, len(this.Puk)+len(this.Sign))
		bs = append(bs, this.Puk...)
		bs = append(bs, this.Sign...)
		return &bs
	} else {
		bs := make([]byte, 0, len(this.Txid)+8+len(this.Puk)+len(this.Sign))
		bs = append(bs, this.Txid...)
		bs = append(bs, utils.Uint64ToBytes(this.Vout)...)
		bs = append(bs, this.Puk...)
		bs = append(bs, this.Sign...)
		return &bs
	}

	// buf := bytes.NewBuffer(nil)
	// if this.Txid != nil {
	// 	buf.Write(this.Txid)
	// 	buf.Write(utils.Uint64ToBytes(this.Vout))
	// }
	// buf.Write(this.Puk)
	// buf.Write(this.Sign)
	// //	buf.Write(this.VoutSign)
	// bs := buf.Bytes()
	// return &bs
}

/*
	格式化成json字符串
*/
// func (this *Vin) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	验证地址是否属于自己
*/
func (this *Vin) ValidateAddr() (*crypto.AddressCoin, bool) {
	addr := crypto.BuildAddr(config.AddrPre, this.Puk)

	// addr, err := keystore.BuildAddrByPubkey(this.Puk)
	// if err != nil {
	// 	return nil, false
	// }
	//验证和自己相关的地址

	// fmt.Println("是否是自己的地址", addr)

	// if !keystore.FindAddress(addr) {
	// 	return &addr, false
	// }

	_, ok := keystore.FindAddress(addr)
	if !ok {
		return &addr, false
	}

	// validate := keystore.ValidateByAddress(addr.B58String())
	// if !validate.IsVerify || !validate.IsMine {
	// 	return nil, false
	// }
	return &addr, true
}

/*
	UTXO输出
*/
type Vout struct {
	Value        uint64             `json:"value"`         //输出金额 = 实际金额 * 100000000
	Address      crypto.AddressCoin `json:"address"`       //钱包地址
	FrozenHeight uint64             `json:"frozen_height"` //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
	AddrIsSelf   int                `json:"-"`             //缓存地址是否属于自己钱包。0=未检查;1=不属于自己;2=是属于自己;
	AddrStr      string             `json:"-"`             //缓存地址字符串
	// Txids        []byte             `json:"txid"`          //本输出被使用后的交易id
}

type VoutVO struct {
	Value        uint64 `json:"value"`         //输出金额 = 实际金额 * 100000000
	Address      string `json:"address"`       //钱包地址
	FrozenHeight uint64 `json:"frozen_height"` //冻结高度。在冻结高度以下，未花费的交易余额不能使用
	// Txids        string `json:"txid"`          //本输出被使用后的交易id
}

/*

 */
func (this *Vout) ConversionVO() *VoutVO {
	return &VoutVO{
		Value:        this.Value,               //输出金额 = 实际金额 * 100000000
		Address:      this.Address.B58String(), //钱包地址
		FrozenHeight: this.FrozenHeight,        //
		// Txids:        hex.EncodeToString(this.Txids), //本输出被使用后的交易id
	}
}

/*
	检查地址是否属于自己
*/
func (this *Vout) CheckIsSelf() bool {
	// return true
	if this.AddrIsSelf == 0 {
		_, ok := keystore.FindAddress(this.Address)
		if ok {
			this.AddrIsSelf = 2
		} else {
			this.AddrIsSelf = 1
		}
	}
	if this.AddrIsSelf == 1 {
		return false
	} else {
		return true
	}
}

/*
	获取地址字符串
*/
func (this *Vout) GetAddrStr() string {
	if this.AddrStr == "" {
		this.AddrStr = this.Address.B58String()
	}
	return this.AddrStr
}

/*
	将需要签名的字段序列化
*/
func (this *Vout) Serialize() *[]byte {
	bs := make([]byte, 0, len(this.Address)+8+8)
	bs = append(bs, utils.Uint64ToBytes(this.Value)...)
	bs = append(bs, this.Address...)
	bs = append(bs, utils.Uint64ToBytes(this.FrozenHeight)...)

	// buf := bytes.NewBuffer(nil)
	// buf.Write(utils.Uint64ToBytes(this.Value))
	// buf.Write(this.Address)
	// buf.Write(utils.Uint64ToBytes(this.FrozenHeight))
	// bs := buf.Bytes()
	return &bs
}

/*
	格式化成json字符串
*/
// func (this *Vout) Json() (*[]byte, error) {
// 	bs, err := json.Marshal(this)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &bs, err
// }

/*
	全网广播交易
*/
func MulticastTx(txItr TxItr) {
	//		engine.NLog.Debug(engine.LOG_console, "是超级节点发起投票")
	//		log.Println("是超级节点发起投票")
	utils.Go(func() {
		goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
		_, file, line, _ := runtime.Caller(0)
		engine.AddRuntime(file, line, goroutineId)
		defer engine.DelRuntime(file, line, goroutineId)
		// bs, err := txItr.Json()
		bs, err := txItr.Proto()
		if err != nil {
			// engine.Log.Warn("交易json格式化错误，取消广播 %s", txItr.GetHashStr())
			return
		}
		mc.SendMulticastMsg(config.MSGID_multicast_transaction, bs)
	})

}

/*
	查询数据库和解析交易
*/
// func FindTxBase(txid []byte) (TxItr, error) {
// 	var err error
// 	var txItr TxItr
// 	ok := false
// 	//是否启用缓存
// 	if config.EnableCache {
// 		//先判断缓存中是否存在
// 		// txItr, ok = TxCache.FindTxInCache(hex.EncodeToString(txid))
// 		txItr, ok = TxCache.FindTxInCache(txid)
// 	}
// 	if !ok {
// 		// engine.Log.Error("未命中缓存 FindTxInCache")
// 		var bs *[]byte
// 		bs, err = db.Find(txid)
// 		if err != nil {
// 			return nil, err
// 		}

// 		txItr, err = ParseTxBase(ParseTxClass(txid), bs)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return txItr, err

// 	// bs, err := db.Find(txid)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }
// 	// return ParseTxBase(ParseTxClass(txid), bs)
// }

/*
	通过交易hash解析交易类型
*/
func ParseTxClass(txid []byte) uint64 {
	// engine.Log.Info("交易id 111 %s", hex.EncodeToString(txid))
	classBs := txid[:8]
	// engine.Log.Info("交易id 222 %v", classBs)
	return utils.BytesToUint64(classBs)
}
