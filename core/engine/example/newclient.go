package main

import (
	"fmt"
	"time"

	"engine"
)

func main() {

	example1()
}

var CryKey = []byte{0, 0, 0, 0, 0, 0, 0, 0}

func example1() {
	engine.GlobalInit("console", "", "debug", 1)
	engine.InitEngine("client_server")
	//	eg := engine.NewEngine("client_server")
	engine.RegisterMsg(101, hello)

	for {
		go func() {
			for i := 0; i < 3; i++ {
				session, err := engine.AddClientConn("192.168.1.18", 9981, false)
				if err != nil {
					panic(err)
				}
				data := []byte("hello")
				session.Send(101, 0, 0, &data)
				session.Close()
			}
		}()
		time.Sleep(time.Second)
	}

	//	for {

	//		for i := 0; i < 100; i++ {
	//			go func() {
	//				for j := 0; j < 1000; j++ {

	//					//					session, err := engine.AddClientConn("127.0.0.1", 9981, false)
	//					session, err := engine.AddClientConn("192.168.1.18", 9981, false)
	//					if err != nil {
	//						panic(err)
	//						return
	//					}
	//					data := []byte("hello")
	//					session.Send(101, 0, 0, &data)
	//					session.Close()
	//					time.Sleep(time.Nanosecond * 100)
	//				}
	//			}()
	//		}
	//		time.Sleep(time.Second * 60)
	//	}
}

func hello(c engine.Controller, msg engine.Packet) {
	fmt.Println(string(msg.Data), "client")
}

type FindNode struct {
	Name string `json:"name"`
}
