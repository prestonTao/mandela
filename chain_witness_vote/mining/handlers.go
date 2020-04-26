package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"time"
)

func RegisteMSG() {

	// fmt.Println("---------- 注册区块链模块消息 ----------")
	// engine.RegisterMsg(config.MSGID_multicast_vote_recv, MulticastVote_recv) //接收投票旷工广播
	message_center.Register_multicast(config.MSGID_multicast_vote_recv, MulticastVote_recv)           //接收投票旷工广播
	message_center.Register_multicast(config.MSGID_multicast_blockhead, MulticastBlockHead_recv)      //接收区块头广播
	message_center.Register_neighbor(config.MSGID_heightBlock, FindHeightBlock)                       //查询邻居节点区块高度
	message_center.Register_neighbor(config.MSGID_heightBlock_recv, FindHeightBlock_recv)             //查询邻居节点区块高度_返回
	message_center.Register_neighbor(config.MSGID_getBlockHead, GetBlockHead)                         //查询起始区块头
	message_center.Register_neighbor(config.MSGID_getBlockHead_recv, GetBlockHead_recv)               //查询起始区块头_返回
	message_center.Register_neighbor(config.MSGID_getTransaction, GetTransaction)                     //查询交易
	message_center.Register_neighbor(config.MSGID_getTransaction_recv, GetTransaction_recv)           //查询交易_返回
	message_center.Register_multicast(config.MSGID_multicast_transaction, MulticastTransaction_recv)  //接收交易广播
	message_center.Register_neighbor(config.MSGID_getUnconfirmedBlock, GetUnconfirmedBlock)           //从邻居节点获取未确认的区块
	message_center.Register_neighbor(config.MSGID_getUnconfirmedBlock_recv, GetUnconfirmedBlock_recv) //从邻居节点获取未确认的区块_返回

	message_center.Register_neighbor(config.MSGID_multicast_return, MulticastReturn_recv) //收到广播消息回复_返回

	message_center.Register_neighbor(config.MSGID_getblockforwitness, GetBlockForWitness)           //从邻居节点获取指定见证人的区块
	message_center.Register_neighbor(config.MSGID_getblockforwitness_recv, GetBlockForWitness_recv) //从邻居节点获取指定见证人的区块_返回

	flowControllerWaiteSecount()

}

type BlockForWitness struct {
	GroupHeight uint64             //见证人组高度
	Addr        crypto.AddressCoin //见证人地址
}

/*
	从邻居节点获取指定见证人的区块
*/
func GetBlockForWitness(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	engine.Log.Info("recv GetBlockForWitness message")
	bs := []byte{}
	if message.Body.Content == nil {
		engine.Log.Warn("GetBlockForWitness message.Body.Content is nil")
		message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	bfw := new(BlockForWitness)
	decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	decoder.UseNumber()
	err := decoder.Decode(bfw)
	if err != nil {
		engine.Log.Warn("GetBlockForWitness decoder error: %s", err.Error())
		message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	witnessGroup := GetLongChain().witnessChain.witnessGroup
	for witnessGroup.Height > bfw.GroupHeight && witnessGroup.PreGroup != nil {
		witnessGroup = witnessGroup.PreGroup
	}
	for witnessGroup.Height < bfw.GroupHeight && witnessGroup.NextGroup != nil {
		witnessGroup = witnessGroup.NextGroup
	}
	if witnessGroup.Height != bfw.GroupHeight {
		engine.Log.Warn("GetBlockForWitness not find group height")
		message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	//找到了组高度，遍历这个组中的见证人
	for _, one := range witnessGroup.Witness {
		if !bytes.Equal(*one.Addr, bfw.Addr) {
			continue
		}
		//找到了这个见证人
		//这个见证人还未出块
		if one.Block == nil {
			message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		} else {
			bh, tx, err := one.Block.LoadTxs()
			if err != nil {
				message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
			}
			bhvo := CreateBlockHeadVO(bh, *tx)
			newbs, err := bhvo.Json()
			if err != nil {
				engine.Log.Warn("GetBlockForWitness bhvo encoding json error: %s", err.Error())
			}
			bs = *newbs
			engine.Log.Info("GetBlockForWitness bhvo encoding Success")
			message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		}
		break
	}
	engine.Log.Warn("GetBlockForWitness not find this witness")
	message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)

}

/*
	从邻居节点获取指定见证人的区块_返回
*/
func GetBlockForWitness_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// bs := []byte("ok")
	flood.ResponseWait(config.CLASS_wallet_getblockforwitness, hex.EncodeToString(message.Body.Hash), message.Body.Content)
}

/*
	收到广播消息回复_返回
*/
func MulticastReturn_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	bs := []byte("ok")
	flood.ResponseWait(config.CLASS_wallet_broadcast_return, hex.EncodeToString(message.Body.Hash), &bs)
}

