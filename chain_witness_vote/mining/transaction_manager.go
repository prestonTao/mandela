package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/hex"
	"math/big"
	"sort"
	"sync"
	"time"
)

//保存网络中的交易
//var unpackedTransactions = new(sync.Map) //未打包的交易,key:string=交易hahs id；value=&TxItr

type TransactionManager struct {
	witnessBackup *WitnessBackup //候选见证人
	depositin     *sync.Map      //见证人缴押金,key:string=见证人公钥；value:&TxItr=;见证人不能有重复，因此单独管理
	unpacked      *sync.Map      //未打包的交易,key:string=交易hahs id；value:TransactionRatio=;
	tempTxLock    *sync.RWMutex  //锁
	tempTx        []TxItr        //未验证的交易。有的交易验证需要花很长时间，导致阻塞，因此，先保存在这里，再一条一条的验证。
	tempTxsignal  chan bool      //有未验证的交易时，发送一个信号。
}

/*
	添加一个未打包的交易
*/
func (this *TransactionManager) AddTx(txItr TxItr) bool {
	// engine.Log.Info("添加一个未打包的交易 111111111")
	this.tempTxLock.Lock()
	this.tempTx = append(this.tempTx, txItr)
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
	验证一个未验证的交易
*/
func (this *TransactionManager) checkTx(txItr TxItr) bool {
	start := time.Now()

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

	// fmt.Println("添加一个交易", hex.EncodeToString(*txItr.GetHash()))
	//判断双花的交易
	for _, one := range *txItr.GetVin() {
		//是区块奖励
		if txItr.Class() == config.Wallet_tx_type_mining {
			continue
		}
		//先判断数据库
		txBs, err := db.Find(one.Txid)
		if err != nil {
			// fmt.Println("111111111111111111", err)
			return false
		}
		txItr, err := ParseTxBase(txBs)
		if err != nil {
			// fmt.Println("2222222222222222")
			return false
		}
		if (*txItr.GetVout())[one.Vout].Tx != nil {
			//该笔交易已经被使用
			// fmt.Println("3333333333333333333")
			return false
		}

		//判断押金交易是否有双花的交易
		have := false
		this.depositin.Range(func(k, v interface{}) bool {
			txBase := v.(TransactionRatio)
			for _, two := range *txBase.tx.GetVin() {
				if bytes.Equal(one.Txid, two.Txid) && one.Vout == two.Vout {
					have = true
					return false
				}
			}
			return true
		})
		if have {
			// fmt.Println("44444444444444")
			return false
		}
		//判断未打包的交易里是否有双花交易
		this.unpacked.Range(func(k, v interface{}) bool {
			txBase := v.(TransactionRatio)
			for _, two := range *txBase.tx.GetVin() {
				if bytes.Equal(one.Txid, two.Txid) && one.Vout == two.Vout {
					have = true
					return false
				}
			}
			return true
		})
		if have {
			// fmt.Println("555555555555555555")
			return false
		}
	}

	//见证人押金不能多次提交
	if txItr.Class() == config.Wallet_tx_type_deposit_in {

		addr := crypto.BuildAddr(config.AddrPre, (*txItr.GetVin())[0].Puk)

		// addr, err := keystore.ParseHashByPubkey((*txItr.GetVin())[0].Puk)
		// if err != nil {
		// 	// fmt.Println("666666666666666666")
		// 	return false
		// }

		//判断是否已经交了押金
		if this.witnessBackup.haveWitness(&addr) {
			// fmt.Println("7777777777777777777777")
			return false
		}

		//判断是否有重复的未打包的押金
		pukStr := hex.EncodeToString((*txItr.GetVin())[0].Puk)
		_, ok := this.depositin.Load(pukStr)
		if ok {
			// fmt.Println("888888888888888")
			return false
		}
		//		fmt.Println("--------1111添加一个押金交易", hex.EncodeToString(*txItr.GetHash()))
		this.depositin.Store(pukStr, ratio)
		return true
	}

	txItr.BuildHash()
	this.unpacked.Store(hex.EncodeToString(*txItr.GetHash()), ratio)
	//	fmt.Println("--------2222添加一个普通交易")
	engine.Log.Info("Verifying transactions %s Use time %s", hex.EncodeToString(*txItr.GetHash()), time.Now().Sub(start))

	return true
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
	将未验证的交易，一条一条的验证后，再添加到未打包的交易中去。
	此方法异步执行
*/
func (this *TransactionManager) loopCheckTxs() {
	go func() {
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
			this.tempTx = make([]TxItr, 0)
			this.tempTxLock.Unlock()
			// engine.Log.Info("开始验证交易 2222222222222")
			//开始验证交易
			for _, one := range temp {
				this.checkTx(one)
			}
		}
	}()
}

/*
	删除一个已经打包的交易
*/
func (this *TransactionManager) DelTx(txs []TxItr) {
	for _, one := range txs {
		str := hex.EncodeToString(*one.GetHash())
		if one.Class() == config.Wallet_tx_type_deposit_in {
			pukStr := hex.EncodeToString((*one.GetVin())[0].Puk)
			this.depositin.Delete(pukStr)
		}
		this.unpacked.Delete(str)
	}
}

/*
	打包交易
	控制每个区块大小，给交易手续费排序。
*/
func (this *TransactionManager) Package(blocks []Block) ([]TxItr, [][]byte) {

	//未确认的交易
	unacknowledgedTxs := make([]TxItr, 0)

	//排除已经打包的交易
	exclude := make(map[string]string)
	for _, one := range blocks {
		_, txs, err := one.LoadTxs()
		if err != nil {
			return nil, nil
		}
		for _, txOne := range *txs {
			exclude[hex.EncodeToString(*txOne.GetHash())] = ""
			unacknowledgedTxs = append(unacknowledgedTxs, txOne)
		}
	}

	//打包见证人押金
	txRatios := make([]TransactionRatio, 0)
	this.depositin.Range(func(k, v interface{}) bool {
		txid := k.(string)
		//判断是否有排除的交易，有则不加入列表
		_, ok := exclude[txid]
		if ok {
			return true
		}

		txItr := v.(TransactionRatio)
		//判断交易的合法性
		if err := txItr.tx.Check(); err != nil {
			return true
		}

		txRatios = append(txRatios, txItr)
		// txids = append(txids, *txItr.GetHash())
		// fmt.Println("===111打包押金交易", hex.EncodeToString(*txItr.GetHash()))
		return true
	})

	//打包普通交易
	this.unpacked.Range(func(k, v interface{}) bool {
		txid := k.(string)
		//判断是否有排除的交易，有则不加入列表
		_, ok := exclude[txid]
		if ok {
			return true
		}

		txItr := v.(TransactionRatio)

		if err := txItr.tx.Check(); err != nil {
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
	txs := make([]TxItr, 0)
	txids := make([][]byte, 0)
	sizeTotal := uint64(0)
	for _, one := range txRatios {
		if sizeTotal+one.size > config.Block_size_max {
			//这里用continue是因为，排在后面的交易有可能占用空间小，可以打包到区块中去，使交易费最大化。
			continue
		}

		//判断重复的交易
		if !one.tx.CheckRepeatedTx(unacknowledgedTxs...) {
			continue
		}

		txs = append(txs, one.tx)
		txids = append(txids, *one.tx.GetHash())
		sizeTotal = sizeTotal + one.size
		unacknowledgedTxs = append(unacknowledgedTxs, one.tx)
	}

	return txs, txids
}

/*
	清除过期的交易
*/
func (this *TransactionManager) CleanIxOvertime(height uint64) {
	keys := make([]string, 0)
	this.depositin.Range(func(k, v interface{}) bool {
		txBase := v.(TransactionRatio)
		lockheight := txBase.tx.GetLockHeight()
		if lockheight < height {
			keys = append(keys, k.(string))
		}
		return true
	})
	for _, one := range keys {
		this.depositin.Delete(one)
	}
	keys = make([]string, 0)
	this.unpacked.Range(func(k, v interface{}) bool {
		txBase := v.(TransactionRatio)
		lockheight := txBase.tx.GetLockHeight()
		if lockheight < height {
			keys = append(keys, k.(string))
		}
		return true
	})
	for _, one := range keys {
		this.unpacked.Delete(one)
	}
}

/*
	查询见证人是否缴纳押金
*/
func (this *TransactionManager) FindDeposit(puk string) bool {
	_, ok := this.depositin.Load(puk)
	return ok
}

/*
	创建交易管理
*/
func NewTransactionManager(wb *WitnessBackup) *TransactionManager {
	tm := TransactionManager{
		witnessBackup: wb,                 //
		depositin:     new(sync.Map),      //见证人缴押金,key:string=交易hahs id；value=&TxItr
		unpacked:      new(sync.Map),      //未打包的交易,key:string=交易hahs id；value:&TxItr=;
		tempTxLock:    new(sync.RWMutex),  //
		tempTx:        make([]TxItr, 0),   //
		tempTxsignal:  make(chan bool, 1), //
	}
	tm.loopCheckTxs()

	return &tm
}
