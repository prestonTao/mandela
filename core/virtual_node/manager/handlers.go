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
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func RegVnode() {
	message_center.Register_p2pHE(config.MSGID_vnode_getstate, GetVnodeOpenState)                //获取节点的虚拟节点开通状态
	message_center.Register_p2pHE(config.MSGID_vnode_getstate_recv, GetVnodeOpenState_recv)      //获取节点的虚拟节点开通状态 返回
	message_center.Register_p2pHE(config.MSGID_vnode_getNearSuperIP, GetNearSuperAddr)           //获取相邻节点的Vnode地址
	message_center.Register_p2pHE(config.MSGID_vnode_getNearSuperIP_recv, GetNearSuperAddr_recv) //获取相邻节点的Vnode地址 返回

	virtual_node.LoadVnode()
	go LoopGetVnodeinfo()
	go FindNearVnode()

}

/*
	获取节点的虚拟节点开通状态
*/
func GetVnodeOpenState(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	n := len(virtual_node.GetVnodeNumber())
	bs := utils.Uint64ToBytes(uint64(n))
	message_center.SendP2pReplyMsgHE(message, config.MSGID_vnode_getstate_recv, &bs)

}

/*
	获取节点的虚拟节点开通状态 返回
*/
func GetVnodeOpenState_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(message_center.CLASS_vnode_getstate, hex.EncodeToString(message.Body.Hash), message.Body.Content)
}

/*
	获取相邻节点的Vnode地址
*/
func GetNearSuperAddr(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// fmt.Println("-----------------接收到邻居节点虚拟节点查询消息")
	// fmt.Println(string(*message.Body.Content))
	findVnodeVO := new(virtual_node.FindVnodeVO)
	// err := json.Unmarshal(*message.Body.Content, findVnodeVO)
	decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	decoder.UseNumber()
	err := decoder.Decode(findVnodeVO)
	if err != nil {
		fmt.Println("json格式化错误", err)
		return
	}

	//验证消息发送方节点id
	if !findVnodeVO.Self.Check() {
		fmt.Println("验证发送方节点不合法")
		return
	}

	// fmt.Println("目标id", findVnodeVO.Target.B58String())

	//验证Target参数是否在自己的节点中
	have := false
	vnodeinfo := virtual_node.GetVnodeSelf()
	for _, one := range vnodeinfo {
		// fmt.Println("本节点id", one.Vid.B58String())
		if bytes.Equal(one.Vid, findVnodeVO.Target.Vid) {
			have = true
			break
		}
	}
	// if !have && bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, findVnodeVO.Target) {
	// 	have = true
	// }

	//检查发送目标节点是否在本节点中，不在本节点中，不处理
	if !have {
		fmt.Println("不在本节点中")
		return
	}

	//将对方节点保存到自己的逻辑节点中，此协议用于节点发现。
	virtual_node.AddLogicVnodeinfo(findVnodeVO.Self)
	// fmt.Println("是否添加成功", ok)
	// for _, one := range virtual_node.GetVnodeLogical() {
	// 	fmt.Println("自己节点的逻辑节点地址", one.Vid.B58String())
	// }

	//获取本节点保存的所有id
	idsMap := make(map[string]nodeStore.AddressNet)
	vnodeinfos := virtual_node.GetVnodeLogical()
	//添加自己的虚拟节点中的逻辑节点
	for _, one := range vnodeinfos {
		temp := nodeStore.AddressNet(one.Vid)
		idsMap[temp.B58String()] = temp
	}
	// fmt.Println("添加自己的逻辑虚拟节点之后", idsMap)
	//添加自己的虚拟节点
	for _, one := range vnodeinfo {
		temp := nodeStore.AddressNet(one.Vid)
		idsMap[temp.B58String()] = temp
	}
	// fmt.Println("添加自己的虚拟节点之后", idsMap)
	// //包含自己真实节点id，必须要包含真实节点，否则网络会形成孤岛。
	// for _, one := range nodeStore.GetAllNodes() {
	// 	idsMap[one.B58String()] = *one
	// }
	//不包括自己
	// delete(idsMap, findVnodeVO.Target.B58String())
	//不包括消息发送方id
	delete(idsMap, findVnodeVO.Self.Vid.B58String())

	// fmt.Println("需要查询的地址", len(idsMap), idsMap)

	//查找对方节点所需要的id
	selfVid := nodeStore.AddressNet(findVnodeVO.Self.Vid)
	idsm := nodeStore.NewIds(selfVid, nodeStore.NodeIdLevel)
	for _, one := range idsMap {
		// tempId := nodeStore.AddressNet(one.Vid)
		idsm.AddId(one)
	}
	ids := idsm.GetIds()

	// fmt.Println("查询到的节点id", len(ids))
	vinfos := make([]virtual_node.Vnodeinfo, 0)
	for _, one := range ids {
		temp := virtual_node.AddressNetExtend(one)
		// fmt.Println("要查询的id", temp.B58String())
		vinfo := virtual_node.FindVnodeinfo(temp)
		if vinfo == nil {
			// fmt.Println("查询vinfo为空 1111111111111")
			// fmt.Println("自己的id", nodeStore.NodeSelf.IdInfo.Id.B58String())

			// if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *one) {
			// 	vinfo = &Vnodeinfo{
			// 		Nid:   *one, //节点真实网络地址
			// 		Index: 0,    //节点第几个空间，从1开始,下标为0的节点为实际节点。
			// 		Vid:   temp, //vid，虚拟节点网络地址
			// 	}
			// 	vinfos = append(vinfos, *vinfo)
			// 	fmt.Println("查询vinfo为空 2222222222222222222")
			// 	continue
			// }
			// //在自己的真实邻居节点中查找
			// nodeOne := nodeStore.FindNode(one)
			// if nodeOne != nil {
			// 	vinfo = &Vnodeinfo{
			// 		Nid:   *one, //节点真实网络地址
			// 		Index: 0,    //节点第几个空间，从1开始,下标为0的节点为实际节点。
			// 		Vid:   temp, //vid，虚拟节点网络地址
			// 	}
			// 	vinfos = append(vinfos, *vinfo)
			// }
			continue
		}
		vinfos = append(vinfos, *vinfo)
		// fmt.Println("添加到列表中去 3333333333333333333333")
	}
	bs, _ := json.Marshal(vinfos)

	// fmt.Println("查找到的节点id", len(vinfos), string(bs))
	message_center.SendP2pReplyMsgHE(message, config.MSGID_vnode_getNearSuperIP_recv, &bs)

}

/*
	获取相邻节点的Vnode地址 返回
*/
func GetNearSuperAddr_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// fmt.Println("-----------------接收到邻居节点虚拟节点查询消息 返回")
	vnodeinfos := make([]virtual_node.Vnodeinfo, 0)

	// fmt.Println(string(*message.Body.Content))

	// err := json.Unmarshal(*message.Body.Content, &vnodeinfos)
	decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	decoder.UseNumber()
	err := decoder.Decode(&vnodeinfos)
	if err != nil {
		fmt.Println("不能解析json错误", err)
		return
	}
	for _, one := range vnodeinfos {
		// fmt.Println("添加一个虚拟节点")
		virtual_node.AddLogicVnodeinfo(one)
	}

	// for _, one := range virtual_node.GetVnodeLogical() {
	// 	fmt.Println("查询自己的逻辑虚拟节点", one.Vid.B58String())
	// }
}
