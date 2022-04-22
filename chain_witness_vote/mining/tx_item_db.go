package mining

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/sqlite3_db"
	"sync"
)

type TxItemManagerDB struct {
	lock *sync.RWMutex //
	// NotSpentBalance map[string]map[string]*TxItem //保存各个状态的txitem，解锁、冻结等只是状态的改变。key:string=收款地址;value:*sync.Map(key:string=[txid]_[voutIndex];value:*TxItem=TxItem;)=;
	notSpentValue uint64 //未花费的余额
	lockValue     uint64 //等待上链，临时冻结的余额
	frozenValue   uint64 //锁仓，等待解锁的余额
}

/*
	批量添加和删除TxItem
*/
func (this *TxItemManagerDB) CountTxItem(itemCount TxItemCount, blockHeight uint64, blockTime int64) {
	// start := time.Now()
	// this.lock.Lock()

	// this.NotSpentBalanceFotAddrKeyLock.Lock()
	//删除余额
	keys := make([][]byte, 0)
	for _, one := range itemCount.SubItems {
		key := append(utils.Hash_SHA3_256(one.Txid), utils.Uint64ToBytes(one.VoutIndex)...)
		keys = append(keys, key)
		// new(sqlite3_db.TxItem).RemoveOne(one.Txid, one.Addr, one.VoutIndex)
	}
	new(sqlite3_db.TxItem).RemoveMoreKey(keys)
	//添加余额
	txItemDBs := make([]sqlite3_db.TxItem, 0)
	for _, item := range itemCount.Additems {
		if item.Value == 0 {
			continue
		}
		//根据冻结高度判断余额是否可用
		// lockHeight := atomic.LoadUint64(&item.LockupHeight)
		// if lockHeight > config.Wallet_frozen_time_min {
		// 	//按时间锁仓
		// 	if int64(lockHeight) > blockTime {
		// 		atomic.StoreInt32(&item.Status, txItem_status_frozen)
		// 	} else {
		// 		atomic.StoreInt32(&item.Status, txItem_status_notSpent)
		// 	}
		// } else {
		// 	//按块高度锁仓
		// 	if lockHeight > blockHeight {
		// 		atomic.StoreInt32(&item.Status, txItem_status_frozen)
		// 	} else {
		// 		atomic.StoreInt32(&item.Status, txItem_status_notSpent)
		// 	}
		// }
		key := append(utils.Hash_SHA3_256(item.Txid), utils.Uint64ToBytes(item.VoutIndex)...)

		// addr := crypto.AddressCoin(item.Addr)
		txItemDB := sqlite3_db.TxItem{
			Key:          key,               //
			Addr:         *item.Addr,        //收款地址
			Value:        item.Value,        //余额
			Txid:         item.Txid,         //交易id
			VoutIndex:    item.VoutIndex,    //交易输出index，从0开始
			Height:       item.Height,       //区块高度，排序用
			VoteType:     item.VoteType,     //投票类型
			FrozenHeight: item.LockupHeight, //锁仓高度/锁仓时间
			// LockupHeight: item.LockupHeight, //锁仓高度/锁仓时间
			// Status:       item.Status,       //状态
		}
		txItemDBs = append(txItemDBs, txItemDB)

	}
	new(sqlite3_db.TxItem).AddTxItems(&txItemDBs)
	// this.lock.Unlock()
	// engine.Log.Info("CountTxItem spend time: %s", time.Now().Sub(start))
}

/*
	添加一个TxItem
*/
// func (this *TxItemManagerDB) AddTxItem(item TxItem) {
// 	if item.Value == 0 {
// 		return
// 	}
// 	atomic.StoreInt32(&item.Status, txItem_status_notSpent)
// 	txItemDB := sqlite3_db.TxItem{
// 		Addr:         *item.Addr,        //收款地址
// 		Value:        item.Value,        //余额
// 		Txid:         item.Txid,         //交易id
// 		VoutIndex:    item.VoutIndex,    //交易输出index，从0开始
// 		Height:       item.Height,       //区块高度，排序用
// 		VoteType:     item.VoteType,     //投票类型
// 		LockupHeight: item.LockupHeight, //锁仓高度/锁仓时间
// 		Status:       item.Status,       //状态
// 	}

// 	new(sqlite3_db.TxItem).Add(&txItemDB)
// 	// this.lock.Lock()
// 	// atomic.StoreInt32(&item.Status, txItem_status_notSpent)
// 	// key := utils.Bytes2string(item.Txid) + "_" + strconv.Itoa(int(item.VoutIndex))

