package mining

import (
	"mandela/config"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"strconv"
	"sync"
	"time"
)

type TxItemManagerOld struct {
	NotSpentBalanceFotAddrKeyLock *sync.RWMutex                 //
	NotSpentBalanceFotAddrKey     map[string]map[string]*TxItem //保存各个地址的余额，key:string=收款地址;value:*Balance=收益列表;
	FrozenBalance                 map[string]map[string]*TxItem //保存冻结各个地址的余额，key:string=收款地址;value:*Balance=收益列表;
	FrozenBalanceLockHeight       map[string]uint64             //冻结交易的锁定回滚高度
	BalanceLockup                 map[string]map[string]*TxItem //锁仓的交易
	BalanceLockupLockHeight       map[string]uint64             //锁仓的交易对应的锁仓高度，超过高度会解锁
}

/*
	批量添加和删除TxItem
*/
func (this *TxItemManagerOld) CountTxItem(itemCount TxItemCount, blockHeight uint64, blockTime int64) {
	this.NotSpentBalanceFotAddrKeyLock.Lock()
	//删除余额
	for _, one := range itemCount.SubItems {
		key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))

		items, ok := this.NotSpentBalanceFotAddrKey[utils.Bytes2string(one.Addr)]
		if !ok {
			items = make(map[string]*TxItem)
		}
		delete(items, key)
		this.NotSpentBalanceFotAddrKey[utils.Bytes2string(one.Addr)] = items
	}
	//添加余额
	// currentHeight := GetLongChain().GetCurrentBlock()
	// currentHeight := uint64(0)
	for _, item := range itemCount.Additems {
		key := utils.Bytes2string(item.Txid) + "_" + strconv.Itoa(int(item.VoutIndex))

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

		//根据冻结高度判断余额是否可用
		if item.LockupHeight > blockHeight {
			//冻结高度大于当前高度，则锁仓
			items, ok := this.BalanceLockup[utils.Bytes2string(*item.Addr)]
			if !ok {
				items = make(map[string]*TxItem)
				this.BalanceLockup[utils.Bytes2string(*item.Addr)] = items
			}
			items[key] = item
			//保存锁仓高度
			this.BalanceLockupLockHeight[key] = item.LockupHeight
		} else {
			// engine.Log.Info("添加余额 111 %+v", item)
			items, ok := this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)]
			if !ok {
				items = make(map[string]*TxItem)
				this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)] = items
			}
			items[key] = item
			// engine.Log.Info("添加余额 333 %+v", this.NotSpentBalanceFotAddrKey)
		}
	}
	//删除冻结item
	for _, item := range itemCount.SubItems {
		key := utils.Bytes2string(item.Txid) + "_" + strconv.Itoa(int(item.VoutIndex))
		// engine.Log.Debug("删除冻结item %s %s", item.Addr.B58String(), hex.EncodeToString([]byte(key)))
		items, ok := this.FrozenBalance[utils.Bytes2string(item.Addr)]
		if !ok {
			items = make(map[string]*TxItem)
		}
		delete(items, key)
		this.FrozenBalance[utils.Bytes2string(item.Addr)] = items
	}
	this.NotSpentBalanceFotAddrKeyLock.Unlock()
}

/*
	添加一个TxItem
*/
func (this *TxItemManagerOld) AddTxItem(item TxItem) {
	key := utils.Bytes2string(item.Txid) + "_" + strconv.Itoa(int(item.VoutIndex))

	this.NotSpentBalanceFotAddrKeyLock.Lock()
	items, ok := this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)]
	if !ok {
		items = make(map[string]*TxItem)
	}
	items[key] = &item
	this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)] = items
	this.NotSpentBalanceFotAddrKeyLock.Unlock()
}

/*
	删除一个TxItem
*/
// func (this *TxItemManager) DeleteTxItem(txidStr string, voutIndex uint64, addrStr string) {
// 	key := txidStr + "_" + strconv.Itoa(int(voutIndex))
// 	// this.NotSpentBalance.Delete(addrStr)

