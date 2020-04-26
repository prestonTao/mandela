package main

import (
	"mandela/core/utils"
	"fmt"
	"time"
)

func main() {
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
