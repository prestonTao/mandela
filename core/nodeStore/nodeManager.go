package nodeStore

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type NodeClass int

const (
	Node_min = 0 //每个节点的最少连接数量

	Node_type_all        NodeClass = 0 //包含所有类型
	Node_type_logic      NodeClass = 1 //自己需要的逻辑节点
	Node_type_client     NodeClass = 2 //保存其他逻辑节点连接到自己的节点，都是超级节点
	Node_type_proxy      NodeClass = 3 //被代理的节点
	Node_type_other      NodeClass = 4 //每个节点有最少连接数量
	Node_type_white_list NodeClass = 5 //连接白名单
)

var (
	NodeSelf *Node = new(Node) //保存自己的id信息和ip地址和端口号

	nodesLock = new(sync.RWMutex)      //
	nodes     = make(map[string]*Node) //保存所有类型的节点，通过Type参数区分开

	// Nodes                             = new(sync.Map)                //超级节点。key:string=节点id,value:*Node=节点信息;
	// Proxys                            = new(sync.Map)                //被代理的节点。key:string=节点id;value:*Node=节点信息;
	// OtherNodes                        = new(sync.Map)                //每个节点有最少连接数量。key:string=节点id;value:*AddressNet=节点地址;
	// NodesClient                       = new(sync.Map)                //保存其他逻辑节点连接到自己的节点，都是超级节点
	OutFindNode      chan *AddressNet = make(chan *AddressNet, 1000) //需要查询的逻辑节点
	OutCloseConnName                  = make(chan *AddressNet, 1000) //废弃的nodeid，需要询问是否关闭
	SuperPeerId      *AddressNet                                     //超级节点名称
	NodeIdLevel      uint             = 256                          //节点id长度
	HaveNewNode                       = make(chan *Node, 100)        //当添加新的超级节点时，给他个信号
	// LanNode                           = new(sync.Map)                //局域网节点。key:string=节点id;value:*Node=节点信息;
	// Groups           *NodeGroup       = NewNodeGroup()               //组
	// once             sync.Once
	// Key              *ECCKey
	WhiteList = new(sync.Map) //连接白名单
)

func Init() int {
	//加载本地网络id
	idinfo, err := BuildIdinfo(config.Wallet_keystore_default_pwd)
	if err != nil {
		fmt.Println(err)
		return 1
	}
	NodeSelf.IdInfo = *idinfo
	NodeSelf.MachineID = utils.GetRandomOneInt64()
	NodeSelf.Version = config.Version_0
	// fmt.Println("id =", idinfo.Id.B58String())
	return 0

}

/*
	加载本地私钥生成idinfo
*/
func BuildIdinfo(pwd string) (*IdInfo, error) {
	fileAbsPath := filepath.Join(config.Path_configDir, config.Core_keystore)

	exist, err := utils.PathExists(fileAbsPath)
	if err != nil {
		return nil, err
	}
	if exist {
		err := keystore.Load(fileAbsPath)
		if err != nil {
			return nil, err
		}
	} else {
		err = keystore.CreateKeystore(fileAbsPath, pwd)
		if err != nil {
			return nil, err
		}
	}

	// var dhPuk dh.PublicKey
	keyPair := keystore.GetDHKeyPair()

	// cpuk := keystore.GetDHKeyPair().KeyPair.GetPublicKey()
	// dhPuk := keyPair.KeyPair.PublicKey
	// fmt.Println("cpuk:", hex.EncodeToString(dhPuk[:]), hex.EncodeToString(cpuk[:]))

	prk, puk, err := keystore.GetNetAddr(pwd)

	// addr := keystore.GetCoinbase()
	// _, prk, puk, err := keystore.GetKeyByAddr(addr.Addr, pwd)
	if err != nil {
		return nil, err
	}
	addrNet := BuildAddr(puk)
	// sign := ed25519.Sign(prk, append(dhPuk[:], keyPair.Index))

	// pair, err := dh.GenerateKeyPair(rand)
	// if err != nil {
	// 	return nil, err
	// }

	idinfo := IdInfo{
		Id:   addrNet,                   //id，网络地址，临时地址，由updatetime字符串产生MD5值
		EPuk: puk,                       //ed25519公钥
		CPuk: keyPair.KeyPair.PublicKey, //curve25519公钥
		V:    uint32(keyPair.Index),     //dh密钥版本
		// Sign: sign,                      //ed25519私钥签名
	}
	idinfo.SignDHPuk(prk)
	return &idinfo, nil
}

