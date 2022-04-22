package core

import (
	"mandela/config"
	gconfig "mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/virtual_node/manager"
	"mandela/protos/go_protos"
	"mandela/sqlite3_db"
	"bytes"
	"strconv"
	"sync"
	"time"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	//	FindNodeNum    = iota + 101 //查找结点服务id
	//	SendMessageNum              //发送消息服务id

	//	SaveKeyValueReqNum
	//	SaveKeyValueRspNum
	MsgTextId  = 1 //文本消息
	MsgEmotId  = 2 //表情包消息
	MsgPicId   = 3 //图片消息
	MsgFileId  = 4 //文件消息
	MsgVedioId = 5 //视频消息
	MsgLinkId  = 6 //链接消息
	MsgPayId   = 7 //付款消息
)

var MsgChannl = make(chan *MessageVO, 10000)

type MessageVO struct {
	Name     string //消息记录name
	Id       string //发送消息者id
	Index    int64  //unix时间排序
	Time     string //接收时间
	Content  string //消息内容
	Path     string //图片、文件路径
	FileName string //文件名
	Size     int64  //文件大小
	Cate     int    //消息类型
	DBID     int64  //数据量id
}

func init() {
	//TODO 正式发布将这个模拟函数去掉
	//模拟每10秒钟收到一个消息
	//	go func() {

	//		for {
	//			time.Sleep(time.Second * 10)
	//			now := time.Now()
	//			msgOne := MessageVO{
	//				Name:    "haha",
	//				Id:      "123456789",
	//				Index:   now.Unix(),
	//				Time:    utils.FormatTimeToSecond(now),
	//				Content: "nihaoa",
	//			}
	//			MsgChannl <- &msgOne
	//		}
	//	}()
}

func RegisterCoreMsg() {
	message_center.RegisterMsgVersion()

	message_center.Register_search_super(gconfig.MSGID_checkNodeOnline, findSuperID)             //检查节点是否在线
	message_center.Register_p2p(gconfig.MSGID_checkNodeOnline_recv, findSuperID_recv)            //检查节点是否在线_返回
	message_center.Register_neighbor(gconfig.MSGID_getNearSuperIP, GetNearSuperAddr)             //从邻居节点得到自己的逻辑节点
	message_center.Register_neighbor(gconfig.MSGID_getNearSuperIP_recv, GetNearSuperAddr_recv)   //从邻居节点得到自己的逻辑节点_返回
	message_center.Register_multicast(gconfig.MSGID_multicast_online_recv, MulticastOnline_recv) //接收节点上线广播
	message_center.Register_neighbor(gconfig.MSGID_ask_close_conn_recv, AskCloseConn_recv)       //询问关闭连接
	message_center.Register_p2pHE(gconfig.MSGID_TextMsg, TextMsg)                                //接收文本消息
	message_center.Register_p2pHE(gconfig.MSGID_TextMsg_recv, TextMsg_recv)                      //接收文本消息_返回

	message_center.Register_p2p(gconfig.MSGID_SearchAddr, message_center.SearchAddress)                  //搜索节点，返回节点真实地址
	message_center.Register_p2p(gconfig.MSGID_SearchAddr_recv, message_center.SearchAddress_recv)        //搜索节点，返回节点真实地址_返回
	message_center.Register_p2p(gconfig.MSGID_security_create_pipe, message_center.CreatePipe)           //创建加密通道
	message_center.Register_p2p(gconfig.MSGID_security_create_pipe_recv, message_center.CreatePipe_recv) //创建加密通道_返回
	message_center.Register_p2pHE(gconfig.MSGID_security_pipe_error, message_center.Pipe_error)          //解密错误

	message_center.Register_p2p(gconfig.MSGID_search_node, SearchNode)           //查询一个节点是否在线
	message_center.Register_p2p(gconfig.MSGID_search_node_recv, SearchNode_recv) //查询一个节点是否在线_返回

}

/*
	查询一个节点是否在线
*/
func SearchNode(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	if !message.CheckSendhash() {
		return
	}

	//回复消息
	data := utils.Uint64ToBytes(uint64(nodeStore.NodeSelf.MachineID))

	message_center.SendP2pReplyMsg(message, config.MSGID_search_node_recv, &data)

}

func SearchNode_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	if !message.CheckSendhash() {
		return
	}
	// mid := utils.BytesToInt64(*message.Body.Content)
	// flood.ResponseWait(config.CLASS_get_MachineID, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(config.CLASS_get_MachineID, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

