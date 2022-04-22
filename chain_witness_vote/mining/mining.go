package mining

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/sqlite3_db"
	"bytes"
	"encoding/hex"
	"log"
	"runtime"
	"sync"
	"time"
)

const (
	Mining_Status_Start           = 1 //开始状态，节点启动
	Mining_Status_WaitMulticas    = 2 //
	Mining_Status_WaitImportBlock = 3 //
	Mining_Status_ImportBlock     = 4 //
	// Mining_Status_WaitMulticas = 5 //

)

var MiningStatusLock = new(sync.Mutex)
var MiningStatus = Mining_Status_Start
var BhvoMulticasCache = make(map[string]*BlockHeadVO)

func init() {

}

func SetMiningStatus_() {

}

func AddBlockToCache(bhvo *BlockHeadVO) {
	//先清理缓存
	MiningStatusLock.Lock()
	now := time.Now()
	for k, bhvo := range BhvoMulticasCache {
		if bhvo.BH.Time < now.Unix()-(config.Mining_block_time*10) {
			delete(BhvoMulticasCache, k)
		}
	}
	//再添加区块
	BhvoMulticasCache[utils.Bytes2string(bhvo.BH.Hash)] = bhvo
	MiningStatusLock.Unlock()
}

/*
	从缓存中找到对应区块导入
*/
func ImportBlockByCache(hash *[]byte) {
	var bhvo, prebhvo *BlockHeadVO
	// ok := false
	MiningStatusLock.Lock()
	bhvo, _ = BhvoMulticasCache[utils.Bytes2string(*hash)]
	MiningStatusLock.Unlock()
	if bhvo == nil {
		return
	}
	//导入区块，先查找本区块是否也在缓存中

	MiningStatusLock.Lock()
	prebhvo, _ = BhvoMulticasCache[utils.Bytes2string(bhvo.BH.Previousblockhash)]
	MiningStatusLock.Unlock()
	if prebhvo != nil {
		//先导入前置区块
		err := forks.AddBlockHead(prebhvo)
		if err == nil {
			//广播区块
			go MulticastBlock(*prebhvo)
			// utils.Go(MulticastBlock(*prebhvo))
			MiningStatusLock.Lock()
			delete(BhvoMulticasCache, utils.Bytes2string(prebhvo.BH.Hash))
			MiningStatusLock.Unlock()
		}
	}
	//再导入本区块
	err := forks.AddBlockHead(bhvo)
	if err == nil {
		//广播区块
		go MulticastBlock(*bhvo)
		// utils.Go(MulticastBlock(*bhvo))
		MiningStatusLock.Lock()
		delete(BhvoMulticasCache, utils.Bytes2string(bhvo.BH.Hash))
		MiningStatusLock.Unlock()
	}
}

// /*
// 	开始挖矿
// 	当每个组见证人选出来之后，启动挖矿程序，按顺序定时出块
// */
// func Mining() {
// 	//判断是否同步完成
// 	if GetHighestBlock() <= 0 {
// 		fmt.Println("区块未同步完成，不能挖矿 GetHighestBlock", GetHighestBlock())
// 		return
// 	}
// 	if GetHighestBlock() > GetCurrentBlock() {
// 		fmt.Println("区块未同步完成，不能挖矿 GetCurrentBlock", GetCurrentBlock(), GetHighestBlock())
// 		return
// 	}
// 	if !config.Miner {
// 		fmt.Println("本节点不是旷工节点")
// 		return
// 	}

// 	fmt.Println("启动挖矿程序")

// 	addr := keystore.GetCoinbase()

// 	//用见证人方式出块
// 	fmt.Println("用见证人方式出块")
// 	group := forks.GetLongChain().witnessChain.group
// 	//判断是否已经安排了任务
// 	if group.Task {
// 		// fmt.Println("已经安排了任务，退出")
// 		return
// 	}
// 	group.Task = true

// 	//判断自己出块顺序的时间
// 	for i, one := range forks.GetLongChain().witnessChain.group.Witness {
// 		//自己是见证人才能出块，否则自己出块了，其他节点也不会承认
// 		if bytes.Equal(*one.Addr, addr) {
// 			fmt.Println("自己多少秒钟后出块", config.Mining_block_time*(i+1))
// 			utils.AddTimetask(utils.GetNow()+int64(config.Mining_block_time*(i+1)),
// 				TaskBuildBlock, Task_class_buildBlock, "")
// 		}

