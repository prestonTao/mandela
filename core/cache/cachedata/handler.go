package cachedata

import (
	// "fmt"
	"mandela/core/engine"
	mc "mandela/core/message_center"
	"mandela/core/nodeStore"
	"mandela/core/utils"
)

func syncData(c engine.Controller, msg engine.Packet) {
	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		// fmt.Println(err)
		return
	}
	form, _ := utils.FromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//发送给自己的，自己处理
	if err := message.ParserContent(); err != nil {
		// fmt.Println(err)
		return
	}
	content := message.Body.Content
	//fmt.Println("接受到的数据", string(*content))
	cachedata, err := Parse(*content)
	if err != nil {
		// fmt.Println("数据解析错误", err)
		return
	}
	if cachedata.Del {
		Del(cachedata.Kbyte)
	}
	//fmt.Printf("%+v", cachedata)
	//保存并广播出去
	// fmt.Println("接收到数据", string(cachedata.Kbyte), string(cachedata.Value))
	//	//加入定时器,广播出去
	//	if len(Get(cachedata.Kbyte)) == 0 {
	//		go AddSyncDataTask(cachedata.Kbyte)
	//	}
	//如果是共享者，则更新时间
	if checkOwn(nodeStore.NodeSelf.IdInfo.Id, cachedata.Ownid) {
		cachedata.SetTime()
		Save(cachedata)
	} else {
		Save(cachedata)
	}
	//更新时间
	UpTime()
	//回复给发送者
	//	mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	//	mbody := mc.NewMessageBody(content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	//	message = mc.NewMessage(mhead, mbody)
	//	message.Reply(MSGID_syncData_recv)
}
func syncData_recv(c engine.Controller, msg engine.Packet) {
	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		// fmt.Println("error  1", err)
		return
	}
	form, _ := utils.FromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//发送给自己的，自己处理
	if err := message.ParserContent(); err != nil {
		engine.NLog.Error(engine.LOG_file, "%s", err.Error())
		engine.NLog.Error(engine.LOG_file, "%s", string(msg.Dataplus))
		return
	}
	mc.ResponseWait(mc.CLASS_syncdata, message.Body.Hash.B58String(), message.Body.Content)
}
