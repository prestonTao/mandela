package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"mandela/sqlite3_db"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/shirou/gopsutil/v3/mem"
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
	message_center.Register_neighbor(config.MSGID_multicast_return, MulticastReturn_recv)             //收到广播消息回复_返回
	message_center.Register_neighbor(config.MSGID_getblockforwitness, GetBlockForWitness)             //从邻居节点获取指定见证人的区块
	message_center.Register_neighbor(config.MSGID_getblockforwitness_recv, GetBlockForWitness_recv)   //从邻居节点获取指定见证人的区块_返回
	message_center.Register_neighbor(config.MSGID_getTransaction_one, GetTransactionOne)              //查询交易
	message_center.Register_neighbor(config.MSGID_getTransaction_one_recv, GetTransactionOne_recv)    //查询交易_返回
	message_center.Register_multicast(config.MSGID_multicast_find_witness, MulticastWitness)          //广播寻找见证人地址
	message_center.Register_p2pHE(config.MSGID_multicast_find_witness_recv, MulticastWitness_recv)    //广播寻找见证人地址 返回
	message_center.Register_neighbor(config.MSGID_getBlockLastCurrent, GetBlockLastCurrent)           //从邻居节点获取已经确认的最高区块
	message_center.Register_neighbor(config.MSGID_getBlockLastCurrent_recv, GetBlockLastCurrent_recv) //从邻居节点获取已经确认的最高区块_返回

	message_center.Register_neighbor(config.MSGID_multicast_witness_blockhead, MulticastBlockHeadHash)              //接收见证人之间的区块广播
	message_center.Register_neighbor(config.MSGID_multicast_witness_blockhead_recv, MulticastBlockHeadHash_recv)    //接收见证人之间的区块广播_返回
	message_center.Register_neighbor(config.MSGID_multicast_witness_blockhead_get, GetMulticastBlockHead)           //接收见证人之间的区块广播
	message_center.Register_neighbor(config.MSGID_multicast_witness_blockhead_get_recv, GetMulticastBlockHead_recv) //接收见证人之间的区块广播_返回

	message_center.Register_neighbor(config.MSGID_uniformity_multicast_witness_blockhead, UniformityMulticastBlockHeadHash)            //接收见证人之间的区块广播
	message_center.Register_neighbor(config.MSGID_uniformity_multicast_witness_blockhead_recv, UniformityMulticastBlockHeadHash_recv)  //接收见证人之间的区块广播_返回
	message_center.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_get, UniformityGetMulticastBlockHead)             //接收见证人之间的区块同步
	message_center.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_get_recv, UniformityGetMulticastBlockHead_recv)   //接收见证人之间的区块同步_返回
	message_center.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_import, UniformityMulticastBlockImport)           //接收见证人之间的区块导入指令
	message_center.Register_neighbor(config.MSGID_uniformity_multicast_witness_block_import_recv, UniformityMulticastBlockImport_recv) //接收见证人之间的区块导入指令_返回

	// flowControllerWaiteSecount()
	// flowControllerWaiteSecountToo()

}

type BlockForWitness struct {
	GroupHeight uint64             //见证人组高度
	Addr        crypto.AddressCoin //见证人地址
}

func (this *BlockForWitness) Proto() (*[]byte, error) {
	bwp := go_protos.BlockForWitness{
		GroupHeight: this.GroupHeight,
		Addr:        this.Addr,
	}
	bs, err := bwp.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, nil
	// return bwp.Marshal()
}

func ParseBlockForWitness(bs *[]byte) (*BlockForWitness, error) {
	if bs == nil {
		return nil, nil
	}
	bwp := new(go_protos.BlockForWitness)
	err := proto.Unmarshal(*bs, bwp)
	if err != nil {
		return nil, err
	}
	bw := &BlockForWitness{
		GroupHeight: bwp.GroupHeight,
		Addr:        bwp.Addr,
	}
	return bw, nil
}

