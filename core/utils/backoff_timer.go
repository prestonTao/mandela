/*
	指数退避算法
*/
package utils

import (
	"sync/atomic"
	"time"
)

type BackoffTimer struct {
	interval []int64 //退避间隔时间，单位：秒
	index    int32   //当前间隔时间下标
}

/*
	等待时间
	@return    int64    等待的时间
*/
func (this *BackoffTimer) Wait() int64 {
	n := this.interval[this.index]
	if int(atomic.LoadInt32(&this.index)+1) < len(this.interval) {
		atomic.AddInt32(&this.index, 1)
	}
	time.Sleep(time.Second * time.Duration(n))
	return n
}

/*
	重置
	间隔时间从头开始
*/
func (this *BackoffTimer) Reset() {
	atomic.StoreInt32(&this.index, 0)
}

/*
	间隔n秒后发送一个信号
*/
func NewBackoffTimer(n ...int64) *BackoffTimer {
	return &BackoffTimer{
		interval: n,
	}
}
