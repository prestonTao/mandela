package sqlite3_db

import (
	"fmt"

	_ "github.com/go-xorm/xorm"
)

type Downprogress struct {
	Id         uint64  `xorm:"pk autoincr unique 'id'"`   //id
	Hash       string  `xorm:"varchar(25) unique 'hash'"` //文件hash
	State      int     `xorm:"'state'"`                   //下载状态  0 暂停下载 1 下载中 2下载完成
	FileInfo   []byte  `xorm:"varchar(255) 'fileinfo'"`   //文件索引
	Rate       float32 `xorm:"'rate'"`                    //下载进度
	Speed      uint64  `xorm:"'speed'"`                   //下载速度
	UpdateTime uint64  `xorm:"updated 'updated'"`         //修改后自动更新时间
	CreateTime uint64  `xorm:"created 'created'"`         //创建时间
}

func (this *Downprogress) Add() error {
	ok, err := table_downloadprogress.Where("hash=?", this.Hash).Exist(&Downprogress{})
	if err != nil {
		return err
	}
	if ok {
		_, err = table_downloadprogress.Where("hash=?", this.Hash).Update(this)
	} else {
		_, err = table_downloadprogress.Insert(this)
	}

	return err
}
func (this *Downprogress) Update() error {
	_, err := table_downloadprogress.Where("hash=?", this.Hash).Update(this)
	return err
}
func (this *Downprogress) Delete() error {
	_, err := table_downloadprogress.Where("hash=?", this.Hash).Delete(this)
	return err
}
func (this *Downprogress) Get(hash string) (dp Downprogress) {
	_, err := table_downloadprogress.Where("hash = ?", hash).Get(&dp)
	if err != nil {
		fmt.Println(err)
	}
	return
}
func (this *Downprogress) List() (dps []Downprogress) {
	err := table_downloadprogress.Find(&dps)
	if err != nil {
		fmt.Println(err)
	}
	return
}
func (this *Downprogress) Listcomplete() (dps []Downprogress) {
	err := table_downloadprogress.Where("Rate=?", 100).Find(&dps)
	if err != nil {
		fmt.Println(err)
	}
	return
}