/*
	接收备用见证人投票广播
*/
func MulticastVote_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收见证人投票广播")
	//	engine.NLog.Debug(engine.LOG_console, "接收投票旷工广播")
	//	log.Println("接收投票旷工广播", msg.Session.GetName())

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// //自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// if !message.CheckSendhash() {
	// 	return
	// }

	//	fmt.Println("--接收见证人投票广播", hex.EncodeToString(bt.Deposit))

	//TODO 先验证选票是否合法

	//TODO 再判断是否要为他投票

	//	bt := ParseBallotTicket(message.Body.Content)
	//	AddBallotTicket(bt)

	// //继续广播给其他节点
	// if nodeStore.NodeSelf.IsSuper {
	// 	//广播给其他超级节点
	// 	//		mh := utils.Multihash(*message.Body.Content)
	// 	ids := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
	// 	for _, one := range ids {
	// 		//			log.Println("发送给", one.B58String())
	// 		if ss, ok := engine.GetSession(one.B58String()); ok {
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}

	// 	//广播给代理对象
	// 	pids := nodeStore.GetProxyAll()
	// 	for _, one := range pids {
	// 		if ss, ok := engine.GetSession(one); ok {
	// 			//				ss.Send(MSGID_multicast_online_recv, &msg.Data, false)
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}

	// }

}

/*
	接收区块广播
	当矿工挖到一个新的区块后，会广播这个区块
*/
func MulticastBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if !forks.GetLongChain().SyncBlockFinish {
		return
	}

	engine.Log.Info("Receiving block head broadcast")
	// engine.NLog.Debug(engine.LOG_console, "接收区块头广播")
	//	log.Println("接收区块头广播", msg.Session.GetName())

	//判断重复消息
	// if !message.CheckSendhash() {
	// 	engine.Log.Info("This message is repeated")
	// 	return
	// }

	// fmt.Println("接收区块头广播", string(*message.Body.Content))

	bhVO, err := ParseBlockHeadVO(message.Body.Content)
	if err != nil {
		// fmt.Println("解析区块广播错误", err)
		engine.Log.Info("Parse block broadcast error: %s", err.Error())
		return
	}

	// jsonBs, _ := bhVO.Json()
	// fmt.Println("解析区块---------------------:\n", string(*jsonBs))

	if !bhVO.Verify() {
		// panic("区块不合法")
		return
	}

	// engine.Log.Info("接收区块广播，区块高度 %d %s", bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash))

	// fmt.Println("接收区块广播，区块高度", bhVO.BH.Height)
	go forks.AddBlockHead(bhVO)
	//	go ImportBlock(bhVO)

	// //继续广播给其他节点
	// if nodeStore.NodeSelf.IsSuper {
	// 	//广播给其他超级节点
	// 	//		mh := utils.Multihash(*message.Body.Content)
	// 	ids := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
	// 	for _, one := range ids {
	// 		//			log.Println("发送给", one.B58String())
	// 		if ss, ok := engine.GetSession(one.B58String()); ok {
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}

	// 	//广播给代理对象
	// 	pids := nodeStore.GetProxyAll()
	// 	for _, one := range pids {
	// 		if ss, ok := engine.GetSession(one); ok {
	// 			//				ss.Send(MSGID_multicast_online_recv, &msg.Data, false)
	// 			ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
	// 		}
	// 	}
	// }

}

