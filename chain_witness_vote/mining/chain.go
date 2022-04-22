package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hyahm/golog"
)

/*
	查找一个地址的余额
*/
func FindSurplus(addr utils.Multihash) uint64 {
	return 0
}

/*
	查找最后一组旷工地址
*/
func FindLastGroupMiner() []utils.Multihash {
	return []utils.Multihash{}
}

type Chain struct {
	No                 uint64              //分叉编号
	StartingBlock      uint64              //区块开始高度
	StartBlockTime     uint64              //起始区块时间
	CurrentBlock       uint64              //内存中已经同步到的区块高度
	PulledStates       uint64              //正在同步的区块高度
	SyncBlockLock      *sync.RWMutex       //同步区块锁
	SyncBlockFinish    bool                //同步区块是否完成，关系能否接收新的区块
	witnessBackup      *WitnessBackup      //备用见证人
	witnessChain       *WitnessChain       //见证人组链
	balance            *BalanceManager     //
	transactionManager *TransactionManager //交易管理器
	// history            *BalanceHistory     //
	StopSyncBlock uint32       //当区块分叉或出现重大错误时候，终止同步功能。初始值为0,1=暂停
	Temp          *BlockHeadVO //收到的区块高度最大的区块
}

/*
	获取区块开始高度
*/
func (this *Chain) GetStartingBlock() uint64 {
	return atomic.LoadUint64(&this.StartingBlock)
}

/*
	获取起始区块出块时间
*/
func (this *Chain) GetStartBlockTime() uint64 {
	return atomic.LoadUint64(&this.StartBlockTime)
}

/*
	设置区块开始高度
*/
func (this *Chain) SetStartingBlock(n, startBlockTime uint64) {
	atomic.StoreUint64(&this.StartingBlock, n)
	atomic.StoreUint64(&this.StartBlockTime, startBlockTime)
}

/*
	获取已经同步到的区块高度
*/
func (this *Chain) GetCurrentBlock() uint64 {
	return atomic.LoadUint64(&this.CurrentBlock)
}

func (this *Chain) SetCurrentBlock(n uint64) {
	// engine.Log.Warn("设置同步区块高度 %d", n)
	atomic.StoreUint64(&this.CurrentBlock, n)
}

/*
	获取正在同步的区块高度
*/
func (this *Chain) GetPulledStates() uint64 {
	return atomic.LoadUint64(&this.PulledStates)
}

func (this *Chain) SetPulledStates(n uint64) {
	// engine.Log.Warn("设置正在同步的区块高度 %d", n)
	atomic.StoreUint64(&this.PulledStates, n)
}

/*
	获取区块组高度
*/
// func (this *Chain) GetGroupHeights() uint64 {
// 	// if forks.GetLongChain() == nil {
// 	// 	return 0
// 	// }
// 	return this.GetLastBlock().Group.Height
// }

func NewChain() *Chain {
	chain := &Chain{}
	chain.StopSyncBlock = 0
	chain.SyncBlockLock = new(sync.RWMutex)
	chain.SyncBlockFinish = false
	// chain.No = no
	wb := NewWitnessBackup(chain)
	wc := NewWitnessChain(wb, chain)
	tm := NewTransactionManager(wb)
	b := NewBalanceManager(wb, tm, chain)

	// chain.lastBlock = block
	chain.witnessBackup = wb
	chain.witnessChain = wc
	chain.balance = b
	chain.transactionManager = tm
	// chain.history = NewBalanceHistory()
	// go chain.GoSyncBlock()
	utils.Go(chain.GoSyncBlock)
	return chain
}

/*
	克隆一个链
*/
//func (this *Chain) Clone() *Chain {
//Chain{
//	witnessBackup : this.witnessBackup ,     *WitnessBackup      //备用见证人
//	witnessChain       *WitnessChain       //见证人组链
//	lastBlock          *Block              //最新块
//	balance            *BalanceManager     //
//	transactionManager *TransactionManager //交易管理器
//}
//}

type Group struct {
	PreGroup  *Group   //前置组
	NextGroup *Group   //下一个组，有分叉，下标为0的是最长链
	Height    uint64   //组高度
	Blocks    []*Block //组中的区块
}

type Block struct {
	Id         []byte   //区块id
	PreBlockID []byte   //前置区块id
	PreBlock   *Block   //前置区块高度
	NextBlock  *Block   //下一个区块高度
	Group      *Group   //所属组
	Height     uint64   //区块高度
	witness    *Witness //是哪个见证人出的块
	// IdStr      string   //
	// LocalTime time.Time //
}

