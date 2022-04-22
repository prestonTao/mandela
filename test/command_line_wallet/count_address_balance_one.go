package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	_ "mandela/chain_witness_vote/mining/tx_name_in"
	_ "mandela/chain_witness_vote/mining/tx_name_out"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/sqlite3_db"
	"strconv"

	// "mandela/test/map_chain_tools/sqlite"
	"bytes"
	"encoding/hex"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	txItem_status_notSpent = int32(0) //未花费的交易余额，可以正常支付
	txItem_status_frozen   = int32(1) //锁仓,区块达到指定高度才能使用
	txItem_status_lock     = int32(2) //冻结高度，指定高度还未上链，则转为未花费的交易
)

/*
	统计某个地址余额及交易日志，统计一个地址，速度会快一些
*/
var CountAddrCoin = crypto.AddressFromB58String("TESTHj94iJaS245y5WhCzXPrEtRqhkTVBj8Dd4")

// var CountAddrCoin = crypto.AddressFromB58String("TESTGn11Ka45vZhzpdv5qyrsniYnnCEu8wiGg4")
var heightStart = uint64(0)
var heightEnd = uint64(516831) //解锁高度

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

	// path :=
	mNotSpend, tim, timOld := CountAddrBalance(filepath.Join("wallet", "data"))

	PrintBalanceMap(mNotSpend, tim, timOld)

	//等待3秒钟再关闭，让sqlite数据库处理完
	time.Sleep(time.Second * 3)
}