/*
	从邻居节点获取指定见证人的区块
*/
func GetBlockForWitness(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// engine.Log.Info("recv GetBlockForWitness message")
	bs := []byte{}
	if message.Body.Content == nil {
		engine.Log.Warn("GetBlockForWitness message.Body.Content is nil")
		message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	bfw, err := ParseBlockForWitness(message.Body.Content)
	if err != nil {
		engine.Log.Warn("GetBlockForWitness decoder error: %s", err.Error())
		message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
		return
	}

	// bfw := new(BlockForWitness)
	// // var jso = jsoniter.ConfigCompatibleWithStandardLibrary
	// // err := json.Unmarshal(*message.Body.Content, bfw)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(bfw)
	// if err != nil {
	// 	engine.Log.Warn("GetBlockForWitness decoder error: %s", err.Error())
	// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getblockforwitness_recv, &bs, msg.Session)
	// 	return
	// }

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
			bhvo := CreateBlockHeadVO(nil, bh, *tx)
			newbs, err := bhvo.Proto() // bhvo.Json()
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
	// flood.ResponseWait(config.CLASS_wallet_getblockforwitness, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(config.CLASS_wallet_getblockforwitness, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	收到广播消息回复_返回
*/
func MulticastReturn_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	bs := []byte("ok")
	// flood.ResponseWait(config.CLASS_wallet_broadcast_return, hex.EncodeToString(message.Body.Hash), &bs)
	flood.ResponseWait(config.CLASS_wallet_broadcast_return, utils.Bytes2string(message.Body.Hash), &bs)
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
	chain := GetLongChain()
	if chain == nil {
		return
	}
	if !chain.SyncBlockFinish {
		return
	}

	// engine.NLog.Debug(engine.LOG_console, "接收区块头广播")
	//	log.Println("接收区块头广播", msg.Session.GetName())

	//判断重复消息
	// if !message.CheckSendhash() {
	// 	engine.Log.Info("This message is repeated")
	// 	return
	// }

	// fmt.Println("接收区块头广播", string(*message.Body.Content))

	bhVO, err := ParseBlockHeadVOProto(message.Body.Content) //ParseBlockHeadVO(message.Body.Content)
	if err != nil {
		// fmt.Println("解析区块广播错误", err)
		engine.Log.Warn("Parse block broadcast error: %s", err.Error())
		return
	}

	// sessionName := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// jsonBs, _ := bhVO.Json()
	// engine.Log.Info("Receiving block head broadcast from %s", sessionName.B58String())

	if !bhVO.Verify(bhVO.StaretBlockHash) {
		// panic("区块不合法")
		return
	}

	//此区块已存在
	_, err = db.LevelDB.Find(bhVO.BH.Hash)
	if err == nil {
		engine.Log.Warn("this block exist")
		return
	}

	bhVO.FromBroadcast = true

	// engine.Log.Info("接收区块广播，区块高度 %d %s %v", bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash), bhVO.FromBroadcast)
	bhVO.BH.BuildBlockHash()
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

	// flood.ResponseWait(mc.CLASS_findHeightBlock, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(mc.CLASS_findHeightBlock, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	接收邻居节点起始区块头查询
*/
func GetBlockHead(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("++++接收邻居节点区块头查询")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块头查询")
	//	log.Println("接收邻居节点区块头查询", msg.Session.GetName())

	// var bs *[]byte
	bhash, err := db.LevelDB.Find(config.Key_block_start)
	if err != nil {
		return
	}
	// bs = bhash

	chainInfo := ChainInfo{
		StartBlockHash: *bhash,                  //创始区块hash
		HightBlock:     forks.GetHighestBlock(), //最高区块
	}

	bs, err := chainInfo.Proto() // json.Marshal(chainInfo)
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

func (this *ChainInfo) Proto() ([]byte, error) {
	cip := go_protos.ChainInfo{
		StartBlockHash: this.StartBlockHash,
		HightBlock:     this.HightBlock,
	}
	return cip.Marshal()
}

func ParseChainInfo(bs *[]byte) (*ChainInfo, error) {
	if bs == nil {
		return nil, nil
	}
	cip := new(go_protos.ChainInfo)
	err := proto.Unmarshal(*bs, cip)
	if err != nil {
		return nil, err
	}

	ci := ChainInfo{
		StartBlockHash: cip.StartBlockHash, //创始区块hash
		HightBlock:     cip.HightBlock,     //最高区块
	}

	return &ci, nil
}

/*
	接收邻居节点起始区块头查询_返回本次从邻居节点同步区块
*/
func GetBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收邻居节点区块头查询_返回")
	//	engine.NLog.Debug(engine.LOG_console, "接收邻居节点区块头查询_返回")
	//	log.Println("接收邻居节点区块头查询_返回", msg.Session.GetName())

	// flood.ResponseWait(mc.CLASS_getBlockHead, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(mc.CLASS_getBlockHead, utils.Bytes2string(message.Body.Hash), message.Body.Content)

}

/*
	接收查询交易
	此接口频繁调用占用带宽较多，需要控制好流量
*/
// var flowController = make(chan bool, 1)
// var flowControllerRelax = make(chan bool, 1)

// //等待1秒钟
// func flowControllerWaiteSecount() {
// 	go func() {
// 		for range time.NewTicker(config.Wallet_sync_block_interval_time).C {
// 			select {
// 			case flowController <- false:
// 			default:
// 			}
// 		}
// 	}()

// 	go func() {
// 		for range time.NewTicker(config.Wallet_sync_block_interval_time_relax).C {
// 			select {
// 			case flowControllerRelax <- false:
// 			default:
// 			}
// 		}
// 	}()
// }
func GetTransaction(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	utils.SetTimeToken(config.TIMETOKEN_GetTransaction, config.Wallet_sync_block_interval_time)
	utils.SetTimeToken(config.TIMETOKEN_GetTransactionRelax, config.Wallet_sync_block_interval_time_relax)
	//未同步完成则不给别人同步，因为在启动节点的时候，大量的请求需要处理，导致自己启动非常慢，缺点会导致节点为同步完成之前，连接它的节点会获取不到创始块而崩溃
	//未同步完成，只给同步初始区块
	// if !GetSyncFinish() && !bytes.Equal(config.StartBlockHash, *message.Body.Content) {
	// 	// engine.Log.Warn("未同步完成则不分配奖励")
	// 	engine.Log.Warn("If the synchronization is not completed, no reward will be allocated")
	// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	// 	return
	// }
	chain := GetLongChain()
	if chain == nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	}
	//节点限速
	if !chain.SyncBlockFinish {
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else if CheckNameStore() { //判断是否是存储节点
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else if chain.witnessChain.FindWitness(keystore.GetCoinbase().Addr) { //是见证人
		utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
		// <-flowController
	} else {
		utils.GetTimeToken(config.TIMETOKEN_GetTransactionRelax, true)
		// <-flowControllerRelax
	}
	//这里查看内存，控制速度
	memInfo, _ := mem.VirtualMemory()
	if memInfo.UsedPercent > 92 {
		time.Sleep(time.Second)
	}
	// <-flowController

	//获取存储超级节点地址
	// nameinfo := name.FindName(config.Name_store)
	// if nameinfo != nil {
	// 	//判断域名是否是自己的
	// 	have := false
	// 	for _, one := range nameinfo.NetIds {
	// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, one) {
	// 			have = true
	// 			break
	// 		}
	// 	}
	// 	//在列表里，则退出，存储节点避免网络占用，不作同步服务器
	// 	if have {
	// 		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	// 		return
	// 	}
	// }

	//	timeout := time.NewTimer(time.Millisecond * 20) //20毫秒
	//	select {
	//	case <-timeout.C:
	//		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
	//		return
	//	case <-flowController:
	//		timeout.Stop()
	//	}
	// defer flowControllerWaiteSecount()

	//收到邻居节点查询区块消息
	// engine.Log.Info("Received neighbor node query block message")

	netid := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// addrNet := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// engine.Log.Info("Received neighbor node query block message %s", netid.B58String())

	if message.Body.Content == nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	}

	bhvo := new(BlockHeadVO)
	bhvo.Txs = make([]TxItr, 0)

	//通过区块hash查找区块头
	// bs, err := db.Find(*message.Body.Content)
	bh, err := LoadBlockHeadByHash(message.Body.Content)
	if err != nil {
		// engine.Log.Error("querying transaction or block Error: %s", err.Error())
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		engine.Log.Info("return nil message")
		return
	} else {
		// fmt.Println("查询区块或交易结果2", len(*bs))
		// bh, err := ParseBlockHead(bs)
		// if err != nil {
		// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		// 	// engine.Log.Info("return error message")
		// 	return
		// }
		bhvo.BH = bh
		for _, one := range bh.Tx {
			// txOne, err := FindTxBase(one, hex.EncodeToString(one))
			// txOne, err := FindTxBase(one)
			txOne, err := LoadTxBase(one)
			// bs, err := db.Find(one)
			// if err != nil {
			// 	message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
			// 	return
			// }
			// txOne, err := ParseTxBase(ParseTxClass(one), bs)
			if err != nil {
				message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
				engine.Log.Info("return error message")
				return
			}
			bhvo.Txs = append(bhvo.Txs, txOne)
		}
	}
	if bhvo.BH.Nextblockhash == nil {

		if GetHighestBlock() > bhvo.BH.Height+1 {
			engine.Log.Info("neighbor %s find next block %d hash nil. hight:%d", netid.B58String(), bhvo.BH.Height, GetHighestBlock())
		}
		//TODO Nextblockhash为空，则补充一个，临时解决方案
		tempGroup := GetLongChain().witnessChain.witnessGroup
		for tempGroup != nil {

			if tempGroup.Height < bhvo.BH.GroupHeight {
				break
			}
			if tempGroup.Height > bhvo.BH.GroupHeight {
				tempGroup = tempGroup.PreGroup
				continue
			}
			for _, one := range tempGroup.Witness {
				if one.Block == nil {
					continue
				}
				if one.Block.Height == bhvo.BH.Height {
					if one.Block.NextBlock != nil {
						engine.Log.Info("neighbor %s find next block %d hash nil.", netid.B58String(), bhvo.BH.Height)
						bhvo.BH.Nextblockhash = one.Block.NextBlock.Id
					}
					tempGroup = nil
					break
				}
			}
			break
		}
	}

	bs, err := bhvo.Proto() // bhvo.Json()
	if err != nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, nil, msg.Session)
		engine.Log.Info("return json fialt message")
		return
	}
	err = message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_recv, bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction or block message Error: %s", err.Error())
	} else {
		// engine.Log.Info("return success message: %s", netid.B58String())
	}
}

