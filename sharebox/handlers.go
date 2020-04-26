package sharebox

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	sconfig "mandela/sharebox/config"
	"bytes"
	"encoding/hex"

	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

/*
	注册这些消息id
*/
func RegisterMsgid() {
	// engine.RegisterMsg(MSGID_addFileShare, AddFileShare)
	message_center.Register_search_all(config.MSGID_sharebox_addFileShare, AddFileShare)
	// engine.RegisterMsg(MSGID_addFileShare_recv, AddFileShare_recv)
	message_center.Register_p2p(config.MSGID_sharebox_addFileShare_recv, AddFileShare_recv)
	// engine.RegisterMsg(MSGID_findFileinfo, FindFileinfoHandler)
	message_center.Register_search_all(config.MSGID_sharebox_findFileinfo, FindFileinfoHandler)
	// engine.RegisterMsg(MSGID_findFileinfo_recv, FindFileinfo_recv)
	message_center.Register_p2p(config.MSGID_sharebox_findFileinfo_recv, FindFileinfo_recv)
	// engine.RegisterMsg(MSGID_getFilesize, FindFilesize)
	message_center.Register_p2p(config.MSGID_sharebox_getFilesize, FindFilesize)
	// engine.RegisterMsg(MSGID_getFilesize_recv, FindFilesize_recv)
	message_center.Register_p2p(config.MSGID_sharebox_getFilesize_recv, FindFilesize_recv)
	// engine.RegisterMsg(MSGID_downloadFileChunk, DownloadFilechunk)
	message_center.Register_p2p(config.MSGID_sharebox_downloadFileChunk, DownloadFilechunk)
	// engine.RegisterMsg(MSGID_downloadFileChunk_recv, DownloadFilechunk_recv)
	message_center.Register_p2p(config.MSGID_sharebox_downloadFileChunk_recv, DownloadFilechunk_recv)
	// engine.RegisterMsg(MSGID_getUploadinfo, Uploadinfo)
	message_center.Register_p2p(config.MSGID_sharebox_getUploadinfo, Uploadinfo)
	// engine.RegisterMsg(MSGID_getUploadinfo_recv, Uploadinfo_recv)
	message_center.Register_p2p(config.MSGID_sharebox_getUploadinfo_recv, Uploadinfo_recv)
	// engine.RegisterMsg(MSGID_syncFileInfo, syncFileInfo)
	// engine.RegisterMsg(MSGID_syncFileInfo_recv, syncFileInfo_recv)
	// engine.RegisterMsg(MSGID_getfourNodeinfo, GetfourNodeinfo)
	// engine.RegisterMsg(MSGID_getfourNodeinfo_recv, GetfourNodeinfo_recv)

	// engine.RegisterMsg(MSGID_getNodeWalletReceiptAddress, GetWalletAddr)
	message_center.Register_p2p(config.MSGID_sharebox_getNodeWalletReceiptAddress, GetWalletAddr)
	// engine.RegisterMsg(MSGID_getNodeWalletReceiptAddress_recv, GetWalletAddr_recv)
	message_center.Register_p2p(config.MSGID_sharebox_getNodeWalletReceiptAddress_recv, GetWalletAddr_recv)

	message_center.Register_p2p(config.MSGID_sharebox_getsharefolderlist, GetShareFolderList)
	message_center.Register_p2p(config.MSGID_sharebox_getsharefolderlist_recv, GetShareFolderList_recv)

}

/*
	收到共享文件消息
*/
func AddFileShare(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	fmt.Println("收到共享文件消息")

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	fi, err := ParseFileindex(*message.Body.Content)
	if err != nil {
		// fmt.Println(err)
	}

	fmt.Println("本节点保存文件索引", string(fi.JSON()))

	//判断本地网络是否存在文件，若不存在则添加
	filocal := FindFileindexToNet(fi.Hash.B58String())
	if filocal == nil {
		//添加文件
		err = AddFileindexToNet(fi, true)
		if err != nil {
			// fmt.Println(err)
		}
		fmt.Println("文件索引保存到本地")
	} else {
		fmt.Println("本地有文件索引")
		//文件中添加共享用户
		fi.AddShareUser(message.Head.Sender)
		// for _, v := range fi.FileChunk.GetAll() {
		// 	one := v.(*FileChunk)
		// 	filocal.AddShareUser(one.No, message.Head.Sender)
		// }
	}
	//回复给发送者
	message_center.SendP2pReplyMsg(message, config.MSGID_sharebox_addFileShare_recv, nil)
	// mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	// mbody := mc.NewMessageBody(nil, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// message = mc.NewMessage(mhead, mbody)
	// message.Reply(MSGID_addFileShare_recv)
}

