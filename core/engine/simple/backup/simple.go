package main

import (
	// _ "./cxv"
	server "gameServerEngine"
	"gameServerEngine/net"
	// "github.com/cuixin/csv4g"
	"fmt"
)

func main() {
	//指定配置文件的路径，若不指定，默认为config.ini
	server.ConfigPath = "config/config.ini"
	//添加一个消息处理逻辑
	server.AddRouter(6, Login)
	//添加一个启动钩子
	server.AddStartHook(CSVHook)
	//启动服务器
	server.StartUP()
}

func Login(c server.Controller, msg net.Packet) {
	//得到数据库id和表格id
	// c.GetDatabaseId()
	//得到数据库连接
	// c.GetDBConn()
	//得到网络连接
	// c.GetNet()
	//得到其他服务器连接
	// c.GetClient()
	//得到值
	// c.GetAttribute("item")
}

func CSVHook(c server.Controller) {
	item := new(Item)
	c.SetAttribute("item", item)
	fmt.Println("运行了启动钩子")
}

type Item struct {
	SeqID   int32  //道具ID
	Name    string //道具名
	Type    int32  //道具类型
	Count   int32  //数量
	Skill   int32  //对应技能
	Life    int32  //生命
	Magic   int32  //魔法
	Attack  int32  //攻击
	Defense int32  //防御
	dodge   int32  //闪避
}