/*
	接收查询交易_返回
*/
func GetTransaction_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// flood.ResponseWait(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(mc.CLASS_getTransaction, utils.Bytes2string(message.Body.Hash), message.Body.Content)

}

/*
	接收交易广播
*/
func MulticastTransaction_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// engine.Log.Debug("MulticastTransaction_recv %s", string(*message.Body.Content))

	//判断重复消息
	// if !message.CheckSendhash() {
	// 	engine.Log.Info("这个消息重复了")
	// 	return
	// }

	//自己处理
	txbase, err := ParseTxBaseProto(0, message.Body.Content) // ParseTxBase(0, message.Body.Content)
	if err != nil {
		// fmt.Println("解析广播的交易错误", err)
		engine.Log.Warn("Broadcast transaction format error %s", err.Error())
		return
	}
	engine.Log.Debug("Broadcast transaction received %s", hex.EncodeToString(*txbase.GetHash()))
	// bs, _ := json.Marshal(txbase.GetVOJSON())
	// engine.Log.Debug("接收到广播交易 %s", string(bs))
	txbase.BuildHash()
	// engine.Log.Debug("接收到广播交易 222 %s", txbase.GetHashStr())
	//判断区块是否同步完成，如果没有同步完成则无法验证交易合法性，则不保存交易

	chain := GetLongChain()
	if chain == nil {
		return
	}
	if !chain.SyncBlockFinish {
		return
	}

	checkTxQueue <- txbase
	//

	// if !GetSyncFinish() {
	// 	// engine.Log.Warn("未同步完成则无法验证交易合法性")
	// 	return
	// }
	// if len(*txbase.GetVin()) > config.Mining_pay_vin_max {
	// 	//交易太大了
	// 	engine.Log.Warn(config.ERROR_pay_vin_too_much.Error())
	// 	return
	// }
	// //验证交易
	// if err := txbase.CheckLockHeight(GetHighestBlock()); err != nil {
	// 	// engine.Log.Warn("验证交易锁定高度失败")
	// 	engine.Log.Warn("Failed to verify transaction lock height")
	// 	return
	// }
	// // txbase.CheckFrozenHeight(GetHighestBlock())

	// //加载相关交易到缓存
	// keys := make(map[string]uint64, 0) //记录加载了哪些交易到缓存
	// for _, one := range *txbase.GetVin() {
	// 	//已经有了就不用重复查询了
	// 	if _, ok := TxCache.FindTxInCache(one.Txid); !ok {
	// 		continue
	// 	}

	// 	// txItr, err := FindTxBase(one.Txid)
	// 	txItr, err := LoadTxBase(one.Txid)
	// 	if err != nil {
	// 		return
	// 	}
	// 	TxCache.AddTxInCache(one.Txid, txItr)
	// 	// key := utils.Bytes2string(one.Txid) //one.GetTxidStr()
	// 	keys[utils.Bytes2string(one.Txid)] = one.Vout
	// 	// keys = append(keys, key)
	// }

	// // bs, _ := txbase.Json()
	// // engine.Log.Info("交易\n%s", string(*bs))

	// if GetHighestBlock() > config.Mining_block_start_height+config.Mining_block_start_height_jump {
	// 	if err := txbase.Check(); err != nil {
	// 		//交易不合法，则不发送出去
	// 		//验证未通过，删除缓存
	// 		// for k, v := range keys {
	// 		// 	TxCache.RemoveTxInCache(k, v)
	// 		// }
	// 		runtime.GC()
	// 		engine.Log.Warn("Failed to verify transaction signature %s %s", hex.EncodeToString(*txbase.GetHash()), err.Error())
	// 		return
	// 	}
	// 	runtime.GC()
	// }

	// //判断是否有重复交易，并且检查是否有无效区块的交易的标记
	// if db.LevelDB.CheckHashExist(*txbase.GetHash()) && !db.LevelDB.CheckHashExist(config.BuildTxNotImport(*txbase.GetHash())) {
	// 	//验证未通过，删除缓存
	// 	// for k, v := range keys {
	// 	// 	TxCache.RemoveTxInCache(k, v)
	// 	// }
	// 	engine.Log.Warn("Transaction hash collision is the same %s", hex.EncodeToString(*txbase.GetHash()))
	// 	return
	// }

	// forks.GetLongChain().transactionManager.AddTx(txbase)

}

