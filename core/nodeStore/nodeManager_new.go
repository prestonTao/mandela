package nodeStore

//import (
//	"encoding/hex"
//	"fmt"
//	"math/big"
//	"strconv"
//	"sync"
//	"time"
//	gconfig "mandela/config"
//	"mandela/core/utils"
//)

//var Manager *NodeManager

//type NodeManager struct {
//	NodeIdLevel      uint                  //节点id长度
//	NodeSelf         *Node                 //保存自己的id信息和ip地址和端口号
//	nodes            *sync.Map             //超级节点，id字符串为键 key:string,value:*Node
//	Proxys           *sync.Map             //被代理的节点，id字符串为键 key:string,value:*Node
//	OutFindNode      chan *utils.Multihash //需要查询的逻辑节点
//	OutCloseConnName chan *utils.Multihash //废弃的nodeid，需要询问是否关闭
//	SuperPeerId      *utils.Multihash      //超级节点名称
//	once             sync.Once
//}

//func InitNodeStore(node *Node) {
//	//	once.Do(run)
//	Manager = NewNodeManager(gconfig.NodeIDLevel, node)
//}

///*
//	添加一个代理节点
//*/
//func (this *NodeManager) AddProxyNode(node *Node) {
//	this.Proxys.Store(node.IdInfo.Id.B58String(), node)
//}

////得到一个代理节点
//func (this *NodeManager) GetProxyNode(id string) (node *Node, ok bool) {
//	var v interface{}
//	v, ok = this.Proxys.Load(id)
//	if v == nil {
//		return nil, ok
//	}
//	node = v.(*Node)
//	return
//}

///*
//	获得所有代理节点
//*/
//func (this *NodeManager) GetProxyAll() []string {
//	ids := make([]string, 0)
//	this.Proxys.Range(func(key, value interface{}) bool {
//		ids = append(ids, key.(string))
//		return true
//	})
//	return ids
//}

///*
//	添加一个超级节点
//	检查这个节点是否是自己的逻辑节点，如果是，则保存
//	不保存自己
//*/
//func (this *NodeManager) AddNode(node *Node) {
//	//	fmt.Println("添加一个节点", new(big.Int).SetBytes(node.IdInfo.Id.Data()).Int64())

//	//是本身节点不添加
//	if node.IdInfo.Id.B58String() == this.NodeSelf.IdInfo.Id.B58String() {
//		return
//	}

//	idm := NewIds(this.NodeSelf.IdInfo.Id, this.NodeIdLevel)
//	ids := this.GetAllNodes()
//	for _, one := range ids {
//		idm.AddId(one)
//	}
//	//	idm.AddId(node.IdInfo.Id)

//	ok, removeIDs := idm.AddId(node.IdInfo.Id)
//	if ok {
//		//		fmt.Println("添加成功", new(big.Int).SetBytes(node.IdInfo.Id.Data()).Int64())
//		node.lastContactTimestamp = time.Now()
//		this.nodes.Store(node.IdInfo.Id.B58String(), node)
//		//修改超级节点，普通节点经常切换影响网络
//		this.SuperPeerId = idm.GetIndex(0)

//		//删除被替换的id
//		for _, one := range removeIDs {
//			//			idOne := hex.EncodeToString(one)
//			//			delete(nodes, idOne)
//			//			OutCloseConnName <- idOne

//			this.nodes.Delete(one.B58String())
//			this.OutCloseConnName <- one
//		}
//	}
//	return
//}

///*
//	删除一个节点，包括超级节点和代理节点
//*/
//func (this *NodeManager) DelNode(id *utils.Multihash) {
//	this.nodes.Delete(id.B58String())
//	this.Proxys.Delete(id.B58String())
//}

///*
//	通过id查找一个节点
//*/
//func (this *NodeManager) FindNode(id *utils.Multihash) *Node {
//	v, ok := this.nodes.Load(id.B58String())
//	if ok {
//		return v.(*Node)
//	}
//	v, ok = this.Proxys.Load(id.B58String())
//	if ok {
//		return v.(*Node)
//	}
//	return nil
//}

///*
//	在超级节点中找到最近的节点，不包括代理节点
//	@nodeId         要查找的节点
//	@outId          排除一个节点
//	@includeSelf    是否包括自己
//	@return         查找到的节点id，可能为空
//*/
//func (this *NodeManager) FindNearInSuper(nodeId, outId *utils.Multihash, includeSelf bool) *utils.Multihash {
//	kl := NewKademlia()
//	if includeSelf {
//		kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id.Data()))
//	}
//	outIdStr := ""
//	if outId != nil {
//		outIdStr = outId.B58String()
//	}
//	this.nodes.Range(func(k, v interface{}) bool {
//		if k.(string) == outIdStr {
//			return true
//		}
//		value := v.(*Node)
//		//		fmt.Println("+", value.IdInfo.Id.B58String())
//		kl.Add(new(big.Int).SetBytes(value.IdInfo.Id.Data()))
//		return true
//	})

