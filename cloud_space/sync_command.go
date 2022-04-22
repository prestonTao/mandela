package cloud_space

import (

	//	"mandela/core/utils"
	//	gconfig "mandela/config"
	// "mandela/store/fs"
	"mandela/config"
	// mc "mandela/core/message_center"
	// "mandela/core/message_center/flood"
	// "mandela/core/virtual_node"
	"bytes"
	// "encoding/hex"
	// "errors"
	"fmt"
	"os"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	FileInfoType  = 1 //索引文件
	FileChunkType = 2 //块文件
)

type SyncData struct {
	Type   uint8  //类型 1，索引 2，文件块
	Data   []byte //数据内容
	Timing bool   //是否是定时器发送的数据
}

func (sd *SyncData) Json() []byte {
	res, err := json.Marshal(sd)
	if err != nil {
		// fmt.Println(err)
		return nil
	}
	return res
}

//解析字节为SyncData
func ParseSyncData(data []byte) (*SyncData, error) {
	sd := new(SyncData)
	// err := json.Unmarshal(data, sd)
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err := decoder.Decode(sd)
	if err != nil {
		// fmt.Println("RaftTeam:", err)
		return sd, err
	}
	return sd, nil
}

//获取1/4节点id
//func getQuarterLogicIds(id *nodeStore.AddressNet) []*nodeStore.AddressNet {
//	return nodeStore.GetQuarterLogicAddrNetByAddrNet(id)
//}

/*
	发送文件拥有者消息，给4个节点
*/
func syncFileInfoToPeer(fileHashBs []byte) error {
	// fmt.Println("syncFileInfoToPeer 参数", fileHashBs)
	// filehash := virtual_node.AddressNetExtend(fileHashBs)
	// fmt.Println("syncFileInfoToPeer 参数", filehash.B58String())
	// // fiTable := fs.FileindexSelf{}
	// fiLocal, vid, err := FindFileindexToSelf(filehash)
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("syncFileInfoToPeer 查询结果", fiLocal)

	// ids := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(&filehash)
	// fmt.Println("####同步给#####")
	// for _, id := range ids {
	// 	content := fiLocal.JSON()
	// 	//f, _ := ParseChunkInfoData(data.Data)
	// 	fmt.Printf("发送======：%s", string(content))
	// 	if message, ok := mc.SendVnodeSearchMsg(config.MSGID_store_addFileOwner, &vid, id, &content); ok {
	// 		bs := flood.WaitRequest(mc.CLASS_syncfileinfo, hex.EncodeToString(message.Body.Hash), 0)
	// 		if bs == nil {
	// 			// fmt.Println("发送共享文件消息失败，可能超时")
	// 			return errors.New("发送共享文件消息失败，可能超时")
	// 		}
	// 		return nil
	// 	}

	// 	//fmt.Println(k, id.B58String())
	// 	// sd := new(SyncData)
	// 	// sd.Type = FileInfoType
	// 	// //索引数据
	// 	// fid := new(FileInfoData)
	// 	// fid.FHash = cid.FHash
	// 	// fid.CHash = cid.CHash
	// 	// fid.FInfo = cid.FInfo
	// 	// fid.Time = time.Now()

	// 	// sd.Data = fid.Json()
	// 	// sendData(id, sd)
	// }
	// //fmt.Println("#########")
	return nil
}

////同步所有块索引到相应节点
//func SyncFileChunkToPeer(fi *FileIndex) error {
//	//把上传者加入默认共享者
//	for _, v := range fi.FileChunk.GetAll() {
//		one := v.(*FileChunk)
//		fi.AddShareUser(one.No, &nodeStore.NodeSelf.IdInfo.Id)
//	}
//	//把上传者设为所有者
//	fi.AddFileUser(nodeStore.NodeSelf.IdInfo.Id.B58String())
//	chunks := fi.FileChunk.GetAll()
//	for _, v := range chunks {
//		fc := v.(*FileChunk)
//		cid := new(ChunkInfoData)
//		cid.FHash = fi.Hash
//		cid.CHash = fc.Hash
//		cid.FInfo = fi.JSON()
//		cid.AddShareUser(&nodeStore.NodeSelf.IdInfo.Id)
//		cid.First = true //第一次同步
//		//本节点(上传者)参与管理块索引
//		FD.AddFileChunk(cid)
//		//同步块索引到相应节点
//		sendChunkInfoToPeer(cid)
//	}
//	return nil
//}

