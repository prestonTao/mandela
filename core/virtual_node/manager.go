package virtual_node

import (
	// "mandela/config"
	// "mandela/core/message_center"
	// "mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"sync"
)

type VnodeManager struct {
	lock              *sync.RWMutex    //
	Vnodes            []Vnode          //多个虚拟节点
	findNearVnodeChan chan FindVnodeVO //需要查找的虚拟节点
}

func (this *VnodeManager) Run() {
	// fmt.Println("添加锁")
	this.lock = new(sync.RWMutex)
	this.findNearVnodeChan = make(chan FindVnodeVO, 1000)
	// Reg()
	//需要添加一个index=0的虚拟节点，映射为真实节点

	// vnode := NewVnode(0, nodeStore.NodeSelf.IdInfo.Id, this.findNearVnodeChan)
	// this.Vnodes = append(this.Vnodes, *vnode)

	// vnode = NewVnode(1, nodeStore.NodeSelf.IdInfo.Id, this.findNearVnodeChan)
	// this.Vnodes = append(this.Vnodes, *vnode)
	// vnode = NewVnode(2, nodeStore.NodeSelf.IdInfo.Id, this.findNearVnodeChan)
	// this.Vnodes = append(this.Vnodes, *vnode)
}

// /*
// 	添加一个新节点，发送消息看这个节点是否开通了虚拟节点
// 	已开通则添加这个节点，未开通则抛弃。
// */
// func (this *VnodeManager) AddNewNode(addr nodeStore.AddressNet) {

// 	go func() {
// 		this.lock.Lock()
// 		defer this.lock.Unlock()
// 		if len(this.Vnodes) <= 0 {
// 			return
// 		}

// 	}()

// }

/*
	添加一个虚拟节点
*/
func (this *VnodeManager) AddVnode() Vnodeinfo {
	this.lock.Lock()
	defer this.lock.Unlock()

	index := uint64(len(this.Vnodes))

	nodeOne := NewVnode(index, nodeStore.NodeSelf.IdInfo.Id, this.findNearVnodeChan)
	this.Vnodes = append(this.Vnodes, *nodeOne)
	return nodeOne.Vnode
}

/*
	删除一个虚拟节点
*/
func (this *VnodeManager) DelVnode() (nodeinfo Vnodeinfo) {
	this.lock.Lock()
	defer this.lock.Unlock()

	index := uint64(len(this.Vnodes))

	newvnodes := make([]Vnode, 0)

	for i, _ := range this.Vnodes {
		if uint64(i+1) >= index {
			nodeinfo = this.Vnodes[i].Vnode
			break
		}
		newvnodes = append(newvnodes, this.Vnodes[i])
	}
	this.Vnodes = newvnodes
	return
}

/*
	调整云存储大小，多了的就减少，少了的就增加。
*/
func (this *VnodeManager) SetupVnodeNumber(n uint64) {
	// if n <= 0 {
	// 	return
	// }
	this.lock.Lock()
	defer this.lock.Unlock()
	//空间大小合适，不需要调整
	if uint64(len(this.Vnodes)) == n {
		return
	}
	//空间太大，需要减少空间
	if uint64(len(this.Vnodes)) > n {
		newvnodes := make([]Vnode, 0)
		for i, _ := range this.Vnodes {
			if uint64(i+1) > n {
				break
			}
			newvnodes = append(newvnodes, this.Vnodes[i])
		}
		this.Vnodes = newvnodes
	} else {
		//空间太小，需要增加空间。
		for i := uint64(len(this.Vnodes)); i < n; i++ {
			nodeOne := NewVnode(i, nodeStore.NodeSelf.IdInfo.Id, this.findNearVnodeChan)
			this.Vnodes = append(this.Vnodes, *nodeOne)
		}
	}

}

/*
	查询云存储大小
*/
func (this *VnodeManager) GetVnodeNumber() []Vnodeinfo {
	return this.GetVnodeSelf()
}

/*
	添加虚拟节点的逻辑节点
*/
func (this *VnodeManager) AddLogicVnodeinfo(vnode Vnodeinfo) (ok bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ok = false
	for _, one := range this.Vnodes {
		if success := one.AddLogicVnodeinfo(vnode); success {
			ok = true
		}
	}
	return
}

/*
	获得所有节点，包括自己节点
*/
func (this *VnodeManager) GetVnodeLogical() map[string]Vnodeinfo {
	this.lock.RLock()
	defer this.lock.RUnlock()

	vnodeinfoMap := make(map[string]Vnodeinfo)
	for _, one := range this.Vnodes {
		selfOne := one.GetSelfVnodeinfo()
		vnodeinfoMap[selfOne.Vid.B58String()] = selfOne

		one.LogicalNode.Range(func(k, v interface{}) bool {
			vnodeinfo := v.(Vnodeinfo)
			vnodeinfoMap[vnodeinfo.Vid.B58String()] = vnodeinfo
			return true
		})
		// for j, two := range one.LogicalNode {
		// 	vnodeinfoMap[two.Vid.B58String()] = one.LogicalNode[j]
		// }
	}
	//删除自己节点
	for _, one := range this.Vnodes {
		delete(vnodeinfoMap, one.Vnode.Vid.B58String())

	}
	return vnodeinfoMap
}

/*
	获得自己管理的节点info
*/
func (this *VnodeManager) GetVnodeSelf() []Vnodeinfo {
	this.lock.RLock()
	defer this.lock.RUnlock()

	vnodeinfo := make([]Vnodeinfo, 0)
	for _, one := range this.Vnodes {
		vnodeinfo = append(vnodeinfo, one.Vnode)
		// //----------------------------
		// engine.Log.Info("----------------------------")
		// engine.Log.Info("自己节点id %s", one.Vnode.Vid.B58String())
		// one.LogicalNode.Range(func(k, v interface{}) bool {
		// 	engine.Log.Info("逻辑节点id %s", k.(string))
		// 	return true
		// })
		// //-----------------------
	}

	return vnodeinfo
}

/*
	在逻辑节点中查找Vnodeinfo
*/
func (this *VnodeManager) FindVnodeinfo(vid AddressNetExtend) *Vnodeinfo {
	this.lock.RLock()
	defer this.lock.RUnlock()

	for _, one := range this.Vnodes {
		vnodeinfo := one.FindVnodeinfo(vid)
		if vnodeinfo == nil {
			continue
		}
		return vnodeinfo
	}
	return nil
}

func (this *VnodeManager) Test() {

}