/*
	收到共享文件消息 返回
*/
func AddFileShare_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("收到共享文件消息 返回")

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	//	fmt.Println("===", string(msg.Data), "\n", string(*message.Body.JSON()))

	//	message := new(mc.Message)
	//	err := json.Unmarshal(msg.Data, message)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	//	form, _ := hex.DecodeString(msg.Session.GetName())

	//	mh, _ := utils.FromB58String(msg.Session.GetName())

	//	if ok := mc.IsSendToOtherSuper(message, msg.MsgID, &mh); ok {
	//		fmt.Println("发给其他小伙伴了")
	//		return
	//	}
	//	fmt.Println("是本节点的")
	flood.ResponseWait(message_center.CLASS_sharefile, hex.EncodeToString(message.Body.Hash), &[]byte{})
}

/*
	收到查询文件信息消息
*/
func FindFileinfoHandler(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("收到查询文件信息消息")

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println("解析查询索引文件消息错误：", err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println("解析消息内容错误", err)
	// 	return
	// }
	var hashid nodeStore.AddressNet
	if message.Body.Content != nil {
		// umul := utils.Multihash(*message.Body.Content)
		hashid = nodeStore.AddressNet(*message.Body.Content)
	} else {
		hashid = *message.Head.RecvId
	}
	var bs []byte
	//fileinfo := FindFileinfoToNet(message.Head.RecvId.B58String())
	fileinfo := FindFileindexToNet(hashid.B58String())
	if fileinfo != nil {
		bs = fileinfo.JSON()
		// fmt.Println("查询到了文件", string(bs))
	} else {
		// fmt.Println("没有找到文件索引")
	}
	// fmt.Println("发回：", message.Head.Sender.B58String(), message.Head.SenderSuperId.B58String())
	message_center.SendP2pReplyMsg(message, config.MSGID_sharebox_findFileinfo_recv, &bs)
	// mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, false)
	// mbody := mc.NewMessageBody(&bs, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// message = mc.NewMessage(mhead, mbody)
	// message.Reply(MSGID_findFileinfo_recv)

}

/*
	收到查询文件索引 返回
*/
func FindFileinfo_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// fmt.Println("收到查询文件索引 返回")

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	//				fmt.Println("---2", err)
	// 	//		fmt.Println(string(msg.Dataplus))
	// 	//		engine.Log.Debug("%s", err.Error())
	// 	//		engine.Log.Debug("%s", string(msg.Dataplus))
	// 	//		engine.NLog.Error(engine.LOG_file, "%s", err.Error())
	// 	//		engine.NLog.Error(engine.LOG_file, "%s", string(msg.Dataplus))
	// 	return
	// }

	//	message := new(mc.Message)
	//	err := json.Unmarshal(msg.Data, message)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	//	form, _ := hex.DecodeString(msg.Session.GetName())

	//	mh, _ := utils.FromB58String(msg.Session.GetName())

	//	if ok := mc.IsSendToOtherSuper(message, msg.MsgID, &mh); ok {
	//		fmt.Println("发给其他小伙伴了")
	//		return
	//	}

	flood.ResponseWait(message_center.CLASS_findfileinfo, hex.EncodeToString(message.Body.Hash), message.Body.Content)

}

/*
	收到查询文件长度
*/
func FindFilesize(c engine.Controller, msg engine.Packet, message *message_center.Message) {

}

/*
	收到查询文件长度 返回
*/
func FindFilesize_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

}

