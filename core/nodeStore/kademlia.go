package nodeStore

import (
	"math/big"
	"sort"
)

type Kademlia struct {
	findNode *big.Int
	nodes    []*big.Int
}

func (this *Kademlia) Len() int {
	return len(this.nodes)
}

func (this *Kademlia) Less(i, j int) bool {
	a := new(big.Int).Xor(this.findNode, this.nodes[i])
	b := new(big.Int).Xor(this.findNode, this.nodes[j])
	if a.Cmp(b) > 0 {
		return false
	} else {
		return true
	}
}

func (this *Kademlia) Swap(i, j int) {
	this.nodes[i], this.nodes[j] = this.nodes[j], this.nodes[i]
}

/*
	添加节点
*/
func (this *Kademlia) Add(nodes ...*big.Int) {
	for _, idOne := range nodes {
		//判断重复的
		has := false
		for _, node := range this.nodes {
			if node.Cmp(idOne) == 0 {
				has = true
				break
			}
		}
		if !has {
			this.nodes = append(this.nodes, idOne)
		}
	}
}

/*
	获得这个节点由近到远距离排序的节点列表
*/
func (this *Kademlia) Get(nodeId *big.Int) []*big.Int {
	this.findNode = nodeId
	sort.Sort(this)
	return this.nodes
}

/*
	保留逻辑节点，清理掉不需要的节点id
*/
func (this *Kademlia) clear() {
}

func NewKademlia(length int) *Kademlia {
	k := &Kademlia{
		nodes: make([]*big.Int, 0, length),
	}
	return k
}