/*
	接收邻居节点区块高度
*/
func FindHeightBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收邻居节点区块高度")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块高度")
	//	log.Println("接收邻居节点区块高度", msg.Session.GetName())

	dataBuf := bytes.NewBuffer([]byte{})
	binary.Write(dataBuf, binary.LittleEndian, forks.GetLongChain().GetStartingBlock())
	binary.Write(dataBuf, binary.LittleEndian, forks.GetLongChain().GetCurrentBlock())
	bs := dataBuf.Bytes()

	message_center.SendNeighborReplyMsg(message, config.MSGID_heightBlock_recv, &bs, msg.Session)

}

/*
	接收邻居节点区块高度_返回
*/
func FindHeightBlock_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收邻居节点区块高度")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块高度")
	//	log.Println("接收邻居节点区块高度", msg.Session.GetName())

	flood.ResponseWait(mc.CLASS_findHeightBlock, hex.EncodeToString(message.Body.Hash), message.Body.Content)
}

/*
	接收邻居节点起始区块头查询
*/
func GetBlockHead(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("++++接收邻居节点区块头查询")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块头查询")
	//	log.Println("接收邻居节点区块头查询", msg.Session.GetName())

	// var bs *[]byte
	bhash, err := db.Find(config.Key_block_start)
	if err != nil {
		return
	}
	// bs = bhash

	chainInfo := ChainInfo{
		StartBlockHash: *bhash,                  //创始区块hash
		HightBlock:     forks.GetHighestBlock(), //最高区块
	}

	bs, err := json.Marshal(chainInfo)
	if err != nil {
		return
	}

	message_center.SendNeighborReplyMsg(message, config.MSGID_getBlockHead_recv, &bs, msg.Session)

}

/*
	区块同步信息
*/
type ChainInfo struct {
	StartBlockHash []byte //创始区块hash
	HightBlock     uint64 //最高区块
}

/*
	接收邻居节点起始区块头查询_返回本次从邻居节点同步区块
*/
func GetBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收邻居节点区块头查询_返回")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块头查询_返回")
	//	log.Println("接收邻居节点区块头查询_返回", msg.Session.GetName())

	flood.ResponseWait(mc.CLASS_getBlockHead, hex.EncodeToString(message.Body.Hash), message.Body.Content)

}

/*
	接收查询交易
	此接口频繁调用占用带宽较多，需要控制好流量
*/
var flowController = make(chan bool, 1)

//等待1秒钟
func flowControllerWaiteSecount() {
	go func() {
		time.Sleep(config.Wallet_sync_block_interval_time)
		select {
		case flowController <- false:
		default:
		}
	}()
}
func GetTransaction(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	//	timeout := time.NewTimer(time.Millisecond * 20) //20毫秒
	//	select {
	//	case <-timeout.C:
	//		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	//		return
	//	case <-flowController:
	//		timeout.Stop()
	//	}
	defer flowControllerWaiteSecount()
	<-flowController

	if message.Body.Content == nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		return
	}

	bhvo := new(BlockHeadVO)
	bhvo.Txs = make([]TxItr, 0)

	//通过区块hash查找区块头
	bs, err := db.Find(*message.Body.Content)
	if err != nil {
		// engine.Log.Error("querying transaction or block Error: %s", err.Error())
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		return
	} else {
		// fmt.Println("查询区块或交易结果2", len(*bs))
		bh, err := ParseBlockHead(bs)
		if err != nil {
			message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
			return
		}
		bhvo.BH = bh
		for _, one := range bh.Tx {
			bs, err := db.Find(one)
			if err != nil {
				message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
				return
			}
			txOne, err := ParseTxBase(bs)
			if err != nil {
				message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
				return
			}
			bhvo.Txs = append(bhvo.Txs, txOne)
		}
	}
	bs, err = bhvo.Json()
	if err != nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		return
	}
	err = message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction or block message Error: %s", err.Error())
	}
}

