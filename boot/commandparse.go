package boot

import (
	"mandela/config"
	gconfig "mandela/core/config"
	"mandela/core/engine"
	"mandela/rpc"
	"bytes"
	"flag"
	"io/ioutil"
	"os"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	Init        = flag.String("init", "", "创建创始区块(默认：genesis.json)")
	Conf        = flag.String("conf", "conf/config.json", "指定配置文件(默认：conf/config.json)")
	Port        = flag.Int("port", 9811, "本地监听端口(默认：9811)")
	NetId       = flag.Int("netid", 20, "网络id(默认：20)")
	Ip          = flag.String("ip", "0.0.0.0", "本地IP地址(默认：0.0.0.0)")
	WebAddr     = flag.String("webaddr", "0.0.0.0", "本地web服务器IP地址(默认：0.0.0.0)")
	WebPort     = flag.Int("webport", 2080, "web服务器端口(默认：2080)")
	WebStatic   = flag.String("westatic", "", "本地web静态文件目录")
	WebViews    = flag.String("webviews", "", "本地web Views文件目录")
	DataDir     = flag.String("datadir", "", "指定数据目录")
	DbCache     = flag.Int("dbcache", 25, "设置数据库缓存大小，单位为兆字节（MB）（默认：25）")
	TimeOut     = flag.Int("timeout", 0, "设置连接超时，单位为毫秒")
	RpcServer   = flag.Bool("rpcserver", false, "打开或关闭JSON-RPC true/false(默认：false)")
	RpcUser     = flag.String("rpcuser", "", "JSON-RPC 连接使用的用户名")
	RpcPassword = flag.String("rpcpassword", "", "JSON-RPC 连接使用的密码")
	WalletPwd   = flag.String("walletpwd", config.Wallet_keystore_default_pwd, "钱包密码")
	classpath   = flag.String("classpath", "", "jar包路径")
	Load        = flag.String("load", "", "从历史区块拉起链端")
)

func Step() {
	parseConfig()
	flag.Parse()
	parseParam()
	parseNonFlag()
}
func parseNonFlag() {
	for _, param := range flag.Args() {
		switch param {
		case "init":
			// config.InitNode = true
			// startblock.BuildFirstBlock()
		case "testnet":
			// fmt.Println("testnet")
		}
	}
}
func parseParam() {
	flag.VisitAll(func(v *flag.Flag) {
		switch v.Name {
		case "port":
			gconfig.Init_LocalPort = uint16(*Port)
			gconfig.Init_GatewayPort = gconfig.Init_LocalPort
		case "netid":
			engine.Netid = uint32(*NetId)
		case "webaddr":
			config.WebAddr = *WebAddr
		case "webport":
			config.WebPort = uint16(*WebPort)
		case "westatic":
			config.Web_path_static = *WebStatic
		case "webviews":
			config.Web_path_views = *WebViews
		case "datadir":
			//datadir
		case "dbcache":
			//dbcache
		case "timeout":
			//timeout
		case "rpcserver":
			rpc.Server = *RpcServer
		case "rpcuser":
			rpc.User = *RpcUser
		case "rpcpassword":
			rpc.Password = *RpcPassword
		case "walletpwd":
			config.Wallet_keystore_default_pwd = *WalletPwd
			//rpc
		}
	})

}

func parseConfig() {
	if !exists(*Conf) {
		return
	}
	confpath := flag.Lookup("conf").Value.String()
	bs, err := ioutil.ReadFile(confpath)
	if err != nil {
		panic("Read conf error: " + err.Error())
		return
	}
	cfi := new(config.Config)
	// err = json.Unmarshal(bs, cfi)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(cfi)

	if err != nil {
		panic("Parse conf error: " + err.Error())
		return
	}
	*Port = int(cfi.Port)
	*NetId = int(cfi.Netid)
	*Ip = cfi.IP
	*WebAddr = cfi.WebAddr
	*WebPort = int(cfi.WebPort)
	*WebStatic = cfi.WebStatic
	*WebViews = cfi.WebViews
	*RpcServer = cfi.RpcServer
	*RpcUser = cfi.RpcUser
	*RpcPassword = cfi.RpcPassword

}
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}
