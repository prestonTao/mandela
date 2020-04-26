package cachedata

import (
	// "fmt"
	"mandela/core/utils"
	"time"
)

var (
	Caches *Cache
)

//定时同步数据
func TimingSynchronization(hash []byte) {
	//每10秒同步一次数据
	// fmt.Println("加入定时器")
	task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_share_cache_data, string(hash))
}
func initCache() {
	Caches = NewCache()
}

//根据key/value生成cachedata
func BuildCacheData(key, value []byte) *CacheData {
	return newCacheData(key, value)
}

//保存cachedata
func Save(cd *CacheData) {
	Caches.Save(cd)
}

//删除cachedata
func Del(key []byte) {
	Caches.Del(key)
}

//根据原始key获取data
func Get(key []byte) []byte {
	return Caches.Get(key)
}

//根据原始hash id获取data
func GetByHash(id *utils.Multihash) []byte {
	return Caches.GetByHash(id)
}

//根据key获取cachedata
func GetCacheData(key []byte) *CacheData {
	return Caches.GetCacheData(key)
}

//根据hash id 获取cachedata
func GetCacheDataByHash(id *utils.Multihash) *CacheData {
	return Caches.GetCacheDataByHash(id)
}
func Hash(key []byte) *utils.Multihash {
	return buildHash(key)
}
func UpTime() {
	Caches.UpTime()
}