/*
	统计链上所有地址余额
*/
func CountAddrBalance(dir string) (map[string]*map[string]*mining.TxItem, *mining.TxItemManager, *mining.TxItemManagerOld) {

	db.InitDB(config.DB_path, config.DB_path_temp)
	sqlite3_db.Init()
	// timOld := mining.NewTxItemManagerOld()
	tim := mining.NewTxItemManager()

	NotSpentBalance := make(map[string]*map[string]*mining.TxItem) //保存各个状态的txitem，解锁、冻结等只是状态的改变。key:string=收款地址;value:*sync.Map(key:string=[txid]_[voutIndex];value:*TxItem=TxItem;)=;

	// balanceMap := make(map[string]uint64)
	// balanceNotSpend := make(map[string]uint64)
	// balanceFrozen := make(map[string]uint64)

	// db := utils.CreateLevelDB(dir)

	// db.InitDB(dir)
	beforBlockHash, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		engine.Log.Info("111 查询起始块id错误 " + err.Error())
		return nil, nil, nil
	}

	// balanceTotal := uint64(0)
	// rewardTotal := uint64(0)

	// beforGroupHeight := uint64(0)

	for beforBlockHash != nil {

		bhvo, err := mining.LoadBlockHeadVOByHash(beforBlockHash)
		if err != nil {
			engine.Log.Info("查询第 个块错误:%s", err.Error())
			return nil, nil, nil
		}

		engine.Log.Info("第%d个块 -----------------------------------\n%s\nnext区块个数", bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Hash))

		for _, txBase := range bhvo.Txs {

			// bs, _ := json.Marshal(txBase.GetVOJSON())
			// engine.Log.Info("%s", string(bs))

			//如果是区块奖励，则计算奖励总和
			if txBase.Class() == config.Wallet_tx_type_mining {
				// rewardTotal := uint64(0)
				// for _, one := range *txBase.GetVout() {
				// 	rewardTotal += one.Value
				// }
				// engine.Log.Info("区块奖励 %d", rewardTotal)
			} else {
				//计算所有交易中的地址余额
				for _, one := range *txBase.GetVin() {

					addrCoin := crypto.BuildAddr(config.AddrPre, one.Puk)
					ok := bytes.Equal(addrCoin, CountAddrCoin)
					if !ok {
						continue
					}

					// addrCoin := crypto.BuildAddr(config.AddrPre, one.Puk)
					addrStr := utils.Bytes2string(addrCoin)
					// value, ok := balanceMap[addrStr]
					// if !ok {
					// 	panic("找不到这个地址的记录" + addrCoin.B58String())
					// 	return nil, nil, nil, nil
					// }

					//找上一个交易
					// txBase, err := mining.LoadTxBase(one.Txid)
					// if err != nil {
					// 	panic("error:查询交易错误2")
					// 	return nil, nil, nil
					// }

					// vout := (*txBase.GetVout())[one.Vout]
					//判定锁定高度能不能使用
					// free := mining.CheckFrozenHeightFree(vout.FrozenHeight-1, bhvo.BH.Height, bhvo.BH.Time)
					// if free {

					// } else {
					// 	engine.Log.Error("这个交易被锁定，不能使用:%s %d %d %d %d", hex.EncodeToString(one.Txid), one.Vout, vout.FrozenHeight, bhvo.BH.Height, bhvo.BH.Time)
					// 	// panic("这个交易不能使用")
					// 	// return nil, nil, nil
					// }

					key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.Vout))

					items, ok := NotSpentBalance[addrStr]
					if !ok {
						panic("这个地址不存在")
						return nil, nil, nil
					}
					_, ok = (*items)[key]
					if !ok {
						engine.Log.Info("%s %d", hex.EncodeToString(one.Txid), one.Vout)
						// panic("这个txitem不存在")
						// return nil, nil, nil
					}
					delete(*items, key)
					// engine.Log.Info("%s %d del item:%d", hex.EncodeToString(one.Txid), one.Vout, item.Value)
					// engine.Log.Info("修改余额 %s %d %d", addr, value, vout.Value)
					// balanceTotal = balanceTotal - vout.Value

				}
			}

			//计算输出中的余额
			for voutIndex, _ := range *txBase.GetVout() {

				//见证人押金和投票押金不计入余额
				if txBase.Class() == config.Wallet_tx_type_deposit_in || txBase.Class() == config.Wallet_tx_type_vote_in || txBase.Class() == config.Wallet_tx_type_account {
					if voutIndex == 0 {
						continue
					}
				}
				one := (*txBase.GetVout())[voutIndex]

				ok := bytes.Equal(one.Address, CountAddrCoin)
				if !ok {
					continue
				}

				addrStr := utils.Bytes2string(one.Address)
				key := utils.Bytes2string(*txBase.GetHash()) + "_" + strconv.Itoa(int(voutIndex))

				txItem := mining.TxItem{
					Addr: &one.Address, //  &vout.Address,
					// AddrStr: vout.GetAddrStr(),                      //
					Value: one.Value,         //余额
					Txid:  *txBase.GetHash(), //交易id
					// TxidStr:      txHashStr,                              //
					VoutIndex:    uint64(voutIndex), //交易输出index，从0开始
					Height:       bhvo.BH.Height,    //
					LockupHeight: one.FrozenHeight,  //锁仓高度
				}

				items, ok := NotSpentBalance[addrStr]
				if !ok {
					itemsTemp := make(map[string]*mining.TxItem)
					items = &itemsTemp
					NotSpentBalance[addrStr] = items
				}
				(*items)[key] = &txItem
				// if mining.CheckFrozenHeightFree(txItem.LockupHeight, bhvo.BH.Height, bhvo.BH.Time) {
				// 	engine.Log.Info("%s %d add item:%d", hex.EncodeToString(*txBase.GetHash()), voutIndex, txItem.Value)
				// }
				// balanceTotal = balanceTotal + one.Value

			}

		}

		//开始解锁
		for _, items := range NotSpentBalance {
			for _, item := range *items {
				if mining.CheckFrozenHeightFree(item.LockupHeight, bhvo.BH.Height, bhvo.BH.Time) {
					// engine.Log.Info("txItem_status_notSpent:%+v", item)
					item.Status = txItem_status_notSpent
				} else {
					item.Status = txItem_status_frozen
					// engine.Log.Info("txItem_status_frozen:%+v", item)
				}
			}
		}

		// engine.Log.Info("下一个块hash %s \n", hex.EncodeToString(bh.Nextblockhash))

		if bhvo.BH.Nextblockhash != nil && len(bhvo.BH.Nextblockhash) > 0 {
			beforBlockHash = &bhvo.BH.Nextblockhash
		} else {
			beforBlockHash = nil
		}

		tic := CountBalances(bhvo)
		tim.CountTxItem(tic, bhvo.BH.Height, bhvo.BH.Time)
		// timOld.CountTxItem(tic, bhvo.BH.Height, bhvo.BH.Time)
		tim.Unfrozen(bhvo.BH.Height, bhvo.BH.Time)
		// timOld.Unfrozen(bhvo.BH.Height, bhvo.BH.Time)

		// engine.Log.Info("------------------------------")
		//开始对比余额是否一样
		timNotSpendTotal := uint64(0)
		items := tim.FindBalanceAll()
		for _, one := range items {
			engine.Log.Info("txItem_status_notSpent:%+v", one)
			timNotSpendTotal += one.Value
		}
		// engine.Log.Info("------------------------------")
		thatNotSpendTotal := uint64(0)
		for _, items := range NotSpentBalance {
			for _, item := range *items {
				if item.Status == txItem_status_notSpent {
					engine.Log.Info("txItem_status_notSpent:%+v", item)
					thatNotSpendTotal += item.Value
				}
			}
		}
		if timNotSpendTotal != thatNotSpendTotal {
			engine.Log.Info("tim:%d that:%d", timNotSpendTotal, thatNotSpendTotal)

			engine.Log.Info("------------------------------")
			//开始对比余额是否一样
			timNotSpendTotal := uint64(0)
			items := tim.FindBalanceAll()
			for _, one := range items {
				engine.Log.Info("txItem_status_notSpent:%+v", one)
				timNotSpendTotal += one.Value
			}
			engine.Log.Info("------------------------------")
			thatNotSpendTotal := uint64(0)
			for _, items := range NotSpentBalance {
				for _, item := range *items {
					if item.Status == txItem_status_notSpent {
						engine.Log.Info("txItem_status_notSpent:%+v", item)
						thatNotSpendTotal += item.Value
					}
				}
			}

			panic("余额不相等")
		}

	}
	// engine.Log.Info("挖矿奖励总量:%d 链上共流通币数量:%d", rewardTotal, balanceTotal)

	// tim.FindBalanceAll()

	return NotSpentBalance, tim, nil
}

