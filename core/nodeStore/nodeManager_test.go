package nodeStore

import (
	"fmt"
	// "math/big"
	// "math/rand"
	"testing"
	// "time"
)

func TestNodeManager(t *testing.T) {
	// nodeManagerTest()
	// imitationSum()
	checkNeedNodeTest()
}

func checkNeedNodeTest() {

	bs, _ := hex.DecodeString("4c4a60967f7cae228b2b1355b01c98976c1517131c25fe30ef77c0051976113e")
	nodeStore.NodeSelf = &nodeStore.Node{IdInfo: nodeStore.IdInfo{Id: nodeStore.NewIdAddress(bs)}}

	bs, _ = hex.DecodeString("6ea6ed79c7f34e560d127301d2451fabaf942d75b8de5e0bffc2124baf5eb427")
	nodeStore.AddNode(&nodeStore.Node{IdInfo: nodeStore.IdInfo{Id: nodeStore.NewIdAddress(bs)}})

	bs, _ = hex.DecodeString("1fcf6a5144aae28a09042cd10f56c69532b47268f3edea15665a7977e19fd84a")
	nodeStore.AddNode(&nodeStore.Node{IdInfo: nodeStore.IdInfo{Id: nodeStore.NewIdAddress(bs)}})

	bs, _ = hex.DecodeString("9d1406a76433f43ab752f468e0ca58baf73c9adddfc505a617c5756ea310e44f")
	nodeStore.AddNode(&nodeStore.Node{IdInfo: nodeStore.IdInfo{Id: nodeStore.NewIdAddress(bs)}})

	bs, _ = hex.DecodeString("edc392f218058cfe0da907a53a120950541cb7d31fd5e0d908326f7f9562fea6")
	ok, repl := nodeStore.CheckNeedNode(bs)
	fmt.Println("---------", ok, repl)

}

// func imitationSum2() {
// 	findNodeRoot, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
// 	fmt.Println("本节点id为：", findNodeRoot.String())
// 	findNode := NewNodeManager(&Node{NodeId: findNodeRoot}, 256)

// 	rootId, _ := new(big.Int).SetString("59422813065590763321187925186011450884940934337897117431794152839561407098597", 10)
// 	fmt.Println("本节点id为：", rootId.String())
// 	nodeManager := NewNodeManager(&Node{NodeId: rootId}, 256)

// 	// idA, _ := new(big.Int).SetString("38985264161753223911670476475859110596857691769085279908018319674400729625595", 10)
// 	// nodeManager.AddNode(&Node{NodeId: idA})
// 	// idB, _ := new(big.Int).SetString("59422813065590763321187925186011450884940934337897117431794152839561407098597", 10)
// 	// nodeManager.AddNode(&Node{NodeId: idB})
// 	idC, _ := new(big.Int).SetString("31622036050853307757176718873676335712993063093791913422933189278586653352673", 10)
// 	nodeManager.AddNode(&Node{NodeId: idC})
// 	// idD, _ := new(big.Int).SetString("38879061860890225964363770808076149471375052911854164467748691902681942298885", 10)
// 	// nodeManager.AddNode(&Node{NodeId: idD})

// 	for key, _ := range findNode.getNodeNetworkNum() {
// 		targetNode := nodeManager.Get(key, true, findNodeRoot.String())
// 		if targetNode.NodeId.String() == rootId.String() {
// 			fmt.Println("包含这个节点")
// 		}
// 		fmt.Println(targetNode.NodeId.String())
// 	}

// }

// func imitationSum() {
// 	findNodeRoot, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
// 	fmt.Println("本节点id为：", findNodeRoot.String())
// 	findNode := NewNodeManager(&Node{NodeId: findNodeRoot}, 256)

// 	rootId, _ := new(big.Int).SetString("38879061860890225964363770808076149471375052911854164467748691902681942298885", 10)
// 	fmt.Println("本节点id为：", rootId.String())
// 	nodeManager := NewNodeManager(&Node{NodeId: rootId}, 256)

// 	idA, _ := new(big.Int).SetString("38985264161753223911670476475859110596857691769085279908018319674400729625595", 10)
// 	nodeManager.AddNode(&Node{NodeId: idA})
// 	idB, _ := new(big.Int).SetString("59422813065590763321187925186011450884940934337897117431794152839561407098597", 10)
// 	nodeManager.AddNode(&Node{NodeId: idB})
// 	idC, _ := new(big.Int).SetString("31622036050853307757176718873676335712993063093791913422933189278586653352673", 10)
// 	nodeManager.AddNode(&Node{NodeId: idC})
// 	// idD, _ := new(big.Int).SetString("38879061860890225964363770808076149471375052911854164467748691902681942298885", 10)
// 	// nodeManager.AddNode(&Node{NodeId: idD})

// 	for key, _ := range findNode.getNodeNetworkNum() {
// 		targetNode := nodeManager.Get(key, true, findNodeRoot.String())
// 		if targetNode.NodeId.String() == rootId.String() {
// 			fmt.Println("包含这个节点")
// 		}
// 		// fmt.Println(targetNode.NodeId.String())
// 	}

// }

// func nodeManagerTest() {
// 	rootId := RandNodeId(4)
// 	fmt.Println("本节点id为：", rootId.String())

// 	node := &Node{
// 		NodeId:  rootId,
// 		IsSuper: true, //是超级节点
// 		Addr:    "1111",
// 		TcpPort: 8080,
// 		UdpPort: 0,
// 	}

// 	nodeManager := NewNodeManager(node, 4)

// 	idA, _ := new(big.Int).SetString("1", 10)
// 	idB, _ := new(big.Int).SetString("4", 10)
// 	idC, _ := new(big.Int).SetString("10", 10)
// 	idD, _ := new(big.Int).SetString("13", 10)

// 	nodeManager.AddNode(&Node{NodeId: idA})
// 	nodeManager.AddNode(&Node{NodeId: idB})
// 	nodeManager.AddNode(&Node{NodeId: idC})
// 	nodeManager.AddNode(&Node{NodeId: idD})
// 	ok, repl := nodeManager.CheckNeedNode("9")
// 	fmt.Println(ok, repl)

// }

// //得到指定长度的节点id
// //@return 10进制字符串
// func RandNodeId(lenght int) *big.Int {
// 	min := rand.New(rand.NewSource(99))
// 	timens := int64(time.Now().Nanosecond())
// 	min.Seed(timens)
// 	maxId := new(big.Int).Lsh(big.NewInt(1), uint(lenght))
// 	randInt := new(big.Int).Rand(min, maxId)
// 	return randInt
// }
