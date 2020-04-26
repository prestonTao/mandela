package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	mc "mandela/core/message_center"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"sync"

	"golang.org/x/crypto/ed25519"
)

const (
	BlockTx_Gas       = "gas"
	BlockTx_Hash      = "hash"
	BlockTx_Vout      = "vout"
	BlockTx_Vout_Tx   = "tx"
	BlockTx_Blockhash = "blockhash"
)

type TxItr interface {
	Class() uint64                                                                    //交易类型
	BuildHash()                                                                       //构建交易hash
	GetHash() *[]byte                                                                 //获得交易hash
	Check() error                                                                     //检查交易是否合法
	Json() (*[]byte, error)                                                           //将交易格式化成json字符串
	Serialize() *[]byte                                                               //将需要签名的字段序列化
	Balance() *sync.Map                                                               //查询交易输出，统计输出地址余额key:utils.Multihash=收款地址;value:TxItem=地址余额;
	GetVin() *[]Vin                                                                   //
	GetVout() *[]Vout                                                                 //
	GetGas() uint64                                                                   //
	SetTxid(bs *[]byte, index uint64, txid *[]byte) error                             //这个交易输出被使用之后，需要把UTXO输出标记下
	UnSetTxid(bs *[]byte, index uint64) error                                         //区块回滚，把之前标记为已经使用过的交易的标记去掉
	GetVoutSignSerialize(voutIndex uint64) *[]byte                                    //获取交易输出序列化
	GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte //获取签名
	SetBlockHash(bs []byte)                                                           //设置本交易所属的区块hash
	GetBlockHash() *[]byte                                                            //
	CheckHashExist() bool                                                             //检查hash在数据库中是否已经存在
	GetLockHeight() uint64                                                            //获取锁定区块高度
	SetSign(index uint64, bs []byte) bool                                             //修改签名
	SetPayload(bs []byte)                                                             //修改备注
	GetPayload() []byte                                                               //获取备注内容
	CheckRepeatedTx(txs ...TxItr) bool                                                //是否验证通过
	GetVOJSON() interface{}                                                           //用于地址和txid格式化显示
}

/*
	交易
*/
type TxBase struct {
	Hash       []byte `json:"hash"`        //本交易hash，不参与区块hash，只用来保存
	Type       uint64 `json:"type"`        //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
	Vin_total  uint64 `json:"vin_total"`   //输入交易数量
	Vin        []Vin  `json:"vin"`         //交易输入
	Vout_total uint64 `json:"vout_total"`  //输出交易数量
	Vout       []Vout `json:"vout"`        //交易输出
	Gas        uint64 `json:"gas"`         //交易手续费，此字段不参与交易hash
	LockHeight uint64 `json:"lock_height"` //本交易锁定在低于这个高度的块中，超过这个高度，块将不被打包到区块中。
	Payload    []byte `json:"payload"`     //备注信息
	BlockHash  []byte `json:"blockhash"`   //本交易属于的区块hash，不参与区块hash，只用来保存
	// CreateTime int64  `json:"lock_time"`   //创建时间,交易不需要创建时间，直接查询出块时间就OK
}

/*
	交易
*/
type TxBaseVO struct {
	Hash       string   `json:"hash"`        //本交易hash，不参与区块hash，只用来保存
	Type       uint64   `json:"type"`        //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
	Vin_total  uint64   `json:"vin_total"`   //输入交易数量
	Vin        []VinVO  `json:"vin"`         //交易输入
	Vout_total uint64   `json:"vout_total"`  //输出交易数量
	Vout       []VoutVO `json:"vout"`        //交易输出
	Gas        uint64   `json:"gas"`         //交易手续费，此字段不参与交易hash
	LockHeight uint64   `json:"lock_height"` //本交易锁定在低于这个高度的块中，超过这个高度，块将不被打包到区块中。
	Payload    string   `json:"payload"`     //备注信息
	BlockHash  string   `json:"blockhash"`   //本交易属于的区块hash，不参与区块hash，只用来保存
	// CreateTime int64    `json:"lock_time"`   //创建时间
}

/*
	转化为VO对象
*/
func (this *TxBase) ConversionVO() TxBaseVO {
	vins := make([]VinVO, 0)
	for _, one := range this.Vin {
		vins = append(vins, one.ConversionVO())
	}

	vouts := make([]VoutVO, 0)
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
		// CreateTime: this.CreateTime,                    //
	}
}

