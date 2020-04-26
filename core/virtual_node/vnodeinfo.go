package virtual_node

import (
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"bytes"
)

const (
	max_vnode_index = 100000000
)

type Vnodeinfo struct {
	Nid   nodeStore.AddressNet `json:"nid"`   //节点真实网络地址
	Index uint64               `json:"index"` //节点第几个空间，从1开始,下标为0的节点为实际节点。
	Vid   AddressNetExtend     `json:"vid"`   //vid，虚拟节点网络地址
	// lastContactTimestamp time.Time  //最后检查的时间戳
}

/*
	验证节点id是否合法
*/
func (this *Vnodeinfo) Check() bool {

	var newAddressNetExtend AddressNetExtend

	if this.Index == 0 {
		newAddressNetExtend = AddressNetExtend(this.Nid)
	} else if this.Index > max_vnode_index {
		return false
	} else {
		buf := bytes.NewBuffer(utils.Uint64ToBytes(this.Index))
		buf.Write(this.Nid)
		hashBs := utils.Hash_SHA3_256(buf.Bytes())
		newAddressNetExtend = AddressNetExtend(hashBs)
	}

	if bytes.Equal(newAddressNetExtend, this.Vid) {
		return true
	}
	return false
}

func BuildNodeinfo(index uint64, addrNet nodeStore.AddressNet) *Vnodeinfo {
	vnodeInfo := Vnodeinfo{
		Nid:   addrNet, //
		Index: index,   //节点第几个空间，从0开始。
		// Vid:   addressNetExtend, //vid，虚拟节点网络地址
	}
	if index == 0 {
		vnodeInfo.Vid = AddressNetExtend(addrNet)
	} else if index > max_vnode_index {
		return nil
	} else {
		buf := bytes.NewBuffer(utils.Uint64ToBytes(index))
		buf.Write(addrNet)

		hashBs := utils.Hash_SHA3_256(buf.Bytes())
		addressNetExtend := AddressNetExtend(hashBs)
		vnodeInfo.Vid = addressNetExtend
	}
	return &vnodeInfo
}
