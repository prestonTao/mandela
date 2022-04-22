package mining

import (
	_ "mandela/core/engine"
	"mandela/core/utils"
	"sync"
)

var TxCache *Cache

func init() {
	TxCache = &Cache{
		txitemLock:         new(sync.RWMutex),         //
		txitem:             utils.NewCache(40000),     //保存txitem中的交易引用计数。key:string=交易id;value:*sync.Map=交易输出引用;//40000
		txCacheLock:        new(sync.RWMutex),         //
		txCache:            utils.NewCache(40000 * 2), //保存txitem中的交易引用计数。key:string=交易id;value:*sync.Map=交易输出引用;//40000
		blockHeadCacheLock: new(sync.RWMutex),
		blockHeadCache:     utils.NewCache(100), //区块头缓存。key:string=区块hash;value:*BlockHead=区块内容;
	}
}

type Cache struct {
	txitemLock         *sync.RWMutex //
	txitem             *utils.Cache  //保存txitem中的交易引用计数。key:string=交易id;value:*sync.Map=交易输出引用;
	txCacheLock        *sync.RWMutex //
	txCache            *utils.Cache  //保存交易引用计数。key:string=交易id;value:*sync.Map=交易输出引用;
	blockHeadCacheLock *sync.RWMutex //
	blockHeadCache     *utils.Cache  //区块头缓存。key:string=区块hash;value:*BlockHead=区块内容;
}

// type TxIndexAndHeight struct {
// 	txid  string //交易id
// 	index uint64 //输出索引号
// }

/*
	检查一个交易的引用计数，计数为0则删除这个交易
*/
// func (this *Cache) CheckCleanKey(key string) {
// 	//检查itemRef
// 	v, ok := this.txitemRef.Load(key)
// 	if ok {
// 		ref := v.(*sync.Map)
// 		//如果引用计数为0，则删除这个交易
// 		have := false
// 		ref.Range(func(k, v interface{}) bool {
// 			have = true
// 			return false
// 		})
// 		if have {
// 			return
// 		}
// 		this.txitemRef.Delete(key)
// 	}
// 	//检查txCacheRef
// 	v, ok = this.txCacheRef.Load(key)
// 	if ok {
// 		ref := v.(*sync.Map)
// 		//如果引用计数为0，则删除这个交易
// 		have := false
// 		ref.Range(func(k, v interface{}) bool {
// 			have = true
// 			return false
// 		})
// 		if have {
// 			return
// 		}
// 		this.txCacheRef.Delete(key)
// 	}
// 	this.txs.Delete(key)
// }

func (this *Cache) AddTxInTxItem(keybs []byte, txItr TxItr) {
	// engine.Log.Info("添加TxItem缓存 %s", key)
	key := utils.Bytes2string(keybs)
	this.txitemLock.Lock()
	_, ok := this.txitem.Get(key)
	if !ok {
		this.txitem.Add(key, txItr)
	}
	this.txitemLock.Unlock()
	// this.txs.LoadOrStore(key, txItr)
	// v, ok := this.txitemRef.Load(key)
	// if ok {
	// 	ref := v.(*sync.Map)
	// 	ref.Store(index, index)
	// } else {
	// 	ref := new(sync.Map)
	// 	ref.Store(index, index)
	// 	this.txitemRef.Store(key, ref)
	// }
}

// func (this *Cache) RemoveTxInTxItem(key string, index uint64) {
// 	// engine.Log.Info("删除TxItem缓存 %s %d", key, index)
// 	v, ok := this.txitemRef.Load(key)
// 	if ok {
// 		ref := v.(*sync.Map)
// 		ref.Delete(index)
// 	}
// 	this.CheckCleanKey(key)
// }

func (this *Cache) FindTxInCache(keybs []byte) (TxItr, bool) {
	// engine.Log.Info("查找缓存 %s", key)
	key := utils.Bytes2string(keybs)
	this.txCacheLock.RLock()
	v, ok := this.txCache.Get(key)
	this.txCacheLock.RUnlock()
	if !ok {
		this.txitemLock.RLock()
		v, ok = this.txitem.Get(key)
		this.txitemLock.RUnlock()
	}
	if ok {
		tx := v.(TxItr)
		return tx, ok
	}
	return nil, ok
}

/*
	缓存中有就刷新，没有不做任何操作
*/
// func (this *Cache) FlashTxInTxItem(key string, txItr TxItr) {
// 	_, ok := txInTxitem.Load(key)
// 	if !ok {
// 		//没有不做任何操作
// 		return
// 	}
// 	//有就刷新
// 	AddTxInTxItem(key, txItr)
// }

