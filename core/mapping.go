package core

import (
	"mandela/core/config"
	"mandela/core/engine"
	"mandela/core/nodeStore"
	"mandela/core/utils"

	"github.com/prestonTao/upnp"
)

var sys_mapping = new(upnp.Upnp) //端口映射程序

/*
	根据网络情况自己确定节点角色
*/
func AutoRole() bool {
	//判断是否联网
	if _, ok := utils.GetLocalIntenetIp(); ok {
		config.IsNotInternet = false
	} else {
		config.IsNotInternet = true
	}
	//本地地址是全球唯一公网地址
	if utils.IsOnlyIp(config.Init_LocalIP) {
		config.Init_IsGlobalOnlyAddress = true
		nodeStore.NodeSelf.IsSuper = true
		nodeStore.NodeSelf.Addr = config.Init_LocalIP
		nodeStore.NodeSelf.TcpPort = config.Init_LocalPort
		engine.Log.Debug("本机ip是公网全球唯一地址")
		return true
	}
	//尝试端口映射
	if portMapping() {
		nodeStore.NodeSelf.Addr = config.Init_GatewayAddress
		nodeStore.NodeSelf.TcpPort = config.Init_GatewayPort

		nodeStore.NodeSelf.IsSuper = true
	} else {
		nodeStore.NodeSelf.IsSuper = false
	}
	return true

	//	if _, ok := utils.GetLocalIntenetIp(); ok {
	//		config.Init_LocalIP = address
	//		config.IsNotInternet = false
	//		//本地地址是全球唯一公网地址
	//		if utils.IsOnlyIp(config.Init_LocalIP) {
	//			config.Init_IsGlobalOnlyAddress = true
	//			nodeStore.NodeSelf.IsSuper = true
	//			nodeStore.NodeSelf.Addr = address
	//			nodeStore.NodeSelf.TcpPort = config.Init_LocalPort
	//			engine.Log.Debug("本机ip是公网全球唯一地址")
	//			return true
	//		}
	//		//尝试端口映射
	//		if portMapping() {
	//			nodeStore.NodeSelf.Addr = config.Init_GatewayAddress
	//			nodeStore.NodeSelf.TcpPort = config.Init_GatewayPort

	//			nodeStore.NodeSelf.IsSuper = true
	//		} else {
	//			nodeStore.NodeSelf.IsSuper = false
	//		}
	//		//判断是否是超级节点

	//		return true
	//	} else {
	//		config.IsNotInternet = true
	//		config.Init_LocalIP = utils.GetLocalHost()
	//		return false
	//	}

}

/*
	判断自己是否有公网ip地址
	若支持upnp协议，则添加一个端口映射
*/
func portMapping() bool {
	// engine.Log.Debug("监听一个本地地址：%s:%d", Init_LocalIP, Init_LocalPort)

	// fmt.Println("监听一个本地地址：", Init_LocalIP, ":", Init_LocalPort)
	//本地地址是全球唯一公网地址
	//	if utils.IsOnlyIp(config.Init_LocalIP) {
	//		config.Init_IsGlobalOnlyAddress = true
	//		engine.Log.Debug("本机ip是公网全球唯一地址")
	//		return
	//	}
	//获得网关公网地址
	err := sys_mapping.ExternalIPAddr()
	if err != nil {
		// fmt.Println(err.Error())
		engine.Log.Warn("网关不支持端口映射", err)
		return false
	}

	//网关外网地址是不是全球唯一公网地址，不是则可能在双层网关内
	if !utils.IsOnlyIp(sys_mapping.GatewayOutsideIP) {
		engine.Log.Warn("本机在双层网关内")
		return false
	}

	engine.Log.Debug("正在尝试端口映射")
	for i := 0; i < 1000; i++ {
		if err := sys_mapping.AddPortMapping(int(config.Init_LocalPort), int(config.Init_GatewayPort), "TCP"); err == nil {
			config.Init_IsMapping = true
			engine.Log.Debug("映射到公网地址：%s:%d", config.Init_GatewayAddress, config.Init_GatewayPort)
			return true
		}
		config.Init_GatewayPort = config.Init_GatewayPort + 1
	}
	engine.Log.Warn("端口映射失败")
	// fmt.Println("端口映射失败")
	return false
}

/*
	关闭服务器时回收端口
*/
func Reclaim() {
	sys_mapping.Reclaim()
}
