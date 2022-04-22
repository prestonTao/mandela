package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"bytes"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

/*
	从数据库中加载区块
	先找到内存中最高区块，从区块由低到高开始加载
*/
func (this *Chain) LoadBlockChain() error {

	engine.Log.Info("Start loading blocks in database")
	forks.SetHighestBlock(db.GetHighstBlock())

	_, lastBlock := this.GetLastBlock()
	headid := lastBlock.Id
	bh, _, err := loadBlockForDB(&headid)
	if err != nil {
		return err
	}

	if bh.Nextblockhash == nil {
		return nil
	}
	blockhash := bh.Nextblockhash
	var bhvo *BlockHeadVO
	for blockhash != nil && len(blockhash) > 0 {
		//这里查看内存，控制速度
		memInfo, _ := mem.VirtualMemory()
		if memInfo.UsedPercent > config.Wallet_Memory_percentage_max {
			runtime.GC()
			time.Sleep(time.Second)
		}

		bhvo = this.deepCycleLoadBlock(&blockhash)
		if bhvo == nil {
			break
		}
		//临时改变nexthash
		// if nextHash, ok := config.BlockNextHash.Load(utils.Bytes2string(bhvo.BH.Hash)); ok {
		// 	nextHashBs := nextHash.(*[]byte)
		// 	blockhash = *nextHashBs
		// 	continue
		// }
		// if bhvo.BH.Height == 817445 {
		// 	bhvo.BH.Nextblockhash = config.SpecialBlockHash
		// }

		if bhvo.BH.Nextblockhash == nil || len(bhvo.BH.Nextblockhash) <= 0 {
			break
		}
		blockhash = bhvo.BH.Nextblockhash
	}
	engine.Log.Info("end loading blocks in database")
	return nil
}

/*
	深度循环加载区块，包括分叉的链的加载
	加载到出错或者加载完成为止
*/
func (this *Chain) deepCycleLoadBlock(bhash *[]byte) *BlockHeadVO {
	if len(config.BlockHashs) > 0 && bytes.Equal(*bhash, config.BlockHashs[0]) {
		bhash = this.testLoadBlock()
		// return
	}

	bh, txItrs, err := loadBlockForDB(bhash)
	if err != nil {
		engine.Log.Info("load block for db error:%s", err.Error())
		return nil
	}
	// engine.Log.Info("--------深度循环加载区块 %d %s", bh.Height, hex.EncodeToString(bh.Hash))
	bhvo := &BlockHeadVO{FromBroadcast: false, BH: bh, Txs: txItrs}
	err = this.AddBlock(bhvo)
	if err != nil {
		engine.Log.Info("add block error:%s", err.Error())
		return nil
	}
	//	chain.AddBlock(bh, &txItrs)
	if bh.Nextblockhash == nil {
		engine.Log.Info("load block next blockhash nil")
		return nil
	}
	// for i, _ := range bh.Nextblockhash {
	// 	// engine.Log.Info("--深度循环加载区块的下一个区块 %d %s", bh.Height, hex.EncodeToString(bh.Nextblockhash[i]))
	// 	this.deepCycleLoadBlock(&bh.Nextblockhash[i])
	// }
	// this.deepCycleLoadBlock(&bh.Nextblockhash)
	return bhvo
}

/*
	深度循环加载区块，包括分叉的链的加载
	加载到出错或者加载完成为止
*/
func (this *Chain) testLoadBlock() *[]byte {
	engine.Log.Info("testLoadBlock---------------------------")
	var nextBlockhash *[]byte
	for _, one := range config.BlockHashs {
		bh, txItrs, err := loadBlockForDB(&one)
		if err != nil {
			continue
		}
		// engine.Log.Info("--------深度循环加载区块 %d %s", bh.Height, hex.EncodeToString(bh.Hash))
		bhvo := &BlockHeadVO{FromBroadcast: false, BH: bh, Txs: txItrs}
		err = this.AddBlock(bhvo)
		if err != nil {
			continue
		}
		//	chain.AddBlock(bh, &txItrs)
		if bh.Nextblockhash == nil {
			continue
		}
		nextBlockhash = &bh.Nextblockhash
	}
	return nextBlockhash
}

/*
	从数据库中加载一个区块
*/
// func loadBlockForDB(bhash *[]byte) (*BlockHead, []TxItr, error) {
// 	head, err := db.Find(*bhash)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	hB, err := ParseBlockHead(head)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	txItrs := make([]TxItr, 0)
// 	for _, one := range hB.Tx {
// 		// txItr, err := FindTxBase(one, hex.EncodeToString(one))
// 		txItr, err := FindTxBase(one)

// 		// txBs, err := db.Find(one)
// 		if err != nil {
// 			// fmt.Println("3333", err)
// 			return nil, nil, err
// 		}
// 		// txItr, err := ParseTxBase(ParseTxClass(one), txBs)
// 		txItrs = append(txItrs, txItr)
// 	}

// 	return hB, txItrs, nil
// }

/*
	加载数据库中的初始块
*/
func LoadStartBlock() *BlockHeadVO {
	headid, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		//认为这是一个空数据库
		engine.Log.Info("This is an empty database")
		return nil
	}
	bh, txItrs, err := loadBlockForDB(headid)
	if err != nil {
		return nil
	}
	bhvo := BlockHeadVO{
		BH:  bh,     //区块
		Txs: txItrs, //交易明细
	}
	return &bhvo
}
