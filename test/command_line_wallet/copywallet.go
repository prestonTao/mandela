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
	"mandela/test/map_chain_tools/sqlite"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {
	Start()

}

/*
	统计链上所有地址余额
	保存在数据库里面
*/
func Start() {
	// sqlite.Init()
	path := filepath.Join("copywallet", "data")
	CopyWallet(path)

	//等待3秒钟再关闭，让sqlite数据库处理完
	time.Sleep(time.Second * 3)
}

/*

 */
func CopyWallet(dir string) {

	db.InitDB(dir)
	beforBlockHash, err := db.Find(config.Key_block_start)
	if err != nil {
		engine.Log.Info("111 查询起始块id错误 " + err.Error())
		return nil, nil
	}

	balanceTotal := uint64(0)

	// beforGroupHeight := uint64(0)

	for beforBlockHash != nil {
		bs, err := db.Find(*beforBlockHash)
		if err != nil {
			engine.Log.Info("查询第 个块错误" + err.Error())
			return nil, nil
		}
		bh, err := mining.ParseBlockHead(bs)
		if err != nil {

			engine.Log.Info("查询第 个块错误" + err.Error())
			// fmt.Println(string(*bs))
			engine.Log.Info(string(*bs))
			return nil, nil
		}
		if bh.Nextblockhash == nil {
			engine.Log.Info("第%d个块 -----------------------------------\n%s\nnext区块个数", bh.Height,
				hex.EncodeToString(bh.Hash))
		} else {
			engine.Log.Info("第%d个块 -----------------------------------\n%s\nnext区块个数%d", bh.Height,
				hex.EncodeToString(bh.Hash), len(bh.Nextblockhash))
		}

		bhvo := mining.BlockHeadVO{
			// FromBroadcast   bool       `json:"-"`   //是否来自于广播的区块
			// StaretBlockHash []byte     `json:"sbh"` //创始区块hash
			BH:  bh,                      //区块
			Txs: make([]mining.TxItr, 0), //交易明细
		}

		// txs := make([]string, 0)
		// for _, one := range bh.Tx {
		// 	txs = append(txs, hex.EncodeToString(one))
		// }
		// bhvo := rpc.BlockHeadVO{
		// 	Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		// 	Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		// 	GroupHeight:       bh.GroupHeight,                           //矿工组高度
		// 	GroupHeightGrowth: bh.GroupHeightGrowth,                     //
		// 	Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		// 	Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		// 	NTx:               bh.NTx,                                   //交易数量
		// 	MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		// 	Tx:                txs,                                      //本区块包含的交易id
		// 	Time:              bh.Time,                                  //出块时间，unix时间戳
		// 	Witness:           bh.Witness.B58String(),                   //此块见证人地址
		// 	Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		// }
		// *bs, _ = json.Marshal(bhvo)
		// engine.Log.Info(string(*bs))

		//计算是否跳过了组
		// intervalGroup := bhvo.GroupHeight - beforGroupHeight
		// if intervalGroup > 1 {
		// 	// engine.Log.Warn("跳过了组高度 %d", intervalGroup)
		// }
		// beforGroupHeight = bhvo.GroupHeight

		// for _, one := range bh.Nextblockhash {
		// 	fmt.Println("下一个块hash", hex.EncodeToString(one))
		// }

		for _, one := range bh.Tx {
			tx, err := db.Find(one)
			if err != nil {
				engine.Log.Info("查询第 %d 个块的交易错误"+err.Error(), bh.Height)
				panic("error:查询交易错误")
				return nil, nil
			}
			txBase, err := mining.ParseTxBase(mining.ParseTxClass(one), tx)
			if err != nil {
				engine.Log.Info("解析第 %d 个块的交易错误:%s %s", bh.Height, hex.EncodeToString(one), err.Error())
				panic("error:解析交易错误")
				return nil, nil
			}

			bhvo.Txs = append(bhvo.Txs, txBase)

			// txid := txBase.GetHash()
			// engine.Log.Info("交易id " + string(hex.EncodeToString(*txid)))

			// itr := txBase.GetVOJSON()
			// bs, _ := json.Marshal(itr)
			// engine.Log.Info(string(bs))

			//如果是区块奖励，则计算奖励总和
			if txBase.Class() == config.Wallet_tx_type_mining {
				rewardTotal := uint64(0)
				for _, one := range *txBase.GetVout() {
					rewardTotal += one.Value
				}
				engine.Log.Info("区块奖励 %d", rewardTotal)
			}

			//计算所有交易中的地址余额
			if txBase.Class() != config.Wallet_tx_type_mining {
				for _, one := range *txBase.GetVin() {
					addrCoin := crypto.BuildAddr(config.AddrPre, one.Puk)
					addrStr := utils.Bytes2string(addrCoin)
					value, ok := balanceMap[addrStr]
					if !ok {
						panic("找不到这个地址的记录" + addrCoin.B58String())
						return nil, nil
					}

					// addr := addrCoin.B58String()
					// balance, _ := new(sqlite.Balance).FindByAddr(addr)
					// if balance == nil {
					// 	panic("找不到这个地址的记录" + addr)
					// 	return
					// }
					tx, err := db.Find(one.Txid)
					if err != nil {
						// engine.Log.Info("查询第 %d 个块的交易错误"+err.Error(), bh.Height)
						panic("error:查询交易错误2")
						return nil, nil
					}
					txBase, err := mining.ParseTxBase(mining.ParseTxClass(one.Txid), tx)
					if err != nil {
						// engine.Log.Info("解析第 %d 个块的交易错误"+err.Error(), bh.Height)
						// fmt.Println("解析第", bh.Height, "个块的交易错误", err)
						panic("error:解析交易错误2")
						return nil, nil
					}
					vout := (*txBase.GetVout())[one.Vout]
					// engine.Log.Info("修改余额 %s %d %d", addr, value, vout.Value)
					balanceTotal = balanceTotal - vout.Value
					balanceMap[addrStr] = value - vout.Value
					// err = new(sqlite.Balance).UpdateBalance(addr, balance.Balance-vout.Value)
					// if err != nil {
					// 	engine.Log.Info("修改余额错误 " + err.Error())
					// 	panic(err)
					// 	return
					// }
				}
			}

			//计算输出中的余额
			for _, one := range *txBase.GetVout() {
				balanceTotal = balanceTotal + one.Value
				addrStr := utils.Bytes2string(one.Address)
				value, ok := balanceMap[addrStr]
				if !ok {
					balanceMap[addrStr] = one.Value
				} else {
					balanceMap[addrStr] = value + one.Value
				}

				//计算可用余额
				now := time.Now().Unix()
				if uint64(now) > one.FrozenHeight {
					value, ok := balanceNotSpend[addrStr]
					if !ok {
						balanceNotSpend[addrStr] = one.Value
					} else {
						balanceNotSpend[addrStr] = value + one.Value
					}
				}

				// addr := one.Address.B58String()
				// balance, _ := new(sqlite.Balance).FindByAddr(addr)
				// if balance == nil {
				// 	//新地址，添加记录
				// 	err := new(sqlite.Balance).Add(addr, one.Value)
				// 	if err != nil {
				// 		engine.Log.Info("添加新地址错误 " + err.Error())
				// 		panic(err)
				// 		return
				// 	}
				// 	continue
				// }
				// err := new(sqlite.Balance).UpdateBalance(addr, one.Value+balance.Balance)
				// if err != nil {
				// 	engine.Log.Info("2修改余额错误 " + err.Error())
				// 	panic(err)
				// 	return
				// }
			}

			// switch txBase.Class() {
			// case config.Wallet_tx_type_vote_in:
			// 	tx := txBase.(*mining.Tx_vote_in)
			// 	engine.Log.Info("%d %s %s", tx.VoteType, hex.EncodeToString((*tx.GetVout())[0].Address), tx.Vote.B58String())
			// case config.Wallet_tx_type_vote_out:
			// 	tx := txBase.(*mining.Tx_vote_out)
			// 	engine.Log.Info("%s", hex.EncodeToString((*tx.GetVin())[0].Txid))
			// }
		}

		// engine.Log.Info("下一个块hash %s \n", hex.EncodeToString(bh.Nextblockhash))

		if bh.Nextblockhash != nil {
			beforBlockHash = &bh.Nextblockhash
		} else {
			beforBlockHash = nil
		}

		tic := CountBalances(&bhvo)
		tim.CountTxItem(tic, bhvo.BH.Height, bhvo.BH.Time)
		timOld.CountTxItem(tic, bhvo.BH.Height, bhvo.BH.Time)
		tim.Unfrozen(bhvo.BH.Height, bhvo.BH.Time)
		timOld.Unfrozen(bhvo.BH.Height, bhvo.BH.Time)

		txItems := tim.FindBalance(CountAddrCoin)
		// txItems = append(txItems, tim.FindBalanceFrozen(CountAddrCoin)...)

		balanceAll := uint64(0)
		for _, one := range txItems {
			balanceAll = balanceAll + one.Value
		}
		if balanceAll == 0 {
			continue
		}

		txItemsOld := timOld.FindBalance(CountAddrCoin)
		// txItems = append(txItems, tim.FindBalanceFrozen(CountAddrCoin)...)

		balanceAllOld := uint64(0)
		for _, one := range txItemsOld {
			balanceAllOld = balanceAllOld + one.Value
		}

		if balanceAll != balanceAllOld {
			engine.Log.Info("不相等 %d %d", balanceAll, balanceAllOld)
			itemsMap := make(map[string]*mining.TxItem)
			for _, one := range txItems {
				item := one
				// txidStr := hex.EncodeToString(item.Txid)
				// if txidStr == "040000000000000034a13dc53e827de642ad34925e4debb6de86cf2771a104713b808367b10ef010" && item.VoutIndex == 14 {
				// 	engine.Log.Info("Unfrozen: %s %d %d %d", txidStr, item.VoutIndex, item.LockupHeight, item.Status)
				// }
				key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))
				itemsMap[key] = one
			}
			itemsOldMap := make(map[string]*mining.TxItem)
			for _, one := range txItemsOld {
				key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))
				itemsOldMap[key] = one
			}

			for _, one := range txItemsOld {
				key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))
				_, ok := itemsMap[key]
				if !ok {
					engine.Log.Info("%+v", one)

					bs, _ := db.Find(one.Txid)
					txBase, _ := mining.ParseTxBase(mining.ParseTxClass(one.Txid), bs)
					tx := txBase.GetVOJSON()
					bsTx, _ := json.Marshal(tx)
					engine.Log.Info("%s", string(bsTx))

				}
			}

			// engine.Log.Info("txItems")
			// for _, one := range txItems {
			// 	engine.Log.Info("%+v", one)
			// }

			// engine.Log.Info("txItemsOld")
			// for _, one := range txItemsOld {
			// 	engine.Log.Info("%+v", one)
			// }
			panic("不相等")
		}

		// value, ok := balanceMap[utils.Bytes2string(CountAddrCoin)]
		// if !ok {
		// 	panic("不存在")
		// }
		// if balanceAll != value {
		// 	engine.Log.Info("不相等 %d %d", balanceAll, value)
		// 	panic("不存在")
		// }

	}
	engine.Log.Info("链上共流通币数量：%d", balanceTotal)
	return balanceMap, balanceNotSpend
}

