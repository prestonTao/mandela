package routers

import (
	"mandela/rpc"
	"mandela/web/controllers"
	"mandela/web/controllers/anonymousnet"
	"mandela/web/controllers/sharebox"
	"mandela/web/controllers/store"
	"mandela/web/controllers/wallet"
	"sync"

	"github.com/astaxie/beego"
)

var routerLock = new(sync.RWMutex)

func Router(rootpath string, c beego.ControllerInterface, mappingMethods ...string) (app *beego.App) {
	routerLock.Lock()
	app = beego.Router(rootpath, c, mappingMethods...)
	routerLock.Unlock()
	return
}

func Start() {
	Router("/", &controllers.MainController{}, "get:Test") //云存储首页
	// Router("/", &wallet.Index{}, "get:Index")   //首页
	// Router("/", &sharebox.Index{}, "get:Index") //首页

	// Router("/", &controllers.MainController{})
	// Router("/self/msg", &controllers.MsgController{}, "get:MsgPage")           //打开消息页面
	// Router("/self/sendtextmsg", &controllers.MainController{}, "post:SendMeg") //给节点发送文本消息
	// Router("/self/msg/getmsg", &controllers.MsgController{}, "post:GetMsg")    //轮询获取消息
	// Router("/self/friend/add", &controllers.MsgController{}, "post:AddFriend") //添加一个好友

	Router("/self/test", &controllers.MainController{}, "get:Test") //
	// //	Router("/self/applyname", &controllers.MainController{}, "post:ApplyName")         //申请一个域名
	// //	Router("/self/sendmsgtoname", &controllers.MainController{}, "post:SendMegToName") //给一个域名发送消息

	// Router("/self/bttest", &controllers.MainController{}, "post:BtTest") //

}

//云存储模块
func RegisterStore() {
	Router("/store/getlist", &store.Index{}, "get:GetList")            //获取文件列表
	Router("/store/addfile", &store.Index{}, "post:AddFile")           //添加一个文件
	Router("/store/addcryptfile", &store.Index{}, "post:AddCryptFile") //添加一个加密文件
	Router("/store/:hash", &store.Index{}, "get:GetFile")              //获取一个文件
	Router("/store/:down/:hash", &store.Index{}, "get:GetFile")        //获取一个文件
}

//共享存储模块
// func RegisterSharestore() {
// 	Router("/", &store.Index{}, "get:Index")                      //云存储首页
// 	Router("/store/getlist", &sharestore.Index{}, "get:GetList")  //获取文件列表
// 	Router("/store/addfile", &sharestore.Index{}, "post:AddFile") //添加一个文件
// 	Router("/store/:hash", &sharestore.Index{}, "get:GetFile")    //获取一个文件
// }

//共享盒子
func RegisterSharebox() {
	Router("/sharebox/page", &sharebox.Index{}, "get:Index")       //云存储首页
	Router("/sharebox/getlist", &sharebox.Index{}, "get:GetList")  //获取文件列表
	Router("/sharebox/addfile", &sharebox.Index{}, "post:AddFile") //添加一个文件
	Router("/sharebox/:hash", &sharebox.Index{}, "get:GetFile")    //获取一个文件
}

func RegisterWallet() {
	// Router("/self/wallet", &wallet.Index{}, "get:Index")        //钱包首页
	Router("/self/getinfo", &wallet.Index{}, "post:Getinfo")            //获取节点信息
	Router("/self/block", &wallet.Index{}, "post:Block")                //获取区块头及交易
	Router("/self/witnesslist", &wallet.Index{}, "post:GetWitnessList") //获取见证人列表
}
func RegisterRpc() {
	Router("/rpc", &rpc.Bind{}, "post:Index") //rpc调用
}

func RegisterAnonymousNet() {
	Router("/*", &anonymousnet.MainController{}, "*:Agent")          //
	Router("/:urls", &anonymousnet.MainController{}, "get:AgentToo") //
}
