package mining

import (
	"mandela/core/engine"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type TxItemManager struct {
	lock            *sync.RWMutex                 //
	NotSpentBalance map[string]map[string]*TxItem //保存各个状态的txitem，解锁、冻结等只是状态的改变。key:string=收款地址;value:*sync.Map(key:string=[txid]_[voutIndex];value:*TxItem=TxItem;)=;
	notSpentValue   uint64                        //未花费的余额
	lockValue       uint64                        //等待上链，临时冻结的余额
	frozenValue     uint64                        //锁仓，等待解锁的余额
}

/*
	批量添加和删除TxItem
*/
func (this *TxItemManager) CountTxItem(itemCount TxItemCount, blockHeight uint64, blockTime int64) {

	this.lock.Lock()
	// start := time.Now()

	// this.NotSpentBalanceFotAddrKeyLock.Lock()
	//删除余额
	for index, _ := range itemCount.SubItems {
		one := itemCount.SubItems[index]
		// engine.Log.Info("del item:%s_%d", hex.EncodeToString(one.Txid), one.VoutIndex)
		key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))

		items, ok := this.NotSpentBalance[utils.Bytes2string(one.Addr)]
		if !ok {
			items = make(map[string]*TxItem)
			this.NotSpentBalance[utils.Bytes2string(one.Addr)] = items
			continue
		}
		txItem, ok := items[key]
		if !ok {
			continue
		}
		switch txItem.Status {
		case txItem_status_notSpent:
			// engine.Log.Info("%s voutIndex:%d old:%d new:%d change:-%d", hex.EncodeToString(one.Txid), one.VoutIndex, this.notSpentValue, (this.notSpentValue - txItem.Value), txItem.Value)
			this.notSpentValue -= txItem.Value
		case txItem_status_lock:
			this.lockValue -= txItem.Value
		case txItem_status_frozen:
			this.frozenValue -= txItem.Value
		}
		delete(items, key)
	}
	//添加余额
	for _, item := range itemCount.Additems {
		// engine.Log.Info("add item:%s_%d %d %d", hex.EncodeToString(item.Txid), item.VoutIndex, item.LockupHeight, item.Value)

		key := utils.Bytes2string(item.Txid) + "_" + strconv.Itoa(int(item.VoutIndex))

		//根据冻结高度判断余额是否可用
		lockHeight := atomic.LoadUint64(&item.LockupHeight)

		if CheckFrozenHeightFree(lockHeight, blockHeight, blockTime) {
			atomic.StoreInt32(&item.Status, txItem_status_notSpent)
		} else {
			atomic.StoreInt32(&item.Status, txItem_status_frozen)
		}

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

		items, ok := this.NotSpentBalance[utils.Bytes2string(*item.Addr)]
		if !ok {
			items = make(map[string]*TxItem)
			this.NotSpentBalance[utils.Bytes2string(*item.Addr)] = items
		}
		switch item.Status {
		case txItem_status_notSpent:
			// engine.Log.Info("add notSpent %s voutIndex:%d old:%d new:%d change:+%d", hex.EncodeToString(item.Txid), item.VoutIndex, this.notSpentValue, (this.notSpentValue + item.Value), item.Value)
			this.notSpentValue += item.Value
		case txItem_status_lock:
			this.lockValue += item.Value
		case txItem_status_frozen:
			// engine.Log.Info("add frozen %s voutIndex:%d old:%d new:%d change:+%d", hex.EncodeToString(item.Txid), item.VoutIndex, this.notSpentValue, (this.notSpentValue + item.Value), item.Value)
			this.frozenValue += item.Value
		}
		items[key] = item

	}
	this.lock.Unlock()

	// engine.Log.Info("CountTxItem spend time: %s", time.Now().Sub(start))
}

/*
	添加一个TxItem
*/
// func (this *TxItemManager) AddTxItem(item TxItem) {
// 	this.lock.Lock()
// 	atomic.StoreInt32(&item.Status, txItem_status_notSpent)
// 	key := utils.Bytes2string(item.Txid) + "_" + strconv.Itoa(int(item.VoutIndex))