//超级节点之间查询的间隔时间
//var SpacingInterval time.Duration = time.Second * 30

//id字符串格式为16进制字符串
//var IdStrBit int = 16

func InitNodeStore() int {
	return Init()
	// once.Do(run)
}

/*
	定期检查所有节点状态
	一个小时查询所有逻辑节点，非超级节点查询相邻节点
	5分钟清理一次已经不在线的节点
*/
// func run() {
// 	//	fmt.Println("启动循环查询逻辑节点")
// 	go func() {

// 		//		bt := utils.NewBackoffTimer(config.Time_find_network_peer...)
// 		//		//查询和自己相关的逻辑节点
// 		//		for {
// 		//			ids := getNodeNetworkNum()
// 		//			//			if !NodeSelf.IsSuper {
// 		//			//				ids = ids[:3]
// 		//			//			}
// 		//			//			ids = ids[:8]
// 		//			for _, idOne := range ids {
// 		//				OutFindNode <- idOne
// 		//				time.Sleep(time.Second * 1)
// 		//			}
// 		//			//			fmt.Println("完成一轮查找")
// 		//			bt.Wait()
// 		//		}
// 	}()
// }

//func initIds() {
//	if netNodes == nil {
//		netNodes = NewIds(NodeSelf.IdInfo.Id, NodeIdLevel)
//	}
//}

//添加一个代理节点
func AddProxyNode(node Node) {
	// Proxys.Store(node.IdInfo.Id.B58String(), node)
	engine.Log.Info("add proxys nodeid %s", node.IdInfo.Id.B58String())
	node.Type = Node_type_proxy
	nodesLock.Lock()
	nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
	nodesLock.Unlock()
	// Proxys.Store(utils.Bytes2string(node.IdInfo.Id), &node)
}

//得到一个代理节点
func GetProxyNode(id string) (node *Node, ok bool) {

	nodesLock.RLock()
	node, ok = nodes[id]
	if node != nil && node.Type != Node_type_proxy {
		node = nil
		ok = false
	}
	nodesLock.RUnlock()

	return

	// var v interface{}
	// v, ok = Proxys.Load(id)
	// if v == nil {
	// 	return nil, ok
	// }
	// node = v.(*Node)
	// return
}

/*
	局域网连接不能直接保持，比如一个局域网有255台电脑，每台电脑有其他电脑的连接，一共有255 * 255 = 65025个连接。
*/
/*
	添加一个局域网连接
*/
// func AddLanNode(node *Node) {
// 	LanNode.Store(node.IdInfo.Id.B58String(), node)
// }

// /*
// 	查询一个局域网节点
// */
// func GetLanNode(id string) (node *Node, ok bool) {
// 	var v interface{}
// 	v, ok = LanNode.Load(id)
// 	if v == nil {
// 		return nil, ok
// 	}
// 	node = v.(*Node)
// 	return
// }

// /*
// 	删除一个局域网节点
// */
// func DelLanNode(id AddressNet) {
// 	LanNode.Delete(id.B58String())
// 	engine.RemoveSession(id.B58String())
// }

