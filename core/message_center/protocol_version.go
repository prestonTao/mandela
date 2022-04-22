package message_center

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center/flood"
	"mandela/core/message_center/security_signal/doubleratchet"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/protos/go_protos"
	"mandela/sqlite3_db"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

//消息协议版本号
const (
	version_neighbor            = 1  //邻居节点消息，本消息不转发
	version_multicast           = 2  //广播消息
	version_search_super        = 3  //搜索节点消息
	version_search_all          = 4  //搜索节点消息
	version_p2p                 = 5  //点对点消息
	version_p2pHE               = 6  //点对点可靠传输加密消息
	version_vnode_search        = 7  //搜索虚拟节点消息
	version_vnode_p2pHE         = 8  //虚拟节点间点对点可靠传输加密消息
	version_multicast_sync      = 9  //查询邻居节点的广播消息
	version_multicast_sync_recv = 10 //查询广播消息的回应
)

// var do = new(sync.Once)

func RegisterMsgVersion() {
	engine.RegisterMsg(version_neighbor, NeighborHandler)
	engine.RegisterMsg(version_multicast, multicastHandler)
	engine.RegisterMsg(version_search_super, searchSuperHandler)
	engine.RegisterMsg(version_search_all, searchAllHandler)
	engine.RegisterMsg(version_p2p, p2pHandler)
	engine.RegisterMsg(version_p2pHE, p2pHEHandler)
	engine.RegisterMsg(version_vnode_search, vnodeSearchHandler)
	engine.RegisterMsg(version_vnode_p2pHE, vnodeP2pHEHandler)
	engine.RegisterMsg(version_multicast_sync, multicastSyncHandler)
	engine.RegisterMsg(version_multicast_sync_recv, multicastSyncRecvHandler)

}

type MsgHandler func(c engine.Controller, msg engine.Packet, message *Message)

/*
	查询邻居节点的广播消息
*/
func multicastSyncHandler(c engine.Controller, msg engine.Packet) {
	//	fmt.Println("接收到查询最近超级节点请求")
	fromAddr := nodeStore.AddressNet(msg.Session.GetName())
	is := false
	if nodeStore.FindWhiteList(&fromAddr) {
		is = true
	}
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		if is {
			engine.Log.Info("get multicast message error from:%s %s", fromAddr.B58String(), err.Error())
		}
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		if is {
			engine.Log.Info("get multicast message error from:%s %s", fromAddr.B58String(), err.Error())
		}
		return
	}

	messageCache, err := new(sqlite3_db.MessageCache).FindByHash(*message.Body.Content) //.Add(message.KeyDB(), headBs, bodyBs)
	if err != nil {
		if is {
			engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		}
		engine.Log.Error("get multicast message :%s error:%s", hex.EncodeToString(*message.Body.Content), err.Error())
		return
	}

	mmp := go_protos.MessageMulticast{
		Head: messageCache.Head,
		Body: messageCache.Body,
	}

	content, err := mmp.Marshal()
	if err != nil {
		if is {
			engine.Log.Info("get multicast message error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		}
		engine.Log.Error(err.Error())
		return
	}

	head := NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	body := NewMessageBody(0, &content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	newmessage := NewMessage(head, body)
	newmessage.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)

	mheadBs := head.Proto()
	mbodyBs := body.Proto()
	err = msg.Session.Send(version_multicast_sync_recv, &mheadBs, &mbodyBs, false)
	if err != nil {
		if is {
			engine.Log.Info("get multicast message success error from:%s %s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content), err.Error())
		}
	} else {
		if is {
			// engine.Log.Info("get multicast message success from:%s %s", fromAddr.B58String(), hex.EncodeToString(*message.Body.Content))
		}
	}
	// SendNeighborReplyMsg(message, config.MSGID_multicast_return, nil, msg.Session)
	// //
	// msg.Session.Send(version_multicast_sync_recv, messageCache.Head, messageCache.Body, false)
}

func multicastSyncRecvHandler(c engine.Controller, msg engine.Packet) {
	//	fmt.Println("接收到查询最近超级节点请求")
	fromAddr := nodeStore.AddressNet(msg.Session.GetName())
	is := false
	if nodeStore.FindWhiteList(&fromAddr) {
		is = true
		engine.Log.Info("recv multicast message content from:%s", fromAddr.B58String())
	}

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		if is {
			engine.Log.Info("recv multicast message content error from:%s", fromAddr.B58String(), err.Error())
		}
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		if is {
			engine.Log.Info("recv multicast message content error from:%s", fromAddr.B58String(), err.Error())
		}
		return
	}

	//自己处理
	flood.ResponseWait(config.CLASS_engine_multicast_sync, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	邻居节点消息控制器，不转发消息
*/
func NeighborHandler(c engine.Controller, msg engine.Packet) {
	//	fmt.Println("接收到查询最近超级节点请求")
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }

	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		return
	}
	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This neighbor message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
	广播消息控制器
*/
func multicastHandler(c engine.Controller, msg engine.Packet) {
	// engine.Log.Debug("收到广播消息")

	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	//广播消息头解析失败
	// 	engine.Log.Warn("Parsing of this broadcast header failed")
	// 	return
	// }
	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	engine.Log.Info("Content parsing of this broadcast message failed %s", err.Error())
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		//广播消息头解析失败
		engine.Log.Warn("Parsing of this broadcast header failed")
		return
	}
	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		engine.Log.Info("Content parsing of this broadcast message failed %s", err.Error())
		return
	}
	//第一时间回复广播消息
	// go SendNeighborReplyMsg(message, config.MSGID_multicast_return, nil, msg.Session)

	// engine.Log.Info("广播消息hash %s", hex.EncodeToString(message.Body.Hash))

	//判断重复的广播
	// if !message.CheckSendhash() {
	// 	engine.Log.Info("This multicast message is repeated")
	// 	// engine.Log.Info("有重复的广播")
	// 	return
	// }
	// engine.Log.Info("这个广播不重复")

	//先同步，同步了再广播给其他节点
	mhOne := MsgHolderOne{
		MsgHash: *message.Body.Content,
		Addr:    nodeStore.AddressNet([]byte(msg.Session.GetName())), // *message.Head.SenderSuperId, // nodeStore.AddressNet
		Message: message,
		Session: msg.Session,
	}
	select {
	case msgchannl <- &mhOne:
	default:
		engine.Log.Error("msgchannl is full")
	}

	return

	// //继续广播给其他节点
	// // engine.Log.Info("自己是否是超级节点 %v", nodeStore.NodeSelf.IsSuper)
	// if nodeStore.NodeSelf.IsSuper {
	// 	//广播给其他超级节点
	// 	go func() {
	// 		//先发送给超级节点
	// 		superNodes := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
	// 		//广播给代理对象
	// 		proxyNodes := nodeStore.GetProxyAll()
	// 		broadcastsAll(superNodes, proxyNodes, message)

	// 		return
	// 	}()
	// }

	// //自己处理
	// h := router.GetHandler(message.Body.MessageId)
	// if h == nil {
	// 	engine.Log.Info("This broadcast message is not registered:", message.Body.MessageId)
	// 	return
	// }
	// // engine.Log.Info("有广播消息，消息编号 %d", message.Body.MessageId)
	// h(c, msg, message)

}

