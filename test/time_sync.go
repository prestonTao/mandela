package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {

	systemTime := time.Now().Unix()
	fmt.Println("系统时间", systemTime, time.Unix(systemTime, 0).Format("2006-01-02 15:04:05"))

	otherTime, _ := getSuningTime()
	fmt.Println("苏宁时间", otherTime, time.Unix(otherTime, 0).Format("2006-01-02 15:04:05"))

	fmt.Println(time.Unix(systemTime, 0).Format("2006-01-02 15:04:05"))
}

func simple1() {
	fmt.Println("start")
	err := utils.StartOtherTime()
	if err != nil {
		fmt.Println("获取系统时间失败", err.Error())
		return
	}
	fmt.Println("成功获取其它源时间")
	systemTime := time.Now().Unix() - 60*60*8
	otherTime := utils.GetNow()
	fmt.Println("系统时间", systemTime, time.Unix(systemTime, 0).Format("2006-01-02 15:04:05"))
	fmt.Println("其它源时间", otherTime, time.Unix(otherTime, 0).Format("2006-01-02 15:04:05"))
	fmt.Println("时间差", systemTime-otherTime)
	for range time.NewTicker(time.Minute).C {
		systemTime = time.Now().Unix() - 60*60*8
		otherTime = utils.GetNow()
		fmt.Println("系统时间", systemTime, time.Unix(systemTime, 0).Format("2006-01-02 15:04:05"))
		fmt.Println("其它源时间", otherTime, time.Unix(otherTime, 0).Format("2006-01-02 15:04:05"))
		fmt.Println("时间差", systemTime-otherTime)
	}
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
