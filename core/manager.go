package core

import (
	gconfig "mandela/config"
	"mandela/core/addr_manager"
	addrm "mandela/core/addr_manager"
	"mandela/core/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/nodeStore"
	_ "mandela/core/persistence"
	"mandela/core/utils"
	"mandela/core/virtual_node/manager"
	"mandela/protos/go_protos"
	"mandela/sqlite3_db"
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	isStartCore           = false
	isOnline              = make(chan bool, 1) //当连入网络，给他个信号
	Error_different_netid = errors.New("netid different")
	findNearNodeTimer     = utils.NewBackoffTimer(60 * 5)
)

func init() {
	go startUp()
	go read()
	go getNearSuperIP()
	go SendNearLogicSuperIP()
	go readOutCloseConnName()
	go loopCleanMessageCache()
}

/*
	有新地址就连接到网络中去
*/
func startUp() {
	// engine.Log.Debug("开始接收新地址")
	for addr := range addrm.SubscribesChan {
		//接收到超级节点地址消息
		// engine.Log.Debug("有新的地址 %s", addr)
		host, portStr, _ := net.SplitHostPort(addr)
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		//判断节点地址是否是自己
		// engine.Log.Debug("self node %s %d %s %d", ip, port, nodeStore.NodeSelf.Addr, nodeStore.NodeSelf.TcpPort)
		if nodeStore.NodeSelf.Addr == host && nodeStore.NodeSelf.TcpPort == uint16(port) {
			continue
		}
		//查询是否已经有这个连接了，有了就不连接
		session := engine.GetSessionByHost(host + ":" + strconv.Itoa(int(port)))
		if session != nil {
			continue
		}
		go connectNet(host, uint16(port))
	}
}

/*
	初始化消息引擎
*/
//func InitEngine() error {
//	engine.InitEngine(string(nodeStore.NodeSelf.IdInfo.Build()))
//	engine.SetAuth(new(Auth))
//	engine.SetCloseCallback(closeConnCallback)
//	msg.InitMsgRouter()

//	return nil
//}

/*
	启动消息服务器
*/
func StartEngine() bool {
	defer func() {
		//非创始节点，需要等待连接入网络，才能启动程序
		if !gconfig.ParseInitFlag() {

			// fmt.Println("等待连接到网络")
			// time.Sleep(time.Second * 5)
			select {
			case <-isOnline:
			}
		}
		manager.RegVnode()
		engine.Log.Info("Local netid is: %s", nodeStore.NodeSelf.IdInfo.Id.B58String())

	}()
	// engine.InitEngine(string(nodeStore.NodeSelf.IdInfo.JSON()))
	engine.InitEngine(nodeStore.NodeSelf.IdInfo.Id.B58String())
	engine.SetAuth(new(Auth))
	engine.SetCloseCallback(closeConnCallback)
	RegisterCoreMsg()
	// security_signal.Init()

	//	engine.ListenByListener(config.TCPListener, true)
	//占用本机一个端口
	var err error
	for i := 0; i < 100; i++ {
		//		fmt.Println(runtime.GOARCH, runtime.GOOS)
		//		if runtime.GOOS == "windows" {
		//			err = engine.Listen(config.Init_LocalIP, uint32(config.Init_LocalPort+uint16(i)), true)
		//		} else {
		//			err = engine.Listen("0.0.0.0", uint32(config.Init_LocalPort+uint16(i)), true)
		//		}
		err = engine.Listen("0.0.0.0", uint32(config.Init_LocalPort+uint16(i)), true)
		if err != nil {
			continue
		} else {
			//得到本机可用端口
			config.Init_LocalPort = config.Init_LocalPort + uint16(i)
			if !config.Init_IsMapping {
				nodeStore.NodeSelf.TcpPort = config.Init_LocalPort
			}

			//加载超级节点ip地址
			addrm.Init()
			return true
		}
	}
	return false
}

