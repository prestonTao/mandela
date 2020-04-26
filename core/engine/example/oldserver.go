package main

import (
	"fmt"
	"time"

	"github.com/prestonTao/engine"
)

func main() {
	example1()
}

var CryKey = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func example1() {
	engine.GlobalInit("console", "", "debug", 1)

	engine.InitEngine("file_server")
	engine.RegisterMsg(101, hello)

	// engine.SetAuth(new(handlers.NoneAuth))
	// engine.SetCloseCallback(handlers.CloseConnHook)
	err := engine.Listen("127.0.0.1", int32(9981))
	if err != nil {
		panic(err)
	}
	time.Sleep(time.Minute * 10)
}

func hello(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "sever")
	data := []byte("hello")
	msg.Session.Send(101, 0, 0, CryKey, &data)
}

type FindNode struct {
	Name string `json:"name"`
}
