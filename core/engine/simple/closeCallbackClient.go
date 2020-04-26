package main

import (
	"fmt"
	msgE "messageEngine"
	"time"
)

func main() {
	client()
}

//---------------------------------------------
//          client
//---------------------------------------------
func client() {
	engine := msgE.NewEngine("interClient")
	engine.RegisterMsg(111, ClientRecvMsg)

	//设置回调函数
	engine.SetCloseCallback(CloseConnCall)
	engine.AddClientConn("test", "127.0.0.1", 9090, false)

	//给服务器发送消息
	session, _ := engine.GetController().GetSession("test")
	hello := []byte("hello, I'm client")
	session.Send(111, &hello)
	time.Sleep(time.Second * 2000)

}

func ClientRecvMsg(c msgE.Controller, msg msgE.GetPacket) {
	fmt.Println(string(msg.Date))
}

func CloseConnCall(name string) {
	fmt.Println("调用了我", name)
}
