package mining

import (
	"mandela/core/utils/crypto"
	"sync"
)

type TxController interface {
	Factory() interface{}                                                                 //
	CountBalance(balance *sync.Map, deposit *sync.Map, bhvo *BlockHeadVO, txIndex uint64) //统计余额断
	RollbackBalance()                                                                     //
	//前一个交易
	//
	BuildTx(balance *sync.Map, deposit *sync.Map, addr *crypto.AddressCoin, amount,
		gas uint64, pwd string, params ...interface{}) (TxItr, error) //
	//Check(txItr TxItr) bool //
}

//管理其他
var txCtrlMap = new(sync.Map) //key:uint64=交易类型;value:TxController=交易控制接口;

/*
	注册一个交易
*/
func RegisterTransaction(class uint64, txCtrl TxController) {
	txCtrlMap.Store(class, txCtrl)
}

/*
	获得一个交易的控制器
*/
func GetTransactionCtrl(class uint64) TxController {
	itr, ok := txCtrlMap.Load(class)
	if ok {
		txCtrl := itr.(TxController)
		return txCtrl
	} else {
		return nil
	}
}

/*
	获得一个新的交易对象
*/
func GetNewTransaction(class uint64) interface{} {
	itr, ok := txCtrlMap.Load(class)
	if ok {
		txCtrl := itr.(TxController)
		return txCtrl.Factory()
	} else {
		return nil
	}
}