// 	}

// }

// var logblockheight = uint64(152020)

/*
	查找未确认的区块
	获取其中的交易，用于验证交易
	@return    *Block     出块时，应该链接的上一个块
	@return    []Block    出块时，应该链接的上一个组的块
*/
func (this *Witness) FindUnconfirmedBlock() (*Block, []Block) {
	//找到上一个块
	var preBlock *Block
	//判断是否是该组第一个块
	isFirst := false
	// engine.Log.Info("FindUnconfirmedBlock SelectionChain")
	group := this.Group.SelectionChain(nil)
	if group == nil {
		isFirst = true
	} else {
		isFirst = false
		//取本组最后一个块
		// fmt.Println("获取本组最后一个块", len(group.Blocks))
		preBlock = group.Blocks[len(group.Blocks)-1]
		// engine.Log.Info("1前置区块 %+v", preBlock)
	}
	// engine.Log.Info("是否是本组第一个块 %v", isFirst)

	//找到上一个组
	preGroup := this.Group
	var preGroupBlock *Group
	var ok bool
	for {
		// if preGroup.Height > logblockheight || preGroup.Height < 145137 {
		// 	engine.Log.Info("-----------寻找上一组 1111 %d", preGroup.Height)
		// }
		ok = false
		preGroup = preGroup.PreGroup
		ok, preGroupBlock = preGroup.CheckBlockGroup(nil)
		if ok {
			// engine.Log.Info("-----------寻找上一组 222222222")
			if isFirst {
				//取本组最后一个块
				// engine.Log.Info("获取上一组最后一个块")
				preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
				// engine.Log.Info("2前置区块", preBlock)
			}
			break
		}
		// engine.Log.Info("-----------寻找上一组 3333333")
	}

	//查找出未确认的块
	blocks := make([]Block, 0)
	if preGroup.Height != this.Group.Height {
		for _, one := range preGroupBlock.Blocks {
			blocks = append(blocks, *one)
		}
	}
	if group != nil {
		for _, one := range group.Blocks {
			blocks = append(blocks, *one)
		}
	}
	return preBlock, blocks
}

/*
	验证未确认的区块
	获取其中的交易，用于验证交易
	@return    *Block     出块时，应该链接的上一个块
	@return    []Block    出块时，应该链接的上一个组的块
*/
func (this *Witness) CheckUnconfirmedBlock(blockhash *[]byte) (*Block, []Block) {

	//找到上一个块
	var preBlock *Block
	//判断是否是该组第一个块
	isFirst := false

	// engine.Log.Info("CheckUnconfirmedBlock SelectionChain")
	group := this.Group.SelectionChain(blockhash)

	if group == nil {
		isFirst = true
	} else {
		isFirst = false
		//取本组最后一个块
		// fmt.Println("获取本组最后一个块", len(group.Blocks))
		preBlock = group.Blocks[len(group.Blocks)-1]
		// engine.Log.Info("1前置区块 %+v", preBlock)
	}
	// engine.Log.Info("是否是本组第一个块 %v", isFirst)

	//找到上一个组
	preGroup := this.Group
	var preGroupBlock *Group
	var ok bool
	for {
		// if preGroup.Height > logblockheight || preGroup.Height < 145137 {
		// 	engine.Log.Info("-----------寻找上一组 1111 %d", preGroup.Height)
		// }
		ok = false
		preGroup = preGroup.PreGroup

		ok, preGroupBlock = preGroup.CheckBlockGroup(blockhash)

		if ok {
			// engine.Log.Info("-----------寻找上一组 222222222")
			if isFirst {
				//取本组最后一个块
				// engine.Log.Info("获取上一组最后一个块")
				preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
				// engine.Log.Info("2前置区块", preBlock)
			}
			break
		}
		// engine.Log.Info("-----------寻找上一组 3333333")
	}

	// for {
	// 	if bytes.Equal(preBlock.Id, bh.Previousblockhash) {
	// 		break
	// 	} else {
	// 		if preBlock.Height <= bh.Height-1 {
	// 			//找不到这个区块了，因为高度都不一样了
	// 			break
	// 		} else {
	// 			preBlock = preBlock.PreBlock
	// 		}
	// 	}
	// }

	preWitness := this.PreWitness
	for {
		if preWitness == nil {
			break
		}
		if preWitness.Block == nil {
			preWitness = preWitness.PreWitness
			continue
		}

		if bytes.Equal(preWitness.Block.Id, *blockhash) {

			preBlock = preWitness.Block
			break
		}
		preWitness = preWitness.PreWitness
		// engine.Log.Info("-----------寻找上一个块 444444444")
	}

	//查找出未确认的块
	blocks := make([]Block, 0)
	if preGroup.Height != this.Group.Height {
		for _, one := range preGroupBlock.Blocks {
			blocks = append(blocks, *one)
		}
	}
	if group != nil {
		for _, one := range group.Blocks {
			blocks = append(blocks, *one)
		}
	}
	return preBlock, blocks
}

