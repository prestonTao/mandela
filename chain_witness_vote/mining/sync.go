package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

//同步保存区块头队列
var syncSaveBlockHead = make(chan *BlockHeadVO, 1)

//保存区块不连续信号
var syncForNeighborChain = make(chan bool, 1)

//var syncHeightBlock = new(sync.Map)
//var heightBlockGroup = new(sync.WaitGroup)

func (this *Chain) GoSyncBlock() {
	// fmt.Println("--开始循环接收同步链端消息")

	//如果是创始节点，不需要同步其他节点的区块数据
	// if config.InitNode {
	// 	return
	// }
	go func() {
		//创始节点启动需要一次加载
		tag := false

		l := new(sync.RWMutex)
		for force := range syncForNeighborChain {
			//自己是见证人，已经开始出块了，就不需要同步区块
			// engine.Log.Info("打印内存 %b %b", config.InitNode, config.AlreadyMining)
			if config.AlreadyMining && tag {
				continue
			}

			l.Lock()
			// fmt.Println("------------- start")
			this.LoadBlockChain()
			this.SyncBlockHead()
			//同步出块时间
			this.witnessChain.StopAllMining()
			this.witnessChain.BuildMiningTime(force)
			// fmt.Println("------------- end")

			tag = true

			l.Unlock()
		}
	}()
}

/*
	通知加载区块到内存的信号
*/
func (this *Chain) NoticeLoadBlockForDB(force bool) {
	select {
	case syncForNeighborChain <- force:
		engine.Log.Info("Put in sync queue")
	default:
	}
}

/*
	保存一个区块，返回这个区块是否连续，判断依据是能够找到并修改前置区块
	@return    bool    是否找到前置区块
*/
func saveBlockHead(bhvo *BlockHeadVO) (bool, error) {
	// fmt.Println("------------保存区块及交易")

	ok := true
	//保存区块中的交易
	for i, one := range bhvo.Txs {
		//新保存的块，交易输出标记为未使用
		for j, _ := range *one.GetVout() {
			(*bhvo.Txs[i].GetVout())[j].Tx = nil
		}

		// fmt.Println("改变前", hex.EncodeToString(*bhvo.Txs[i].GetHash()))
		bhvo.Txs[i].BuildHash()
		bhvo.Txs[i].SetBlockHash(bhvo.BH.Hash)
		// fmt.Println("改变后", hex.EncodeToString(*bhvo.Txs[i].GetHash()))
		bs, err := bhvo.Txs[i].Json()
		if err != nil {
			//TODO 严谨的错误处理
			fmt.Println("Save block error, transaction JSON format error", err)
			// ok = false
			return false, err
		}
		// fmt.Println("保存交易", hex.EncodeToString(*bhvo.Txs[i].GetHash()), string(*bs))
		err = db.Save(*bhvo.Txs[i].GetHash(), bs)
		if err != nil {
			fmt.Println("Save transaction error", bhvo.Txs[i].GetHash())
			return false, err
		}
		// fmt.Println("保存交易成功", hex.EncodeToString(*bhvo.Txs[i].GetHash()))
	}
	//		fmt.Println("222222222222222")

	//查询前一个区块
	_, err := db.Find(bhvo.BH.Previousblockhash)
	// if err == nil {
	// 	bh, err := ParseBlockHead(bs)
	if err != nil {
		// fmt.Println("查询前置区块错误", err)
		ok = false
	}
	// 	bh.Nextblockhash = bhvo.BH.Hash
	// 	bs, err = bh.Json()
	// 	if err != nil {
	// 		// fmt.Println("格式化前置区块错误", err)
	// 		ok = false
	// 	}
	// 	err = db.Save(bh.Hash, bs)
	// 	if err != nil {
	// 		//TODO 严谨的错误处理
	// 		// fmt.Println("保存前置区块错误", err)
	// 		ok = false
	// 	}
	// } else {
	// 	ok = false
	// }

	//保存区块
	bs, err := bhvo.BH.Json()
	if err != nil {
		//TODO 严谨的错误处理
		// fmt.Println("格式化新区块错误", err)
		// ok = false
		return false, err
	}
	err = db.Save(bhvo.BH.Hash, bs)
	if err != nil {
		//TODO 严谨的错误处理
		// fmt.Println("保存新区块错误", err)
		// ok = false
		return false, err
	}
	return ok, nil
}