/*
	从超级节点中搜索目标节点消息控制器
*/
func searchSuperHandler(c engine.Controller, msg engine.Packet) {
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// // form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	if message.IsSendOther(&form) {
		return
	}

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}

	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This searchsuper message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
	从所有节点中搜索目标节点消息控制器
*/
func searchAllHandler(c engine.Controller, msg engine.Packet) {
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// // form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// form := nodeStore.AddressNet(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	form := nodeStore.AddressNet(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}

	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This searchAll message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
	点对点消息控制器
*/
func p2pHandler(c engine.Controller, msg engine.Packet) {
	// fmt.Println("head", string(msg.Data), string(msg.Dataplus))
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println("收到点对点控制消息 1111", err)
	// 	return
	// }

	// // form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.IsSendOther(&form) {
	// 	// fmt.Println("收到点对点控制消息 发送给其他人")
	// 	return
	// }
	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println("收到点对点控制消息 22222222", err)
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	// fmt.Println("head", string(msg.Data), string(msg.Dataplus))
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println("收到点对点控制消息 1111", err)
	// 	return
	// }

	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	if message.IsSendOther(&form) {
		// fmt.Println("收到点对点控制消息 发送给其他人")
		return
	}
	//解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println("收到点对点控制消息 22222222", err)
	// 	return
	// }

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}
	// fmt.Println("==========\n", message.Body.MessageId, message.Head)
	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This p2p message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
	点对点消息控制器
*/
func p2pHEHandler(c engine.Controller, msg engine.Packet) {
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// // form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}

	headSize := binary.LittleEndian.Uint64((*message.Body.Content)[:8])
	headData := (*message.Body.Content)[8 : headSize+8]
	// bodySize := binary.LittleEndian.Uint64((*message.Body.Content)[headSize+8 : headSize+8+8])
	bodyData := (*message.Body.Content)[headSize+8+8:]

	msgHE := doubleratchet.MessageHE{
		Header:     headData,
		Ciphertext: bodyData,
	}
	// fmt.Println("head", hex.EncodeToString(msgHE.Header))
	// fmt.Println("body", hex.EncodeToString(msgHE.Ciphertext))

	sessionHE := sessionManager.GetRecvRatchet(*message.Head.Sender)
	if sessionHE == nil {
		// fmt.Println("双棘轮接收key未找到")
		//双棘轮key未找到
		go SendP2pMsgHE(config.MSGID_security_pipe_error, message.Head.Sender, nil)
		return
	}
	bs, err := sessionHE.RatchetDecrypt(msgHE, nil)
	if err != nil {
		// fmt.Println("解密消息出错", err)
		return
	}
	*message.Body.Content = bs
	// fmt.Println("开始解密消息 33333333333333")
	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This p2pHE message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
	// fmt.Println("开始解密消息 44444444444444444")
}

/*
	从所有虚拟节点中搜索目标节点消息控制器
*/
func vnodeSearchHandler(c engine.Controller, msg engine.Packet) {
	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}

	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This searchAll message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}

/*
	虚拟节点间点对点可靠传输加密消息
*/
func vnodeP2pHEHandler(c engine.Controller, msg engine.Packet) {
	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// // form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	message, err := ParserMessageProto(msg.Data, msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	// message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	form := nodeStore.AddressNet([]byte(msg.Session.GetName()))
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	//解析包体内容
	if err = message.ParserContentProto(); err != nil {
		// fmt.Println(err)
		return
	}

	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		fmt.Println("This searchAll message is not registered:", message.Body.MessageId)
		return
	}
	h(c, msg, message)
}
