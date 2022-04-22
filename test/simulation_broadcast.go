package main

import (
	"mandela/config"
	gconfig "mandela/config"
	"mandela/core/nodeStore"
	"bytes"
	"fmt"
	"math/big"
	"strconv"
)

func main() {
	// BuildIds()
	Run2()
	fmt.Println("=============================")

	// ns := BuildIdsTree()
	// Select(ns)

	//==========================
	// run4()
}

/*
	测试查询逻辑节点
*/
func BuildIds() []*NodeOne {

	// fmt.Println("---------- BuildIds ----------")

	// ids := []string{
	// 	"2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk",
	// 	"8eWgP57sAKepA7h4FJV3ig4KekgNqE3exvhLLWtgDP57",
	// 	"Bsyuy8Cpg5VWi69axQKaU6pLbHkWHffCDjcQEFJC1qEr",
	// 	"4V2haHmFRdS5hp9VELnnVNkKwHmoUD41TSgUys28pskj",
	// 	"7JBa2oUeYgYSp9FUsHy6wt5WuqigJkDcsYoAntig7eTt",
	// 	"FPFEKi4MmDi9PssqkYbYjwTbzFnwqmanFZ7fwo7DNC1x",
	// 	// "BsVoRNtDaPZh2w6H48Y1fS2QRgjo4z6jmbJGVoTRqxMt",
	// 	"84yEekKXynEx3SSaQjEQUr5JDf6B1Fp34Kn2hBNQmNZS",
	// 	"DNDywcPsJqsWq2gn7gH4yZg5GrAZbR5JvbpxoJDhyoAs",
	// }

	ids := []string{
		"8cWMZz7FvkcRr3EA1VbzGZWdMKR7Uy4VYB2bEyzqdb29",
		// "ExTx6F4KD7buH3JZEuNNKd3Kzrf25P9ZSu2AeoupCR9w",
		"8NjRXxa6KKSv9NtP2KnwzRAQGuMnqARKznCUa1U9LmR1",
		"x8sXmm9HKqyit7Pad5U5Y4o8WFPx9s3aNxQCmJfZWNk",
		"D3BbKJJyaQhWiT6eKiSryX22ZdX8AfduE1sb4Bp4PQYQ",
		"BqfZe98Sb8iiajFKHf9ndSuyapd5jXrkUSoDbCLFsfuW",
		"AhSUaV5hJ6U5Pgf7fwimpZPK7cmMwmpfmcv1fXxKJHwx",
		// "FQyov18orsY553JHYhHXgs5u6uPxh7bM7484aPKenpnJ",
	}

	nodes := make([]*NodeOne, 0)

	for n := 0; n < len(ids); n++ {
		nodeOne := &NodeOne{
			Nodes: make([]nodeStore.AddressNet, 0),
		}
		fmt.Println(n+1, "本节点为", ids[n])
		index := n

		idMH := nodeStore.AddressFromB58String(ids[index])
		idsm := nodeStore.NewIds(idMH, gconfig.NodeIDLevel)
		nodeOne.Self = idMH
		for i, one := range ids {
			if i == index {
				continue
			}

			idMH := nodeStore.AddressFromB58String(one)
			idsm.AddId(idMH)
			//		ok, remove := idsm.AddId(&idMH)
			//		if ok {
			//			fmt.Println(one, remove)
			//		}
		}

		is := idsm.GetIds()
		for _, one := range is {

			idOne := nodeStore.AddressNet(one)
			nodeOne.Nodes = append(nodeOne.Nodes, idOne)

			fmt.Println("    --逻辑节点", idOne.B58String())
		}
		fmt.Println()
		nodes = append(nodes, nodeOne)
	}

	return nodes

}

//============================================
//模拟节点广播，消息到达情况
func Run2() {
	nodes := BuildIdsTwo()
	PrintNodes(nodes)
	// fmt.Println("--------------------")
	broadcast(nodes)
}