func PrintBalanceMap(notSpentBalance map[string]*map[string]*mining.TxItem, tim *mining.TxItemManager, timOld *mining.TxItemManagerOld) {
	// if balances == nil {
	// 	engine.Log.Info("打印结果为空")
	// 	return
	// }
	// for k, v := range balances {
	// 	addrCoin := crypto.AddressCoin([]byte(k))
	// 	engine.Log.Info("%s %d", addrCoin.B58String(), v)
	// }

	for addrStr, items := range notSpentBalance {
		addrCoin := crypto.AddressCoin([]byte(addrStr))

		notSpend := uint64(0)
		frozen := uint64(0)
		for _, item := range *items {
			if item.Status == txItem_status_frozen {
				notSpend += item.Value
			} else {
				frozen += item.Value
			}
		}

		//查询对比
		notSpend2 := uint64(0)
		frozen2 := uint64(0)
		items := tim.FindBalance(addrCoin)
		for _, one := range items {
			notSpend2 += one.Value
		}
		items = tim.FindBalanceFrozen(addrCoin)
		for _, one := range items {
			frozen2 += one.Value
		}

		engine.Log.Info("地址:%s 可用:%d 锁定:%d  可用2:%d 锁定2:%d", addrCoin.B58String(), notSpend, frozen, notSpend2, frozen2)
	}

	// engine.Log.Info("打印已经解锁的余额")

	// if balanceNotSpend == nil {
	// 	engine.Log.Info("打印结果为空")
	// 	return
	// }
	// for k, v := range balanceNotSpend {
	// 	addrCoin := crypto.AddressCoin([]byte(k))
	// 	engine.Log.Info("%s %d", addrCoin.B58String(), v)
	// }

	// engine.Log.Info("打印 TxItemManager")
	// for k, v := range balances {
	// 	addrOne := crypto.AddressCoin([]byte(k))
	// 	notSpent1 := uint64(0)
	// 	txItems := tim.FindBalance(addrOne)
	// 	for _, one := range txItems {
	// 		notSpent1 += one.Value
	// 	}
	// 	notSpent2 := uint64(0)
	// 	if timOld != nil {
	// 		txItems = timOld.FindBalance(addrOne)
	// 		for _, one := range txItems {
	// 			notSpent2 += one.Value
	// 		}
	// 	}

	// 	notF1 := uint64(0)
	// 	txItems = tim.FindBalanceFrozen(addrOne)
	// 	for _, one := range txItems {
	// 		notF1 += one.Value
	// 	}
	// 	notF2 := uint64(0)
	// 	if timOld != nil {
	// 		txItems = timOld.FindBalanceFrozen(addrOne)
	// 		for _, one := range txItems {
	// 			notF2 += one.Value
	// 		}
	// 	}

	// 	engine.Log.Info("地址:%s 总余额:%d 可用余额1:%d 可用余额2:%d 锁仓余额1:%d 锁仓余额2:%d", addrOne.B58String(), v, notSpent1, notSpent2, notF1, notF2)
	// }

}

