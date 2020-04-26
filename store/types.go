package store

import (
	"mandela/core/virtual_node"
	"time"
)

//文件所有者
type FileOwner struct {
	virtual_node.Vnodeinfo
	//	Hash       *nodeStore.AddressNet          //用户hash
	//	Vid        *virtual_node.AddressNetExtend //用户hash
	UpdateTime int64 //最后在线时间
}

//更新在线时间
func (fu *FileOwner) Update() error {
	fu.UpdateTime = time.Now().Unix()
	return nil
}