// func StartServices() {
// 	//启动核心组件
// 	StartUpCore()

// }

/*
	启动核心组件
*/
// func StartUpCore() {

// 	//	if len(*nodeStore.NodeSelf.IdInfo.Id.GetId()) == 0 {
// 	//		//		GetId()
// 	//		if len(*nodeStore.NodeSelf.IdInfo.Id.GetId()) == 0 {
// 	//			return
// 	//		}
// 	//	}
// 	if nodeStore.NodeSelf.IdInfo.Id == nil {
// 		return
// 	}
// 	engine.Log.Debug("启动服务器核心组件")
// 	engine.Log.Debug("本机id为：\n%s", nodeStore.NodeSelf.IdInfo.Id)

// 	isSuperPeer := config.CheckIsSuperPeer()
// 	//是超级节点
// 	// node := &nodeStore.Node{
// 	// 	IdInfo:  nodeStore.NodeSelf.IdInfo,
// 	// 	IsSuper: isSuperPeer, //是否是超级节点
// 	// 	//		UdpPort: 0,
// 	// }
// 	addr, port := config.GetHost()
// 	// node.Addr = addr
// 	// node.TcpPort = uint16(port)

// 	/*
// 		启动消息服务器
// 	*/
// 	engine.InitEngine(string(nodeStore.NodeSelf.IdInfo.JSON()))
// 	/*
// 		生成密钥文件
// 	*/
// 	//	var err error
// 	//	//生成密钥
// 	//	privateKey, err = rsa.GenerateKey(rand.Reader, 512)
// 	//	if err != nil {
// 	//		fmt.Println("生成密钥错误", err.Error())
// 	//		return
// 	//	}
// 	/*
// 		启动分布式哈希表
// 	*/
// 	//	nodeStore.InitNodeStore(node)
// 	/*
// 		设置关闭连接回调函数后监听
// 	*/
// 	engine.SetAuth(new(Auth))
// 	engine.SetCloseCallback(closeConnCallback)
// 	//	engine.ListenByListener(config.TCPListener, true)
// 	//	engine.Listen(config.TCPListener)
// 	//自己是超级节点就把自己添加到超级节点地址列表中去
// 	if isSuperPeer {
// 		addrm.AddSuperPeerAddr(addr + ":" + strconv.Itoa(int(port)))
// 	}

// 	isStartCore = true

// 	/*
// 		连接到超级节点
// 	*/
// 	ip, port, err := addrm.GetSuperAddrOne(false)
// 	if err == nil {
// 		// fmt.Println("准备连接到网络中去", ip, port)
// 		connectNet(ip, port)
// 	}

// 	//	go read()

// }

/*
	链接到网络中去
*/
func connectNet(ip string, port uint16) {
	//判断节点地址是否是自己
	engine.Log.Debug("self node %s %d %s %d", ip, port, nodeStore.NodeSelf.Addr, nodeStore.NodeSelf.TcpPort)
	if nodeStore.NodeSelf.Addr == ip && nodeStore.NodeSelf.TcpPort == port {
		return
	}
	//查询是否已经有这个连接了，有了就不连接
	session := engine.GetSessionByHost(ip + ":" + strconv.Itoa(int(port)))
	if session != nil {
		return
	}

	//开始连接新节点
	engine.Log.Debug("Start connecting new nodes %s %d", ip, port)

	session, err := engine.AddClientConn(ip, uint32(port), false)
	if err != nil {
		//连接失败
		engine.Log.Error("connection failed %s %d %v", ip, port, err)
		if err.Error() == Error_different_netid.Error() {
			addr_manager.RemoveIP(ip, port)
		}
		return
	}

	// mh := nodeStore.AddressFromB58String(session.GetName())
	mh := nodeStore.AddressNet([]byte(session.GetName()))
	nodeStore.SuperPeerId = &mh

	// engine.Log.Debug("超级节点为: %s", nodeStore.SuperPeerId.B58String())
	config.IsOnline = true

	select {
	case isOnline <- true:
	default:
	}

}

