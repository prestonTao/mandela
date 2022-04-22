package token

import (
	"mandela/core/utils"
	"mandela/core/utils/crypto"

	// "mandela/core/utils/crypto"
	// "bytes"
	// "encoding/hex"
	// "mandela/core/engine"
	"mandela/chain_witness_vote/mining"
	"sort"
	"strconv"
	"sync"
)

type TokenTxItemManager struct {
	lock                      *sync.RWMutex                                   //
	NotSpentBalanceFotAddrKey map[string]map[string]map[string]*mining.TxItem //保存各个地址的余额，最外层key:string=token合约地址;第二层key:string=收款地址;第三层key:string=交易txid+vout;
	FrozenBalance             map[string]map[string]map[string]*mining.TxItem //保存冻结各个地址的余额，最外层key:string=token合约地址;第二层key:string=收款地址;第三层key:string=交易txid+vout;
	FrozenBalanceLockHeight   map[string]uint64                               //冻结交易的锁定回滚高度
	BalanceLockup             map[string]map[string]map[string]*mining.TxItem //锁仓的交易
	BalanceLockupLockHeight   map[string]uint64                               //锁仓的交易的锁仓高度
}

/*
	统计交易输入中要处理的参数
*/
type TokenTxItemCount struct {
	PublishTxidStr string         //token发布的交易id
	Name           string         //名称（全称）
	Symbol         string         //单位
	Supply         uint64         //发行总量
	Additems       *mining.TxItem //交易
}

type TokenTxSubItems struct {
	PublishTxid []byte             //token发布的交易id
	Addr        crypto.AddressCoin //收款地址
	Txid        []byte             //交易id
	VoutIndex   uint64             //交易输出索引
}

/*
	添加token信息
*/
// func (this *TokenTxItemManager) AddTokenInfo(tokenInfo *TokenInfo) {
// 	this.lock.Lock()
// 	this.TokenInfo[]
// 	this.lock.Unlock()
// }

/*
	批量添加和删除TxItem
*/
func (this *TokenTxItemManager) CountTxItem(addItemCount []TokenTxItemCount, subItem []TokenTxSubItems) {
	// engine.Log.Info("批量添加和删除TxItem start")
	this.lock.Lock()
	//删除余额
	for _, one := range subItem {
		addrs, ok := this.NotSpentBalanceFotAddrKey[utils.Bytes2string(one.PublishTxid)]
		if !ok {
			addrs = make(map[string]map[string]*mining.TxItem)
			this.NotSpentBalanceFotAddrKey[utils.Bytes2string(one.Addr)] = addrs
		}
		items, ok := addrs[utils.Bytes2string(one.Addr)]
		if !ok {
			items = make(map[string]*mining.TxItem)
			addrs[utils.Bytes2string(one.Addr)] = items
		}
		key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))
		delete(items, key)

		//如果余额为0，则删除相关token信息
		// if len(items) <= 0 {
		// 	delete(this.TokenInfo, one.PublishTxidStr)
		// }
	}
	//添加余额
	currentHeight := mining.GetLongChain().GetCurrentBlock()
	for _, item := range addItemCount {
		key := utils.Bytes2string(item.Additems.Txid) + "_" + strconv.Itoa(int(item.Additems.VoutIndex))
		//根据冻结高度判断余额是否可用
		if item.Additems.LockupHeight > currentHeight {
			//冻结高度大于当前高度，则锁仓
			addrs, ok := this.BalanceLockup[item.PublishTxidStr]
			if !ok {
				addrs = make(map[string]map[string]*mining.TxItem)
				this.BalanceLockup[item.PublishTxidStr] = addrs
			}
			items, ok := addrs[utils.Bytes2string(*item.Additems.Addr)]
			if !ok {
				items = make(map[string]*mining.TxItem)
				addrs[utils.Bytes2string(*item.Additems.Addr)] = items
			}
			items[key] = item.Additems
			//保存锁仓高度
			this.BalanceLockupLockHeight[key] = item.Additems.LockupHeight
		} else {
			addrs, ok := this.NotSpentBalanceFotAddrKey[item.PublishTxidStr]
			if !ok {
				addrs = make(map[string]map[string]*mining.TxItem)
				this.NotSpentBalanceFotAddrKey[item.PublishTxidStr] = addrs
			}
			items, ok := addrs[utils.Bytes2string(*item.Additems.Addr)]
			if !ok {
				items = make(map[string]*mining.TxItem)
				addrs[utils.Bytes2string(*item.Additems.Addr)] = items
			}
			items[key] = item.Additems
			// engine.Log.Info("添加余额 333 %+v", this.NotSpentBalanceFotAddrKey)

		}
		//添加tokeninfo
		// _, ok := this.TokenInfo[item.PublishTxidStr]
		// if !ok {
		// 	tokeninfo := TokenInfo{
		// 		Txid:   item.PublishTxidStr, //发布交易地址
		// 		Name:   item.Name,           //名称（全称）
		// 		Symbol: item.Symbol,         //单位
		// 		Supply: item.Supply,         //发行总量
		// 	}
		// 	// if item.Name != ""{
		// 	// 	tokeninfo.Name = item.Name
		// 	// 	tokeninfo.Symbol = item.Symbol
		// 	// 	tokeninfo.Supply = item.Supply
		// 	// }
		// 	this.TokenInfo[item.PublishTxidStr] = &tokeninfo
		// }
	}
	//删除冻结item
	for _, one := range subItem {
		addrs, ok := this.FrozenBalance[utils.Bytes2string(one.PublishTxid)]
		if !ok {
			addrs = make(map[string]map[string]*mining.TxItem)
			this.FrozenBalance[utils.Bytes2string(one.PublishTxid)] = addrs
		}
		items, ok := addrs[utils.Bytes2string(one.Addr)]
		if !ok {
			items = make(map[string]*mining.TxItem)
			addrs[utils.Bytes2string(one.Addr)] = items
		}

		key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))
		delete(items, key)
	}
	this.lock.Unlock()
	// engine.Log.Info("批量添加和删除TxItem end")
}