//修改签名
func (this *TxBase) SetSign(index uint64, bs []byte) bool {
	this.Vin[index].Sign = bs

	// for k, val := range this.Vin {
	// 	// index := append(val.Txid, utils.Int64ToBytes(int64(k))...)
	// 	if bytes.Equal(val.Txid, txid) && val.Vout == index {
	// 		this.Vin[k].Sign = bs
	// 	}
	// }
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
	buf := bytes.NewBuffer(nil)
	buf.Write(utils.Uint64ToBytes(this.Type))
	buf.Write(utils.Uint64ToBytes(this.Vin_total))
	if this.Vin != nil {
		for _, one := range this.Vin {
			bs := one.SerializeVin()
			buf.Write(*bs)
		}
	}
	buf.Write(utils.Uint64ToBytes(this.Vout_total))
	if this.Vout != nil {
		for _, one := range this.Vout {
			bs := one.Serialize()
			buf.Write(*bs)
		}
	}
	//	buf.Write(utils.Int64ToBytes(this.CreateTime))
	// buf.Write(utils.Uint64ToBytes(this.Gas))
	buf.Write(utils.Uint64ToBytes(this.LockHeight))
	buf.Write(this.Payload)
	// buf.Write(utils.Int64ToBytes(this.CreateTime))
	bs := buf.Bytes()
	return &bs
}

func (this *TxBase) GetVin() *[]Vin {
	return &this.Vin
}
func (this *TxBase) GetVout() *[]Vout {
	return &this.Vout
}

func (this *TxBase) GetGas() uint64 {
	return this.Gas
}

func (this *TxBase) GetHash() *[]byte {
	return &this.Hash
}

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

	// //上一个交易 所属的区块hash
	// buf := bytes.NewBuffer(this.BlockHash)
	// //上一个交易的hash
	// buf.Write(this.Hash)
	buf := bytes.NewBuffer(nil)

	//上一个交易的指定输出序列化
	buf.Write(utils.Uint64ToBytes(voutIndex))
	vout := this.Vout[voutIndex]
	bs := vout.Serialize()
	buf.Write(*bs)
	*bs = buf.Bytes()
	return bs
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
	buf := bytes.NewBuffer(nil)
	if voutBs != nil {
		buf.Write(*voutBs)
	}
	// buf := bytes.NewBuffer(*voutBs)
	buf.Write(utils.Uint64ToBytes(this.Type))
	buf.Write(utils.Uint64ToBytes(this.Vin_total))
	buf.Write(utils.Uint64ToBytes(vinIndex))
	buf.Write(utils.Uint64ToBytes(this.Vout_total))
	bs := make([]byte, 0)
	for _, one := range this.Vout {
		bs = append(bs, *one.Serialize()...)
	}
	buf.Write(bs)
	buf.Write(utils.Uint64ToBytes(this.LockHeight))
	buf.Write(this.Payload)
	// buf.Write(utils.Int64ToBytes(this.CreateTime))
	bs = buf.Bytes()
	return &bs
}

