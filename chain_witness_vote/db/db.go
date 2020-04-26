package db

import (
	"mandela/config"
	"mandela/core/engine"
	"bytes"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	// "github.com/syndtr/goleveldb/leveldb/util"
)

var Once_ConnLevelDB sync.Once
var db *leveldb.DB

//链接leveldb
func InitDB(name string) (err error) {
	Once_ConnLevelDB.Do(func() {
		// engine.Log.Info("这个方法执行了多少次")
		//没有db目录会自动创建
		db, err = leveldb.OpenFile(name, nil)
		//	defer db.Close()
		if err != nil {
			return
		}
		cleanDB()
		return
	})
	return
}

/*
	初始化数据库的时候，清空一些数据
*/
func cleanDB() {
	_, err := Tags([]byte(config.Name))
	if err == nil {
		// for _, one := range keys {
		// 	engine.Log.Info("开始删除域名 %s", hex.EncodeToString(one))
		// 	err = Remove(one)
		// 	if err != nil {
		// 		engine.Log.Info("删除错误 %s", err.Error())
		// 	}
		// }
		// for _, one := range keys {
		// 	value, _ := Find(one)
		// 	if value != nil {
		// 		engine.Log.Info("查找域名 %s", hex.EncodeToString(one))

		// 	}

		// }
	}
	// engine.Log.Info("删除域名 end")

	// db.
}

// 根据Tags遍历
func Tags(tag []byte) ([][]byte, error) {
	// keys := make([][]byte, 0)
	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), tag) {
			// engine.Log.Info("匹配的 %s", iter.Key())
			// keys = append(keys, iter.Key())
			db.Delete(iter.Key(), nil)
		}
	}
	iter.Release()
	err := iter.Error()
	return nil, err
}

/*
	连接levelDB
*/
func connLevelDB() {

}

/*
	保存
*/
func Save(id []byte, bs *[]byte) error {
	//levedb保存相同的key，原来的key保存的数据不会删除，因此保存之前先删除原来的数据
	err := db.Delete(id, nil)
	if err != nil {
		engine.Log.Error("Delete error while saving leveldb", err)
		return err
	}
	err = db.Put(id, *bs, nil)
	if err != nil {
		engine.Log.Error("Leveldb save error", err)
	}
	return err
}

/*
	查找
*/
func Find(txId []byte) (*[]byte, error) {
	value, err := db.Get(txId, nil)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

/*
	删除
*/
func Remove(id []byte) error {
	return db.Delete(id, nil)
}

/*
	检查是否是空数据库
*/
func CheckNullDB() (bool, error) {
	_, err := Find(config.Key_block_start)
	if err != nil {
		if err == leveldb.ErrNotFound {
			//认为这是一个空数据库
			return true, nil
		}
		return false, err
	}
	return false, nil
}

/*
	检查key是否存在
*/
func CheckHashExist(hash []byte) bool {
	// fmt.Println(hex.EncodeToString(hash))
	_, err := Find(hash)
	if err != nil {
		if err == leveldb.ErrNotFound {
			// fmt.Println("db 没找到")
			return false
		}
		// fmt.Println("db 错误")
		return true
	}
	// fmt.Println("db 找到了")
	return true
}

/*
	保存区块高度对应的区块hahs
*/
//func SaveBlockHeight(height uint64, id *[]byte) error {
//	return Save([]byte(config.BlockHeight+strconv.Itoa(int(height))), id)
//}

/*
	查询区块高度对应的区块hahs
*/
//func FindBlockHeight(height uint64) (*[]byte, error) {
//	return Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
//}