/*
	添加一个TxItem
*/
// func (this *TokenTxItemManager) AddTxItem(item mining.TxItem) {
// 	key := item.GetTxidStr() + "_" + strconv.Itoa(int(item.OutIndex))

// 	this.lock.Lock()
// 	items, ok := this.NotSpentBalanceFotAddrKey[item.GetAddrStr()]
// 	if !ok {
// 		items = make(map[string]*mining.TxItem)
// 	}
// 	items[key] = &item
// 	this.NotSpentBalanceFotAddrKey[item.GetAddrStr()] = items
// 	this.lock.Unlock()
// }

/*
	删除一个TxItem
*/
// func (this *TokenTxItemManager) DeleteTxItem(txidStr string, voutIndex uint64, addrStr string) {
// 	key := txidStr + "_" + strconv.Itoa(int(voutIndex))
// 	// this.NotSpentBalance.Delete(addrStr)

// 	this.lock.Lock()
// 	items, ok := this.NotSpentBalanceFotAddrKey[addrStr]
// 	if !ok {
// 		items = make(map[string]*TxItem)
// 	}
// 	delete(items, key)
// 	this.NotSpentBalanceFotAddrKey[addrStr] = items
// 	this.lock.Unlock()
// }

/*
	查询地址余额
*/
func (this *TokenTxItemManager) FindBalanceByAddr(txid, addrStr string) *TokenBalance {
	// engine.Log.Info("查询地址余额 start")
	tb := new(TokenBalance)
	this.lock.RLock()
	defer func() {
		this.lock.RUnlock()
		// engine.Log.Info("查询地址余额 end")
	}()

	// ti, _ := this.TokenInfo[txid]
	// if ti == nil {
	// 	return tb
	// }
	// tb.Name = ti.Name
	// tb.Supply = ti.Supply
	// tb.Symbol = ti.Symbol
	// tb.TokenId = txid

	//查询余额
	addrs, ok := this.NotSpentBalanceFotAddrKey[txid]
	if !ok {
		return tb
	}
	items, ok := addrs[addrStr]
	if !ok {
		return tb
	}
	for _, v := range items {
		tb.Balance = tb.Balance + v.Value
	}
	//查询冻结余额
	addrs, ok = this.FrozenBalance[txid]
	if !ok {
		return tb
	}
	items, ok = addrs[addrStr]
	if !ok {
		return tb
	}
	for _, v := range items {
		tb.BalanceFrozen = tb.BalanceFrozen + v.Value
	}

	//锁仓的余额
	addrs, ok = this.BalanceLockup[txid]
	if !ok {
		return tb
	}
	items, ok = addrs[addrStr]
	if !ok {
		return tb
	}
	for _, v := range items {
		tb.BalanceLockup = tb.BalanceLockup + v.Value
	}
	return tb
}

