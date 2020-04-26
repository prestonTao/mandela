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

const (
	Node_min = 50 //每个节点的最少连接数量
)

var (
	NodeSelf         *Node            = new(Node)                    //保存自己的id信息和ip地址和端口号
	Nodes                             = new(sync.Map)                //超级节点。key:string=节点id,value:*Node=节点信息;
	Proxys                            = new(sync.Map)                //被代理的节点。key:string=节点id;value:*Node=节点信息;
	OtherNodes                        = new(sync.Map)                //每个节点有最少连接数量。key:string=节点id;value:*AddressNet=节点地址;
	OutFindNode      chan *AddressNet = make(chan *AddressNet, 1000) //需要查询的逻辑节点
	OutCloseConnName                  = make(chan *AddressNet, 1000) //废弃的nodeid，需要询问是否关闭
	SuperPeerId      *AddressNet                                     //超级节点名称
	NodeIdLevel      uint             = 256                          //节点id长度
	// LanNode                           = new(sync.Map)                //局域网节点。key:string=节点id;value:*Node=节点信息;
	// Groups           *NodeGroup       = NewNodeGroup()               //组
	// once             sync.Once
	// Key              *ECCKey
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

	addr := keystore.GetCoinbase()
	_, prk, puk, err := keystore.GetKeyByAddr(addr, pwd)
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
func AddProxyNode(node *Node) {
	Proxys.Store(node.IdInfo.Id.B58String(), node)
}

//得到一个代理节点
func GetProxyNode(id string) (node *Node, ok bool) {
	var v interface{}
	v, ok = Proxys.Load(id)
	if v == nil {
		return nil, ok
	}
	node = v.(*Node)
	return
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
func AddNode(node *Node) {
	//	initIds()

	//不添加自己
	if bytes.Equal(node.IdInfo.Id, NodeSelf.IdInfo.Id) {
		return
	}

	//检查是否够最少连接数量
	total := 0
	OtherNodes.Range(func(k, v interface{}) bool {
		total++
		return true
	})
	//不够就保存
	if total < Node_min {
		//保存
		OtherNodes.Store(node.IdInfo.Id.B58String(), node)
	}

	idm := NewIds(NodeSelf.IdInfo.Id, NodeIdLevel)
	ids := GetLogicNodes()
	for _, one := range ids {
		idm.AddId(*one)
	}

	ok, removeIDs := idm.AddId(node.IdInfo.Id)
	if ok {
		//		fmt.Println("添加成功", new(big.Int).SetBytes(node.IdInfo.Id.Data()).Int64())
		node.lastContactTimestamp = time.Now()
		Nodes.Store(node.IdInfo.Id.B58String(), node)
		//修改超级节点，普通节点经常切换影响网络
		addrNet := AddressNet(idm.GetIndex(0))
		SuperPeerId = &addrNet

		//删除被替换的id
		for _, one := range removeIDs {
			//			idOne := hex.EncodeToString(one)
			//			delete(Nodes, idOne)
			//			OutCloseConnName <- idOne
			addrNet := AddressNet(one)
			Nodes.Delete(addrNet.B58String())
			OutCloseConnName <- &addrNet
		}

	}
	//	fmt.Println("添加一个node", node.IdInfo.Id.B58String())
	return

}

/*
	删除一个节点，包括超级节点和代理节点
*/
func DelNode(id *AddressNet) {
	idStr := id.B58String()
	Nodes.Delete(idStr)
	Proxys.Delete(idStr)
	OtherNodes.Delete(idStr)
	engine.RemoveSession(idStr)
}

/*
	通过id查找一个节点
*/
func FindNode(id *AddressNet) *Node {
	v, ok := Nodes.Load(id.B58String())
	if ok {
		return v.(*Node)
	}
	v, ok = OtherNodes.Load(id.B58String())
	if ok {
		return v.(*Node)
	}

	v, ok = Proxys.Load(id.B58String())
	if ok {
		return v.(*Node)
	}
	return nil
}

/*
	在超级节点中找到最近的节点，不包括代理节点
	@nodeId         要查找的节点
	@outId          排除一个节点
	@includeSelf    是否包括自己
	@return         查找到的节点id，可能为空
*/
func FindNearInSuper(nodeId, outId *AddressNet, includeSelf bool) *AddressNet {
	kl := NewKademlia()
	if includeSelf {
		kl.Add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
	}
	outIdStr := ""
	if outId != nil {
		outIdStr = outId.B58String()
	}
	Nodes.Range(func(k, v interface{}) bool {
		if k.(string) == outIdStr {
			return true
		}
		value := v.(*Node)
		kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
		return true
	})

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
	kl := NewKademlia()
	if includeSelf {
		kl.Add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
	}
	outIdStr := ""
	if outId != nil {
		outIdStr = outId.B58String()
	}
	Nodes.Range(func(k, v interface{}) bool {
		if k.(string) == outIdStr {
			return true
		}
		value := v.(*Node)
		kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
		return true
	})
	//代理节点
	Proxys.Range(func(k, v interface{}) bool {
		if k.(string) == outIdStr {
			return true
		}
		value := v.(*Node)
		//过滤APP节点
		if value.IsApp {
			return true
		}
		kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
		return true
	})

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

//得到所有逻辑节点，不包括本节点，也不包括代理节点
func GetLogicNodes() []*AddressNet {
	ids := make([]*AddressNet, 0)
	Nodes.Range(func(k, v interface{}) bool {
		value := v.(*Node)
		ids = append(ids, &value.IdInfo.Id)
		return true
	})
	return ids
}

/*
	获得所有代理节点
*/
func GetProxyAll() []*AddressNet {
	// ids := make([]string, 0)
	// Proxys.Range(func(key, value interface{}) bool {
	// 	ids = append(ids, key.(string))
	// 	return true
	// })
	// return ids

	ids := make([]*AddressNet, 0)
	Proxys.Range(func(k, v interface{}) bool {
		value := v.(*Node)
		ids = append(ids, &value.IdInfo.Id)
		return true
	})
	return ids
}

/*
	获取额外的节点连接
*/
func GetOtherNodes() []*AddressNet {
	ids := make([]*AddressNet, 0)
	OtherNodes.Range(func(k, v interface{}) bool {
		value := v.(*Node)
		ids = append(ids, &value.IdInfo.Id)
		return true
	})
	return ids
}

/*
	获得本机所有逻辑节点的ip地址
*/
func GetSuperNodeIps() (ips []string) {
	ips = make([]string, 0)
	Nodes.Range(func(k, v interface{}) bool {
		value := v.(*Node)
		ips = append(ips, value.Addr+":"+strconv.Itoa(int(value.TcpPort)))
		return true
	})

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
	// if nodeId.B58String() == NodeSelf.IdInfo.Id.B58String() {
	if bytes.Equal(*nodeId, NodeSelf.IdInfo.Id) {
		//		fmt.Println("2不添加")
		return false
	}

	ids := NewIds(NodeSelf.IdInfo.Id, NodeIdLevel)
	for _, one := range GetLogicNodes() {
		ids.AddId(*one)
	}
	ok, _ := ids.AddId(*nodeId)
	return ok

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
func GetIdsForFar(id *AddressNet) []*AddressNet {
	//计算来源的逻辑节点地址
	kl := NewKademlia()
	kl.Add(new(big.Int).SetBytes(NodeSelf.IdInfo.Id))
	kl.Add(new(big.Int).SetBytes(*id))

	Nodes.Range(func(k, v interface{}) bool {
		value := v.(*Node)
		kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
		return true
	})

	list := kl.Get(new(big.Int).SetBytes(*id))

	out := make([]*AddressNet, 0)
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
				out = append(out, &mh)
			}
		}

	}

	return out
}
