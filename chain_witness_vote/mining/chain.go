package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"encoding/hex"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
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
	CurrentBlock       uint64              //内存中已经同步到的区块高度
	PulledStates       uint64              //正在同步的区块高度
	SyncBlockFinish    bool                //同步区块是否完成
	witnessBackup      *WitnessBackup      //备用见证人
	witnessChain       *WitnessChain       //见证人组链
	balance            *BalanceManager     //
	transactionManager *TransactionManager //交易管理器
	history            *BalanceHistory     //
}

/*
	获取区块开始高度
*/
func (this *Chain) GetStartingBlock() uint64 {
	return atomic.LoadUint64(&this.StartingBlock)
}

/*
	获取区块开始高度
*/
func (this *Chain) SetStartingBlock(n uint64) {
	atomic.StoreUint64(&this.StartingBlock, n)
}

/*
	获取已经同步到的区块高度
*/
func (this *Chain) GetCurrentBlock() uint64 {
	return atomic.LoadUint64(&this.CurrentBlock)
}

func (this *Chain) SetCurrentBlock(n uint64) {
	atomic.StoreUint64(&this.CurrentBlock, n)
}

/*
	获取正在同步的区块高度
*/
func (this *Chain) GetPulledStates() uint64 {
	return atomic.LoadUint64(&this.PulledStates)
}

func (this *Chain) SetPulledStates(h uint64) {
	atomic.StoreUint64(&this.PulledStates, h)
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
	chain.history = NewBalanceHistory()
	go chain.GoSyncBlock()
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
	// LocalTime time.Time //
}

