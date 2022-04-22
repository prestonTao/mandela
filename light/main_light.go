package main

import (
	"mandela/chain_witness_vote"
	"mandela/core"
	"mandela/core/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/web"
	"mandela/web/routers"
	"fmt"

	"github.com/astaxie/beego"
)

func main() {
	StartUP()
	fmt.Println("end")
}

func StartUP() {

	//	ip := beego.AppConfig.DefaultString("ip", "0.0.0.0")
	//	portStr := beego.AppConfig.DefaultString("port", "0")

	nodeStore.NodeSelf.Addr = config.Init_LocalIP
	nodeStore.NodeSelf.TcpPort = config.Init_LocalPort
	nodeStore.NodeSelf.IsSuper = true

	errInt := nodeStore.InitNodeStore()
	if errInt == 1 {
		panic("password fail")
	}
	// engine.Log.Info("本机netid为：%s", nodeStore.NodeSelf.IdInfo.Id.B58String())

	//	if err := core.InitEngine(); err != nil {
	//		panic(err)
	//	}
	core.StartEngine()

	web.Start()

	//启动socks5代理服务
	//	go core.InitSocks5Server()

	//	message_center.Root_AgreeRegisterName(config.C_root_name, nodeStore.PublicKey)

	//	ids := make([][]byte, 0)
	//	ids = append(ids, nodeStore.NodeSelf.IdInfo.Id)
	//	tempId := nodeStore.NewTempId(nodeStore.NodeSelf.IdInfo.Id, nodeStore.NodeSelf.IdInfo.Id)

	//启动其他模块
	// go func() {
	// 	//启动云存储模块
	// 	store.RegsterStore()
	// 	routers.RegisterStore()
	// 	//		engine.Log.Debug("各个模块加载完成")
	// }()

	go func() {
		// return
		//启动区块链模块
		routers.RegisterWallet()
		err := chain_witness_vote.Register()
		if err != nil {
			fmt.Println(err)
		}
	}()

	//启动RPC模块
	routers.RegisterRpc()
	go func() {
	}()

	//启动匿名web模块
	// routers.RegisterAnonymousNet()
	// go func() {
	// 	proxyhttp.Register()
	// }()

	//启动共享盒子模块
	// routers.RegisterSharebox()
	// go func() {
	// 	sharebox.RegsterStore()

	// 	//添加一个测试目录
	// 	// sharebox.AddLocalShareFolders("E:/share/upload")

	// }()

	// go func() {
	// 	//启动云存储模块
	// 	cloud_space.RegsterCloudSpace()
	// 	// routers.RegisterStore()
	// 	//		engine.Log.Debug("各个模块加载完成")
	// }()

	//启动IM模块
	// go func() {
	// 	im.RegisterIM()
	// }()

	//启动web模块
	go func() {
		beego.Run()
	}()
	<-utils.GetStopService()

}
