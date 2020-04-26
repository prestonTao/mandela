package store

import (
	"mandela/config"
	"mandela/core/engine"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/virtual_node"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"sync"
	"time"
)

var (
	downchunkthread = 5 //最多一次5个协程同时下载
)

// /*
// 	共享本节点的所有文件块索引
// 	把文件索引上传到网络中去，并且增加本节点共享
// */
// func UpNetFileindex(fi *FileIndex, vid *virtual_node.AddressNetExtend) error {
// 	//	for _, v := range fi.FileChunk.GetAll() {
// 	//		one := v.(*FileChunk)

// 	//		fi.AddShareUser(one.No, &nodeStore.NodeSelf.IdInfo.Id)
// 	//	}

// 	//	recvId := fi.Hash
// 	//fmt.Println(recvId)

// 	// virtual_node.SendSearchAllMsg()

// 	// mc.SendVnodeSearchMsg(config.MSGID_store_addFileShare, recvId, &content)

// 	fmt.Println("hash", fi.Hash.B58String())

// 	logicIds := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(fi.Hash)
// 	content := fi.JSON()

// 	for _, one := range logicIds {

// 		fmt.Println("one ", vid.B58String(), one.B58String())

// 		if message, ok := mc.SendVnodeSearchMsg(config.MSGID_store_addFileShare, vid, one, &content); ok {
// 			fmt.Println("发给其他小伙伴了----")
// 			bs := flood.WaitRequest(mc.CLASS_sharefile, hex.EncodeToString(message.Body.Hash))
// 			fmt.Println("有消息返回了啊")
// 			if bs == nil {
// 				fmt.Println("发送共享文件消息失败，可能超时")
// 				return errors.New("发送共享文件消息失败，可能超时")
// 			}
// 			return nil
// 		}
// 	}
// 	return nil
// }

/*
	网络中查找一个文件信息
*/
func FindFileindexOpt(hash string) (fi *FileIndex, err error) {
	err = ErrNotFindCode
	// mh, errs := utils.FromB58String(hash)
	// if errs != nil {
	// 	err = errs
	// 	return
	// }
	vnodes := virtual_node.GetVnodeSelf()
	if len(vnodes) <= 0 {
		return nil, errors.New("没有加入存储网络")
	}

	mh := virtual_node.AddressFromB58String(hash)
	ids := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(&mh)
	for _, id := range ids {
		content := []byte(mh)
		if message, ok := mc.SendVnodeSearchMsg(config.MSGID_store_findFileinfo, &vnodes[0].Vid, id, &content); ok {
			//		fmt.Println("开始等待查找返回")
			bs := flood.WaitRequest(mc.CLASS_findfileinfo, hex.EncodeToString(message.Body.Hash), 0)
			//		fmt.Println("等待查找已经返回", string(*bs))
			if bs != nil {
				fi, err = ParseFileindex(*bs)
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
func DownloadFileOpt(fileindex *FileIndex, isdown bool) error {
	var dp *DownProc
	//下载时才加入下载列表
	if isdown {
		dp = NewDownProc(fileindex) //添加到下载列表
	} else {
		dp = nil
	}
	file := NewFile(fileindex)

	group := new(sync.WaitGroup)
	//分块下载
	for k, one := range fileindex.FileChunk {
		group.Add(1)
		go func(one *FileChunk, dp *DownProc) {
			renum := 0
		begin:
			if err := DownloadFilechunkToLocal(fileindex, one.No, dp); err == nil {
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

////同步文件索引到邻居节点
//func SyncFiletoNearId(fi *FileIndex) error {
//	recvId := nodeStore.FindNearInSuper(&nodeStore.NodeSelf.IdInfo.Id, nil, false)
//	if recvId == nil {
//		return errors.New("没有附近节点")
//	}
//	//加入自己为共享用户
//	for _, v := range fi.FileChunk.GetAll() {
//		one := v.(*FileChunk)
//		fi.AddShareUser(one.No, recvId)
//	}
//	content := fi.JSON()
//	if message, ok, _ := mc.SendP2pMsg(config.MSGID_store_addFileShare, recvId, &content); ok {
//		bs := flood.WaitRequest(mc.CLASS_sharefile, message.Body.Hash.B58String())
//		if bs == nil {
//			// fmt.Println("发送共享文件消息失败，可能超时")
//			return errors.New("发送共享文件消息失败，可能超时")
//		}
//		return nil
//	}
//	return nil
//}

////检查是否还有剩余空间
//func CheckSpace(size uint64) bool {
//	totalSize := fs.GetSpaceSize()

//	if totalSize >= fs.GetUseSpaceSize()+size {
//		return true
//	}
//	return false

//	//	sp := getSpaceSize()
//	//	space := sp + size
//	//	if space < uint64(config.Spacenum) {
//	//		return true
//	//	}
//	//	return false
//}

/*
	一个节点的所有文件列表
*/
type NodeFileindexList []VnodeFileindexList

/*
	一个虚拟节点的文件列表
*/
type VnodeFileindexList struct {
	virtual_node.Vnodeinfo
	List []FileIndex
}

/*
	解析一个节点的所有文件列表
*/
func ParseNodeFileindexList(bs *[]byte) NodeFileindexList {
	var nfl NodeFileindexList = make([]VnodeFileindexList, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(&nfl)
	if err != nil {
		return nil
	}
	return nfl
}

/*
	检查文件拥有者是否保存了文件，同时验证空间大小是否能放下文件
	@return    bool    false=验证不通过；true=验证通过；
*/
func CheckOwerFileindex(nid *nodeStore.AddressNet, vid *virtual_node.AddressNetExtend, fi FileIndex) bool {
	nfl := GetFileindexListForNode(nid)
	if nfl == nil {
		return false
	}
	for _, one := range nfl {
		if !bytes.Equal(one.Vid, *vid) {
			continue
		}
		totalSize := uint64(0)
		have := false
		for _, two := range one.List {
			totalSize = totalSize + two.Size
			if bytes.Equal(*two.Hash, *fi.Hash) {
				have = true
			}
		}
		//判断保存的文件占用空间大小是否超出限制
		if totalSize > config.Spacenum {
			return false
		}
		return have
	}
	return false
}

/*
	查询一个网络节点的文件列表
*/
func GetFileindexListForNode(nid *nodeStore.AddressNet) NodeFileindexList {
	content := []byte("ok")
	if message, ok, _ := mc.SendP2pMsg(config.MSGID_store_getFileindexList, nid, &content); ok {
		bs := flood.WaitRequest(mc.CLASS_getfileindexlist, hex.EncodeToString(message.Body.Hash), 0)
		if bs == nil {
			//fmt.Println("发送共享文件消息失败，可能超时1")
			// err = errors.New("发送共享文件消息失败，可能超时")
			return nil
		}

		return ParseNodeFileindexList(bs)

		// spacelist := ParseSpaceList(*bs)
		// var size uint64
		// for _, val := range spacelist.List {
		// 	size = size + val.Size
		// }
		// if size < config.Spacenum {
		// 	bl = true
		// 	err = nil
		// } else {
		// 	bl = false
		// 	err = errors.New("存储空间不足")
		// }

		// return
	}
	// err = errors.New("发送共享文件消息失败，可能超时")
	return nil
}
