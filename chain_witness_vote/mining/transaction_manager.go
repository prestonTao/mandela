package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"encoding/hex"
	"math/big"
	"sort"
	"strconv"
	"sync"
)

//保存网络中的交易
//var unpackedTransactions = new(sync.Map) //未打包的交易,key:string=交易hahs id；value=&TxItr

type TransactionManager struct {
	witnessBackup   *WitnessBackup     //候选见证人
	depositin       *sync.Map          //见证人缴押金,key:string=见证人公钥；value:&TxItr=;见证人不能有重复，因此单独管理
	unpacked        *sync.Map          //未打包的交易,key:string=交易hash id；value:TransactionRatio=;
	ActiveVoutIndex *sync.Map          //活动的交易输出，key:string=[txid]_[vout index];value:=;
	tempTxLock      *sync.RWMutex      //锁
	tempTx          []TransactionRatio //未验证的交易。有的交易验证需要花很长时间，导致阻塞，因此，先保存在这里，再一条一条的验证。
	tempTxsignal    chan bool          //有未验证的交易时，发送一个信号。
}

/*
	添加一个未打包的交易
*/
func (this *TransactionManager) AddTx(txItr TxItr) bool {
	// engine.Log.Info("添加一个未打包的交易 111111111")
	//已有的交易不重复保存，避免把交易的BlockHash字段覆盖掉
	txItr.BuildHash()
	if !db.LevelDB.CheckHashExist(*txItr.GetHash()) {
		//把交易放到数据库，并且标记为未确认的交易。
		err := SaveTempTx(txItr)
		if err != nil {
			return false
		}
	}

	txbs := txItr.Serialize()

	div1 := new(big.Int).Mul(big.NewInt(int64(txItr.GetGas())), big.NewInt(100000000))
	div2 := big.NewInt(int64(len(*txbs)))

	ratioValue := new(big.Int).Div(div1, div2)
	// ratioValue := txItr.GetGas() / uint64(len(*txbs))
	ratio := TransactionRatio{
		tx:    txItr,              //交易
		size:  uint64(len(*txbs)), //交易总大小
		gas:   txItr.GetGas(),     //手续费
		Ratio: ratioValue,         //价值比
	}

	this.tempTxLock.Lock()
	this.tempTx = append(this.tempTx, ratio)
	this.tempTxLock.Unlock()
	// engine.Log.Info("添加一个未打包的交易 222222222222")
	select {
	case this.tempTxsignal <- false:
		// engine.Log.Info("添加一个未打包的交易 3333333333333")
	default:
		// engine.Log.Info("添加一个未打包的交易 4444444444444")
	}
	// engine.Log.Info("添加一个未打包的交易 555555555555555")
	return true
}

/*
	将未验证的交易，一条一条的验证后，再添加到未打包的交易中去。
	此方法异步执行
*/
func (this *TransactionManager) loopCheckTxs() {
	utils.Go(func() {
		// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
		// engine.Log.Info("add Goroutine:%s", goroutineId)
		// defer engine.Log.Info("del Goroutine:%s", goroutineId)
		for range this.tempTxsignal {
			// engine.Log.Info("开始验证交易 1111111111111")
			this.tempTxLock.Lock()
			if len(this.tempTx) <= 0 {
				this.tempTxLock.Unlock()
				continue
			}
			//取出交易
			temp := this.tempTx
			//删除交易
			this.tempTx = make([]TransactionRatio, 0)
			this.tempTxLock.Unlock()
			// engine.Log.Info("开始验证交易 2222222222222")
			//开始验证交易
			for _, one := range temp {

				// engine.Log.Debug("开始验证交易 %s", hex.EncodeToString(*one.tx.GetHash()))
				//验证失败的交易
				err := this.checkTx(one)
				if err != nil {
					// engine.Log.Debug("交易验证失败 %s %s", hex.EncodeToString(*one.tx.GetHash()), err.Error())
					engine.Log.Debug("Transaction validation failed %s %s", hex.EncodeToString(*one.tx.GetHash()), err.Error())
					// bs, _ := json.Marshal(one.GetVOJSON())
					// engine.Log.Debug("%s", string(bs))
					continue
				}
				// engine.Log.Debug("结束验证交易 %s", hex.EncodeToString(*one.tx.GetHash()))
			}
		}
	})
}