/*
	接收查询交易_返回
*/
func GetTransaction_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	flood.ResponseWait(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), message.Body.Content)

}

/*
	接收交易广播
*/
func MulticastTransaction_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	//判断重复消息
	// if !message.CheckSendhash() {
	// 	engine.Log.Info("这个消息重复了")
	// 	return
	// }

	//自己处理
	txbase, err := ParseTxBase(message.Body.Content)
	if err != nil {
		// fmt.Println("解析广播的交易错误", err)
		return
	}
	txbase.BuildHash()
	//判断区块是否同步完成，如果没有同步完成则不验证交易合法性
	if GetSyncFinish() {
		//验证交易
		if err := txbase.Check(); err != nil {
			//交易不合法，则不发送出去
			// fmt.Println("交易不合法，则不发送出去")
			return
		}

	}

	//判断是否有重复交易
	if db.CheckHashExist(*txbase.GetHash()) {
		return
	}

	forks.GetLongChain().transactionManager.AddTx(txbase)

}

/*
	从邻居节点获取未确认的区块
*/
func GetUnconfirmedBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	height := utils.BytesToUint64(*message.Body.Content)

	engine.Log.Info("Get the unconfirmed block, the height of this synchronization block %d", height)

	witnessGroup := GetLongChain().witnessChain.witnessGroup

	group := witnessGroup
	for {
		group = group.PreGroup
		if group.BlockGroup != nil {
			break
		}
	}
	block := group.BlockGroup.Blocks[0]
	for i := 0; i < config.Mining_group_max*5; i++ {
		if block == nil || block.Height == height {
			break
		}
		if block.Height > height {
			block = block.PreBlock
		}
		if block.Height < height {
			block = block.NextBlock
		}
	}

	var bs []byte
	bhvos := make([]*BlockHeadVO, 0)
	var err error
	//先查询已经确认的块
	for block != nil {
		bh, txs, e := block.LoadTxs()
		if e != nil {
			err = e
			break
		}
		bhvo := &BlockHeadVO{BH: bh, Txs: *txs}
		bhvos = append(bhvos, bhvo)
		block = block.NextBlock
	}
	if err == nil {
		for {
			if witnessGroup == nil {
				break
			}

			for _, one := range witnessGroup.Witness {
				if one.Block == nil {
					continue
				}

				// fmt.Println("这个block.id怎么会为空", one.Block, witnessGroup.Height)
				bh, txs, e := one.Block.LoadTxs()
				if e != nil {
					err = e
					break
				}
				bhvo := &BlockHeadVO{BH: bh, Txs: *txs}
				bhvos = append(bhvos, bhvo)
				// block = block.NextBlock
			}
			if err != nil {
				break
			}
			witnessGroup = witnessGroup.NextGroup
		}
	}
	if err == nil {
		bsOne, err := json.Marshal(bhvos)
		if err == nil {
			bs = bsOne
		}
	}

	// engine.Log.Info("查询到的区块", bhvos)
	err = message_center.SendNeighborReplyMsg(message, config.MSGID_getUnconfirmedBlock_recv, &bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction or block message Error %s", err.Error())
	}
}

/*
	从邻居节点获取未确认的区块_返回
*/
func GetUnconfirmedBlock_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	flood.ResponseWait(mc.CLASS_getUnconfirmedBlock, hex.EncodeToString(message.Body.Hash), message.Body.Content)

}