// var flowControllerToo = make(chan bool, 1)

// //等待1秒钟
// func flowControllerWaiteSecountToo() {
// 	go func() {
// 		for range time.NewTicker(time.Second).C {
// 			select {
// 			case flowControllerToo <- false:
// 			default:
// 			}
// 		}
// 	}()
// }

/*
	从邻居节点获取未确认的区块
*/
func GetUnconfirmedBlock(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	utils.SetTimeToken(config.TIMETOKEN_GetUnconfirmedBlock, time.Second)
	utils.GetTimeToken(config.TIMETOKEN_GetUnconfirmedBlock, true)

	// <-flowControllerToo
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
		// bsOne, err := json.Marshal(bhvos)
		// if err == nil {
		// 	bs = bsOne
		// }

		rbsp := go_protos.RepeatedBytes{
			Bss: make([][]byte, 0),
		}
		for _, one := range bhvos {
			bsOne, err := one.Proto()
			if err != nil {
				return
			}
			rbsp.Bss = append(rbsp.Bss, *bsOne)
		}
		bsOne, err := rbsp.Marshal()

		// bsOne, err := json.Marshal(bhvos)
		if err == nil {
			bs = bsOne
		}
	}

	// for _, one := range bhvos {
	// 	engine.Log.Info("find block one:%d", one.BH.Height)
	// }
	err = message_center.SendNeighborReplyMsg(message, config.MSGID_getUnconfirmedBlock_recv, &bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction or block message Error %s", err.Error())
	}
}