// /*
// 	广播节点上线
// */
// func MulticastOnline() {
// 	//间隔一分钟广播一次，广播5次
// 	for i := 0; i < 5; i++ {
// 		if nodeStore.NodeSelf.IsSuper {
// 			//			fmt.Println("开始广播")
// 			content := []byte(nodeStore.NodeSelf.IdInfo.Id)
// 			SendMulticastMsg(gconfig.MSGID_multicast_online_recv, &content)

// 			// head := NewMessageHead(nil, nil, false)
// 			// content := []byte(nodeStore.NodeSelf.IdInfo.Id)
// 			// body := NewMessageBody(&content, "", nil, 0)
// 			// message := NewMessage(head, body)
// 			// message.BuildHash()

// 			// //			message := &Message{
// 			// //				//			ReceSuperId:   ,                             //接收者的超级节点id
// 			// //				CreateTime:    utils.TimeFormatToNanosecond(), //消息创建时间unix
// 			// //				SenderSuperId: nodeStore.NodeSelf.IdInfo.Id,   //发送者超级节点id
// 			// //				//				ReplyHash:     "",                                    //回复消息的hash
// 			// //				Accurate: false,                         //是否准确发送给一个节点
// 			// //				Content:  *nodeStore.NodeSelf.IdInfo.Id, //发送的内容
// 			// //			}
// 			// //			data := message.JSON()
// 			// //广播给其他节点
// 			// //		ids := nodeStore.GetIdsForFar(message.Content)
// 			// for _, one := range nodeStore.GetAllNodes() {
// 			// 	//				fmt.Println()
// 			// 	if ss, ok := engine.GetSession(one.B58String()); ok {
// 			// 		ss.Send(gconfig.MSGID_multicast_online_recv, head.JSON(), body.JSON(), false)
// 			// 	}
// 			// }
// 		} else {
// 			//非超级节点不需要广播
// 			break
// 		}
// 		time.Sleep(time.Second * config.Time_Multicast_online)
// 	}

// }

/*
	接收上线的广播
*/
func MulticastOnline_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("接收到有节点上线的广播")

	newNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	// newNode := new(nodeStore.Node)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(newNode)
	if err != nil {
		//		fmt.Println("解析失败", err)
		return
	}

	// engine.Log.Info("检查这个节点是否需要 %s",)
	// engine.Log.Info("GetNearSuperAddr_recv check need: %s", newNode.IdInfo.Id.B58String())

	manager.AddNewNode(newNode.IdInfo.Id)

	//查询是否已经有这个连接了，有了就不连接
	//避免地址和node真实地址不对应的情况
	session := engine.GetSessionByHost(newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort)))
	if session != nil {
		// name := session.GetName()
		// addrOne := nodeStore.AddressNet([]byte(session.GetName()))
		// engine.Log.Info("GetNearSuperAddr_recv conn exist: %s", addrOne.B58String())
		return
	}

	//		if newNode.IdInfo.Id.B58String() == "5duDDfkY1tChLKGxbAtPdPysp9ghYn" || newNode.IdInfo.Id.B58String() == "5dqqnW3YTxTw9EESzT63qM63zf9BYj" {
	//		}
	//		if message.Head.Sender.B58String() == "5dsEMMhVbww4hUXV6VzaeRfHKv1nhh" {
	//			fmt.Println("查找结果", newNode.IdInfo.Id.B58String())
	//		}
	//		fmt.Println("查找结果", newNode.IdInfo.Id.B58String(), newNode.Addr, newNode.TcpPort)
	//TODO 当真实P2P网络地址和ip端口不对应的情况，容易漏连接某些节点
	//检查是否需要这个逻辑节点
	ok := nodeStore.CheckNeedNode(&newNode.IdInfo.Id)
	if !ok {
		//		fmt.Println("不需要这个逻辑节点")
		return
	}
	//	fmt.Println("需要这个节点")
	// engine.Log.Info("GetNearSuperAddr_recv need: %s", newNode.IdInfo.Id.B58String())
	//检查是否有这个连接
	// _, ok = engine.GetSession(newNode.IdInfo.Id.B58String())
	_, ok = engine.GetSession(utils.Bytes2string(newNode.IdInfo.Id))
	if !ok {
		//查询是否已经有这个连接了，有了就不连接
		//避免地址和node真实地址不对应的情况
		// session := engine.GetSessionByHost(newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort)))
		// if session != nil {
		// 	// name := session.GetName()
		// 	addrOne := nodeStore.AddressNet([]byte(session.GetName()))
		// 	engine.Log.Info("GetNearSuperAddr_recv conn exist: %s", addrOne.B58String())
		// 	continue
		// }
		//		fmt.Println("没有这个连接", message.Hash)
		_, err := engine.AddClientConn(newNode.Addr, uint32(newNode.TcpPort), false)
		if err != nil {
			//			fmt.Println("连接失败", err)
			// engine.Log.Info("GetNearSuperAddr_recv conn fail: %s %s", newNode.IdInfo.Id.B58String(), err.Error())
			return
		}
		// if session.GetName() != newNode
		// session.GetName()
		// nodeStore.FindNode()
	} else {
		//		fmt.Println("有连接", message.Hash)
	}
	// engine.Log.Info("add super nodeid: %s %+v", newNode.IdInfo.Id.B58String(), newNode)
	// nodeStore.AddNode(newNode)

	//非超级节点判断超级节点是否改变
	if !nodeStore.NodeSelf.IsSuper {
		nearId := nodeStore.FindNearInSuper(&nodeStore.NodeSelf.IdInfo.Id, nil, false)
		//		nearIdStr := hex.EncodeToString(nearId)
		// fmt.Println("判断是否需要替换超级节点", nearId.B58String(), nodeStore.SuperPeerId.B58String())
		if bytes.Equal(*nearId, *nodeStore.SuperPeerId) {
			return
		}
		nodeStore.SuperPeerId = nearId
		//		nodeStore.SuperPeerIdStr = hex.EncodeToString(nearId)
		// fmt.Println("超级节点换为:", nodeStore.SuperPeerId.B58String())
	}

}