// func (this *Block) GetIdStr() string {
// 	if this.IdStr == "" {
// 		this.IdStr = hex.EncodeToString(this.Id)
// 	}
// 	return this.IdStr
// }

func (this *Block) Load() (*BlockHead, error) {

	//先判断缓存中是否存在
	blockHead, ok := TxCache.FindBlockHeadCache(this.Id)
	if !ok {
		// // engine.Log.Error("未命中缓存 111 FindBlockHeadCache")
		// bh, err := db.Find(this.Id)
		// if err != nil {
		// 	//		if err == leveldb.ErrNotFound {
		// 	//			return
		// 	//		} else {
		// 	//		}
		// 	return nil, err
		// }
		// blockHead, err = ParseBlockHead(bh)
		var err error
		blockHead, err = LoadBlockHeadByHash(&this.Id)
		if err != nil {
			return nil, err
		}
		TxCache.AddBlockHeadCache(this.Id, blockHead)
	}

	return blockHead, nil
}

/*
	加载本区块的所有交易
*/
func (this *Block) LoadTxs() (*BlockHead, *[]TxItr, error) {
	bh, err := this.Load()
	if err != nil {
		// fmt.Println("加载区块错误", err)
		return nil, nil, err
	}
	txs := make([]TxItr, 0)
	for i, one := range bh.Tx {

		// key := ""
		var txItr TxItr
		ok := false
		//是否启用缓存
		if config.EnableCache {
			// key = hex.EncodeToString(one)
			// key = utils.Bytes2string(one)
			//先判断缓存中是否存在
			txItr, ok = TxCache.FindTxInCache(bh.Tx[i])
		}
		if !ok {
			// engine.Log.Error("未命中缓存 222 FindTxInCache")
			// bs, err := db.Find(one)
			// if err != nil {
			// 	return nil, nil, err
			// }

			// txItr, err = ParseTxBase(ParseTxClass(one), bs)
			var err error
			txItr, err = LoadTxBase(one)
			if err != nil {
				return nil, nil, err
			}
			//如果缓存已经启用，则把交易放入缓存
			if config.EnableCache {
				TxCache.AddTxInCache(bh.Tx[i], txItr)
			}
		}

		txs = append(txs, txItr)
	}
	return bh, &txs, nil
}

/*
	添加新的区块到分叉中
	返回(true,chain)  区块添加到分叉链上
	返回(false,chain) 区块添加到主链上的
	返回(true,nil)    区块分叉超过了区块确认数量
	@return    bool      是否有分叉
	@return    *Chain    区块添加到的链
*/
var AddBlockLock = new(sync.RWMutex)

// var stopBlockHeight = config.Mining_block_start_height + utils.GetRandNum(10) + 3