/*
	验证一个未验证的交易
*/
func (this *TransactionManager) checkTx(ratio TransactionRatio) error {
	// startTime := time.Now()
	txItr := ratio.tx

	// engine.Log.Debug("开始验证双花交易")
	//验证双花交易
	activiVoutIndexs := make(map[string]interface{})
	for _, vin := range *txItr.GetVin() {
		//是区块奖励
		if txItr.Class() == config.Wallet_tx_type_mining {
			continue
		}
		preTxidStr := utils.Bytes2string(vin.Txid) // vin.GetTxidStr()

		//先验证数据库里的交易是否有双花
		if !db.LevelTempDB.CheckHashExist(BuildKeyForUnspentTransaction(vin.Txid, vin.Vout)) {
			//已经使用过了
			// engine.Log.Warn("这个交易已经使用过了 %s_%d", hex.EncodeToString(vin.Txid), vin.Vout)
			engine.Log.Warn("This transaction has been used %s_%d", hex.EncodeToString(vin.Txid), vin.Vout)
			return config.ERROR_tx_is_use
		}
		//再验证未上链的交易中是否有双花
		_, ok := this.ActiveVoutIndex.Load(preTxidStr + "_" + strconv.Itoa(int(vin.Vout)))
		if ok {
			// db.CheckHashExist(*txItr.GetHash())

			//已经存在
			// engine.Log.Warn("未上链的交易有双花 %s", hex.EncodeToString(vin.Txid)+"_"+strconv.Itoa(int(vin.Vout)))
			engine.Log.Warn("Transactions that are not linked have double flowers %s", hex.EncodeToString(vin.Txid)+"_"+strconv.Itoa(int(vin.Vout)))
			return config.ERROR_tx_is_use
		}
		activiVoutIndexs[preTxidStr+"_"+strconv.Itoa(int(vin.Vout))] = nil
	}

	//见证人押金不能多次提交
	if txItr.Class() == config.Wallet_tx_type_deposit_in {
		vin := (*txItr.GetVin())[0]

		//判断是否已经交了押金
		if this.witnessBackup.haveWitness(vin.GetPukToAddr()) {
			// engine.Log.Warn("已经缴纳了押金了 %s", hex.EncodeToString(*txItr.GetHash()))
			engine.Log.Warn("The deposit has been paid %s", hex.EncodeToString(*txItr.GetHash()))
			return config.ERROR_deposit_exist
		}

		//判断是否有重复的未打包的押金
		_, ok := this.depositin.Load(utils.Bytes2string(vin.Puk))
		if ok {
			// engine.Log.Warn("有重复的未打包的押金 %s", hex.EncodeToString(*txItr.GetHash()))
			engine.Log.Warn("There is a duplicate unpackaged deposit %s", hex.EncodeToString(*txItr.GetHash()))
			return config.ERROR_deposit_exist
		}
		// engine.Log.Info("Verifying 777 Use time %s", time.Now().Sub(startTime))
		this.depositin.Store(utils.Bytes2string(vin.Puk), ratio)
		// engine.Log.Warn("见证人押金验证通过 %s", txItr.GetHashStr())
		return nil
	}

	//检查重复的交易
	txCtrl := GetTransactionCtrl(txItr.Class())
	if txCtrl != nil {
		err := txCtrl.CheckMultiplePayments(txItr)
		if err != nil {
			engine.Log.Warn("CheckMultiplePayments %s", hex.EncodeToString(*txItr.GetHash()))
			return err
		}
	}

	txItr.BuildHash()
	this.unpacked.Store(utils.Bytes2string(*txItr.GetHash()), ratio)

	//保存活动的交易输出
	for k, v := range activiVoutIndexs {
		// engine.Log.Info("添加活动的交易 %s", k)
		this.ActiveVoutIndex.Store(k, v)
	}
	// engine.Log.Info("Verifying transactions Use time %s", time.Now().Sub(startTime))
	return nil
}

