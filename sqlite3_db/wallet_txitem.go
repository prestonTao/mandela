package sqlite3_db

import (
	"time"

	_ "github.com/go-xorm/xorm"
)

// var txitemLock = new(sync.Mutex)

type TxItem struct {
	Id           int64     `xorm:"pk autoincr unique 'id'"` //id
	Key          []byte    `xorm:"Blob 'key'"`              //txitem key
	Addr         []byte    `xorm:"Blob 'addr'"`             //收款地址
	Value        uint64    `xorm:"uint64 'value'"`          //余额
	Txid         []byte    `xorm:"Blob 'txid'"`             //交易id
	VoutIndex    uint64    `xorm:"uint64 'voutindex'"`      //交易输出index，从0开始
	Height       uint64    `xorm:"uint64 'height'"`         //区块高度，排序用
	VoteType     uint16    `xorm:"uint64 'votetype'"`       //投票类型
	FrozenHeight uint64    `xorm:"uint64 'frozenheight'"`   //锁仓高度/锁仓时间
	LockupHeight uint64    `xorm:"uint64 'lockupheight'"`   //锁仓高度/锁仓时间
	Status       int32     `xorm:"int32 'status'"`          //状态
	CreateTime   time.Time `xorm:"created 'createtime'"`    //创建时间，这个Field将在Insert时自动赋值为当前时间
}

/*
	添加一个未花费的交易记录
	@return    int64    数据库id
*/
// func (this *TxItem) Add(txitem *TxItem) error {
// 	_, err := table_wallet_txitem.Insert(txitem)
// 	return err
// }

/*
	添加一个未花费的交易记录
	@return    int64    数据库id
*/
func (this *TxItem) AddTxItems(txitems *[]TxItem) error {
	if txitems == nil || len(*txitems) <= 0 {
		return nil
	}
	lenght := len(*txitems)
	pageOne := 100
	var err error

	// txitemLock.Lock()
	for i := 0; i < lenght/pageOne; i++ {
		items := (*txitems)[i*pageOne : (i+1)*pageOne]
		_, err = engineDB.Insert(&items)
		if err != nil {
			// txitemLock.Unlock()
			return err
		}
	}
	if lenght%pageOne > 0 {
		i := lenght / pageOne
		items := (*txitems)[i*pageOne : lenght]
		_, err = engineDB.Insert(&items)
		if err != nil {
			// txitemLock.Unlock()
			return err
		}
	}
	// txitemLock.Unlock()
	return nil
}

/*
	删除多个已经使用了的余额记录
*/
func (this *TxItem) RemoveMoreKey(keys [][]byte) error {
	if keys == nil || len(keys) <= 0 {
		return nil
	}
	// txitemLock.Lock()
	_, err := engineDB.In("key = ?", keys).Unscoped().Delete(this)
	// txitemLock.Unlock()
	return err
}

/*
	设置冻结高度
*/
// func (this *TxItem) FrozenKeys(keys [][]byte, lockTime uint64) error {
// 	txitem := TxItem{LockupHeight: lockTime}
// 	_, err := table_wallet_txitem.In("key", keys).Update(&txitem)
// 	return err
// }

/*
	冻结部分txitem
*/
func (this *TxItem) UpdateLockHeight(ids *[]int64, lockTime uint64) error {
	if ids == nil || len(*ids) <= 0 {
		return nil
	}
	txitem := TxItem{LockupHeight: lockTime}
	// txitemLock.Lock()
	_, err := engineDB.In("id", *ids).Update(&txitem)
	// txitemLock.Unlock()
	return err
}

/*
	对未使用的余额求和
*/
func (this *TxItem) FindNotSpentSum(height uint64) (uint64, error) {
	ss := new(TxItem)
	// txitemLock.Lock()
	total, err := engineDB.Where("frozenheight < ? and lockupheight < ?", height, height).SumInt(ss, "value")
	// txitemLock.Unlock()
	if err != nil {
		return 0, err
	}
	return uint64(total), nil
}