func (this *Chain) AddBlock(bhvo *BlockHeadVO) error {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)

	start := time.Now()

	AddBlockLock.Lock()
	defer AddBlockLock.Unlock()
	// fmt.Println("保存区块", bhvo)
	// engine.Log.Info("保存区块 00000000000 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

	engine.Log.Info("====== Import block group:%d block:%d prehash:%s hash:%s witness:%s", bhvo.BH.GroupHeight,
		bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash), hex.EncodeToString(bhvo.BH.Hash), bhvo.BH.Witness.B58String())

	// engine.Log.Info("====== Import block group:%d block:%d prehash:%s hash:%s", bhvo.BH.GroupHeight,
	// 	bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash), hex.EncodeToString(bhvo.BH.Hash))

	IsBackup := this.witnessChain.FindWitness(keystore.GetCoinbase().Addr)
	if !IsBackup {
		//区块不连续，卡住了就不再导入区块了
		if atomic.LoadUint32(&this.StopSyncBlock) == 1 {
			engine.Log.Info("not import")
			return nil
		}
	}
	// engine.Log.Info("this block witness %s", bhvo.BH.Witness.B58String())
	// engine.Log.Info("Import block %s", hex.EncodeToString(bhvo.BH.Hash))

	// engine.Log.Info("添加区块 111 耗时 %s", time.Now().Sub(start))
	// engine.Log.Info("添加区块 111 %v", bhvo.FromBroadcast)

	bhvo.BH.BuildBlockHash()

	// if !config.InitNode && bhvo.BH.Height == uint64(stopBlockHeight) {
	// 	stopBlockHeight = 0
	// 	return
	// }

	// engine.Log.Info("保存区块 11111111111111 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

	//先保存区块
	ok, err := SaveBlockHead(bhvo)
	if err != nil {
		engine.Log.Warn("save block error %s", err.Error())
		return err
	}
	// engine.Log.Info("保存区块 3333333333333 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

	if !ok {
		// engine.Log.Info("保存区块 444444444444444 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

		// engine.Log.Info("------------------------------------------------------------ 111111111111111")
		//收到的区块不连续，则从邻居节点同步
		this.NoticeLoadBlockForDB()
		return nil
	}
	// engine.Log.Info("保存区块 55555555555555 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

	// engine.Log.Info("保存区块 6666666666666 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	//检查出块时间和本机时间相比，出块只能滞后，不能提前
	now := utils.GetNow() // time.Now().Unix()
	if bhvo.BH.Time > now+config.Mining_block_time {
		engine.Log.Warn("Build block It's too late %d %d %s", bhvo.BH.Time, now, time.Unix(bhvo.BH.Time, 0).String())
		//出块时间提前了
		return errors.New("Build block It's too late")
	}

	// engine.Log.Info("add block spend time 333 %s", time.Now().Sub(start))

	//自己是见证人，那么要判断出块时间是否提前了
	// if config.AlreadyMining && bhvo.BH.Time < now-(config.Mining_block_time/2) {
	// 	engine.Log.Warn("出块时间提前了 %d %d %s", bhvo.BH.Time, now, time.Unix(bhvo.BH.Time, 0).String())
	// 	//出块时间提前了
	// 	return
	// }

	// engine.Log.Info("保存区块 77777777777777 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	//更新网络广播块高度
	if bhvo.BH.Height > forks.GetHighestBlock() {
		// atomic.StoreUint64(&forks.HighestBlock, bhvo.BH.Height)
		forks.SetHighestBlock(bhvo.BH.Height)
		this.Temp = bhvo
	}

	//排除上链但不合法的交易
	for _, one := range config.Exclude_Tx {
		if bhvo.BH.Height != one.Height {
			continue
		}
		for j, two := range bhvo.Txs {
			if !bytes.Equal(one.TxByte, *two.GetHash()) {
				// engine.Log.Info("交易hash不相同 %d %s %d %s", len(one.TxByte),
				// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
				continue
			}

			// engine.Log.Info("排除交易前 %d 排除第 %d 个交易", len(bhvo.Txs), j)

			notExcludeTx := bhvo.Txs[:j]
			bhvo.Txs = append(notExcludeTx, bhvo.Txs[j+1:]...)

			// engine.Log.Info("排除交易后 %d", len(bhvo.Txs))
			break
		}
	}

	//检查区块是否已经导入过了
	if this.witnessChain.CheckRepeatImportBlock(bhvo) {
		// engine.Log.Info("重复导入区块")
		engine.Log.Info("Repeat import block")
		return ERROR_repeat_import_block
	}

	//查找前置区块是否存在，并且合法
	preWitness := this.witnessChain.FindPreWitnessForBlock(bhvo.BH.Previousblockhash)
	if preWitness == nil {
		//有见证人节点分叉了，还在继续出块，但是分叉的长度没有自己的链长
		if this.GetCurrentBlock() > bhvo.BH.Height {
			return nil
		}
		// this.Temp = bhvo
		// engine.Log.Warn("找不到前置区块，新区块不连续 新高度：%d", bhvo.BH.Height)
		engine.Log.Warn("The front block cannot be found, and the new block is discontinuous with a new height:%d preblockhash:%s v:%+v", bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Previousblockhash), bhvo.BH)
		//这里有可能分叉，有可能有断掉的区块未同步到位，现在都默认为区块未同步到位
		//TODO 有可能产生分叉，处理分叉的情况

		//从邻居节点同步区块
		this.NoticeLoadBlockForDB()
		return ERROR_fork_import_block
	} else {
		//判断区块高度是否连续
		engine.Log.Info("pre witness height:%d bhvo height:%d", preWitness.Block.Height, bhvo.BH.Height)
		if preWitness.Block.Height+1 != bhvo.BH.Height {
			engine.Log.Error("new block height fail,pre height:%d new height:%d", preWitness.Block.Height, bhvo.BH.Height)
			return ERROR_import_block_height_not_continuity
		}
	}

	// engine.Log.Info("add block spend time 444 %s", time.Now().Sub(start))

	//
	var currentWitness *Witness
	var isOverWitnessGroupChain bool
	//是首个区块，这里不构建前面的组
	if bhvo.BH.GroupHeight != config.Mining_group_start_height {
		//检查新区块是否在备用见证人组中
		currentWitness, isOverWitnessGroupChain = this.witnessChain.FindWitnessForBlockOnly(bhvo)
		if bhvo.BH.Height < config.FixBuildGroupBUGHeightMax {
			if currentWitness == nil {
				//当跳过了很多组高度，超出了现有已创建的组高度，获得的currentWitness为空，这里需要创建足够的组高度，用以导入新区块。
				// engine.Log.Info("找不到这个区块的见证人 新高度：%d", bhvo.BH.Height)
				engine.Log.Info("The new height of the witness for this block cannot be found:%d", bhvo.BH.Height)
				//跳过多个组高度，则构建后面的备用见证人组
				//先统计之前的区块
				this.witnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
				// engine.Log.Info("add block spend time 555 %s", time.Now().Sub(start))
				//再构建后面的备用见证人组
				this.witnessChain.CompensateWitnessGroupByGroupHeight(bhvo.BH.GroupHeight)
				// engine.Log.Info("add block spend time 666 %s", time.Now().Sub(start))
				this.witnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
				// engine.Log.Info("add block spend time 777 %s", time.Now().Sub(start))
			}
			this.witnessChain.BuildBlockGroup(bhvo, preWitness)
		} else {
			if isOverWitnessGroupChain {

				this.witnessChain.CompensateWitnessGroupByGroupHeight(bhvo.BH.GroupHeight)

				// currentWitness, isOverWitnessGroupChain = this.witnessChain.FindWitnessForBlockOnly(bhvo)
				// if currentWitness == nil {
				// 	//找不到这个见证人
				// 	return errors.New("not font witness")
				// }
			}
		}

		// if currentWitness == nil {
		// 	//当跳过了很多组高度，超出了现有已创建的组高度，获得的currentWitness为空，这里需要创建足够的组高度，用以导入新区块。
		// 	// engine.Log.Info("找不到这个区块的见证人 新高度：%d", bhvo.BH.Height)
		// 	engine.Log.Info("The new height of the witness for this block cannot be found:%d", bhvo.BH.Height)
		// 	//跳过多个组高度，则构建后面的备用见证人组
		// 	//先统计之前的区块
		// 	this.witnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
		// 	// engine.Log.Info("add block spend time 555 %s", time.Now().Sub(start))
		// 	//再构建后面的备用见证人组
		// 	this.witnessChain.CompensateWitnessGroupByGroupHeight(bhvo.BH.GroupHeight)
		// 	// engine.Log.Info("add block spend time 666 %s", time.Now().Sub(start))
		// 	this.witnessChain.BuildBlockGroupForGroupHeight(bhvo.BH.GroupHeight-1, &bhvo.BH.Previousblockhash)
		// 	// engine.Log.Info("add block spend time 777 %s", time.Now().Sub(start))
		// }
		//当出块跳过了很多个组高度，超出了现有已创建的组高度，上一次获得的currentWitness为空，需要再次获得
		// currentWitness = this.witnessChain.FindWitnessForBlockOnly(bhvo)

		//先统计再构建
		//---------开始统计交易------
		// if bhvo.BH.Height >= config.FixBuildGroupBUGHeightMax {
		// 	//本组有2个以上的见证人出块，那么本组才有可能被确认
		// 	engine.Log.Info("111111111 %d", currentWitness.Group.Height)
		// 	//判断这个组是否多人出块
		// 	ok, group := currentWitness.Group.CheckBlockGroup(nil)
		// 	if ok {
		// 		engine.Log.Info("222222222 %s", hex.EncodeToString(group.Blocks[0].PreBlockID))
		// 		//找到上一个未确认的组最后一个见证人
		// 		witness := this.witnessChain.FindPreWitnessForBlock(group.Blocks[0].PreBlockID)
		// 		engine.Log.Info("3333333333 %s", hex.EncodeToString(witness.Block.Id))
		// 		wg := witness.Group.BuildGroup(&witness.Block.Id)
		// 		// wg := preWitness.Group.BuildGroup(&bhvo.BH.Previousblockhash)
		// 		//找到上一组到本组的见证人组，开始查找没有出块的见证人
		// 		if wg != nil {
		// 			engine.Log.Info("44444444444")
		// 			for _, one := range wg.Witness {
		// 				if !one.CheckIsMining {
		// 					if one.Block == nil {
		// 						this.witnessBackup.AddBlackList(*one.Addr)
		// 					} else {
		// 						this.witnessBackup.SubBlackList(*one.Addr)
		// 					}
		// 					one.CheckIsMining = true
		// 				}
		// 			}
		// 		}
		// 		engine.Log.Info("5555555555")
		// 		this.CountBlock(witness.Group)
		// 	}
		// }

		// // engine.Log.Info("add block spend time 888 %s", time.Now().Sub(start))
		// if bhvo.BH.Height < config.FixBuildGroupBUGHeightMax {
		// 	this.witnessChain.BuildBlockGroup(bhvo, preWitness)
		// }
		// // engine.Log.Info("add block spend time 999 %s", time.Now().Sub(start))
	}

	// engine.Log.Info("add block spend time 555 %s", time.Now().Sub(start))

	currentWitness, _ = this.witnessChain.FindWitnessForBlockOnly(bhvo)
	if currentWitness == nil {
		//找不到这个见证人
		return errors.New("not font witness")
	}
	//验证区块见证人签名
	// if !bhvo.BH.CheckBlockHead(currentWitness.Puk) {
	// 	//区块验证不通过，区块不合法
	// 	engine.Log.Warn("Block verification failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// 	// this.chain.NoticeLoadBlockForDB(false)
	// 	return
	// }

	// engine.Log.Info("add block spend time 666 %s", time.Now().Sub(start))
	//把见证人设置为已出块
	ok = this.witnessChain.SetWitnessBlock(bhvo)
	if !ok {
		//从邻居节点同步
		// engine.Log.Info("------------------------------------------------------------ 22222222222222")
		engine.Log.Debug("Setting witness block failed")
		// this.NoticeLoadBlockForDB(false)
		return errors.New("Setting witness block failed")
	}

	if bhvo.BH.Height >= config.FixBuildGroupBUGHeightMax {
		//是首个区块，这里不构建前面的组
		if bhvo.BH.GroupHeight == config.Mining_group_start_height {
			this.witnessChain.witnessGroup.BuildGroup(nil)
			this.witnessChain.BuildWitnessGroup(false, false)
			this.witnessChain.witnessGroup = this.witnessChain.witnessGroup.NextGroup
			this.witnessChain.BuildWitnessGroup(false, true)
		} else {
			//本组有2个以上的见证人出块，那么本组才有可能被确认
			// engine.Log.Info("111111111 %d", currentWitness.Group.Height)
			//判断这个组是否多人出块
			ok, group := currentWitness.Group.CheckBlockGroup(nil)
			if ok {
				// engine.Log.Info("222222222 %s", hex.EncodeToString(group.Blocks[0].PreBlockID))
				//找到上一个未确认的组最后一个见证人
				witness := this.witnessChain.FindPreWitnessForBlock(group.Blocks[0].PreBlockID)
				// engine.Log.Info("3333333333 %s", hex.EncodeToString(witness.Block.Id))
				wg := witness.Group.BuildGroup(&witness.Block.Id)
				// wg := preWitness.Group.BuildGroup(&bhvo.BH.Previousblockhash)
				//找到上一组到本组的见证人组，开始查找没有出块的见证人
				if wg != nil {
					// engine.Log.Info("44444444444")
					for _, one := range wg.Witness {
						if !one.CheckIsMining {
							if one.Block == nil {
								this.witnessBackup.AddBlackList(*one.Addr)
							} else {
								this.witnessBackup.SubBlackList(*one.Addr)
							}
							one.CheckIsMining = true
						}
					}
				}
				// engine.Log.Info("5555555555")
				this.CountBlock(witness.Group)
				// this.witnessChain.BuildBlockGroup(bhvo, preWitness)

				this.witnessChain.witnessGroup = currentWitness.Group

				this.witnessChain.BuildWitnessGroup(false, true)
			}
		}
	}

	//TODO 检查从此时开始，未来还未出块的见证人组有多少，太少则创建新的组，避免跳过太多组后出块暂停

	// engine.Log.Info("添加区块 666 耗时 %s", time.Now().Sub(start))
	// engine.Log.Info("保存区块 99999999999999 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	// SaveTxToBlockHead(bhvo)
	engine.Log.Info("Save block Time spent %s", time.Now().Sub(start))

	//是首个区块，这里不构建前面的组
	// if bhvo.BH.GroupHeight != config.Mining_group_start_height {
	// 	this.witnessChain.BuildBlockGroup(bhvo, preWitness)
	// }
	// engine.Log.Info("添加区块 777 耗时 %s", time.Now().Sub(start))

	this.witnessChain.BuildMiningTime()

	//回收内存，将前n个见证人之前的见证人链删除。
	this.witnessChain.GCWitnessOld()

	if bhvo.BH.Height == config.Witness_backup_group_overheight {
		config.Witness_backup_group = config.Witness_backup_group_new
	}

	// engine.Log.Info("添加区块 888 耗时 %s", time.Now().Sub(start))

	//删除之前区块的交易缓存
	// TxCache.RemoveHeightTxInCache(bhvo.BH.Height - (config.Mining_group_max * 2))

	//判断自己是否是见证人，自己是见证人，则添加其他见证人白名单连接

	// if this.witnessChain.FindWitness(keystore.GetCoinbase().Addr) {

	// }

	return nil

}