/*
	查询地址冻结的余额
*/
// func (this *TokenTxItemManager) FindBalanceFrozen(addrStr ...string) []*mining.TxItem {
// 	itemsResult := make([]*mining.TxItem, 0)
// 	this.lock.RLock()
// 	// engine.Log.Info("查询冻结余额 %+v", this.FrozenBalance)
// 	for _, one := range addrStr {
// 		ba, ok := this.FrozenBalance[one]
// 		if !ok {
// 			continue
// 		}
// 		for _, v := range ba {
// 			itemsResult = append(itemsResult, v)
// 		}
// 	}
// 	this.lock.RUnlock()
// 	return itemsResult
// }

/*
	查询一种token的所有地址余额
	@return    map[string]uint64    可用余额。key:string=地址;value:uint64=余额;
	@return    map[string]uint64    冻结余额。key:string=地址;value:uint64=余额;
	@return    map[string]uint64    锁仓余额。key:string=地址;value:uint64=余额;
*/
func (this *TokenTxItemManager) FindTokenBalanceForTxid(txidbs []byte) (map[string]uint64, map[string]uint64, map[string]uint64) {
	txid := utils.Bytes2string(txidbs)
	balance := make(map[string]uint64)
	balanceFrozen := make(map[string]uint64)
	balanceLockup := make(map[string]uint64)
	this.lock.RLock()
	addrs, ok := this.NotSpentBalanceFotAddrKey[txid]
	if ok {
		for addrOne, items := range addrs {
			value := uint64(0)
			for _, itemOne := range items {
				value += itemOne.Value
			}
			balance[addrOne] = value
		}
	}
	addrs, ok = this.FrozenBalance[txid]
	if ok {
		for addrOne, items := range addrs {
			value := uint64(0)
			for _, itemOne := range items {
				value += itemOne.Value
			}
			balanceFrozen[addrOne] = value
		}
	}
	addrs, ok = this.BalanceLockup[txid]
	if ok {
		for addrOne, items := range addrs {
			value := uint64(0)
			for _, itemOne := range items {
				value += itemOne.Value
			}
			balanceLockup[addrOne] = value
		}
	}
	this.lock.RUnlock()
	return balance, balanceFrozen, balanceLockup
}

/*
	查询所有token余额
*/
func (this *TokenTxItemManager) FindTokenBalanceForAll() []TokenBalance {
	// engine.Log.Info("查询所有token余额 start")
	tbs := make([]TokenBalance, 0)
	this.lock.RLock()

	for k, addrs := range this.NotSpentBalanceFotAddrKey {
		tb := new(TokenBalance)
		tb.TokenId = []byte(k)

		// ti, ok := this.TokenInfo[k]
		// if ok {
		// 	tb.Name = ti.Name
		// 	tb.Supply = ti.Supply
		// 	tb.Symbol = ti.Symbol
		// 	tb.TokenId = ti.Txid
		// }
		for _, items := range addrs {
			for _, item := range items {
				tb.Balance = tb.Balance + item.Value
			}
		}

		//查找冻结的余额
		addrs, ok := this.FrozenBalance[k]
		if ok {
			for _, items := range addrs {
				for _, item := range items {
					tb.BalanceFrozen = tb.BalanceFrozen + item.Value
				}
			}
		}

		//查找锁仓高度
		addrs, ok = this.BalanceLockup[k]
		if ok {
			for _, items := range addrs {
				for _, item := range items {
					tb.BalanceLockup = tb.BalanceLockup + item.Value
				}
			}
		}
		tbs = append(tbs, *tb)
	}
	this.lock.RUnlock()
	// engine.Log.Info("查询所有token余额 end")
	return tbs
}

/*
	查询所有地址可用余额
*/
func (this *TokenTxItemManager) GetReadyPayToken(txidbs []byte, srcAddr crypto.AddressCoin) []*mining.TxItem {
	txid := utils.Bytes2string(txidbs)
	// engine.Log.Info("查询所有地址可用余额 start")
	itemsPay := make([]*mining.TxItem, 0)

	if srcAddr != nil {
		this.lock.RLock()
		addrs, ok := this.NotSpentBalanceFotAddrKey[txid]
		if ok {
			items, ok := addrs[utils.Bytes2string(srcAddr)]
			if ok {
				for _, item := range items {
					itemsPay = append(itemsPay, item)
				}
			}
		}
		this.lock.RUnlock()
		return itemsPay
	}

	this.lock.RLock()
	addrs, ok := this.NotSpentBalanceFotAddrKey[txid]
	if ok {
		for _, items := range addrs {
			for _, item := range items {
				itemsPay = append(itemsPay, item)
			}
		}
	}
	this.lock.RUnlock()
	// engine.Log.Info("查询所有地址可用余额 end")
	return itemsPay
}

