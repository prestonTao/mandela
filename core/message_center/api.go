package message_center

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/core/utils/crypto/dh"
	"mandela/core/virtual_node"
	"mandela/sqlite3_db"
	"bytes"
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

var router = NewRouter()

/*
	注册邻居节点消息，不转发
*/
func Register_neighbor(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送一个新邻居节点消息
	@return    bool    消息是否发送成功
*/
func SendNeighborMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*Message, error) {
	//直接查找最近的超级节点
	mhead := NewMessageHead(recvid, recvid, false)
	mbody := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(mhead, mbody)
	message.BuildHash()
	// session, ok := engine.GetSession(recvid.B58String())
	session, ok := engine.GetSession(utils.Bytes2string(*recvid))
	if ok {
		// fmt.Println("给这个session发送消息", recvid.B58String())
		// err := session.Send(version_neighbor, mhead.JSON(), mbody.JSON(), false)
		mheadBs := mhead.Proto()
		mbodyBs := mbody.Proto()
		err := session.Send(version_neighbor, &mheadBs, &mbodyBs, false)
		if err != nil {
			engine.Log.Error("msg send error %s", err.Error())
			return message, err
		}
		return message, nil
	} else {
		engine.Log.Error("get node conn fail %s", recvid.B58String())
		nodeStore.DelNode(recvid)
		return message, config.ERROR_get_node_conn_fail
	}
}

/*
	对某个消息回复
*/
func SendNeighborReplyMsg(message *Message, msgid uint64, content *[]byte, session engine.Session) error {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)

	head := NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	// engine.Log.Info("看看是哪个指针为空 %v", message.Body)
	body := NewMessageBody(msgid, content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	newmessage := NewMessage(head, body)
	newmessage.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// return session.Send(version_neighbor, head.JSON(), body.JSON(), false)
	mheadBs := head.Proto()
	mbodyBs := body.Proto()
	return session.Send(version_neighbor, &mheadBs, &mbodyBs, false)
}

func SendNeighborWithReplyMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte, waitRequestClass string, timeout int64) (*[]byte, error) {
	message, err := SendNeighborMsg(msgid, recvid, content)
	if err != nil {
		return nil, err
	}
	bs, err := flood.WaitRequest(waitRequestClass, utils.Bytes2string(message.Body.Hash), timeout)
	if err != nil {
		return nil, err
	}
	return bs, nil
}

