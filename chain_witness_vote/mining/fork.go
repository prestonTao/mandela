/*
	区块分叉管理
*/
package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"errors"
	"sync/atomic"
)

var forks = new(Forks)

func init() {
	// forks.chainss = new(sync.Map)
}

type Forks struct {
	Init bool //是否是创世节点
	// MaxForkNo    uint64 //分叉链自增长最大编号
	LongChain    *Chain //最高区块引用
	HighestBlock uint64 //网络节点广播的区块最高高度
	// chainss      *sync.Map //保存各个分叉链key:string=链最高块hash;value:*Block=各个分叉链引用;
}

/*
	添加一个区块到添加队列中去
*/
func (this *Forks) AddBlockHead(bhvo *BlockHeadVO) error {

	chain := forks.GetLongChain()

	//添加到内存
	// start := time.Now()
	return chain.AddBlock(bhvo)
	// engine.Log.Info("导入区块耗时 %s", time.Now().Sub(start))

}

/*
	获得最长链
*/
func (this *Forks) GetLongChain() *Chain {
	return this.LongChain
}

/*
	获得最长链
*/
// func (this *Forks) FindChain(bh []byte) *Chain {
// 	v, ok := this.chainss.Load(hex.EncodeToString(bh))
// 	if ok {
// 		c := v.(*Chain)
// 		return c
// 	}
// 	return nil
// }

/*
	获得最长链
*/
func GetLongChain() *Chain {
	return forks.LongChain
}

/*
	通过从邻居节点获取创始块来创建链端
*/
func GetFirstBlock() error {
	// engine.Log.Info("获取起始区块")
	//获得本节点的最新块失败，重新同步
	//从令居节点同步起始区块hash值
	chainInfo := FindStartBlockForNeighbor()
	if chainInfo == nil {
		return errors.New("Synchronization start block hash failed")
	}
	// engine.Log.Info("同步到的创世区块hash %s", hex.EncodeToString(chainInfo.StartBlockHash))
	db.LevelDB.Save(config.Key_block_start, &chainInfo.StartBlockHash)
	config.StartBlockHash = chainInfo.StartBlockHash

	peerBlockinfo, _ := FindRemoteCurrentHeight()
	bhvo, _ := syncBlockFlashDB(&chainInfo.StartBlockHash, peerBlockinfo)
	if bhvo == nil {
		return nil
	}

	bhvo.BH.BuildBlockHash()
	// engine.Log.Info("打印同步到的区块 %s", hex.EncodeToString(bhvo.BH.Hash))
	BuildFirstChain(bhvo)

	//
	if forks.GetHighestBlock() < bhvo.BH.Height {
		forks.SetHighestBlock(bhvo.BH.Height)
	}

	return nil
}

/*
	通过已有的创始块启动链端
*/
func BuildFirstChain(bhvo *BlockHeadVO) {
	forks.buildFirstChain(bhvo)

}

/*
	通过创始块创建一个链
*/
func (this *Forks) buildFirstChain(bhvo *BlockHeadVO) {

	newChain := NewChain()
	newChain.SetStartingBlock(bhvo.BH.Height, uint64(bhvo.BH.Time))
	// newChain.StartingBlock = bhvo.BH.Height

	this.LongChain = newChain

	//计算余额
	newChain.balance.CountBalanceForBlock(bhvo)

	//统计交易中的备用见证人以及见证人投票
	newChain.witnessBackup.CountWitness(&bhvo.Txs)

	//
	newChain.witnessChain.BuildWitnessGroup(true, true)

	//把见证人设置为已出块
	newChain.witnessChain.SetWitnessBlock(bhvo)

	newChain.witnessChain.BuildBlockGroup(bhvo, nil)

	// SaveTxToBlockHead(bhvo)

	return
}

/*
	获得最长链
*/
//func (this *Forks) GetChain(beforeHash string) *Chain {
//	chainItr, ok := this.chains.Load(beforeHash)
//	if ok {
//		chain := chainItr.(*Chain)
//		return chain
//	}
//	return nil
//}