/*
	查询所有地址冻结的余额
*/
// func (this *TokenTxItemManager) FindBalanceFrozenAll() []*mining.TxItem {
// 	items := make([]*mining.TxItem, 0)
// 	this.lock.RLock()
// 	for _, v := range this.FrozenBalance {
// 		for _, item := range v {
// 			items = append(items, item)
// 		}
// 	}
// 	this.lock.RUnlock()
// 	return items
// }

/*
	冻结交易余额
*/
func (this *TokenTxItemManager) FrozenToken(txidBs []byte, items []*mining.TxItem, tx mining.TxItr) {
	txid := utils.Bytes2string(txidBs)
	// engine.Log.Info("冻结交易余额 start %s %s", txid, tx.GetHashStr())
	this.lock.Lock()
	defer func() {
		this.lock.Unlock()
		// engine.Log.Info("冻结交易余额 end")
	}()
	addrs, ok := this.NotSpentBalanceFotAddrKey[txid]
	if !ok {
		return
	}
	for i, one := range items {
		//找到余额位置
		itemsMap, ok := addrs[utils.Bytes2string(*one.Addr)]
		if !ok {
			continue
		}
		keyStr := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))
		_, ok = itemsMap[keyStr]
		if !ok {
			continue
		}
		//删除余额
		delete(itemsMap, keyStr)
		//放入冻结的余额里面
		addrsF, ok := this.FrozenBalance[txid]
		if !ok {
			addrsF = make(map[string]map[string]*mining.TxItem)
			this.FrozenBalance[txid] = addrsF
		}
		itemsR, ok := addrsF[utils.Bytes2string(*one.Addr)]
		if !ok {
			itemsR = make(map[string]*mining.TxItem)
			addrsF[utils.Bytes2string(*one.Addr)] = itemsR
		}
		itemsR[keyStr] = items[i]
		//保存冻结高度
		this.FrozenBalanceLockHeight[keyStr] = tx.GetLockHeight()
	}
}

/*
	解冻回滚冻结的交易
*/
func (this *TokenTxItemManager) Unfrozen(blockHeight uint64) {
	this.lock.Lock()
	// engine.Log.Info("开始回滚冻结的交易")
	for txidStr, addrs := range this.FrozenBalance {
		for addr, items := range addrs {
			for itemKey, item := range items {
				lockHeight, ok := this.FrozenBalanceLockHeight[itemKey]

				// keyStr := item.GetTxidStr() + "_" + strconv.Itoa(int(item.OutIndex))
				// lockHeight, ok := this.FrozenBalanceLockHeight[keyStr]
				// engine.Log.Info("开始回滚冻结的交易 判断 %d %d", blockHeight, lockHeight)
				if lockHeight > blockHeight {
					continue
				}
				delete(this.FrozenBalanceLockHeight, itemKey)
				//开始回滚交易
				//添加到余额列表
				addrsB, ok := this.NotSpentBalanceFotAddrKey[txidStr]
				if !ok {
					addrsB = make(map[string]map[string]*mining.TxItem)
					this.NotSpentBalanceFotAddrKey[txidStr] = addrsB
				}
				itemsB, ok := addrsB[addr]
				if !ok {
					itemsB = make(map[string]*mining.TxItem)
					addrsB[addr] = itemsB
				}
				itemsB[itemKey] = item
				//删除锁定的交易
				delete(items, itemKey)
			}
		}
	}

	for txidStr, addrs := range this.BalanceLockup {
		for addr, items := range addrs {
			for itemKey, item := range items {
				lockHeight, ok := this.BalanceLockupLockHeight[itemKey]

				// keyStr := item.GetTxidStr() + "_" + strconv.Itoa(int(item.OutIndex))
				// lockHeight, ok := this.FrozenBalanceLockHeight[keyStr]
				// engine.Log.Info("开始回滚冻结的交易 判断 %d %d", blockHeight, lockHeight)
				if lockHeight > blockHeight {
					continue
				}
				delete(this.BalanceLockupLockHeight, itemKey)

				//添加到余额列表
				addrsB, ok := this.NotSpentBalanceFotAddrKey[txidStr]
				if !ok {
					addrsB = make(map[string]map[string]*mining.TxItem)
					this.NotSpentBalanceFotAddrKey[txidStr] = addrsB
				}
				itemsB, ok := addrsB[addr]
				if !ok {
					itemsB = make(map[string]*mining.TxItem)
					addrsB[addr] = itemsB
				}
				itemsB[itemKey] = item
				//删除锁定的交易
				delete(items, itemKey)
			}
		}
	}
	this.lock.Unlock()
}

