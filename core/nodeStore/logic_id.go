package nodeStore

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"sync"
)

type Ids struct {
	root           []byte
	ids            [][]byte
	count          int64
	logicNumBuider *LogicNumBuider
	lock           *sync.RWMutex
}

/*
	添加一个id
*/
func (this *Ids) AddId(id []byte) (ok bool, removeIDs [][]byte) {

	if this.count <= 0 {
		for i := 0; i < len(this.ids); i++ {
			this.ids[i] = id
		}
		this.count++
		ok = true
		return
	}

	//非逻辑节点不要添加
	netIDs := this.logicNumBuider.GetNodeNetworkNum()

	delId := make([][]byte, 0)
	for i, one := range this.ids {
		kl := NewKademlia()
		kl.Add(new(big.Int).SetBytes(one))
		kl.Add(new(big.Int).SetBytes(id))
		nearId := kl.Get(new(big.Int).SetBytes(*netIDs[i]))
		//		fmt.Println(hex.EncodeToString(nearId[0].Bytes()))

		// if hex.EncodeToString(*one) == hex.EncodeToString(nearId[0].Bytes()) {
		if bytes.Equal(one, nearId[0].Bytes()) {
			continue
		}
		//		fmt.Println("删除的节点id", i, one, "替换", node.IdInfo.Id)
		delId = append(delId, one)
		//		netNodes[i] = node.IdInfo.Id
		this.ids[i] = id
		ok = true
	}
	//找到删除的节点
	removeIDs = make([][]byte, 0)
	for _, one := range delId {
		find := false
		for _, netOne := range this.ids {
			if bytes.Equal(one, netOne) {
				find = true
				break
			}
		}
		if !find {
			removeIDs = append(removeIDs, one)
		}
	}
	if ok {
		this.count++
	}
	return
}

/*
	删除一个id
*/
func (this *Ids) RemoveId(id []byte) {
	have := false

	netIDs := this.logicNumBuider.GetNodeNetworkNum()
	for i, one := range this.ids {

		if bytes.Equal(one, id) {
			continue
		}
		ids := this.GetIds()

		kl := NewKademlia()

		for _, one := range ids {
			kl.Add(new(big.Int).SetBytes(one))
		}

		nearId := kl.Get(new(big.Int).SetBytes(*netIDs[i]))
		// mhbs, _ := utils.Encode(nearId[1].Bytes(), gconfig.HashCode)
		// idmh := utils.Multihash(mhbs)
		idAddr := []byte(nearId[1].Bytes())
		this.ids[i] = idAddr

		have = true
	}
	if have {
		this.count--
	}
}

/*
	获取所有id
*/
func (this *Ids) GetIds() [][]byte {
	m := make(map[string][]byte)
	if this.count <= 0 {
		return make([][]byte, 0)
	}
	//去重复
	for _, one := range this.ids {
		m[hex.EncodeToString(one)] = one
	}
	//组装成数组
	ids := make([][]byte, 0)
	for _, v := range m {
		ids = append(ids, v)
	}
	return ids
}

/*
	通过下标获取id
*/
func (this *Ids) GetIndex(index int) []byte {
	return this.ids[index]
}

func NewIds(id []byte, level uint) *Ids {
	lb := NewLogicNumBuider(id, level)
	return &Ids{
		root:           id,
		ids:            make([][]byte, level),
		logicNumBuider: lb,
		lock:           new(sync.RWMutex),
	}
}
