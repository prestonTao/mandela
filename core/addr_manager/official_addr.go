/*
	获取超级节点地址方式：
		1.本地配置文件方式获取
		2.官方目录服务器获取

	工作流程：
		1.判断配置文件夹是否存在，不存在则创建空文件夹。
		2.读取本地超级节点地址文件，添加配置中的地址。
		3.添加官方地址。
		4.启动心跳检查本地地址是否可用。
*/
package addr_manager

import (
	// "fmt"
	gconfig "mandela/config"
	"io/ioutil"
	"net/http"
	"os"
)

const (
// Local_
)

var (
	//官方节点地址
	// Path_SuperPeerdomain = "mandela.io:9981"
	//官方目录服务器地址
	// Path_DirectotyServerAddr = []string{"mandela.io:19981"}
	Path_DirectotyServerAddr = []string{}
)

func init() {
	//	registerFunc()
}

func InitSuperPeer() {
	// Path_SuperPeerdomain = config.Init_LocalIP + ":9981"
	// Path_DirectotyServerAddr = []string{config.Init_LocalIP + ":19981"}
	//判断文件夹是否存在
	if _, err := os.Stat(gconfig.Path_configDir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(gconfig.Path_configDir, 0755)
		}
		panic(err.Error())
	}
}

/*
	开始加载超级节点地址
*/
func StartLoadSuperPeer() {
	// Sys_superNodeEntry[Path_SuperPeerdomain] = ""
	// Sys_superNodeEntry.Store(Path_SuperPeerdomain, "")
	CheckAddr()
	// go func() {
	// 	//获得一个心跳
	// 	for range time.NewTicker(Sys_cleanAddressTicker).C {
	// 		CheckAddr()
	// 	}
	// }()
}

/*
	从目录服务器获取超级节点地址
	@ ds   目录服务器地址列表
*/
func pullSuperPeerAddrForDS(ds []string) {
	for _, addrOne := range ds {
		resp, _ := http.Get("http://" + addrOne)
		body, _ := ioutil.ReadAll(resp.Body)
		// fmt.Println("返回结果：", string(body))
		parseSuperPeerEntry(body)
	}
}

/*
	关闭并重启读取并解析本地的超级节点列表文件程序
*/
// func reloadSuperPeerEntry() {
// 	Sys_StopCleanSuperPeerEntry <- true
// 	loadSuperPeerEntry()
// }