/*
	从邻居节点获取未确认的区块_返回
*/
func GetUnconfirmedBlock_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// flood.ResponseWait(mc.CLASS_getUnconfirmedBlock, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(mc.CLASS_getUnconfirmedBlock, utils.Bytes2string(message.Body.Hash), message.Body.Content)

}

func GetTransactionOne(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	utils.SetTimeToken(config.TIMETOKEN_GetTransaction, config.Wallet_sync_block_interval_time)
	utils.GetTimeToken(config.TIMETOKEN_GetTransaction, true)
	// defer flowControllerWaiteSecount()
	// <-flowController

	if message.Body.Content == nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_one_recv, nil, msg.Session)
		return
	}

	bs, err := db.LevelDB.Find(*message.Body.Content)
	if err != nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_one_recv, nil, msg.Session)
		return
	}

	err = message_center.SendNeighborReplyMsg(message, config.MSGID_getTransaction_one_recv, bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning query transaction message Error: %s", err.Error())
	}
}

/*
	接收查询交易_返回
*/
func GetTransactionOne_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	flood.ResponseWait(mc.CLASS_getTransaction, utils.Bytes2string(message.Body.Hash), message.Body.Content)

	// flood.ResponseWait(mc.CLASS_getTransaction_one,  utils.Bytes2string(message.Body.Hash), message.Body.Content)

}

/*
	接收见证人地址发现广播
*/
func MulticastWitness(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	chain := GetLongChain()
	if chain == nil {
		return
	}

	IsBackup := chain.witnessChain.FindWitness(keystore.GetCoinbase().Addr)
	if !IsBackup {
		return
	}

	// ipv4, err := utils.IPV4String2Long(nodeStore.NodeSelf.Addr)
	// if err != nil {
	// 	return
	// }
	// ipBs := utils.Uint32ToBytes(ipv4)

	// portBs := utils.Uint16ToBytes(nodeStore.NodeSelf.TcpPort)
	// bs := append(ipBs, portBs...)
	// _, ok := message_center.SendP2pMsg(message, config.MSGID_multicast_find_witness_recv, &bs, msg.Session)
	// if ok {
	// 	engine.Log.Info("returning query transaction message Error: %s", err.Error())
	// }

}

/*
	接收见证人地址发现广播_返回
*/
func MulticastWitness_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// flood.ResponseWait(mc.CLASS_getTransaction, utils.Bytes2string(message.Body.Hash), message.Body.Content)
	// flood.ResponseWait(mc.CLASS_getTransaction_one,  utils.Bytes2string(message.Body.Hash), message.Body.Content)

}

