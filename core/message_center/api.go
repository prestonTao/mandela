package message_center

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils/crypto"
	"mandela/core/utils/crypto/dh"
	"mandela/core/virtual_node"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
)

var router = NewRouter()

/*
	注册邻居节点消息，不转发
*/
func Register_neighbor(msgid uint64, handler MsgHandler) {
	router.Register(msgid, handler)
}

/*
	发送一个新邻居节点消息
*/
func SendNeighborMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) (*Message, bool) {
	//直接查找最近的超级节点
	mhead := NewMessageHead(recvid, recvid, false)
	mbody := NewMessageBody(msgid, content, "", nil, 0)
	message := NewMessage(mhead, mbody)
	message.BuildHash()
	session, ok := engine.GetSession(recvid.B58String())
	if ok {
		//				fmt.Println("给这个session发送消息成功", key.B58String())
		session.Send(version_neighbor, mhead.JSON(), mbody.JSON(), false)
	}
	return message, ok
}

/*
	对某个消息回复
*/
func SendNeighborReplyMsg(message *Message, msgid uint64, content *[]byte, session engine.Session) error {
	head := NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	// engine.Log.Info("看看是哪个指针为空 %v", message.Body)
	body := NewMessageBody(msgid, content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	newmessage := NewMessage(head, body)
	newmessage.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	return session.Send(version_neighbor, head.JSON(), body.JSON(), false)
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
func SendMulticastMsg(msgid uint64, content *[]byte) bool {
	//TODO 这里查询缓存中的节点信息，拿到节点的超级节点信息
	head := NewMessageHead(nil, nil, false)
	body := NewMessageBody(msgid, content, "", nil, 0)
	message := NewMessage(head, body)
	message.BuildHash()

	failNode := append(nodeStore.GetLogicNodes(), nodeStore.GetProxyAll()...)
	failNode = append(failNode, nodeStore.GetOtherNodes()...)
	broadcasts := make([]*nodeStore.AddressNet, 0)
	//广播失败的节点最多广播5次
	for i := 0; i < 5; i++ {
		if len(failNode) <= 0 {
			break
		}
		// engine.Log.Info("第 %d 轮发送广播", i)
		broadcasts = failNode
		failNode = make([]*nodeStore.AddressNet, 0)

		// lock := new(sync.RWMutex)
		for j, _ := range broadcasts {
			//区块广播给节点
			engine.Log.Warn("Block broadcast to node %s", broadcasts[j].B58String())

			//异步广播消息
			// go func() {
			// 	for i := 0; i < 5; i++ {
			// lock.Lock()
			// defer lock.Unlock()
			// for _, one := range append(nodeStore.GetAllNodes(), nodeStore.GetProxyAll()...) {
			// 	// fmt.Println("区块广播给节点", one.B58String())
			// >>>>>>> im
			if ss, ok := engine.GetSession(broadcasts[j].B58String()); ok {
				err := ss.Send(version_multicast, head.JSON(), body.JSON(), false)
				if err != nil {
					engine.Log.Warn("send multicast message fail %s", err.Error()) // fmt.Println("发送广播消息失败", err)
					// lock.Unlock()
					continue
				} else {
					// fmt.Println("发送广播消息成功!")
					bs := flood.WaitRequest(config.CLASS_wallet_broadcast_return, hex.EncodeToString(message.Body.Hash), int64(i+1))
					if bs == nil {
						engine.Log.Warn("Timeout receiving broadcast reply message %s", broadcasts[j].B58String())
						// lock.Unlock()
						failNode = append(failNode, broadcasts[j])
						continue
					}
					// engine.Log.Warn("收到广播回复消息  %s", broadcasts[j].B58String())
					// lock.Unlock()
					// break
				}
			} else {
				engine.Log.Warn("get node conn fail")
				// fmt.Println("获取节点连接失败")
				// engine.Log.Warn("获取节点连接失败")
				// lock.Unlock()
				// break
			}
			// }
			// }()

		}
	}

	// for _, one := range nodeStore.GetAllNodes() {
	// 	fmt.Println("区块广播给节点", one.B58String())
	// 	if ss, ok := engine.GetSession(one.B58String()); ok {
	// 		err := ss.Send(version_multicast, head.JSON(), body.JSON(), false)
	// 		if err != nil {
	// 			fmt.Println("发送广播消息失败", err)
	// 		} else {
	// 			fmt.Println("发送广播消息成功!")
	// 		}
	// 	} else {
	// 		fmt.Println("获取节点连接失败")
	// 	}

	// }
	return true
}

func broadcastsOne() {

}

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
	mbody := NewMessageBody(msgid, content, "", nil, 0)
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
	mbody := NewMessageBody(msgid, content, "", nil, 0)
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
	body := NewMessageBody(msgid, content, "", nil, 0)
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
	body := NewMessageBody(msgid, content, "", nil, 0)
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

	if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *recvid) {
		return nil, false, true
	}

	if content == nil {
		bs := []byte{0}
		content = &bs
	}
	//这里查询缓存中的节点信息，拿到节点的超级节点信息
	ratchet := sessionManager.GetSendRatchet(*recvid)
	v, ok := securityStore.Load(recvid.B58String())
	if ratchet == nil || !ok {

		// fmt.Println("============ 发送加密消息 11111")
		//发送获取节点信息消息
		//获取对方身份公钥
		message, ok, _ := SendP2pMsg(config.MSGID_SearchAddr, recvid, nil)
		bs := flood.WaitRequest(config.CLASS_security_searchAddr, hex.EncodeToString(message.Body.Hash), 0)
		if bs == nil {
			// fmt.Println("============ 发送加密消息 22222")
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 33333333")
		sni, err := ParserSearchNodeInfo(bs)
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

		*bs, err = json.Marshal(shareKey)
		if err != nil {
			return nil, false, false
		}
		// fmt.Println("============ 发送加密消息 101010")
		//给对方发送自己的公钥，用于创建加密通道消息
		message, ok = SendP2pMsgEX(config.MSGID_security_create_pipe, sni.Id, sni.SuperId, bs)
		if !ok {
			return nil, false, false
		} else {
			bs := flood.WaitRequest(config.CLASS_im_security_create_pipe, hex.EncodeToString(message.Body.Hash), 0)
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
		securityStore.Store(recvid.B58String(), sni)
	}
	// fmt.Println("============ 发送加密消息 22222")
	v, ok = securityStore.Load(recvid.B58String())
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
	body := NewMessageBody(msgid, &bs, "", nil, 0)
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
	mbody := NewMessageBody(msgid, content, "", nil, 0)
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
	mbody := NewMessageBody(msgid, content, "", nil, 0)
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
	return session.Send(version_vnode_p2pHE, head.JSON(), body.JSON(), false)
}
