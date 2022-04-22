package chain_witness_vote

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/startblock"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"errors"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hyahm/golog"
	// "crypto/ecdsa"
)

// var prk *ecdsa.PrivateKey

// var coinbase *keystore.Address

func Register() error {
	golog.InitLogger("logs/randeHash.txt", 0, true)
	golog.Infof("start %s", "log")

	engine.Log.Info("CPUNUM :%d", config.CPUNUM)
	// if config.CPUNUM < 8 {
	// 	config.CPUNUM = 8
	// }
	go func() {
		for {
			engine.Log.Info("NumGoroutine:%d", runtime.NumGoroutine())
			time.Sleep(time.Minute)
			// log.Error(http.ListenAndServe(":6060", nil))
		}
	}()

	err := utils.StartSystemTime() //StartOtherTime()
	if err != nil {
		return err
	}

	config.ParseInitFlag()

	// mining.ModuleEnable = true
	//检查目录是否存在，不存在则创建
	utils.CheckCreateDir(filepath.Join(config.Wallet_path))

	//启动和链接leveldb数据库
	err = db.InitDB(config.DB_path, config.DB_path_temp)
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

	//创始节点方式启动
	if config.InitNode {
		bhvo, err = startblock.BuildFirstBlock()
		if err != nil {
			return err
		}
		engine.Log.Info("create initiation block build chain")
		config.StartBlockHash = bhvo.BH.Hash
		//构建创始区块成功
		mining.BuildFirstChain(bhvo)
		mining.SetHighestBlock(config.Mining_block_start_height)
		mining.GetLongChain().SyncBlockFinish = true
		mining.GetLongChain().NoticeLoadBlockForDB()
		return nil
	}

	//拉起节点方式启动
	if config.LoadNode {
		// engine.Log.Info("拉起见证人节点")

		bhvo := mining.LoadStartBlock()
		if bhvo == nil {
			return errors.New("加载本地区块数据失败")
		}
		engine.Log.Info("load db initiation block build chain")
		config.StartBlockHash = bhvo.BH.Hash
		//从本地数据库创始区块构建链
		mining.BuildFirstChain(bhvo)
		// config.InitNode = true
		mining.SetHighestBlock(db.GetHighstBlock())

		mining.FindBlockHeight()

		if err := mining.GetLongChain().LoadBlockChain(); err != nil {
			return err
		}
		mining.FinishFirstLoadBlockChain()

		// if err := mining.GetLongChain().FirstDownloadBlock(); err != nil {
		// 	return err
		// }
		return nil
	}
	//普通启动方式
	bhvo = mining.LoadStartBlock()
	if bhvo == nil {
		engine.Log.Info("neighbor initiation block build chain")
		//从邻居节点同步区块
		err = mining.GetFirstBlock()
		if err != nil {
			engine.Log.Error("get first block error: %s", err.Error())
			panic(err.Error())
		}
		// engine.Log.Info("用邻居节点区块构建链2")
		mining.FindBlockHeight()
	} else {
		engine.Log.Info("load db initiation block build chain")
		config.StartBlockHash = bhvo.BH.Hash
		//从本地数据库创始区块构建链
		mining.BuildFirstChain(bhvo)
		mining.FindBlockHeight()
	}
	if err := mining.GetLongChain().FirstDownloadBlock(); err != nil {
		return err
	}

	engine.Log.Info("build chain success")

	mining.GetLongChain().NoticeLoadBlockForDB()
	return nil

	//------------------------------
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

			//拉起见证人节点
			if config.LoadNode {
				engine.Log.Info("拉起见证人节点")
				config.InitNode = true
				// return nil
			}

			mining.SetHighestBlock(db.GetHighstBlock())
			if config.InitNode {
				mining.GetLongChain()
			}
			// engine.Log.Info("从本地数据库创始区块构建链2")
			mining.FindBlockHeight()
			// if err := mining.GetLongChain().FirstDownloadBlock(); err != nil {
			// 	return err
			// }
			// mining.GetLongChain().NoticeLoadBlockForDB()
			// return nil
		} else {
			engine.Log.Info("neighbor initiation block build chain")
			//从邻居节点同步区块
			mining.GetFirstBlock()
			// engine.Log.Info("用邻居节点区块构建链2")
			mining.FindBlockHeight()

		}

		if err := mining.GetLongChain().FirstDownloadBlock(); err != nil {
			return err
		}
	}
	engine.Log.Info("build chain success")

	//如果是创世节点，不用同步区块，直接开始挖矿
	// if config.InitNode && config.DB_is_null {
	// 	mining.GetLongChain().SetHighestBlock(1)
	// }
	// mining.GetLongChain().GetHighestBlock()

	mining.GetLongChain().NoticeLoadBlockForDB()
	return nil

}
