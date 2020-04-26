package nodeStore

import (
	"math/big"
	"sort"
)

// //从大到小排序
// type IdDESC []*big.Int

// func (this IdDESC) Len() int {
// 	return len(this)
// }

// func (this IdDESC) Less(i, j int) bool {
// 	qu := new(big.Int).Sub(this[i], this[j])

// 	quInt := qu.Cmp(big.NewInt(0))
// 	//从大到小排序
// 	return quInt > 0

// 	// return this[i].NodeId < this[j].Val // 按值排序
// 	//return ms[i].Key < ms[j].Key // 按键排序
// }

// func (this IdDESC) Swap(i, j int) {
// 	this[i], this[j] = this[j], this[i]
// }

//从小到大排序
type IdASC struct {
	findNode *big.Int
	nodes    []*big.Int
}

func (this IdASC) Len() int {
	return len(this.nodes)
}

func (this IdASC) Less(i, j int) bool {
	a := new(big.Int).Xor(this.findNode, this.nodes[i])
	b := new(big.Int).Xor(this.findNode, this.nodes[j])
	if a.Cmp(b) > 0 {
		return false
	} else {
		return true
	}
}

func (this IdASC) Swap(i, j int) {
	this.nodes[i], this.nodes[j] = this.nodes[j], this.nodes[i]
}

func (this IdASC) Sort() []*big.Int {
	sort.Sort(this)
	return this.nodes
}

func NewIdASC(findNode *big.Int, nodes []*big.Int) *IdASC {
	return &IdASC{
		findNode: findNode,
		nodes:    nodes,
	}
}
