package main

import (
	"fmt"
	"time"
)

func main() {
	plant_10s()
}

//10秒出块
func plant_10s() {
	blockTime := int64(10)
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", "2020-12-03 10:20:00", time.Local)
	fmt.Println("开始时间", err, startTime.Format("2006-01-02 15:04:05"))

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", "2020-12-08 00:00:00", time.Local)
	fmt.Println("结束时间", err, startTime.Format("2006-01-02 15:04:05"))

	jiange := endTime.Unix() - startTime.Unix()
	fmt.Println("间隔时间 ", jiange)

	lostHeight := jiange / blockTime

	engHeight := plant_20s()

	fmt.Println("预计出块高度 ", engHeight-lostHeight)

}

//20秒出块
func plant_20s() int64 {
	// startHeight := int64(120530)
	startHeight := int64(381960)
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", "2020-12-01 11:12:05", time.Local)
	fmt.Println("开始时间", err, startTime.Format("2006-01-02 15:04:05"))

	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", "2020-12-08 00:00:00", time.Local)
	fmt.Println("结束时间", err, startTime.Format("2006-01-02 15:04:05"))

	jiange := endTime.Unix() - startTime.Unix()
	fmt.Println("间隔时间 ", jiange)

	blockTime := int64(20)
	lostHeight := jiange / blockTime

	fmt.Println("预计出块高度 ", startHeight+lostHeight)

	// end := newTime.Unix() + int64(60*60*24)

	// str = newTime.Format("2006-01-02 15:04:05")
	// fmt.Println(str, newTime.Unix(), start.Unix(), end)
	// //		fmt.Println(newTime.)
	// fmt.Println((end - time.Now().Unix()) / 60 / 60)

	// // time.Sleep(time.Duration((end - time.Now().Unix())) * time.Second)

	return startHeight + lostHeight
}