/*
	添加一个超级节点
	检查这个节点是否是自己的逻辑节点，如果是，则保存
	不保存自己
*/
func AddNode(node Node) bool {
	//	initIds()
	engine.Log.Info("add node: %s", node.IdInfo.Id.B58String())

	//不添加自己
	if bytes.Equal(node.IdInfo.Id, NodeSelf.IdInfo.Id) {
		return false
	}

	//检查是否够最少连接数量
	// total := 0
	// OtherNodes.Range(func(k, v interface{}) bool {
	// 	total++
	// 	return true
	// })
	// //不够就保存
	// if total < Node_min {
	// 	//保存
	// 	OtherNodes.Store(utils.Bytes2string(node.IdInfo.Id), node)
	// }

	//检查是否在白名单中
	if FindWhiteList(&node.IdInfo.Id) {
		nodesLock.Lock()
		node.Type = Node_type_white_list
		nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
		nodesLock.Unlock()
		return true
	}

	idm := NewIds(NodeSelf.IdInfo.Id, NodeIdLevel)
	// ids := GetLogicNodes()

	nodesLock.Lock()

	//检查是否够最少连接数量
	total := 0
	for _, one := range nodes {
		if one.Type != Node_type_logic {
			continue
		}
		total += 1
		// ids = append(ids, one.IdInfo.Id)
		idm.AddId(one.IdInfo.Id)
	}
	//不够就保存
	if total < Node_min {
		node.Type = Node_type_logic
		nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
		nodesLock.Unlock()
		return true
	}

	// for _, one := range ids {
	// 	idm.AddId(one)
	// }

	ok, removeIDs := idm.AddId(node.IdInfo.Id)
	if ok {
		//		fmt.Println("添加成功", new(big.Int).SetBytes(node.IdInfo.Id.Data()).Int64())
		node.lastContactTimestamp = time.Now()
		// engine.Log.Info("Nodes super add: %s %d", node.IdInfo.Id.B58String(), len(removeIDs))

		nodeLoad, ok := nodes[utils.Bytes2string(node.IdInfo.Id)]
		if ok {
			nodeLoad.Type = Node_type_logic
		} else {
			node.Type = Node_type_logic
			nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
		}

		// Nodes.Store(utils.Bytes2string(node.IdInfo.Id), &node)

		select {
		case HaveNewNode <- &node:
		default:
		}

		//修改超级节点，普通节点经常切换影响网络
		addrNet := AddressNet(idm.GetIndex(0))
		SuperPeerId = &addrNet

		//删除被替换的id
		for _, one := range removeIDs {
			//			idOne := hex.EncodeToString(one)
			//			delete(Nodes, idOne)
			//			OutCloseConnName <- idOne
			addrNet := AddressNet(one)
			addrStr := utils.Bytes2string(addrNet)
			//如果自己是对方节点的逻辑节点，则不删除，保存到ClientNodes中

			nodeLoad, ok := nodes[addrStr]
			if ok {
				if nodeLoad.Type == Node_type_white_list {
					continue
				}
				// engine.Log.Info("add client nodeid %s", node.IdInfo.Id.B58String())
				nodeLoad.Type = Node_type_client
			}
			if !ok {
				continue
			}

			// nodeOne, ok := Nodes.Load(addrStr)
			// if !ok {
			// 	continue
			// }
			// Nodes.Delete(addrStr)
			// NodesClient.Store(addrStr, nodeOne)

			// engine.Log.Info("Nodes super del: %s", addrNet.B58String())

			//如果排除的节点在自己的白名单中，就不询问关闭连接了
			if FindWhiteList(&addrNet) {
				continue
			}
			//询问对方，自己是否是对方的逻辑节点，如果是则保留连接，如果不是，则关闭连接
			select {
			case OutCloseConnName <- &addrNet:
			default:
			}
			//TODO 以上是询问对方的方式自治网络，需要对方配合，如果对方不配合，可以直接关闭连接，让对方重新连接自己

		}

	}
	nodesLock.Unlock()
	//	fmt.Println("添加一个node", node.IdInfo.Id.B58String())
	return ok

}

/*
	删除一个节点，包括超级节点和代理节点
*/
func DelNode(id *AddressNet) {
	idStr := utils.Bytes2string(*id) //  id.B58String()
	engine.Log.Info("delete client nodeid %s", id.B58String())

	nodesLock.Lock()

	delete(nodes, idStr)

	nodesLock.Unlock()

	// // engine.Log.Info("Nodes super del: %s", id.B58String())
	// Nodes.Delete(idStr)
	// // engine.Log.Info("delete proxys nodeid %s", id.B58String())
	// Proxys.Delete(idStr)
	// // Proxys.Range(func(k, v interface{}) bool {
	// // 	keyStr := k.(string)
	// // 	id := AddressNet([]byte(keyStr))
	// // 	engine.Log.Info("proxys list one %s", id.B58String())
	// // 	return true
	// // })
	// OtherNodes.Delete(idStr)
	// NodesClient.Delete(idStr)
	engine.RemoveSession(idStr)
}

/*
	通过id查找一个节点
*/
func FindNode(id *AddressNet) *Node {
	idStr := utils.Bytes2string(*id)
	// engine.Log.Info("Nodes super add: %s", id.B58String())

	var node *Node
	// var ok bool
	nodesLock.RLock()
	node, _ = nodes[idStr]
	nodesLock.RUnlock()
	return node
	// v, ok := Nodes.Load(idStr)
	// if ok {
	// 	return v.(*Node)
	// }
	// v, ok = OtherNodes.Load(idStr)
	// if ok {
	// 	return v.(*Node)
	// }

	// v, ok = Proxys.Load(idStr)
	// if ok {
	// 	return v.(*Node)
	// }

	// v, ok = NodesClient.Load(idStr)
	// if ok {
	// 	return v.(*Node)
	// }
	// return nil
}