/*
	注册广播消息
*/
func Register_multicast(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送一个新的广播消息
*/
// func SendMulticastMsg(msgid uint64, content *[]byte) bool {
// 	//TODO 这里查询缓存中的节点信息，拿到节点的超级节点信息
// 	head := NewMessageHead(nil, nil, false)
// 	body := NewMessageBody(msgid, content, "", nil, 0)
// 	message := NewMessage(head, body)
// 	message.BuildHash()

// 	failNode := append(nodeStore.GetLogicNodes(), nodeStore.GetProxyAll()...)
// 	failNode = append(failNode, nodeStore.GetOtherNodes()...)
// 	// engine.GetAllSession()
// 	//排除重复的地址
// 	failNode = nodeStore.RemoveDuplicateAddress(failNode)

// 	broadcasts := make([]*nodeStore.AddressNet, 0)
// 	//广播失败的节点最多广播5次
// 	for i := 0; i < 1; i++ {
// 		if len(failNode) <= 0 {
// 			break
// 		}
// 		// engine.Log.Info("第 %d 轮发送广播", i)
// 		broadcasts = failNode
// 		failNode = make([]*nodeStore.AddressNet, 0)

// 		for j, _ := range broadcasts {
// 			//区块广播给节点
// 			// engine.Log.Warn("广播给节点 %d %s", j, addrStr)
// 			// if ss, ok := engine.GetSession(broadcasts[j].B58String()); ok {
// 			if ss, ok := engine.GetSession(utils.Bytes2string(*broadcasts[j])); ok {
// 				err := ss.Send(version_multicast, head.JSON(), body.JSON(), false)
// 				if err != nil {
// 					engine.Log.Warn("send multicast message fail %s", err.Error()) // fmt.Println("发送广播消息失败", err)
// 					continue
// 				} else {
// 					// hashStr := hex.EncodeToString(message.Body.Hash)
// 					// engine.Log.Warn("广播消息 1111111111111 %s", hashStr)
// 					// bs := flood.WaitRequest(config.CLASS_wallet_broadcast_return, hex.EncodeToString(message.Body.Hash), int64(i+1))
// 					bs := flood.WaitRequest(config.CLASS_wallet_broadcast_return, utils.Bytes2string(message.Body.Hash), int64(i+1))
// 					if bs == nil {
// 						engine.Log.Warn("Timeout receiving broadcast reply message %s %s", broadcasts[j].B58String(), hex.EncodeToString(message.Body.Hash))
// 						failNode = append(failNode, broadcasts[j])
// 						continue
// 					}
// 					// engine.Log.Warn("收到广播回复消息  %s %s", addrStr, hashStr)
// 				}
// 			} else {
// 				engine.Log.Warn("get node conn fail")
// 			}

// 		}
// 	}
// 	return true
// }

/*
	发送一个新的广播消息
*/
func SendMulticastMsg(msgid uint64, content *[]byte) bool {
	// engine.Log.Info("发送广播 start")
	//TODO 这里查询缓存中的节点信息，拿到节点的超级节点信息
	head := NewMessageHead(nil, nil, false)
	body := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(head, body)
	message.BuildHash()

	// headBs := message.Head.JSON()
	// bodyBs := message.Body.JSON()

	//先保存这个消息到缓存
	err := new(sqlite3_db.MessageCache).Add(message.Body.Hash, head.Proto(), body.Proto())
	if err != nil {
		engine.Log.Error(err.Error())
		return false
	}

	//先发送给超级节点
	// superNodes := nodeStore.GetLogicNodes()
	// superNodes = append(superNodes, nodeStore.GetNodesClient()...)

	// //再发送给内网节点
	// proxyNodes := nodeStore.GetProxyAll()

	// broadcastsAll(superNodes, proxyNodes, message)
	//先发送给超级节点
	// whiltlistNodes := nodeStore.GetWhiltListNodes()

	superNodes := nodeStore.GetLogicNodes()
	superNodes = append(superNodes, nodeStore.GetNodesClient()...)

	//再发送给内网节点
	proxyNodes := nodeStore.GetProxyAll()

	BroadcastsAll(version_multicast, 0, nil, superNodes, proxyNodes, &message.Body.Hash)
	// engine.Log.Info("发送广播 end")
	return true

}

func BroadcastsAll(msgid, p2pMsgid uint64, whiltlistNodes, superNodes, proxyNodes []nodeStore.AddressNet, hash *[]byte) error {

	// start := time.Now()
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
			ok := broadcastsOne(msgid, p2pMsgid, cs, sessionid, hash, group, true)
			if ok {
				return
			}
			timeouterrorlock.Lock()
			timeouterror = config.ERROR_wait_msg_timeout
			timeouterrorlock.Lock()
		})
	}
	group.Wait()
	// engine.Log.Info("multicast whilt list node time %s", time.Now().Sub(start))

	// start = time.Now()
	for i, _ := range superNodes {
		sessionid := superNodes[i]
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
		go broadcastsOne(msgid, p2pMsgid, cs, sessionid, hash, group, false)
	}
	group.Wait()
	// engine.Log.Info("multicast super node time %s", time.Now().Sub(start))
	// start = time.Now()

	//再发送给内网节点
	// proxyNodes := nodeStore.GetProxyAll()
	//排除重复的地址
	// proxyNodes = nodeStore.RemoveDuplicateAddress(proxyNodes)
	for i, _ := range proxyNodes {
		sessionid := proxyNodes[i]
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
		go broadcastsOne(msgid, p2pMsgid, cs, sessionid, hash, group, false)
	}
	group.Wait()
	// engine.Log.Info("multicast proxy node time %s", time.Now().Sub(start))
	return timeouterror
}

