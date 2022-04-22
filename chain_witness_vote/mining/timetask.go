package mining

import (
	"mandela/core/engine"
	"mandela/core/utils"
	"runtime"
	"time"
)

// const (
// 	Task_class_buildBlock = "Task_class_buildBlock" //定时生成块
// )

/*
	定时生成块
	@n    int64    间隔时间
*/
func (this *Witness) SyncBuildBlock(n int64) {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)
	// engine.Log.Info("定时出块 %d", this.Group.Height)
	// this.StopMining = make(chan bool, 1)
	timer := time.NewTimer(time.Second * time.Duration(n))
	select {
	case <-timer.C:
	case <-this.StopMining:
		// fmt.Println("暂停出块")
		// engine.Log.Info("停止出块 %d", this.Group.Height)
		timer.Stop()
		return
	}

	// time.Sleep(time.Second * time.Duration(config.Mining_block_time*n))
	// start := time.Now()
	// this.txCache = utils.NewCache(80000)
	this.BuildBlock()
	// this.txCache.Clear()
	// this.txCache = nil
	// engine.Log.Info("构建区块耗时 %s", time.Now().Sub(start))
}
