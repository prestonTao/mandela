package proto

import (
	"mandela/core/engine"
)

const (
	MSGID_safemsginfo      = 1015 //返回节点上传地址信息
	MSGID_safemsginfo_recv = 1016 //返回节点上传地址信息 返回
)

func Register() {
	engine.RegisterMsg(MSGID_safemsginfo, safemsginfo)
	engine.RegisterMsg(MSGID_safemsginfo_recv, safemsginfo_recv)
}