// var testBug = false

/*
	见证人方式出块
	出块并广播
	@gh    uint64    出块的组高度
	@id    []byte    押金id
*/
func (this *Witness) BuildBlock() {
	// var this *Witness
	addrInfo := keystore.GetCoinbase()

	//自己是见证人才能出块，否则自己出块了，其他节点也不会承认
	if !bytes.Equal(*this.Addr, addrInfo.Addr) {
		return
	}

	//本节点未同步完成，则不应该出块
	if !GetLongChain().SyncBlockFinish {
		return
	}

	// start := time.Now()
	// fmt.Println("===准备出块===")
	engine.Log.Info("=== start building blocks === group height:%d", this.Group.Height)

	// fmt.Println("前置组", preGroup)
	// fmt.Println("前置块", preBlock, preBlock.Height, "\n", hex.EncodeToString(preBlock.Id))

	//查找出未确认的块
	preBlock, blocks := this.FindUnconfirmedBlock()
	// engine.Log.Info("上一个块高度 %d %d %d", preBlock.Height, blocks[0].Height, blocks[0].witness.Group.Height)
	// engine.Log.Info("获取上一个块高度消耗时间 %s", time.Now().Sub(start))

	//存放交易
	tx := make([]TxItr, 0)
	txids := make([][]byte, 0)

	var reward *Tx_reward
	//检查本组是否给上一组见证人奖励
	if this.WitnessBackupGroup != preBlock.witness.WitnessBackupGroup {
		// engine.Log.Info("开始构建上一组见证人奖励 %s %d", fmt.Sprintf("%p", preBlock.witness.WitnessBackupGroup), preBlock.witness.Group.Height)
		reward = preBlock.witness.WitnessBackupGroup.CountRewardToWitnessGroup(preBlock.Height+1, blocks, preBlock)
		tx = append(tx, reward)
		txids = append(txids, reward.Hash)
	}

	// engine.Log.Info("构建区块奖励消耗时间 %s", time.Now().Sub(start))

	//打包所有交易
	chain := forks.GetLongChain()

	txs, ids := chain.transactionManager.Package(reward, preBlock.Height+1, blocks, this.CreateBlockTime)
	tx = append(tx, txs...)
	txids = append(txids, ids...)

	// engine.Log.Info("打包消耗时间 %s", time.Now().Sub(start))

	//准备块中的交易
	// fmt.Println("准备块中的交易")
	coinbase := keystore.GetCoinbase()

	var bh *BlockHead
	now := utils.GetNow() //time.Now().Unix()
	for i := int64(0); i < (config.Mining_block_time*2)-1; i++ {
		//开始生成块
		bh = &BlockHead{
			Height:            preBlock.Height + 1, //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       this.Group.Height,   // preGroup.Height + 1,               //矿工组高度
			Previousblockhash: preBlock.Id,         //上一个区块头hash
			NTx:               uint64(len(tx)),     //交易数量
			Tx:                txids,               //本区块包含的交易id
			Time:              now + i,             //unix时间戳
			Witness:           coinbase.Addr,       //此块矿工地址
		}

		// if !testBug && bh.Height > config.Mining_block_start_height+12 && this.Group.FirstWitness() {
		// 	engine.Log.Error("把区块高度减1")
		// 	testBug = true
		// 	bh.Height = bh.Height - 1
		// }

		bh.BuildMerkleRoot()
		bh.BuildSign(coinbase.Addr)
		bh.BuildBlockHash()
		if bh.CheckHashExist() {
			bh = nil
			continue
		} else {
			break
		}
	}
	if bh == nil {
		engine.Log.Info("Block out failed, all hash have collisions")
		//出块失败，所有hash都有碰撞
		return
	}

	bhvo := CreateBlockHeadVO(config.StartBlockHash, bh, tx)

	engine.Log.Info("=== build block Success === group height:%d block height:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	engine.Log.Info("=== build block Success === Block hash %s", hex.EncodeToString(bhvo.BH.Hash))
	engine.Log.Info("=== build block Success === pre Block hash %s", hex.EncodeToString(bhvo.BH.Previousblockhash))
	//先保存到数据库再广播，否则其他节点查询不到
	// SaveBlockHead(bhvo)

	bhvo.FromBroadcast = true

	// <<<<<<< HEAD
	// 	err := MulticastBlockAndImport(bhvo)
	// 	if err != nil {
	// 		return
	// 	}

	// 	forks.AddBlockHead(bhvo)

	// 	//广播区块
	// 	go MulticastBlock(*bhvo)
	// =======
	UniformityMulticastBlock(bhvo)

	// bhvo.FromBroadcast = true

	// err := MulticastBlockAndImport(bhvo)
	// if err != nil {
	// 	return
	// }

	// forks.AddBlockHead(bhvo)

	// //广播区块
	// go MulticastBlock(*bhvo)
	// >>>>>>> dev
}

// /*
// 	POW方式出块并广播
// 	@gh    uint64    出块的组高度
// 	@id    []byte    押金id
// */
// func BuildBlockForPOW() {
// 	addr := keystore.GetCoinbase()

// 	fmt.Println("===准备pow方式出块===")

// 	chain := forks.GetLongChain()
// 	lastBlock := chain.GetLastBlock()

// 	//打包交易
// 	txs := make([]TxItr, 0)
// 	txids := make([][]byte, 0)

// 	//打包10秒内的所有交易
// 	txss, ids := chain.transactionManager.Package()
// 	fmt.Println("打包的交易", len(txss))

// 	allGas := uint64(0)
// 	for _, one := range txss {
// 		allGas = allGas + one.GetGas()
// 	}

// 	//第一个块产出80个币
// 	//每增加一定块数，产出减半，直到为0
// 	//最多减半9次，第10次减半后产出为0
// 	oneReward := uint64(config.Mining_reward)
// 	n := (lastBlock.Height + 1) / config.Mining_block_cycle
// 	if n < 10 {
// 		for i := uint64(0); i < n; i++ {
// 			oneReward = oneReward / 2
// 		}
// 	} else {
// 		oneReward = 0
// 	}
// 	allReward := oneReward + allGas

// 	//构造出块奖励
// 	if allReward > 0 {
// 		vouts := make([]Vout, 0)
// 		vouts = append(vouts, Vout{
// 			Value:   allReward, //输出金额 = 实际金额 * 100000000
// 			Address: addr,      //钱包地址
// 		})
// 		base := TxBase{
// 			Type:       config.Wallet_tx_type_mining, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
// 			Vout_total: uint64(len(vouts)),           //输出交易数量
// 			Vout:       vouts,                        //交易输出
// 			LockHeight: lastBlock.Height + 100,       //锁定高度
// 			//			CreateTime: time.Now().Unix(),            //创建时间
// 		}
// 		reward := Tx_reward{
// 			TxBase: base,
// 		}
// 		txs = append(txs, &reward)
// 		reward.BuildHash()
// 		txids = append(txids, reward.Hash)
// 	}

// 	//判断上一个组是否是见证人方式出块，是见证人方式出块，计算上一组出块奖励。
// 	if chain.witnessChain.beforeGroup != nil {
// 		reward := chain.witnessChain.beforeGroup.CountReward(txss)
// 		txs = append(txs, reward)
// 		txids = append(txids, reward.Hash)
// 	}

// 	txs = append(txs, txss...)
// 	txids = append(txids, ids...)

// 	//准备块中的交易
// 	fmt.Println("准备块中的交易")
// 	coinbase := keystore.GetCoinbase()

// 	//开始生成块
// 	bh := BlockHead{
// 		Height:            lastBlock.Height + 1,       //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
// 		GroupHeight:       lastBlock.Group.Height + 1, //矿工组高度
// 		Previousblockhash: [][]byte{lastBlock.Id},     //上一个区块头hash
// 		NTx:               uint64(len(txs)),           //交易数量
// 		Tx:                txids,                      //本区块包含的交易id
// 		Time:              time.Now().Unix(),          //unix时间戳
// 		//		BackupMiner:       bmId,                            //备用矿工选举结果hash
// 		//		DepositId: this.DepositId, //预备矿工组高度
// 		Witness: coinbase, //此块矿工地址
// 	}
// 	bh.BuildMerkleRoot()

// 	if !findNonce(&bh) {
// 		fmt.Println("因中断而退出")
// 		return
// 	}

// 	bhvo := CreateBlockHeadVO(&bh, txs)
// 	fmt.Println("========POW 出块完成 高度为", bhvo.BH.Height, base64.StdEncoding.EncodeToString(bhvo.BH.Hash))

// 	//广播区块
// 	MulticastBlock(bhvo)

// 	AddBlockHead(bhvo)
// }

/*
	发起投票，广播
*/
func Seekvote() {
	//	log.Println("发起投票")
	//	engine.NLog.Debug(engine.LOG_console, "发起投票")
	if nodeStore.NodeSelf.IsSuper {
		//		engine.NLog.Debug(engine.LOG_console, "是超级节点发起投票")
		log.Println("是超级节点发起投票")

		//		coinbase := "1234567890"

		//		ele := NewElection(coinbase)
		//		content := ele.JSON()
		//		if content == nil {
		//			return
		//		}

		//添加自己为竞选
		//		AddElection(ele)

		// ele := NewElection(&nodeStore.NodeSelf.IdInfo.Id)

		// message_center.SendMulticastMsg(config.MSGID_multicast_vote_recv, ele.JSON())

		// //		content := []byte(*nodeStore.NodeSelf.IdInfo.Id)
		// head := mc.NewMessageHead(nil, nil, false)
		// body := mc.NewMessageBody(ele.JSON(), "", nil, 0)
		// message := mc.NewMessage(head, body)
		// message.BuildHash()

		// //广播给其他节点
		// //		ids := nodeStore.GetIdsForFar(message.Content)
		// for _, one := range nodeStore.GetAllNodes() {
		// 	log.Println("发送给", one.B58String())
		// 	if ss, ok := engine.GetSession(one.B58String()); ok {
		// 		ss.Send(config.MSGID_multicast_vote_recv, head.JSON(), body.JSON(), false)
		// 	} else {
		// 		engine.NLog.Debug(engine.LOG_console, "发送消息失败")
		// 	}
		// }
	} else {
		//非超级节点不需要广播
	}
}

/*
	广播挖到的区块
*/
func MulticastBlock(bhVO BlockHeadVO) {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	bs, err := bhVO.Proto() //bhVO.Json()
	if err != nil {
		return
	}
	message_center.SendMulticastMsg(config.MSGID_multicast_blockhead, bs)
}

/*
	广播挖到的区块，当各个见证人都收到后，导入区块
*/
func MulticastBlockAndImport(bhVO *BlockHeadVO) error {
	// goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	// _, file, line, _ := runtime.Caller(0)
	// engine.AddRuntime(file, line, goroutineId)
	// defer engine.DelRuntime(file, line, goroutineId)
	bs, err := bhVO.Proto() //bhVO.Json()
	if err != nil {
		return err
	}

	head := message_center.NewMessageHead(nil, nil, false)
	body := message_center.NewMessageBody(config.MSGID_multicast_witness_blockhead, bs, 0, nil, 0)
	message := message_center.NewMessage(head, body)
	message.BuildHash()
	//先保存这个消息到缓存
	err = new(sqlite3_db.MessageCache).Add(message.Body.Hash, head.Proto(), body.Proto())
	if err != nil {
		engine.Log.Error(err.Error())
		return err
	}
	engine.Log.Info("multicast message hash:%s", hex.EncodeToString(message.Body.Hash))
	return MulticastBlockSync(message)
}

/*
	广播挖到的区块，当各个见证人都收到后，导入区块
*/
func MulticastBlockSync(message *message_center.Message) error {
	whiltlistNodes := nodeStore.GetWhiltListNodes()
	return message_center.BroadcastsAll(1, config.MSGID_multicast_witness_blockhead, whiltlistNodes, nil, nil, &message.Body.Hash)
}

func UniformityMulticastBlock(bhVO *BlockHeadVO) {

	bs, err := bhVO.Proto() //bhVO.Json()
	if err != nil {
		return
	}

	head := message_center.NewMessageHead(nil, nil, false)
	body := message_center.NewMessageBody(config.MSGID_multicast_witness_blockhead, bs, 0, nil, 0)
	message := message_center.NewMessage(head, body)
	message.BuildHash()
	//先保存这个消息到缓存
	// engine.Log.Info("保存消息到缓存:%s", hex.EncodeToString(message.Body.Hash))
	err = new(sqlite3_db.MessageCache).Add(message.Body.Hash, head.Proto(), body.Proto())
	if err != nil {
		engine.Log.Error("save message cache error:%s", err.Error())
		return
	}

	AddBlockToCache(bhVO)
	//先广播区块并确认到达
	// engine.Log.Info("广播区块hash:%s", hex.EncodeToString(bhVO.BH.Hash))
	err = UniformityBroadcasts(&message.Body.Hash, config.MSGID_uniformity_multicast_witness_blockhead, config.CLASS_uniformity_witness_multicas_blockhead, config.Wallet_sync_block_timeout)
	if err != nil {
		engine.Log.Info("multcas blockhash error:%s", err.Error())
		return
	}
	//再发送导入命令
	err = UniformityBroadcasts(&bhVO.BH.Hash, config.MSGID_uniformity_multicast_witness_block_import, config.CLASS_uniformity_witness_multicas_block_import, config.Wallet_sync_block_timeout)
	if err != nil {
		engine.Log.Info("multcas import block error:%s", err.Error())
		return
	}
	engine.Log.Info("start import block")
	ImportBlockByCache(&bhVO.BH.Hash)

}

/*
	发送广播
*/
func UniformityBroadcasts(hash *[]byte, msgid uint64, waitRequestClass string, timeout int64) error {
	// timeout := 4
	// timeoutloopTotal := config.Wallet_multicas_block_time / timeout
	whiltlistNodes := nodeStore.GetWhiltListNodes()
	//给已发送的节点放map里，避免重复发送
	allNodes := make(map[string]bool)

	var timeouterrorlock = new(sync.Mutex)
	var timeouterror error

	//先发送给超级节点
	// superNodes := nodeStore.GetLogicNodes()
	//排除重复的地址
	// superNodes = nodeStore.RemoveDuplicateAddress(superNodes)
	cs := make(chan bool, config.CPUNUM)
	group := new(sync.WaitGroup)
	for i, _ := range whiltlistNodes {
		sessionid := whiltlistNodes[i]
		//不要发送给自己
		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, sessionid) {
			continue
		}
		_, ok := allNodes[utils.Bytes2string(sessionid)]
		if ok {
			// engine.Log.Info("repeat node addr: %s", sessionid.B58String())
			continue
		}
		allNodes[utils.Bytes2string(sessionid)] = false
		cs <- false
		group.Add(1)
		//区块广播给节点
		// engine.Log.Info("multcast super node:%s", sessionid.B58String())
		utils.Go(func() {
			success := false
			_, err := message_center.SendNeighborWithReplyMsg(msgid, &sessionid, hash, waitRequestClass, timeout)
			if err == nil {
				success = true
			} else {
				if err.Error() == config.ERROR_wait_msg_timeout.Error() {
				} else {
					//其他错误不管，当作发送成功
					success = true
				}
			}
			if !success {
				timeouterrorlock.Lock()
				timeouterror = config.ERROR_wait_msg_timeout
				timeouterrorlock.Lock()
			}
			<-cs
			group.Done()
		})
	}
	group.Wait()
	// engine.Log.Info("multicast whilt list node time %s", time.Now().Sub(start))

	// engine.Log.Info("multicast proxy node time %s", time.Now().Sub(start))
	return timeouterror
}
