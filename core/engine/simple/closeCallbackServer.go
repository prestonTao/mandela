package main

import (
	"fmt"
	msgE "messageEngine"
	"time"
)

func main() {
	server()
}

//---------------------------------------------
//          server
//---------------------------------------------
func server() {
	engine := msgE.NewEngine("interServer")
	engine.RegisterMsg(111, RecvMsg)

	//设置回调函数
	engine.SetCloseCallback(CloseConnCall)
	engine.Listen("127.0.0.1", 9090)
	time.Sleep(time.Second * 1000)
	// session, _ := engine.GetController().GetSession("1")
	// session.Close()
}

func RecvMsg(c msgE.Controller, msg msgE.Packet) {
	fmt.Println(string(msg.Date))
	session, ok := c.GetSession(msg.Name)
	if ok {
		hello := []byte("hello, I'm server")
		session.Send(111, &hello)
		session.Close()
	}
}

func CloseConnCall(name string) {
	fmt.Println("调用了我", name)
}