/*
	从邻居节点获取已经确认的最高区块
*/
func GetBlockLastCurrent(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	_, block := GetLongChain().GetLastBlock()

	bh, err := block.Load()
	if err != nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, nil, msg.Session)
		return
	}
	bs, err := bh.Proto()
	if err != nil {
		message_center.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, nil, msg.Session)
		return
	}
	err = message_center.SendNeighborReplyMsg(message, config.MSGID_getBlockLastCurrent_recv, bs, msg.Session)
	if err != nil {
		engine.Log.Info("returning GetBlockLastCurrent Error %s", err.Error())
	}
}

/*
	从邻居节点获取已经确认的最高区块_返回
*/
func GetBlockLastCurrent_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(mc.CLASS_getBlockLastCurrent, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

var syncHashLock = new(sync.Mutex)
var blockheadHashMap = make(map[string]int)

/*
	接收见证人出块的区块hash
*/
func MulticastBlockHeadHash(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// engine.Log.Info("11111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	if !forks.GetLongChain().SyncBlockFinish {
		//区块未同步好，直接回复对方已经收到区块
		// engine.Log.Info("222222222222222222222")
		message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, nil, msg.Session)
		return
	}

	// newmessage, err := message_center.ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	//广播消息头解析失败
	// 	engine.Log.Warn("Parsing of this broadcast header failed")
	// 	return
	// }
	// //解析包体内容
	// if err = newmessage.ParserContentProto(); err != nil {
	// 	engine.Log.Info("Content parsing of this broadcast message failed %s", err.Error())
	// 	return
	// }
	var bhVO *BlockHeadVO
	success := false
	syncHashLock.Lock()
	_, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	if err == nil {
		// engine.Log.Info("333333333333333333333")
		//已经下载好了，就回复对方
		message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, nil, msg.Session)
	} else {
		// engine.Log.Info("4444444444444444444444")
		//去下载
		addrNet := nodeStore.AddressNet(msg.Session.GetName())
		newmsg, _ := message_center.SendNeighborMsg(config.MSGID_multicast_witness_blockhead_get, &addrNet, message.Body.Content)
		bs, err := flood.WaitRequest(config.CLASS_witness_get_blockhead, utils.Bytes2string(newmsg.Body.Hash), int64(4))
		if err != nil {
			// engine.Log.Info("4444444444444444444444 111111111111111111")
			engine.Log.Warn("Timeout receiving broadcast reply message %s %s", addrNet.B58String(), hex.EncodeToString(newmsg.Body.Hash))
			// failNode = append(failNode, broadcasts[j])
			// continue
		} else {
			//下载成功
			// engine.Log.Info("4444444444444444444444 222222222222222222222")
			//验证同步到的消息
			mmp := new(go_protos.MessageMulticast)
			err = proto.Unmarshal(*bs, mmp)
			if err != nil {
				engine.Log.Error("proto unmarshal error %s", err.Error())
			} else {

				//解析获取到的交易
				bhVOmessage, err := message_center.ParserMessageProto(mmp.Head, mmp.Body, 0)
				if err != nil {
					engine.Log.Error("proto unmarshal error %s", err.Error())
				} else {

					err = bhVOmessage.ParserContentProto()
					if err != nil {
						engine.Log.Error("proto unmarshal error %s", err.Error())
					} else {

						bhVO, err = ParseBlockHeadVOProto(bhVOmessage.Body.Content) //ParseBlockHeadVO(message.Body.Content)
						if err != nil {
							engine.Log.Warn("Parse block broadcast error: %s", err.Error())
						} else {
							//回复消息
							message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, nil, msg.Session)
							//
							err = new(sqlite3_db.MessageCache).Add(*message.Body.Content, mmp.Head, mmp.Body)
							if err != nil {
								engine.Log.Error(err.Error())
							} else {
								// engine.Log.Info("4444444444444444444444 vvvvvvvvvvvvvvvvvvvvv")
								success = true
							}
						}
					}
				}
			}
		}
	}
	syncHashLock.Unlock()

	if !success {
		// engine.Log.Info("555555555555555555555555")
		return
	}
	// engine.Log.Info("666666666666666666666")
	//广播给其他人
	whiltlistNodes := nodeStore.GetWhiltListNodes()
	err = message_center.BroadcastsAll(1, config.MSGID_multicast_witness_blockhead, whiltlistNodes, nil, nil, message.Body.Content)
	if err != nil {
		// engine.Log.Info("7777777777777777777777")
		//广播超时，网络不好，则不导入这个区块
		return
	}
	// engine.Log.Info("888888888888888888888")
	// bhVO, err := ParseBlockHeadVOProto(message.Body.Content) //ParseBlockHeadVO(message.Body.Content)
	// if err != nil {
	// 	// fmt.Println("解析区块广播错误", err)
	// 	engine.Log.Warn("Parse block broadcast error: %s", err.Error())
	// 	return
	// }

	if !bhVO.Verify(bhVO.StaretBlockHash) {
		// panic("区块不合法")
		return
	}

	//此区块已存在
	_, err = db.LevelDB.Find(bhVO.BH.Hash)
	if err == nil {
		engine.Log.Warn("this block exist")
		return
	}

	bhVO.FromBroadcast = true

	// engine.Log.Info("接收区块广播，区块高度 %d %s %v", bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash), bhVO.FromBroadcast)
	bhVO.BH.BuildBlockHash()
	go forks.AddBlockHead(bhVO)

	//广播区块
	go MulticastBlock(*bhVO)

	// MulticastBlockAndImport(bhVO)

	// err = message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, bs, msg.Session)
	// if err != nil {
	// 	engine.Log.Info("returning GetBlockLastCurrent Error %s", err.Error())
	// }

}