/*
	创建一个txitem管理器
*/
func NewTokenTxItemManager() *TokenTxItemManager {
	return &TokenTxItemManager{
		lock: new(sync.RWMutex), //
		// TokenInfo:                 make(map[string]*TokenInfo),                           //
		NotSpentBalanceFotAddrKey: make(map[string]map[string]map[string]*mining.TxItem), //
		FrozenBalance:             make(map[string]map[string]map[string]*mining.TxItem), //保存冻结各个地址的余额，key:string=收款地址;value:*Balance=收益列表;
		FrozenBalanceLockHeight:   make(map[string]uint64),                               //
		BalanceLockup:             make(map[string]map[string]map[string]*mining.TxItem), //
		BalanceLockupLockHeight:   make(map[string]uint64),                               //
	}
}

//----------------------------------------------

var tokenManager = NewTokenTxItemManager()

/*
	添加token的收益
*/
func CountToken(addItemCount []TokenTxItemCount, subItem []TokenTxSubItems) {
	tokenManager.CountTxItem(addItemCount, subItem)
}

/*
	通过地址查询余额
*/
func FindTokenBalanceForTxid(txidbs []byte) (map[string]uint64, map[string]uint64, map[string]uint64) {
	return tokenManager.FindTokenBalanceForTxid(txidbs)
}

/*
	通过地址查询余额
*/
func FindBalanceByAddr(txid string, addr string) *TokenBalance {
	return tokenManager.FindBalanceByAddr(txid, addr)
}

/*
	查询所有token余额
*/
func FindTokenBalanceForAll() []TokenBalance {
	tbs := tokenManager.FindTokenBalanceForAll()
	for i, one := range tbs {
		tokeninfo, err := FindTokenInfo(one.TokenId)
		if err != nil {
			continue
		}
		tbs[i].Name = tokeninfo.Name
		tbs[i].Symbol = tokeninfo.Symbol
		tbs[i].Supply = tokeninfo.Supply
	}
	return tbs
}

/*
	冻结token
*/
func FrozenToken(txid []byte, items []*mining.TxItem, tx mining.TxItr) {
	tokenManager.FrozenToken(txid, items, tx)
}

/*
	解冻回滚token
*/
func UnfrozenToken(blockHeight uint64) {
	tokenManager.Unfrozen(blockHeight)
}

/*
	获得可以支付的TxItem
*/
func GetReadyPayToken(txid []byte, srcAddr crypto.AddressCoin, amount uint64) (total uint64, txItems []*mining.TxItem) {

	var tis mining.TxItemSort = tokenManager.GetReadyPayToken(txid, srcAddr)

	if len(tis) <= 0 {
		return
	}

	sort.Sort(&tis)

	total = uint64(0)
	items := make([]*mining.TxItem, 0)
	if tis[0].Value >= amount {
		item := tis[0]
		for i, one := range tis {
			if one.Value < amount {
				break
			} else if one.Value == amount {
				item = tis[i]
				break
			} else {
				item = tis[i]
			}
		}
		items = append(items, item)
		total = item.Value
	} else {
		for i, one := range tis {
			items = append(items, tis[i])
			total = total + one.Value
			if total >= amount {
				break
			}
		}
	}
	return total, items

}

type TokenBalance struct {
	TokenId       []byte //
	Name          string //名称（全称）
	Symbol        string //单位
	Supply        uint64 //发行总量
	Balance       uint64 //可用余额
	BalanceFrozen uint64 //冻结的余额
	BalanceLockup uint64 //锁仓的余额
}

type TokenBalanceVO struct {
	TokenId       string //
	Name          string //名称（全称）
	Symbol        string //单位
	Supply        uint64 //发行总量
	Balance       uint64 //可用余额
	BalanceFrozen uint64 //冻结的余额
	BalanceLockup uint64 //锁仓的余额
}
