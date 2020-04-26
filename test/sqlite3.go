package main

import (
	// "mandela/store/fs"
	"database/sql"
	"fmt"
	"sync"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var once sync.Once

var db *sql.DB

var engineDB *xorm.Engine

func main() {
	Init()

	example()

}

func Init() {
	var err error
	engineDB, err = xorm.NewEngine("sqlite3", "file:sqlite3.db?cache=shared")
	if err != nil {
		fmt.Println(err)
	}
	engineDB.ShowSQL(true)

	ok, err := engineDB.IsTableExist(FileindexSelf{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(FileindexSelf{}).CreateTable(FileindexSelf{}) //创建表格
	}
}

func example() {
	vidBs := []byte("vid")
	valueBs := []byte("value")
	fileidBs := []byte("fileid")
	table := FileindexSelf{
		Name:   "123",
		Vid:    vidBs,    //虚拟节点id
		FileId: fileidBs, //索引哈市值
		Value:  valueBs,  //内容
	}
	table.Add(&table)

	one, _ := table.FindByVid(vidBs)
	fmt.Println("查询结果", one)
}

type FileindexSelf struct {
	Id     uint64 `xorm:"pk autoincr unique 'id'"` //id
	Name   string `xorm:"Text 'name'"`             //真实文件名称
	Vid    []byte `xorm:"Blob 'vid'"`              //虚拟节点id
	FileId []byte `xorm:"Blob 'fileid'"`           //索引哈市值
	Value  []byte `xorm:"Blob 'value'"`            //内容
	//	Status int    `xorm:"int 'status'"`            //好友状态.1=添加好友时，用户不在线;2=申请添加好友状态;3=同意添加;4=;5=;6=;
}

func (this *FileindexSelf) TableName() string {
	return "fileindex_self"
}

func (this *FileindexSelf) Add(f *FileindexSelf) error {
	_, err := engineDB.Insert(f)
	return err
}

func (this *FileindexSelf) Del(fid string) error {
	_, err := engineDB.Where("fileid = ?", fid).Unscoped().Delete(this)
	return err
}

func (this *FileindexSelf) Update() error {
	_, err := engineDB.Where("nodeid = ?", this.FileId).Update(this)
	return err
}

//修改
// func (this *FileindexSelf) UpdateValue(fileid string, value []byte) error {
// 	this.Value = hex. value
// 	_, err := engineDB.Where("fileid = ?", fileid).Update(this)
// 	return err
// }
func (this *FileindexSelf) Getall() ([]FileindexSelf, error) {

	fs := make([]FileindexSelf, 0)
	err := engineDB.Find(&fs)
	return fs, err
}

/*
	检查用户id是否存在
*/
func (this *FileindexSelf) FindByVid(vid []byte) (*FileindexSelf, error) {
	fs := make([]FileindexSelf, 0)
	err := engineDB.Where("vid = ?", vid).Find(&fs)
	if err != nil {
		return nil, err
	}
	if len(fs) <= 0 {
		return nil, nil
	}
	return &fs[0], nil
}
