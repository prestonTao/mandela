package nodeStore

// import (
// 	"math/big"
// 	"sync"
// )

// type NodeGroup struct {
// 	lock   *sync.RWMutex
// 	groups map[string][]big.Int //保存所有组
// }

// //添加到一个组
// func (this *NodeGroup) AddNode(groupName string, nodeId big.Int) {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()
// 	group, ok := this.groups[groupName]
// 	if ok {
// 		group = append(group, nodeId)
// 		this.groups[groupName] = group
// 		return
// 	}
// 	group = make([]big.Int, 0)
// 	this.groups[groupName] = group
// }

// //
// func (this *NodeGroup) DelGroup(groupName string) {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()
// 	delete(this.groups, groupName)
// }

// func NewNodeGroup() *NodeGroup {
// 	nodeGroup := new(NodeGroup)
// 	nodeGroup.lock = new(sync.RWMutex)
// 	nodeGroup.groups = make(map[string][]big.Int)
// 	return nodeGroup
// }
