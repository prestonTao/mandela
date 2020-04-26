package main

import (
	"fmt"
	"time"
	"mandela/core/utils"
)

func main() {
	utils.AddTimetask(time.Now().Unix()+int64(10), haha, "add", "nihao")
	time.Sleep(time.Minute)
	fmt.Println("end")
}

func haha(class, params string) {
	fmt.Println(class, params)
	utils.AddTimetask(time.Now().Unix()+int64(10), haha, "add", "nihao")
}