/*
	添加一个未打包的交易
*/
func (this *TransactionManager) AddTxs(txs ...TxItr) {
	for _, one := range txs {
		this.AddTx(one)
	}
}

/*
	删除一个已经打包的交易
*/
func (this *TransactionManager) DelTx(txs []TxItr) {
	for _, one := range txs {
		// str := hex.EncodeToString(*one.GetHash())
		// str := one.GetHashStr()
		if one.Class() == config.Wallet_tx_type_deposit_in {
			pukStr := (*one.GetVin())[0].Puk //.GetPukStr() // hex.EncodeToString((*one.GetVin())[0].Puk)
			this.depositin.Delete(utils.Bytes2string(pukStr))
		}
		this.unpacked.Delete(utils.Bytes2string(*one.GetHash()))

		//删除交易输出中的索引
		for _, vin := range *one.GetVin() {
			// keyStr := hex.EncodeToString(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
			keyStr := utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
			// engine.Log.Info("删除活动的交易 %s", keyStr)
			this.ActiveVoutIndex.Delete(keyStr)
		}
	}
}

/*
	打包交易
	控制每个区块大小，给交易手续费排序。
*/
func (this *TransactionManager) Package(reward *Tx_reward, height uint64, blocks []Block, createBlockTime int64) ([]TxItr, [][]byte) {

	//未确认的交易
	unacknowledgedTxs := make([]TxItr, 0)

	// start := time.Now()
	//排除已经打包的交易
	exclude := make(map[string]string)
	for _, one := range blocks {
		// engine.Log.Info("打包加载区块信息")
		_, txs, err := one.LoadTxs()
		if err != nil {
			return nil, nil
		}
		for _, txOne := range *txs {
			// exclude[hex.EncodeToString(*txOne.GetHash())] = ""
			exclude[utils.Bytes2string(*txOne.GetHash())] = ""
			unacknowledgedTxs = append(unacknowledgedTxs, txOne)
		}
	}

	// engine.Log.Info("排除打包的交易花费耗时 %s", time.Now().Sub(start))

	//打包见证人押金
	txRatios := make([]TransactionRatio, 0)
	this.depositin.Range(func(k, v interface{}) bool {
		// engine.Log.Debug("开始打包见证人押金交易 111")
		txid := k.(string)
		//判断是否有排除的交易，有则不加入列表
		_, ok := exclude[txid]
		if ok {
			// engine.Log.Debug("开始打包见证人押金交易 222")
			return true
		}

		txItr := v.(TransactionRatio)
		//判断交易锁定高度
		if err := txItr.tx.CheckLockHeight(height); err != nil {
			// engine.Log.Debug("开始打包见证人押金交易 333")
			return true
		}
		//判断余额冻结高度
		if err := txItr.tx.CheckFrozenHeight(height, createBlockTime); err != nil {
			return true
		}

		//判断交易签名是否正确
		txRatios = append(txRatios, txItr)
		// txids = append(txids, *txItr.GetHash())
		// engine.Log.Debug("开始打包见证人押金交易 444")
		return true
	})

	//打包普通交易
	this.unpacked.Range(func(k, v interface{}) bool {
		txid := k.(string)
		// engine.Log.Info("判断本次普通交易 %s", txid)
		//判断是否有排除的交易，有则不加入列表
		_, ok := exclude[txid]
		if ok {
			// engine.Log.Info("排除已经打包的交易 %s", txid)
			return true
		}

		txItr := v.(TransactionRatio)

		if err := txItr.tx.CheckLockHeight(height); err != nil {
			// engine.Log.Info("排除锁定高度不正确的交易 %s %s", txid, err.Error())
			return true
		}

		if err := txItr.tx.CheckFrozenHeight(height, createBlockTime); err != nil {
			// engine.Log.Info("排除冻结余额高度不正确的交易 %s %s", txid, err.Error())
			return true
		}

		txRatios = append(txRatios, txItr)
		// txids = append(txids, *txItr.GetHash())
		// fmt.Println("===222打包普通交易", hex.EncodeToString(*txItr.GetHash()))
		return true
	})
	//排序，把手续费多的排前面
	rss := &RatioSort{txRatio: txRatios}
	sort.Sort(rss)
	txRatios = rss.txRatio

	//获取排在前面的
	// start = time.Now()
	txs := make([]TxItr, 0)
	txids := make([][]byte, 0)
	sizeTotal := uint64(0)
	//多个见证人的时候，有的区块没有奖励
	if reward != nil {
		sizeTotal = sizeTotal + uint64(len(*reward.Serialize()))
	}
	for _, one := range txRatios {
		if sizeTotal+one.size > config.Block_size_max {
			//这里用continue是因为，排在后面的交易有可能占用空间小，可以打包到区块中去，使交易费最大化。
			continue
		}

		//判断重复的交易
		if !one.tx.CheckRepeatedTx(unacknowledgedTxs...) {
			// engine.Log.Info("打包时有重复的交易 %s", one.tx.GetHashStr())
			continue
		}

		txs = append(txs, one.tx)
		txids = append(txids, *one.tx.GetHash())
		sizeTotal = sizeTotal + one.size
		unacknowledgedTxs = append(unacknowledgedTxs, one.tx)
	}

	// engine.Log.Info("本次打包大小 %d 交易个数 %d", sizeTotal, len(txids))

	// engine.Log.Info("获取排在前面的交易耗时 %s", time.Now().Sub(start))

	return txs, txids
}

