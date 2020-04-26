package main

import (
	"fmt"
	"mandela/core/utils"
	"time"
)

var poll = utils.NewPollManager()

func main() {
	//	poll.InitClass("vc")
	//	fmt.Println("打印是为等待")

	fmt.Println(poll.Vote("vc", "vote", "1"))
	fmt.Println(poll.Vote("vc", "vote", "2"))
	fmt.Println(poll.Vote("vc", "vote", "3"))
	fmt.Println(poll.Vote("vc", "vote", "4"))
	fmt.Println(poll.Vote("vc", "vote", "5"))
	fmt.Println(poll.Vote("vc", "vote", "6"))
	fmt.Println(poll.Vote("vc", "vote", "7"))
	fmt.Println(poll.Vote("vc", "vote", "8"))
	fmt.Println(poll.Vote("vc", "vote", "9"))
	fmt.Println(poll.Vote("vc", "vote", "10"))
	fmt.Println(poll.Vote("vc", "vote", "11"))
	fmt.Println(poll.Vote("vc", "vote", "12"))
	time.Sleep(time.Hour)
}