/*
	关闭服务器回调函数
*/
func ShutdownCallback() {
	//回收映射的端口
	Reclaim()
	// addrm.CloseBroadcastServer()
	// fmt.Println("Close over")
}

/*
	一个连接断开后的回调方法
*/
func closeConnCallback(name string) {

	mh := nodeStore.AddressNet([]byte(name))
	engine.Log.Debug("Node offline %s", mh.B58String())

	//检查此节点下线，对自己的逻辑节点是否有影响
	logicNodes := nodeStore.GetLogicNodes()

	DelNodeAddrSpeed(mh)
	nodeStore.DelNode(&mh)

	//对比删除此节点，前后，是否有变化
	if nodeStore.EqualLogicNodes(logicNodes) {
		// engine.Log.Info("有变化，重新查询逻辑节点")
		//有变化就重新查询自己的逻辑节点
		findNearNodeTimer.Release()
	}
	// _, ok := engine.GetSession(name)
	// if ok {
	// 	engine.Log.Debug("这个session还存在")
	// }
	if nodeStore.SuperPeerId == nil {
		engine.Log.Debug("-------------------------- Offline")
		return
	}
	// if  name == nodeStore.SuperPeerId.B58String() {
	if bytes.Equal([]byte(name), *nodeStore.SuperPeerId) {
		nearId := nodeStore.FindNearInSuper(&nodeStore.NodeSelf.IdInfo.Id, nil, false)
		if nearId == nil {
			//该节点没有邻居节点，已经离开了网络，没有连入网站中。
			fmt.Println("-------------------------- Left the network")
			nodeStore.SuperPeerId = nil
			//启动定时重连机制

		} else {
			nodeStore.SuperPeerId = nearId
		}
	}

	//断线重连机制
	// go func() {
	// 	tiker := utils.NewBackoffTimer(1)
	// 	for {
	// 		if nodeStore.SuperPeerId != nil {
	// 			break
	// 		}
	// 		tiker.Wait()
	// 		go addr_manager.LoadAddrForAll()
	// 	}
	// }()

}

/*
	处理查找节点的请求
	定期查询已知节点是否在线，更新节点信息
*/
func read() {
	for {
		nodeIdStr := <-nodeStore.OutFindNode

		message_center.SendSearchSuperMsg(gconfig.MSGID_checkNodeOnline, nodeIdStr, nil)

		// mhead := msg.NewMessageHead(nodeIdStr, nodeIdStr, true)
		// mbody := msg.NewMessageBody(nil, "", nil, 0)
		// message := msg.NewMessage(mhead, mbody)
		// message.Send(gconfig.MSGID_checkNodeOnline)

	}
}

/*
	定时获得相邻节点的超级节点ip地址
*/
func getNearSuperIP() {
	total := 0
	for {
		logicNodes := nodeStore.GetLogicNodes()
		clientNodes := nodeStore.GetNodesClient()
		nodesAll := append(logicNodes, clientNodes...)
		if len(nodesAll) <= 0 {
			time.Sleep(time.Second * 1)
			continue
		}

		engine.Log.Info("check near logic nodes: %d", total)
		//检查连接到本机的节点是否可以作为逻辑节点
		nodeStore.CheckClientNodeIsLogicNode()
		// for _, one := range clientNodes {
		// 	if nodeStore.CheckNeedNode(&one) {
		// 		node := nodeStore.FindNode(&one)
		// 		nodeStore.SwitchNodesClientToLogic(*node)
		// 	}
		// }

		haveFail := false
		for i, _ := range nodesAll {
			_, err := message_center.SendNeighborMsg(gconfig.MSGID_getNearSuperIP, &nodesAll[i], nil)

			// bs := flood.WaitRequest(gconfig.CLASS_near_find_logic_node, utils.Bytes2string(message.Body.Hash), int64(1))
			// if bs == nil {
			// 	engine.Log.Warn("Timeout receiving near reply message %s %s", logicNodes[i].B58String(), hex.EncodeToString(message.Body.Hash))
			// }

			//这里等待是保证异步操作完成
			time.Sleep(time.Second * 1)
			if err != nil {
				haveFail = true
			}
		}
		if haveFail {
			total = 0
			continue
		}

		//检查逻辑节点是否有变化，如果两次无变化，则停止寻找逻辑节点
		if nodeStore.EqualLogicNodes(logicNodes) {
			total = 0
		} else {
			total++
		}

		//		fmt.Println("完成一轮查找邻居节点地址")

		//定时广播自己在线
		// broadcastSelfOnline()
		// time.Sleep(time.Second * config.Time_getNear_super_ip)
		if total >= 2 {
			engine.Log.Info("check near logic nodes ok: %d", total)
			findNearNodeTimer.Wait()
			total = 1
		}
	}
}