/*
	给一个节点发送广播消息
	@return    bool    白名单种节点是否都收到了消息
*/
func broadcastsOne(msgid, p2pMsgid uint64, cs chan bool, sid nodeStore.AddressNet, hash *[]byte, group *sync.WaitGroup, reply bool) (success bool) {
	// engine.Log.Info("multicast to %s", sid.B58String())
	//先查询对方的版本,看协议是否有升级
	// node := nodeStore.FindNode(&sid)
	// if node == nil {
	// 	<-cs
	// 	group.Done()
	// 	return
	// }
	// if nodeStore.NodeSelf.Version >= config.Version_1 && node.Version >= config.Version_1 {

	// }
	head := NewMessageHead(nil, nil, false)
	body := NewMessageBody(p2pMsgid, hash, 0, nil, 0) //广播采用同步机制后，不需要真实msgid，所以设置为0
	msg := NewMessage(head, body)
	msg.BuildHash()
	//如果是普通节点，则不需要对方回复
	if !reply {
		if ss, ok := engine.GetSession(utils.Bytes2string(sid)); ok {
			// ss.Send(version_multicast, msg.Head.JSON(), msg.Body.JSON(), false)
			mheadBs := msg.Head.Proto()
			mbodyBs := msg.Body.Proto()
			ss.Send(version_multicast, &mheadBs, &mbodyBs, false)
		} else {
			engine.Log.Warn("get node conn fail %s", sid.B58String())
		}
		<-cs
		group.Done()
		return true
	}
	// engine.Log.Info("multicast to:%s %s", sid.B58String(), hex.EncodeToString(msg.Body.Hash))

	if ss, ok := engine.GetSession(utils.Bytes2string(sid)); ok {
		// err := ss.Send(version_multicast, msg.Head.JSON(), msg.Body.JSON(), false)
		mheadBs := msg.Head.Proto()
		mbodyBs := msg.Body.Proto()
		err := ss.Send(msgid, &mheadBs, &mbodyBs, false)
		if err != nil {
			// engine.Log.Warn("send multicast message fail %s", err.Error()) // fmt.Println("发送广播消息失败", err)
			// continue
		} else {
			// hashStr := hex.EncodeToString(message.Body.Hash)
			// engine.Log.Warn("广播消息 1111111111111 %s", hashStr)
			// bs := flood.WaitRequest(config.CLASS_wallet_broadcast_return, hex.EncodeToString(message.Body.Hash), int64(i+1))
			_, err := flood.WaitRequest(config.CLASS_wallet_broadcast_return, utils.Bytes2string(msg.Body.Hash), int64(4))
			if err != nil {
				// engine.Log.Warn("Timeout receiving broadcast reply message %s %s", sid.B58String(), hex.EncodeToString(msg.Body.Hash))
				// failNode = append(failNode, broadcasts[j])
				// continue
			} else {
				success = true
			}
			// engine.Log.Warn("收到广播回复消息  %s %s", addrStr, hashStr)
		}
	} else {
		engine.Log.Warn("get node conn fail :%s", sid.B58String())
	}
	<-cs
	group.Done()
	return
}

// /*
// 	注册广播消息
// */
// func Register_multicast_OneLevel(msgid uint64, handler MsgHandler) {
// 	router.Register(msgid, handler)
// }

// /*
// 	发送一个新的广播消息
// */
// func SendMulticastMsg_OneLevel(msgid uint64, content *[]byte) bool {
// 	// engine.Log.Info("发送广播 start")
// 	//TODO 这里查询缓存中的节点信息，拿到节点的超级节点信息
// 	head := NewMessageHead(nil, nil, false)
// 	body := NewMessageBody(msgid, content, 0, nil, 0)
// 	message := NewMessage(head, body)
// 	message.BuildHash()