//	targetIds := kl.Get(new(big.Int).SetBytes(nodeId.Data()))
//	if len(targetIds) == 0 {
//		return nil
//	}
//	//	for _, one := range targetIds {
//	//		idbs, _ := utils.Encode(one.Bytes(), utils.SHA1)
//	//		idmh := utils.Multihash(idbs)
//	//		fmt.Println("-", idmh.B58String())
//	//	}
//	targetId := targetIds[0]
//	if targetId == nil {
//		return nil
//	}
//	bs, _ := utils.Encode(targetId.Bytes(), utils.SHA1)
//	mh := utils.Multihash(bs)
//	return &mh
//}

////得到所有的节点，不包括本节点，也不包括代理节点
//func (this *NodeManager) GetAllNodes() []*utils.Multihash {
//	ids := make([]*utils.Multihash, 0)
//	this.nodes.Range(func(k, v interface{}) bool {
//		value := v.(*Node)
//		ids = append(ids, value.IdInfo.Id)
//		return true
//	})
//	return ids
//}

///*
//	获得本机所有逻辑节点的ip地址
//*/
//func (this *NodeManager) GetSuperNodeIps() (ips []string) {
//	ips = make([]string, 0)
//	this.nodes.Range(func(k, v interface{}) bool {
//		value := v.(*Node)
//		ips = append(ips, value.Addr+":"+strconv.Itoa(int(value.TcpPort)))
//		return true
//	})

//	return
//}

///*
//	检查节点是否是本节点的逻辑节点
//	只检查，不保存
//*/
//func (this *NodeManager) CheckNeedNode(nodeId *utils.Multihash) (isNeed bool) {
//	/*
//		1.找到已有节点中与本节点最近的节点
//		2.计算两个节点是否在同一个网络
//		3.若在同一个网络，计算谁的值最小
//	*/
//	if len(this.GetAllNodes()) == 0 {
//		return true
//	}
//	//是本身节点不添加
//	//	if hex.EncodeToString(nodeId) == NodeSelf.IdInfo.Id.GetIdStr() {
//	if nodeId.B58String() == this.NodeSelf.IdInfo.Id.B58String() {
//		//		fmt.Println("2不添加")
//		return false
//	}

//	ids := NewIds(this.NodeSelf.IdInfo.Id, this.NodeIdLevel)
//	for _, one := range this.GetAllNodes() {
//		ids.AddId(one)
//	}
//	ok, _ := ids.AddId(nodeId)
//	return ok

//}

///*
//	获得一个节点更远的节点中，比自己更远的节点
//*/
//func (this *NodeManager) GetIdsForFar(id *utils.Multihash) []*utils.Multihash {
//	kl := NewKademlia()
//	kl.Add(new(big.Int).SetBytes(this.NodeSelf.IdInfo.Id.Data()))
//	//	nodesLock.RLock()
//	//	for _, value := range nodes {
//	//		kl.add(new(big.Int).SetBytes(*value.IdInfo.Id.GetId()))
//	//	}
//	//	nodesLock.RUnlock()
//	this.nodes.Range(func(k, v interface{}) bool {
//		value := v.(*Node)
//		kl.Add(new(big.Int).SetBytes(value.IdInfo.Id.Data()))
//		return true
//	})
//	list := kl.Get(new(big.Int).SetBytes(id.Data()))
//	out := make([]*utils.Multihash, 0)
//	for i, one := range list {
//		if hex.EncodeToString(one.Bytes()) == this.NodeSelf.IdInfo.Id.HexString() {
//			for j := i + 1; j < len(list); j++ {
//				bs, err := utils.Encode(list[j].Bytes(), utils.SHA1)
//				if err != nil {
//					fmt.Println("编码失败")
//					continue
//				}
//				mh := utils.Multihash(bs)
//				out = append(out, &mh)
//			}
//			break
//		}
//	}
//	return out
//}

//func NewNodeManager(level uint, root *Node) *NodeManager {
//	return &NodeManager{
//		NodeIdLevel:      level,                             //节点id长度
//		NodeSelf:         root,                              //保存自己的id信息和ip地址和端口号
//		nodes:            new(sync.Map),                     //超级节点，id字符串为键 key:string,value:*Node
//		OutFindNode:      make(chan *utils.Multihash, 1000), //需要查询的逻辑节点
//		Proxys:           new(sync.Map),                     //被代理的节点，id字符串为键 key:string,value:*Node
//		SuperPeerId:      nil,                               //超级节点名称
//		OutCloseConnName: make(chan *utils.Multihash, 1000), //废弃的nodeid，需要询问是否关闭
//	}
//}
