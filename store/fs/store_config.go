package fs

import (
	"mandela/core/virtual_node"
)

/*
	共享目录
*/
type StoreSpaceConfig struct {
	Id         uint64                        `xorm:"pk autoincr unique 'id'"` //id
	VnodeIndex uint64                        `xorm:"int 'vnodeindex'"`        //虚拟节点编号
	VNodeId    virtual_node.AddressNetExtend `xorm:"Blob 'vnodeid'"`          //虚拟节点id
	DBAbsPath  string                        `xorm:"varchar(25) 'dbabspath'"` //空间存储地址
	//	UseSize    uint64 `xorm:"int 'usesize'"`           //已经使用的空间大小
	//	Status       int    `xorm:"int 'status'"`                //好友状态.1=添加好友时，用户不在线;2=申请添加好友状态;3=同意添加;4=;5=;6=;
}

/*
	添加一个共享目录
*/
func (this *StoreSpaceConfig) Add(vnodeIndex uint64, vnodeId virtual_node.AddressNetExtend, dbAbsPath string) error {
	this.VnodeIndex = vnodeIndex
	this.VNodeId = vnodeId
	this.DBAbsPath = dbAbsPath
	_, err := engineDB.Insert(this)
	return err
}

func (this *StoreSpaceConfig) Del(vid virtual_node.AddressNetExtend) {
	engineDB.Where("vnodeid=?", vid).Unscoped().Delete(this)
}

func (this *StoreSpaceConfig) GetAll() ([]StoreSpaceConfig, error) {
	sf := make([]StoreSpaceConfig, 0)
	err := engineDB.Find(&sf)
	return sf, err
}

func (this *StoreSpaceConfig) Count() (int64, error) {
	return engineDB.Count(this)
}