// 	// headBs := message.Head.JSON()
// 	// bodyBs := message.Body.JSON()

// 	//先保存这个消息到缓存
// 	err := new(sqlite3_db.MessageCache).Add(message.Body.Hash, head.Proto(), body.Proto())
// 	if err != nil {
// 		engine.Log.Error(err.Error())
// 		return false
// 	}

// 	//先发送给超级节点
// 	// superNodes := nodeStore.GetLogicNodes()
// 	// superNodes = append(superNodes, nodeStore.GetNodesClient()...)

// 	// //再发送给内网节点
// 	// proxyNodes := nodeStore.GetProxyAll()

// 	// broadcastsAll(superNodes, proxyNodes, message)
// 	//先发送给超级节点
// 	whiltlistNodes := nodeStore.GetWhiltListNodes()

// 	superNodes := nodeStore.GetLogicNodes()
// 	superNodes = append(superNodes, nodeStore.GetNodesClient()...)

// 	//再发送给内网节点
// 	proxyNodes := nodeStore.GetProxyAll()

// 	BroadcastsAll(version_multicast, whiltlistNodes, superNodes, proxyNodes, &message.Body.Hash)
// 	// engine.Log.Info("发送广播 end")
// 	return true

// }

// func broadcastsAll_OneLevel(whiltlistNodes, superNodes, proxyNodes []nodeStore.AddressNet, hash *[]byte) {

// 	start := time.Now()
// 	//给已发送的节点放map里，避免重复发送
// 	allNodes := make(map[string]bool)

// 	//先发送给超级节点
// 	// superNodes := nodeStore.GetLogicNodes()
// 	//排除重复的地址
// 	// superNodes = nodeStore.RemoveDuplicateAddress(superNodes)
// 	cs := make(chan bool, config.CPUNUM)
// 	group := new(sync.WaitGroup)
// 	for i, _ := range whiltlistNodes {
// 		sessionid := whiltlistNodes[i]
// 		//不要发送给自己
// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, sessionid) {
// 			continue
// 		}
// 		_, ok := allNodes[utils.Bytes2string(sessionid)]
// 		if ok {
// 			engine.Log.Info("repeat node addr: %s", sessionid.B58String())
// 			continue
// 		}
// 		allNodes[utils.Bytes2string(sessionid)] = false
// 		cs <- false
// 		group.Add(1)
// 		//区块广播给节点
// 		// engine.Log.Info("multcast super node:%s", sessionid.B58String())
// 		go broadcastsOne(cs, sessionid, hash, group, true)
// 	}
// 	group.Wait()
// 	engine.Log.Info("multicast whilt list node time %s", time.Now().Sub(start))

// 	start = time.Now()
// 	for i, _ := range superNodes {
// 		sessionid := superNodes[i]
// 		//不要发送给自己
// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, sessionid) {
// 			continue
// 		}
// 		_, ok := allNodes[utils.Bytes2string(sessionid)]
// 		if ok {
// 			engine.Log.Info("repeat node addr: %s", sessionid.B58String())
// 			continue
// 		}
// 		allNodes[utils.Bytes2string(sessionid)] = false
// 		cs <- false
// 		group.Add(1)
// 		//区块广播给节点
// 		// engine.Log.Info("multcast super node:%s", sessionid.B58String())
// 		go broadcastsOne(cs, sessionid, hash, group, false)
// 	}
// 	group.Wait()
// 	engine.Log.Info("multicast super node time %s", time.Now().Sub(start))
// 	start = time.Now()

