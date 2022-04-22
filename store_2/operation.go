package store

import (
	"mandela/config"
	"mandela/core/engine"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	downchunkthread = 5 //最多一次5个协程同时下载
)

/*
	共享本节点的所有文件块索引
	把文件索引上传到网络中去，并且增加本节点共享
*/
func UpNetFileinfo(fi *FileInfo) error {
	for _, v := range fi.FileChunk.GetAll() {
		one := v.(*FileChunk)
		fi.AddShareUser(one.No, &nodeStore.NodeSelf.IdInfo.Id)
	}
	//这个文件索引归自己管理
	AddFileinfoToNet(fi, true)
	recvId := fi.Hash
	content := fi.JSON()
	//fmt.Println(recvId)

	// virtual_node.SendSearchAllMsg()

	// mc.SendVnodeSearchMsg(config.MSGID_store_addFileShare, recvId, &content)

	if message, ok := mc.SendSearchAllMsg(config.MSGID_store_addFileShare, recvId, &content); ok {
		fmt.Println("发给其他小伙伴了----")
		bs := flood.WaitRequest(mc.CLASS_sharefile, message.Body.Hash.B58String())
		//		fmt.Println("有消息返回了啊")
		if bs == nil {
			// fmt.Println("发送共享文件消息失败，可能超时")
			return errors.New("发送共享文件消息失败，可能超时")
		}
		return nil
	}
	return nil
}

/*
	网络中查找一个文件信息
*/
func FindFileinfoOpt(hash string) (fi *FileInfo, err error) {
	err = ErrNotFindCode
	// mh, errs := utils.FromB58String(hash)
	// if errs != nil {
	// 	err = errs
	// 	return
	// }
	mh := nodeStore.AddressFromB58String(hash)
	ids := getQuarterLogicIds(&mh)
	for _, id := range ids {
		content := []byte(mh)
		if message, ok := mc.SendSearchAllMsg(config.MSGID_store_findFileinfo, id, &content); ok {
			//		fmt.Println("开始等待查找返回")
			bs := flood.WaitRequest(mc.CLASS_findfileinfo, message.Body.Hash.B58String())
			//		fmt.Println("等待查找已经返回", string(*bs))
			if bs != nil {
				fi, err = ParseFileinfo(*bs)
				return
			}

		}
	}
	//	fmt.Println("这里直接就返回错误了")
	return

}

/*
	网络中下载一个文件到本地
*/
func DownloadFileOpt(fileinfo *FileInfo, isdown bool) error {
	var dp *DownProc
	//下载时才加入下载列表
	if isdown {
		dp = NewDownProc(fileinfo) //添加到下载列表
	} else {
		dp = nil
	}
	file := NewFile(fileinfo)

	group := new(sync.WaitGroup)
	//分块下载
	for k, v := range fileinfo.FileChunk.GetAll() {
		one := v.(*FileChunk)
		group.Add(1)
		go func(one *FileChunk, dp *DownProc) {
			renum := 0
		begin:
			if err := DownloadFilechunkToLocal(fileinfo, one.No, dp); err == nil {
				//				fmt.Println("下载文件成功了")
				//文件块下载成功
				fc := NewFileChunk(one.No, one.Hash)
				file.AddFileChunk(fc)

			} else {
				engine.Log.Warn("下载文件块错误 %s", err.Error())
				//继续偿试下载
				time.Sleep(time.Second)
				renum++
				if renum < Renum {
					goto begin
				}
			}
			group.Done()
		}(one, dp)
		//最多5个协程同时下载
		if k%downchunkthread == 0 {
			group.Wait()
		}
	}
	group.Wait()
	if !file.Check() {
		//		fmt.Println("文件分片下载失败")
		return errors.New("文件分片下载失败")
	}
	err := file.Assemble()
	if err != nil {
		// fmt.Println("组装文件失败", err)
		return err
	}
	return nil
}

//同步文件索引到邻居节点
func SyncFiletoNearId(fi *FileInfo) error {
	recvId := nodeStore.FindNearInSuper(&nodeStore.NodeSelf.IdInfo.Id, nil, false)
	if recvId == nil {
		return errors.New("没有附近节点")
	}
	//加入自己为共享用户
	for _, v := range fi.FileChunk.GetAll() {
		one := v.(*FileChunk)
		fi.AddShareUser(one.No, recvId)
	}
	content := fi.JSON()
	if message, ok, _ := mc.SendP2pMsg(config.MSGID_store_addFileShare, recvId, &content); ok {
		bs := flood.WaitRequest(mc.CLASS_sharefile, message.Body.Hash.B58String())
		if bs == nil {
			// fmt.Println("发送共享文件消息失败，可能超时")
			return errors.New("发送共享文件消息失败，可能超时")
		}
		return nil
	}
	return nil
}

//检查是否还有剩余空间
func CheckSpace(size uint64) bool {
	sp := getSpaceSize()
	space := sp + size
	if space < uint64(spacenum) {
		return true
	}
	return false
}

//发送空间大小验证消息
func sendSpaceinfo(recvId *nodeStore.AddressNet) (bl bool, err error) {
	content := []byte("ok")
	if message, ok, _ := mc.SendP2pMsg(config.MSGID_store_checkspaceinfo, recvId, &content); ok {
		bs := flood.WaitRequest(mc.CLASS_checkspaceinfo, message.Body.Hash.B58String())
		if bs == nil {
			//fmt.Println("发送共享文件消息失败，可能超时1")
			err = errors.New("发送共享文件消息失败，可能超时")
			return
		}
		spacelist := ParseSpaceList(*bs)
		var size uint64
		for _, val := range spacelist.List {
			size = size + val.Size
		}
		if size < spacenum {
			bl = true
			err = nil
		} else {
			bl = false
			err = errors.New("存储空间不足")
		}

		return
	}
	err = errors.New("发送共享文件消息失败，可能超时")
	return
}