/*
	查询一个id最近的超级节点id
*/
func findSuperID(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// data, _ := json.Marshal(nodeStore.NodeSelf)
	data, _ := nodeStore.NodeSelf.Proto()
	message_center.SendSearchSuperReplyMsg(message, gconfig.MSGID_checkNodeOnline_recv, &data)

}

func findSuperID_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	newNode, err := nodeStore.ParseNodeProto(*message.Body.Content)
	// newNode := new(nodeStore.Node)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(&newNode)
	// if err := json.Unmarshal(*message.Body.Content, &newNode); err != nil {
	if err != nil {
		//		fmt.Println("解析失败", err)
		return
	}

	node := nodeStore.FindNode(&newNode.IdInfo.Id)
	node.FlashOnlineTime()

}

/*
	获取相邻节点的超级节点地址
*/
func GetNearSuperAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	nodes := make([]nodeStore.Node, 0)
	ns := nodeStore.GetLogicNodes()
	ns = append(ns, nodeStore.GetNodesClient()...)
	idsm := nodeStore.NewIds(*message.Head.Sender, nodeStore.NodeIdLevel)
	for _, one := range ns {

		if bytes.Equal(*message.Head.Sender, one) {
			continue
		}
		idsm.AddId(one)
	}

	ids := idsm.GetIds()
	for _, one := range ids {
		//		if message.Head.Sender.B58String() == "5dtPgE32MoURWep7QEViBkZh5iVLDZ" {
		//			fmt.Println("查询到的节点", one.B58String())
		//		}
		addrNet := nodeStore.AddressNet(one)
		node := nodeStore.FindNode(&addrNet)
		if node != nil {
			nodes = append(nodes, *node)
		} else {
			// fmt.Println("这个节点为空")
		}
	}
	// data, _ := json.Marshal(nodes)
	nodeRepeated := go_protos.NodeRepeated{
		Nodes: make([]*go_protos.Node, 0),
	}
	for _, one := range nodes {
		idinfo := go_protos.IdInfo{
			Id:   one.IdInfo.Id,
			EPuk: one.IdInfo.EPuk,
			CPuk: one.IdInfo.CPuk[:],
			V:    one.IdInfo.V,
			Sign: one.IdInfo.Sign,
		}
		nodeOne := go_protos.Node{
			IdInfo:    &idinfo,
			IsSuper:   one.IsSuper,
			Addr:      one.Addr,
			TcpPort:   uint32(one.TcpPort),
			IsApp:     one.IsApp,
			MachineID: one.MachineID,
			Version:   one.Version,
		}
		nodeRepeated.Nodes = append(nodeRepeated.Nodes, &nodeOne)
	}

	data, _ := nodeRepeated.Marshal()

	message_center.SendNeighborReplyMsg(message, gconfig.MSGID_getNearSuperIP_recv, &data, msg.Session)

}

var GetNearSuperAddr_recvLock = new(sync.Mutex)