type FileChunkVO struct {
	FileHash *nodeStore.AddressNet //完整文件hash
	// No            uint64           //文件块编号
	// ChunkHash     *utils.Multihash //块 hash
	Index   uint64 //下载块起始位置
	Length  uint64 //下载块长度
	Content []byte //数据块内容
	// ContentLength uint64           //数据块总大小
}

func (this *FileChunkVO) JSON() []byte {
	bs, _ := json.Marshal(this)
	return bs
}
func ParseFileChunkVO(bs []byte) *FileChunkVO {
	fcvo := new(FileChunkVO)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(fcvo)
	// if json.Unmarshal(bs, fcvo) != nil {
	if err != nil {
		return nil
	}
	return fcvo
}

/*
	收到下载文件块
*/
func DownloadFilechunk(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//fmt.Println("收到下载文件块")
	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	filechunk := ParseFileChunkVO(*message.Body.Content)

	var resultErrorMsgFun = func() {
		//给发送者返回错误消息
		if message_center.SendP2pReplyMsg(message, config.MSGID_sharebox_downloadFileChunk_recv, nil) {
			return
		}
		// mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
		// mbody := mc.NewMessageBody(nil, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
		// message = mc.NewMessage(mhead, mbody)
		// if message.Reply(MSGID_downloadFileChunk_recv) {
		// 	return
		// }

	}

	absPath := ""
	fileinfo := FindFile(filechunk.FileHash.B58String())
	if fileinfo != nil {
		absPath = fileinfo.Path
	} else {
		absPath = filepath.Join(sconfig.Store_dir, filechunk.FileHash.B58String())
	}

	//bs, err := ioutil.ReadFile(filepath.Join(gconfig.Store_dir, filechunk.ChunkHash.B58String()))
	f, err := os.Open(absPath)
	defer f.Close()
	if err != nil {
		// fmt.Println(err)
		resultErrorMsgFun()
		return
	}
	fi, err := f.Stat()
	if err != nil {
		// fmt.Println(err)
		resultErrorMsgFun()
		return
	}
	//请求下载断点错误
	datalength := uint64(fi.Size())
	if filechunk.Index > datalength {
		resultErrorMsgFun()
		return
	}
	//请求下载长度错误
	if filechunk.Index+filechunk.Length > datalength {
		resultErrorMsgFun()
		return
	}

	//判断下载长度
	var length uint64
	if filechunk.Length > 1024*1024 {
		length = 1024 * 1024 //1M
	} else {
		length = filechunk.Length
	}
	// if filechunk.Length == uint64(0) || filechunk.Length > datalength {
	// 	length = 1024 * 1024 //1M
	// } else {
	// 	length = filechunk.Length
	// }
	bs := make([]byte, length)
	_, err = f.ReadAt(bs, int64(filechunk.Index))
	if err != nil {
		f.Close()
		// fmt.Println(err)
		resultErrorMsgFun()
		return
	}
	f.Close()
	filechunk.Content = bs
	fmt.Println("**********收到块下载信息********")
	fmt.Println("块", filechunk.FileHash.B58String())
	fmt.Println("-------- 从这里下载的文件块 -------")
	fmt.Println(filechunk.Index, length)
	fmt.Println("发送给", message.Head.Sender.B58String())
	fmt.Println("预计发送大小", len(bs))
	fmt.Println("*********end***********")
	content := filechunk.JSON()
	message_center.SendP2pReplyMsg(message, config.MSGID_sharebox_downloadFileChunk_recv, &content)
	// mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	// content := filechunk.JSON()
	// mbody := mc.NewMessageBody(&content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// message = mc.NewMessage(mhead, mbody)
	// message.Reply(MSGID_downloadFileChunk_recv)
}

