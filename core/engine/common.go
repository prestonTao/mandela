package engine

import (
	"fmt"
	// "fmt"
	"runtime"
	"time"
)

type TimeOut struct {
	isTimeOutChan chan bool
	duration      time.Duration
	f             func()
}

func (this *TimeOut) Do(duration time.Duration) bool {
	this.duration = duration
	go this.run()

	select {
	case <-this.isTimeOutChan:
		close(this.isTimeOutChan)
		return false
	case <-time.After(this.duration):
		return true
	}
}

func (this *TimeOut) run() {
	this.f()
	select {
	case this.isTimeOutChan <- false:
	default:
	}

}

func NewTimeOut(f func()) *TimeOut {
	to := TimeOut{
		isTimeOutChan: make(chan bool),
		f:             f,
	}
	return &to
}

/*
	错误处理
*/
func PrintPanicStack() {
	if x := recover(); x != nil {
		fmt.Println(x)
		for i := 0; i < 10; i++ {
			funcName, file, line, ok := runtime.Caller(i)
			if ok {
				// fmt.Println("frame :[func:%s,file:%s,line:%d]\n", i, runtime.FuncForPC(funcName).Name(), file, line)
				Log.Error("%d frame :[func:%s,file:%s,line:%d]\n", i, runtime.FuncForPC(funcName).Name(), file, line)

				//				fmt.Println("frame " + strconv.Itoa(i) + ":[func:" + runtime.FuncForPC(funcName).Name() + ",file:" + file + ",line:" + line + "]\n")
			}
		}
	}
}
