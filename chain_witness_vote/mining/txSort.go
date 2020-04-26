package mining

import (
	"math/big"
)

/*
	区块中的交易价值比
*/
type TransactionRatio struct {
	tx    TxItr    //交易
	size  uint64   //交易总大小
	gas   uint64   //手续费
	Ratio *big.Int //价值比
}

/*
	区块中的交易价值比排序
*/
type RatioSort struct {
	txRatio []TransactionRatio
}

func (this *RatioSort) Len() int {
	return len(this.txRatio)
}

func (this *RatioSort) Less(i, j int) bool {
	if this.txRatio[i].Ratio.Cmp(this.txRatio[j].Ratio) < 0 {
		return false
	} else {
		return true
	}
}

func (this *RatioSort) Swap(i, j int) {
	this.txRatio[i], this.txRatio[j] = this.txRatio[j], this.txRatio[i]
}