/*
	获取相邻节点的超级节点地址返回
*/
func GetNearSuperAddr_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	GetNearSuperAddr_recvLock.Lock()
	defer GetNearSuperAddr_recvLock.Unlock()

	//	fmt.Println("获取相邻节点的超级节点地址返回")

	nodes, err := nodeStore.ParseNodesProto(message.Body.Content)
	// nodes := make([]nodeStore.Node, 0)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(&nodes)
	// if err := json.Unmarshal(*message.Body.Content, &nodes); err != nil {
	if err != nil {
		//		fmt.Println("解析失败", err)
		return
	}

	//	fmt.Println("查找结果", newNode.IdInfo.Id.B58String())
	// engine.Log.Info("GetNearSuperAddr_recv find result total %d", len(nodes))
	for i, _ := range nodes {
		newNode := nodes[i]
		// engine.Log.Info("检查这个节点是否需要 %s",)
		// engine.Log.Info("GetNearSuperAddr_recv check need: %s", newNode.IdInfo.Id.B58String())

		manager.AddNewNode(newNode.IdInfo.Id)

		//查询是否已经有这个连接了，有了就不连接
		//避免地址和node真实地址不对应的情况
		session := engine.GetSessionByHost(newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort)))
		if session != nil {
			// name := session.GetName()
			// addrOne := nodeStore.AddressNet([]byte(session.GetName()))
			// engine.Log.Info("GetNearSuperAddr_recv conn exist: %s", addrOne.B58String())
			continue
		}

		//		if newNode.IdInfo.Id.B58String() == "5duDDfkY1tChLKGxbAtPdPysp9ghYn" || newNode.IdInfo.Id.B58String() == "5dqqnW3YTxTw9EESzT63qM63zf9BYj" {
		//		}
		//		if message.Head.Sender.B58String() == "5dsEMMhVbww4hUXV6VzaeRfHKv1nhh" {
		//			fmt.Println("查找结果", newNode.IdInfo.Id.B58String())
		//		}
		//		fmt.Println("查找结果", newNode.IdInfo.Id.B58String(), newNode.Addr, newNode.TcpPort)
		//TODO 当真实P2P网络地址和ip端口不对应的情况，容易漏连接某些节点
		//检查是否需要这个逻辑节点
		ok := nodeStore.CheckNeedNode(&newNode.IdInfo.Id)
		if !ok {
			//		fmt.Println("不需要这个逻辑节点")
			continue
		}
		//	fmt.Println("需要这个节点")
		// engine.Log.Info("GetNearSuperAddr_recv need: %s", newNode.IdInfo.Id.B58String())
		//检查是否有这个连接
		// _, ok = engine.GetSession(newNode.IdInfo.Id.B58String())
		_, ok = engine.GetSession(utils.Bytes2string(newNode.IdInfo.Id))
		if !ok {
			//查询是否已经有这个连接了，有了就不连接
			//避免地址和node真实地址不对应的情况
			// session := engine.GetSessionByHost(newNode.Addr + ":" + strconv.Itoa(int(newNode.TcpPort)))
			// if session != nil {
			// 	// name := session.GetName()
			// 	addrOne := nodeStore.AddressNet([]byte(session.GetName()))
			// 	engine.Log.Info("GetNearSuperAddr_recv conn exist: %s", addrOne.B58String())
			// 	continue
			// }
			//		fmt.Println("没有这个连接", message.Hash)
			_, err := engine.AddClientConn(newNode.Addr, uint32(newNode.TcpPort), false)
			if err != nil {
				//			fmt.Println("连接失败", err)
				// engine.Log.Info("GetNearSuperAddr_recv conn fail: %s %s", newNode.IdInfo.Id.B58String(), err.Error())
				continue
			}
			// if session.GetName() != newNode
			// session.GetName()
			// nodeStore.FindNode()
		} else {
			//		fmt.Println("有连接", message.Hash)
		}
		// engine.Log.Info("add super nodeid: %s %+v", newNode.IdInfo.Id.B58String(), newNode)
		// nodeStore.AddNode(newNode)

		//非超级节点判断超级节点是否改变
		if !nodeStore.NodeSelf.IsSuper {
			nearId := nodeStore.FindNearInSuper(&nodeStore.NodeSelf.IdInfo.Id, nil, false)
			//		nearIdStr := hex.EncodeToString(nearId)
			// fmt.Println("判断是否需要替换超级节点", nearId.B58String(), nodeStore.SuperPeerId.B58String())
			if bytes.Equal(*nearId, *nodeStore.SuperPeerId) {
				continue
			}
			nodeStore.SuperPeerId = nearId
			//		nodeStore.SuperPeerIdStr = hex.EncodeToString(nearId)
			// fmt.Println("超级节点换为:", nodeStore.SuperPeerId.B58String())
		}
	}

}

