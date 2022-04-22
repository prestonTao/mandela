package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/sqlite3_db"
	"bytes"
	"errors"
	"runtime"
	"sort"
	"strconv"
	"sync"

	"github.com/go-xorm/xorm"
)

/*
	地址余额管理器
*/
type BalanceManager struct {
	countTotal        uint64              //
	chain             *Chain              //链引用
	syncBlockHead     chan *BlockHeadVO   //正在同步的余额，准备导入到余额中
	notspentBalance   *TxItemManager      //未花费的余额索引
	notspentBalanceDB *TxItemManagerDB    //未花费的余额索引，保存到数据库，降低内存消耗
	depositin         *TxItem             //保存成为见证人押金交易
	votein            *sync.Map           //保存本节点投票的押金额度，key:string=见证人地址;value:*Balance=押金列表;
	witnessBackup     *WitnessBackup      //
	txManager         *TransactionManager //
	otherDeposit      *sync.Map           //其他押金，key:uint64=交易类型;value:*sync.Map=押金列表;
}

func NewBalanceManager(wb *WitnessBackup, tm *TransactionManager, chain *Chain) *BalanceManager {
	bm := &BalanceManager{
		chain:             chain,
		syncBlockHead:     make(chan *BlockHeadVO, 1), //正在同步的余额，准备导入到余额中
		notspentBalance:   NewTxItemManager(),         //
		notspentBalanceDB: NewTxItemManagerDB(),       //
		witnessBackup:     wb,                         //
		txManager:         tm,                         //
		votein:            new(sync.Map),              //
		otherDeposit:      new(sync.Map),              //
	}
	utils.Go(bm.run)
	// go bm.run()
	return bm
}

/*
	获取一个地址的余额列表
*/
// func (this *BalanceManager) FindBalanceOne(addr *crypto.AddressCoin) (*Balance, *Balance) {
// 	bas, bafrozens := this.FindBalance(addr)
// 	var ba, bafrozen *Balance
// 	if bas != nil || len(bas) >= 1 {
// 		// fmt.Println("这里错误222")
// 		ba = bas[0]
// 	}
// 	if bafrozens != nil || len(bafrozens) >= 1 {
// 		bafrozen = bafrozens[0]
// 	}
// 	return ba, bafrozen
// }

/*
	获取一个地址的押金列表
*/
func (this *BalanceManager) GetDepositIn() *TxItem {
	return this.depositin
}

/*
	获取一个地址的押金列表
*/
func (this *BalanceManager) GetVoteIn(witnessAddrBs VoteAddress) *Balance {
	witnessAddr := utils.Bytes2string(witnessAddrBs)
	// engine.Log.Debug("查询一个见证人 %s %s", hex.EncodeToString(witnessAddrBs), witnessAddr)
	v, ok := this.votein.Load(witnessAddr)
	if !ok {
		return nil
	}
	b := v.(*Balance)
	return b
}

/*
	获取一个地址的押金列表
*/
func (this *BalanceManager) GetVoteInByTxid(txid []byte) *TxItem {
	var tx *TxItem
	this.votein.Range(func(k, v interface{}) bool {
		b := v.(*Balance)
		b.Txs.Range(func(txidItr, v interface{}) bool {
			// dstTxid := txidItr.(string)
			txItem := v.(*TxItem)
			//0600000000000000b027d84883693a16de4df892c4d856cbf103ed0e28a2d5d98277199ea2d79345_0
			if bytes.Equal(txid, txItem.Txid) {
				tx = txItem
				return false
			}

			// if utils.Bytes2string(txid) == strings.SplitN(dstTxid, "_", 2)[0] {
			// 	tx = v.(*TxItem)
			// 	return false
			// }
			return true
		})
		if tx != nil {
			return false
		}
		return true
	})
	return tx
}

/*
	统计多个地址的余额
*/
func (this *BalanceManager) FindBalance(addrs crypto.AddressCoin) (uint64, uint64) {

	if config.Wallet_txitem_save_db {
		return 0, 0
	} else {
		var count, countf uint64
		bas := this.notspentBalance.FindBalanceNotSpent(addrs)
		//统计冻结的余额
		bafrozens := this.notspentBalance.FindBalanceFrozen(addrs)
		// txitems, bfs := chain.balance.FindBalance(addr)
		for _, item := range bas {
			count = count + item.Value
		}

		for _, item := range bafrozens {
			// engine.Log.Info("统计冻结 111 %d", item.Value)
			countf = countf + item.Value
		}
		// engine.Log.Info("统计冻结 222 %d", countf)
		return count, countf
	}
}

/*
	查询各种状态的余额
*/
func (this *BalanceManager) FindBalanceValue() (n, f, l uint64) {
	if config.Wallet_txitem_save_db {
		return this.notspentBalanceDB.FindBalanceValue(this.chain.GetCurrentBlock())
	} else {
		return this.notspentBalance.FindBalanceValue()
	}
	// return 0, 0, 0
}

/*
	从最后一个块开始统计多个地址的余额
*/
func (this *BalanceManager) FindBalanceAll() (map[string]uint64, map[string]uint64, map[string]uint64) {
	if config.Wallet_txitem_save_db {
		return this.notspentBalance.FindBalanceAllAddrs()
	} else {
		return this.notspentBalance.FindBalanceAllAddrs()
	}

	// bas := this.notspentBalance.FindBalanceNotSpentAll()
	// //统计冻结的余额
	// bafrozens := this.notspentBalance.FindBalanceFrozenAll()
	// //统计锁仓的余额
	// baLockup := this.notspentBalance.FindBalanceLockupAll()
	// return bas, bafrozens, baLockup
}