// 	this.NotSpentBalanceFotAddrKeyLock.Lock()
// 	items, ok := this.NotSpentBalanceFotAddrKey[addrStr]
// 	if !ok {
// 		items = make(map[string]*TxItem)
// 	}
// 	delete(items, key)
// 	this.NotSpentBalanceFotAddrKey[addrStr] = items
// 	this.NotSpentBalanceFotAddrKeyLock.Unlock()
// }

/*
	查询地址余额
*/
func (this *TxItemManagerOld) FindBalance(addrStr ...crypto.AddressCoin) []*TxItem {
	itemsResult := make([]*TxItem, 0)
	this.NotSpentBalanceFotAddrKeyLock.RLock()
	for _, one := range addrStr {
		items, ok := this.NotSpentBalanceFotAddrKey[utils.Bytes2string(one)]
		if !ok {
			continue
		}
		for _, v := range items {
			itemsResult = append(itemsResult, v)
		}
	}
	this.NotSpentBalanceFotAddrKeyLock.RUnlock()
	return itemsResult
}

/*
	查询地址冻结的余额
*/
func (this *TxItemManagerOld) FindBalanceFrozen(addrStr ...crypto.AddressCoin) []*TxItem {
	itemsResult := make([]*TxItem, 0)
	this.NotSpentBalanceFotAddrKeyLock.RLock()
	// engine.Log.Info("查询冻结余额 %+v", this.FrozenBalance)
	for _, one := range addrStr {
		ba, ok := this.FrozenBalance[utils.Bytes2string(one)]
		if !ok {
			continue
		}
		for _, v := range ba {
			itemsResult = append(itemsResult, v)
		}
	}
	this.NotSpentBalanceFotAddrKeyLock.RUnlock()
	return itemsResult
}

/*
	查询所有地址余额
*/
func (this *TxItemManagerOld) FindBalanceAll() []*TxItem {
	items := make([]*TxItem, 0)
	this.NotSpentBalanceFotAddrKeyLock.RLock()
	for _, v := range this.NotSpentBalanceFotAddrKey {
		for _, item := range v {
			items = append(items, item)
		}
	}
	this.NotSpentBalanceFotAddrKeyLock.RUnlock()
	return items
}

/*
	查询所有地址冻结的余额
*/
func (this *TxItemManagerOld) FindBalanceFrozenAll() []*TxItem {
	items := make([]*TxItem, 0)
	this.NotSpentBalanceFotAddrKeyLock.RLock()
	for _, v := range this.FrozenBalance {
		for _, item := range v {
			items = append(items, item)
		}
	}
	this.NotSpentBalanceFotAddrKeyLock.RUnlock()
	return items
}

/*
	查询所有地址锁仓的余额
*/
func (this *TxItemManagerOld) FindBalanceLockupAll() []*TxItem {
	items := make([]*TxItem, 0)
	this.NotSpentBalanceFotAddrKeyLock.RLock()
	for _, v := range this.BalanceLockup {
		for _, item := range v {
			items = append(items, item)
		}
	}
	this.NotSpentBalanceFotAddrKeyLock.RUnlock()
	return items
}

/*
	冻结交易余额
*/
func (this *TxItemManagerOld) Frozen(items []*TxItem, tx TxItr) {
	this.NotSpentBalanceFotAddrKeyLock.Lock()
	for i, one := range items {
		addrStr := utils.Bytes2string(*one.Addr) // one.GetAddrStr()
		keyStr := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))
		// engine.Log.Debug("添加冻结item %s %s", one.Addr.B58String(), hex.EncodeToString([]byte(keyStr)))

		//删除余额里面的交易
		itemMap, ok := this.NotSpentBalanceFotAddrKey[addrStr]
		if !ok {
			itemMap = make(map[string]*TxItem)
		}
		delete(itemMap, keyStr)
		// engine.Log.Debug("添加冻结 111 %+v", itemMap)
		this.NotSpentBalanceFotAddrKey[addrStr] = itemMap
		// engine.Log.Debug("添加冻结 222 %+v", this.NotSpentBalanceFotAddrKey)

		//放入冻结的余额里面
		itemMap, ok = this.FrozenBalance[addrStr]
		if !ok {
			itemMap = make(map[string]*TxItem)
		}
		itemMap[keyStr] = items[i]
		this.FrozenBalance[addrStr] = itemMap

		//保存冻结高度
		this.FrozenBalanceLockHeight[keyStr] = tx.GetLockHeight()

		// engine.Log.Debug("添加冻结 333 %+v", this.FrozenBalance)
	}

	this.NotSpentBalanceFotAddrKeyLock.Unlock()

}

