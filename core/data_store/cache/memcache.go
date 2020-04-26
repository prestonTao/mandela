package cache

import (
// "github.com/HouzuoGuo/tiedot/db"
)

var myDBDir = "conf"

type Memcache struct {
	DBDir string `conf` //数据库文件存放文件夹
	// col   *db.Col
	// lock  *sync.RWMutex
	// cache *Cache
}

func (this *Memcache) Add(data map[string]interface{}) {
	// this.col.Insert(data)
}

// func NewMencache() *Memcache {
// 	cache := new(Memcache)
// 	cache.DBDir = "conf"
// 	// os.RemoveAll(cache.DBDir)
// 	db, _ := db.OpenDB(cache.DBDir)
// 	db.Create("db")
// 	cache.col = db.Use("db")
// 	return cache
// }
