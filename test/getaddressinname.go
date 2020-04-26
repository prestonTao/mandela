package main

import (
	"fmt"
	"mandela/core/cache_store"
	"mandela/core/nodeStore"
	"strconv"
)

func main() {
	go read()
	Start()
}

//var lock = new(sync.RWMutex)
//var count int64 = 0

func Start() {
	//	c := make(chan bool, 1)

	for i := 0; i < 10; i++ {

		name := cache_store.GetAddressInName(strconv.Itoa(i))
		fmt.Println(name.SuperPeerId.GetIdStr())
	}

}

func read() {
	for one := range cache_store.OutFindName {
		//		fmt.Println("开始查询域名 11111111")
		id := []byte{1, 1, 1, 1, 1}
		tid := nodeStore.NewTempId(id, id)
		name := cache_store.NewName(one, []*nodeStore.TempId{tid}, id)
		//		fmt.Println("开始查询域名 22222222222")
		cache_store.AddAddressInName(one, name)
		//		fmt.Println("开始查询域名 3333333333")

		//		cache_store.NoticeNameNotExist(one)
	}

	//	cache_store.GetAddressInName(name)
	//	atomic.AddInt64(&count, 1)
}