/*
	接收见证人出块的区块hash
*/
func MulticastBlockHeadHash_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(config.CLASS_wallet_broadcast_return, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	接收见证人出块的区块hash
*/
func GetMulticastBlockHead(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// if !forks.GetLongChain().SyncBlockFinish {
	// 	//区块未同步好，直接回复对方已经收到区块
	// 	return
	// }
	// engine.Log.Info("GetMulticastBlockHead 11111111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	messageCache, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		// engine.Log.Info("GetMulticastBlockHead 2222222222222222222222")
		engine.Log.Error("find message hash error:%s", err.Error())
		return
	}

	mmp := go_protos.MessageMulticast{
		Head: messageCache.Head,
		Body: messageCache.Body,
	}

	content, err := mmp.Marshal()
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		engine.Log.Error(err.Error())
		return
	}
	// engine.Log.Info("GetMulticastBlockHead 4444444444444444444")
	//回复消息
	message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_get_recv, &content, msg.Session)

}

/*
	接收见证人出块的区块hash_返回
*/
func GetMulticastBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(config.CLASS_witness_get_blockhead, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

var syncUniformityHashLock = new(sync.Mutex)

// var blockheadHashMap = make(map[string]int)

/*
	接收见证人出块的区块hash
*/
func UniformityMulticastBlockHeadHash(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// second := utils.GetRandNum(10)
	// time.Sleep(time.Second * time.Duration(second))

	// engine.Log.Info("11111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	if !forks.GetLongChain().SyncBlockFinish {
		//区块未同步好，直接回复对方已经收到区块
		// engine.Log.Info("222222222222222222222")
		message_center.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, nil, msg.Session)
		return
	}
	var bhVO *BlockHeadVO
	success := false
	syncUniformityHashLock.Lock()
	_, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	if err == nil {
		// engine.Log.Info("333333333333333333333")
		//已经下载好了，就回复对方
		message_center.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, nil, msg.Session)
	} else {
		// engine.Log.Info("4444444444444444444444")
		//去下载
		addrNet := nodeStore.AddressNet(msg.Session.GetName())

		bs, err := message_center.SendNeighborWithReplyMsg(config.MSGID_uniformity_multicast_witness_block_get, &addrNet, message.Body.Content,
			config.CLASS_uniformity_witness_get_blockhead, config.Wallet_sync_block_timeout)

		// newmsg, _ := message_center.SendNeighborMsg(config.MSGID_multicast_witness_blockhead_get, &addrNet, message.Body.Content)
		// bs, err := flood.WaitRequest(config.CLASS_uniformity_witness_get_blockhead, utils.Bytes2string(newmsg.Body.Hash), int64(4))
		if err != nil {
			// engine.Log.Info("4444444444444444444444 111111111111111111")
			engine.Log.Warn("Timeout receiving broadcast reply message %s", addrNet.B58String())
			// failNode = append(failNode, broadcasts[j])
			// continue
		} else {
			//下载成功
			// engine.Log.Info("4444444444444444444444 222222222222222222222")
			//验证同步到的消息
			mmp := new(go_protos.MessageMulticast)
			err = proto.Unmarshal(*bs, mmp)
			if err != nil {
				engine.Log.Error("proto unmarshal error %s", err.Error())
			} else {

				//解析获取到的交易
				bhVOmessage, err := message_center.ParserMessageProto(mmp.Head, mmp.Body, 0)
				if err != nil {
					engine.Log.Error("proto unmarshal error %s", err.Error())
				} else {

					err = bhVOmessage.ParserContentProto()
					if err != nil {
						engine.Log.Error("proto unmarshal error %s", err.Error())
					} else {

						bhVO, err = ParseBlockHeadVOProto(bhVOmessage.Body.Content) //ParseBlockHeadVO(message.Body.Content)
						if err != nil {
							engine.Log.Warn("Parse block broadcast error: %s", err.Error())
						} else {
							//回复消息
							message_center.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_blockhead_recv, nil, msg.Session)
							//
							err = new(sqlite3_db.MessageCache).Add(*message.Body.Content, mmp.Head, mmp.Body)
							if err != nil {
								engine.Log.Error(err.Error())
							} else {
								// engine.Log.Info("4444444444444444444444 vvvvvvvvvvvvvvvvvvvvv")
								success = true
							}
						}
					}
				}
			}
		}
	}
	syncUniformityHashLock.Unlock()

	if !success {
		// engine.Log.Info("555555555555555555555555")
		return
	}
	// engine.Log.Info("666666666666666666666")
	//广播给其他人
	// whiltlistNodes := nodeStore.GetWhiltListNodes()
	// err = message_center.BroadcastsAll(1, config.MSGID_multicast_witness_blockhead, whiltlistNodes, nil, nil, message.Body.Content)
	// if err != nil {
	// 	// engine.Log.Info("7777777777777777777777")
	// 	//广播超时，网络不好，则不导入这个区块
	// 	return
	// }
	// engine.Log.Info("888888888888888888888")
	// bhVO, err := ParseBlockHeadVOProto(message.Body.Content) //ParseBlockHeadVO(message.Body.Content)
	// if err != nil {
	// 	// fmt.Println("解析区块广播错误", err)
	// 	engine.Log.Warn("Parse block broadcast error: %s", err.Error())
	// 	return
	// }

	if !bhVO.Verify(bhVO.StaretBlockHash) {
		// panic("区块不合法")
		return
	}

	//此区块已存在
	_, err = db.LevelDB.Find(bhVO.BH.Hash)
	if err == nil {
		engine.Log.Warn("this block exist")
		return
	}

	bhVO.FromBroadcast = true
	bhVO.BH.BuildBlockHash()

	AddBlockToCache(bhVO)

	//放入缓存

	// engine.Log.Info("接收区块广播，区块高度 %d %s %v", bhVO.BH.Height, hex.EncodeToString(bhVO.BH.Hash), bhVO.FromBroadcast)
	// go forks.AddBlockHead(bhVO)

	//广播区块
	// go MulticastBlock(*bhVO)

	// MulticastBlockAndImport(bhVO)

	// err = message_center.SendNeighborReplyMsg(message, config.MSGID_multicast_witness_blockhead_recv, bs, msg.Session)
	// if err != nil {
	// 	engine.Log.Info("returning GetBlockLastCurrent Error %s", err.Error())
	// }

}

