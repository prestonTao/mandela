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
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

var (
	//	privateKey  *rsa.PrivateKey
	isStartCore = false

	isOnline              = make(chan bool, 1) //当连入网络，给他个信号
	Error_different_netid = errors.New("netid different")
)

func init() {

	go startUp()
	go read()
	go getNearSuperIP()
}

/*
	有新地址就连接到网络中去
*/
func startUp() {
	//	fmt.Println("这个方法究竟有没有执行啊")
	//	one := make(chan string, 0)
	//	addrm.AddSubscribe(one)
	for addr := range addrm.SubscribesChan {
		//		fmt.Println("这个方法究竟有没有执行啊222")
		//		engine.Log.Debug("开始接收新地址")
		//		fmt.Println("这个方法究竟有没有执行啊333")
		//接收到超级节点地址消息
		//		addr := <-addrm.SubscribesChan
		// engine.Log.Debug("有新的地址 %s", addr)
		//		fmt.Println("有新的地址 %s", addr)
		host, portStr, _ := net.SplitHostPort(addr)
		port, err := strconv.Atoi(portStr)
		if err != nil {
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
		engine.Log.Info("Local netid is：%s", nodeStore.NodeSelf.IdInfo.Id.B58String())

	}()
	engine.InitEngine(string(nodeStore.NodeSelf.IdInfo.JSON()))
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
			go addrm.LoadAddrForAll()
			return true
		}
	}
	return false
}

func StartService() {
	//启动核心组件
	StartUpCore()

}

/*
	启动核心组件
*/
func StartUpCore() {

	//	if len(*nodeStore.NodeSelf.IdInfo.Id.GetId()) == 0 {
	//		//		GetId()
	//		if len(*nodeStore.NodeSelf.IdInfo.Id.GetId()) == 0 {
	//			return
	//		}
	//	}
	if nodeStore.NodeSelf.IdInfo.Id == nil {
		return
	}
	engine.Log.Debug("启动服务器核心组件")
	engine.Log.Debug("本机id为：\n%s", nodeStore.NodeSelf.IdInfo.Id)

	isSuperPeer := config.CheckIsSuperPeer()
	//是超级节点
	node := &nodeStore.Node{
		IdInfo:  nodeStore.NodeSelf.IdInfo,
		IsSuper: isSuperPeer, //是否是超级节点
		//		UdpPort: 0,
	}
	addr, port := config.GetHost()
	node.Addr = addr
	node.TcpPort = uint16(port)

	/*
		启动消息服务器
	*/
	engine.InitEngine(string(nodeStore.NodeSelf.IdInfo.JSON()))
	/*
		生成密钥文件
	*/
	//	var err error
	//	//生成密钥
	//	privateKey, err = rsa.GenerateKey(rand.Reader, 512)
	//	if err != nil {
	//		fmt.Println("生成密钥错误", err.Error())
	//		return
	//	}
	/*
		启动分布式哈希表
	*/
	//	nodeStore.InitNodeStore(node)
	/*
		设置关闭连接回调函数后监听
	*/
	engine.SetAuth(new(Auth))
	engine.SetCloseCallback(closeConnCallback)
	//	engine.ListenByListener(config.TCPListener, true)
	//	engine.Listen(config.TCPListener)
	//自己是超级节点就把自己添加到超级节点地址列表中去
	if isSuperPeer {
		addrm.AddSuperPeerAddr(addr + ":" + strconv.Itoa(int(port)))
	}

	isStartCore = true

	/*
		连接到超级节点
	*/
	ip, port, err := addrm.GetSuperAddrOne(false)
	if err == nil {
		// fmt.Println("准备连接到网络中去", ip, port)
		connectNet(ip, port)
	}

	//	go read()

}

/*
	链接到网络中去
*/
func connectNet(ip string, port uint16) {
	//判断节点地址是否是自己
	// engine.Log.Debug("本机节点 %s %d %s %d", ip, port, nodeStore.NodeSelf.Addr, nodeStore.NodeSelf.TcpPort)
	if nodeStore.NodeSelf.Addr == ip && nodeStore.NodeSelf.TcpPort == port {
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

	mh := nodeStore.AddressFromB58String(session.GetName())
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
	engine.Log.Debug("Node offline %s", name)
	// fmt.Println("节点下线", name)

	mh := nodeStore.AddressFromB58String(name)

	nodeStore.DelNode(&mh)
	// _, ok := engine.GetSession(name)
	// if ok {
	// 	engine.Log.Debug("这个session还存在")
	// }
	if nodeStore.SuperPeerId == nil {
		engine.Log.Debug("-------------------------- Offline")
		return
	}
	if name == nodeStore.SuperPeerId.B58String() {
		nearId := nodeStore.FindNearInSuper(&nodeStore.NodeSelf.IdInfo.Id, nil, false)
		if nearId == nil {
			//该节点没有邻居节点，已经离开了网络，没有连入网站中。
			fmt.Println("-------------------------- Left the network")
			//启动定时重连机制
			nodeStore.SuperPeerId = nil

		} else {
			nodeStore.SuperPeerId = nearId
			//			nodeStore.SuperPeerIdStr = hex.EncodeToString(nearId)

		}
		//		if nodeStore.NodeSelf.IsSuper {
		//			//			nodeStore.Get(nodeStore.NodeSelf.IdInfo.Id, false, []byte{})
		//		} else {
		//		}
	}

	//断线重连机制
	go func() {
		tiker := utils.NewBackoffTimer(1)
		for {
			if nodeStore.SuperPeerId != nil {
				break
			}
			tiker.Wait()
			go addr_manager.LoadAddrForAll()
		}
	}()

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
	for {
		for _, key := range nodeStore.GetLogicNodes() {

			message_center.SendNeighborMsg(gconfig.MSGID_getNearSuperIP, key, nil)
			time.Sleep(time.Second * 1)

		}
		//		fmt.Println("完成一轮查找邻居节点地址")
		time.Sleep(time.Second * config.Time_getNear_super_ip)
	}
}
