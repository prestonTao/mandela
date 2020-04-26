package cache

import (
	"mandela/core/engine"
)

const (
	MSGID_syncData      = 2001 //数据同步
	MSGID_syncData_recv = 2002 //数据同步 返回
)

func Register() {
	//初始化Cache
	initCache()
	//注册消息ID
	engine.RegisterMsg(MSGID_syncData, syncData)
	engine.RegisterMsg(MSGID_syncData_recv, syncData_recv)
}
