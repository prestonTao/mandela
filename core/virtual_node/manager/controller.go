package manager

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/virtual_node"
	"bytes"
	"runtime"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
	查询邻居节点
*/
func FindNearVnode() {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	c := virtual_node.GetFindVnodeChan()

	// engine.Log.Info("开始接收查询邻居节点消息信号")

	for one := range c {
		// bs, err := json.Marshal(one)
		bs, err := one.Proto()
		if err != nil {
			continue
		}
		message_center.SendP2pMsgHE(config.MSGID_vnode_getNearSuperIP, &one.Target.Nid, &bs)
	}
	// engine.Log.Info("停止接收查询邻居节点消息信号")
}

/*
	定时获取邻居节点的虚拟节点地址
*/
func LoopGetVnodeinfo() {
	// vnodeinfo := virtual_node.GetVnodeLogical()

	newTiker := utils.NewBackoffTimer(config.VNODE_get_neighbor_vnode_tiker...)

	open := false

	for {
		newTiker.Wait()

		//功能未开启则退出
		if len(virtual_node.GetVnodeNumber()) <= 0 {
			open = false
			continue
		}

		//虚拟节点功能刚打开，重新设置同步时间
		if !open {
			newTiker.Reset()
		}
		open = true

		nodes := nodeStore.GetLogicNodes()
		for _, one := range nodes {
			nodeinfo := virtual_node.Vnodeinfo{
				Nid:   one,                                //节点真实网络地址
				Index: 0,                                  //节点第几个空间，从1开始,下标为0的节点为实际节点。
				Vid:   virtual_node.AddressNetExtend(one), //vid，虚拟节点网络地址
			}
			virtual_node.AddLogicVnodeinfo(nodeinfo)
		}
	}
}

/*
	添加一个新节点，发送消息看这个节点是否开通了虚拟节点
	已开通则添加这个节点，未开通则抛弃。
*/
func AddNewNode(addr nodeStore.AddressNet) {
	utils.Go(func() {
		goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
		_, file, line, _ := runtime.Caller(0)
		engine.AddRuntime(file, line, goroutineId)
		defer engine.DelRuntime(file, line, goroutineId)
		//自己节点没开通虚拟节点，就不需要添加
		if len(virtual_node.GetVnodeNumber()) <= 0 {
			return
		}

		//判断这个节点是否已经添加，已经添加则不需要重复添加
		vnodeinfoMap := virtual_node.GetVnodeLogical()
		for _, v := range vnodeinfoMap {
			if bytes.Equal(v.Nid, addr) {
				return
			}
		}

		//查询节点是否开通了虚拟节点
		message, ok, isSelf := message_center.SendP2pMsgHE(config.MSGID_vnode_getstate, &addr, nil)
		if !ok || isSelf {
			return
		}
		// bs := flood.WaitRequest(message_center.CLASS_vnode_getstate, hex.EncodeToString(message.Body.Hash), 0)
		bs, _ := flood.WaitRequest(message_center.CLASS_vnode_getstate, utils.Bytes2string(message.Body.Hash), 0)
		if bs == nil || len(*bs) <= 0 {
			return
		}
		index := utils.BytesToUint64(*bs)

		for i := uint64(0); i < index; i++ {
			vnodeinfo := virtual_node.BuildNodeinfo(i, addr)
			virtual_node.AddLogicVnodeinfo(*vnodeinfo)
		}
	})
}
