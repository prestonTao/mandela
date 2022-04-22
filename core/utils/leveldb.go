package utils

import (
	"mandela/core/engine"
	"bytes"
	"encoding/hex"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

type LevelDB struct {
	path string
	db   *leveldb.DB
	once sync.Once
}

func CreateLevelDB(path string) (*LevelDB, error) {
	lldb := LevelDB{
		path: path,
		// db   *leveldb.DB
		once: sync.Once{},
	}
	err := lldb.InitDB()
	if err != nil {
		return nil, err
	}
	return &lldb, nil
}

//链接leveldb
func (this *LevelDB) InitDB() (err error) {
	this.once.Do(func() {
		// engine.Log.Info("这个方法执行了多少次")
		//没有db目录会自动创建
		this.db, err = leveldb.OpenFile(this.path, nil)
		//	defer db.Close()
		if err != nil {
			return
		}
		// cleanDB()
		return
	})
	return
}

/*
	保存
*/
func (this *LevelDB) Save(id []byte, bs *[]byte) error {

	// engine.Log.Debug("保存到leveldb %s %s", hex.EncodeToString(id), string(*bs))

	//levedb保存相同的key，原来的key保存的数据不会删除，因此保存之前先删除原来的数据
	err := this.db.Delete(id, nil)
	if err != nil {
		// engine.Log.Error("Delete error while saving leveldb", err)
		return err
	}
	if bs == nil {
		err = this.db.Put(id, nil, nil)
	} else {
		err = this.db.Put(id, *bs, nil)
	}
	// if err != nil {
	// 	// engine.Log.Error("Leveldb save error", err)
	// }
	return err
}

/*
	查找
*/
func (this *LevelDB) Find(txId []byte) (*[]byte, error) {
	value, err := this.db.Get(txId, nil)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

/*
	删除
*/
func (this *LevelDB) Remove(id []byte) error {
	return this.db.Delete(id, nil)
}

/*
	初始化数据库的时候，清空一些数据
*/
func (this *LevelDB) cleanDB(name string) {
	// _, err := this.Tags([]byte(config.Name))
	_, err := this.Tags([]byte(name))
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
func (this *LevelDB) Tags(tag []byte) ([][]byte, error) {
	// keys := make([][]byte, 0)
	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
	iter := this.db.NewIterator(nil, nil)
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), tag) {
			// engine.Log.Info("匹配的 %s", iter.Key())
			// keys = append(keys, iter.Key())
			this.db.Delete(iter.Key(), nil)
		}
	}
	iter.Release()
	err := iter.Error()
	return nil, err
}

/*
	打印所有key
*/
func (this *LevelDB) PrintAll() ([][]byte, error) {
	// keys := make([][]byte, 0)
	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
	iter := this.db.NewIterator(nil, nil)
	for iter.Next() {
		engine.Log.Info("%s", hex.EncodeToString(iter.Key()))
	}
	iter.Release()
	err := iter.Error()
	return nil, err
}

/*
	查询指定前缀的key
*/
func (this *LevelDB) FindPrefixKeyAll(tag []byte) ([][]byte, [][]byte, error) {
	keys := make([][]byte, 0)
	values := make([][]byte, 0)
	// iter := db.NewIterator(util.BytesPrefix(tag), nil)
	iter := this.db.NewIterator(nil, nil)
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), tag) {
			// engine.Log.Info("匹配的 %s", iter.Key())
			// engine.Log.Info("匹配的 %s", iter.Value())
			keys = append(keys, iter.Key())
			// db.Delete(iter.Key(), nil)
			value, err := this.db.Get(iter.Key(), nil)
			if err != nil {
				return nil, nil, err
			}
			values = append(values, value)
			// engine.Log.Info("查询的结果 %s", value)
		}
	}
	iter.Release()
	err := iter.Error()
	return keys, values, err
}

/*
	检查是否是空数据库
*/
func (this *LevelDB) CheckNullDB(key []byte) (bool, error) {
	// _, err := this.Find(config.Key_block_start)
	_, err := this.Find(key)
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
	@return    bool    true:存在;false:不存在;
*/
func (this *LevelDB) CheckHashExist(hash []byte) bool {
	// fmt.Println(hex.EncodeToString(hash))
	_, err := this.Find(hash)
	if err != nil {
		if err == leveldb.ErrNotFound {
			// fmt.Println("db 没找到")
			// engine.Log.Debug("db 没找到 %s", hex.EncodeToString(hash))
			return false
		}
		// fmt.Println("db 错误")
		// engine.Log.Debug("db 错误 %s", hex.EncodeToString(hash))
		return true
	}
	// fmt.Println("db 找到了")
	// engine.Log.Debug("db 找到了 %s", hex.EncodeToString(hash))
	return true
}