/*
	清除过期的交易
*/
func (this *TransactionManager) CleanIxOvertime(height uint64) {
	//清除过期的见证人押金交易
	this.depositin.Range(func(k, v interface{}) bool {
		txBase := v.(TransactionRatio)
		lockheight := txBase.tx.GetLockHeight()
		if lockheight < height {
			//删除
			this.depositin.Delete(k.(string))
			//删除交易输出中的索引和缓存
			for _, vin := range *txBase.tx.GetVin() {
				// txidStr := vin.GetTxidStr() //hex.EncodeToString(vin.Txid)
				keyStr := utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
				// engine.Log.Info("删除过期的活动交易 111 %s", keyStr)
				this.ActiveVoutIndex.Delete(keyStr)
				//删除缓存
				// TxCache.RemoveTxInCache(txidStr, vin.Vout)
			}
		}
		return true
	})
	//清除过期的普通交易
	//不验证的区块不清除锁定高度
	if GetHighestBlock() < config.Mining_block_start_height+config.Mining_block_start_height_jump {
		return
	}
	this.unpacked.Range(func(k, v interface{}) bool {
		txBase := v.(TransactionRatio)
		lockheight := txBase.tx.GetLockHeight()
		if lockheight < height {
			//删除
			this.unpacked.Delete(k.(string))
			//删除交易输出中的索引和缓存
			for _, vin := range *txBase.tx.GetVin() {
				// txidStr := vin.GetTxidStr() // hex.EncodeToString(vin.Txid)
				keyStr := utils.Bytes2string(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout))
				// engine.Log.Info("删除过期的活动的交易 222 %s", hex.EncodeToString(vin.Txid))
				this.ActiveVoutIndex.Delete(keyStr)
				//删除缓存
				// TxCache.RemoveTxInCache(txidStr, vin.Vout)
			}
		}
		return true
	})

}

/*
	查询见证人是否缴纳押金
*/
func (this *TransactionManager) FindDeposit(puk []byte) bool {
	_, ok := this.depositin.Load(utils.Bytes2string(puk))
	return ok
}

/*
	创建交易管理
*/
func NewTransactionManager(wb *WitnessBackup) *TransactionManager {
	tm := TransactionManager{
		witnessBackup:   wb,                          //
		depositin:       new(sync.Map),               //见证人缴押金,key:string=交易hahs id；value=&TxItr
		unpacked:        new(sync.Map),               //未打包的交易,key:string=交易hahs id；value:&TxItr=;
		ActiveVoutIndex: new(sync.Map),               //
		tempTxLock:      new(sync.RWMutex),           //
		tempTx:          make([]TransactionRatio, 0), //
		tempTxsignal:    make(chan bool, 1),          //
	}
	tm.loopCheckTxs()

	return &tm
}