// 	items, ok := this.NotSpentBalance[utils.Bytes2string(*item.Addr)]
// 	if !ok {
// 		items = make(map[string]*TxItem)
// 		this.NotSpentBalance[utils.Bytes2string(*item.Addr)] = items
// 	}
// 	switch item.Status {
// 	case txItem_status_notSpent:
// 		engine.Log.Info("old:%d new:%d change:%d", this.notSpentValue, (this.notSpentValue + item.Value), item.Value)
// 		this.notSpentValue += item.Value
// 	case txItem_status_lock:
// 		this.lockValue += item.Value
// 	case txItem_status_frozen:
// 		this.frozenValue += item.Value
// 	}
// 	items[key] = &item
// 	this.lock.Unlock()
// }

/*
	查询地址余额
*/
func (this *TxItemManager) FindBalanceNotSpent(addrStr ...crypto.AddressCoin) []*TxItem {
	return this.findBalanceForStatusByAddrs(addrStr, txItem_status_notSpent)
}

/*
	查询地址冻结的余额
*/
func (this *TxItemManager) FindBalanceFrozen(addrStr ...crypto.AddressCoin) []*TxItem {
	return this.findBalanceForStatusByAddrs(addrStr, txItem_status_frozen)
}

/*
	通过部分地址查询不同状态的余额
*/
func (this *TxItemManager) findBalanceForStatusByAddrs(addrStr []crypto.AddressCoin, status int32) []*TxItem {
	this.lock.RLock()
	itemsResult := make([]*TxItem, 0)
	for _, one := range addrStr {

		items, ok := this.NotSpentBalance[utils.Bytes2string(one)]
		// itemsItr, ok := this.NotSpentBalance.Load(utils.Bytes2string(one))
		if !ok {
			continue
		}
		for _, item := range items {
			if atomic.LoadInt32(&item.Status) != status {
				continue
			}
			itemsResult = append(itemsResult, item)
		}
	}
	this.lock.RUnlock()
	return itemsResult
}

func (this *TxItemManager) FindBalanceValue() (notspent, frozen, lock uint64) {
	this.lock.RLock()
	notspent = this.notSpentValue
	frozen = this.frozenValue
	lock = this.lockValue
	this.lock.RUnlock()
	return
}

/*
	查询所有地址余额
*/
func (this *TxItemManager) FindBalanceNotSpentAll() []*TxItem {
	return this.findBalanceForStatus(txItem_status_notSpent)
}

/*
	查询所有地址的余额明细
*/
func (this *TxItemManager) FindBalanceAllAddrs() (map[string]uint64, map[string]uint64, map[string]uint64) {
	basMap := make(map[string]uint64)      //可用余额
	fbasMap := make(map[string]uint64)     //冻结的余额
	baLockupMap := make(map[string]uint64) //锁仓的余额

	// itemsResult := make([]*TxItem, 0)
	this.lock.RLock()
	for _, items := range this.NotSpentBalance {
		for _, item := range items {
			switch atomic.LoadInt32(&item.Status) {
			case txItem_status_notSpent: // = int32(0) //未花费的交易余额，可以正常支付
				value, _ := basMap[utils.Bytes2string(*item.Addr)]
				basMap[utils.Bytes2string(*item.Addr)] = value + item.Value
			case txItem_status_frozen: //   = int32(1) //锁仓,区块达到指定高度才能使用
				value, _ := fbasMap[utils.Bytes2string(*item.Addr)]
				fbasMap[utils.Bytes2string(*item.Addr)] = value + item.Value
			case txItem_status_lock: //    = int32(2) //冻结高度，指定高度还未上链，则转为未花费的交易
				value, _ := baLockupMap[utils.Bytes2string(*item.Addr)]
				baLockupMap[utils.Bytes2string(*item.Addr)] = value + item.Value
			}
		}
	}
	this.lock.RUnlock()
	return basMap, fbasMap, baLockupMap
}

