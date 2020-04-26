package proto

import (
	"errors"
	// "fmt"
	"mandela/core/engine"
	mc "mandela/core/message_center"
	"mandela/core/utils"
)

func getMessage(id *utils.Multihash, content []byte) *mc.Mess {
	mhead := mc.NewMessageHead(id, id, true)
	mbody := mc.NewMessageBody(&content, "", nil, 0)
	message := mc.NewMessage(mhead, mbody)
	message.BuildHash()
	mess := new(mc.Mess)
	mess.Message = message
	return mess
}

//发送消息
func sendMessage(message *mc.Mess) error {
	updateNum(message.Body.Hash, 1)
	if message.Send(MSGID_safemsginfo) {
		//fmt.Println("数据发送成功", id.B58String())
		bs := mc.WaitRequest(mc.CLASS_safemsginfo, message.Body.Hash.B58String())
		if bs != nil {
			//消息回调
			msgid := utils.Multihash(*bs)
			return callbackSuccess(&msgid)
		}
		//fmt.Println("有消息返回", string(*bs))
		//发送消息失败，可能超时
		return errors.New("Failed to send message, may timeout")
	}
	//fmt.Println("数据发送失败", id.B58String())
	//消息发送失败
	return errors.New("Message sending failed")
}

//发送消息 回调
func safemsginfo(c engine.Controller, msg engine.Packet) {
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
	var content []byte
	content = *message.Body.Content
	// fmt.Println(string(content))
	replycontent := []byte(message.Body.Hash)
	//回复给发送者
	mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	mbody := mc.NewMessageBody(&replycontent, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	message = mc.NewMessage(mhead, mbody)
	message.Reply(MSGID_safemsginfo_recv)
}

//发送消息 返回
func safemsginfo_recv(c engine.Controller, msg engine.Packet) {
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
	mc.ResponseWait(mc.CLASS_safemsginfo, message.Body.Hash.B58String(), message.Body.Content)
}