/*
	定时解冻交易
	及时解冻交易存在加载区块效率低的问题
*/
func (this *TxItemManagerOld) TimingUnfrozen() {
	for range time.NewTicker(time.Second * (config.Mining_block_time / 2)).C {
		this.Unfrozen(GetLongChain().GetCurrentBlock(), time.Now().Unix())
	}
}

/*
	解冻回滚冻结的交易
*/
func (this *TxItemManagerOld) Unfrozen(blockHeight uint64, blockTime int64) {
	this.NotSpentBalanceFotAddrKeyLock.Lock()
	// engine.Log.Info("开始回滚冻结的交易")
	for _, fba := range this.FrozenBalance {
		for key, item := range fba {
			// keyStr := item.GetTxidStr() + "_" + strconv.Itoa(int(item.OutIndex))
			lockHeight, ok := this.FrozenBalanceLockHeight[key]
			// engine.Log.Info("开始回滚冻结的交易 判断 %d %d", blockHeight, lockHeight)
			if lockHeight > blockHeight {
				continue
			}
			delete(this.FrozenBalanceLockHeight, key)
			// engine.Log.Info("开始回滚冻结的交易 回滚")
			//开始回滚交易
			//添加到余额列表
			ba, ok := this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)]
			if !ok {
				ba = make(map[string]*TxItem)
			}
			ba[key] = item
			this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)] = ba
			//删除锁定的交易
			delete(fba, key)
			// this.FrozenBalance[addrStr] = fba
		}
	}
	//开始解冻锁仓的余额
	for _, fba := range this.BalanceLockup {
		for key, item := range fba {
			// keyStr := item.GetTxidStr() + "_" + strconv.Itoa(int(item.OutIndex))
			lockHeight, ok := this.BalanceLockupLockHeight[key]
			// engine.Log.Info("开始回滚冻结的交易 判断 %d %d", blockHeight, lockHeight)

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

			delete(this.BalanceLockupLockHeight, key)
			// engine.Log.Info("开始回滚冻结的交易 回滚")
			//开始回滚交易
			//添加到余额列表
			ba, ok := this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)]
			if !ok {
				ba = make(map[string]*TxItem)
			}
			ba[key] = item
			this.NotSpentBalanceFotAddrKey[utils.Bytes2string(*item.Addr)] = ba
			//删除锁定的交易
			delete(fba, key)
			// this.BalanceLockup[addrStr] = fba
		}
	}
	this.NotSpentBalanceFotAddrKeyLock.Unlock()
}

/*
	创建一个txitem管理器
*/
func NewTxItemManagerOld() *TxItemManagerOld {
	tm := &TxItemManagerOld{
		NotSpentBalanceFotAddrKeyLock: new(sync.RWMutex),                   //
		NotSpentBalanceFotAddrKey:     make(map[string]map[string]*TxItem), //
		FrozenBalance:                 make(map[string]map[string]*TxItem), //保存冻结各个地址的余额，key:string=收款地址;value:*Balance=收益列表;
		FrozenBalanceLockHeight:       make(map[string]uint64),             //
		BalanceLockup:                 make(map[string]map[string]*TxItem), //锁仓的交易
		BalanceLockupLockHeight:       make(map[string]uint64),             //锁仓的交易对应的锁仓高度，超过高度会解锁
	}
	// go tm.TimingUnfrozen()
	return tm
}