/*
	查询邻居节点区块高度
	从邻居节点中查找最高区块高度
*/
func FindBlockHeight() {
	syncHeightBlock := new(sync.Map)

	//	heightBlockGroup = new(sync.WaitGroup)
	//	count := 0

	for _, key := range nodeStore.GetLogicNodes() {
		sessionName := ""
		session, ok := engine.GetSession(key.B58String())
		if ok {
			sessionName = session.GetName()
		}
		message, ok := mc.SendNeighborMsg(config.MSGID_heightBlock, key, nil)
		if ok {
			bs := flood.WaitRequest(mc.CLASS_findHeightBlock, hex.EncodeToString(message.Body.Hash), 0)
			//		fmt.Println("有消息返回了啊")
			if bs == nil {
				// fmt.Println("11111 发送共享文件消息失败，可能超时")
				continue
			}
			chain := forks.GetLongChain()
			//				startHeight := binary.LittleEndian.Uint64((*bs)[:8])
			heightBlock := binary.LittleEndian.Uint64((*bs)[8:])
			//收到的区块高度比自己低，则不保存
			if chain.GetCurrentBlock() > heightBlock {
				continue
			}

			if GetHighestBlock() < heightBlock {
				SetHighestBlock(heightBlock)
			}

			syncHeightBlock.Store(sessionName, heightBlock)
		}

	}
	//以下情况都返回
	//1.没有邻居节点。
	//2.查询邻居节点全部失败。
	//3.邻居节点都未同步完成。
	count := 0
	syncHeightBlock.Range(func(key, value interface{}) bool {
		count++
		return false //只要有数据就够了
	})
	if count <= 0 {
		return
	}

	//统计区块高度结果，给结果投票
	heightBlockVote := new(sync.Map)
	syncHeightBlock.Range(func(key, value interface{}) bool {
		//		fmt.Println("统计投票", key, value)
		height := value.(uint64)
		v, ok := heightBlockVote.Load(height)
		if ok {
			total := v.(uint64)
			heightBlockVote.Store(height, uint64(total+1))
		} else {
			heightBlockVote.Store(height, uint64(1))
		}
		return true
	})

	//查看区块高度投票结果，采用票数高的，票数都一样，采用区块高度最高的。
	heightBlockMax := uint64(0) //区块最高高度
	heightBlock := uint64(0)    //票数最高的区块高度
	heightTotal := uint64(0)    //最高票数
	isEqual := false            //最高票数是否有不同的区块高度
	heightBlockVote.Range(func(k, v interface{}) bool {
		//		fmt.Println("投票结果", k, v)
		height := k.(uint64)
		if height == 0 {
			return true
		}
		if height > heightBlockMax {
			heightBlockMax = height
		}
		total := v.(uint64)
		if total > heightTotal {
			heightTotal = total
			heightBlock = height
		} else if total == heightTotal {
			isEqual = true
		}
		return true
	})
	//TODO 考虑相同票数该选哪个
	//直接使用票数最多的区块高度
	//	atomic.StoreUint64(&chain.StartingBlock, 1)
	SetHighestBlock(heightBlock)
	// atomic.StoreUint64(&forks.GetLongChain().HighestBlock, heightBlock)
	// fmt.Println("收到的区块高度", heightBlock, "自己的高度", atomic.LoadUint64(&forks.CurrentBlock))

}

// var SyncBlockHead_Lock = new(sync.RWMutex)

/*
	从邻居节点同步区块
*/
func (this *Chain) SyncBlockHead() error {
	engine.Log.Info("Start synchronizing blocks from neighbor nodes")
	//获得本节点的最新块hash
	// var bhash *[]byte
	chain := forks.GetLongChain()
	_, block := chain.GetLastBlock()
	bhash := block.Id

	//最新块一定要去邻居节点拉取一次，更新next
	//	bhvo := FindBlockForNeighbor(bhash)
	bhvo := syncBlockFlashDB(&bhash)
	if bhvo == nil {
		return nil
	}

	bhvo.BH.BuildHash()
	// fmt.Println("打印同步到的区块", hex.EncodeToString(bh.Hash))
	engine.Log.Info("Print blocks synchronized to %d %s", bhvo.BH.Height, hex.EncodeToString(bhvo.BH.Hash))

	// this.AddBlock(bhvo.BH, &bhvo.Txs)

	// if bhvo.BH.Nextblockhash == nil {
	// 	return nil
	// }

	tiker := time.NewTicker(time.Minute)

	// for _, one := range bhvo.BH.Nextblockhash {
	// 	this.deepCycleSyncBlock(&one, tiker.C)
	// }
	this.deepCycleSyncBlock(&bhvo.BH.Nextblockhash, tiker.C, bhvo.BH.Height+1)

	tiker.Stop()

	_, block = this.GetLastBlock()
	bhvos := GetUnconfirmedBlockForNeighbor(block.Height)
	for _, one := range bhvos {
		this.AddBlock(one)
	}

	this.SyncBlockFinish = true
	engine.Log.Info("Sync block complete")
	return nil
}