/*
	查询所有地址冻结的余额
*/
func (this *TxItemManager) FindBalanceFrozenAll() []*TxItem {
	return this.findBalanceForStatus(txItem_status_frozen)
}

/*
	查询所有地址锁仓的余额
*/
func (this *TxItemManager) FindBalanceLockupAll() []*TxItem {
	return this.findBalanceForStatus(txItem_status_lock)
}

/*
	根据状态查询所有地址余额
*/
func (this *TxItemManager) findBalanceForStatus(status int32) []*TxItem {
	itemsResult := make([]*TxItem, 0)
	this.lock.RLock()
	for _, items := range this.NotSpentBalance {
		for _, item := range items {
			if atomic.LoadInt32(&item.Status) != status {
				continue
			}
			itemsResult = append(itemsResult, item)
		}
	}
	this.lock.RUnlock()
	return itemsResult
}

/*
	冻结交易余额
*/
func (this *TxItemManager) Frozen(items []*TxItem, tx TxItr) {
	lockHeight := tx.GetLockHeight()
	lockTotal := uint64(0)
	for _, one := range items {
		lockTotal += one.Value
		atomic.StoreInt32(&one.Status, txItem_status_lock)
		atomic.StoreUint64(&one.LockupHeight, lockHeight)
	}
	this.lock.Lock()
	this.lockValue += lockTotal
	// engine.Log.Info("old:%d new:%d change:%d", this.notSpentValue, (this.notSpentValue - lockTotal), lockTotal)
	this.notSpentValue -= lockTotal
	this.lock.Unlock()
}

/*
	定时解冻交易
	及时解冻交易存在加载区块效率低的问题
*/
// func (this *TxItemManager) TimingUnfrozen() {
// 	for range time.NewTicker((time.Second * config.Mining_block_time) / 2).C {
// 		this.Unfrozen(GetLongChain().GetCurrentBlock(), time.Now().Unix())
// 	}
// }

/*
	解冻冻结的交易
*/
func (this *TxItemManager) Unfrozen(blockHeight uint64, blockTime int64) {
	this.lock.RLock()
	start := time.Now()

	for _, items := range this.NotSpentBalance {
		for _, item := range items {
			status := atomic.LoadInt32(&item.Status)
			if status != txItem_status_frozen && status != txItem_status_lock {
				continue
			}
			//开始解冻锁仓的余额,开始回滚冻结的交易
			lockHeight := atomic.LoadUint64(&item.LockupHeight)

			if !CheckFrozenHeightFree(lockHeight, blockHeight, blockTime) {
				continue
			}

			// if lockHeight > config.Wallet_frozen_time_min {
			// 	//按时间锁仓
			// 	if int64(lockHeight) > blockTime {
			// 		continue
			// 	}
			// } else {
			// 	//按块高度锁仓
			// 	if lockHeight > blockHeight {
			// 		continue
			// 	}
			// }
			switch item.Status {
			case txItem_status_lock:
				this.lockValue -= item.Value
			case txItem_status_frozen:
				this.frozenValue -= item.Value
			}

			// engine.Log.Info("old:%d new:%d change:%d", this.notSpentValue, (this.notSpentValue + item.Value), item.Value)
			this.notSpentValue += item.Value
			// engine.Log.Info("Unfrozen notSpent: %s %d %d %d", txidStr, item.OutIndex, lockHeight, blockHeight)
			atomic.StoreInt32(&item.Status, txItem_status_notSpent)
		}
	}
	this.lock.RUnlock()
	engine.Log.Info("Unfrozen spend time 222: %s", time.Now().Sub(start))
}

/*
	创建一个txitem管理器
*/
func NewTxItemManager() *TxItemManager {
	tm := &TxItemManager{
		lock:            new(sync.RWMutex),                   //
		NotSpentBalance: make(map[string]map[string]*TxItem), //
	}
	// go tm.TimingUnfrozen()
	return tm
}
