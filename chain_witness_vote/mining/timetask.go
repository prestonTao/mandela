package mining

import (
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
	// this.StopMining = make(chan bool, 1)
	timer := time.NewTimer(time.Second * time.Duration(n))
	select {
	case <-timer.C:
	case <-this.StopMining:
		// fmt.Println("暂停出块")
		timer.Stop()
		return
	}

	// time.Sleep(time.Second * time.Duration(config.Mining_block_time*n))
	this.BuildBlock()
}
