package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
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
	engine.RegisterMsg(111, CreateUserMsg)
	engine.Listen("127.0.0.1", 9090)

	//初始化数据库
	mysqlEngine, err := xorm.NewEngine("mysql", "root:root@/test?charset=utf8")
	if err != nil {
		fmt.Println("连接数据库失败")
	}
	//设置连接池
	mysqlEngine.SetMaxOpenConns(5)
	//创建表
	mysqlEngine.CreateTables(&User{})

	ctrl := engine.GetController()
	//把mysqlEngine放入数据共享区域
	ctrl.SetAttribute("mysqlEngine", mysqlEngine)

	time.Sleep(time.Second * 10)
}

func CreateUserMsg(c msgE.Controller, msg msgE.Packet) {
	fmt.Println(string(msg.Date))
	user := new(User)
	user.Name = string(msg.Date)

	//从数据共享区域取出mysqlEngine
	obj := c.GetAttribute("mysqlEngine")
	mysqlEngine, _ := obj.(*xorm.Engine)

	mysqlEngine.Insert(user)
	fmt.Println("保存数据成功")
}

type User struct {
	Id   int64
	Name string `xorm:"c_name"`
}

//---------------------------------------------
//          client
//---------------------------------------------
func client() {
	engine := msgE.NewEngine("interClient")
	engine.AddClientConn("test", "127.0.0.1", 9090, false)
	session, _ := engine.GetController().GetSession("test")
	hello := []byte("taopopo@126.com")
	session.Send(111, &hello)

	time.Sleep(time.Second * 10)

}
