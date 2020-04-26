package sqlite3_db

import (
	_ "github.com/go-xorm/xorm"
)

/*
	奖励快照
*/
type Snapshot struct {
	Id          uint64 `xorm:"pk autoincr unique 'id'"` //id
	Addr        string `xorm:"varchar(25) 'addr'"`      //社区节点地址
	StartHeight uint64 `xorm:"int64 'startheight'"`     //快照开始高度
	EndHeight   uint64 `xorm:"int64 'endheight'"`       //快照结束高度
	Reward      uint64 `xorm:"int64 'rewadr'"`          //此快照的总共奖励
	LightNum    uint64 `xorm:"int64 'lightnum'"`        //奖励的轻节点数量
	CreateTime  uint64 `xorm:"created 'createtime'"`    //创建时间，这个Field将在Insert时自动赋值为当前时间
}

/*
	添加一个快照
*/
func (this *Snapshot) Add(s *Snapshot) error {
	_, err := engineDB.Insert(s)
	return err
}

/*
	查询一个地址的最新快照
*/
func (this *Snapshot) Find(addr string) (*Snapshot, error) {
	ss := make([]Snapshot, 0)
	err := engineDB.Where("addr = ?", addr).Limit(1, 0).Desc("endheight").Find(&ss)
	if err != nil {
		return nil, err
	}
	if len(ss) <= 0 {
		return nil, nil
	}
	return &ss[0], nil
}

/*
	轻节点奖励
*/
type Reward struct {
	Id           uint64 `xorm:"pk autoincr unique 'id'"` //id
	Sort         uint64 `xorm:"int64 'sort'"`            //排序字段id
	SnapshotId   uint64 `xorm:"int64 'snapshotid'"`      //奖励快照id
	Addr         string `xorm:"varchar(25) 'addr'"`      //轻节点地址
	Reward       uint64 `xorm:"int64 'rewadr'"`          //自己奖励多少
	Distribution uint64 `xorm:"int64 'distribution'"`    //已经分配的奖励
	CreateTime   uint64 `xorm:"created 'createtime'"`    //创建时间，这个Field将在Insert时自动赋值为当前时间
}

/*
	添加一个轻节点奖励记录
*/
func (this *Reward) Add(r *Reward) error {
	_, err := engineDB.Insert(r)
	return err
}

/*
	查询未分配的奖励记录
*/
func (this *Reward) FindNotSend(id uint64) (*[]Reward, error) {
	// ssAll := make([]Reward, 0)
	// err := engineDB.Where("snapshotid = ?", id).Desc("endheight").Find(&ssAll)
	// if err != nil {
	// 	return nil, nil, err
	// }

	ssNotSend := make([]Reward, 0)
	err := engineDB.Where("snapshotid = ? and rewadr <> distribution", id).Desc("endheight").Find(&ssNotSend)
	if err != nil {
		return nil, err
	}

	return &ssNotSend, nil
}
