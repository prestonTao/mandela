package main

import (
	"fmt"
	msgE "messageEngine"
	"time"
)

func main() {
	go server()
	time.Sleep(time.Second * 5)
	client()
}

//---------------------------------------------
//          server
//---------------------------------------------
func server() {
	engine := msgE.NewEngine("interServer")
	engine.RegisterMsg(111, RecvMsg)
	engine.Listen("127.0.0.1", 9090)
	time.Sleep(time.Second * 10)
}

func RecvMsg(c msgE.Controller, msg msgE.GetPacket) {
	fmt.Println(string(msg.Date))
	session, ok := c.GetSession(msg.Name)
	if ok {
		hello := []byte("hello, I'm server")
		session.Send(111, &hello)
		session.Close()
	}
}

//---------------------------------------------
//          client
//---------------------------------------------
func client() {
	engine := msgE.NewEngine("interClient")
	engine.RegisterMsg(111, ClientRecvMsg)
	engine.AddClientConn("test", "127.0.0.1", 9090, false)

	//给服务器发送消息
	session, _ := engine.GetController().GetSession("test")
	hello := []byte("hello, I'm client")
	session.Send(111, &hello)
	time.Sleep(time.Second * 10)

}

func ClientRecvMsg(c msgE.Controller, msg msgE.GetPacket) {
	fmt.Println(string(msg.Date))
}
