package nodeStore

// import (
// 	"math/big"
// 	"sort"
// 	"sync"
// )

// type ConsistentHash struct {
// 	lock  sync.Mutex
// 	nodes IdDESC
// }

// //10进制字符串
// func (this *ConsistentHash) add(nodes ...*big.Int) {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()
// 	for _, idOne := range nodes {
// 		//判断重复的
// 		has := false
// 		for _, node := range this.nodes {
// 			if node.Cmp(idOne) == 0 {
// 				has = true
// 				break
// 			}
// 		}
// 		if !has {
// 			this.nodes = append(this.nodes, idOne)
// 		}
// 	}
// 	sort.Sort(this.nodes)
// }

// func (this *ConsistentHash) get(nodeId *big.Int) *big.Int {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()
// 	if len(this.nodes) == 0 {
// 		return nil
// 	}
// 	isFind := false
// 	for i, idOne := range this.nodes {
// 		switch nodeId.Cmp(idOne) {
// 		case 0:
// 			return idOne
// 		case -1:
// 			// fmt.Println("haha")
// 			isFind = true
// 		case 1:
// 			if i == 0 {
// 				firstDistanceInt := new(big.Int).Xor(nodeId, idOne)
// 				lastDistanceInt := new(big.Int).Xor(nodeId, this.nodes[len(this.nodes)-1])
// 				switch firstDistanceInt.Cmp(lastDistanceInt) {
// 				case 0:
// 					return idOne
// 				case -1:
// 					return idOne
// 				case 1:
// 					return this.nodes[len(this.nodes)-1]
// 				}
// 			}
// 			if isFind {
// 				startDistanceInt := new(big.Int).Xor(nodeId, this.nodes[i-1])
// 				lastDistanceInt := new(big.Int).Xor(nodeId, idOne)
// 				switch startDistanceInt.Cmp(lastDistanceInt) {
// 				case 0:
// 					return idOne
// 				case -1:
// 					return this.nodes[i-1]
// 				case 1:
// 					return idOne
// 				}
// 			}
// 		}
// 	}

// 	firstDistanceInt := new(big.Int).Xor(this.nodes[0], nodeId)
// 	lastDistanceInt := new(big.Int).Xor(nodeId, this.nodes[len(this.nodes)-1])
// 	switch firstDistanceInt.Cmp(lastDistanceInt) {
// 	case 0:
// 		return this.nodes[0]
// 	case -1:
// 		return this.nodes[0]
// 	case 1:
// 		return this.nodes[len(this.nodes)-1]
// 	}
// 	return nil
// }

// //获得左边最近的节点
// //@nodeId     要查询的节点id
// //@maxId      查询的id数量
// func (this *ConsistentHash) getLeftLow(nodeId *big.Int, count int) []*big.Int {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()
// 	if len(this.nodes) == 0 {
// 		return nil
// 	}
// 	if len(this.nodes) <= count {
// 		return this.nodes
// 	}
// 	maxId := new(big.Int).Lsh(big.NewInt(1), NodeIdLevel)
// 	var idsTemp IdASC = make([]*big.Int, 0)
// 	for _, idOne := range this.nodes {
// 		switch idOne.Cmp(nodeId) {
// 		case 0:
// 		case -1:
// 			idsTemp = append(idsTemp, new(big.Int).Add(idOne, maxId))
// 		case 1:
// 			idsTemp = append(idsTemp, idOne)
// 		}
// 	}
// 	sort.Sort(idsTemp)
// 	ids := make([]*big.Int, 0)
// 	if len(idsTemp) > count {
// 		idsTemp = idsTemp[:count]
// 	}
// 	for _, idOne := range idsTemp {
// 		switch idOne.Cmp(maxId) {
// 		case 0:
// 			ids = append(ids, big.NewInt(0))
// 		case -1:
// 			ids = append(ids, idOne)
// 		case 1:
// 			ids = append(ids, new(big.Int).Sub(idOne, maxId))
// 		}
// 	}
// 	// fmt.Println("left:", ids)
// 	return ids
// }

// //获得右边最近的节点，不包括被查询节点
// //@nodeId     要查询的节点id
// //@maxId      查询的id数量
// func (this *ConsistentHash) getRightLow(nodeId *big.Int, count int) []*big.Int {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()
// 	if len(this.nodes) == 0 {
// 		return nil
// 	}
// 	if len(this.nodes) <= count {
// 		return this.nodes
// 	}
// 	maxId := new(big.Int).Lsh(big.NewInt(1), NodeIdLevel)
// 	var idsTemp IdDESC = make([]*big.Int, 0)
// 	for _, idOne := range this.nodes {
// 		switch idOne.Cmp(nodeId) {
// 		case 0:
// 		case -1:
// 			idsTemp = append(idsTemp, idOne)
// 		case 1:
// 			idsTemp = append(idsTemp, new(big.Int).Sub(idOne, maxId))
// 		}
// 	}
// 	sort.Sort(idsTemp)
// 	ids := make([]*big.Int, 0)
// 	if len(idsTemp) > count {
// 		idsTemp = idsTemp[:count]
// 	}
// 	for _, idOne := range idsTemp {
// 		switch idOne.Cmp(big.NewInt(0)) {
// 		case 0:
// 			ids = append(ids, big.NewInt(0))
// 		case -1:
// 			ids = append(ids, new(big.Int).Add(idOne, maxId))
// 		case 1:
// 			ids = append(ids, idOne)
// 		}
// 	}
// 	// fmt.Println("right:", ids)
// 	return ids
// }

// //删除一个节点
// func (this *ConsistentHash) del(node *big.Int) {
// 	this.lock.Lock()
// 	defer this.lock.Unlock()
// 	//判断重复的
// 	for i, nodeOne := range this.nodes {
// 		if nodeOne.Cmp(node) == 0 {
// 			this.nodes = append(this.nodes[:i], this.nodes[i+1:]...)
// 			return
// 		}
// 	}
// }

// //得到hash表中保存的所有节点
// func (this *ConsistentHash) getAll() []*big.Int {
// 	return this.nodes
// }

// //创建一个新的一致性hash表
// func NewHash() *ConsistentHash {
// 	chash := &ConsistentHash{nodes: []*big.Int{}}
// 	return chash
// }