/*
	接收见证人出块的区块hash
*/
func UniformityMulticastBlockHeadHash_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(config.CLASS_uniformity_witness_multicas_blockhead, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	接收见证人出块的区块hash
*/
func UniformityGetMulticastBlockHead(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// second := utils.GetRandNum(10)
	// time.Sleep(time.Second * time.Duration(second))
	// if !forks.GetLongChain().SyncBlockFinish {
	// 	//区块未同步好，直接回复对方已经收到区块
	// 	return
	// }
	// engine.Log.Info("GetMulticastBlockHead 11111111111111111111111 %s", hex.EncodeToString(*message.Body.Content))
	messageCache, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		// engine.Log.Info("GetMulticastBlockHead 2222222222222222222222")
		engine.Log.Error("find message hash error:%s", err.Error())
		return
	}

	mmp := go_protos.MessageMulticast{
		Head: messageCache.Head,
		Body: messageCache.Body,
	}

	content, err := mmp.Marshal()
	if err != nil {
		// if is {
		// engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		// }
		engine.Log.Error(err.Error())
		return
	}
	// engine.Log.Info("GetMulticastBlockHead 4444444444444444444")
	//回复消息
	message_center.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_block_get_recv, &content, msg.Session)

}

/*
	接收见证人出块的区块hash_返回
*/
func UniformityGetMulticastBlockHead_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(config.CLASS_uniformity_witness_get_blockhead, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	接收见证人导入区块命令
*/
func UniformityMulticastBlockImport(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	message_center.SendNeighborReplyMsg(message, config.MSGID_uniformity_multicast_witness_block_import_recv, nil, msg.Session)
	if !forks.GetLongChain().SyncBlockFinish {
		//区块未同步好，直接回复对方已经收到区块
		return
	}

	ImportBlockByCache(message.Body.Content)
	// engine.Log.Info("GetMulticastBlockHead 11111111111111111111111 %s", hex.EncodeToString(*message.Body.Content))

	// engine.Log.Info("GetMulticastBlockHead 4444444444444444444")
	//回复消息

}

/*
	接收见证人导入区块命令_返回
*/
func UniformityMulticastBlockImport_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(config.CLASS_uniformity_witness_multicas_block_import, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}
