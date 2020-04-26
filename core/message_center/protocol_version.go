package message_center

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center/security_signal/doubleratchet"
	"mandela/core/nodeStore"
	"encoding/binary"
	"fmt"
)

//消息协议版本号
const (
	version_neighbor     = 1 //邻居节点消息，本消息不转发
	version_multicast    = 2 //广播消息
	version_search_super = 3 //搜索节点消息
	version_search_all   = 4 //搜索节点消息
	version_p2p          = 5 //点对点消息
	version_p2pHE        = 6 //点对点可靠传输加密消息
	version_vnode_search = 7 //搜索虚拟节点消息
	version_vnode_p2pHE  = 8 //虚拟节点间点对点可靠传输加密消息
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

}

type MsgHandler func(c engine.Controller, msg engine.Packet, message *Message)

/*
	邻居节点消息控制器，不转发消息
*/
func NeighborHandler(c engine.Controller, msg engine.Packet) {
	//	fmt.Println("接收到查询最近超级节点请求")
	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}

	//解析包体内容
	if err := message.ParserContent(); err != nil {
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

	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		// fmt.Println(err)
		// engine.Log.Info("这个广播消息头解析失败")
		return
	}
	//解析包体内容
	if err := message.ParserContent(); err != nil {
		// fmt.Println(err)
		// engine.Log.Info("这个广播消息内容解析失败")
		return
	}
	//第一时间回复广播消息
	go SendNeighborReplyMsg(message, config.MSGID_multicast_return, nil, msg.Session)

	// engine.Log.Info("广播消息hash %s", hex.EncodeToString(message.Body.Hash))

	//判断重复的广播
	if !message.CheckSendhash() {
		// engine.Log.Info("This message is repeated")
		// engine.Log.Info("有重复的广播")
		return
	}
	// engine.Log.Info("这个广播不重复")

	//继续广播给其他节点
	// engine.Log.Info("自己是否是超级节点 %v", nodeStore.NodeSelf.IsSuper)
	if nodeStore.NodeSelf.IsSuper {
		//广播给其他超级节点
		//		mh := utils.Multihash(*message.Body.Content)
		ids := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
		// engine.Log.Info("广播给节点地址列表 %v", ids)
		for _, one := range ids {
			// engine.Log.Info("广播发送给 %s", one.B58String())
			if ss, ok := engine.GetSession(one.B58String()); ok {
				ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
			}
		}

		//广播给代理对象
		pids := nodeStore.GetProxyAll()
		for _, one := range pids {
			if ss, ok := engine.GetSession(one.B58String()); ok {
				//				ss.Send(MSGID_multicast_online_recv, &msg.Data, false)
				ss.Send(msg.MsgID, &msg.Data, &msg.Dataplus, false)
			}
		}

	}

	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		engine.Log.Info("This broadcast message is not registered:", message.Body.MessageId)
		return
	}
	// engine.Log.Info("有广播消息，消息编号 %d", message.Body.MessageId)
	h(c, msg, message)

}

/*
	从超级节点中搜索目标节点消息控制器
*/
func searchSuperHandler(c engine.Controller, msg engine.Packet) {
	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}
	form := nodeStore.AddressFromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	if err := message.ParserContent(); err != nil {
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
	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}
	form := nodeStore.AddressFromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	if err := message.ParserContent(); err != nil {
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
	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		// fmt.Println("收到点对点控制消息 1111", err)
		return
	}

	form := nodeStore.AddressFromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		// fmt.Println("收到点对点控制消息 发送给其他人")
		return
	}
	//解析包体内容
	if err := message.ParserContent(); err != nil {
		// fmt.Println("收到点对点控制消息 22222222", err)
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
	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}
	form := nodeStore.AddressFromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	if err := message.ParserContent(); err != nil {
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
	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}
	form := nodeStore.AddressFromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	if err := message.ParserContent(); err != nil {
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
	message, err := ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		return
	}
	form := nodeStore.AddressFromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//解析包体内容
	if err := message.ParserContent(); err != nil {
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