////发送块索引到相应节点
//func sendChunkInfoToPeer(cid *ChunkInfoData) {
//	ids := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(cid.CHash)
//	var idtmp nodeStore.AddressNet // := new(utils.Multihash)
//	for _, id := range ids {
//		sd := new(SyncData)
//		sd.Type = FileChunkType
//		sd.Data = cid.Json()
//		//获取最近的节点，包括代理节点
//		receiveid := nodeStore.FindNearNodeId(id, nil, false)
//		if receiveid == nil {
//			receiveid = id
//		}
//		//防止重复发送
//		if idtmp != nil {
//			if bytes.Equal(idtmp, *receiveid) {
//				continue
//			}
//		}
//		idtmp = *receiveid
//		sendData(receiveid, sd)
//	}
//}

// //广播数据消息
// func sendData(id *virtual_node.AddressNetExtend, data *SyncData) error {
// 	//如果发给自己,则特殊处理
// 	if bytes.Equal(*id, nodeStore.NodeSelf.IdInfo.Id) {
// 		return nil
// 	}
// 	content := data.Json()
// 	//f, _ := ParseChunkInfoData(data.Data)
// 	//fmt.Printf("发送======：%s", f.FInfo)
// 	if message, ok := mc.SendVnodeSearchMsg(config.MSGID_store_syncFileInfo, id, &content); ok {
// 		bs := flood.WaitRequest(mc.CLASS_syncfileinfo, message.Body.Hash.B58String())
// 		if bs == nil {
// 			// fmt.Println("发送共享文件消息失败，可能超时")
// 			return errors.New("发送共享文件消息失败，可能超时")
// 		}
// 		return nil
// 	}
// 	return errors.New("数据发送失败")
// }

////同步块数据到相应节点
//func syncFileChunkData(cid *ChunkInfoData) {
//	chash := cid.CHash
//	//优先判断本地是否有文件块的缓存
//	ok, err := utils.PathExists(filepath.Join(gconfig.Store_dir, chash.B58String()))
//	if err != nil {
//		return
//	}
//	if ok {
//		//修改本块的共享者时间(主要处理上传者共享时间更新)
//		cid.AddShareUser(&nodeStore.NodeSelf.IdInfo.Id)
//		FD.AddFileChunk(cid)
//		//同步文件索引到相应节点，并且把本节点加入共享者
//		syncFileInfoToPeer(cid)
//		return
//	}

//	fileinfo := FD.GetFileIndex(chash)
//	if fileinfo == nil {
//		return
//	}
//	chunk := fileinfo.FindChunk(chash.B58String())
//	//如果没有找到块，则退出
//	if len(*chunk.Hash) == 0 {
//		return
//	}
//	if err := DownloadFilechunkToLocal(fileinfo, chunk.No); err == nil {
//		//文件块下载成功
//		// fmt.Println("块数据同步完成", chunk.Hash.B58String())
//		//同步文件索引到相应节点，并且把本节点加入共享者
//		syncFileInfoToPeer(cid)
//	} else {
//		// fmt.Println("下载文件块错误", err)
//	}
//}

//加入共享块用户为文件共享者
//func addFileindexToMyself(shareid *nodeStore.AddressNet, fid *FileInfoData) {
//	//fmt.Println("加入共享者", shareid.B58String())
//	fi, err := ParseFileindex(fid.FInfo)
//	if err != nil {
//		return
//	}
//	finew := new(FileIndex)

//	fin := fs.FileindexNet{}
//	fin, _ = fin.FindByFileid(fid.FHash.B58String())
//	fio, _ := ParseFileindex(finBs.Value)

//	//	fio, ok := netFileinfo.Load(fid.FHash.B58String())
//	if fio != nil {
//		finew = fio
//	} else {
//		finew = fi
//	}
//	//fmt.Printf("####收到的：%+v", fi)
//	one := finew.FindChunk(fid.CHash.B58String())
//	finew.AddShareUser(one.No, shareid) //增加文件块共享者
//	AddFileindexToNet(finew, true)
//	//fmt.Println("%%%%添加共享用户%%%%")
//	// fmt.Println(one.Hash.B58String(), shareid.B58String())
//}

//加入1/4节点为块的默认共享用户
//func addQuarterUser(fi *FileIndex, chunkhash *nodeStore.AddressNet) *FileIndex {
//	one := fi.FindChunk(chunkhash.B58String())
//	ids := getQuarterLogicIds(chunkhash)
//	for _, id := range ids {
//		fi.AddShareUser(one.No, id)
//	}
//	return fi
//}

//删除块文件
func DelChunkFile(hash string) error {
	file := config.Store_path_dir + "/" + hash
	err := os.Remove(file)
	fmt.Printf("删除块文件：%s %v", file, err)
	return err
}
