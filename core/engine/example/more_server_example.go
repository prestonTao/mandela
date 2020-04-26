package main

import (
	"fmt"
	"github.com/prestonTao/engine"
	"time"
)

func main() {
	example1()
}

var CryKey = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func example1() {
	engine.GlobalInit("console", "", "debug", 1)

	engineOne := engine.NewEngine("one_server")
	engineOne.RegisterMsg(101, helloOne)

	engineTwo := engine.NewEngine("two_server")
	engineTwo.RegisterMsg(101, helloTwo)

	// engine.SetAuth(new(handlers.NoneAuth))
	// engine.SetCloseCallback(handlers.CloseConnHook)
	err := engineOne.Listen("127.0.0.1", int32(9991))
	if err != nil {
		panic(err)
	}

	err = engineTwo.Listen("127.0.0.1", int32(9992))
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Minute * 10)
}

func helloOne(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "one sever")
	data := []byte("hello")
	msg.Session.Send(101, 0, 0, CryKey, &data)
}

func helloTwo(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "two sever")
	data := []byte("hello")
	msg.Session.Send(101, 0, 0, CryKey, &data)
}
