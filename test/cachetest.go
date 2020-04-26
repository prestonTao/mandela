package main

import (
	"fmt"
	//"time"
	"mandela/core"
	"mandela/core/cache"
	"mandela/core/nodeStore"
)

func main() {
	c := make(chan int)
	core.StartEngine()
	cache.Register()
	cachedata := cache.BuildCacheData([]byte("key"), []byte("value"))
	cachedata.AddOwnId(nodeStore.NodeSelf.IdInfo.Id)
	cachedata.SetTime()
	cache.Save(cachedata)
	cache.AddSyncDataTask([]byte("key"))
	//cache.SyncDataToQuarterLogicIds("key", []byte("value"))
	fmt.Println(cache.Get([]byte("key")))
	fmt.Printf("%+v", cache.GetCacheData([]byte("key")))
	<-c
}
