package main

import (
	"mandela/core/utils"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

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
		//		cleanDB()
		return
	})
	return
}
func main() {
	InitDB("D:/test/hzzfiles/data1111")

	oldKey := []byte("key")

	for {
		key := utils.GetRandomDomain()

		fmt.Println("key", key)

		n := 1024 * 1024 //每块1M
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			fmt.Println(err)
			return
		}
		db.Delete(oldKey, nil)
		db.Put([]byte(key), b, nil)
		time.Sleep(time.Second * 5)
		oldKey = []byte(key)
	}

	fmt.Println("end")
}