// /*
// 	添加本区块的下一个区块
// */
// func (this *Block) AddNextBlock(bhash []byte) error {
// this.NextBlock
// 	return nil
// }

// /*
// 	修改本区块的下一个区块中最长区块下标为0
// */
// func (this *Block) UpdateNextIndex(bhash [][]byte) error {
// 	this.NextBlock

// 	return nil
// }

/*
	修改next区块顺序
*/
func (this *Block) FlashNextblockhash() error {
	engine.Log.Info("000 update block nextblockhash,this %d blockid:%s nextid:%s", this.Height, hex.EncodeToString(this.Id), hex.EncodeToString(this.NextBlock.Id))
	bh, err := this.Load()
	if err != nil {
		return err
	}

	bh.Nextblockhash = this.NextBlock.Id

	// bs, err := bh.Json()
	bs, err := bh.Proto()
	if err != nil {
		return err
	}

	TxCache.FlashBlockHeadCache(this.Id, bh)

	if bh.Nextblockhash == nil {
		engine.Log.Error("save block nextblockhash nil %s", string(*bs))
	}
	err = db.LevelDB.Save(this.Id, bs)
	if err != nil {
		return err
	}

	// bs, _ = db.Find(this.Id)
	// engine.Log.Info("打印刚保存的区块 \n %s", string(*bs))

	return nil
	// return &bs, nil
}

