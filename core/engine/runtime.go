package engine

import (
	// "strconv"
	"sync"
	"time"
)

var runtimeMap = new(sync.Map)

func init() {
	go func() {
		time.Sleep(time.Minute * 30)
		runtimeMap.Range(func(k, v interface{}) bool {
			t := v.(*time.Time)
			Log.Info("RuntimeMapOne :%s %s", k.(string), t)
			return true
		})
	}()
}

func AddRuntime(file string, line int, goroutineId string) {
	// now := time.Now()
	// key := file + "_" + strconv.Itoa(line) + "_" + goroutineId
	// Log.Info("AddRuntime :%s", key)
	// runtimeMap.Store(key, &now)
}

func DelRuntime(file string, line int, goroutineId string) {
	// key := file + "_" + strconv.Itoa(line) + "_" + goroutineId
	// Log.Info("DelRuntime :%s", key)
	// runtimeMap.Delete(key)
}