func BuildIdsTwo() []*NodeOne {

	ids := []string{
		"2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk",
		"8eWgP57sAKepA7h4FJV3ig4KekgNqE3exvhLLWtgDP57",
		"Bsyuy8Cpg5VWi69axQKaU6pLbHkWHffCDjcQEFJC1qEr",
		"4V2haHmFRdS5hp9VELnnVNkKwHmoUD41TSgUys28pskj",
		"5sok55G8osoiUJDbPLFpJbZUoowtxeWMHVyK9CvCS45g",
	}

	nodes := make([]*NodeOne, 0)

	for n := 0; n < len(ids); n++ {
		nodeOne := &NodeOne{
			Nodes: make([]nodeStore.AddressNet, 0),
		}
		idMH := nodeStore.AddressFromB58String(ids[n])
		// idsm := nodeStore.NewIds(idMH, gconfig.NodeIDLevel)
		nodeOne.Self = idMH
		nodes = append(nodes, nodeOne)
	}

	node0 := nodes[0]
	node0.Nodes = append(node0.Nodes, nodes[2].Self)
	node0.Nodes = append(node0.Nodes, nodes[3].Self)

	node1 := nodes[1]
	node1.Nodes = append(node1.Nodes, nodes[4].Self)
	// node1.Nodes = append(node1.Nodes, nodes[3].Self)

	node2 := nodes[2]
	node2.Nodes = append(node2.Nodes, nodes[0].Self)
	node2.Nodes = append(node2.Nodes, nodes[3].Self)

	node3 := nodes[3]
	node3.Nodes = append(node3.Nodes, nodes[0].Self)
	node3.Nodes = append(node3.Nodes, nodes[2].Self)

	// node4 := nodes[4]
	// node4.Nodes = append(node4.Nodes, nodes[2].Self)
	// node4.Nodes = append(node4.Nodes, nodes[3].Self)

	return nodes
}

func PrintNodes(nodes []*NodeOne) {
	for _, one := range nodes {
		fmt.Println(one.Self.B58String(), one.Msg)
		for _, two := range one.Nodes {
			fmt.Println("  ", two.B58String())
		}
	}
}

func broadcast(nodes []*NodeOne) {
	for i, one := range nodes {
		//初始化状态
		for _, temp := range nodes {
			temp.Msg = false
		}
		fmt.Println("第" + strconv.Itoa(i+1) + "轮消息")
		one.Msg = true

		//开始广播
		for j, _ := range one.Nodes {
			tree := one.Nodes[j]
			for x, _ := range nodes {
				temp := nodes[x]
				if bytes.Equal(tree, temp.Self) {
					fmt.Println("src", tree.B58String(), "->", temp.Self.B58String())
					temp.Msg = true
					loopSend(temp, tree, nodes)
					break
				}
			}
		}

		PrintNodes(nodes)
	}
}
func loopSend(self *NodeOne, src nodeStore.AddressNet, nodes []*NodeOne) {
	farNodes := self.GetIdsForFar(src)
	for j, _ := range farNodes {
		tree := farNodes[j]
		for x, _ := range nodes {
			temp := nodes[x]
			if temp.Msg {
				continue
			}
			if bytes.Equal(tree, temp.Self) {
				fmt.Println("src", tree.B58String(), "->", temp.Self.B58String())
				temp.Msg = true
				loopSend(temp, tree, nodes)
				break
			}
		}
	}

}

// nodeStore.GetIdsForFar()

type NodeOne struct {
	Self  nodeStore.AddressNet
	Nodes []nodeStore.AddressNet
	Msg   bool
}

func (this *NodeOne) GetIdsForFar(id nodeStore.AddressNet) []nodeStore.AddressNet {
	//计算来源的逻辑节点地址
	kl := nodeStore.NewKademlia()
	kl.Add(new(big.Int).SetBytes(this.Self))
	kl.Add(new(big.Int).SetBytes(id))

	for i, _ := range this.Nodes {
		kl.Add(new(big.Int).SetBytes(this.Nodes[i]))
	}
	// Nodes.Range(func(k, v interface{}) bool {
	// 	value := v.(*Node)
	// 	kl.Add(new(big.Int).SetBytes(value.IdInfo.Id))
	// 	return true
	// })

	list := kl.Get(new(big.Int).SetBytes(id))

	out := make([]nodeStore.AddressNet, 0)
	find := false
	for _, one := range list {

		// if hex.EncodeToString(one.Bytes()) == hex.EncodeToString(NodeSelf.IdInfo.Id.Data()) {
		if bytes.Equal(one.Bytes(), this.Self) {
			find = true
		} else {
			if find {
				// bs, err := utils.Encode(one.Bytes(), config.HashCode)
				// if err != nil {
				// 	// fmt.Println("编码失败")
				// 	continue
				// }
				// mh := utils.Multihash(bs)
				mh := nodeStore.AddressNet(one.Bytes())
				out = append(out, mh)
			}
		}

	}

	return out
}

//--------------------------