/*
	定时广播自己在线
*/
// func broadcastSelfOnline() {
// 	//TODO 应该只初始化一次
// 	bs, _ := json.Marshal(nodeStore.NodeSelf)
// 	message_center.SendMulticastMsg(gconfig.MSGID_multicast_online_recv, &bs)

// }

/*
	通过事件驱动，给邻居节点推送对方需要的逻辑节点
*/
func SendNearLogicSuperIP() {
	for nodeOne := range nodeStore.HaveNewNode {
		// nodes := make([]nodeStore.Node, 0)
		// nodes = append(nodes, *nodeOne)
		// data, _ := json.Marshal(nodes)

		engine.Log.Info("SendNearLogicSuperIP :%s", nodeOne.IdInfo.Id.B58String())

		for _, session := range engine.GetAllSession() {
			sessionAddr := nodeStore.AddressNet([]byte(session.GetName()))
			ns := nodeStore.GetLogicNodes()
			ns = append(ns, nodeStore.GetNodesClient()...)
			ns = append(ns, nodeOne.IdInfo.Id)

			idsm := nodeStore.NewIds(sessionAddr, nodeStore.NodeIdLevel)
			for _, one := range ns {
				if bytes.Equal(sessionAddr, one) {
					continue
				}
				idsm.AddId(one)
			}
			ids := idsm.GetIds()

			nodes := make([]nodeStore.Node, 0)
			have := false //标记是否有这个新节点
			for _, one := range ids {
				if bytes.Equal(one, nodeOne.IdInfo.Id) {
					have = true
					nodes = append(nodes, *nodeOne)
					continue
				}
				addrNet := nodeStore.AddressNet(one)
				node := nodeStore.FindNode(&addrNet)
				if node != nil {
					nodes = append(nodes, *node)
				} else {
					// fmt.Println("这个节点为空")
				}
			}
			if !have {
				//没有新节点,则不发送推送消息
				continue
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

			// message_center.SendNeighborReplyMsg(message, gconfig.MSGID_getNearSuperIP_recv, &data, msg.Session)
			message_center.SendNeighborMsg(gconfig.MSGID_getNearSuperIP_recv, &sessionAddr, &data)

		}

	}
}

/*
	读取需要询问关闭的连接名称
*/
func readOutCloseConnName() {
	for name := range nodeStore.OutCloseConnName {
		// message_center.AskCloseConn(name.B58String())
		message_center.SendNeighborMsg(gconfig.MSGID_ask_close_conn_recv, name, nil)
	}
}

/*
	定时删除数据库中过期的消息缓存
*/
func loopCleanMessageCache() {
	for range time.NewTicker(time.Hour).C {
		//计算24小时以前的时间UNIX
		overtime := time.Now().Unix() - message_center.MsgCacheTimeOver
		new(sqlite3_db.MessageCache).Remove(overtime)
	}
}