func FindNodeInLogic(id *AddressNet) *Node {
	idStr := utils.Bytes2string(*id)

	var node *Node
	// var ok bool
	nodesLock.RLock()
	node, _ = nodes[idStr]
	nodesLock.RUnlock()
	if node != nil && node.Type != Node_type_logic && node.Type != Node_type_white_list {
		return nil
	}
	return node

	// v, ok := Nodes.Load(idStr)
	// if ok {
	// 	return v.(*Node)
	// }
	// return nil
}

/*
	对比逻辑节点是否有变化
*/
func EqualLogicNodes(ids []AddressNet) bool {
	newLogicNodes := GetLogicNodes()
	if len(ids) != len(newLogicNodes) {
		return true
	}
	isChange := false
	nodesLock.RLock()
	for _, one := range ids {
		idStr := utils.Bytes2string(one)
		// _, ok := Nodes.Load(idStr)
		_, ok := nodes[idStr]
		if !ok {
			isChange = true
			break
			// return true
		}
	}
	nodesLock.RUnlock()
	return isChange
}

/*
	在超级节点中找到最近的节点，不包括代理节点
	@nodeId         要查找的节点
	@outId          排除一个节点
	@includeSelf    是否包括自己
	@return         查找到的节点id，可能为空
*/
func FindNearInSuper(nodeId, outId *AddressNet, includeSelf bool) *AddressNet {
	kl := NewKademlia(len(nodes) + 1)
	if includeSelf {
		kl.Add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
	}
	// outIdStr := ""
	// if outId != nil {
	// 	outIdStr = utils.Bytes2string(*outId)
	// }

	nodesLock.RLock()
	for _, one := range nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		if outId != nil && bytes.Equal(one.IdInfo.Id, *outId) {
			continue
		}
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
	}
	nodesLock.RUnlock()

	// Nodes.Range(func(k, v interface{}) bool {
	// 	if k.(string) == outIdStr {
	// 		return true
	// 	}
	// 	value := v.(*Node)
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	mh := AddressNet(targetId.Bytes())
	return &mh
}