func BuildIdsTree() []*Node {
	ids := []string{
		"2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk",
		"8eWgP57sAKepA7h4FJV3ig4KekgNqE3exvhLLWtgDP57",
		"Bsyuy8Cpg5VWi69axQKaU6pLbHkWHffCDjcQEFJC1qEr",
		"4V2haHmFRdS5hp9VELnnVNkKwHmoUD41TSgUys28pskj",
		"5sok55G8osoiUJDbPLFpJbZUoowtxeWMHVyK9CvCS45g",
	}

	nodes := make([]*Node, 0)

	for n := 0; n < len(ids); n++ {
		nodeOne := &Node{
			Nodes: make([]*Node, 0),
		}
		idMH := nodeStore.AddressFromB58String(ids[n])
		nodeOne.Self = idMH

		idBuilder := nodeStore.NewLogicNumBuider(nodeOne.Self, config.NodeIDLevel)
		nodeOne.LogicIds = idBuilder.GetNodeNetworkNum()
		nodes = append(nodes, nodeOne)
	}

	node0 := nodes[0]
	node0.Nodes = append(node0.Nodes, nodes[2])
	node0.Nodes = append(node0.Nodes, nodes[3])

	node1 := nodes[1]
	node1.Nodes = append(node1.Nodes, nodes[4])
	// node1.Nodes = append(node1.Nodes, nodes[3].Self)

	node2 := nodes[2]
	node2.Nodes = append(node2.Nodes, nodes[0])
	node2.Nodes = append(node2.Nodes, nodes[3])

	node3 := nodes[3]
	node3.Nodes = append(node3.Nodes, nodes[0])
	node3.Nodes = append(node3.Nodes, nodes[2])

	// node4 := nodes[4]
	// node4.Nodes = append(node4.Nodes, nodes[2].Self)
	// node4.Nodes = append(node4.Nodes, nodes[3].Self)

	return nodes
}

func Select(nodes []*Node) {
	for n := 0; n < 1; n++ {
		fmt.Println("第" + strconv.Itoa(n) + "伦查询")
		//开始查询
		for _, nodeOne := range nodes {

			for _, idOne := range nodeOne.LogicIds {
				// for _,logicIdOne := range idOne.
				nodeOne.Send(nodeOne.Self, *idOne)
				// return
			}
			// return
			// one.FindNearNodeId()
		}
		//
		// Print(nodes)
	}
}
func Print(nodes []*Node) {
	for _, one := range nodes {
		fmt.Println(one.Self.B58String())
		for _, two := range one.Nodes {
			fmt.Println("    ", two.Self.B58String())
		}
	}
}

type Node struct {
	Self        nodeStore.AddressNet
	Nodes       []*Node
	NodesClient []*Node
	LogicIds    []*[]byte
	Msg         bool
}

func (this *Node) Send(src, findId nodeStore.AddressNet) {
	id := this.FindNearNodeId(findId, nil, true)
	// fmt.Println(this.Self.B58String(), "Send", id.B58String())
	if bytes.Equal(id, this.Self) {
		// fmt.Println
		this.ReturnMsg(src, this)
		return
	}
	for _, one := range this.Nodes {
		if bytes.Equal(one.Self, id) {
			one.Send(src, findId)
			return
		}
	}
}

func (this *Node) ReturnMsg(src nodeStore.AddressNet, findNode *Node) {
	id := this.FindNearNodeId(src, nil, true)
	if bytes.Equal(id, this.Self) {
		// fmt.Println
		// this.ReturnMsg(src, this)
		//不添加自己
		if bytes.Equal(this.Self, findNode.Self) {
			return
		}

		//是否已经添加过了
		have := false
		for _, one := range this.Nodes {
			if bytes.Equal(one.Self, findNode.Self) {
				have = true
				break
			}
		}
		if have {
			return
		}
		fmt.Println(this.Self.B58String(), "添加", findNode.Self.B58String())
		return
	}
	for _, one := range this.Nodes {
		if bytes.Equal(one.Self, id) {
			one.ReturnMsg(src, findNode)
			return
		}
	}
}

//在节点中找到最近的节点，包括代理节点
func (this *Node) FindNearNodeId(nodeId, outId nodeStore.AddressNet, includeSelf bool) nodeStore.AddressNet {
	kl := nodeStore.NewKademlia()
	if includeSelf {
		kl.Add(new(big.Int).SetBytes(this.Self))
	}

	for _, one := range this.Nodes {
		if bytes.Equal(one.Self, outId) {
			continue
		}
		kl.Add(new(big.Int).SetBytes(one.Self))
	}

	targetIds := kl.Get(new(big.Int).SetBytes(nodeId))
	if len(targetIds) == 0 {
		return nil
	}
	targetId := targetIds[0]
	if targetId == nil {
		return nil
	}
	mh := nodeStore.AddressNet(targetId.Bytes())
	return mh
}

//----------------------------------------

func run4() {
	nodes := BuildIds4()
	Select4(nodes)
}