/*
	从最后一个块开始统计多个地址的余额
*/
// func (this *BalanceManager) FindBalanceAll() ([]*TxItem, []*TxItem, []*TxItem) {
// 	bas := this.notspentBalance.FindBalanceNotSpentAll()
// 	//统计冻结的余额
// 	bafrozens := this.notspentBalance.FindBalanceFrozenAll()
// 	//统计锁仓的余额
// 	baLockup := this.notspentBalance.FindBalanceLockupAll()
// 	return bas, bafrozens, baLockup
// }

/*
	构建付款输入
	当扣款账户地址参数不为空，则从指定的扣款账户扣款，
	即使金额不够，也不会从其他账户扣款

	@srcAddress    *crypto.AddressCoin    扣款账户地址
	@amount        uint64                 需要的金额
*/
func (this *BalanceManager) BuildPayVin(srcAddress crypto.AddressCoin, amount uint64) (uint64, []*TxItem) {

	var tis TxItemSort = make([]*TxItem, 0)

	//当指定扣款账户地址，则从指定地址扣款
	if srcAddress != nil {
		if config.Wallet_txitem_save_db {
			tis = this.notspentBalanceDB.FindBalanceByAddr(srcAddress, this.chain.GetCurrentBlock(), amount)
		} else {
			tis = this.notspentBalance.FindBalanceNotSpent(srcAddress)
		}
	} else {
		if config.Wallet_txitem_save_db {
			tis = this.notspentBalanceDB.FindBalanceAll(this.chain.GetCurrentBlock(), amount)
		} else {
			tis = this.notspentBalance.FindBalanceNotSpentAll()
		}
	}

	// fmt.Printf("查询到的有余额的交易 %d %+v\n", tis)

	if len(tis) <= 0 {
		return 0, nil
	}

	sort.Sort(&tis)

	total := uint64(0)
	items := make([]*TxItem, 0)
	if tis[0].Value >= amount {
		item := tis[0]
		for i, one := range tis {
			if amount >= one.Value {
				item = tis[i]
				if i == len(tis)-1 {
					items = append(items, item)
					total = item.Value
				}
			} else {
				items = append(items, item)
				total = item.Value
				break
			}
		}
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

/*
	引入最新的块
	将交易计入余额
	使用过的UTXO余额删除
*/
func (this *BalanceManager) CountBalanceForBlock(bhvo *BlockHeadVO) {
	this.countBlock(bhvo)

	//给已经确认的区块建立高度索引
	// engine.Log.Info("保存索引 %s", config.BlockHeight+strconv.Itoa(int(bhvo.BH.Height)))
	db.LevelDB.Save([]byte(config.BlockHeight+strconv.Itoa(int(bhvo.BH.Height))), &bhvo.BH.Hash)
}

func (this *BalanceManager) run() {
	for bhvo := range this.syncBlockHead {
		this.countBlock(bhvo)
	}
}

/*
	开始统计余额
*/
func (this *BalanceManager) countBlock(bhvo *BlockHeadVO) {
	this.countTotal++
	if this.countTotal == bhvo.BH.Height {
		engine.Log.Info("count block group:%d height:%d witness:%s total:%d", bhvo.BH.GroupHeight, bhvo.BH.Height, bhvo.BH.Witness.B58String(), this.countTotal)
	} else {
		engine.Log.Info("count block group:%d height:%d witness:%s total:%d Unequal", bhvo.BH.GroupHeight, bhvo.BH.Height, bhvo.BH.Witness.B58String(), this.countTotal)
	}

	// start := time.Now()

	// witness, _ := GetLongChain().GetLastBlock()
	// if witness != nil {
	// 	engine.Log.Info("打印最新区块出块时间 %s", utils.FormatTimeToSecond(time.Unix(witness.CreateBlockTime, 0)))
	// }

	// for _, one := range bhvo.Txs {
	// 	engine.Log.Info("统计到的交易hash %s", one.GetHashStr())
	// }

	SaveTxToBlockHead(bhvo)

	//统计社区奖励
	// this.countCommunityReward(bhvo)
	// engine.Log.Info("统计社区奖励耗时 %d %s", bhvo.BH.Height, time.Now().Sub(start))

	//及时将已花费的交易作废，新的余额添加索引
	this.MarkTransactionAsUsed(bhvo)
	// engine.Log.Info("余额添加索引耗时 %d %s", bhvo.BH.Height, time.Now().Sub(start))

	CountBalanceOther(this.notspentBalance, this.otherDeposit, bhvo)
	// engine.Log.Info("统计其他类型的交易耗时 %d %s", bhvo.BH.Height, time.Now().Sub(start))

	//及时统计押金及投票
	this.countDepositAndVote(bhvo)
	// engine.Log.Info("及时统计押金及投票耗时 %d %s", bhvo.BH.Height, time.Now().Sub(start))

	//异步统计，包括可用余额累计和历史记录
	this.SyncCountOthre(bhvo)
	// engine.Log.Info("统计交易及其他耗时 %d %s", bhvo.BH.Height, time.Now().Sub(start))
}

/*
	异步统计其他的
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) SyncCountOthre(bhvo *BlockHeadVO) {
	// start := time.Now()
	//统计交易余额
	this.countBalances(bhvo)
	// engine.Log.Info("区块高度 %d 统计交易余额耗时 %s", bhvo.BH.Height, time.Now().Sub(start))

	//统计交易历史记录
	this.countTxHistory(bhvo)
	// engine.Log.Info("区块高度 %d 统计交易历史记录耗时 %s", bhvo.BH.Height, time.Now().Sub(start))

	//冻结的交易未上链，回滚
	if !this.chain.SyncBlockFinish {
		//未同步完成，不回滚交易
		return
	}
	this.Unfrozen(bhvo.BH.Height-1, bhvo.BH.Time)
	// engine.Log.Info("区块高度 %d 交易回滚耗时 %s", bhvo.BH.Height, time.Now().Sub(start))

}

/*
	及时将已花费的交易作废，新的余额添加索引
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) MarkTransactionAsUsed(bhvo *BlockHeadVO) {

	//查询排除的交易
	// excludeTx := make([]config.ExcludeTx, 0)
	// for i, one := range config.Exclude_Tx {
	// 	if bhvo.BH.Height == one.Height {
	// 		excludeTx = append(excludeTx, config.Exclude_Tx[i])
	// 	}
	// }

	for _, txItr := range bhvo.Txs {
		// engine.Log.Info("开始标记交易 %s", hex.EncodeToString(*txItr.GetHash()))
		//排除的交易不验证
		// for _, two := range excludeTx {
		// 	if bytes.Equal(two.TxByte, *txItr.GetHash()) {
		// 		continue
		// 	}
		// }

		//删除未导入区块的交易无效的标记。
		// txidBs := txItr.GetHash()
		db.LevelDB.Remove(config.BuildTxNotImport(*txItr.GetHash()))

		// txHashStr := txItr.GetHashStr()
		// engine.Log.Info("开始标记交易余额 %s", txHashStr)
		for _, vin := range *txItr.GetVin() {
			//是区块奖励
			if txItr.Class() == config.Wallet_tx_type_mining {
				continue
			}
			// preTxHashStr := vin.GetTxidStr()
			//删除一个已经使用了的交易输出
			db.LevelTempDB.Remove(BuildKeyForUnspentTransaction(vin.Txid, vin.Vout))
			// engine.Log.Info("删除一个已经使用过的交易输出 %s", BuildKeyForUnspentTransaction(preTxHashStr, vin.Vout))
		}

		//生成新的UTXO收益，保存到列表中
		for voutIndex, vout := range *txItr.GetVout() {
			//保存未使用的vout索引
			bs := utils.Uint64ToBytes(vout.FrozenHeight)
			// db.Save(BuildKeyForUnspentTransaction(txHashStr, uint64(voutIndex)), &bs)
			db.LevelTempDB.Save(BuildKeyForUnspentTransaction(*txItr.GetHash(), uint64(voutIndex)), &bs)
			// engine.Log.Info("添加一个交易输出 %s", BuildKeyForUnspentTransaction(txHashStr, uint64(voutIndex)))
		}
	}
}

/*
	统计押金和投票
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countDepositAndVote(bhvo *BlockHeadVO) {
	// engine.Log.Info("开始统计交易中的投票 000")
	for _, txItr := range bhvo.Txs {

		txItr.BuildHash()
		// txHashStr := txItr.GetHashStr()

		txCtrl := GetTransactionCtrl(txItr.Class())
		if txCtrl != nil {
			// txCtrl.CountBalance(this.notspentBalance, this.otherDeposit, bhvo, uint64(txIndex))
			return
		}

		vinAddrs := make([]*crypto.AddressCoin, 0)

		//将之前的UTXO标记为已经使用，余额中减去。
		for _, vin := range *txItr.GetVin() {

			//检查地址是否是自己的
			isSelf := vin.CheckIsSelf()
			vinAddrs = append(vinAddrs, vin.GetPukToAddr())
			if !isSelf {
				continue
			}
			//是区块奖励
			if txItr.Class() == config.Wallet_tx_type_mining {
				continue
			}
			// preTxItr, err := FindTxBase(vin.Txid)
			preTxItr, err := LoadTxBase(vin.Txid)
			if err != nil {
				//TODO 不能解析上一个交易，程序出错退出
				continue
			}

			switch txItr.Class() {
			case config.Wallet_tx_type_mining:
			case config.Wallet_tx_type_deposit_in:
			case config.Wallet_tx_type_deposit_out:

				if vin.Vout == 0 && preTxItr.Class() == config.Wallet_tx_type_deposit_in {

					if this.depositin != nil {
						if bytes.Equal(*vin.GetPukToAddr(), *this.depositin.Addr) {
							this.depositin = nil
						}
					}
				}
			case config.Wallet_tx_type_pay:

			case config.Wallet_tx_type_vote_in:
			case config.Wallet_tx_type_vote_out:

				if vin.Vout == 0 && preTxItr.Class() == config.Wallet_tx_type_vote_in {
					votein := preTxItr.(*Tx_vote_in)
					// engine.Log.Info("删除社区节点开始高度的区块:%s", (*preTxItr.GetVout())[vin.Vout].Address.B58String())
					//保存社区节点开始高度的区块
					// db.LevelTempDB.Remove(BuildCommunityAddrStartHeight((*preTxItr.GetVout())[vin.Vout].Address))

					voteinAddr := utils.Bytes2string(votein.Vote) //votein.Vote.B58String()
					b, ok := this.votein.Load(voteinAddr)
					if ok {
						ba := b.(*Balance)
						// ba.Txs.Delete(hex.EncodeToString(*preTxItr.GetHash()) + "_" + strconv.Itoa(int(vin.Vout)))
						ba.Txs.Delete(utils.Bytes2string(*preTxItr.GetHash()) + "_" + strconv.Itoa(int(vin.Vout)))
						// engine.Log.Debug("保存一个见证人1 %s %s", hex.EncodeToString(votein.Vote), votein.Vote.B58String())
						this.votein.Store(voteinAddr, ba)
					}
				}
			}
		}

		//生成新的UTXO收益，保存到列表中
		for voutIndex, vout := range *txItr.GetVout() {
			//找出需要统计余额的地址

			//和自己无关的地址
			ok := vout.CheckIsSelf()
			if !ok {
				continue
			}

			switch txItr.Class() {
			case config.Wallet_tx_type_mining:
				//保存历史记录
			case config.Wallet_tx_type_deposit_in:
				//
				if voutIndex == 0 {

					txItem := TxItem{
						Addr:      &(*txItr.GetVout())[voutIndex].Address, //  &vout.Address,
						Value:     vout.Value,                             //余额
						Txid:      *txItr.GetHash(),                       //交易id
						VoutIndex: uint64(voutIndex),                      //交易输出index，从0开始
						Height:    bhvo.BH.Height,                         //
					}
					this.depositin = &txItem
					//创始节点直接打开挖矿
					if config.InitNode {
						config.AlreadyMining = true
					}
					//自己提交见证人押金后，再打开出块的开关
					if config.SubmitDepositin {
						config.AlreadyMining = true
					}
					continue
				}
			case config.Wallet_tx_type_deposit_out:
				//保存历史记录

			case config.Wallet_tx_type_pay:
				//判断是否是找零地址，判断依据是输入地址是否有自己钱包的地址

			case config.Wallet_tx_type_vote_in:
				if voutIndex == 0 {

					txItem := TxItem{
						Addr:      &(*txItr.GetVout())[voutIndex].Address, //  &vout.Address,
						Value:     vout.Value,                             //余额
						Txid:      *txItr.GetHash(),                       //交易id
						VoutIndex: uint64(voutIndex),                      //交易输出index，从0开始
						Height:    bhvo.BH.Height,                         //
					}

					voteIn := txItr.(*Tx_vote_in)

					witnessAddr := utils.Bytes2string(voteIn.Vote) //voteIn.Vote.B58String()

					// engine.Log.Info("----------------------------------保存投票" + witnessAddr + "end")

					v, ok := this.votein.Load(witnessAddr)
					var ba *Balance
					if ok {
						ba = v.(*Balance)
					} else {
						ba = new(Balance)
						addr := voteIn.Vote.GetAddress()
						ba.Addr = &addr
						ba.Txs = new(sync.Map)
					}
					txItem.VoteType = voteIn.VoteType
					if voteIn.VoteType == VOTE_TYPE_community {
						//给见证人投票，成为社区节点
						addr := (*txItr.GetVout())[voutIndex].Address
						// engine.Log.Info("保存社区节点开始高度的区块:%s", addr.B58String())
						//保存社区节点开始高度的区块
						blockhash, _ := db.GetTxToBlockHash(txItr.GetHash())
						db.LevelTempDB.Save((addr), blockhash)
					}
					// ba.Txs.Store(txHashStr+"_"+strconv.Itoa(voutIndex), &txItem)
					ba.Txs.Store(utils.Bytes2string(*txItr.GetHash())+"_"+strconv.Itoa(voutIndex), &txItem)
					// engine.Log.Debug("保存一个见证人2 %s %s", hex.EncodeToString(voteIn.Vote), witnessAddr)
					this.votein.Store(witnessAddr, ba)
					continue
				}
			case config.Wallet_tx_type_vote_out:
				//保存历史记录
			}
		}
	}
}

/*
	统计交易余额
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countBalances(bhvo *BlockHeadVO) {

	//将txitem集中起来，一次性添加
	itemCount := TxItemCount{
		Additems: make([]*TxItem, 0),
		SubItems: make([]*TxSubItems, 0),
		// deleteKey: make([]string, 0),
	}
	itemsChan := make(chan *TxItemCount, len(bhvo.Txs))
	wg := new(sync.WaitGroup)
	wg.Add(len(bhvo.Txs))
	utils.Go(
		func() {
			goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
			_, file, line, _ := runtime.Caller(0)
			engine.AddRuntime(file, line, goroutineId)
			defer engine.DelRuntime(file, line, goroutineId)
			for i := 0; i < len(bhvo.Txs); i++ {
				one := <-itemsChan
				if one != nil {
					itemCount.Additems = append(itemCount.Additems, one.Additems...)
					itemCount.SubItems = append(itemCount.SubItems, one.SubItems...)
					// itemCount.deleteKey = append(itemCount.deleteKey, one.deleteKey...)
				}
				wg.Done()
			}
			// for one := range itemsChan {
			// 	if one != nil {
			// 		itemCount.Additems = append(itemCount.Additems, one.Additems...)
			// 		itemCount.SubItems = append(itemCount.SubItems, one.SubItems...)
			// 		// itemCount.deleteKey = append(itemCount.deleteKey, one.deleteKey...)
			// 	}
			// 	wg.Done()
			// }
		})

	//查询排除的交易
	// excludeTx := make([]config.ExcludeTx, 0)
	// for i, one := range config.Exclude_Tx {
	// 	if bhvo.BH.Height == one.Height {
	// 		excludeTx = append(excludeTx, config.Exclude_Tx[i])
	// 	}
	// }

	NumCPUTokenChan := make(chan bool, runtime.NumCPU()*6)
	for _, txItr := range bhvo.Txs {
		// engine.Log.Info("开始统计交易:%s", hex.EncodeToString(*txItr.GetHash()))
		//排除的交易不统计
		// for _, two := range excludeTx {
		// 	if bytes.Equal(two.TxByte, *txItr.GetHash()) {
		// 		continue
		// 	}
		// }
		go this.countBalancesTxOne(txItr, bhvo.BH.Height, NumCPUTokenChan, itemsChan)
	}

	wg.Wait()
	// start := time.Now()
	if config.Wallet_txitem_save_db {
		this.notspentBalanceDB.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)
	} else {
		this.notspentBalance.CountTxItem(itemCount, bhvo.BH.Height, bhvo.BH.Time)
	}
	// engine.Log.Info("统计交易 耗时 %s", time.Now().Sub(start))
}

/*
	统计单个交易余额，方便异步统计
*/
func (this *BalanceManager) countBalancesTxOne(txItr TxItr, height uint64, tokenCPU chan bool, itemChan chan *TxItemCount) {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	tokenCPU <- false
	// start := time.Now()
	// itemCount := TxItemCount{
	// 	Additems: make([]*TxItem, 0),
	// 	SubItems: make([]*TxSubItems, 0),
	// }
	// start := time.Now()
	txItr.BuildHash()
	// txHashStr := txItr.GetHashStr()
	//将之前的UTXO标记为已经使用，余额中减去。
	// for _, vin := range *txItr.GetVin() {

	// 	// engine.Log.Info("统计余额 111 %d", txItr.Class())
	// 	//是区块奖励
	// 	if txItr.Class() == config.Wallet_tx_type_mining {
	// 		continue
	// 	}
	// 	// engine.Log.Info("查看vin中的状态 %d", vin.PukIsSelf)
	// 	ok := vin.CheckIsSelf()
	// 	if !ok {
	// 		continue
	// 	}
	// 	// engine.Log.Info("统单易1耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	// 	//查找这个地址的余额列表，没有则创建一个
	// 	itemCount.SubItems = append(itemCount.SubItems, &TxSubItems{
	// 		Txid:      vin.Txid, //utils.Bytes2string(vin.Txid), //  vin.GetTxidStr(),
	// 		VoutIndex: vin.Vout,
	// 		Addr:      *vin.GetPukToAddr(), // utils.Bytes2string(*vin.GetPukToAddr()), // vin.GetPukToAddrStr(),
	// 	})

	// }
	// // engine.Log.Info("统单易3耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	// txCtrl := GetTransactionCtrl(txItr.Class())
	// if txCtrl != nil {

	// 	txCtrl.SyncCount()
	// 	itemChan <- &itemCount
	// 	<-tokenCPU
	// 	return
	// }
	// // engine.Log.Info("统计单个交易 444 耗时 %s", time.Now().Sub(start))

	// //生成新的UTXO收益，保存到列表中
	// for voutIndex, vout := range *txItr.GetVout() {

	// 	//找出需要统计余额的地址

	// 	//和自己无关的地址
	// 	ok := vout.CheckIsSelf()
	// 	// addrInfo, ok := keystore.FindAddress(vout.Address)
	// 	if !ok {
	// 		continue
	// 	}
	// 	//见证人押金和投票押金不计入余额
	// 	if txItr.Class() == config.Wallet_tx_type_deposit_in || txItr.Class() == config.Wallet_tx_type_vote_in {
	// 		if voutIndex == 0 {
	// 			continue
	// 		}
	// 	}

	// 	// engine.Log.Info("统单易5耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	// 	txItem := TxItem{
	// 		Addr: &(*txItr.GetVout())[voutIndex].Address, //  &vout.Address,
	// 		// AddrStr: vout.GetAddrStr(),                      //
	// 		Value: vout.Value,       //余额
	// 		Txid:  *txItr.GetHash(), //交易id
	// 		// TxidStr:      txHashStr,                              //
	// 		VoutIndex:    uint64(voutIndex), //交易输出index，从0开始
	// 		Height:       height,            //
	// 		LockupHeight: vout.FrozenHeight, //锁仓高度
	// 	}

	// 	//计入余额列表
	// 	// this.notspentBalance.AddTxItem(txItem)
	// 	itemCount.Additems = append(itemCount.Additems, &txItem)

	// 	//保存到缓存
	// 	// engine.Log.Info("开始统计交易余额 区块高度 %d 保存到缓存", bhvo.BH.Height)
	// 	// TxCache.AddTxInTxItem(txHashStr, txItr)
	// 	TxCache.AddTxInTxItem(*txItr.GetHash(), txItr)

	// }
	// engine.Log.Info("统计余额及奖励 101010 耗时 %s", time.Now().Sub(start))
	itemCount := txItr.CountTxItems(height)
	itemChan <- itemCount
	// engine.Log.Info("统单易6耗时 %s %s", txItr.GetHashStr(), time.Now().Sub(start))
	<-tokenCPU
}

/*
	统计社区奖励
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countCommunityReward(bhvo *BlockHeadVO) {

	var err error
	var addr crypto.AddressCoin
	var ok bool
	var cs *CommunitySign
	var sn *sqlite3_db.SnapshotReward
	var rt *RewardTotal
	var r *[]sqlite3_db.RewardLight
	for _, txItr := range bhvo.Txs {
		//判断交易类型
		if txItr.Class() != config.Wallet_tx_type_pay {
			continue
		}
		//检查签名
		addr, ok, cs = CheckPayload(txItr)
		if !ok {
			//签名不正确
			continue
		}
		//判断地址是否属于自己
		_, ok = keystore.FindAddress(addr)
		if !ok {
			//签名者地址不属于自己
			continue
		}

		//判断有没有这个快照
		sn, _, err = FindNotSendReward(&addr)
		if err != nil && err.Error() != xorm.ErrNotExist.Error() {
			engine.Log.Error("querying database Error %s", err.Error())
			return
		}
		//同步快照
		if sn == nil || sn.EndHeight < cs.EndHeight {
			//创建快照
			rt, r, err = GetRewardCount(&addr, cs.StartHeight, cs.EndHeight)
			if err != nil {
				return
			}
			err = CreateRewardCount(addr, rt, *r)
			if err != nil {
				return
			}
		}
	}

}

/*
	统计交易历史记录
*/
func (this *BalanceManager) countTxHistory(bhvo *BlockHeadVO) {
	for _, txItr := range bhvo.Txs {

		// start := time.Now()
		txItr.BuildHash()
		// txHashStr := hex.EncodeToString(*txItr.GetHash())
		// engine.Log.Info("统计余额 000")
		//将之前的UTXO标记为已经使用，余额中减去。

		// engine.Log.Info("统计余额及奖励 666 耗时 %s", time.Now().Sub(start))

		txItr.CountTxHistory(bhvo.BH.Height)

	}
}

/*
	统计交易历史记录
*/
// func (this *BalanceManager) countTxHistory(bhvo *BlockHeadVO) {
// 	for _, txItr := range bhvo.Txs {

// 		// start := time.Now()
// 		txItr.BuildHash()
// 		// txHashStr := hex.EncodeToString(*txItr.GetHash())
// 		// engine.Log.Info("统计余额 000")
// 		//将之前的UTXO标记为已经使用，余额中减去。

// 		// engine.Log.Info("统计余额及奖励 666 耗时 %s", time.Now().Sub(start))

// 		txCtrl := GetTransactionCtrl(txItr.Class())
// 		if txCtrl != nil {
// 			// txCtrl.CountBalance(this.notspentBalance, this.otherDeposit, bhvo, uint64(txIndex))
// 			continue
// 		}

// 		// engine.Log.Info("统计余额及奖励 777 耗时 %s", time.Now().Sub(start))

// 		//其他类型交易，自己节点不支持，直接当做普通交易处理

// 		// voutAddrs := make([]*crypto.AddressCoin, 0)
// 		// for i, one := range *txItr.GetVout() {
// 		// 	//如果地址是自己的，就可以不用显示
// 		// 	if keystore.FindAddress(one.Address) {
// 		// 		continue
// 		// 	}
// 		// 	voutAddrs = append(voutAddrs, &(*txItr.GetVout())[i].Address)
// 		// }

// 		// engine.Log.Info("统计余额及奖励 888 耗时 %s", time.Now().Sub(start))

// 		vinAddrs := make([]*crypto.AddressCoin, 0)

// 		hasPayOut := false //是否有支付类型转出记录
// 		//将之前的UTXO标记为已经使用，余额中减去。
// 		for _, vin := range *txItr.GetVin() {

// 			isSelf := vin.CheckIsSelf()

// 			//检查地址是否是自己的
// 			// addrInfo, isSelf := keystore.FindPuk(vin.Puk)

// 			// addr, isSelf := vin.ValidateAddr()
// 			// fmt.Println(addr)
// 			vinAddrs = append(vinAddrs, vin.GetPukToAddr())
// 			if !isSelf {
// 				continue
// 			}

// 			//是区块奖励
// 			if txItr.Class() == config.Wallet_tx_type_mining {
// 				continue
// 			}

// 			switch txItr.Class() {
// 			case config.Wallet_tx_type_mining:
// 			case config.Wallet_tx_type_deposit_in:
// 			case config.Wallet_tx_type_deposit_out:

// 				// if preTxItr.Class() == config.Wallet_tx_type_deposit_in {
// 				// 	if this.depositin != nil {
// 				// 		if bytes.Equal(addrInfo.Addr, *this.depositin.Addr) {
// 				// 			this.depositin = nil
// 				// 		}
// 				// 	}
// 				// }
// 			case config.Wallet_tx_type_pay:
// 				if !hasPayOut {
// 					//和自己相关的输入地址
// 					vinAddrsSelf := make([]*crypto.AddressCoin, 0)
// 					for _, vin := range *txItr.GetVin() {
// 						//检查地址是否是自己的
// 						addrInfo, isSelf := keystore.FindPuk(vin.Puk)
// 						// addr, isSelf := vin.ValidateAddr()
// 						// fmt.Println(addr)
// 						if isSelf {
// 							vinAddrsSelf = append(vinAddrsSelf, &addrInfo.Addr)
// 						}
// 					}

// 					temp := make([]*crypto.AddressCoin, 0) //转出地址
// 					amount := uint64(0)                    //转出金额
// 					for i, one := range *txItr.GetVout() {
// 						//如果地址是自己的，就可以不用显示
// 						// if keystore.FindAddress(one.Address) {
// 						// 	continue
// 						// }
// 						_, ok := keystore.FindAddress(one.Address)
// 						if ok {
// 							continue
// 						}
// 						temp = append(temp, &(*txItr.GetVout())[i].Address)
// 						amount = amount + one.Value
// 					}
// 					if amount > 0 {
// 						//将转出保存历史记录
// 						hi := HistoryItem{
// 							IsIn:    false,         //资金转入转出方向，true=转入;false=转出;
// 							Type:    txItr.Class(), //交易类型
// 							InAddr:  temp,          //输入地址
// 							OutAddr: vinAddrsSelf,  //输出地址
// 							// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
// 							Value:  amount,           //交易金额
// 							Txid:   *txItr.GetHash(), //交易id
// 							Height: bhvo.BH.Height,   //
// 							// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
// 						}
// 						this.chain.history.Add(hi)
// 						// engine.Log.Info("转出记录", bhvo.BH.Height, hi, (*preTxItr.GetVout())[vin.Vout].Value)
// 					}
// 					hasPayOut = true
// 				}
// 			case config.Wallet_tx_type_vote_in:
// 			case config.Wallet_tx_type_vote_out:

// 				// if preTxItr.Class() == config.Wallet_tx_type_vote_in {
// 				// 	votein := preTxItr.(*Tx_vote_in)
// 				// 	b, ok := this.votein.Load(votein.Vote.B58String())
// 				// 	if ok {
// 				// 		ba := b.(*Balance)
// 				// 		ba.Txs.Delete(hex.EncodeToString(*preTxItr.GetHash()) + "_" + strconv.Itoa(int(vin.Vout)))
// 				// 		this.votein.Store(votein.Vote.B58String(), ba)
// 				// 	}
// 				// }
// 			}
// 		}

// 		// engine.Log.Info("统计余额及奖励 999 耗时 %s", time.Now().Sub(start))
// 		//生成新的UTXO收益，保存到列表中
// 		for _, vout := range *txItr.GetVout() {

// 			//和自己无关的地址
// 			// if !keystore.FindAddress(vout.Address) {
// 			// 	continue
// 			// }
// 			ok := vout.CheckIsSelf()
// 			// _, ok := keystore.FindAddress(vout.Address)
// 			if !ok {
// 				continue
// 			}

// 			// fmt.Println("放入内存的txid为", base64.StdEncoding.EncodeToString(*txItr.GetHash()))

// 			switch txItr.Class() {
// 			case config.Wallet_tx_type_mining:
// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  []*crypto.AddressCoin{&vout.Address}, //输入地址
// 					OutAddr: nil,                                  //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				this.chain.history.Add(hi)
// 			case config.Wallet_tx_type_deposit_in:

// 			case config.Wallet_tx_type_deposit_out:
// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  vinAddrs,                             //输入地址
// 					OutAddr: []*crypto.AddressCoin{&vout.Address}, //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				this.chain.history.Add(hi)
// 			case config.Wallet_tx_type_pay:
// 				//判断是否是找零地址，判断依据是输入地址是否有自己钱包的地址
// 				have := false
// 				for _, one := range vinAddrs {
// 					// if keystore.FindAddress(*one) {
// 					// 	have = true
// 					// 	break
// 					// }
// 					_, ok := keystore.FindAddress(*one)
// 					if ok {
// 						have = true
// 						break
// 					}
// 				}

// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  []*crypto.AddressCoin{&vout.Address}, //
// 					OutAddr: vinAddrs,                             //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				if !have {
// 					hi.InAddr = []*crypto.AddressCoin{&vout.Address}
// 					this.chain.history.Add(hi)
// 				}

// 			case config.Wallet_tx_type_vote_in:

// 			case config.Wallet_tx_type_vote_out:
// 				//保存历史记录
// 				//*如果是找零的记录不用保存历史记录
// 				hi := HistoryItem{
// 					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
// 					Type:    txItr.Class(),                        //交易类型
// 					InAddr:  []*crypto.AddressCoin{&vout.Address}, //输入地址
// 					OutAddr: vinAddrs,                             //输出地址
// 					Value:   vout.Value,                           //交易金额
// 					Txid:    *txItr.GetHash(),                     //交易id
// 					Height:  bhvo.BH.Height,                       //
// 					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
// 				}
// 				this.chain.history.Add(hi)
// 			}

// 		}

// 	}
// }

/*
	缴纳押金，并广播
*/
func (this *BalanceManager) DepositIn(amount, gas uint64, pwd, payload string) error {
	addrInfo := keystore.GetCoinbase()

	//不能重复提交押金
	if this.depositin != nil {
		return errors.New("Deposit cannot be paid repeatedly")
	}
	if this.txManager.FindDeposit(addrInfo.Puk) {
		return errors.New("Deposit cannot be paid repeatedly")
	}
	if amount < config.Mining_deposit {
		return errors.New("Deposit not less than" + strconv.Itoa(int(uint64(config.Mining_deposit)/Unit)))
	}

	deposiIn, err := CreateTxDepositIn(amount, gas, pwd, payload)
	if err != nil {
		return err
	}
	if deposiIn == nil {
		return errors.New("Failure to pay deposit")
	}
	deposiIn.BuildHash()
	MulticastTx(deposiIn)

	this.txManager.AddTx(deposiIn)
	return nil
}

/*
	退还押金，并广播
*/
func (this *BalanceManager) DepositOut(addr string, amount, gas uint64, pwd string) error {

	if this.depositin == nil {
		return errors.New("I didn't pay the deposit")
	}

	deposiOut, err := CreateTxDepositOut(addr, amount, gas, pwd)
	if err != nil {
		return err
	}
	if deposiOut == nil {
		return errors.New("Failure to pay deposit")
	}
	deposiOut.BuildHash()

	MulticastTx(deposiOut)

	this.txManager.AddTx(deposiOut)
	return nil
}

/*
	投票押金，并广播
	不能自己给自己投票，统计票的时候会造成循环引用
	给见证人投票的都是社区节点，一个社区节点只能给一个见证人投票。
	给社区节点投票的都是轻节点，轻节点投票前先缴押金。
	轻节点可以给轻节点投票，相当于一个轻节点尾随另一个轻节点投票。
	引用关系不能出现死循环
	@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
*/
func (this *BalanceManager) VoteIn(voteType uint16, witnessAddr crypto.AddressCoin, addr crypto.AddressCoin, amount, gas uint64, pwd, payload string) error {

	//不能自己给自己投票
	if bytes.Equal(witnessAddr, addr) {
		return errors.New("You can't vote for yourself")
	}
	dstAddr := addr

	isWitness := this.witnessBackup.haveWitness(&dstAddr)
	_, isCommunity := this.witnessBackup.haveCommunityList(&dstAddr)
	_, isLight := this.witnessBackup.haveLight(&dstAddr)
	// fmt.Println("查看自己的角色", addr, isWitness, isCommunity, isLight)
	switch voteType {
	case 1: //1=给见证人投票
		if isLight || isWitness {
			return errors.New("The voting address is already another role")
		}
		vs, ok := this.witnessBackup.haveCommunityList(&dstAddr)
		if ok {
			if bytes.Equal(*vs.Witness, witnessAddr) {
				return errors.New("Can't vote again")
			}
			return errors.New("Cannot vote for multiple witnesses")
		}
		//检查押金
		if amount != config.Mining_vote {
			return errors.New("Community node deposit is " + strconv.Itoa(int(config.Mining_vote/1e8)))
		}

	case 2: //2=给社区节点投票

		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		//检查是否成为轻节点
		if !isLight {
			//先成为轻节点
			return errors.New("Become a light node first")
		}

		vs, ok := this.witnessBackup.haveVoteList(&dstAddr)
		if ok {
			if !bytes.Equal(*vs.Witness, witnessAddr) {
				//不能给多个社区节点投票
				return errors.New("Cannot vote for multiple community nodes")
			}
		}

	case 3: //3=轻节点押金
		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		if isLight {
			//已经是轻节点了
			return errors.New("It's already a light node")
		}
		// engine.Log.Info("轻节点押金是 %d %d", amount, config.Mining_light_min)

		if amount != config.Mining_light_min {
			//轻节点押金是
			return errors.New("Light node deposit is " + strconv.Itoa(int(config.Mining_light_min/1e8)))
		}
		witnessAddr = nil
	default:
		//不能识别的投票类型
		return errors.New("Unrecognized voting type")

	}

	voetIn, err := CreateTxVoteIn(voteType, witnessAddr, addr, amount, gas, pwd, payload)
	if err != nil {
		return err
	}
	if voetIn == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	voetIn.BuildHash()
	// bs, err := voetIn.Json()
	// if err != nil {
	// 	//		fmt.Println("33333333333333 33333")
	// 	return err
	// }
	//	fmt.Println("4444444444444444")
	//	fmt.Println("5555555555555555")
	// txbase, err := ParseTxBase(bs)
	// if err != nil {
	// 	return err
	// }
	// voetIn.BuildHash()
	//	fmt.Println("66666666666666")
	//验证交易
	// if err := voetIn.CheckLockHeight(GetHighestBlock()); err != nil {
	// 	return err
	// }
	// if err := voetIn.Check(); err != nil {
	// 	//交易不合法，则不发送出去
	// 	// fmt.Println("交易不合法，则不发送出去")
	// 	return err
	// }
	MulticastTx(voetIn)
	this.txManager.AddTx(voetIn)
	// fmt.Println("添加投票押金是否成功", ok)
	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	//	fmt.Println("7777777777777777")
	return nil
}

// /*
// 	退还一笔投票押金，并广播
// */
// func (this *BalanceManager) VoteOutOne(txid, addr string, amount, gas uint64, pwd string) error {
// 	tx := this.GetVoteInByTxid(txid)
// 	if tx == nil {
// 		return errors.New("没有找到这个交易")
// 	}
// 	deposiOut, err := CreateTxVoteOutOne(tx, addr, amount, gas, pwd)
// 	if err != nil {
// 		return err
// 	}
// 	if deposiOut == nil {
// 		//		fmt.Println("33333333333333 22222")
// 		return errors.New("退押金失败")
// 	}
// 	deposiOut.BuildHash()
// 	bs, err := deposiOut.Json()
// 	if err != nil {
// 		//		fmt.Println("33333333333333 33333")
// 		return err
// 	}
// 	//	fmt.Println("4444444444444444")
// 	MulticastTx(bs)
// 	//	fmt.Println("5555555555555555")
// 	txbase, err := ParseTxBase(bs)
// 	if err != nil {
// 		return err
// 	}
// 	txbase.BuildHash()
// 	//	fmt.Println("66666666666666")
// 	//验证交易
// 	if !txbase.Check() {
// 		//交易不合法，则不发送出去
// 		// fmt.Println("交易不合法，则不发送出去")
// 		return errors.New("交易不合法，则不发送出去")
// 	}
// 	this.txManager.AddTx(txbase)
// 	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
// 	//	fmt.Println("7777777777777777")
// 	return nil
// 	return nil

// }

/*
	退还投票押金，并广播
*/
func (this *BalanceManager) VoteOut(txid []byte, addr crypto.AddressCoin, amount, gas uint64, pwd string) error {

	// waddr := witnessAddr

	// if witnessAddr != nil && witnessAddr.B58String() != "" {
	// 	waddr = witnessAddr
	// }
	// engine.Log.Info("---------------------------查询这个见证人" + waddr + "end")
	// balance := this.GetVoteIn(*waddr)
	// if balance == nil {
	// 	//没有对这个见证人投票
	// 	return errors.New("No vote for this witness")
	// }

	deposiOut, err := CreateTxVoteOut(txid, addr, amount, gas, pwd)
	if err != nil {
		return err
	}
	if deposiOut == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	deposiOut.BuildHash()
	MulticastTx(deposiOut)
	this.txManager.AddTx(deposiOut)
	return nil
}

/*
	构建一个其他交易，并广播
*/
func (this *BalanceManager) BuildOtherTx(class uint64, srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, params ...interface{}) (TxItr, error) {

	ctrl := GetTransactionCtrl(class)
	txItr, err := ctrl.BuildTx(this.notspentBalance, this.otherDeposit, srcAddr, addr, amount, gas, frozenHeight, pwd, comment, params...)
	if err != nil {
		return nil, err
	}
	txItr.BuildHash()
	MulticastTx(txItr)

	this.txManager.AddTx(txItr)
	return txItr, nil
}

/*
	获得自己轻节点押金列表
*/
func (this *BalanceManager) GetVoteList() []*Balance {
	balances := make([]*Balance, 0)
	this.votein.Range(func(k, v interface{}) bool {
		b := v.(*Balance)
		balances = append(balances, b)
		return true
	})
	return balances
}
