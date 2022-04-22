/*
	指数退避算法
*/
package utils

import (
	"sync/atomic"
	"time"
)

type BackoffTimer struct {
	interval []int64   //退避间隔时间，单位：秒
	index    int32     //当前间隔时间下标
	release  chan bool //
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
	timer := time.NewTimer(time.Second * time.Duration(n))
	select {
	case <-timer.C:
	case <-this.release:
		timer.Stop()
	}
	// time.Sleep(time.Second * time.Duration(n))
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
	立即释放暂停的程序
*/
func (this *BackoffTimer) Release() {
	select {
	case this.release <- false:
	default:
	}
}

/*
	间隔n秒后发送一个信号
*/
func NewBackoffTimer(n ...int64) *BackoffTimer {
	return &BackoffTimer{
		interval: n,
		release:  make(chan bool, 1),
	}
}