func (this *Cache) AddTxInCache(keybs []byte, txItr TxItr) {
	if txItr == nil {
		return
	}
	key := utils.Bytes2string(keybs)
	this.txCacheLock.Lock()
	_, ok := this.txCache.Get(key)
	if !ok {
		this.txCache.Add(key, txItr)
	}
	this.txCacheLock.Unlock()
}

func (this *Cache) FlashTxInCache(keybs []byte, txItr TxItr) {
	if txItr == nil {
		return
	}
	key := utils.Bytes2string(keybs)
	this.txCacheLock.Lock()
	this.txCache.Add(key, txItr)
	this.txCacheLock.Unlock()
}

/*
	删除缓存中的交易
*/
// func (this *Cache) RemoveTxInCache(key string, index uint64) {
// 	v, ok := this.txCacheRef.Load(key)
// 	if ok {
// 		ref := v.(*sync.Map)
// 		ref.Delete(index)
// 	}
// 	this.CheckCleanKey(key)

// }

/*
	将缓存的交易添加指定高度引用
*/
// func (this *Cache) TransferTxInCache(key string, index, height uint64) {
// 	//添加区块高度的交易
// 	var heightMap *sync.Map
// 	v, ok := this.txCacheHeightRef.Load(height)
// 	// engine.Log.Info("%b %+v", ok, v)
// 	if !ok {
// 		heightMap = new(sync.Map)
// 		this.txCacheHeightRef.Store(height, heightMap)
// 	} else {
// 		heightMap = v.(*sync.Map)
// 	}
// 	heightMap.Store(key, index)

// 	//删除未打包的缓存引用
// 	// v, ok := this.txCacheRef.Load(key)
// 	// if ok {
// 	// 	ref := v.(*sync.Map)
// 	// 	ref.Delete(index)
// 	// }

// }

/*
	删除小于等于某个区块高度的所有交易缓存
*/
// func (this *Cache) RemoveHeightTxInCache(height uint64) {
// 	this.txCacheHeightRef.Range(func(k, v interface{}) bool {
// 		if height < k.(uint64) {
// 			return true
// 		}
// 		heightMap := v.(*sync.Map)
// 		heightMap.Range(func(k, v interface{}) bool {
// 			txid := k.(string)
// 			voutIndex := v.(uint64)
// 			this.RemoveTxInCache(txid, voutIndex)
// 			return true
// 		})
// 		this.txCacheHeightRef.Delete(k)
// 		return true
// 	})
// }

// func (this *Cache) FindTxInCache(key string) (TxItr, bool) {
// 	v, ok := txCache.Load(key)
// 	if ok {
// 		tx := v.(TxItr)
// 		return tx, ok
// 	}
// 	return nil, ok
// }

/*
	缓存中有就刷新，没有不做任何操作
*/
// func (this *Cache) FlashTxInCache(key string, txItr TxItr) {
// 	v, ok := this.txCache.Get(key)
// 	if !ok {
// 		this.txCache.Add(key, txItr)
// 		//没有不做任何操作
// 		return
// 	}
// 	//有就刷新
// 	txi :=v.(TxItr)
// 	txi.set
// 	this.txs.Store(key, txItr)
// 	// AddTxInCache(key, txItr)
// }

// func (this *Cache)

func (this *Cache) AddBlockHeadCache(keybs []byte, bh *BlockHead) {
	key := utils.Bytes2string(keybs)
	this.blockHeadCacheLock.Lock()
	_, ok := this.blockHeadCache.Get(key)
	if !ok {
		this.blockHeadCache.Add(key, bh)
	}
	this.blockHeadCacheLock.Unlock()
}

// func (this *Cache) RemoveBlockHeadCache(key string) {
// 	this.blockHeadCache.Delete(key)
// }

func (this *Cache) FindBlockHeadCache(keybs []byte) (*BlockHead, bool) {
	key := utils.Bytes2string(keybs)
	this.blockHeadCacheLock.RLock()
	v, ok := this.blockHeadCache.Get(key)
	this.blockHeadCacheLock.RUnlock()
	if ok {
		tx := v.(*BlockHead)
		return tx, ok
	}
	return nil, ok
}

/*
	缓存中有就刷新，没有不做任何操作
*/
func (this *Cache) FlashBlockHeadCache(keybs []byte, bh *BlockHead) {
	key := utils.Bytes2string(keybs)
	this.blockHeadCacheLock.RLock()
	v, ok := this.blockHeadCache.Get(key)
	this.blockHeadCacheLock.RUnlock()
	if !ok {
		this.AddBlockHeadCache(keybs, bh)
		//没有不做任何操作
		return
	}
	//有就刷新
	bhOld := v.(*BlockHead)
	bhOld.Nextblockhash = bh.Nextblockhash
}