/*
	统计已经确认的组中的区块
*/
func (this *Chain) CountBlock(witnessGroup *WitnessGroup) {
	// engine.Log.Info("开始统计")

	// start := time.Now()

	//如果本组没有评选出最多出块组，则不统计本组
	if witnessGroup.BlockGroup == nil {
		return
	}
	if witnessGroup.IsCount {
		return
	}

	//获取本组中的出块
	for _, one := range witnessGroup.BlockGroup.Blocks {
		// engine.Log.Info("开始统计 22222222222 %d %d", one.Group.Height, one.Height)

		// engine.Log.Info("统计交易 111 耗时 %s", time.Now().Sub(start))

		//创始块不需要统计
		if one.Height == config.Mining_block_start_height {
			continue
		}
		// engine.Log.Info("统计已经确认的组中的区块 CountBlock")
		bh, txs, err := one.LoadTxs()
		if err != nil {
			//TODO 是个严重的错误
			continue
		}
		bhvo := &BlockHeadVO{BH: bh, Txs: *txs}

		//排除上链但不合法的交易
		for _, one := range config.Exclude_Tx {
			if bhvo.BH.Height != one.Height {
				continue
			}
			for j, two := range bhvo.Txs {
				if !bytes.Equal(one.TxByte, *two.GetHash()) {
					// engine.Log.Info("交易hash不相同 %d %s %d %s", len(one.TxByte),
					// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
					continue
				}

				// engine.Log.Info("排除交易前 %d 排除第 %d 个交易", len(bhvo.Txs), j)

				notExcludeTx := bhvo.Txs[:j]
				bhvo.Txs = append(notExcludeTx, bhvo.Txs[j+1:]...)

				// engine.Log.Info("排除交易后 %d", len(bhvo.Txs))
				break
			}
		}

		//把上一个块的 Nextblockhash 修改为最长链的区块
		one.PreBlock.FlashNextblockhash()

		// engine.Log.Info("统计交易 222 耗时 %s", time.Now().Sub(start))

		//计算余额
		this.balance.CountBalanceForBlock(bhvo)

		// engine.Log.Info("统计交易 333 耗时 %s", time.Now().Sub(start))
		//统计交易中的备用见证人以及见证人投票
		this.witnessBackup.CountWitness(&bhvo.Txs)
		// engine.Log.Info("统计交易 444 耗时 %s", time.Now().Sub(start))
		//删除已经打包了的交易
		this.transactionManager.DelTx(bhvo.Txs)
		// engine.Log.Info("统计交易 555 耗时 %s", time.Now().Sub(start))
		//设置最新高度
		this.SetCurrentBlock(one.Height)
		//清除掉内存中已经过期的交易
		this.transactionManager.CleanIxOvertime(one.Height)
		// engine.Log.Info("统计交易 666 耗时 %s", time.Now().Sub(start))
	}
	witnessGroup.IsCount = true

}