/*
	收到下载文件块 返回
*/
func DownloadFilechunk_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//	fmt.Println("收到下载文件块 返回", string(msg.Data))
	// fmt.Println("收到下载文件块 返回", len(msg.Dataplus))

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println("error  1", err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	//		fmt.Println("error  2", err)
	// 	//		fmt.Println(string(msg.Dataplus) + "end")
	// 	engine.NLog.Error(engine.LOG_file, "%s", err.Error())
	// 	engine.NLog.Error(engine.LOG_file, "%s", string(msg.Dataplus))
	// 	return
	// }

	//	message := new(mc.Message)
	//	err := json.Unmarshal(msg.Data, message)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//	//	fmt.Println("111查看是否被改变", string(*message.JSON()))
	//	//	form, _ := hex.DecodeString(msg.Session.GetName())

	//	form, _ := utils.FromB58String(msg.Session.GetName())

	//	if ok := mc.IsSendToOtherSuper(message, msg.MsgID, &form); ok {
	//		fmt.Println("发给其他小伙伴了")
	//		return
	//	}
	//	fmt.Println("222查看是否被改变", string(*message.JSON()))

	// fmt.Println("返回的文件块内容大小", len(*message.Body.Content))

	flood.ResponseWait(message_center.CLASS_downloadfile, hex.EncodeToString(message.Body.Hash), message.Body.Content)

}

//上传地址信息
type UpInfo struct {
	Scheme string
	Ip     string
	Port   uint16
	Path   string
	Field  string
}

func (u *UpInfo) Json() []byte {
	res, err := json.Marshal(u)
	if err != nil {
		// fmt.Println("upinfo marshal:", err)
		return nil
	}

	return res
}

//获取上传地址信息
func Uploadinfo(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	upinfo := UpInfo{}
	upinfo.Scheme = sconfig.UploadScheme
	upinfo.Ip = nodeStore.NodeSelf.Addr
	upinfo.Port = config.WebPort
	upinfo.Path = sconfig.UploadPath
	upinfo.Field = sconfig.UploadField
	content := upinfo.Json()
	//回复给发送者
	message_center.SendP2pReplyMsg(message, config.MSGID_sharebox_getUploadinfo_recv, &content)
	// mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	// mbody := mc.NewMessageBody(&content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// message = mc.NewMessage(mhead, mbody)
	// message.Reply(MSGID_getUploadinfo_recv)
}

//获取上传地址信息 返回
func Uploadinfo_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println("error  1", err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	engine.NLog.Error(engine.LOG_file, "%s", err.Error())
	// 	engine.NLog.Error(engine.LOG_file, "%s", string(msg.Dataplus))
	// 	return
	// }
	flood.ResponseWait(message_center.CLASS_uploadinfo, hex.EncodeToString(message.Body.Hash), message.Body.Content)
}

// //根据文件hash获取1/4节点地址信息（app用）
// func GetfourNodeinfo(c engine.Controller, msg engine.Packet) {
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
// 	var idstr []string
// 	ids := getQuarterLogicIds(message.Head.RecvId)
// 	for _, v := range ids {
// 		idstr = append(idstr, v.B58String())
// 	}
// 	content, err := json.Marshal(idstr)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	//回复给发送者
// 	mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
// 	mbody := mc.NewMessageBody(&content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
// 	message = mc.NewMessage(mhead, mbody)
// 	message.Reply(MSGID_getfourNodeinfo_recv)
// }

// //根据文件hash获取1/4节点地址信息（app用）
// func GetfourNodeinfo_recv(c engine.Controller, msg engine.Packet) {
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
// 	mc.ResponseWait(mc.CLASS_getfourNodeinfo, message.Body.Hash.B58String(), message.Body.Content)
// }

/*
	获取节点共享文件夹中的文件列表
*/
func GetShareFolderList(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	divVO := GetShareFolderRootsDetail()
	//获取云盘文件
	dir := getListFileFromSelf()
	divVO.Dirs = append(divVO.Dirs, dir)

	bs, err := json.Marshal(divVO)
	if err != nil {
		return
	}

	//回复给发送者
	message_center.SendP2pReplyMsg(message, config.MSGID_sharebox_getsharefolderlist_recv, &bs)

	// mhead := message_center.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	// mbody := message_center.NewMessageBody(&bs, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// message = message_center.NewMessage(mhead, mbody)
	// message.Reply(MSGID_getsharefolderlist_recv)
}

/*
	获取节点共享文件夹中的文件列表 返回
*/
func GetShareFolderList_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {

	flood.ResponseWait(message_center.CLASS_getRemoteFolderList, hex.EncodeToString(message.Body.Hash), message.Body.Content)

}