// /*
// 	判断分叉链是否长于当前最长链
// 	如果分叉链长于当前链，则找出分叉链从主链上的分叉路径
// 	@chain        *Chain      最高链
// 	@forkChain    *Chain      分叉链
// 	@return       bool        是否分叉链最长
// 	@return       [][]byte    分叉链区块头hash
// */
// func (this *Forks) ContrastLongBlock(chain *Chain) (ok bool, hs [][]byte) {
// 	//判断最新块是不是添加在最长链上
// 	if bytes.Equal(this.LongChain.GetLastBlock().Id, chain.GetLastBlock().Id) {
// 		//
// 		return false, nil
// 	} else {
// 		if this.LongChain.GetLastBlock().Height >= chain.GetLastBlock().Height {
// 			return false, nil
// 		}
// 		//保存主链所有未确认块hash
// 		hs := make([][]byte, 0)
// 		oneBlock := this.LongChain.GetLastBlock()
// 		groupHeight := oneBlock.Group.Height
// 		for i := 0; i < config.Block_confirm; {
// 			hs = append(hs, oneBlock.Id)
// 			oneBlock = oneBlock.PreBlock[0]
// 			if oneBlock.Group.Height < groupHeight {
// 				i++
// 				groupHeight = oneBlock.Group.Height
// 			}
// 		}
// 		//保存分叉链hash值
// 		forkBlockHashs := make([][]byte, 0)
// 		//找到主链和分叉链的分叉点
// 		oneBlock = this.LongChain.GetLastBlock()
// 		groupHeight = oneBlock.Group.Height
// 		//分叉链最多查找未确认的块，如果找完都未找到，则是应该是被删除的链，有问题
// 		for i := 0; i < config.Block_confirm; {
// 			forkBlockHashs = append(forkBlockHashs, oneBlock.Id)
// 			oneBlock = oneBlock.PreBlock[0]
// 			if len(oneBlock.NextBlock) > 1 {
// 				//找到分叉点，和主链上的块对比
// 				for _, one := range hs {
// 					if bytes.Equal(one, oneBlock.Id) {
// 						//找到了分叉点
// 						return true, forkBlockHashs
// 					}
// 				}
// 			}

// 			if oneBlock.Group.Height < groupHeight {
// 				i++
// 				groupHeight = oneBlock.Group.Height
// 			}
// 		}
// 		return true, nil
// 	}

// }

// /*
// 	查找目标链和主链的交叉点，返回分叉链区块
// 	@return    uint64      主链回滚区块数量
// 	@return    [][]byte    新分叉主链区块路径，从区块高度由高到低的顺序返回区块hash
// */
// func (this *Forks) FindIntersection(forkBlock *Block) (uint64, [][]byte) {
// 	//保存主链所有未确认块hash
// 	hs := make([][]byte, 0)
// 	oneBlock := this.LongChain.GetLastBlock()
// 	groupHeight := oneBlock.Group.Height
// 	for i := 0; i < config.Block_confirm; {
// 		hs = append(hs, oneBlock.Id)
// 		oneBlock = oneBlock.PreBlock[0]
// 		if oneBlock.Group.Height < groupHeight {
// 			i++
// 			groupHeight = oneBlock.Group.Height
// 		}
// 	}
// 	//保存分叉链hash值
// 	forkBlockHashs := make([][]byte, 0)
// 	//找到主链和分叉链的分叉点
// 	oneBlock = forkBlock
// 	groupHeight = oneBlock.Group.Height
// 	//分叉链最多查找未确认的块，如果找完都未找到，则是应该是被删除的链，有问题
// 	for i := 0; i < config.Block_confirm; {
// 		forkBlockHashs = append(forkBlockHashs, oneBlock.Id)
// 		oneBlock = oneBlock.PreBlock[0]
// 		if len(oneBlock.NextBlock) > 1 {
// 			//找到分叉点，和主链上的块对比
// 			for j, one := range hs {
// 				if bytes.Equal(one, oneBlock.Id) {
// 					//找到了分叉点
// 					return uint64(j + 1), forkBlockHashs
// 				}
// 			}
// 		}

// 		if oneBlock.Group.Height < groupHeight {
// 			i++
// 			groupHeight = oneBlock.Group.Height
// 		}
// 	}
// 	return 0, forkBlockHashs
// }

