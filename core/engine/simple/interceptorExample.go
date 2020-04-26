package main

import (
	// "bytes"
	// "encoding/binary"
	// "errors"
	"fmt"
	msgE "messageEngine"
	// "io"
	// "net"
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
	//添加一个拦截器
	engine.AddInterceptor(new(MyInterceptor))
	engine.Listen("127.0.0.1", 9090)
	time.Sleep(time.Second * 10)
}

//定义一个消息拦截器
type MyInterceptor struct{}

//消息处理前执行
func (this *MyInterceptor) In(c msgE.Controller, msg msgE.Packet) bool {
	fmt.Println("消息处理前执行")
	//返回false，继续执行
	//返回true，不再执行消息模块
	return false
}

//消息处理后执行
func (this *MyInterceptor) Out(c msgE.Controller, msg msgE.Packet) {
	fmt.Println("消息处理后执行")
}

func RecvMsg(c msgE.Controller, msg msgE.Packet) {
	fmt.Println(string(msg.Date))
	session, ok := c.GetSession(msg.Name)
	if ok {
		session.Close()
	}
}

//---------------------------------------------
//          client
//---------------------------------------------
func client() {
	engine := msgE.NewEngine("interClient")
	engine.RegisterMsg(111, RecvMsg)
	engine.AddClientConn("test", "127.0.0.1", 9090, false)

	//给服务器发送消息
	session, _ := engine.GetController().GetSession("test")
	hello := []byte("hello, I'm client")
	session.Send(111, &hello)
	time.Sleep(time.Second * 10)

}