/*
	深度循环同步区块，包括分叉的链的同步
	加载到出错或者加载完成为止
*/
func (this *Chain) deepCycleSyncBlock(bhash *[]byte, c <-chan time.Time, height uint64) {
	// fmt.Println("本次同步hash", hex.EncodeToString(*bhash))
	if bhash == nil || len(*bhash) <= 0 {
		return
	}
	engine.Log.Info("Synchronize block from neighbor node this time %d hash %s", height, hex.EncodeToString(*bhash))
	bh, txItrs, err := this.syncBlockForDBAndNeighbor(bhash)
	if err != nil {
		engine.Log.Info("Error synchronizing block %s", err.Error())
		return
	}

	bhvo := &BlockHeadVO{BH: bh, Txs: txItrs}
	this.AddBlock(bhvo)

	//同步最新高度
	if GetHighestBlock() < bh.Height {
		SetHighestBlock(bh.Height)
	}

	//定时同步区块最新高度
	select {
	case <-c:
		go FindBlockHeight()
	default:
	}

	// fmt.Println("区块的next个数", len(bh.Nextblockhash), "高度", bh.Height)
	// engine.Log.Info("区块的next个数 %d 高度 %d", len(bh.Nextblockhash), bh.Height)
	// for _, one := range bh.Nextblockhash {
	// 	this.deepCycleSyncBlock(&one, c)
	// }
	engine.Log.Info("Next block %s", hex.EncodeToString(bh.Nextblockhash))

	this.deepCycleSyncBlock(&bh.Nextblockhash, c, bh.Height+1)
}

/*
	从数据库查询区块，如果数据库没有，从网络邻居节点查询区块
	查询到区块后，修改他们的指向hash值和UTXO输出的指向
*/
func (this *Chain) syncBlockForDBAndNeighbor(bhash *[]byte) (*BlockHead, []TxItr, error) {
	//此注释代码会导致区块同步中断。本地查到区块，next区块值为null，导致同步中断。
	//先查询数据库
	// head, err := db.Find(*bhash)
	// if err == nil {
	// 	hB, err := ParseBlockHead(head)
	// 	if err == nil {
	// 		fmt.Println("查询到本地数据库中的区块，下一个区块", len(hB.Nextblockhash))
	// 		return hB, nil
	// 	}
	// }
	//再查找邻居节点
	bhvo := FindBlockForNeighbor(bhash)
	if bhvo == nil {
		//同步失败，未找到区块
		return nil, nil, config.ERROR_chain_sysn_block_fail
	}
	//保存区块中的交易
	for i, _ := range bhvo.Txs {
		bhvo.Txs[i].BuildHash()
		bhvo.Txs[i].SetBlockHash(*bhash)
		bs, err := bhvo.Txs[i].Json()
		if err != nil {
			//TODO 严谨的错误处理
			// fmt.Println("严重错误1", err)
			return nil, nil, err
		}
		//			fmt.Println("保存交易", hex.EncodeToString(*bhvo.Txs[i].GetHash()))
		db.Save(*bhvo.Txs[i].GetHash(), bs)

		//将之前的UTXO输出标记为已使用
		// for _, two := range *one.GetVin() {
		// 	//是区块奖励
		// 	if one.Class() == config.Wallet_tx_type_mining {
		// 		continue
		// 	}
		// 	txbs, err := db.Find(two.Txid)
		// 	if err != nil {
		// 		//TODO 区块未同步完整可以查找不到交易
		// 		return nil, nil, err
		// 	}
		// 	txItr, err := ParseTxBase(txbs)
		// 	if err != nil {
		// 		// fmt.Println("严重错误3", err)
		// 		return nil, nil, err
		// 	}
		// 	err = txItr.SetTxid(txbs, two.Vout, one.GetHash())
		// 	if err != nil {
		// 		// fmt.Println("严重错误4", err)
		// 		return nil, nil, err
		// 	}
		// }
	}

	//先将前一个区块修改next
	if this.GetStartingBlock() > config.Mining_block_start_height {
		bs, err := db.Find(bhvo.BH.Previousblockhash)
		if err != nil {
			//TODO 区块未同步完整可以查找不到之前的区块
			return nil, nil, err
		}
		bh, err := ParseBlockHead(bs)
		if err != nil {
			// fmt.Println("严重错误5", err)
			return nil, nil, err
		}

		bh.Nextblockhash = bhvo.BH.Hash

		bs, err = bh.Json()
		if err != nil {
			// fmt.Println("严重错误6", err)
			return nil, nil, err
		}
		db.Save(bh.Hash, bs)
	}

	//保存区块
	bs, err := bhvo.BH.Json()
	if err != nil {
		//TODO 严谨的错误处理
		// fmt.Println("严重错误7", err)
		return nil, nil, err
	}
	db.Save(bhvo.BH.Hash, bs)

	return bhvo.BH, bhvo.Txs, nil
}

/*
	同步区块并刷新本地数据库
*/
func syncBlockFlashDB(bhash *[]byte) *BlockHeadVO {
	bhvo := FindBlockForNeighbor(bhash)
	if bhvo == nil {
		return nil
	}
	bhvo.BH.BuildHash()
	bs, err := bhvo.BH.Json()
	if err != nil {
		return nil
	}
	db.Save(*bhash, bs)
	for i, one := range bhvo.Txs {
		for j, _ := range *one.GetVout() {
			(*bhvo.Txs[i].GetVout())[j].Tx = nil
		}
		one.SetBlockHash(*bhash)
		bs, err := one.Json()
		if err != nil {
			return nil
		}
		db.Save(*one.GetHash(), bs)
	}
	return bhvo
}