/*
	接收发送的文本消息
*/
func TextMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	engine.Log.Debug("收到文本消息")

	//	fmt.Println("这个文本消息是自己的")
	//发送给自己的，自己处理
	content := string(*message.Body.Content)
	//判断自己有没有这个好友，没有这个好友添加到申请好友列表
	sendId := message.Head.Sender.B58String()

	now := time.Now()
	msgVO := MessageVO{
		Id:      sendId,
		Index:   now.Unix(),
		Time:    utils.FormatTimeToSecond(now),
		Content: content,
		Cate:    MsgTextId,
	}

	// fmt.Println("收到文本消息  11111111111111", sendId)

	f := new(sqlite3_db.Friends)
	f, err := f.FindById(sendId)
	if err != nil {
		return
	}
	//这里xorm框架有个BUG，生成的sql语句有问题，需要重复查询一次
	if f == nil {
		f, err = f.FindById(sendId)
		if err != nil {
			return
		}
	}
	// fmt.Println("收到文本消息  222222222222222")
	if f == nil {
		// fmt.Println("收到文本消息  33333333333333")
		//未添加对方好友
		f = &sqlite3_db.Friends{
			NodeId:   sendId,  //网络节点id
			Nickname: "",      //昵称
			Notename: "",      //备注昵称
			Note:     "",      //备注信息
			Status:   2,       //好友状态.1=添加好友时，用户不在线;2=申请添加好友状态;3=同意添加;4=;5=;6=;
			IsAdd:    1,       //是否自己主动添加的好友.1=别人添加的自己;2=自己主动添加的别人;
			Hello:    content, //打招呼内容
			Read:     1,       //添加好友消息自己是否已读。1=未读;2=已读;
		}
		err = f.Add(f)
		if err != nil {
			// fmt.Println("错误", err)
			return
		}
		// fmt.Println("收到文本消息  44444444444444444")
		select {
		case MsgChannl <- &msgVO:
		default:
		}
		//回复发送者，自己已经收到消息
		message_center.SendP2pReplyMsgHE(message, config.MSGID_TextMsg_recv, nil)
		return
	} else if f != nil && f.Status == 2 {
		// fmt.Println("收到文本消息  55555555555555555")
		f = &sqlite3_db.Friends{
			NodeId: sendId,
			Hello:  content,
			Read:   1, //添加好友消息自己是否已读。1=未读;2=已读;
		}
		err := f.Update()
		if err != nil {
			return
		}
		// fmt.Println("收到文本消息  6666666666666666666")
		select {
		case MsgChannl <- &msgVO:
		default:
		}
		//回复发送者，自己已经收到消息
		message_center.SendP2pReplyMsgHE(message, config.MSGID_TextMsg_recv, nil)
		return
	}
	//------------------
	//以下情况为互相都是好友，添加聊天记录
	// fmt.Println("收到文本消息  777777777777777777")
	//回复发送者，自己已经收到消息
	ok := message_center.SendP2pReplyMsgHE(message, config.MSGID_TextMsg_recv, nil)
	if !ok {
		return
	}
	// fmt.Println("收到文本消息  88888888888888888888")
	// recvId := message.Head.RecvId.B58String()
	// fmt.Println(sendId, content)

	// if persistence.Friends_findIdExist(sendId) {
	// 	persistence.Friends_addMsgNum(sendId)
	// } else {
	// 	err := persistence.Friends_add(sendId)
	// 	if err != nil {
	// 		// fmt.Println("添加用户失败", err)
	// 	}
	// }

	ml := sqlite3_db.MsgLog{}
	id, err := ml.Add(sendId, sqlite3_db.Self, content, "", MsgTextId)
	if err == nil {
		ml.IsSuccessful(id)
	}
	msgVO.DBID = id

	select {
	case MsgChannl <- &msgVO:
	default:
	}
	// fmt.Println("收到文本消息  9999999999999999999")
	// err := persistence.SaveMsgLog(sendId, recvId, content)
	// if err != nil {
	// 	// fmt.Println("保存日志失败", err)
	// }
}

/*
	接收发送的文本消息  返回
*/
func TextMsg_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	engine.Log.Debug("收到文本消息 返回")

	//	fmt.Println("这个文本消息是自己的")
	bs := []byte("ok")
	// flood.ResponseWait(config.CLASS_im_msg_come, hex.EncodeToString(message.Body.Hash), &bs)
	flood.ResponseWait(config.CLASS_im_msg_come, utils.Bytes2string(message.Body.Hash), &bs)

}

/*
	询问关闭这个链接
*/
// func AskCloseConn(name string) {
// 	if session, ok := engine.GetSession(name); ok {
// 		session.Send(gconfig.MSGID_ask_close_conn_recv, nil, nil, false)
// 	}
// }