// var CountAddrCoin = crypto.AddressFromB58String("TESTArGrVxQwYQDCoxWVyw9J5xGVWYZVa6ZBE4")

func CountBalances(bhvo *mining.BlockHeadVO) mining.TxItemCount {

	//将txitem集中起来，一次性添加
	itemCount := mining.TxItemCount{
		Additems: make([]*mining.TxItem, 0),
		SubItems: make([]*mining.TxSubItems, 0),
		// deleteKey: make([]string, 0),
	}
	itemsChan := make(chan *mining.TxItemCount, len(bhvo.Txs))
	wg := new(sync.WaitGroup)
	wg.Add(len(bhvo.Txs))
	go func() {
		for i := 0; i < len(bhvo.Txs); i++ {
			one := <-itemsChan
			if one != nil {
				itemCount.Additems = append(itemCount.Additems, one.Additems...)
				itemCount.SubItems = append(itemCount.SubItems, one.SubItems...)
				// itemCount.deleteKey = append(itemCount.deleteKey, one.deleteKey...)
			}
			wg.Done()
		}
	}()

	//查询排除的交易
	// excludeTx := make([]config.ExcludeTx, 0)
	// for i, one := range config.Exclude_Tx {
	// 	if bhvo.BH.Height == one.Height {
	// 		excludeTx = append(excludeTx, config.Exclude_Tx[i])
	// 	}
	// }

	NumCPUTokenChan := make(chan bool, runtime.NumCPU()*6)
	for _, txItr := range bhvo.Txs {
		//排除的交易不统计
		// for _, two := range excludeTx {
		// 	if bytes.Equal(two.TxByte, *txItr.GetHash()) {
		// 		continue
		// 	}
		// }
		go countBalancesTxOne(txItr, bhvo.BH.Height, NumCPUTokenChan, itemsChan)
	}

	wg.Wait()
	// start := time.Now()
	// this.notspentBalance.CountTxItem(itemCount)
	// engine.Log.Info("统计交易 耗时 %s", time.Now().Sub(start))
	return itemCount
}

