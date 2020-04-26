package virtual_node

import (
	"bytes"
)

var vnodeManager = new(VnodeManager)

func init() {

}

/*
	加载节点
*/
func LoadVnode() {
	vnodeManager.Run()
}

/*
	添加一个虚拟节点
*/
func AddVnode() Vnodeinfo {
	return vnodeManager.AddVnode()
}

/*
	删除一个虚拟节点
*/
func DelVnode() Vnodeinfo {
	return vnodeManager.DelVnode()
}

/*
	获取查询vnode管道
*/
func GetFindVnodeChan() chan FindVnodeVO {
	return vnodeManager.findNearVnodeChan
}

// func NewNode() {

// }

/*
	扩展云存储
	系统默认有一个下标为0的虚拟节点映射真实节点。扩展空间从1开始：n > 0
*/
func SetupVnodeNumber(n uint64) {
	vnodeManager.SetupVnodeNumber(n)
}

/*
	查询云存储空间大小
*/
func GetVnodeNumber() []Vnodeinfo {
	return vnodeManager.GetVnodeNumber()
}

/*
	添加虚拟节点的逻辑节点
*/
func AddLogicVnodeinfo(vnode Vnodeinfo) bool {
	return vnodeManager.AddLogicVnodeinfo(vnode)
}

/*
	获得所有逻辑节点，不含自己节点
*/
func GetVnodeLogical() map[string]Vnodeinfo {
	return vnodeManager.GetVnodeLogical()
}

/*
	获得自己管理的所有节点
*/
func GetVnodeSelf() []Vnodeinfo {
	return vnodeManager.GetVnodeSelf()
}

/*
	查找节点id是否是自己的节点
	@return    bool    是否在
*/
func FindInVnodeSelf(id AddressNetExtend) bool {
	for _, one := range vnodeManager.GetVnodeSelf() {
		if bytes.Equal(one.Vid, id) {
			return true
		}
	}
	return false
}

/*
	查找节点id是否是自己的节点
	@return    bool    是否在
*/
func FindInVnodeinfoSelf(id AddressNetExtend) *Vnodeinfo {
	for i, one := range vnodeManager.GetVnodeSelf() {
		if bytes.Equal(one.Vid, id) {
			return &vnodeManager.GetVnodeSelf()[i]
		}
	}
	return nil
}

/*
	在逻辑节点中查找Vnodeinfo
*/
func FindVnodeinfo(vid AddressNetExtend) *Vnodeinfo {
	return vnodeManager.FindVnodeinfo(vid)
}
