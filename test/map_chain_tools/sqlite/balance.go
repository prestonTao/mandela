package sqlite

import (
	_ "github.com/go-xorm/xorm"
)

type Balance struct {
	Id          uint64 `xorm:"pk autoincr unique 'id'"` //id
	Addr        string `xorm:"varchar(25) 'addr'"`      //收款地址
	Balance     uint64 `xorm:"int 'balance'"`           //余额
	IsAllocated int    `xorm:"int 'is_allocated'"`      //是否分配 1=未分配；2=正在分配；3=已分配；
}

func (this *Balance) Add(addr string, balance uint64) error {
	this.Addr = addr
	this.Balance = balance
	this.IsAllocated = 1
	_, err := engineDB.Insert(this)
	return err
}

func (this *Balance) Del(id string) error {
	_, err := engineDB.Where("nodeid = ?", id).Unscoped().Delete(this)
	return err
}

/*
	修改余额
*/
func (this *Balance) UpdateBalance(addr string, balance uint64) error {
	this.Balance = balance
	_, err := engineDB.Cols("balance").Where("addr = ?", addr).Update(this)
	return err
}

/*
	修改为已经分配
*/
func (this *Balance) UpdateIsAllocated(addr string) error {
	this.IsAllocated = 3
	_, err := engineDB.Where("addr = ?", addr).Update(this)
	return err
}

func (this *Balance) Getall() ([]Balance, error) {

	fs := make([]Balance, 0)
	err := engineDB.Find(&fs)
	return fs, err
}

/*
	检查用户id是否存在
*/
func (this *Balance) FindByAddr(addr string) (*Balance, error) {
	fs := make([]Balance, 0)
	err := engineDB.Where("addr = ?", addr).Find(&fs)
	if err != nil {
		return nil, err
	}
	if len(fs) <= 0 {
		return nil, nil
	}
	return &fs[0], nil
}
