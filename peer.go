package main

import (
	"mandela/boot"
	_ "mandela/boot"
	"mandela/chain_witness_vote"
	"mandela/core"
	"mandela/core/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"fmt"

	// "mandela/im"
	// "mandela/proxyhttp"
	// "mandela/sharebox"
	// "mandela/store"
	"mandela/web"
	"mandela/web/routers"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/astaxie/beego"
)

func main() {
	// go pprofMem()
	boot.Step()
	StartUP()

	// 注意，有时候 defer f.Close()， defer pprof.StopCPUProfile() 会执行不到，这时候我们就会看到 prof 文件是空的， 我们需要在自己代码退出的地方，增加上下面两行，确保写文件内容了。
	//	pprof.StopCPUProfile()
	//	f.Close()
}

func pprofMem() {

	//	f, err := os.OpenFile("mem.prof", os.O_RDWR|os.O_CREATE, 0644)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	defer f.Close()
	//	pprof.StartCPUProfile(f)
	//	defer pprof.StopCPUProfile()

	//	pprof.Lookup("")

	//	f, err := os.Create("mem.prof")
	//	if err != nil {
	//		fmt.Fprintf(os.Stderr, "Can not create mem profile output file: %s", err)
	//		return
	//	}
	//	if err = pprof.WriteHeapProfile(f); err != nil {
	//		fmt.Fprintf(os.Stderr, "Can not write %s: %s", *memProfile, err)
	//	}
	//	f.Close()

	runtime.GOMAXPROCS(1)              // 限制 CPU 使用数，避免过载
	runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪
	runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪

	runtime.MemProfileRate = 512 * 1024 //512k

	// startMemProfile()
	time.Sleep(time.Minute * 5)
	stopMemProfile("mem.prof")
}

// func startMemProfile() {

// }

func stopMemProfile(memProfile string) {
	f, err := os.Create(memProfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can not create mem profile output file: %s", err)
		return
	}
	if err = pprof.WriteHeapProfile(f); err != nil {
		fmt.Fprintf(os.Stderr, "Can not write %s: %s", memProfile, err)
	}
	f.Close()
}

func StartUP() {

	//	ip := beego.AppConfig.DefaultString("ip", "0.0.0.0")
	//	portStr := beego.AppConfig.DefaultString("port", "0")

	nodeStore.NodeSelf.Addr = config.Init_LocalIP
	nodeStore.NodeSelf.TcpPort = config.Init_LocalPort
	nodeStore.NodeSelf.IsSuper = true

	errInt := nodeStore.InitNodeStore()
	if errInt == 1 {
		panic("密码错误")
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
		// routers.RegisterWallet()
		err := chain_witness_vote.Register()
		if err != nil {
			fmt.Println(err)
		}
	}()

	go func() {
		//启动RPC模块
		routers.RegisterRpc()
	}()

	// go func() {
	// 	//启动匿名web模块
	// 	routers.RegisterAnonymousNet()
	// 	proxyhttp.Register()
	// }()

	//启动共享盒子模块
	// go func() {
	// 	routers.RegisterSharebox()
	// 	sharebox.RegsterStore()

	// 	//添加一个测试目录
	// 	// sharebox.AddLocalShareFolders("D:/test/share1")

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
