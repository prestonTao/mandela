package nodeStore

//import (
//	"encoding/hex"
//	"math/big"
//	"sort"
//)

///*
//	保存邻居节点，保存比自己大的和最小的，并且距离最近的节点id
//*/
//type RecentNode struct {
//	root     *big.Int //自己节点
//	maxSize  int      //保存的数量
//	preNodes IdASC    //前面的节点 保存的节点,下标越小，离自己越近
//	sufNodes IdDESC   //后面的节点
//}

////检查一个节点是否需要
//func (this *RecentNode) CheckIn(nodeId *big.Int) (bool, string) {
//	if nodeId.Cmp(this.root) == 0 {
//		return false, ""
//	}
//	if len(append(this.preNodes, this.sufNodes...)) < this.maxSize*2 {
//		return true, ""
//	}
//	temp := NewRecentNode(this.root, this.maxSize)
//	for _, idOne := range this.GetAll() {
//		temp.Add(idOne)
//	}
//	switch temp.root.Cmp(nodeId) {
//	case 0: //和root节点相等
//	case -1: //在节点前面
//		temp.preNodes = append(temp.preNodes, nodeId)
//		sort.Sort(temp.preNodes)
//		passId := temp.preNodes[temp.maxSize]
//		if passId.Cmp(nodeId) == 0 {
//			return false, ""
//		}
//		return true, hex.EncodeToString(passId.Bytes())
//	case 1: //在节点后面
//		temp.sufNodes = append(temp.sufNodes, nodeId)
//		sort.Sort(temp.sufNodes)
//		passId := temp.sufNodes[temp.maxSize]
//		if passId.Cmp(nodeId) == 0 {
//			return false, ""
//		}
//		return true, hex.EncodeToString(passId.Bytes())
//	}
//	return false, ""
//}

//func (this *RecentNode) Add(nodeId *big.Int) {
//	switch this.root.Cmp(nodeId) {
//	case 0: //和root节点相等
//	case -1: //在节点前面
//		this.preNodes = append(this.preNodes, nodeId)
//		sort.Sort(this.preNodes)
//		this.preNodes = this.preNodes[:this.maxSize]
//	case 1: //在节点后面
//		this.sufNodes = append(this.sufNodes, nodeId)
//		sort.Sort(this.sufNodes)
//		this.sufNodes = this.sufNodes[:this.maxSize]
//	}

//}

//func (this *RecentNode) GetAll() []*big.Int {
//	return append(this.preNodes, this.sufNodes...)
//}

//func (this *RecentNode) Del(nodeId *big.Int) {
//	switch this.root.Cmp(nodeId) {
//	case 0: //和root节点相等
//	case -1: //在节点前面
//		for i, id := range this.preNodes {
//			if id.Cmp(nodeId) == 0 {
//				this.preNodes = append(this.preNodes[:i], this.preNodes[i+1:]...)
//			}
//		}
//	case 1: //在节点后面
//		for i, id := range this.sufNodes {
//			if id.Cmp(nodeId) == 0 {
//				this.sufNodes = append(this.sufNodes[:i], this.sufNodes[i+1:]...)
//			}
//		}
//	}
//}

//func NewRecentNode(nodeId *big.Int, maxSize int) *RecentNode {
//	recentNode := new(RecentNode)
//	recentNode.root = nodeId
//	recentNode.maxSize = maxSize
//	return recentNode
//}