//在节点中找到最近的节点，包括代理节点
func FindNearNodeId(nodeId, outId *AddressNet, includeSelf bool) *AddressNet {
	kl := NewKademlia(len(nodes) + 1)
	if includeSelf {
		kl.Add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
	}
	// outIdStr := ""
	// if outId != nil {
	// 	outIdStr = utils.Bytes2string(*outId) // outId.B58String()
	// }

	nodesLock.RLock()

	for _, one := range nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_proxy && one.Type != Node_type_white_list {
			continue
		}
		if outId != nil && bytes.Equal(one.IdInfo.Id, *outId) {
			continue
		}
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
	}

	nodesLock.RUnlock()

	// Nodes.Range(func(k, v interface{}) bool {
	// 	if k.(string) == outIdStr {
	// 		return true
	// 	}
	// 	value := v.(*Node)
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })
	// //代理节点
	// Proxys.Range(func(k, v interface{}) bool {
	// 	if k.(string) == outIdStr {
	// 		return true
	// 	}
	// 	value := v.(*Node)
	// 	//过滤APP节点
	// 	if value.IsApp {
	// 		return true
	// 	}
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })

	targetIds := kl.Get(new(big.Int).SetBytes(*nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	mh := AddressNet(targetId.Bytes())
	return &mh
}

/*
	根据节点id得到一个距离最短节点的信息，不包括代理节点
	@nodeId         要查找的节点
	@includeSelf    是否包括自己
	@outId          排除一个节点
	@return         查找到的节点id，可能为空
*/
//func Get(nodeId string, includeSelf bool, outId string) *Node {
//	nodeIdInt, b := new(big.Int).SetString(nodeId, IdStrBit)
//	if !b {
//		fmt.Println("节点id格式不正确，应该为十六进制字符串:")
//		fmt.Println(nodeId)
//		return nil
//	}
//	kl := NewKademlia()
//	if includeSelf {
//		//		temp := new(big.Int).SetBytes(Root.IdInfo.Id)
//		kl.add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
//	}
//	for key, value := range Nodes {
//		if outId != "" && key == outId {
//			continue
//		}
//		kl.add(new(big.Int).SetBytes(value.IdInfo.Id))
//	}
//	// TODO 不安全访问
//	targetId := kl.get(nodeIdInt)[0]

//	if targetId == nil {
//		return nil
//	}
//	if hex.EncodeToString(targetId.Bytes()) == hex.EncodeToString(NodeSelf.IdInfo.Id) {
//		return NodeSelf
//	}
//	return Nodes[hex.EncodeToString(targetId.Bytes())]
//}

/*
	在连接中获得白名单节点
*/
func GetWhiltListNodes() []AddressNet {
	ids := make([]AddressNet, 0)
	nodesLock.RLock()
	for _, one := range nodes {
		if one.Type != Node_type_white_list {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	nodesLock.RUnlock()
	return ids
}

//得到所有逻辑节点，不包括本节点，也不包括代理节点
func GetLogicNodes() []AddressNet {
	ids := make([]AddressNet, 0)
	nodesLock.RLock()
	for _, one := range nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	nodesLock.RUnlock()
	return ids
}

/*
	获得所有代理节点
*/
func GetProxyAll() []AddressNet {
	// ids := make([]string, 0)
	// Proxys.Range(func(key, value interface{}) bool {
	// 	ids = append(ids, key.(string))
	// 	return true
	// })
	// return ids

	ids := make([]AddressNet, 0)
	nodesLock.RLock()
	for _, one := range nodes {
		if one.Type != Node_type_proxy {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	nodesLock.RUnlock()
	// Proxys.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	ids = append(ids, value.IdInfo.Id)
	// 	return true
	// })
	return ids
}

/*
	获取额外的节点连接
*/
// func GetOtherNodes() []AddressNet {
// 	ids := make([]AddressNet, 0)
// 	OtherNodes.Range(func(k, v interface{}) bool {
// 		value := v.(*Node)
// 		ids = append(ids, value.IdInfo.Id)
// 		return true
// 	})
// 	return ids
// }

/*
	添加一个被其他节点当作逻辑节点的连接
*/
func AddNodesClient(node Node) {
	engine.Log.Info("add client nodeid %s", node.IdInfo.Id.B58String())
	node.Type = Node_type_client
	nodesLock.Lock()
	nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
	nodesLock.Unlock()
	// NodesClient.Store(utils.Bytes2string(node.IdInfo.Id), &node)
	select {
	case HaveNewNode <- &node:
	default:
	}
}

/*
	将节点转化为逻辑节点
*/
// func SwitchNodesClientToLogic(node Node) {
// 	node.Type = Node_type_super
// 	engine.Log.Info("SwitchNodesClientToLogic %s", node.IdInfo.Id.B58String())
// 	nodesLock.Lock()
// 	nodeLoad, ok := nodes[utils.Bytes2string(node.IdInfo.Id)]
// 	if ok {
// 		nodeLoad.Type = Node_type_super
// 	} else {
// 		nodes[utils.Bytes2string(node.IdInfo.Id)] = &node
// 	}
// 	nodesLock.Unlock()
// 	// NodesClient.Delete(utils.Bytes2string(node.IdInfo.Id))
// 	// Nodes.Store(utils.Bytes2string(node.IdInfo.Id), &node)
// }

/*
	获取本机被其他当作逻辑节点的连接
*/
func GetNodesClient() []AddressNet {
	ids := make([]AddressNet, 0)
	nodesLock.RLock()
	for _, one := range nodes {
		if one.Type != Node_type_client {
			continue
		}
		ids = append(ids, one.IdInfo.Id)
	}
	nodesLock.RUnlock()
	// NodesClient.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	ids = append(ids, value.IdInfo.Id)
	// 	return true
	// })
	return ids
}

/*
	获得本机所有逻辑节点的ip地址
*/
func GetSuperNodeIps() (ips []string) {
	ips = make([]string, 0)
	nodesLock.RLock()
	for _, one := range nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		// ids = append(ids, one.IdInfo.Id)
		ips = append(ips, one.Addr+":"+strconv.Itoa(int(one.TcpPort)))
	}
	nodesLock.RUnlock()

	// Nodes.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	ips = append(ips, value.Addr+":"+strconv.Itoa(int(value.TcpPort)))
	// 	return true
	// })

	return
}

/*
	检查节点是否是本节点的逻辑节点
	只检查，不保存
*/
func CheckNeedNode(nodeId *AddressNet) (isNeed bool) {
	/*
		1.找到已有节点中与本节点最近的节点
		2.计算两个节点是否在同一个网络
		3.若在同一个网络，计算谁的值最小
	*/

	if len(GetLogicNodes()) == 0 {
		return true
	}
	//是本身节点不添加
	if bytes.Equal(*nodeId, NodeSelf.IdInfo.Id) {
		return false
	}

	ids := NewIds(NodeSelf.IdInfo.Id, NodeIdLevel)
	for _, one := range GetLogicNodes() {
		ids.AddId(one)
	}
	ok, _ := ids.AddId(*nodeId)
	return ok
}

/*
	检查连接到本机的节点是否可以作为逻辑节点
*/
func CheckClientNodeIsLogicNode() {
	nodesLock.Lock()
	for _, one := range nodes {
		if one.Type != Node_type_client {
			continue
		}
		//是本身节点不添加
		if bytes.Equal(one.IdInfo.Id, NodeSelf.IdInfo.Id) {
			continue
		}
		//在白名单中
		if FindWhiteList(&one.IdInfo.Id) {
			one.Type = Node_type_white_list
			continue
		}

		logicNodes := make([]AddressNet, 0)
		for _, two := range nodes {
			if two.Type != Node_type_logic {
				continue
			}
			logicNodes = append(logicNodes, two.IdInfo.Id)
		}
		//检查是否需要
		if len(logicNodes) == 0 {
			one.Type = Node_type_logic
			continue
		}

		ids := NewIds(NodeSelf.IdInfo.Id, NodeIdLevel)
		for _, two := range logicNodes {
			ids.AddId(two)
		}
		ok, _ := ids.AddId(one.IdInfo.Id)
		if ok {
			one.Type = Node_type_logic
		}
	}
	nodesLock.Unlock()
	// for _, one := range clientNodes {
	// 	if nodeStore.CheckNeedNode(&one) {
	// 		node := nodeStore.FindNode(&one)
	// 		nodeStore.SwitchNodesClientToLogic(*node)
	// 	}
	// }
}

type LogicNumBuider struct {
	lock  *sync.RWMutex
	id    *[]byte
	level uint
	idStr string
	ids   []*[]byte
}

/*
	得到每个节点网络的网络号，不包括本节点
	@id        *utils.Multihash    要计算的id
	@level     int                 深度
*/
func (this *LogicNumBuider) GetNodeNetworkNum() []*[]byte {

	this.lock.RLock()
	if this.idStr != "" && this.idStr == hex.EncodeToString(*this.id) {
		this.lock.RUnlock()
		return this.ids
	}
	this.lock.RUnlock()

	this.lock.Lock()
	this.idStr = hex.EncodeToString(*this.id) // .B58String()

	root := new(big.Int).SetBytes(*this.id)

	this.ids = make([]*[]byte, 0)
	for i := 0; i < int(this.level); i++ {
		//---------------------------------
		//将后面的i位置零
		//---------------------------------
		//		startInt := new(big.Int).Lsh(new(big.Int).Rsh(root, uint(i)), uint(i))
		//---------------------------------
		//第i位取反
		//---------------------------------
		networkNum := new(big.Int).Xor(root, new(big.Int).Lsh(big.NewInt(1), uint(i)))

		mhbs := networkNum.Bytes()

		this.ids = append(this.ids, &mhbs)
	}
	this.lock.Unlock()

	return this.ids
}

func NewLogicNumBuider(id []byte, level uint) *LogicNumBuider {
	return &LogicNumBuider{
		lock:  new(sync.RWMutex),
		id:    &id,
		level: level,
		idStr: "",
		ids:   make([]*[]byte, 0),
	}
}

// var (
// 	networkIDsLock   = new(sync.RWMutex)
// 	nodeNetworkIDStr = ""
// 	networkIDs       []*AddressNet
// )

// //得到每个节点网络的网络号，不包括本节点
// func getNodeNetworkNum() []*AddressNet {
// 	networkIDsLock.RLock()
// 	if nodeNetworkIDStr != "" && nodeNetworkIDStr == NodeSelf.IdInfo.Id.B58String() {
// 		networkIDsLock.RUnlock()
// 		return networkIDs
// 	}
// 	networkIDsLock.RUnlock()

// 	// rootInt, _ := new(big.Int).SetString(, IdStrBit)
// 	networkIDsLock.Lock()
// 	nodeNetworkIDStr = NodeSelf.IdInfo.Id.B58String()

// 	root := new(big.Int).SetBytes(NodeSelf.IdInfo.Id)

// 	networkIDs = make([]*AddressNet, 0)
// 	for i := 0; i < int(NodeIdLevel); i++ {
// 		//---------------------------------
// 		//将后面的i位置零
// 		//---------------------------------
// 		//		startInt := new(big.Int).Lsh(new(big.Int).Rsh(root, uint(i)), uint(i))
// 		//---------------------------------
// 		//第i位取反
// 		//---------------------------------
// 		networkNum := new(big.Int).Xor(root, new(big.Int).Lsh(big.NewInt(1), uint(i)))

// 		// bs, err := utils.Encode(networkNum.Bytes(), config.HashCode)
// 		// if err != nil {
// 		// 	// fmt.Println("格式化muhash错误")
// 		// 	continue
// 		// }
// 		// mhbs := utils.Multihash(bs)

// 		mhbs := AddressNet(networkNum.Bytes())
// 		networkIDs = append(networkIDs, &mhbs)
// 	}
// 	networkIDsLock.Unlock()
// 	return networkIDs
// }

/*
	获得一个节点更远的节点中，比自己更远的节点
*/
func GetIdsForFar(id *AddressNet) []AddressNet {
	//计算来源的逻辑节点地址
	kl := NewKademlia(len(nodes) + 2)
	kl.Add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
	kl.Add(new(big.Int).SetBytes(*id))

	nodesLock.RLock()
	for _, one := range nodes {
		if one.Type != Node_type_logic && one.Type != Node_type_white_list {
			continue
		}
		kl.Add(new(big.Int).SetBytes(one.IdInfo.Id))
	}
	nodesLock.RUnlock()

	// Nodes.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })

	list := kl.Get(new(big.Int).SetBytes(*id))

	out := make([]AddressNet, 0)
	find := false
	for _, one := range list {

		// if hex.EncodeToString(one.Bytes()) == hex.EncodeToString(NodeSelf.IdInfo.Id.Data()) {
		if bytes.Equal(one.Bytes(), NodeSelf.IdInfo.Id) {
			find = true
		} else {
			if find {
				// bs, err := utils.Encode(one.Bytes(), config.HashCode)
				// if err != nil {
				// 	// fmt.Println("编码失败")
				// 	continue
				// }
				// mh := utils.Multihash(bs)
				mh := AddressNet(one.Bytes())
				out = append(out, mh)
			}
		}

	}

	return out
}

/*
	添加一个地址到白名单
*/
func AddWhiteList(addr AddressNet) {
	WhiteList.Store(utils.Bytes2string(addr), 0)
}

func DelWhiteList(addr AddressNet) {
	// WhiteList.Delete()
	WhiteList.Delete(utils.Bytes2string(addr))
}

func FindWhiteList(addr *AddressNet) bool {
	_, ok := WhiteList.Load(utils.Bytes2string(*addr))
	return ok
}

func init() {
	p1 := "2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk"
	AddWhiteList(AddressFromB58String(p1))
	p5 := "DNDywcPsJqsWq2gn7gH4yZg5GrAZbR5JvbpxoJDhyoAs"
	AddWhiteList(AddressFromB58String(p5))
	p6 := "5EontzaTP7Ad8ZQS9GviPQfZMVNMEFh5RMDhYUgZuSqB"
	AddWhiteList(AddressFromB58String(p6))
	p7 := "XXmf5vZZ7Nf7XbZsf1YN9hW2KdoKsQqyPMvpPGaxLo6"
	AddWhiteList(AddressFromB58String(p7))
	p8 := "GxCcAvSNRzymqyrRtcXFt9Mz829gFwnZ5TRNBA5Nz2Co"
	AddWhiteList(AddressFromB58String(p8))

	testp1 := "1E1ZBndCZsVDt1QXkk2gy4CTeh2sEQeVKvv7L3mP14S"
	AddWhiteList(AddressFromB58String(testp1))
	testp2 := "j9vydTsmTyC7hx6LmNUXYDtge2R4vhaggotVUF7oCPX"
	AddWhiteList(AddressFromB58String(testp2))
	testp3 := "6RKLimMBb9h2SdXuXooQTgycNdYEVcakxCsMQoWQpy7c"
	AddWhiteList(AddressFromB58String(testp3))
}