func (this *Block) Load() (*BlockHead, error) {
	// fmt.Println("查询的区块id", hex.EncodeToString(this.Id))
	bh, err := db.Find(this.Id)
	if err != nil {
		//		if err == leveldb.ErrNotFound {
		//			return
		//		} else {
		//		}
		return nil, err
	}

	blockHead, err := ParseBlockHead(bh)
	if err != nil {
		return nil, err
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
	for _, one := range bh.Tx {
		bs, err := db.Find(one)
		if err != nil {
			// fmt.Println("加载区块的交易错误", hex.EncodeToString(one), err)
			// panic("123")
			return nil, nil, err
		}
		txItr, err := ParseTxBase(bs)
		if err != nil {
			return nil, nil, err
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

func (this *Chain) AddBlock(bhvo *BlockHeadVO) {
	AddBlockLock.Lock()
	defer AddBlockLock.Unlock()
	// fmt.Println("保存区块", bhvo)
	// engine.Log.Info("保存区块 00000000000 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	engine.Log.Info("Import block group:%d block:%d hash:%s", bhvo.BH.GroupHeight, bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Hash))
	engine.Log.Info("this block witness %s", bhvo.BH.Witness.B58String())
	// engine.Log.Info("Import block %s", hex.EncodeToString(bhvo.BH.Hash))

	start := time.Now()
	//已经保存了，就不需要再保存
	if !db.CheckHashExist(bhvo.BH.Hash) {
		// engine.Log.Info("保存区块 11111111111111 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

		ok, err := saveBlockHead(bhvo)
		if err != nil {
			engine.Log.Warn("save block error %s", err.Error())
			return
		}
		// engine.Log.Info("保存区块 3333333333333 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

		if !ok {
			// engine.Log.Info("保存区块 444444444444444 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

			// engine.Log.Info("------------------------------------------------------------ 111111111111111")
			//收到的区块不连续，则从邻居节点同步
			this.NoticeLoadBlockForDB(false)
			return
		}
		// engine.Log.Info("保存区块 55555555555555 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	}
	// engine.Log.Info("保存区块 6666666666666 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	//检查出块时间和本机时间相比，出块只能滞后，不能提前
	now := utils.GetNow() // time.Now().Unix()
	if bhvo.BH.Time > now+config.Mining_block_time {
		engine.Log.Warn("Build block It's too late %d %d %s", bhvo.BH.Time, now, time.Unix(bhvo.BH.Time, 0).String())
		//出块时间提前了
		return
	}
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
		// fmt.Println("更新网络区块高度  1111")
		forks.SetHighestBlock(bhvo.BH.Height)
	}

	//是首个区块，这里不构建前面的组
	if bhvo.BH.GroupHeight != config.Mining_group_start_height {
		this.witnessChain.BuildBlockGroup(bhvo)
	}

	//把见证人设置为已出块
	ok := this.witnessChain.SetWitnessBlock(bhvo)
	if !ok {
		//从邻居节点同步
		// engine.Log.Info("------------------------------------------------------------ 22222222222222")
		engine.Log.Debug("Setting witness block failed")
		// this.NoticeLoadBlockForDB(false)
		return
	}
	// engine.Log.Info("保存区块 99999999999999 group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)

	engine.Log.Info("Save block Time spent %s", time.Now().Sub(start))

	//是首个区块，这里不构建前面的组
	if bhvo.BH.GroupHeight != config.Mining_group_start_height {
		this.witnessChain.BuildBlockGroup(bhvo)
	}

	this.witnessChain.BuildMiningTime(false)

	return

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
	// engine.Log.Info("修改本区块的next 本节点id %s next id: %s", hex.EncodeToString(this.Id), hex.EncodeToString(this.NextBlock.Id))
	bh, err := this.Load()
	if err != nil {
		return err
	}

	bh.Nextblockhash = this.NextBlock.Id

	bs, err := bh.Json()
	if err != nil {
		return err
	}
	err = db.Save(this.Id, bs)
	if err != nil {
		return err
	}

	bs, _ = db.Find(this.Id)
	// engine.Log.Info("打印刚保存的区块 \n %s", string(*bs))

	return nil
	// return &bs, nil
}

/*
	统计已经确认的组中的区块
*/
func (this *Chain) CountBlock() {
	// engine.Log.Info("开始统计")

	//如果本组没有评选出最多出块组，则不统计本组
	if this.witnessChain.witnessGroup.BlockGroup == nil {
		return
	}

	//获取本组中的出块
	for _, one := range this.witnessChain.witnessGroup.BlockGroup.Blocks {
		// engine.Log.Info("开始统计 22222222222 %d %d", one.Group.Height, one.Height)
		//创始块不需要统计
		if one.Height == config.Mining_block_start_height {
			continue
		}

		bh, txs, err := one.LoadTxs()
		if err != nil {
			//TODO 是个严重的错误
			continue
		}
		bhvo := &BlockHeadVO{BH: bh, Txs: *txs}
		//把上一个块的 Nextblockhash 修改为最长链的区块
		one.PreBlock.FlashNextblockhash()

		//计算余额
		this.balance.CountBalanceForBlock(bhvo)
		//统计交易中的备用见证人以及见证人投票
		this.witnessBackup.CountWitness(&bhvo.Txs)
		//删除已经打包了的交易
		this.transactionManager.DelTx(bhvo.Txs)
		//设置最新高度
		this.SetCurrentBlock(one.Height)
		//清除掉内存中已经过期的交易
		this.transactionManager.CleanIxOvertime(one.Height)
	}

}

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
	获取本链最高的区块
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
	// this.GetLastBlock()
	_, lastBlock := this.GetLastBlock()
	// fmt.Println("lastBlock", lastBlock)

	bs := make([]byte, 0)
	// witness := this.witnessChain.witness
	//链端初始化时候，this.witnessChain.witness为nil
	for i := 0; lastBlock != nil && i < config.Mining_block_hash_count; i++ {
		bs = append(bs, lastBlock.Id...)
		if lastBlock.PreBlock == nil {
			break
		}
		lastBlock = lastBlock.PreBlock
	}
	bs = utils.Hash_SHA3_256(bs)
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
	return this.history.Get(start, total)
}
