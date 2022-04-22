package utils

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

//内置时间，精确到秒
var useSystemTime = true //是否使用系统自带的时间
var now = int64(0)
var timeOnec = new(sync.Once)
var syncTimeFun SyncTimeFun

type SyncTimeFun func() (int64, error)

func StartSystemTime() error {
	useSystemTime = true
	// syncTimeFun = getSystemTime
	// timeOnec.Do(systemTimeTicker)
	return nil
}

func StartOtherTime() error {
	syncTimeFun = getSuningTime
	timeOnec.Do(systemTimeTicker)
	return nil
}

func systemTimeTicker() {
	now, _ = syncTimeFun()
	go func() {
		total := int64(0)
		for range time.NewTicker(time.Second).C {
			atomic.AddInt64(&now, 1)
			total = total + 1
			// count := atomic.AddInt64(total, 1)
			if total >= 60*60 {
				nowTime, _ := syncTimeFun()
				atomic.StoreInt64(&now, nowTime)
				total = 0
			}
		}
	}()
}

/*
	获取时间
*/
func GetNow() int64 {
	if useSystemTime {
		now, _ := getSystemTime()
		return now
	}
	return atomic.LoadInt64(&now)
}

/*
	获取苏宁系统时间
*/
func getSuningTime() (int64, error) {
	rep, err := http.Get("http://quan.suning.com/getSysTime.do")
	if err != nil {
		return 0, err
	}
	if rep.StatusCode != 200 {
		return 0, errors.New("suning rpc return status" + rep.Status)
	}
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, rep.Body)
	result := make(map[string]string)
	err = json.Unmarshal(buf.Bytes(), &result)
	if err != nil {
		return 0, err
	}
	nowStr := result["sysTime2"]
	t, err := time.ParseInLocation("2006-01-02 15:04:05", nowStr, time.Local)
	if err != nil {
		return 0, err
	}

	// _, offset := t.Zone()
	// unix := t.Unix() - int64(offset) //本地时间减去8小时

	unix := t.Unix()

	return unix, nil

}

/*
	获取系统时间
*/
func getSystemTime() (int64, error) {
	// t := time.Now()
	// _, offset := t.Zone()
	// now = t.Unix() - int64(offset)
	// return now, nil
	return time.Now().Unix(), nil
}