// /*
// 	选择最长链，分叉链最长就回滚
// */
// func (this *Forks) SelectLongChain() {
// 	var n uint64
// 	var hs [][]byte
// 	this.chainss.Range(func(k, v interface{}) bool {
// 		//		chain := v.(*Chain)
// 		chain := v.(*Chain)
// 		block := chain.GetLastBlock()
// 		if bytes.Equal(this.LongChain.GetLastBlock().Id, block.Id) {
// 			return true
// 		}
// 		if block.Height > this.LongChain.GetLastBlock().Height {
// 			// fmt.Println("选择最长区块", block.Height, this.LongChain.GetLastBlock().Height)
// 			n, hs = this.FindIntersection(block)
// 			//找到分叉点区块
// 			return false
// 		}

// 		return true
// 	})
// 	if n <= 0 {
// 		return
// 	}

// 	// fmt.Println("开始回滚区块", hs)
// 	//找到分叉点区块
// 	forkBlock := this.LongChain.GetLastBlock()
// 	for i := uint64(0); i < n; i++ {
// 		forkBlock = forkBlock.PreBlock[0]
// 	}
// 	//验证分叉点区块
// 	if !bytes.Equal(forkBlock.Id, hs[len(hs)-1]) {
// 		//验证不通过
// 		// fmt.Println("验证回滚的区块分叉点，不通过")
// 		return
// 	}
// 	//开始回滚
// 	// fmt.Println("开始回滚区块")
// 	this.rollBackBlocks(n)
// 	//把分叉区块连接的下一个块排序，index为0的是最长链

// 	//回滚后重新加载新的区块，这些区块只统计见证人投票
// 	// fmt.Println("开始加载分叉链区块")
// 	// this.CountForkBlocks(n, hs)

// }

// /*
// 	区块回滚，当链分叉的时候，需要回滚区块，添加最长链的区块
// 	@n    uint64    回滚多少个区块
// */
// func (this *Forks) rollBackBlocks(n uint64) {
// 	block := this.LongChain.GetLastBlock()
// 	for i := uint64(0); i < n; i++ {
// 		this.LongChain.RollbackBlock(block.Height)
// 	}

// }

// /*
// 	统计分叉块
// 	@bh    *BlockHead    最新区块
// 	@n    uint64    回滚多少个区块
// */
// func (this *Forks) CountForkBlocks(n uint64, hs [][]byte) {
// 	block := this.LongChain.GetLastBlock()
// 	for i := uint64(0); i < n; i++ {
// 		block = block.PreBlock[0]
// 	}
// 	for _, hbs := range hs {
// 		has := false
// 		for _, one := range block.NextBlock {
// 			if bytes.Equal(one.Id, hbs) {
// 				//TODO 把本块hash修改排序，排在第一位.
// 				has = true
// 				bh, txs, err := one.LoadTxs()
// 				if err != nil {
// 					// fmt.Println("回滚后重新统计分叉链出错-加载区块信息错误", err)
// 					return
// 				}
// 				this.LongChain.CountBlock(bh, txs)
// 				break
// 			}
// 		}
// 		if !has {
// 			// fmt.Println("程序出错，没找到统计的区块")
// 		}
// 	}
// }

/*
	创建一个新的分叉管理器
*/
//func NewForks() *Forks {
//	return &Forks{
//		//			HeightBlock *Block    //最高区块引用
//		HashMap: new(sync.Map), //区块hash对应的区块引用。key:string=区块hash;value:*Block=区块引用;
//	}
//}

/*
	获取网络节点广播的区块最高高度
*/
func (this *Forks) GetHighestBlock() uint64 {
	return atomic.LoadUint64(&this.HighestBlock)
}

/*
	获取所链接的节点的最高高度
*/
func (this *Forks) SetHighestBlock(n uint64) {
	// engine.Log.Warn("设置最新区块高度 %d", n)
	atomic.StoreUint64(&this.HighestBlock, n)
	db.SaveHighstBlock(n)
}

/*
	获取网络节点广播的区块最高高度
*/
func GetHighestBlock() uint64 {
	return forks.GetHighestBlock()
}

/*
	获取所链接的节点的最高高度
*/
func SetHighestBlock(n uint64) {
	forks.SetHighestBlock(n)
}

// var RemoteCurrentBlock uint64 //远程节点已经同步了的高度
// /*
// 	获取网络节点广播的区块最高高度
// */
// func GetRemoteCurrentBlock() uint64 {
// 	return atomic.LoadUint64(&RemoteCurrentBlock)
// }

// /*
// 	获取所链接的节点的最高高度
// */
// func SetRemoteCurrentBlock(n uint64) {
// 	atomic.StoreUint64(&RemoteCurrentBlock, n)
// }
