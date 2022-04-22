package mining

import (
	"mandela/config"
)

/*
	冻结交易余额
*/
func (this *BalanceManager) Frozen(items []*TxItem, tx TxItr) {
	if config.Wallet_txitem_save_db {
		this.notspentBalanceDB.Frozen(items, tx)
	} else {
		this.notspentBalance.Frozen(items, tx)
	}
}

/*
	解冻回滚冻结的交易
*/
func (this *BalanceManager) Unfrozen(blockHeight uint64, blockTime int64) {
	if config.Wallet_txitem_save_db {
		//db存储方式不需要设置解冻
	} else {
		this.notspentBalance.Unfrozen(blockHeight, blockTime)
	}
}

/*
	删除冻结的交易
	@txid         []byte    TxItem中交易txid
	@voutIndex    uint64    TxItem中的vout
*/
// func (this *BalanceManager) DelFrozen(txid []byte, voutIndex uint64) {
// 	this.notspentBalance.DelFrozen(txid, voutIndex)
// }