/*
	询问关闭这个链接
	当双方都没有这个链接的引用时，就关闭这个链接
*/
func AskCloseConn_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	// mh := nodeStore.AddressFromB58String(msg.Session.GetName())
	mh := nodeStore.AddressNet([]byte(msg.Session.GetName()))

	// node := nodeStore.FindNodeInLogic(&mh)
	// node := nodeStore.FindNode(&mh)
	if nodeStore.FindNodeInLogic(&mh) == nil && !nodeStore.FindWhiteList(&mh) {

		//自己也没有这个连接的引用，则关闭这个链接
		engine.Log.Info("Close this session")
		msg.Session.Close()
	}
}

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
*/
//func IsSendToOtherSuper(messageRecv *Message, msgId uint64, form *utils.Multihash) bool {
//	//	fmt.Println(hex.EncodeToString(messageRecv.ReceSuperId))

//	if !messageRecv.Check() {
//		fmt.Println("不能为空", messageRecv)
//		return true
//	}

//	//	recvSuperId := hex.EncodeToString(messageRecv.RecvSuperId)
//	//	recvId := hex.EncodeToString(messageRecv.RecvId)
//	recvSuperId := messageRecv.RecvSuperId
//	recvId := messageRecv.RecvId

//	//	//收消息人就是自己
//	//	if nodeStore.NodeSelf.IdInfo.Id.GetIdStr() == recvId {
//	//		return false
//	//	}

//	//	//自己不是超级节点
//	//	if !nodeStore.NodeSelf.IsSuper {
//	//		//用代理方式发送出去
//	//		if session, ok := engine.GetSession(nodeStore.SuperPeerId.GetIdStr()); ok {
//	//			session.Send(msgId, messageRecv.JSON(), false)
//	//		}
//	//		if msgId == debuf_msgid {
//	//			fmt.Println("发送给超级节点")
//	//		}
//	//		return true
//	//	}

//	//	//接收者超级节点id是自己，接收者不是自己，是自己的代理节点
//	//	if nodeStore.NodeSelf.IdInfo.Id.GetIdStr() == recvSuperId {
//	//		//查找自己的代理节点
//	//		if msgId == debuf_msgid {
//	//			fmt.Println("查找的是自己的代理节点")
//	//		}
//	//		//查找代理节点
//	//		if _, ok := nodeStore.GetProxyNode(recvId); ok {
//	//			//发送给代理节点
//	//			if session, ok := engine.GetSession(recvId); ok {
//	//				if msgId == debuf_msgid {
//	//					fmt.Println("发送出去了111")
//	//				}
//	//				session.Send(msgId, messageRecv.JSON(), false)
//	//			} else {
//	//				//这个链接断开了
//	//				if msgId == debuf_msgid {
//	//					fmt.Println("这个链接断开了")
//	//				}
//	//			}
//	//		} else {
//	//			//该节点不在线了
//	//			if msgId == debuf_msgid {
//	//				fmt.Println("111111111", recvId, recvSuperId)
//	//			}
//	//			//TODO 节点不在线，可能是切换到其他超级节点了，应该回复给消息发送者一个不在线的消息
//	//		}
//	//		return true
//	//	}

//	//	//消息是发送给逻辑节点，不是准确发送给一个节点，离逻辑节点最近的人接收并处理消息
//	//	targetId := nodeStore.FindNearInSuper(messageRecv.RecvSuperId, form, true)
//	//	if !messageRecv.Accurate && hex.EncodeToString(targetId) == nodeStore.NodeSelf.IdInfo.Id.GetIdStr() {
//	//		return false
//	//	}

//	//	//消息转发给其他节点
//	//	//	targetId = nodeStore.FindNearInSuper(messageRecv.RecvSuperId, form, false)
//	//	if session, ok := engine.GetSession(hex.EncodeToString(targetId)); ok {
//	//		session.Send(msgId, messageRecv.JSON(), false)
//	//	} else {
//	//		if msgId == debuf_msgid {
//	//			fmt.Println("1111这个session不存在", hex.EncodeToString(targetId), hex.EncodeToString(messageRecv.RecvSuperId))
//	//		}
//	//	}
//	//	return true

//	//------------------

