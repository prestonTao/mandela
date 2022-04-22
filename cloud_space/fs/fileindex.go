package fs

import (
	"mandela/config"
	"fmt"
	"path/filepath"

	"github.com/go-xorm/xorm"
)

var engineDB *xorm.Engine

func Init() {
	connect()
}

func connect() {
	var err error
	engineDB, err = xorm.NewEngine("sqlite3", "file:"+filepath.Join(config.Store_dir, StoreFileindex)+"?cache=shared")
	if err != nil {
		fmt.Println(err)
	}
	engineDB.ShowSQL(config.SQL_SHOW)

	//虚拟节点存储空间配置信息表格
	ok, err := engineDB.IsTableExist(StoreSpaceConfig{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(StoreSpaceConfig{}).CreateTable(StoreSpaceConfig{}) //创建表格
	}

	//虚拟节点存储空间配置信息表格
	ok, err = engineDB.IsTableExist(FileindexSelf{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(FileindexSelf{}).CreateTable(FileindexSelf{}) //创建表格
	}
}

type Fileindex struct {
	Id     uint64 `xorm:"pk autoincr unique 'id'"` //id
	Vid    string `xorm:"varchar(25) 'vid'"`       //虚拟节点id
	FileId string `xorm:"varchar(25) 'fileid'"`    //索引哈希值
	Value  []byte `xorm:"Blob 'value'"`            //内容
	//	Status int    `xorm:"int 'status'"`            //好友状态.1=添加好友时，用户不在线;2=申请添加好友状态;3=同意添加;4=;5=;6=;
}

func (this *Fileindex) Add(f *Fileindex) error {
	_, err := engineDB.Insert(f)
	return err
}

func (this *Fileindex) Del(id string) error {
	_, err := engineDB.Where("nodeid = ?", id).Unscoped().Delete(this)
	return err
}

func (this *Fileindex) Update() error {
	_, err := engineDB.Where("nodeid = ?", this.FileId).Update(this)
	return err
}

//修改用户备注，可以为空
func (this *Fileindex) UpdateNoteName() error {
	_, err := engineDB.Nullable("note_name").Where("nodeid = ?", this.FileId).Update(this)
	return err
}
func (this *Fileindex) Getall() ([]Fileindex, error) {

	fs := make([]Fileindex, 0)
	err := engineDB.Find(&fs)
	return fs, err
}

/*
	检查用户id是否存在
*/
func (this *Fileindex) FindById(id string) (*Fileindex, error) {
	fs := make([]Fileindex, 0)
	err := engineDB.Where("nodeid = ?", id).Find(&fs)
	if err != nil {
		return nil, err
	}
	if len(fs) <= 0 {
		return nil, nil
	}
	return &fs[0], nil
}