// func (this *Chain) CountBlock() {
// 	// engine.Log.Info("开始统计")

// 	// start := time.Now()

// 	//如果本组没有评选出最多出块组，则不统计本组
// 	if this.witnessChain.witnessGroup.BlockGroup == nil {
// 		return
// 	}

// 	//获取本组中的出块
// 	for _, one := range this.witnessChain.witnessGroup.BlockGroup.Blocks {
// 		// engine.Log.Info("开始统计 22222222222 %d %d", one.Group.Height, one.Height)

// 		// engine.Log.Info("统计交易 111 耗时 %s", time.Now().Sub(start))

// 		//创始块不需要统计
// 		if one.Height == config.Mining_block_start_height {
// 			continue
// 		}
// 		// engine.Log.Info("统计已经确认的组中的区块 CountBlock")
// 		bh, txs, err := one.LoadTxs()
// 		if err != nil {
// 			//TODO 是个严重的错误
// 			continue
// 		}
// 		bhvo := &BlockHeadVO{BH: bh, Txs: *txs}

// 		//排除上链但不合法的交易
// 		for _, one := range config.Exclude_Tx {
// 			if bhvo.BH.Height != one.Height {
// 				continue
// 			}
// 			for j, two := range bhvo.Txs {
// 				if !bytes.Equal(one.TxByte, *two.GetHash()) {
// 					// engine.Log.Info("交易hash不相同 %d %s %d %s", len(one.TxByte),
// 					// 	hex.EncodeToString(one.TxByte), len(*two.GetHash()), hex.EncodeToString(*two.GetHash()))
// 					continue
// 				}

