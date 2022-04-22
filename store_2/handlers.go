package store

import (
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	gconfig "mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
	收到共享文件消息
*/
func AddFileShare(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	fi, err := ParseFileinfo(*message.Body.Content)
	if err != nil {
		fmt.Println(err)
	}
	//	fmt.Println("本节点保存文件索引", string(fi.JSON()))

	//判断本地网络是否存在文件，若不存在则添加
	filocal := FindFileinfoToNet(fi.Hash.B58String())
	if filocal != nil {
		//fmt.Println("本地有文件索引")
		//文件中添加共享用户
		for _, v := range fi.FileChunk.GetAll() {
			one := v.(*FileChunk)
			filocal.AddShareUser(one.No, message.Head.Sender)
		}
	}
	//发送消息的节点为文件所有者
	go func() {
		//判断是否注册域名
		nameinfo := name.FindNameToNet(message.Head.Sender.B58String())
		//fmt.Printf("xxxxx%+v", nameinfo)
		if nameinfo == nil {
			fmt.Println("未注册域名")
			return
		}
		if nameinfo.Deposit < DepositMin {
			fmt.Println("域名冻结押金不足")
			return
		}
		if ok, _ := sendSpaceinfo(message.Head.Sender); ok { //判断空间是否不足
			//增加文件所有者(如存在，则更新时间)
			fi.AddFileUser(message.Head.Sender.B58String())
			AddFileinfoToNet(fi, true)
		}
	}()
	//回复给发送者
	bs := []byte("ok")
	message_center.SendP2pReplyMsg(message, config.MSGID_store_addFileShare_recv, &bs)
}

/*
	收到共享文件消息 返回
*/
func AddFileShare_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(message_center.CLASS_sharefile, message.Body.Hash.B58String(), message.Body.Content)
}

/*
	收到查询文件信息消息
*/
func FindFileinfoHandler(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	var hashid nodeStore.AddressNet
	if message.Body.Content != nil {
		// umul := utils.Multihash(*message.Body.Content)
		hashid = nodeStore.AddressNet(*message.Body.Content)
	} else {
		hashid = *message.Head.RecvId
	}
	var bs []byte
	//fileinfo := FindFileinfoToNet(message.Head.RecvId.B58String())
	fileinfo := FindFileinfoToNet(hashid.B58String())
	if fileinfo != nil {
		bs = fileinfo.JSON()
		fmt.Println("查询到了文件", string(bs))
	} else {
		fmt.Println("没有找到文件索引")
	}
	message_center.SendP2pReplyMsg(message, config.MSGID_store_findFileinfo_recv, &bs)
}

/*
	收到查询文件索引 返回
*/
func FindFileinfo_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(message_center.CLASS_findfileinfo, message.Body.Hash.B58String(), message.Body.Content)

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
	FileHash      *nodeStore.AddressNet //完整文件hash
	No            uint64                //文件块编号
	ChunkHash     *nodeStore.AddressNet //块 hash
	Index         uint64                //下载块起始位置
	Length        uint64                //下载块长度
	Content       []byte                //数据块内容
	ContentLength uint64                //数据块总大小
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

//下载
func DownloadFilechunk(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	filechunk := ParseFileChunkVO(*message.Body.Content)
	var resultErrorMsgFun = func() {
		//给发送者返回错误消息
		message_center.SendP2pReplyMsg(message, config.MSGID_store_downloadFileChunk_recv, nil)

	}
	//bs, err := ioutil.ReadFile(filepath.Join(gconfig.Store_dir, filechunk.ChunkHash.B58String()))
	f, err := os.Open(filepath.Join(gconfig.Store_dir, filechunk.ChunkHash.B58String()))
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		resultErrorMsgFun()
		return
	}
	fi, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		resultErrorMsgFun()
		return
	}
	//start
	//datalength := uint64(len(bs))
	datalength := uint64(fi.Size())
	if filechunk.Index > datalength {
		fmt.Println("err big")
		resultErrorMsgFun()
		return
	}
	var length uint64
	if filechunk.Index+filechunk.Length > datalength {
		length = datalength - filechunk.Index
	} else {
		length = filechunk.Length
	}
	index := filechunk.Index
	//bs = bs[filechunk.Index:length]
	bs := make([]byte, length)
	_, err = f.ReadAt(bs, int64(index))
	if err != nil {
		// fmt.Println(err)
		resultErrorMsgFun()
		return
	}
	f.Close()
	filechunk.Content = bs
	filechunk.ContentLength = datalength
	fmt.Println("**********收到块下载信息********")
	fmt.Println("块", filechunk.ChunkHash.B58String())
	fmt.Println("-------- 从这里下载的文件块 -------")
	fmt.Println(filechunk.Index, filechunk.Index+length)
	fmt.Println("发送给", message.Head.Sender.B58String())
	fmt.Println("预计发送大小", len(bs))
	fmt.Println("*********end***********")
	content := filechunk.JSON()
	message_center.SendP2pReplyMsg(message, config.MSGID_store_downloadFileChunk_recv, &content)
}
func DownloadFilechunk_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(message_center.CLASS_downloadfile, message.Body.Hash.B58String(), message.Body.Content)
}

////上传地址信息
//type UpInfo struct {
//	Scheme string
//	Ip     string
//	Port   uint16
//	Path   string
//	Field  string
//}

//func (u *UpInfo) Json() []byte {
//	res, err := json.Marshal(u)
//	if err != nil {
//		// fmt.Println("upinfo marshal:", err)
//		return nil
//	}

//	return res
//}

////获取上传地址信息
//func Uploadinfo(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	upinfo := UpInfo{}
//	upinfo.Scheme = UploadScheme
//	upinfo.Ip = nodeStore.NodeSelf.Addr
//	upinfo.Port = gconfig.WebPort
//	upinfo.Path = UploadPath
//	upinfo.Field = UploadField
//	content := upinfo.Json()
//	//回复给发送者
//	message_center.SendP2pReplyMsg(message, config.MSGID_store_getUploadinfo_recv, &content)
//}

////获取上传地址信息 返回
//func Uploadinfo_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
//	flood.ResponseWait(message_center.CLASS_uploadinfo, message.Body.Hash.B58String(), message.Body.Content)
//}

//根据文件hash获取1/4节点地址信息（app用）
func GetfourNodeinfo(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	var idstr []string
	ids := getQuarterLogicIds(message.Head.RecvId)
	for _, v := range ids {
		idstr = append(idstr, v.B58String())
	}
	content, err := json.Marshal(idstr)
	if err != nil {
		// fmt.Println(err)
		return
	}
	//回复给发送者
	message_center.SendP2pReplyMsg(message, config.MSGID_store_getfourNodeinfo_recv, &content)
}

//根据文件hash获取1/4节点地址信息（app用）
func GetfourNodeinfo_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(message_center.CLASS_getfourNodeinfo, message.Body.Hash.B58String(), message.Body.Content)
}

//验证空间大小
func CheckSpaceInfo(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	spacelist := SpaceList{}
	spacelist.GetSpaceList()
	content := spacelist.Json()
	//回复给发送者
	message_center.SendP2pReplyMsg(message, config.MSGID_store_checkspaceinfo_recv, &content)
}

//验证空间大小 返回
func CheckSpaceInfo_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	flood.ResponseWait(message_center.CLASS_checkspaceinfo, message.Body.Hash.B58String(), message.Body.Content)
}