// 	//再发送给内网节点
// 	// proxyNodes := nodeStore.GetProxyAll()
// 	//排除重复的地址
// 	// proxyNodes = nodeStore.RemoveDuplicateAddress(proxyNodes)
// 	for i, _ := range proxyNodes {
// 		sessionid := proxyNodes[i]
// 		//不要发送给自己
// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, sessionid) {
// 			continue
// 		}
// 		_, ok := allNodes[utils.Bytes2string(sessionid)]
// 		if ok {
// 			engine.Log.Info("repeat node addr: %s", sessionid.B58String())
// 			continue
// 		}
// 		allNodes[utils.Bytes2string(sessionid)] = false
// 		cs <- false
// 		group.Add(1)
// 		//区块广播给节点
// 		go broadcastsOne(cs, sessionid, hash, group, false)
// 	}
// 	group.Wait()
// 	engine.Log.Info("multicast proxy node time %s", time.Now().Sub(start))
// }

// /*
// 	给一个节点发送广播消息
// 	@return    bool    白名单种节点是否都收到了消息
// */
// func broadcastsOne_OneLevel(cs chan bool, sid nodeStore.AddressNet, hash *[]byte, group *sync.WaitGroup, reply bool) (success bool) {
// 	// engine.Log.Info("multicast to %s", sid.B58String())
// 	//先查询对方的版本,看协议是否有升级
// 	// node := nodeStore.FindNode(&sid)
// 	// if node == nil {
// 	// 	<-cs
// 	// 	group.Done()
// 	// 	return
// 	// }
// 	// if nodeStore.NodeSelf.Version >= config.Version_1 && node.Version >= config.Version_1 {

// 	// }
// 	head := NewMessageHead(nil, nil, false)
// 	body := NewMessageBody(0, hash, 0, nil, 0) //广播采用同步机制后，不需要真实msgid，所以设置为0
// 	msg := NewMessage(head, body)
// 	msg.BuildHash()
// 	//如果是普通节点，则不需要对方回复
// 	if !reply {
// 		if ss, ok := engine.GetSession(utils.Bytes2string(sid)); ok {
// 			// ss.Send(version_multicast, msg.Head.JSON(), msg.Body.JSON(), false)
// 			mheadBs := msg.Head.Proto()
// 			mbodyBs := msg.Body.Proto()
// 			ss.Send(version_multicast, &mheadBs, &mbodyBs, false)
// 		} else {
// 			engine.Log.Warn("get node conn fail %s", sid.B58String())
// 		}
// 		<-cs
// 		group.Done()
// 		return true
// 	}
// 	engine.Log.Info("multicast to:%s %s", sid.B58String(), hex.EncodeToString(msg.Body.Hash))

// 	if ss, ok := engine.GetSession(utils.Bytes2string(sid)); ok {
// 		// err := ss.Send(version_multicast, msg.Head.JSON(), msg.Body.JSON(), false)
// 		mheadBs := msg.Head.Proto()
// 		mbodyBs := msg.Body.Proto()
// 		err := ss.Send(version_multicast, &mheadBs, &mbodyBs, false)
// 		if err != nil {
// 			engine.Log.Warn("send multicast message fail %s", err.Error()) // fmt.Println("发送广播消息失败", err)
// 			// continue
// 		} else {
// 			// hashStr := hex.EncodeToString(message.Body.Hash)
// 			// engine.Log.Warn("广播消息 1111111111111 %s", hashStr)
// 			// bs := flood.WaitRequest(config.CLASS_wallet_broadcast_return, hex.EncodeToString(message.Body.Hash), int64(i+1))
// 			_, err := flood.WaitRequest(config.CLASS_wallet_broadcast_return, utils.Bytes2string(msg.Body.Hash), int64(4))
// 			if err != nil {
// 				engine.Log.Warn("Timeout receiving broadcast reply message %s %s", sid.B58String(), hex.EncodeToString(msg.Body.Hash))
// 				// failNode = append(failNode, broadcasts[j])
// 				// continue
// 			} else {
// 				success = true
// 			}
// 			// engine.Log.Warn("收到广播回复消息  %s %s", addrStr, hashStr)
// 		}
// 	} else {
// 		engine.Log.Warn("get node conn fail :%s", sid.B58String())
// 	}
// 	<-cs
// 	group.Done()
// 	return
// }

