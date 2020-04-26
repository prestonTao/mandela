package chain_witness_vote

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/startblock"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"path/filepath"
	// "crypto/ecdsa"
)

// var prk *ecdsa.PrivateKey

// var coinbase *keystore.Address

func Register() error {

	err := utils.StartOtherTime()
	if err != nil {
		return err
	}

	config.ParseInitFlag()

	// mining.ModuleEnable = true
	//检查目录是否存在，不存在则创建
	utils.CheckCreateDir(filepath.Join(config.Wallet_path))

	//启动和链接leveldb数据库
	err = db.InitDB(config.DB_path)
	if err != nil {
		panic(err)
	}

	//删除数据库中的域名

	//检查leveldb数据库中是否有创始区块
	bhvo := mining.LoadStartBlock()
	if bhvo == nil {
		config.DB_is_null = true
	}

	// _, err = db.Find(config.Key_block_start)
	// if err != nil {
	// 	//认为这是一个空数据库
	// 	// engine.Log.Info("这是一个空数据库")

	// 	config.DB_is_null = true
	// }

	// fmt.Println("数据库是否为空", config.DB_is_null)

	// go client.Start()
	// go server.Start()
	mining.RegisteMSG()

	bhvo, err = startblock.BuildFirstBlock()
	if err != nil {
		return err
	}
	if bhvo != nil {
		engine.Log.Info("create initiation block build chain")
		//构建创始区块成功
		mining.BuildFirstChain(bhvo)
		mining.SetHighestBlock(config.Mining_block_start_height)
		mining.GetLongChain().SyncBlockFinish = true
	} else {
		// engine.Log.Info("先加载本地数据库的创始区块")
		//先加载本地数据库的创始区块
		bhvo := mining.LoadStartBlock()
		if bhvo != nil {
			engine.Log.Info("load db initiation block build chain")
			//从本地数据库创始区块构建链
			mining.BuildFirstChain(bhvo)
			mining.SetHighestBlock(db.GetHighstBlock())
			if config.InitNode {
				mining.GetLongChain()
			}
			// engine.Log.Info("从本地数据库创始区块构建链2")
			mining.FindBlockHeight()
			mining.GetLongChain().NoticeLoadBlockForDB(true)
			return nil
		} else {
			engine.Log.Info("neighbor initiation block build chain")
			//从邻居节点同步区块
			mining.GetFirstBlock()
			// engine.Log.Info("用邻居节点区块构建链2")
			mining.FindBlockHeight()
		}
	}
	engine.Log.Info("build chain success")

	//如果是创世节点，不用同步区块，直接开始挖矿
	// if config.InitNode && config.DB_is_null {
	// 	mining.GetLongChain().SetHighestBlock(1)
	// }
	// mining.GetLongChain().GetHighestBlock()
	mining.GetLongChain().NoticeLoadBlockForDB(false)
	return nil

	//当本地数据库为空时，需要先同步第一个区块，这个只有初始3个矿工需要这个操作

	// fmt.Println("检查区块是否合法")
	// //检查区块是否被篡改，中间是否有不连续的块。
	// ok := mining.CheckBlockDB()
	// if !ok {
	// 	fmt.Println("验证区块失败")
	// 	os.Exit(1)
	// }
	// // fmt.Println("区块合法性检查完成")

	// fmt.Println("开始加载数据库中的区块")
	// //加载数据库中的区块
	// err = mining.LoadBlockChain()
	// if err != nil {
	// 	fmt.Println("加载数据库中的区块错误", err)
	// }
	// fmt.Println("完成加载数据库中的区块", config.InitNode)

	// if config.InitNode && !config.DB_is_null {
	// 	mining.SetHighestBlock(mining.GetLongChain().GetLastBlock().Height)
	// }

	// //如果是创世节点，不用同步区块
	// if config.InitNode {
	// 	fmt.Println("开始启动旷工节点")
	// 	// go mining.Mining()
	// 	return nil
	// }

	//一边同步块，一边加载新块
	//开始同步区块
	// err = mining.SyncBlockHead()
	// if err != nil {
	// 	// fmt.Println("同步区块错误", err)
	// 	engine.Log.Info("同步区块错误 %o", err)
	// 	return nil
	// }
	// mining.LoadBlockChain()
	// mining.SyncBlockHead()
	//同步出块时间
	// fmt.Println(forks.GetLongChain())
	// fmt.Println(forks.GetLongChain().witnessChain)
	// // fmt.Println(forks.GetLongChain())
	// forks.GetLongChain().witnessChain.StopAllMining()
	// forks.GetLongChain().witnessChain.BuildMiningTime()
	// mining.NoticeLoadBlockForDB()
	// fmt.Println("启动链端完成")

	// return nil
}
