package virtual_node

import (
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"bytes"
	"sync"
)

/*
	虚拟节点
*/
type Vnode struct {
	Vnode         Vnodeinfo        //自己的虚拟节点
	LogicalNode   *sync.Map        //逻辑节点 key:string=;value:Vnodeinfo=;
	findVnodeChan chan FindVnodeVO //
}

func (this *Vnode) Run() {
	go this.SearchVnode()
}

/*
	定时搜索其他虚拟节点
*/
func (this *Vnode) SearchVnode() {

	tiker := utils.NewBackoffTimer(config.VNODE_tiker_sync_logical_vnode...)
	for {
		// //先给自己真实邻居节点发送查询消息
		// for _, one := range nodeStore.GetLogicNodes() {
		// 	// tempid := nodeStore.AddressNet(*one)
		// 	targetVnode := BuildNodeinfo(0, *one)
		// 	fvv := FindVnodeVO{
		// 		Self:   this.Vnode,
		// 		Target: *targetVnode,
		// 	}

		// 	// fmt.Println("定时搜索逻辑")
		// 	select {
		// 	case this.findVnodeChan <- fvv:
		// 		// fmt.Println("放进去了")
		// 	default:
		// 		// fmt.Println("没有放进去")
		// 	}

		// 	time.Sleep(time.Second * 2)
		// }
		//再给自己的虚拟邻居节点发送查询消息
		this.LogicalNode.Range(func(k, v interface{}) bool {
			vnodeinfo := v.(Vnodeinfo)
			fvv := FindVnodeVO{
				Self:   this.Vnode,
				Target: vnodeinfo,
			}

			select {
			case this.findVnodeChan <- fvv:
			default:
			}

			tiker.Wait()
			return true
		})

	}
}

/*
	添加Vnode
*/
func (this *Vnode) AddLogicVnodeinfo(vnode Vnodeinfo) (ok bool) {

	//不能添加自己
	if bytes.Equal(vnode.Vid, this.Vnode.Vid) {
		// fmt.Println("自己不能添加自己")
		return false
	}

	idm := nodeStore.NewIds(this.Vnode.Vid, nodeStore.NodeIdLevel)
	this.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeinfo := v.(Vnodeinfo)
		idm.AddId(vnodeinfo.Vid)
		return true
	})

	ok, removeIDs := idm.AddId(vnode.Vid)
	if ok {
		this.LogicalNode.Store(vnode.Vid.B58String(), vnode)

		//		fmt.Println("添加成功", new(big.Int).SetBytes(node.IdInfo.Id.Data()).Int64())

		//删除被替换的id
		for _, one := range removeIDs {
			addrNetExtend := AddressNetExtend(one)
			this.LogicalNode.Delete(addrNetExtend.B58String())
		}
	}
	return true

}

/*
	获得自己的vnodeinfo
*/
func (this *Vnode) GetSelfVnodeinfo() Vnodeinfo {
	return this.Vnode
}

/*
	获得自己节点的所有逻辑节点，不包括自己节点
*/
func (this *Vnode) GetVnodeinfoAllNotSelf() []Vnodeinfo {
	// this.lock.RLock()
	// defer this.lock.RUnlock()
	vns := make([]Vnodeinfo, 0)
	this.LogicalNode.Range(func(k, v interface{}) bool {
		vnodeinfo := v.(Vnodeinfo)
		vns = append(vns, vnodeinfo)
		return true
	})
	return vns

}

/*
	查找Vnodeinfo
*/
func (this *Vnode) FindVnodeinfo(vid AddressNetExtend) *Vnodeinfo {
	value, ok := this.LogicalNode.Load(vid.B58String())
	if ok {
		vnodeinfo := value.(Vnodeinfo)
		return &vnodeinfo
	}
	return nil

	// vnodeinfo, ok := this.vnodesMap[vid.B58String()]
	// if ok {
	// 	return &vnodeinfo
	// }
	// return nil
}

func (this *Vnode) Test() {}

/*
	创建一个虚拟节点
*/
func NewVnode(index uint64, addrNet nodeStore.AddressNet, findNearVnodeChan chan FindVnodeVO) *Vnode {

	vnodeInfo := BuildNodeinfo(index, addrNet)

	vnode := Vnode{
		Vnode:       *vnodeInfo,    //自己的虚拟节点
		LogicalNode: new(sync.Map), // make([]Vnodeinfo, 0),       //逻辑节点
		// lock:          new(sync.RWMutex),          //
		// vnodesMap:     make(map[string]Vnodeinfo), //
		findVnodeChan: findNearVnodeChan, //
	}
	vnode.Run()
	return &vnode
}

/*
	查找虚拟节点
*/
type FindVnodeVO struct {
	Self   Vnodeinfo //自己节点
	Target Vnodeinfo //目标节点
}
