package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
)

/*
	节点启动前，检查数据库中的区块是否合法，区块是否损坏、篡改、连续
	1.先从高度由低到高检查区块头
	2.再从高度由高到低检查交易合法性
	@return    start    起始块高度
	@return    end      数据库中最高块高度
*/
// func CheckBlockDB() bool {

// 	headid, err := db.Find(config.Key_block_start)
// 	if err != nil {
// 		//认为这是一个空数据库
// 		// fmt.Println("这是一个空数据库")
// 		return true
// 	}

// 	bs, err := db.Find(*headid)
// 	if err != nil {
// 		// fmt.Println("1111", err)
// 		return false
// 	}

// 	hB, err := ParseBlockHead(bs)
// 	if err != nil {
// 		// fmt.Println("2222", err)
// 		return false
// 	}

// 	for {
// 		if hB.Nextblockhash == nil {
// 			fmt.Println("没有下一个块了")
// 			break
// 		}
// 		// fmt.Println("开始验证下一个区块", hB.Height+1)
// 		bs, err = db.Find(hB.Nextblockhash)
// 		if err != nil {
// 			//数据库中的区块头查找错误，需要重新下载区块
// 			// fmt.Println("区块头查找错误", hB.Height+1, hex.EncodeToString(*bs))
// 			return false
// 		}
// 		hB, err = ParseBlockHead(bs)
// 		if err != nil {
// 			//数据库中的区块头解析错误，需要重新下载区块
// 			fmt.Println("本区块解析错误", hB.Height)
// 			return false
// 		}
// 		if !hB.CheckBlockHead() {
// 			fmt.Println("本区块不合法", hB.Height)
// 			return false
// 		}
// 		//检查交易是否正确
// 		for _, one := range hB.Tx {
// 			bs, err = db.Find(one)
// 			if err != nil {
// 				//数据库中的交易查找错误，需要重新下载
// 				fmt.Println("查找交易错误", hB.Height, hex.EncodeToString(one))
// 				return false
// 			}
// 			txItr, err := ParseTxBase(bs)
// 			if err != nil {
// 				fmt.Println("解析交易错误", hB.Height, hex.EncodeToString(*bs))
// 				return false
// 			}

// 			if err := txItr.Check(); err != nil {
// 				fmt.Println("验证交易失败，交易不合法")
// 				return false
// 			}
// 		}
// 	}
// 	return true
// }

// var loadBlockChain_Lock = new(sync.RWMutex)

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
	// fmt.Println(bh.Height, hex.EncodeToString(bh.Hash), forks.GetHighestBlock())
	// }
	//	chain.AddBlock(bh, &txItrs)
	// fmt.Println("333333333333")

	if bh.Nextblockhash == nil {
		//		fmt.Println("因Nextblockhash为空退出")
		return nil
	}
	// for i, _ := range bh.Nextblockhash {
	// 	this.deepCycleLoadBlock(&bh.Nextblockhash[i])
	// }
	this.deepCycleLoadBlock(&bh.Nextblockhash)
	return nil
}

/*
	深度循环加载区块，包括分叉的链的加载
	加载到出错或者加载完成为止
*/
func (this *Chain) deepCycleLoadBlock(bhash *[]byte) {
	bh, txItrs, err := loadBlockForDB(bhash)
	if err != nil {
		return
	}
	// engine.Log.Info("--------深度循环加载区块 %d %s", bh.Height, hex.EncodeToString(bh.Hash))
	bhvo := &BlockHeadVO{BH: bh, Txs: txItrs}
	this.AddBlock(bhvo)
	//	chain.AddBlock(bh, &txItrs)
	if bh.Nextblockhash == nil {
		return
	}
	// for i, _ := range bh.Nextblockhash {
	// 	// engine.Log.Info("--深度循环加载区块的下一个区块 %d %s", bh.Height, hex.EncodeToString(bh.Nextblockhash[i]))
	// 	this.deepCycleLoadBlock(&bh.Nextblockhash[i])
	// }
	this.deepCycleLoadBlock(&bh.Nextblockhash)
	return
}

/*
	从数据库中加载一个区块
*/
func loadBlockForDB(bhash *[]byte) (*BlockHead, []TxItr, error) {
	head, err := db.Find(*bhash)
	if err != nil {
		return nil, nil, err
	}
	hB, err := ParseBlockHead(head)
	if err != nil {
		return nil, nil, err
	}
	txItrs := make([]TxItr, 0)
	for _, one := range hB.Tx {
		txBs, err := db.Find(one)
		if err != nil {
			// fmt.Println("3333", err)
			return nil, nil, err
		}
		txItr, err := ParseTxBase(txBs)
		txItrs = append(txItrs, txItr)
	}

	return hB, txItrs, nil
}

/*
	加载数据库中的初始块
*/
func LoadStartBlock() *BlockHeadVO {
	headid, err := db.Find(config.Key_block_start)
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
