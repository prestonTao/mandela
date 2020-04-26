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

	engine.InitEngine("client_server")
	engine.RegisterMsg(101, hello)

	// engine.SetAuth(new(handlers.NoneAuth))
	// engine.SetCloseCallback(handlers.CloseConnHook)
	//	err := engine.Listen("127.0.0.1", int32(9982))
	//	if err != nil {
	//		panic(err)
	//	}

	for {
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 1000; j++ {
					nameOne, _ := engine.AddClientConn("127.0.0.1", int32(9981), false)
					if session, ok := engine.GetSession(nameOne); ok {
						data := []byte("hello")
						session.Send(101, 0, 0, CryKey, &data)
						session.Close()
					}
				}
			}()
		}
		time.Sleep(time.Minute * 2)
	}

}

func hello(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "client")
}

type FindNode struct {
	Name string `json:"name"`
}