/*
	统计单个交易余额，方便异步统计
*/
func countBalancesTxOne(txItr mining.TxItr, height uint64, tokenCPU chan bool, itemChan chan *mining.TxItemCount) {
	tokenCPU <- false
	// start := time.Now()
	itemCount := mining.TxItemCount{
		Additems: make([]*mining.TxItem, 0),
		SubItems: make([]*mining.TxSubItems, 0),
	}
	// start := time.Now()
	txItr.BuildHash()
	// txHashStr := txItr.GetHashStr()
	//将之前的UTXO标记为已经使用，余额中减去。
	for _, vin := range *txItr.GetVin() {

		// engine.Log.Info("统计余额 111 %d", txItr.Class())
		//是区块奖励
		if txItr.Class() == config.Wallet_tx_type_mining {
			continue
		}
		// engine.Log.Info("查看vin中的状态 %d", vin.PukIsSelf)
		addrOne := crypto.BuildAddr("TEST", vin.Puk)
		ok := bytes.Equal(addrOne, CountAddrCoin)
		// ok := vin.CheckIsSelf()
		if !ok {
			continue
		}
		// engine.Log.Info("统单易1耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
		//查找这个地址的余额列表，没有则创建一个
		itemCount.SubItems = append(itemCount.SubItems, &mining.TxSubItems{
			Txid:      vin.Txid, //utils.Bytes2string(vin.Txid), //  vin.GetTxidStr(),
			VoutIndex: vin.Vout,
			Addr:      *vin.GetPukToAddr(), // utils.Bytes2string(*vin.GetPukToAddr()), // vin.GetPukToAddrStr(),
		})

	}
	// engine.Log.Info("统单易3耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	// txCtrl := mining.GetTransactionCtrl(txItr.Class())
	// if txCtrl != nil {

	// 	txCtrl.SyncCount()
	// 	itemChan <- &itemCount
	// 	<-tokenCPU
	// 	return
	// }
	// engine.Log.Info("统计单个交易 444 耗时 %s", time.Now().Sub(start))

	//生成新的UTXO收益，保存到列表中
	for voutIndex, vout := range *txItr.GetVout() {

		//找出需要统计余额的地址

		//和自己无关的地址
		ok := bytes.Equal(vout.Address, CountAddrCoin)
		// // ok := vout.CheckIsSelf()
		// addrInfo, ok := keystore.FindAddress(vout.Address)
		if !ok {
			continue
		}
		//见证人押金和投票押金不计入余额
		if txItr.Class() == config.Wallet_tx_type_deposit_in || txItr.Class() == config.Wallet_tx_type_vote_in || txItr.Class() == config.Wallet_tx_type_account {
			if voutIndex == 0 {
				continue
			}
		}

		// engine.Log.Info("统单易5耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
		txItem := mining.TxItem{
			Addr: &(*txItr.GetVout())[voutIndex].Address, //  &vout.Address,
			// AddrStr: vout.GetAddrStr(),                      //
			Value: vout.Value,       //余额
			Txid:  *txItr.GetHash(), //交易id
			// TxidStr:      txHashStr,                              //
			VoutIndex:    uint64(voutIndex), //交易输出index，从0开始
			Height:       height,            //
			LockupHeight: vout.FrozenHeight, //锁仓高度
		}

		//计入余额列表
		// this.notspentBalance.AddTxItem(txItem)
		itemCount.Additems = append(itemCount.Additems, &txItem)

		//保存到缓存
		// engine.Log.Info("开始统计交易余额 区块高度 %d 保存到缓存", bhvo.BH.Height)
		// TxCache.AddTxInTxItem(txHashStr, txItr)
		// TxCache.AddTxInTxItem(*txItr.GetHash(), txItr)

	}
	// engine.Log.Info("统计余额及奖励 101010 耗时 %s", time.Now().Sub(start))
	itemChan <- &itemCount
	// engine.Log.Info("统单易6耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	<-tokenCPU
}
