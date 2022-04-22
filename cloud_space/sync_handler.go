package cloud_space

import (
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"encoding/hex"
)

//分布式存储信息 回调
func syncFileInfo(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	sd, err := ParseSyncData(*message.Body.Content)
	//	if err != nil {
	//		// fmt.Println(err)
	//		return
	//	}
	//	//var content []byte
	//	if sd.Type == FileChunkType {
	//		//fmt.Println("收到块索引同步信息")
	//		cid, err := ParseChunkInfoData(sd.Data)
	//		if err != nil {
	//			// fmt.Println(err)
	//			return
	//		}
	//		//content = cid.Json()
	//		cid.AddShareUser(message.Head.Sender)
	//		FD.AddFileChunk(cid)
	//		//fmt.Println("收到块索引", cid.CHash.B58String())
	//	}
	//	if sd.Type == FileInfoType {
	//		//fmt.Println("收到文件索引同步信息", message.Head.Sender.B58String())
	//		fid, err := ParseFileInfoData(sd.Data)
	//		if err != nil {
	//			// fmt.Println(err)
	//			return
	//		}
	//		//content = fid.Json()
	//		//FD.AddFileInfo(fid)
	//		//当前文件归自己管理
	//		fmt.Println("@@@@@收到索引@@@@@")
	//		fmt.Println("@", message.Head.Sender.B58String())
	//		addFileindexToMyself(message.Head.Sender, fid)
	//	}
	//	bs := []byte("ok")
	//	message_center.SendP2pReplyMsg(message, config.MSGID_store_syncFileInfo_recv, &bs)
}

//分布式存储信息 返回
func syncFileInfo_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(message_center.CLASS_syncfileinfo, hex.EncodeToString(message.Body.Hash), message.Body.Content)
}
