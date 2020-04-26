package cache_store

import (
	//	"fmt"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"sync"
	"time"
)

var (
	nameCacheLock = new(sync.RWMutex)
	nameCache     = NewCache(10000) //保存域名

	nameCacheMapLock = new(sync.RWMutex)
	nameCacheMap     = make(map[string]chan *nodeStore.TempId)

	OutFindName = make(chan string, 100)
)

/*
	解析一个域名，随机获得域名的地址
*/
func GetAddressInName(name string) *nodeStore.TempId {
	//	fmt.Println("GetAddressInName  11111111111")
	nameCacheLock.Lock()
	value, ok := nameCache.Get(name)
	if !ok {
		//		fmt.Println("GetAddressInName  22222222222")
		//		c := make(chan bool, 1)
		nameCacheMapLock.Lock()
		c, ok := nameCacheMap[name]
		if !ok {
			//			fmt.Println("GetAddressInName  3333333333333")
			c = make(chan *nodeStore.TempId, 1)
			nameCacheMap[name] = c
		}
		nameCacheMapLock.Unlock()
		//		fmt.Println("GetAddressInName  44444444444")
		nameCacheLock.Unlock()
		//		fmt.Println("GetAddressInName  5555555555555")
		OutFindName <- name
		//		fmt.Println("GetAddressInName  6666666666666")
		ticker := time.NewTicker(time.Second * 10)
		select {
		case out := <-c:
			//			fmt.Println("GetAddressInName  7777777777777")
			ticker.Stop()
			return out
		case <-ticker.C:
			//			fmt.Println("GetAddressInName  8888888888888")
			return nil
		}
	}
	nameCacheLock.Unlock()
	//	fmt.Println("GetAddressInName  9999999999999999")
	nameOne := value.(*Name)
	if len(nameOne.Ids) == 0 {
		//		fmt.Println("GetAddressInName  101010101010101010101")
		return nil
	}
	if len(nameOne.Ids) == 1 {
		//		fmt.Println("GetAddressInName  11 11 11 11 11 11")
		return nameOne.Ids[0]
	}
	n := utils.GetRandNum(int64(len(nameOne.Ids)))
	//	fmt.Println("GetAddressInName  end")
	return nameOne.Ids[n]

}

/*
	通知一个域名不存在，不用再等待了
*/
func NoticeNameNotExist(name string) {

	nameCacheMapLock.Lock()
	one, ok := nameCacheMap[name]
	//	fmt.Println("通知域名不存在", name, "1111111111")
	if ok {
		//		fmt.Println("111 删除域名", name)
		delete(nameCacheMap, name)
	}
	nameCacheMapLock.Unlock()
	//	fmt.Println("通知域名不存在", name, "222222222222")
	if !ok {
		//		fmt.Println("通知域名不存在", name, "3333333333333")
		return
	}
ForEnd:
	for {
		//		ticker := time.NewTicker(time.Second * 10)
		select {
		//		case <-ticker.C:
		//			fmt.Println("通知域名不存在", name, "4444444444")
		//			close(one)
		//			break ForEnd
		case one <- nil:
			//			ticker.Stop()
			//			fmt.Println("通知域名不存在", name, "55555555")
		default:
			//			fmt.Println("通知域名不存在", name, "66666666666")
			close(one)
			break ForEnd
		}
	}
}

/*
	添加一个域名
*/
func AddAddressInName(nameStr string, name *Name) {
	//	fmt.Println("添加一个域名", nameStr, "111111111111")
	nameCacheLock.Lock()
	//没有地址id，则不需要保存
	if len(name.Ids) != 0 {
		nameCache.Add(nameStr, name)
	}

	nameCacheMapLock.Lock()
	one, ok := nameCacheMap[nameStr]
	if ok {
		//		fmt.Println("222 删除域名", nameStr)
		delete(nameCacheMap, nameStr)
	}

	nameCacheMapLock.Unlock()
	//	fmt.Println("添加一个域名", nameStr, "222222222222")
	nameCacheLock.Unlock()
	if !ok {
		//		fmt.Println("添加一个域名", nameStr, "333333333333")
		return
	}

	var id *nodeStore.TempId

	if len(name.Ids) == 0 {

	} else if len(name.Ids) == 1 {
		id = name.Ids[0]
	} else {
		n := utils.GetRandNum(int64(len(name.Ids)))
		id = name.Ids[n]
	}
	//	fmt.Println("添加一个域名", nameStr, "444444444444")
ForEnd:
	for {
		//		ticker := time.NewTicker(time.Second * 10)
		select {
		//		case <-ticker.C:
		//			//			close(one)
		//			break ForEnd
		case one <- id:
			//			ticker.Stop()
			break ForEnd
		default:
			break ForEnd
		}
	}
	//	fmt.Println("end")
	//	fmt.Println("添加一个域名", nameStr, "55555555555555")
}

///*
//	等待通道
//*/
//type WaiteChan struct {
//	//	lock *sync.RWMutex
//	ers []chan bool
//}