// 	// items, ok := this.NotSpentBalance[utils.Bytes2string(*item.Addr)]
// 	// if !ok {
// 	// 	items = make(map[string]*TxItem)
// 	// 	this.NotSpentBalance[utils.Bytes2string(*item.Addr)] = items
// 	// }
// 	// items[key] = &item
// 	// this.lock.Unlock()
// }

/*
	查询地址余额
*/
func (this *TxItemManagerDB) FindBalanceAll(height, amount uint64) []*TxItem {
	// addrsBs := make([][]byte, 0)
	// for i, _ := range addrs {
	// 	addrsBs = append(addrsBs, addrs[i])
	// }
	items, err := new(sqlite3_db.TxItem).FindNotSpentTxitem(height, amount, config.Mining_pay_vin_max)
	if err != nil {
		engine.Log.Error("FindBalance txItem_status_notSpent error:", err.Error())
	}

	txItems := make([]*TxItem, 0)
	for i, _ := range *items {
		one := (*items)[i]
		addr := crypto.AddressCoin(one.Addr)
		itemOne := &TxItem{
			Id:           one.Id,           //
			Addr:         &addr,            //收款地址
			Value:        one.Value,        //余额
			Txid:         []byte(one.Txid), //交易id
			VoutIndex:    one.VoutIndex,    //交易输出index，从0开始
			Height:       one.Height,       //区块高度，排序用
			VoteType:     one.VoteType,     //投票类型
			LockupHeight: one.LockupHeight, //锁仓高度/锁仓时间
			Status:       one.Status,       //状态
		}
		txItems = append(txItems, itemOne)
	}
	return txItems
	// return this.findBalanceForStatusByAddrs(addrs, txItem_status_notSpent)
}

/*
	查询地址余额
*/
func (this *TxItemManagerDB) FindBalanceByAddr(addr crypto.AddressCoin, height, amount uint64) []*TxItem {
	// addrsBs := make([][]byte, 0)
	// for i, _ := range addrs {
	// 	addrsBs = append(addrsBs, addrs[i])
	// }
	items, err := new(sqlite3_db.TxItem).FindNotSpentTxitemByAddr(addr, height, amount, config.Mining_pay_vin_max)
	if err != nil {
		engine.Log.Error("FindBalance txItem_status_notSpent error:", err.Error())
	}

	txItems := make([]*TxItem, 0)
	for i, _ := range *items {
		one := (*items)[i]
		addr := crypto.AddressCoin(one.Addr)
		itemOne := &TxItem{
			Id:           one.Id,           //
			Addr:         &addr,            //收款地址
			Value:        one.Value,        //余额
			Txid:         []byte(one.Txid), //交易id
			VoutIndex:    one.VoutIndex,    //交易输出index，从0开始
			Height:       one.Height,       //区块高度，排序用
			VoteType:     one.VoteType,     //投票类型
			LockupHeight: one.LockupHeight, //锁仓高度/锁仓时间
			Status:       one.Status,       //状态
		}
		txItems = append(txItems, itemOne)
	}
	return txItems
	// return this.findBalanceForStatusByAddrs(addrs, txItem_status_notSpent)
}

/*
	查询地址冻结的余额
*/
// func (this *TxItemManagerDB) FindBalanceFrozen(addrs ...crypto.AddressCoin) []*TxItem {
// 	addrsBs := make([][]byte, 0)
// 	for i, _ := range addrs {
// 		addrsBs = append(addrsBs, addrs[i])
// 	}
// 	items, err := new(sqlite3_db.TxItem).FindStateByAddrs(txItem_status_frozen, addrsBs)
// 	if err != nil {
// 		engine.Log.Error("FindBalance txItem_status_frozen error:", err.Error())
// 	}

