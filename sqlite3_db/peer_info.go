package sqlite3_db

import (
	_ "github.com/go-xorm/xorm"
)

type PeerInfo struct {
	Id         string `xorm:"varchar(25) pk notnull unique 'id'"` //网络节点id
	SuperId    string `xorm:"varchar(25) 'sid'"`                  //网络节点的超级节点id
	Puk        string `xorm:"varchar(25) 'puk'"`                  //身份公钥
	SK         string `xorm:"varchar(25) 'sk'"`                   //共享密钥
	SharedHka  string `xorm:"varchar(25) 'ska'"`                  //共享密钥a
	SharedNhkb string `xorm:"varchar(25) 'skb'"`                  //共享密钥b
}

/*
	添加节点信息
*/
func (this *PeerInfo) Add() error {
	_, err := engineDB.Insert(this)
	return err
}

/*
	查询节点信息
*/
func (this *PeerInfo) FindByid(id string) (*PeerInfo, error) {
	fs := make([]PeerInfo, 0)
	err := engineDB.Where("id = ?", id).Find(&fs)
	if err != nil {
		return nil, err
	}
	if len(fs) <= 0 {
		return nil, nil
	}
	return &fs[0], nil
}

/*
	修改节点信息
*/
func (this *PeerInfo) Update() error {
	_, err := engineDB.Id(this.Id).Update(this)
	return err
}
