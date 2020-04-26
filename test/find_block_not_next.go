package main

import (
	"mandela/chain_witness_vote/mining"
	"mandela/core/engine"
	"path/filepath"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	path := filepath.Join("wallet", "data")
	InitDB(path)
	FindNotNext()
}

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
		return
	})
	return
}

// 根据Tags遍历
func FindNotNext() {
	// keys := make([][]byte, 0)
	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		//				bs, err := db.Find(iter.Key())
		//		if err != nil {
		//			engine.Log.Info("1查询第 个块错误" + err.Error())
		//			continue
		//		}
		valueBs := iter.Value()
		bh, err := mining.ParseBlockHead(&valueBs)
		if err != nil {
			//			engine.Log.Info("1查询第 个块错误" + err.Error())
			continue
		}
		//		if bh.Height < 10468 || bh.Height > 10657 {
		//			continue
		//		}
		if bh.Nextblockhash == nil {
			bs, err := bh.Json()
			if err != nil {
				engine.Log.Info("2查询第 个块错误" + err.Error())
				continue
			}
			engine.Log.Info(string(*bs))
		}

	}
	iter.Release()
	err := iter.Error()
	engine.Log.Info("3查询第 个块错误" + err.Error())
}