// 				// engine.Log.Info("排除交易前 %d 排除第 %d 个交易", len(bhvo.Txs), j)

// 				notExcludeTx := bhvo.Txs[:j]
// 				bhvo.Txs = append(notExcludeTx, bhvo.Txs[j+1:]...)

// 				// engine.Log.Info("排除交易后 %d", len(bhvo.Txs))
// 				break
// 			}
// 		}

// 		//把上一个块的 Nextblockhash 修改为最长链的区块
// 		one.PreBlock.FlashNextblockhash()

// 		// engine.Log.Info("统计交易 222 耗时 %s", time.Now().Sub(start))

// 		//计算余额
// 		this.balance.CountBalanceForBlock(bhvo)

// 		// engine.Log.Info("统计交易 333 耗时 %s", time.Now().Sub(start))
// 		//统计交易中的备用见证人以及见证人投票
// 		this.witnessBackup.CountWitness(&bhvo.Txs)
// 		// engine.Log.Info("统计交易 444 耗时 %s", time.Now().Sub(start))
// 		//删除已经打包了的交易
// 		this.transactionManager.DelTx(bhvo.Txs)
// 		// engine.Log.Info("统计交易 555 耗时 %s", time.Now().Sub(start))
// 		//设置最新高度
// 		this.SetCurrentBlock(one.Height)
// 		//清除掉内存中已经过期的交易
// 		this.transactionManager.CleanIxOvertime(one.Height)
// 		// engine.Log.Info("统计交易 666 耗时 %s", time.Now().Sub(start))
// 		// one.witness.
// 	}

// }

/*
	统计已经确认的组中的区块
*/
// func (this *Chain) FlashBlock() {
// 	//如果本组没有评选出最多出块组，则不统计本组
// 	if this.witnessChain.witnessGroup.BlockGroup == nil {
// 		return
// 	}
// }

/*
	统计分叉链上的区块
	@bh    *BlockHead    分叉点区块
	@hs    [][]byte      分叉链hash路径
*/
// func (this *Chain) CountForkBlock(block *Block, hs [][]byte) bool {
// 	block.Load()
// }

// /*
// 	回滚一个区块
// 	@height    uint64    要回滚的区块高度
// */
// func (this *Chain) RollbackBlock(height uint64) {
// 	// fmt.Println("开始回滚区块，回滚区块高度", height)
// 	block := this.GetLastBlock()
// 	for height < block.Height {
// 		block = block.PreBlock[0]
// 	}
// 	bh, txs, err := block.LoadTxs()
// 	if err != nil {
// 		return
// 	}

// 	bhvo := &BlockHeadVO{BH: bh, Txs: *txs}
// 	//回滚余额
// 	this.balance.RollbackBalance(bhvo)

// 	//统计交易中的备用见证人以及见证人投票
// 	this.witnessBackup.RollbackCountWitness(txs)

// 	//把见证人设置为已出块
// 	//	this.witnessChain.SetWitnessBlock(this.GetLastBlock())

// 	//回滚已经打包了的交易
// 	this.transactionManager.AddTxs(bhvo.Txs...)

// }