// 	txItems := make([]*TxItem, 0)
// 	for i, _ := range items {
// 		one := items[i]
// 		addr := crypto.AddressCoin(one.Addr)
// 		itemOne := &TxItem{
// 			Id:           one.Id,           //
// 			Addr:         &addr,            //收款地址
// 			Value:        one.Value,        //余额
// 			Txid:         []byte(one.Txid), //交易id
// 			VoutIndex:    one.VoutIndex,    //交易输出index，从0开始
// 			Height:       one.Height,       //区块高度，排序用
// 			VoteType:     one.VoteType,     //投票类型
// 			LockupHeight: one.LockupHeight, //锁仓高度/锁仓时间
// 			Status:       one.Status,       //状态
// 		}
// 		txItems = append(txItems, itemOne)
// 	}
// 	return txItems
// 	// return this.findBalanceForStatusByAddrs(addrStr, txItem_status_frozen)
// 	// items, err := new(sqlite3_db.TxItem).FindStateByAddrs(txItem_status_frozen, addrs)
// 	// if err != nil {
// 	// 	engine.Log.Error("FindBalance txItem_status_frozen error:", err.Error())
// 	// }
// 	// return items
// }

/*
	通过部分地址查询不同状态的余额
*/
// func (this *TxItemManagerDB) findBalanceForStatusByAddrs(addrStr []crypto.AddressCoin, status int32) []*TxItem {
// 	this.lock.RLock()
// 	itemsResult := make([]*TxItem, 0)
// 	for _, one := range addrStr {

// 		items, ok := this.NotSpentBalance[utils.Bytes2string(one)]
// 		// itemsItr, ok := this.NotSpentBalance.Load(utils.Bytes2string(one))
// 		if !ok {
// 			continue
// 		}
// 		for _, item := range items {
// 			if atomic.LoadInt32(&item.Status) != status {
// 				continue
// 			}
// 			itemsResult = append(itemsResult, item)
// 		}
// 	}
// 	this.lock.RUnlock()
// 	return itemsResult
// }

func (this *TxItemManagerDB) FindBalanceValue(height uint64) (notspent, frozen, lock uint64) {
	// this.lock.RLock()
	// notspent = this.notSpentValue
	// frozen = this.frozenValue
	// lock = this.lockValue
	// this.lock.RUnlock()

	notspent, _ = new(sqlite3_db.TxItem).FindNotSpentSum(height)
	frozen, _ = new(sqlite3_db.TxItem).FindFrozenHeightSum(height)
	lock, _ = new(sqlite3_db.TxItem).FindLockHeightSum(height)
	return
}

/*
	查询所有地址余额
*/
// func (this *TxItemManagerDB) FindBalanceAll() []*TxItem {

// 	items, err := new(sqlite3_db.TxItem).FindState(txItem_status_notSpent)
// 	if err != nil {
// 		engine.Log.Error("FindBalanceAll txItem_status_notSpent error:", err.Error())
// 	}

// 	txItems := make([]*TxItem, 0)
// 	for i, _ := range items {
// 		one := items[i]
// 		addr := crypto.AddressCoin(one.Addr)
// 		itemOne := &TxItem{
// 			Id:           one.Id,           //
// 			Addr:         &addr,            //收款地址
// 			Value:        one.Value,        //余额
// 			Txid:         []byte(one.Txid), //交易id
// 			VoutIndex:    one.VoutIndex,    //交易输出index，从0开始
// 			Height:       one.Height,       //区块高度，排序用
// 			VoteType:     one.VoteType,     //投票类型
// 			LockupHeight: one.LockupHeight, //锁仓高度/锁仓时间
// 			Status:       one.Status,       //状态
// 		}
// 		txItems = append(txItems, itemOne)
// 	}
// 	return txItems
// 	// return this.findBalanceForStatus(txItem_status_notSpent)
// }

/*
	查询所有地址冻结的余额
*/
// func (this *TxItemManagerDB) FindBalanceFrozenAll() []*TxItem {

// 	items, err := new(sqlite3_db.TxItem).FindState(txItem_status_frozen)
// 	if err != nil {
// 		engine.Log.Error("FindBalanceAll txItem_status_frozen error:", err.Error())
// 	}

// 	txItems := make([]*TxItem, 0)
// 	for i, _ := range items {
// 		one := items[i]
// 		addr := crypto.AddressCoin(one.Addr)
// 		itemOne := &TxItem{
// 			Id:           one.Id,           //
// 			Addr:         &addr,            //收款地址
// 			Value:        one.Value,        //余额
// 			Txid:         []byte(one.Txid), //交易id
// 			VoutIndex:    one.VoutIndex,    //交易输出index，从0开始
// 			Height:       one.Height,       //区块高度，排序用
// 			VoteType:     one.VoteType,     //投票类型
// 			LockupHeight: one.LockupHeight, //锁仓高度/锁仓时间
// 			Status:       one.Status,       //状态
// 		}
// 		txItems = append(txItems, itemOne)
// 	}
// 	return txItems
// 	// return this.findBalanceForStatus(txItem_status_frozen)
// }