/*
	对冻结余额求和
*/
func (this *TxItem) FindLockHeightSum(lockTime uint64) (uint64, error) {
	ss := new(TxItem)
	// txitemLock.Lock()
	total, err := engineDB.Where("lockupheight >?", lockTime).SumInt(ss, "value")
	// txitemLock.Unlock()
	if err != nil {
		return 0, err
	}
	return uint64(total), nil
}

/*
	对锁仓余额求和
*/
func (this *TxItem) FindFrozenHeightSum(frozenheight uint64) (uint64, error) {
	ss := new(TxItem)
	// txitemLock.Lock()
	total, err := engineDB.Where("frozenheight >?", frozenheight).SumInt(ss, "value")
	// txitemLock.Unlock()
	if err != nil {
		return 0, err
	}
	return uint64(total), nil
}

/*
	获取未使用的余额
*/
func (this *TxItem) FindNotSpentTxitem(height uint64, amount uint64, limit int) (*[]TxItem, error) {
	tis := make([]TxItem, 0)
	// txitemLock.Lock()
	err := engineDB.Where("frozenheight < ? and lockupheight < ? and value >= ?", height, height, amount).Limit(1, 0).Find(&tis)
	// txitemLock.Unlock()
	if err != nil {
		return &tis, err
	}
	if len(tis) > 0 {
		return &tis, nil
	}
	// txitemLock.Lock()
	err = engineDB.Where("frozenheight < ? and lockupheight < ? and value < ?", height, height, amount).Limit(limit, 0).Find(&tis)
	// txitemLock.Unlock()
	if err != nil {
		return &tis, err
	}
	return &tis, nil
}

/*
	获取未使用的余额
*/
func (this *TxItem) FindNotSpentTxitemByAddr(addr []byte, height uint64, amount uint64, limit int) (*[]TxItem, error) {
	tis := make([]TxItem, 0)
	// txitemLock.Lock()
	err := engineDB.Where("addr = ? and frozenheight < ? and lockupheight < ? and value >= ?", addr, height, height, amount).Limit(1, 0).Find(&tis)
	// txitemLock.Unlock()
	if err != nil {
		return &tis, err
	}
	if len(tis) > 0 {
		return &tis, nil
	}
	// txitemLock.Lock()
	err = engineDB.Where("addr = ? and frozenheight < ? and lockupheight < ? and value < ?", addr, height, height, amount).Limit(limit, 0).Find(&tis)
	// txitemLock.Unlock()
	if err != nil {
		return &tis, err
	}
	return &tis, nil
}

/*
	根据时间将两种冻结状态的数据解冻
*/
// func (this *TxItem) UnfrozenForTime(frozen_status, lock_status, notspent_status int32, lockTime uint64) error {
// 	txitem := TxItem{Status: notspent_status}
// 	_, err := table_wallet_txitem.Where("(status = ? or status = ?) and lockupheight <= ?",
// 		frozen_status, lock_status, lockTime).Update(&txitem)
// 	return err
// }

/*
	根据高度将两种冻结状态的数据解冻
*/
// func (this *TxItem) UnfrozenForHeight(frozen_status, lock_status, notspent_status int32, lockHeight uint64) error {
// 	txitem := TxItem{Status: notspent_status}
// 	_, err := table_wallet_txitem.Where("(status = ? or status = ?) and lockupheight <= ?",
// 		frozen_status, lock_status, lockHeight).Update(&txitem)
// 	return err
// }

/*
	删除多个消息记录
*/
// func (this *TxItem) RemoveOne(txid, addr []byte, voutIndex uint64) error {
// 	_, err := table_wallet_txitem.Where("txid = ? and voutindex = ? and addr = ?",
// 		txid, voutIndex, addr).Unscoped().Delete(this)
// 	return err
// }

/*
	查询
*/
// func (this *TxItem) FindStateByAddrs(status int32, addrs [][]byte) ([]*TxItem, error) {
// 	res := []*TxItem{}
// 	err := engineDB.Where("status = ?", status).In("addr", addrs).Find(&res)
// 	return res, err
// }

// func (this *TxItem) FindState(status int32) ([]*TxItem, error) {
// 	res := []*TxItem{}
// 	err := engineDB.Where("status = ?", status).Find(&res)
// 	return res, err
// }
