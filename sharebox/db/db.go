package db

// import (
// 	// "mandela/config"
// 	"sync"

// 	"github.com/syndtr/goleveldb/leveldb"
// )

// var Once_ConnLevelDB sync.Once
// var db *leveldb.DB

// //链接leveldb
// func InitDB(name string) (err error) {
// 	Once_ConnLevelDB.Do(func() {
// 		//没有db目录会自动创建
// 		db, err = leveldb.OpenFile(name, nil)
// 		//	defer db.Close()
// 		if err != nil {
// 			return
// 		}
// 		return
// 	})
// 	return
// }

// /*
// 	连接levelDB
// */
// func connLevelDB() {

// }

// /*
// 	保存
// */
// func Save(id []byte, bs *[]byte) error {
// 	return db.Put(id, *bs, nil)
// }

// /*
// 	查找
// */
// func Find(txId []byte) (*[]byte, error) {
// 	value, err := db.Get(txId, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &value, nil
// }

// /*
// 	删除
// */
// func Remove(id []byte) error {
// 	return db.Delete(id, nil)
// }

/*
	检查是否是空数据库
*/
// func CheckNullDB() (bool, error) {
// 	_, err := Find(config.Key_block_start)
// 	if err != nil {
// 		if err == leveldb.ErrNotFound {
// 			//认为这是一个空数据库
// 			return true, nil
// 		}
// 		return false, err
// 	}
// 	return false, nil
// }
