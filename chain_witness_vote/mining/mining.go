package mining

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"bytes"
	"encoding/hex"
	"log"
)

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
	group := this.Group.SelectionChain()
	if group == nil {
		isFirst = true
	} else {
		isFirst = false
		//取本组最后一个块
		// fmt.Println("获取本组最后一个块", len(group.Blocks))
		preBlock = group.Blocks[len(group.Blocks)-1]
		// fmt.Println("1前置区块", preBlock)
	}
	// fmt.Println("是否是本组第一个块", isFirst)

	//找到上一个组
	preGroup := this.Group
	var preGroupBlock *Group
	for {
		// fmt.Println("-----------寻找上一组 1111")
		var ok bool
		preGroup = preGroup.PreGroup
		ok, preGroupBlock = preGroup.CheckBlockGroup()
		if ok {
			// fmt.Println("-----------寻找上一组 222222222")
			if isFirst {
				//取本组最后一个块
				// fmt.Println("获取上一组最后一个块", len(group.Blocks))
				preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
				// fmt.Println("2前置区块", preBlock)
			}
			break
		}
		// fmt.Println("-----------寻找上一组 3333333")
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
	见证人方式出块
	出块并广播
	@gh    uint64    出块的组高度
	@id    []byte    押金id
*/
func (this *Witness) BuildBlock() {
	// var this *Witness
	addr := keystore.GetCoinbase()

	//自己是见证人才能出块，否则自己出块了，其他节点也不会承认
	if !bytes.Equal(*this.Addr, addr) {
		return
	}

	// fmt.Println("===准备出块===")
	engine.Log.Info("=== start building blocks === group height:%d", this.Group.Height)

	// fmt.Println("前置组", preGroup)
	// fmt.Println("前置块", preBlock, preBlock.Height, "\n", hex.EncodeToString(preBlock.Id))

	//查找出未确认的块
	preBlock, blocks := this.FindUnconfirmedBlock()

	//存放交易
	tx := make([]TxItr, 0)
	txids := make([][]byte, 0)

	//检查本组是否给上一组见证人奖励
	if this.WitnessBackupGroup != preBlock.witness.WitnessBackupGroup {
		// engine.Log.Info("开始构建上一组见证人奖励 %s %d", fmt.Sprintf("%p", preBlock.witness.WitnessBackupGroup), preBlock.witness.Group.Height)
		reward := preBlock.witness.WitnessBackupGroup.CountRewardToWitnessGroup(preBlock.Height+1, blocks)
		tx = append(tx, reward)
		txids = append(txids, reward.Hash)
	}

	// //如果是本组首个块，计算上一组旷工奖励
	// if isFirst {
	// 	reward := preGroup.CountReward(preBlock.Height + 1)
	// 	tx = append(tx, reward)
	// 	txids = append(txids, reward.Hash)
	// }

	//打包所有交易
	chain := forks.GetLongChain()
	txs, ids := chain.transactionManager.Package(blocks)
	tx = append(tx, txs...)
	txids = append(txids, ids...)

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
			Witness:           coinbase,            //此块矿工地址
		}
		bh.BuildMerkleRoot()
		bh.BuildSign(coinbase)
		bh.BuildHash()
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

	bhvo := CreateBlockHeadVO(bh, tx)

	engine.Log.Info("=== build block Success === group height:%d block height:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
	engine.Log.Info("=== build block Success === Block hash %s", hex.EncodeToString(bhvo.BH.Hash))
	engine.Log.Info("=== build block Success === pre Block hash %s", hex.EncodeToString(bhvo.BH.Previousblockhash))
	//先保存到数据库再广播，否则其他节点查询不到
	forks.AddBlockHead(bhvo)
	//广播区块
	go MulticastBlock(*bhvo)

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

		ele := NewElection(&nodeStore.NodeSelf.IdInfo.Id)

		message_center.SendMulticastMsg(config.MSGID_multicast_vote_recv, ele.JSON())

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
	bs, err := bhVO.Json()
	if err != nil {
		return
	}
	message_center.SendMulticastMsg(config.MSGID_multicast_blockhead, bs)
}