//	if !nodeStore.NodeSelf.IsSuper {
//		if nodeStore.NodeSelf.IdInfo.Id.B58String() == recvId.B58String() {
//			return false
//		} else {
//			if messageRecv.Accurate {
//				//发错节点了
//				fmt.Println("发错节点了", nodeStore.NodeSelf.IdInfo.Id.B58String(), recvSuperId.B58String(), recvId.B58String())
//				return true
//			} else {
//				if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
//					session.Send(msgId, messageRecv.JSON(), false)
//				}
//				if msgId == debuf_msgid {
//					fmt.Println("发送给超级节点")
//				}
//				return true
//			}
//		}
//	}
//	if recvId.B58String() == recvSuperId.B58String() {
//		if recvId.B58String() == nodeStore.NodeSelf.IdInfo.Id.B58String() {
//			return false
//		} else {
//			if msgId == debuf_msgid {
//				fmt.Println("----1111111")
//			}
//			targetId := nodeStore.FindNearInSuper(messageRecv.RecvSuperId, form, true)
//			if msgId == debuf_msgid {
//				fmt.Println("----222222222")
//			}
//			if targetId.B58String() == nodeStore.NodeSelf.IdInfo.Id.B58String() {
//				//查找代理节点
//				_, ok := nodeStore.GetProxyNode(recvId.B58String())
//				if msgId == debuf_msgid {
//					fmt.Println("----333333333")
//				}
//				if ok {
//					//发送给代理节点
//					if session, ok := engine.GetSession(recvId.B58String()); ok {
//						if msgId == debuf_msgid {
//							fmt.Println("发送出去了111")
//						}
//						session.Send(msgId, messageRecv.JSON(), false)
//					} else {
//						//这个链接断开了
//						if msgId == debuf_msgid {
//							fmt.Println("这个链接断开了")
//						}
//					}
//				} else {
//					if !messageRecv.Accurate {
//						return false
//					}

//					if msgId == debuf_msgid {
//						fmt.Println("该节点不在线")
//					}
//					//该节点不在线了
//					if msgId == debuf_msgid {
//						fmt.Println("111111111", recvId, recvSuperId)
//					}
//				}
//				return true
//			}

//			session, ok := engine.GetSession(targetId.B58String())
//			if ok {
//				if msgId == debuf_msgid {
//					fmt.Println("发送出去了222")
//				}
//				session.Send(msgId, messageRecv.JSON(), false)
//			} else {
//				if msgId == debuf_msgid {
//					fmt.Println("这个链接断开了222")
//				}
//				if msgId == debuf_msgid {
//					fmt.Println("-=-=-=-= 这个session已经断开")
//				}
//			}
//		}
//		if msgId == debuf_msgid {
//			fmt.Println("4444444444", recvId, recvSuperId)
//		}
//		return true

//	} else {

//		if nodeStore.NodeSelf.IdInfo.Id.B58String() == recvSuperId.B58String() {
//			if recvId == nil {
//				return false
//			} else {
//				if msgId == debuf_msgid {
//					fmt.Println("----444444444")
//				}
//				_, ok := nodeStore.GetProxyNode(recvId.B58String())
//				if msgId == debuf_msgid {
//					fmt.Println("----555555555")
//				}
//				if ok {
//					if session, ok := engine.GetSession(recvId.B58String()); ok {
//						if msgId == debuf_msgid {
//							fmt.Println("发送出去了")
//						}
//						session.Send(msgId, messageRecv.JSON(), false)
//					} else {
//						if msgId == debuf_msgid {
//							fmt.Println("这个session不存在")
//						}
//					}
//				}
//				//代理节点转移或下线，忽略这个消息
//				if msgId == debuf_msgid {
//					fmt.Println("22222222")
//				}
//				return true
//			}
//		}
//		if msgId == debuf_msgid {
//			fmt.Println("----6666666666")
//		}
//		targetId := nodeStore.FindNearInSuper(messageRecv.RecvSuperId, form, true)
//		if msgId == debuf_msgid {
//			fmt.Println("----777777777777")
//		}
//		// hex.EncodeToString(targetId) == nodeStore.NodeSelf.IdInfo.Id.GetIdStr()
//		if targetId.B58String() == nodeStore.NodeSelf.IdInfo.Id.B58String() {
//			if messageRecv.Accurate {
//				//该节点不在线
//				fmt.Println("该节点不在线，这个包会被丢弃", msgId, targetId.B58String(),
//					messageRecv.RecvSuperId.B58String(), string(*messageRecv.JSON()))
//				if msgId == debuf_msgid {
//					fmt.Println("33333333")
//				}
//				return true
//			} else {
//				return false
//			}
//		}

//		session, ok := engine.GetSession(targetId.B58String())
//		if ok {
//			session.Send(msgId, messageRecv.JSON(), false)
//		}
//		if msgId == debuf_msgid {
//			fmt.Println("5555555555", recvId, recvSuperId)
//		}
//		return true
//	}

//}

// /*
// 	检查该消息是否是自己的
// 	不是自己的则自动转发出去
// */
// func IsSendToOtherSuperToo(messageHead *MessageHead, dataplus *[]byte, version uint64, form *nodeStore.AddressNet) bool {

// 	recvSuperId := messageHead.RecvSuperId
// 	recvId := messageHead.RecvId

