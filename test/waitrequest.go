package main

import (
	"fmt"
	"mandela/core/message_center"
	"time"
)

func main() {
	go func() {
		time.Sleep(time.Second * 4)
		message_center.ResponseWait(message_center.MSG_WAIT_http_request, "123", &[]byte{1, 2, 3})
	}()

	bs := message_center.WaitRequest(message_center.MSG_WAIT_http_request, "123")
	fmt.Println(bs)
}
