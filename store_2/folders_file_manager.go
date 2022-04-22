package store

import (
	db "mandela/sqlite3_db"
	"errors"
	"fmt"
)

//增加目录
func AddFolder(pid uint64, name string) error {
	sf := db.StoreFolder{ParentId: pid, Name: name}
	err := sf.Add()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

//删除目录
func DelFolder(id uint64) error {
	sf := db.StoreFolder{Id: id}
	err := sf.Delete()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

//修改目录
func UpFolder(id, pid uint64, name string) error {
	sf := db.StoreFolder{Id: id, Name: name}
	if pid > 0 {
		sf.ParentId = pid
	}
	ok, dq := sf.Get()
	if !ok {
		return errors.New("记录不存在")
	}
	if dq.Name == name {
		return errors.New("文件夹名已经存在")
	}
	err := sf.Update()
	if err != nil {
		fmt.Println(err)
	}
	return err
}

//目录列表
func ListFolder(parentid uint64) []db.StoreFolder {
	sf := db.StoreFolder{ParentId: parentid}
	return sf.List()
}

//转移文件目录
func Moveto(hash string, cate uint64) bool {
	sff := &db.StoreFolderFile{}
	return sff.Moveto(hash, cate)
}