/*
	注册消息
	从超级节点中搜索目标节点
*/
func Register_search_super(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送一个新的查找超级节点消息
*/
func SendSearchSuperMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*Message, bool) {
	mhead := NewMessageHead(recvid, recvid, false)
	mbody := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(mhead, mbody)
	message.BuildHash()
	return message, message.Send(version_search_super)
}

/*
	对某个消息回复
*/
func SendSearchSuperReplyMsg(message *Message, msgid uint64, content *[]byte) bool {
	return SendP2pReplyMsg(message, msgid, content)
}

/*
	注册消息
	从所有节点中搜索节点，包括普通节点
*/
func Register_search_all(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送一个新的查找超级节点消息
*/
func SendSearchAllMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*Message, bool) {
	mhead := NewMessageHead(recvid, recvid, false)
	mbody := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(mhead, mbody)
	message.BuildHash()
	return message, message.Send(version_search_all)
}

/*
	对某个消息回复
*/
func SendSearchAllReplyMsg(message *Message, msgid uint64, content *[]byte) bool {
	return SendP2pReplyMsg(message, msgid, content)
}

/*
	注册点对点通信消息
*/
func Register_p2p(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送一个新消息
	@return    *Message     返回的消息
	@return    bool         是否发送成功
	@return    bool         消息是发给自己
*/
func SendP2pMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*Message, bool, bool) {
	if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *recvid) {
		return nil, false, true
	}
	//TODO 这里查询缓存中的节点信息，拿到节点的超级节点信息
	// fmt.Println("发送未加密消息 1111111111111", msgid)
	head := NewMessageHead(recvid, recvid, true)
	body := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(head, body)
	message.BuildHash()
	ok := message.Send(version_p2p)
	// fmt.Println("发送未加密消息 2222222222222", msgid, ok)
	return message, ok, false
}

/*
	发送一个新消息
	是SendP2pMsg方法的定制版本，多了recvSuperId参数。
*/
func SendP2pMsgEX(msgid uint64, recvid, recvSuperId *nodeStore.AddressNet, content *[]byte) (*Message, bool) {
	//TODO 这里查询缓存中的节点信息，拿到节点的超级节点信息
	head := NewMessageHead(recvid, recvSuperId, true)
	body := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(head, body)
	message.BuildHash()
	ok := message.Send(version_p2p)
	return message, ok
}

