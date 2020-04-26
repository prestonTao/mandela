package cachedata

import (
	// "fmt"
	"mandela/core/utils"
	"time"
)

var (
	TimeInterval               = 20               //数据同步时间间隔,单位秒
	TimeIntervalClearCache     = 30 * time.Second //节点没有收到数据，清理Cache的时间间隔，1/4节点最近节点重新上线，则本节点将收不到数据，则清理。
	TimeIntervalClearCacheData = 15 * time.Second //节点没有收到数据，清理CachedData的时间间隔,即删除数据清理
)

const (
	Task_class_share_cache_data       = "Task_class_share_cache_data"       //定时发送自己上传的文件索引信息
	Task_class_share_clear_cache      = "Task_class_share_clear_cache"      //节点没有收到数据，清理Cache
	Task_class_share_clear_cache_data = "Task_class_share_clear_cache_data" //节点没有收到数据，清理CacheData
)

var task *utils.Task

func Init() {
	initTask()
}
func initTask() {
	task = utils.NewTask(taskFunc)
	//增加定时清理
	task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_share_clear_cache, "cache")
	task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_share_clear_cache_data, "cache_data")
}

//加入定时器
func AddSyncDataTask(params []byte) {
	//fmt.Println("加入定时器...")
	task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_share_cache_data, string(params))
}
func taskFunc(class, params string) {
	switch class {
	case Task_class_share_cache_data:
		go func() {
			//fmt.Println(class, params)
			taskSyncData([]byte(params))
			task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_share_cache_data, string(params))
		}()
	case Task_class_share_clear_cache:
		go func() {
			taskClearCache()
			task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_share_clear_cache, string(params))
		}()
	case Task_class_share_clear_cache_data:
		go func() {
			taskClearCacheData()
			task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_share_clear_cache_data, string(params))
		}()
	default:
		// fmt.Println("未注册的定时器类型", class)

	}
}
func taskSyncData(key []byte) {
	//fmt.Println("广播数据", string(key))
	//SyncDataToQuarterLogicIds(key)
}

//清理Cache
func taskClearCache() {
	if time.Now().Sub(Caches.Time) > TimeIntervalClearCache {
		Caches = NewCache()
	}
}

//清理CacheData
func taskClearCacheData() {
	Caches.Data.Range(func(k, v interface{}) bool {
		return true
	})
}
