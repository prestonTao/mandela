package addr_manager

import (
	gconfig "mandela/config"
	"mandela/core/config"
	"mandela/core/utils"
	"encoding/json"
	"errors"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (

	//超级节点地址列表文件地址
	Path_SuperPeerAddress = filepath.Join(gconfig.Path_configDir, "nodeEntry.json")

	//超级节点地址最大数量
	Sys_config_entryCount = 1000
	//本地保存的超级节点地址列表
	Sys_superNodeEntry = make(map[string]string, Sys_config_entryCount)
	//清理本地保存的超级节点地址间隔时间
	Sys_cleanAddressTicker = time.Minute * 1
	//需要关闭定时清理超级节点地址列表程序时，向它发送一个信号
	Sys_StopCleanSuperPeerEntry = make(chan bool)

	//保存不同渠道获得超级节点地址方法
	loadAddrFuncs  = make([]func(), 0)
	startLoadChan  = make(chan bool, 1)     //当本机没有可用的超级节点地址，这里会收到一个信号
	SubscribesChan = make(chan string, 100) //保存超级节点的ip地址，当有可用的超级节点地址，这里会收到一个信号
)

/*
	启动本地服务
*/
func init() {
	go smartLoadAddr()
	//	startLoadChan <- true
}

/*
	从所有渠道加载超级节点地址列表
*/
func LoadAddrForAll() {
	//	fmt.Println("haha")
	//	//加载本地文件
	//	//官网获取
	//	//私网获取
	//	//局域网组播获取
	//	LoadByMulticast()

	for _, one := range loadAddrFuncs {
		one()
	}
}

/*
	添加一个获得超级节点地址方法
*/
func registerFunc(newFunc func()) {
	loadAddrFuncs = append(loadAddrFuncs, newFunc)
}

/*
	根据信号加载超级节点地址列表
*/
func smartLoadAddr() {
	for {
		<-startLoadChan
		LoadAddrForAll()
	}
}

/*
	添加一个地址
*/
func AddSuperPeerAddr(addr string) {
	// fmt.Println("添加一个超级节点地址", addr)
	//检查是否重复
	// if _, ok := Sys_superNodeEntry[addr]; ok {
	// 	return
	// }

	//判断对方是否是超级节点
	if gconfig.NetType == gconfig.NetType_release {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return
		}
		if !utils.IsOnlyIp(host) {
			return
		}
	}

	//检查这个地址是否可用
	if !CheckOnline(addr) {
		return
	}
	Sys_superNodeEntry[addr] = ""
	//	BroadcastSubscribe(addr)
	SubscribesChan <- addr
}

/*
	随机得到一个可用的超级节点地址
	这个地址不能是自己的地址
	@bool    contain    是否包含自己的地址
	@return  addr       随机获得的地址
*/
func GetSuperAddrOne(contain bool) (string, uint16, error) {
	addr, port := config.GetHost()
	myaddr := addr + ":" + strconv.Itoa(int(port))
	rand.Seed(int64(time.Now().Nanosecond()))
	for len(Sys_superNodeEntry) != 0 {
		if !contain && len(Sys_superNodeEntry) == 1 {
			if _, ok := Sys_superNodeEntry[myaddr]; ok {
				return "", 0, errors.New("超级节点地址只有自己")
			}
		}
		// 随机取[0-1000)
		r := rand.Intn(len(Sys_superNodeEntry))
		count := 0
		for key, _ := range Sys_superNodeEntry {
			if count == r {
				if key == myaddr {
					if contain {
						host, portStr, _ := net.SplitHostPort(key)
						port, err := strconv.Atoi(portStr)
						if err != nil {
							return "", 0, errors.New("IP地址解析失败")
						}
						return host, uint16(port), nil
					} else {
						break
					}
				}
				if CheckOnline(key) {
					host, portStr, _ := net.SplitHostPort(key)
					port, err := strconv.Atoi(portStr)
					if err != nil {
						return "", 0, errors.New("IP地址解析失败")
					}
					return host, uint16(port), nil
				} else {
					delete(Sys_superNodeEntry, key)
					break
				}
			}
			count = count + 1
		}
	}
	return "", 0, errors.New("没有可用的超级节点地址")
}

/*
	保存超级节点地址列表到本地配置文件
	@path  保存到本地的磁盘路径
*/
func saveSuperPeerEntry(path string) {
	fileBytes, _ := json.Marshal(Sys_superNodeEntry)
	file, _ := os.Create(path)
	file.Write(fileBytes)
	file.Close()
}