/*
	对某个消息回复
*/
func SendP2pReplyMsg(message *Message, msgid uint64, content *[]byte) bool {
	mhead := NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	mbody := NewMessageBody(msgid, content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	newmessage := NewMessage(mhead, mbody)
	// message.BuildHash()
	newmessage.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	return newmessage.Reply(version_p2p)
}

/*
	注册点对点通信消息
*/
func Register_p2pHE(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

var securityStore = new(sync.Map)

/*
	发送一个加密消息，包括消息头也加密
	@return    *Message     返回的消息
	@return    bool         是否发送成功
	@return    bool         消息是发给自己
*/
func SendP2pMsgHE(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*Message, bool, bool) {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *recvid) {
		return nil, false, true
	}

	if content == nil {
		bs := []byte{0}
		content = &bs
	}
	//这里查询缓存中的节点信息，拿到节点的超级节点信息
	ratchet := sessionManager.GetSendRatchet(*recvid)
	v, ok := securityStore.Load(utils.Bytes2string(*recvid))
	if ratchet == nil || !ok {

		// fmt.Println("============ 发送加密消息 11111")
		//发送获取节点信息消息
		//获取对方身份公钥
		message, ok, _ := SendP2pMsg(config.MSGID_SearchAddr, recvid, nil)
		// bs := flood.WaitRequest(config.CLASS_security_searchAddr, hex.EncodeToString(message.Body.Hash), 0)
		bs, _ := flood.WaitRequest(config.CLASS_security_searchAddr, utils.Bytes2string(message.Body.Hash), 0)
		if bs == nil {
			// fmt.Println("============ 发送加密消息 22222")
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 33333333")
		sni, err := ParserSearchNodeInfo(*bs)
		if err != nil {
			// fmt.Println("============ 发送加密消息 4444444")
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 55555555")

		//生成3个密钥，通过共享密钥加密发送两个密钥
		// ivRand, err := crypto.Rand32Byte()
		// if err != nil {
		// 	return nil, false
		// }
		// ivDH, err := dh.GenerateKeyPair(ivRand)
		// if err != nil {
		// 	return nil, false
		// }
		// iv := dh.KeyExchange(dh.DHPair{PrivateKey: ivDH.GetPrivateKey(), PublicKey: sni.CPuk})

		//获取一个随机数
		sharedHkaRand, err := crypto.Rand32Byte()
		if err != nil {
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 666666666")
		//使用随机数生成一个密钥对
		sharedHkaDH, err := dh.GenerateKeyPair(sharedHkaRand[:])
		if err != nil {
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 7777777777")
		//生成一个共享密钥Hka
		sharedHka := dh.KeyExchange(dh.NewDHPair(sharedHkaDH.GetPrivateKey(), sni.CPuk))

		// fmt.Println("SKA prk",hex.EncodeToString())

		//获取一个随机数
		sharedNhkbRand, err := crypto.Rand32Byte()
		if err != nil {
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 88888888")
		//使用随机数生成一个密钥对
		sharedNhkbDH, err := dh.GenerateKeyPair(sharedNhkbRand[:])
		if err != nil {
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 99999")
		//使用自己的私钥和对方公钥生成共享密钥Nhkb
		sharedNhkb := dh.KeyExchange(dh.NewDHPair(sharedNhkbDH.GetPrivateKey(), sni.CPuk))

		//生成共享密钥SK
		sk := dh.KeyExchange(dh.NewDHPair(keystore.GetDHKeyPair().KeyPair.GetPrivateKey(), sni.CPuk))
		// cpuk := keystore.GetDHKeyPair().KeyPair.GetPublicKey()
		// cpukNet := nodeStore.NodeSelf.IdInfo.CPuk

		shareKey := ShareKey{
			Idinfo:   nodeStore.NodeSelf.IdInfo,
			A_DH_PUK: sharedHkaDH.GetPublicKey(),  //
			B_DH_PUK: sharedNhkbDH.GetPublicKey(), //B公钥
		}

		// *bs, err = json.Marshal(shareKey)

		*bs, err = shareKey.Proto() // json.Marshal(shareKey)
		if err != nil {
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 101010")
		//给对方发送自己的公钥，用于创建加密通道消息
		message, ok = SendP2pMsgEX(config.MSGID_security_create_pipe, sni.Id, sni.SuperId, bs)
		if !ok {
			return nil, false, false
		} else {
			// bs := flood.WaitRequest(config.CLASS_im_security_create_pipe, hex.EncodeToString(message.Body.Hash), 0)
			bs, _ := flood.WaitRequest(config.CLASS_im_security_create_pipe, utils.Bytes2string(message.Body.Hash), 0)
			if bs == nil {
				// fmt.Println("============ 发送加密消息 1010101022222")
				return nil, false, false
			}
		}
		// fmt.Println("============ 发送加密消息 111111112222222")
		err = sessionManager.AddSendPipe(*sni.Id, sk, sharedHka, sharedNhkb)
		if err != nil {
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 22222222221111111")
		//
		securityStore.Store(utils.Bytes2string(*recvid), sni)
	}
	// fmt.Println("============ 发送加密消息 22222")
	v, ok = securityStore.Load(utils.Bytes2string(*recvid))
	if !ok {
		// fmt.Println("============ 发送加密消息 333333")
		return nil, false, false
	}
	sni := v.(*SearchNodeInfo)
	// fmt.Println("============ 发送加密消息 444444")
	//开始发送真正的消息
	//先将内容加密
	ratchet = sessionManager.GetSendRatchet(*sni.Id)
	if ratchet == nil {
		// fmt.Println("============ 发送加密消息 5555555")
		return nil, false, false
	}
	// fmt.Println("============ 发送加密消息 6666666")

	msgHE := ratchet.RatchetEncrypt(*content, nil)
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.LittleEndian, uint64(len(msgHE.Header)))
	buf.Write(msgHE.Header)
	binary.Write(buf, binary.LittleEndian, uint64(len(msgHE.Ciphertext)))
	buf.Write(msgHE.Ciphertext)
	bs := buf.Bytes()
	// fmt.Println("head", hex.EncodeToString(msgHE.Header))
	// fmt.Println("body", hex.EncodeToString(msgHE.Ciphertext))

	head := NewMessageHead(sni.Id, sni.SuperId, true)
	body := NewMessageBody(msgid, &bs, 0, nil, 0)
	message := NewMessage(head, body)
	message.BuildHash()
	ok = message.Send(version_p2pHE)
	// fmt.Println("============ 发送加密消息 33333333333")
	return message, ok, false
}

/*
	对某个消息回复
*/
func SendP2pReplyMsgHE(message *Message, msgid uint64, content *[]byte) bool {
	mhead := NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	mbody := NewMessageBody(msgid, content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	newmessage := NewMessage(mhead, mbody)
	newmessage.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	return newmessage.Reply(version_p2p)
}

/*
	注册虚拟节点搜索节点消息
*/
func Register_vnode_search(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送虚拟节点搜索节点消息
*/
func SendVnodeSearchMsg(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, content *[]byte) (*Message, bool) {
	fmt.Println("------------------\n发送一个虚拟节点搜索节点消息")

	//直接查找最近的超级节点
	mhead := NewMessageHeadVnode(sendVnodeid, recvVnodeid, false)
	mbody := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(mhead, mbody)
	ok := message.Send(version_vnode_search)
	return message, ok
}

/*
	对发送虚拟节点搜索节点消息回复
*/
func SendVnodeSearchReplyMsg(message *Message, msgid uint64, content *[]byte, session engine.Session) error {
	return SendVnodeP2pReplyMsgHE(message, msgid, content, session)
}

/*
	注册虚拟节点之间点对点加密消息
*/
func Register_vnode_p2pHE(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送虚拟节点之间点对点消息
*/
func SendVnodeP2pMsgHE(msgid uint64, sendVnodeid, recvVnodeid *virtual_node.AddressNetExtend, content *[]byte) (*Message, bool) {
	fmt.Println("------------------\n发送一个虚拟节点点对点加密消息")

	//直接查找最近的超级节点
	mhead := NewMessageHeadVnode(sendVnodeid, recvVnodeid, true)
	mbody := NewMessageBody(msgid, content, 0, nil, 0)
	message := NewMessage(mhead, mbody)
	ok := message.Send(version_vnode_p2pHE)
	return message, ok
}

/*
	对发送虚拟节点之间点对点消息回复
*/
func SendVnodeP2pReplyMsgHE(message *Message, msgid uint64, content *[]byte, session engine.Session) error {
	head := NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	body := NewMessageBody(msgid, content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	newmessage := NewMessage(head, body)
	newmessage.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// return session.Send(version_vnode_p2pHE, head.JSON(), body.JSON(), false)

	mheadBs := head.Proto()
	mbodyBs := body.Proto()
	return session.Send(version_vnode_p2pHE, &mheadBs, &mbodyBs, false)
}
