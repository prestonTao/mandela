package sqlite3_db

import (
	_ "github.com/go-xorm/xorm"
)

/*
	奖励快照
*/
type SnapshotReward struct {
	Id          uint64 `xorm:"pk autoincr unique 'id'"` //id
	Addr        []byte `xorm:"Blob 'addr'"`             //社区节点地址
	StartHeight uint64 `xorm:"int64 'startheight'"`     //快照开始高度
	EndHeight   uint64 `xorm:"int64 'endheight'"`       //快照结束高度
	Reward      uint64 `xorm:"int64 'reward'"`          //此快照的总共奖励
	LightNum    uint64 `xorm:"int64 'lightnum'"`        //奖励的轻节点数量
	CreateTime  uint64 `xorm:"created 'createtime'"`    //创建时间，这个Field将在Insert时自动赋值为当前时间
}

/*
	添加一个快照
*/
func (this *SnapshotReward) Add(s *SnapshotReward) error {
	_, err := engineDB.Insert(s)
	return err
}

/*
	查询一个地址的最新快照
*/
func (this *SnapshotReward) Find(addr []byte) (*SnapshotReward, error) {
	ss := make([]SnapshotReward, 0)
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
type RewardLight struct {
	Id           uint64 `xorm:"pk autoincr unique 'id'"` //id
	Sort         uint64 `xorm:"int64 'sort'"`            //排序字段id
	SnapshotId   uint64 `xorm:"int64 'snapshotid'"`      //奖励快照id
	Addr         []byte `xorm:"Blob 'addr'"`             //轻节点地址
	Reward       uint64 `xorm:"int64 'reward'"`          //奖励多少
	Txid         []byte `xorm:"Blob 'txid'"`             //交易hash
	LockHeight   uint64 `xorm:"int64 'lock_height'"`     //分配奖励交易锁定高度，超过这个高度，块将不被打包到区块中。
	Distribution uint64 `xorm:"int64 'distribution'"`    //已经分配的奖励
	CreateTime   uint64 `xorm:"created 'createtime'"`    //创建时间，这个Field将在Insert时自动赋值为当前时间
}

/*
	添加一个轻节点奖励记录
*/
func (this *RewardLight) Add(r *RewardLight) error {
	_, err := engineDB.Insert(r)
	return err
}

/*
	查询未分配的奖励记录
*/
func (this *RewardLight) FindNotSend(id uint64) (*[]RewardLight, error) {
	ssNotSend := make([]RewardLight, 0)
	err := engineDB.Where("snapshotid = ? and reward <> distribution", id).Find(&ssNotSend)
	if err != nil {
		return nil, err
	}

	return &ssNotSend, nil
}

/*
	修改正在上链的奖励
*/
func (this *RewardLight) UpdateTxid(id uint64) error {
	_, err := engineDB.Where("id = ?", id).Update(this)
	return err
}

/*
	修改未上链的记录
*/
func (this *RewardLight) RemoveTxid(ids []uint64) error {
	one := new(RewardLight)
	_, err := engineDB.In("id", ids).Cols("txid", "lock_height").Update(one)
	return err
}

/*
	修改已经上链的记录
*/
func (this *RewardLight) UpdateDistribution(id uint64, distribution uint64) error {
	reward := &RewardLight{Distribution: distribution}
	_, err := engineDB.Where("id = ?", id).Update(reward)
	return err
}