func BuildIds4() []*Node {
	ids := []string{
		// "2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk",
		// "8eWgP57sAKepA7h4FJV3ig4KekgNqE3exvhLLWtgDP57",
		// "Bsyuy8Cpg5VWi69axQKaU6pLbHkWHffCDjcQEFJC1qEr",
		// "4V2haHmFRdS5hp9VELnnVNkKwHmoUD41TSgUys28pskj",
		// "5sok55G8osoiUJDbPLFpJbZUoowtxeWMHVyK9CvCS45g",
		// "Cg3iTN8bopEueFnZ6UULZnJDQrBBxfT239WTDYKgoVtk",

		"8eWgP57sAKepA7h4FJV3ig4KekgNqE3exvhLLWtgDP57",
		"2w5QBfujmLTAvesJRyRpxZFj4D4PJTEbhDVQJt1kbDmk",
		"Bsyuy8Cpg5VWi69axQKaU6pLbHkWHffCDjcQEFJC1qEr",
		"4V2haHmFRdS5hp9VELnnVNkKwHmoUD41TSgUys28pskj",
		"7JBa2oUeYgYSp9FUsHy6wt5WuqigJkDcsYoAntig7eTt",
		"FPFEKi4MmDi9PssqkYbYjwTbzFnwqmanFZ7fwo7DNC1x",
		"84yEekKXynEx3SSaQjEQUr5JDf6B1Fp34Kn2hBNQmNZS",
		"DNDywcPsJqsWq2gn7gH4yZg5GrAZbR5JvbpxoJDhyoAs",
	}

	nodes := make([]*Node, 0)

	for n := 0; n < len(ids); n++ {
		nodeOne := &Node{
			Nodes: make([]*Node, 0),
		}
		idMH := nodeStore.AddressFromB58String(ids[n])
		nodeOne.Self = idMH

		idBuilder := nodeStore.NewLogicNumBuider(nodeOne.Self, config.NodeIDLevel)
		nodeOne.LogicIds = idBuilder.GetNodeNetworkNum()
		nodes = append(nodes, nodeOne)
	}

	// node0 := nodes[0]
	// node0.Nodes = append(node0.Nodes, nodes[2])
	// node0.Nodes = append(node0.Nodes, nodes[3])

	for i, _ := range nodes {
		if i == 0 {
			continue
		}
		one := nodes[i]
		one.Nodes = append(one.Nodes, nodes[0])
	}

	return nodes
}

func Select4(nodes []*Node) {
	for n := 0; n < 10; n++ {
		fmt.Println("第" + strconv.Itoa(n) + "伦查询")
		//开始查询
		for _, nodeOne := range nodes {

			for _, logicNodeOne := range nodeOne.Nodes {
				idsm := nodeStore.NewIds(nodeOne.Self, nodeStore.NodeIdLevel)
				for _, one := range append(logicNodeOne.Nodes, logicNodeOne.NodesClient...) {
					if bytes.Equal(nodeOne.Self, one.Self) {
						continue
					}
					idsm.AddId(one.Self)
				}
				ids := idsm.GetIds()
				for _, one := range ids {
					//找到这个node并添加
					for _, findNodeOne := range nodes {
						if bytes.Equal(findNodeOne.Self, one) {
							nodeOne.AddNode(findNodeOne)
							break
						}
					}
				}
			}

		}
		//
		Print(nodes)
	}
}

func (this *Node) AddNode(node *Node) {
	//	initIds()

	//不添加自己
	if bytes.Equal(this.Self, node.Self) {
		return
	}

	idm := nodeStore.NewIds(this.Self, nodeStore.NodeIdLevel)
	// ids := GetLogicNodes()
	for _, one := range this.Nodes {
		idm.AddId(one.Self)
	}

	ok, removeIDs := idm.AddId(node.Self)
	if ok {
		this.Nodes = append(this.Nodes, node)
		node.NodesClient = append(node.NodesClient, this)

		for _, one := range removeIDs {
			for i, nodeOne := range this.Nodes {
				if bytes.Equal(nodeOne.Self, one) {
					temp := this.Nodes[:i]
					temp = append(temp, this.Nodes[i+1:]...)
					this.Nodes = temp
					//从对方的ClientNodes中删除
					for j, clientOne := range nodeOne.NodesClient {
						if bytes.Equal(clientOne.Self, nodeOne.Self) {
							temp := nodeOne.NodesClient[:j]
							temp = append(temp, nodeOne.NodesClient[j+1:]...)
							nodeOne.NodesClient = temp
							break
						}
					}
					break
				}
			}
		}
	}
	//	fmt.Println("添加一个node", node.IdInfo.Id.B58String())
	return

}
