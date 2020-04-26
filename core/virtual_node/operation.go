package virtual_node

import (
	"mandela/core/nodeStore"
	"math/big"
)

/*
	添加文件共享
*/
func AddFileShare() {

}

/*
	发送搜索节点消息
*/
// func SendSearchAllMsg(msgid uint64, recvid *nodeStore.AddressNet, content *[]byte) {
// 	// GetVnodeLogical()

// 	// msg, ok := message_center.SendP2pMsgHE(msgid, recvid, content)

// }

/*
	找到最近的虚拟节点
	@nodeId         要查找的节点
	@outId          排除一个节点
	@includeSelf    是否包括自己
	@return         查找到的节点id，可能为空
*/
func FindNearVnode(nodeId, outId *AddressNetExtend, includeSelf bool) AddressNetExtend {

	//获取本节点的所有逻辑节点
	vnodeMap := GetVnodeLogical()
	//包括自己，就添加自己的虚拟节点
	if includeSelf {
		vnodes := GetVnodeSelf()
		for _, v := range vnodes {
			vnodeMap[v.Vid.B58String()] = v
		}
	} else {
		//不包括自己，逻辑节点中有可能包括自己节点，则删除自己节点
		vnodes := GetVnodeSelf()
		for _, v := range vnodes {
			delete(vnodeMap, v.Vid.B58String())
		}
	}

	//有排除的节点，不添加
	delete(vnodeMap, outId.B58String())

	//构建kad算法，添加逻辑节点
	kl := nodeStore.NewKademlia()
	for _, v := range vnodeMap {
		kl.Add(new(big.Int).SetBytes(v.Vid))
	}

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	mh := AddressNetExtend(targetId.Bytes())
	return mh

	// outIdStr := ""
	// if outId != nil {
	// 	outIdStr = outId.B58String()
	// }
	// Nodes.Range(func(k, v interface{}) bool {
	// 	// if k.(string) == outIdStr {
	// 	// 	return true
	// 	// }
	// 	value := v.(* Node)
	// 	if bytes.Equal(value )
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })

	// targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	// if len(targetIds) == 0 {
	// 	return nil
	// }
	// targetId := targetIds[0]
	// if targetId == nil {
	// 	return nil
	// }
	// mh := AddressNet(targetId.Bytes())
	// return &mh
	// return nil
}

/*
	在自己的虚拟节点中找到最近的虚拟节点
	@nodeId         要查找的节点
*/
func FindNearVnodeInSelf(nodeId *AddressNetExtend) *AddressNetExtend {

	//构建kad算法，添加逻辑节点
	kl := nodeStore.NewKademlia()
	for _, v := range GetVnodeSelf() {
		kl.Add(new(big.Int).SetBytes(v.Vid))
	}

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	mh := AddressNetExtend(targetId.Bytes())
	return &mh
}
