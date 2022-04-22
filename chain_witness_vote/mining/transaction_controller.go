package mining

import (
	"mandela/core/utils/crypto"
	"sync"
)

type TxController interface {
	Factory() interface{}                                                      //
	CountBalance(balance *TxItemManager, deposit *sync.Map, bhvo *BlockHeadVO) //同步统计余额
	SyncCount()                                                                //异步统计
	RollbackBalance()                                                          //
	//前一个交易
	//
	BuildTx(balance *TxItemManager, deposit *sync.Map, srcAddr, addr *crypto.AddressCoin, amount,
		gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) //
	CheckMultiplePayments(txItr TxItr) error //检查多次消费，排除token双重支付。
	//Check(txItr TxItr) bool //
	ParseProto(bs *[]byte) (interface{}, error) //
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
// func GetNewTransaction(class uint64) interface{} {
// 	itr, ok := txCtrlMap.Load(class)
// 	if ok {
// 		txCtrl := itr.(TxController)
// 		return txCtrl.Factory()
// 	} else {
// 		return nil
// 	}
// }

/*
	获得一个新的交易对象
*/
func GetNewTransaction(class uint64, bs *[]byte) interface{} {
	itr, ok := txCtrlMap.Load(class)
	if ok {
		txCtrl := itr.(TxController)
		if bs == nil {
			return txCtrl.Factory()
		}
		if bs == nil || len(*bs) <= 0 {
			return txCtrl
		}
		tx, err := txCtrl.ParseProto(bs)
		if err != nil {
			return nil
		}
		return tx
	} else {
		return nil
	}
}

/*
	统计其他类型交易
*/
func CountBalanceOther(balance *TxItemManager, deposit *sync.Map, bhvo *BlockHeadVO) {
	txCtrlMap.Range(func(k, v interface{}) bool {
		txCtrl := v.(TxController)
		txCtrl.CountBalance(balance, deposit, bhvo)
		return true
	})
}