/*
	获取签名
*/
func (this *TxBase) GetSign(key *ed25519.PrivateKey, txid []byte, voutIndex, vinIndex uint64) *[]byte {

	bs, err := db.Find(txid)
	if err != nil {
		return nil
	}
	txItr, err := ParseTxBase(bs)
	if err != nil {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	//上一个交易 所属的区块hash
	buf.Write(*txItr.GetBlockHash())
	//上一个交易的hash
	buf.Write(*txItr.GetHash())
	//上一个交易的指定输出序列化
	buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	//本交易类型输入输出数量等信息和所有输出
	signBs := buf.Bytes()
	signDst := this.GetSignSerialize(&signBs, vinIndex)

	// fmt.Println("签名前的字节", len(*bs), hex.EncodeToString(*bs), "\n")
	sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return &sign
}

/*
	获取待签名数据
*/
func (this *TxBase) GetWaitSign(txid []byte, voutIndex, vinIndex uint64) *[]byte {
	bs, err := db.Find(txid)
	if err != nil {
		return nil
	}
	txItr, err := ParseTxBase(bs)
	if err != nil {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	//上一个交易 所属的区块hash
	buf.Write(*txItr.GetBlockHash())
	//上一个交易的hash
	buf.Write(*txItr.GetHash())
	//上一个交易的指定输出序列化
	buf.Write(*txItr.GetVoutSignSerialize(voutIndex))
	//本交易类型输入输出数量等信息和所有输出
	signBs := buf.Bytes()
	signDst := this.GetSignSerialize(&signBs, vinIndex)

	//fmt.Println("预签名字节", len(*bs), hex.EncodeToString(*bs), "\n")
	//sign := keystore.Sign(*key, *signDst)

	// fmt.Println("签名字符", len(sign), hex.EncodeToString(sign))

	return signDst
}

/*
	检查交易是否合法
*/
func (this *TxBase) CheckBase() error {
	// fmt.Println("开始验证交易合法性 Tx_deposit_in")

	//不能出现余额为0的转账
	for i, one := range this.Vout {
		if i != 0 && one.Value <= 0 {
			return config.ERROR_amount_zero
		}
	}

	//1.检查输入签名是否正确，2.检查输入输出是否对等，还有手续费
	inTotal := uint64(0)
	for i, one := range this.Vin {
		txbs, err := db.Find(one.Txid)
		if err != nil {
			return config.ERROR_tx_not_exist
		}
		txItr, err := ParseTxBase(txbs)
		if err != nil {
			return config.ERROR_tx_format_fail
		}
		vout := (*txItr.GetVout())[one.Vout]
		//如果这个交易已经被使用，则验证不通过，否则会出现双花问题。
		if vout.Tx != nil {
			engine.Log.Debug("There were two spending mistakes in the deal %s %s %d", hex.EncodeToString(vout.Tx), hex.EncodeToString(one.Txid), one.Vout)
			return config.ERROR_tx_is_use
		}
		inTotal = inTotal + vout.Value

		//验证公钥是否和地址对应
		addr := crypto.BuildAddr(config.AddrPre, one.Puk)
		if !bytes.Equal(addr, (*txItr.GetVout())[one.Vout].Address) {
			return config.ERROR_sign_fail
		}

		//验证签名
		buf := bytes.NewBuffer(nil)
		//上一个交易 所属的区块hash
		buf.Write(*txItr.GetBlockHash())
		//上一个交易的hash
		buf.Write(*txItr.GetHash())
		//上一个交易的指定输出序列化
		buf.Write(*txItr.GetVoutSignSerialize(one.Vout))
		//本交易类型输入输出数量等信息和所有输出
		bs := buf.Bytes()
		sign := this.GetSignSerialize(&bs, uint64(i))

		puk := ed25519.PublicKey(one.Puk)
		// fmt.Printf("txid:%x puk:%x sign:%x", md5.Sum(one.Txid), md5.Sum(one.Puk), md5.Sum(one.Sign))
		if !ed25519.Verify(puk, *sign, one.Sign) {
			return config.ERROR_sign_fail
		}

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

func (this *TxBase) Balance() *sync.Map {
	result := new(sync.Map)
	for i, one := range this.Vout {
		if one.Tx != nil {
			continue
		}
		item := TxItem{
			Addr:     &one.Address,
			Value:    one.Value, //余额
			Txid:     this.Hash, //交易id
			OutIndex: uint64(i), //交易输出index，从0开始
		}
		result.Store(one.Address.B58String(), item)
	}
	return result
}

/*
	这个交易输出被使用之后，需要把UTXO输出标记下
	注意：本方法只会保存
*/
func (this *TxBase) SetTxid(bs *[]byte, index uint64, txid *[]byte) error {
	txMap := make(map[string]interface{})
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(&txMap)
	// err := json.Unmarshal(*bs, &txMap)
	if err != nil {
		return err
	}
	v := txMap["vout"]
	if v == nil {
		//解析失败
		return errors.New("Resolution failure")
	}
	vs := v.([]interface{})
	vouts := make([]Vout, 0)
	for _, one := range vs {
		voutBs, err := json.Marshal(one)
		if err != nil {
			return err
		}
		vout := new(Vout)

		decoder := json.NewDecoder(bytes.NewBuffer(voutBs))
		decoder.UseNumber()
		err = decoder.Decode(vout)

		// err = json.Unmarshal(voutBs, vout)
		if err != nil {
			return err
		}
		vouts = append(vouts, *vout)
	}

	vouts[index].Tx = *txid
	txMap["vout"] = vouts

	txbs, err := json.Marshal(txMap)
	if err != nil {
		return err
	}
	err = db.Save(this.Hash, &txbs)
	return err
}

/*
	区块回滚，把之前标记为已经使用过的交易的标记去掉
*/
func (this *TxBase) UnSetTxid(bs *[]byte, index uint64) error {
	txMap := make(map[string]interface{})
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(&txMap)
	// err := json.Unmarshal(*bs, &txMap)
	if err != nil {
		return err
	}
	v := txMap["vout"]
	if v == nil {
		//解析失败
		return errors.New("Resolution failure")
	}
	vs := v.([]interface{})
	vouts := make([]Vout, 0)
	for _, one := range vs {
		voutBs, err := json.Marshal(one)
		if err != nil {
			return err
		}
		vout := new(Vout)

		decoder := json.NewDecoder(bytes.NewBuffer(voutBs))
		decoder.UseNumber()
		err = decoder.Decode(vout)

		// err = json.Unmarshal(voutBs, vout)
		if err != nil {
			return err
		}
		vouts = append(vouts, *vout)
	}

	vouts[index].Tx = nil
	txMap["vout"] = vouts

	txbs, err := json.Marshal(txMap)
	if err != nil {
		return err
	}
	err = db.Save(this.Hash, &txbs)
	return err
}

/*
	检查这个交易hash是否在数据库中已经存在
*/
func (this *TxBase) CheckHashExist() bool {
	return db.CheckHashExist(this.Hash)
}

/*
	格式化成json字符串
*/
func (this *TxBase) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

func MergeVouts(vs *[]Vout) []Vout {
	voutMap := make(map[string]*Vout)
	for i, one := range *vs {
		if one.Value == 0 {
			continue
		}
		v, ok := voutMap[one.Address.B58String()]
		if ok {
			v.Value = v.Value + one.Value
			continue
		}
		voutMap[one.Address.B58String()] = &(*vs)[i]
	}
	vouts := make([]Vout, 0)
	for _, v := range voutMap {
		vouts = append(vouts, *v)
	}
	return vouts
}

/*
	把相同地址的交易输出合并在一起
*/
func (this *TxBase) MergeVout() {
	// voutMap := make(map[string]*Vout)
	// for i, one := range *this.GetVout() {
	// 	if one.Value == 0 {
	// 		continue
	// 	}
	// 	v, ok := voutMap[one.Address.B58String()]
	// 	if ok {
	// 		v.Value = v.Value + one.Value
	// 		continue
	// 	}
	// 	voutMap[one.Address.B58String()] = &(*this.GetVout())[i]
	// }
	// vouts := make([]Vout, 0)
	// for _, v := range voutMap {
	// 	vouts = append(vouts, *v)
	// }
	this.Vout = MergeVouts(this.GetVout())
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
			m[hex.EncodeToString(two.Txid)+"_"+strconv.Itoa(int(two.Vout))] = nil
		}
	}
	for _, one := range this.Vin {
		key := hex.EncodeToString(one.Txid) + "_" + strconv.Itoa(int(one.Vout))
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
func ParseTxBase(bs *[]byte) (TxItr, error) {

	bh := new(TxBase)
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(bh)

	// err := json.Unmarshal(*bs, bh)
	if err != nil {
		// fmt.Println("1111 这里错误", err, string(*bs))
		return nil, err
	}

	var tx interface{}

	switch bh.Type {
	case config.Wallet_tx_type_mining: //挖矿所得
		tx = new(Tx_reward)
	case config.Wallet_tx_type_deposit_in: //投票参与挖矿输入，余额锁定
		tx = new(Tx_deposit_in)
	case config.Wallet_tx_type_deposit_out: //投票参与挖矿输出，余额解锁
		tx = new(Tx_deposit_out)
	case config.Wallet_tx_type_pay: //普通支付
		tx = new(Tx_Pay)
	case config.Wallet_tx_type_vote_in: //
		tx = new(Tx_vote_in)
	case config.Wallet_tx_type_vote_out: //
		tx = new(Tx_vote_out)
	default:
		tx = GetNewTransaction(bh.Type)
		if tx == nil {
			//未知交易类型
			return nil, errors.New("Unknown transaction type")
		}
	}

	decoder = json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err = decoder.Decode(tx)
	// err = json.Unmarshal(*bs, tx)
	if err != nil {
		// fmt.Println("2222 这里错误", err)
		return nil, err
	}
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
	Vout uint64 `json:"vout"` //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	Puk  []byte `json:"puk"`  //公钥
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
func (this *Vin) ConversionVO() VinVO {
	return VinVO{
		Txid: hex.EncodeToString(this.Txid), //UTXO 前一个交易的id
		Vout: this.Vout,                     //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
		Puk:  hex.EncodeToString(this.Puk),  //公钥
		Sign: hex.EncodeToString(this.Sign), //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
}

/*
	将需要签名的字段序列化
*/
func (this *Vin) SerializeVin() *[]byte {
	buf := bytes.NewBuffer(nil)
	if this.Txid != nil {
		buf.Write(this.Txid)
		buf.Write(utils.Uint64ToBytes(this.Vout))
	}
	buf.Write(this.Puk)
	buf.Write(this.Sign)
	//	buf.Write(this.VoutSign)
	bs := buf.Bytes()
	return &bs
}

/*
	格式化成json字符串
*/
func (this *Vin) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

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

	if !keystore.FindAddress(addr) {
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
	Value   uint64             `json:"value"`   //输出金额 = 实际金额 * 100000000
	Address crypto.AddressCoin `json:"address"` //钱包地址
	Tx      []byte             `json:"tx"`      //本输出被使用后的交易id
}

type VoutVO struct {
	Value   uint64 `json:"value"`   //输出金额 = 实际金额 * 100000000
	Address string `json:"address"` //钱包地址
	Tx      string `json:"tx"`      //本输出被使用后的交易id
}

/*

 */
func (this *Vout) ConversionVO() VoutVO {
	return VoutVO{
		Value:   this.Value,                  //输出金额 = 实际金额 * 100000000
		Address: this.Address.B58String(),    //钱包地址
		Tx:      hex.EncodeToString(this.Tx), //本输出被使用后的交易id
	}
}

/*
	将需要签名的字段序列化
*/
func (this *Vout) Serialize() *[]byte {
	buf := bytes.NewBuffer(nil)
	buf.Write(utils.Uint64ToBytes(this.Value))
	buf.Write(this.Address)
	bs := buf.Bytes()
	return &bs
}

/*
	格式化成json字符串
*/
func (this *Vout) Json() (*[]byte, error) {
	bs, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}
	return &bs, err
}

/*
	全网广播交易
*/
func MulticastTx(bs *[]byte) {
	//		engine.NLog.Debug(engine.LOG_console, "是超级节点发起投票")
	//		log.Println("是超级节点发起投票")

	go mc.SendMulticastMsg(config.MSGID_multicast_transaction, bs)

	// head := mc.NewMessageHead(nil, nil, false)
	// body := mc.NewMessageBody(bs, "", nil, 0)
	// message := mc.NewMessage(head, body)
	// message.BuildHash()

	// //广播给其他节点
	// //		ids := nodeStore.GetIdsForFar(message.Content)
	// for _, one := range nodeStore.GetAllNodes() {
	// 	//		log.Println("发送给", one.B58String())
	// 	if ss, ok := engine.GetSession(one.B58String()); ok {
	// 		ss.Send(config.MSGID_multicast_transaction, head.JSON(), body.JSON(), false)
	// 	} else {
	// 		engine.NLog.Debug(engine.LOG_console, "发送消息失败")
	// 	}
	// }
}

/*
	查询数据库和解析交易
*/
func FindTxBase(txid []byte) (TxItr, error) {
	bs, err := db.Find(txid)
	if err != nil {
		return nil, err
	}
	return ParseTxBase(bs)
}

/*
	通过交易hash解析交易类型
*/
func ParseTxClass(txid []byte) uint64 {
	// fmt.Println("解析交易id，查询交易类型", len(txid))
	// fmt.Println("/TODO 交易id长度判断", len(txid))
	classBs := txid[:8]
	return utils.BytesToUint64(classBs)
}
