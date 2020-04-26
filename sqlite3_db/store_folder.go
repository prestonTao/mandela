package sqlite3_db

import (
	"fmt"

	_ "github.com/go-xorm/xorm"
)

type StoreFolder struct {
	Id         uint64 `xorm:"pk autoincr unique 'id'"` //id
	Name       string `xorm:"'name'"`                  //文件夹名
	ParentId   uint64 `xorm:"'parentid'"`              //父目录ID
	UpdateTime uint64 `xorm:"updated 'updated'"`       //修改后自动更新时间
	CreateTime uint64 `xorm:"created 'created'"`       //创建时间
}

func (sf *StoreFolder) Add() error {
	ok, err := engineDB.Where("name=?", sf.Name).Exist(&StoreFolder{})
	if err != nil {
		return err
	}
	if ok {
		_, err = engineDB.Where("name=?", sf.Name).Update(sf)
	} else {
		_, err = engineDB.Insert(sf)
	}

	return err
}
func (sf *StoreFolder) Update() error {
	_, err := engineDB.Where("id=?", sf.Id).Update(sf)
	return err
}
func (sf *StoreFolder) Delete() error {
	_, err := engineDB.Where("id=?", sf.Id).Delete(sf)
	return err
}
func (sf *StoreFolder) Get() (ok bool, dp StoreFolder) {
	ok, _ = engineDB.Where("id = ?", sf.Id).Get(&dp)
	return
}
func (sf *StoreFolder) List() (dps []StoreFolder) {
	err := engineDB.Where("parentid=?", sf.ParentId).Find(&dps)
	if err != nil {
		fmt.Println(err)
	}
	return
}

type StoreFolderFile struct {
	Id         uint64 `xorm:"pk autoincr unique 'id'"` //id
	Cate       uint64 `xorm:"'cate'"`                  //文件分类ID
	Hash       string `xorm:"'hash'"`                  //文件hash
	FileName   string `xorm:"'filename'"`              //文件名
	Size       int64  `xorm:"'size'"`                  //文件大小
	FileInfo   []byte `xorm:"varchar(255) 'fileinfo'"` //文件索引
	UpdateTime uint64 `xorm:"updated 'updated'"`       //修改后自动更新时间
	CreateTime uint64 `xorm:"created 'created'"`       //创建时间
}

func (sff *StoreFolderFile) Add() error {
	ok, err := engineDB.Where("hash=?", sff.Hash).Exist(&StoreFolderFile{})
	if err != nil {
		return err
	}
	if ok {
		_, err = engineDB.Where("hash=?", sff.Hash).Update(sff)
	} else {
		_, err = engineDB.Insert(sff)
	}
	return err
}
func (sff *StoreFolderFile) Moveto(hash string, cate uint64) bool {
	sf := StoreFolderFile{Cate: cate}
	_, err := engineDB.Where("hash=?", hash).Update(&sf)
	if err != nil {
		return false
	}
	return true
}
func (sff *StoreFolderFile) Delete() error {
	_, err := engineDB.Where("id=?", sff.Id).Delete(sff)
	return err
}

func (sff *StoreFolderFile) List() (sffs []StoreFolderFile) {
	err := engineDB.Where("cate=?", sff.Cate).Find(&sffs)
	if err != nil {
		fmt.Println(err)
	}
	return
}