/*
	获取本链已经确认的最高区块
*/
func (this *Chain) GetLastBlock() (witness *Witness, block *Block) {
	// fmt.Println("查询最后一个区块", this.witnessChain.witnessGroup)

	witnessGroup := this.witnessChain.witnessGroup
	if witnessGroup == nil {
		return
	}

	if witnessGroup.Height != config.Mining_group_start_height {
		// if this.witnessChain.witnessGroup != nil && this.witnessChain.witnessGroup.PreGroup != nil {
		// 	// fmt.Println("查询最后一个区块 11111111111")
		// }
		for {
			witnessGroup = witnessGroup.PreGroup
			if witnessGroup == nil {
				break
			}
			//找到合法的区块组
			if witnessGroup.BlockGroup != nil {
				break
			}
		}
	}

	block = witnessGroup.BlockGroup.Blocks[len(witnessGroup.BlockGroup.Blocks)-1]
	witness = block.witness

	// for _, one := range witnessGroup.Witness {
	// 	// fmt.Println("查询最后一个区块 22222222222222")
	// 	if one.Block != nil {
	// 		// fmt.Println("查询最后一个区块 333333333333")
	// 		witness = one
	// 		block = one.Block
	// 	}
	// }
	return
	// return this.witnessChain.witness.Block
}

/*
	获取本链的收益管理器
*/
func (this *Chain) GetBalance() *BalanceManager {
	return this.balance
}

/*
	打印块列表
*/
func (this *Chain) PrintBlockList() {

	//	start := this.GetLastBlock()
	//	for {
	//		if start.PreBlock == nil {
	//			break
	//		}
	//		start = start.PreBlock
	//	}
	//	for {
	//		fmt.Println("打印块列表", start.Height)
	//		if start.NextBlock == nil {
	//			break
	//		}
	//		start = start.NextBlock
	//	}
}

/*
	依次获取前n个区块的hash，连接起来做一次hash
*/
func (this *Chain) HashRandom() *[]byte {
	_, lastBlock := this.GetLastBlock()

	// if lastBlock == nil || lastBlock.Height > config.RandomHashHeightMin {
	// 	// bs := make([]byte, 0)
	// 	// bs = utils.Hash_SHA3_256(bs)
	// 	// return &bs
	// 	return &config.RandomHashFixed
	// }

	// if lastBlock != nil {
	// 	engine.Log.Info("采用区块高度:%d hash:%s", lastBlock.Height, lastBlock.GetIdStr())
	// }
	if lastBlock != nil {

		// engine.Log.Info("lastBlock :%s", hex.EncodeToString(lastBlock.Id))
		if random, ok := config.RandomMap.Load(utils.Bytes2string(lastBlock.Id)); ok {
			bs := random.(*[]byte)
			// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(*bs))
			return bs
		}
	}

	var preHash *[]byte
	bs := make([]byte, 0)
	// witness := this.witnessChain.witness
	//链端初始化时候，this.witnessChain.witness为nil
	for i := 0; lastBlock != nil && i < config.Mining_block_hash_count; i++ {
		if lastBlock.Height < config.NextHashHeightMax {
			if preHash == nil {
				if random, ok := config.NextHash.Load(utils.Bytes2string(lastBlock.Id)); ok {
					bsOne := random.(*[]byte)
					bs = append(bs, *bsOne...)
					preHash = bsOne
					// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(*bs))
					// return bs
					continue
				}
			} else {
				if random, ok := config.NextHash.Load(utils.Bytes2string(*preHash)); ok {
					bsOne := random.(*[]byte)
					bs = append(bs, *bsOne...)
					preHash = bsOne
					// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(*bs))
					// return bs
					continue
				}
			}
		}

		bs = append(bs, lastBlock.Id...)
		if lastBlock.PreBlock == nil {
			break
		}
		lastBlock = lastBlock.PreBlock
	}
	// engine.Log.Info("HashRandom :%s", hex.EncodeToString(bs))
	bs = utils.Hash_SHA3_256(bs)
	if lastBlock != nil {
		golog.Infof("lastBlock height:%d HashRandom hash:%s", lastBlock.Height, hex.EncodeToString(bs))
		engine.Log.Info("lastBlock height:%d HashRandom hash:%s", lastBlock.Height, hex.EncodeToString(bs))
	}
	// engine.Log.Info("HashRandom hash:%s", hex.EncodeToString(bs))
	return &bs
}

// func SetCurrentBlock(n uint64) {
// 	atomic.StoreUint64(&forks.CurrentBlock, n)
// }

// func GetCurrentBlock(n uint64) {
// 	atomic.StoreUint64(&forks.CurrentBlock, n)
// }

/*
	查询历史交易记录
*/
func (this *Chain) GetHistoryBalance(start *big.Int, total int) []HistoryItem {
	// return this.history.Get(start, total)
	return balanceHistoryManager.Get(start, total)
}
