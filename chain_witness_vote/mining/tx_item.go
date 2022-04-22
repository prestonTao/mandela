package mining

import (
	"mandela/core/utils/crypto"
	"sync"
)

const (
	txItem_status_notSpent = int32(0) //未花费的交易余额，可以正常支付
	txItem_status_frozen   = int32(1) //锁仓,区块达到指定高度才能使用
	txItem_status_lock     = int32(2) //冻结高度，指定高度还未上链，则转为未花费的交易
)

/*
	保存一个地址的余额列表
	一个地址余额等于多个交易输出相加
*/
type Balance struct {
	Addr *crypto.AddressCoin //
	Txs  *sync.Map           //key:string=交易id;value:*TxItem=交易详细
}

/*
	交易列表
*/
type TxItem struct {
	Id           int64               //
	Addr         *crypto.AddressCoin //收款地址
	Value        uint64              //余额
	Txid         []byte              //交易id
	VoutIndex    uint64              //交易输出index，从0开始
	Height       uint64              //区块高度，排序用
	VoteType     uint16              //投票类型
	LockupHeight uint64              //锁仓高度/锁仓时间
	Status       int32               //状态
	// WitnessAddr *crypto.AddressCoin //给谁投票的见证人地址
}

// func (this *TxItem) GetAddrStr() string {
// 	if this.AddrStr == "" {
// 		this.AddrStr = this.Addr.B58String()
// 	}
// 	return this.AddrStr
// }

// func (this *TxItem) GetTxidStr() string {
// 	if this.TxidStr == "" {
// 		this.TxidStr = hex.EncodeToString(this.Txid)
// 	}
// 	return this.TxidStr
// }

/*
	清空对象中的变量
*/
func (this *TxItem) Clean() {
	this.Addr = nil
	// this.AddrStr = ""
	this.Value = 0
	this.Txid = nil
	// this.TxidStr = ""
	this.VoutIndex = 0
	this.Height = 0
	this.VoteType = 0
}

type TxItemSort []*TxItem

func (this *TxItemSort) Len() int {
	return len(*this)
}

/*
	value值大的排在前面
*/
func (this *TxItemSort) Less(i, j int) bool {
	if (*this)[i].Value < (*this)[j].Value {
		return false
	} else {
		return true
	}
}

func (this *TxItemSort) Swap(i, j int) {
	(*this)[i], (*this)[j] = (*this)[j], (*this)[i]
}

/*
	统计交易输入中要处理的参数
*/
type TxItemCount struct {
	Additems []*TxItem     //
	SubItems []*TxSubItems //
	// deleteKey []string      //要删除的冻结交易
}

type TxSubItems struct {
	Txid      []byte
	VoutIndex uint64
	Addr      crypto.AddressCoin
}
