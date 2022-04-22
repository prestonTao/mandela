package db

import (
	"fmt"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

var Once_ConnLevelDB sync.Once
var db *leveldb.DB

//链接leveldb
func InitDB(name string) (err error) {
	Once_ConnLevelDB.Do(func() {
		//没有db目录会自动创建
		db, err = leveldb.OpenFile(name, nil)
		//	defer db.Close()
		if err != nil {
			return
		}
		return
	})
	return
}

/*
	保存
*/
func Save(id, bs []byte) error {
	err := db.Put(id, bs, nil)
	if err != nil {
		fmt.Println("Leveldb save error", err)
	}
	return err
}

/*
	查找
*/
func Find(id []byte) ([]byte, error) {
	value, err := db.Get(id, nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

/*
	删除
*/
func Remove(id []byte) error {
	return db.Delete(id, nil)
}

/*
	检查key是否存在
*/
func CheckHashExist(hash []byte) bool {
	_, err := Find(hash)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return false
		}
		return true
	}
	return true
}
