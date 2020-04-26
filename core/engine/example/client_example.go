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

	engine.InitEngine("client_server")
	engine.RegisterMsg(101, hello)

	// engine.SetAuth(new(handlers.NoneAuth))
	// engine.SetCloseCallback(handlers.CloseConnHook)
	err := engine.Listen("127.0.0.1", int32(9982))
	if err != nil {
		panic(err)
	}

	nameOne, _ := engine.AddClientConn("127.0.0.1", int32(9991), false)
	if session, ok := engine.GetSession(nameOne); ok {
		data := []byte("hello")
		session.Send(101, 0, 0, CryKey, &data)
	}

	nameTwo, _ := engine.AddClientConn("127.0.0.1", int32(9992), false)
	if session, ok := engine.GetSession(nameTwo); ok {
		data := []byte("hello")
		session.Send(101, 0, 0, CryKey, &data)
	}

	time.Sleep(time.Minute * 10)
}

func hello(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "client")
}

type FindNode struct {
	Name string `json:"name"`
}