// 	if !nodeStore.NodeSelf.IsSuper {
// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *recvId) {
// 			return false
// 		} else {
// 			if messageHead.Accurate {
// 				//发错节点了
// 				// fmt.Println("发错节点了", nodeStore.NodeSelf.IdInfo.Id.B58String(), recvSuperId.B58String(), recvId.B58String())
// 				return true
// 			} else {
// 				if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
// 					session.Send(version, messageHead.JSON(), dataplus, false)
// 				}
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("发送给超级节点")
// 				// }
// 				return true
// 			}
// 		}
// 	}
// 	if bytes.Equal(*recvId, *recvSuperId) {
// 		if bytes.Equal(*recvId, nodeStore.NodeSelf.IdInfo.Id) {
// 			return false
// 		} else {
// 			// if msgId == debuf_msgid {
// 			// 	fmt.Println("----1111111")
// 			// }
// 			targetId := nodeStore.FindNearInSuper(messageHead.RecvSuperId, form, true)
// 			// if msgId == debuf_msgid {
// 			// 	fmt.Println("----222222222", targetId.B58String(), nodeStore.NodeSelf.IdInfo.Id.B58String(), form)
// 			// }
// 			if bytes.Equal(*targetId, nodeStore.NodeSelf.IdInfo.Id) {
// 				//查找代理节点
// 				_, ok := nodeStore.GetProxyNode(recvId.B58String())
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("----333333333")
// 				// }
// 				if ok {
// 					//发送给代理节点
// 					if session, ok := engine.GetSession(recvId.B58String()); ok {
// 						// if msgId == debuf_msgid {
// 						// 	fmt.Println("发送出去了111")
// 						// }
// 						session.Send(version, messageHead.JSON(), dataplus, false)
// 					} else {
// 						//这个链接断开了
// 						// if msgId == debuf_msgid {
// 						// 	fmt.Println("这个链接断开了")
// 						// }
// 					}
// 				} else {
// 					if !messageHead.Accurate {
// 						return false
// 					}

// 					// if msgId == debuf_msgid {
// 					// 	fmt.Println("该节点不在线")
// 					// }
// 					// //该节点不在线了
// 					// if msgId == debuf_msgid {
// 					// 	fmt.Println("111111111", recvId, recvSuperId)
// 					// }
// 				}
// 				return true
// 			}

// 			session, ok := engine.GetSession(targetId.B58String())
// 			if ok {
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("发送出去了222")
// 				// }
// 				session.Send(version, messageHead.JSON(), dataplus, false)
// 			} else {
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("这个链接断开了222")
// 				// }
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("-=-=-=-= 这个session已经断开")
// 				// }
// 			}
// 		}
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("4444444444", recvId, recvSuperId)
// 		// }
// 		return true

// 	} else {

// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *recvSuperId) {
// 			if recvId == nil {
// 				return false
// 			} else {
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("----444444444")
// 				// }
// 				_, ok := nodeStore.GetProxyNode(recvId.B58String())
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("----555555555")
// 				// }
// 				if ok {
// 					if session, ok := engine.GetSession(recvId.B58String()); ok {
// 						// if msgId == debuf_msgid {
// 						// 	fmt.Println("发送出去了")
// 						// }
// 						session.Send(version, messageHead.JSON(), dataplus, false)
// 					} else {
// 						// if msgId == debuf_msgid {
// 						// 	fmt.Println("这个session不存在")
// 						// }
// 					}
// 				}
// 				//代理节点转移或下线，忽略这个消息
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("22222222")
// 				// }
// 				return true
// 			}
// 		}
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("----6666666666")
// 		// }
// 		targetId := nodeStore.FindNearInSuper(messageHead.RecvSuperId, form, true)
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("----777777777777")
// 		// }
// 		// hex.EncodeToString(targetId) == nodeStore.NodeSelf.IdInfo.Id.GetIdStr()
// 		if bytes.Equal(*targetId, nodeStore.NodeSelf.IdInfo.Id) {
// 			if messageHead.Accurate {
// 				//该节点不在线
// 				// fmt.Println("该节点不在线，这个包会被丢弃", msgId, targetId.B58String(),
// 				// 	messageHead.RecvSuperId.B58String(), string(*messageHead.JSON()))
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("33333333")
// 				// }
// 				return true
// 			} else {
// 				return false
// 			}
// 		}

// 		session, ok := engine.GetSession(targetId.B58String())
// 		if ok {
// 			session.Send(version, messageHead.JSON(), dataplus, false)
// 		}
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("5555555555", recvId, recvSuperId)
// 		// }
// 		return true
// 	}

// }

/*
	广播给其他人
*/
//func MulticastOther() {

//}