func PrintBalanceMap(balances, balanceNotSpend map[string]uint64) {
	if balances == nil {
		engine.Log.Info("打印结果为空")
		return
	}
	for k, v := range balances {
		addrCoin := crypto.AddressCoin([]byte(k))
		engine.Log.Info("%s %d", addrCoin.B58String(), v)
	}

	engine.Log.Info("打印已经解锁的余额")

	if balanceNotSpend == nil {
		engine.Log.Info("打印结果为空")
		return
	}
	for k, v := range balanceNotSpend {
		addrCoin := crypto.AddressCoin([]byte(k))
		engine.Log.Info("%s %d", addrCoin.B58String(), v)
	}
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
		for one := range itemsChan {
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
		// addrOne := crypto.BuildAddr("TEST", vin.Puk)
		// ok := bytes.Equal(addrOne, CountAddrCoin)
		// // ok := vin.CheckIsSelf()
		// if !ok {
		// 	continue
		// }
		// engine.Log.Info("统单易1耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
		//查找这个地址的余额列表，没有则创建一个
		itemCount.SubItems = append(itemCount.SubItems, &mining.TxSubItems{
			Txid:      vin.Txid, //utils.Bytes2string(vin.Txid), //  vin.GetTxidStr(),
			VoutIndex: vin.Vout,
			Addr:      *vin.GetPukToAddr(), // utils.Bytes2string(*vin.GetPukToAddr()), // vin.GetPukToAddrStr(),
		})

	}
	// engine.Log.Info("统单易3耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	txCtrl := mining.GetTransactionCtrl(txItr.Class())
	if txCtrl != nil {

		txCtrl.SyncCount()
		itemChan <- &itemCount
		<-tokenCPU
		return
	}
	// engine.Log.Info("统计单个交易 444 耗时 %s", time.Now().Sub(start))

	//生成新的UTXO收益，保存到列表中
	for voutIndex, vout := range *txItr.GetVout() {

		//找出需要统计余额的地址

		//和自己无关的地址
		// ok := bytes.Equal(vout.Address, CountAddrCoin)
		// // ok := vout.CheckIsSelf()
		// // addrInfo, ok := keystore.FindAddress(vout.Address)
		// if !ok {
		// 	continue
		// }
		//见证人押金和投票押金不计入余额
		if txItr.Class() == config.Wallet_tx_type_deposit_in || txItr.Class() == config.Wallet_tx_type_vote_in {
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
