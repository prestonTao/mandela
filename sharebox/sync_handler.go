package sharebox

// import (
// 	"fmt"
// 	"mandela/core/engine"
// 	mc "mandela/core/message_center"
// 	"mandela/core/utils"
// )

// //分布式存储信息 回调
// func syncFileInfo(c engine.Controller, msg engine.Packet) {
// 	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	form, _ := utils.FromB58String(msg.Session.GetName())
// 	if message.IsSendOther(&form) {
// 		return
// 	}
// 	//发送给自己的，自己处理
// 	if err := message.ParserContent(); err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	sd, err := ParseSyncData(*message.Body.Content)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	//var content []byte
// 	if sd.Type == FileChunkType {
// 		//fmt.Println("收到块索引同步信息")
// 		cid, err := ParseChunkInfoData(sd.Data)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		//content = cid.Json()
// 		cid.AddShareUser(message.Head.Sender)
// 		FD.AddFileChunk(cid)
// 		fmt.Println("收到块索引", cid.CHash.B58String())
// 	}
// 	if sd.Type == FileInfoType {
// 		//fmt.Println("收到文件索引同步信息", message.Head.Sender.B58String())
// 		fid, err := ParseFileInfoData(sd.Data)
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		//content = fid.Json()
// 		//FD.AddFileInfo(fid)
// 		//当前文件归自己管理
// 		fmt.Println("@@@@@收到索引@@@@@")
// 		fmt.Println("@", message.Head.Sender.B58String())
// 		addFileinfoToMyself(message.Head.Sender, fid)
// 	}
// 	//回复给发送者
// 	//	mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
// 	//	mbody := mc.NewMessageBody(&content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
// 	//	message = mc.NewMessage(mhead, mbody)
// 	//	message.Reply(MSGID_syncFileInfo_recv)
// }

// //分布式存储信息 返回
// func syncFileInfo_recv(c engine.Controller, msg engine.Packet) {
// 	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
// 	if err != nil {
// 		fmt.Println("error  1", err)
// 		return
// 	}
// 	form, _ := utils.FromB58String(msg.Session.GetName())
// 	if message.IsSendOther(&form) {
// 		return
// 	}
// 	//发送给自己的，自己处理
// 	if err := message.ParserContent(); err != nil {
// 		engine.NLog.Error(engine.LOG_file, "%s", err.Error())
// 		engine.NLog.Error(engine.LOG_file, "%s", string(msg.Dataplus))
// 		return
// 	}
// 	mc.ResponseWait(mc.CLASS_syncfileinfo, message.Body.Hash.B58String(), message.Body.Content)
// }