/*
	查询所有地址锁仓的余额
*/
// func (this *TxItemManagerDB) FindBalanceLockupAll() []*TxItem {
// 	items, err := new(sqlite3_db.TxItem).FindState(txItem_status_lock)
// 	if err != nil {
// 		engine.Log.Error("FindBalanceAll txItem_status_lock error:", err.Error())
// 	}

// 	txItems := make([]*TxItem, 0)
// 	for i, _ := range items {
// 		one := items[i]
// 		addr := crypto.AddressCoin(one.Addr)
// 		itemOne := &TxItem{
// 			Id:           one.Id,           //
// 			Addr:         &addr,            //收款地址
// 			Value:        one.Value,        //余额
// 			Txid:         []byte(one.Txid), //交易id
// 			VoutIndex:    one.VoutIndex,    //交易输出index，从0开始
// 			Height:       one.Height,       //区块高度，排序用
// 			VoteType:     one.VoteType,     //投票类型
// 			LockupHeight: one.LockupHeight, //锁仓高度/锁仓时间
// 			Status:       one.Status,       //状态
// 		}
// 		txItems = append(txItems, itemOne)
// 	}
// 	return txItems
// 	// return this.findBalanceForStatus(txItem_status_lock)
// }

/*
	根据状态查询所有地址余额
*/
// func (this *TxItemManagerDB) findBalanceForStatus(status int32) []*TxItem {
// 	itemsResult := make([]*TxItem, 0)
// 	this.lock.RLock()
// 	for _, items := range this.NotSpentBalance {
// 		for _, item := range items {
// 			if atomic.LoadInt32(&item.Status) != status {
// 				continue
// 			}
// 			itemsResult = append(itemsResult, item)
// 		}
// 	}
// 	this.lock.RUnlock()
// 	return itemsResult
// }

/*
	冻结交易余额
*/
func (this *TxItemManagerDB) Frozen(items []*TxItem, tx TxItr) {
	// keys := make([][]byte, 0)
	// for _, one := range itemCount.SubItems {
	// 	key := append(utils.Hash_SHA3_256(one.Txid), utils.Uint64ToBytes(one.VoutIndex)...)
	// 	keys = append(keys, key)
	// 	// new(sqlite3_db.TxItem).RemoveOne(one.Txid, one.Addr, one.VoutIndex)
	// }
	// new(sqlite3_db.TxItem).FrozenKeys(keys)

	ids := make([]int64, 0)
	for _, one := range items {
		ids = append(ids, one.Id)
	}

	lockHeight := tx.GetLockHeight()

	err := new(sqlite3_db.TxItem).UpdateLockHeight(&ids, lockHeight)
	if err != nil {
		engine.Log.Error("FindBalanceAll txItem_status_lock error:", err.Error())
	}

	// for _, one := range items {
	// 	atomic.StoreInt32(&one.Status, txItem_status_lock)
	// 	atomic.StoreUint64(&one.LockupHeight, lockHeight)
	// }
}

/*
	定时解冻交易
	及时解冻交易存在加载区块效率低的问题
*/
// func (this *TxItemManagerDB) TimingUnfrozen() {
// 	for range time.NewTicker((time.Second * config.Mining_block_time) / 2).C {
// 		this.Unfrozen(GetLongChain().GetCurrentBlock(), time.Now().Unix())
// 	}
// }

/*
	解冻回滚冻结的交易
*/
// func (this *TxItemManagerDB) Unfrozen(blockHeight uint64, blockTime int64) {
// 	// this.lock.RLock()
// 	// start := time.Now()

// 	txitemDB := new(sqlite3_db.TxItem)
// 	txitemDB.UnfrozenForHeight(txItem_status_frozen, txItem_status_lock, txItem_status_notSpent, uint64(blockHeight))
// 	txitemDB.UnfrozenForTime(txItem_status_frozen, txItem_status_lock, txItem_status_notSpent, uint64(blockTime))
// 	// engine.Log.Info("Unfrozen spend time 222: %s", time.Now().Sub(start))
// }

/*
	创建一个txitem管理器
*/
func NewTxItemManagerDB() *TxItemManagerDB {
	tm := &TxItemManagerDB{
		lock: new(sync.RWMutex), //
		// NotSpentBalance: make(map[string]map[string]*TxItem), //
	}
	// go tm.TimingUnfrozen()
	return tm
}
