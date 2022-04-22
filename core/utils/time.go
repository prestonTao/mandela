package utils

import (
	"sync"
	"time"
)

func TimeFormatToNanosecond() string {
	return time.Now().Format("2006-01-02 15:04:05.999999999")
}

func TimeFormatToNanosecondStr() string {
	return time.Now().Format("20060102150405999999999")
}

func FormatTimeToSecond(now time.Time) string {
	return now.Format("2006-01-02 15:04:05")
}

var timetokenChanLock = new(sync.RWMutex)
var timetokenChan = make(map[string]chan bool) // new(sync.Map)

func SetTimeToken(class string, t time.Duration) {
	timetokenChanLock.Lock()
	_, ok := timetokenChan[class]
	if !ok {
		flowChan := make(chan bool, 1)
		timetokenChan[class] = flowChan
		go func() {
			for range time.NewTicker(t).C {
				select {
				case flowChan <- false:
				default:
				}
			}
		}()
	}
	timetokenChanLock.Unlock()
}

func GetTimeToken(class string, wait bool) (allow bool) {
	timetokenChanLock.RLock()
	flowChan, ok := timetokenChan[class]
	if ok {
		select {
		case <-flowChan:
			allow = true
		default:
			allow = false
		}
	} else {
		//未添加此种类型的时间令牌
		allow = true
	}
	timetokenChanLock.RUnlock()
	if wait && !allow {
		<-flowChan
		allow = true
	}
	return
}
