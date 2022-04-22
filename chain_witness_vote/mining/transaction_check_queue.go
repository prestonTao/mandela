package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"encoding/hex"
	"runtime"
)

var checkTxQueue = make(chan TxItr, 100000)

func init() {
	go startCheckQueue()
}

func startCheckQueue() {
	utils.Go(func() {
		NumCPUTokenChan := make(chan bool, config.CPUNUM)
		for txItr := range checkTxQueue {
			NumCPUTokenChan <- false
			go checkMulticastTransaction(txItr, NumCPUTokenChan)
		}
	})
}

/*
	验证广播消息
*/
func checkMulticastTransaction(txbase TxItr, tokenCPU chan bool) {
	defer func() {
		<-tokenCPU
	}()
	// if !GetSyncFinish() {
	// 	// engine.Log.Warn("未同步完成则无法验证交易合法性")
	// 	return
	// }
	if len(*txbase.GetVin()) > config.Mining_pay_vin_max {
		//交易太大了
		engine.Log.Warn(config.ERROR_pay_vin_too_much.Error())
		return
	}
	//验证交易
	if err := txbase.CheckLockHeight(GetHighestBlock()); err != nil {
		// engine.Log.Warn("验证交易锁定高度失败")
		engine.Log.Warn("Failed to verify transaction lock height")
		return
	}
	// txbase.CheckFrozenHeight(GetHighestBlock())

	//加载相关交易到缓存
	keys := make(map[string]uint64, 0) //记录加载了哪些交易到缓存
	for _, one := range *txbase.GetVin() {
		//已经有了就不用重复查询了
		if _, ok := TxCache.FindTxInCache(one.Txid); !ok {
			continue
		}

		// txItr, err := FindTxBase(one.Txid)
		txItr, err := LoadTxBase(one.Txid)
		if err != nil {
			return
		}
		TxCache.AddTxInCache(one.Txid, txItr)
		// key := utils.Bytes2string(one.Txid) //one.GetTxidStr()
		keys[utils.Bytes2string(one.Txid)] = one.Vout
		// keys = append(keys, key)
	}

	// bs, _ := txbase.Json()
	// engine.Log.Info("交易\n%s", string(*bs))

	if GetHighestBlock() > config.Mining_block_start_height+config.Mining_block_start_height_jump {
		if err := txbase.Check(); err != nil {
			//交易不合法，则不发送出去
			//验证未通过，删除缓存
			// for k, v := range keys {
			// 	TxCache.RemoveTxInCache(k, v)
			// }
			runtime.GC()
			engine.Log.Warn("Failed to verify transaction signature %s %s", hex.EncodeToString(*txbase.GetHash()), err.Error())
			return
		}
		runtime.GC()
	}

	//判断是否有重复交易，并且检查是否有无效区块的交易的标记
	if db.LevelDB.CheckHashExist(*txbase.GetHash()) && !db.LevelDB.CheckHashExist(config.BuildTxNotImport(*txbase.GetHash())) {
		//验证未通过，删除缓存
		// for k, v := range keys {
		// 	TxCache.RemoveTxInCache(k, v)
		// }
		engine.Log.Warn("Transaction hash collision is the same %s", hex.EncodeToString(*txbase.GetHash()))
		return
	}

	forks.GetLongChain().transactionManager.AddTx(txbase)
}
