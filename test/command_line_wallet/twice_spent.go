package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	_ "mandela/chain_witness_vote/mining/tx_name_in"
	_ "mandela/chain_witness_vote/mining/tx_name_out"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"mandela/sqlite3_db"
	"encoding/hex"
	"path/filepath"
	"runtime"
	"time"
)

func main() {

	go func() {
		for {
			engine.Log.Info("NumGoroutine:%d", runtime.NumGoroutine())
			time.Sleep(time.Minute)
			// log.Error(http.ListenAndServe(":6060", nil))
		}
	}()
	Start()

}

/*
	统计链上所有地址余额
	保存在数据库里面
*/
func Start() {
	CountAddrBalance(filepath.Join("wallet", "data"))

	//等待3秒钟再关闭，让sqlite数据库处理完
	time.Sleep(time.Second * 3)
}

/*
	统计链上所有地址余额
*/
func CountAddrBalance(dir string) (map[string]uint64, map[string]uint64, *mining.TxItemManager, *mining.TxItemManagerOld) {
	sqlite3_db.Init()
	db.InitDB(dir)
	// db.InitDB(dir)
	beforBlockHash, err := db.Find(config.Key_block_start)
	if err != nil {
		engine.Log.Info("111 查询起始块id错误 " + err.Error())
		return nil, nil, nil, nil
	}

	for beforBlockHash != nil {

		bhvo, err := mining.LoadBlockHeadVOByHash(beforBlockHash)
		if err != nil {
			engine.Log.Info("查询第 个块错误:%s", err.Error())
			return nil, nil, nil, nil
		}

		engine.Log.Info("第%d个块 -----------------------------------\n%s\nnext区块个数", bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Hash))

		for _, txBase := range bhvo.Txs {

			//计算所有交易中的地址余额
			if txBase.Class() != config.Wallet_tx_type_mining {
				for _, one := range *txBase.GetVin() {
					key := mining.BuildKeyForUnspentTransaction(one.Txid, one.Vout)
					if !db.CheckHashExist(key) {
						panic("这里有双花")
						return nil, nil, nil, nil
					}

				}
			}

			//计算输出中的余额
			for voutIndex, one := range *txBase.GetVout() {

				bs := utils.Uint64ToBytes(one.FrozenHeight)
				db.Save(mining.BuildKeyForUnspentTransaction(*txBase.GetHash(), uint64(voutIndex)), &bs)

			}

		}

		// engine.Log.Info("下一个块hash %s \n", hex.EncodeToString(bh.Nextblockhash))

		if bhvo.BH.Nextblockhash != nil && len(bhvo.BH.Nextblockhash) > 0 {
			beforBlockHash = &bhvo.BH.Nextblockhash
		} else {
			beforBlockHash = nil
		}

	}

	// tim.FindBalanceAll()

	return nil, nil, nil, nil
}
