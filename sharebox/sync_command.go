package sharebox

import (
	"mandela/core/nodeStore"
)

// const (
// 	FileInfoType  = 1 //索引文件
// 	FileChunkType = 2 //块文件
// )

// type SyncData struct {
// 	Type   uint8  //类型 1，索引 2，文件块
// 	Data   []byte //数据内容
// 	Timing bool   //是否是定时器发送的数据
// }

// func (sd *SyncData) Json() []byte {
// 	res, err := json.Marshal(sd)
// 	if err != nil {
// 		fmt.Println(err)
// 		return nil
// 	}
// 	return res
// }

// //解析字节为SyncData
// func ParseSyncData(data []byte) (*SyncData, error) {
// 	sd := new(SyncData)
// 	err := json.Unmarshal(data, sd)
// 	if err != nil {
// 		fmt.Println("RaftTeam:", err)
// 		return sd, err
// 	}
// 	return sd, nil
// }

//获取1/4节点id
func getQuarterLogicIds(id *nodeStore.AddressNet) []*nodeStore.AddressNet {
	return nodeStore.GetQuarterLogicAddrNetByAddrNet(id)
}

// //同步文件索引到相应节点
// func syncFileInfoToPeer(cid *ChunkInfoData) error {
// 	ids := getQuarterLogicIds(cid.FHash)
// 	//fmt.Println("####同步给#####")
// 	for _, id := range ids {
// 		//fmt.Println(k, id.B58String())
// 		sd := new(SyncData)
// 		sd.Type = FileInfoType
// 		//索引数据
// 		fid := new(FileInfoData)
// 		fid.FHash = cid.FHash
// 		fid.CHash = cid.CHash
// 		fid.FInfo = cid.FInfo
// 		fid.Time = time.Now()

// 		sd.Data = fid.Json()
// 		sendData(id, sd)
// 	}
// 	//fmt.Println("#########")
// 	return nil
// }

// //同步所有块索引到相应节点
// func SyncFileChunkToPeer(fi *FileInfo) error {
// 	//把上传者加入默认共享者
// 	fi.AddShareUser(nodeStore.NodeSelf.IdInfo.Id)

// 	// chunks := fi.FileChunk.GetAll()
// 	// for _, v := range chunks {
// 	// 	fc := v.(*FileChunk)
// 	// 	cid := new(ChunkInfoData)
// 	// 	cid.FHash = fi.Hash
// 	// 	cid.CHash = fc.Hash
// 	// 	cid.FInfo = fi.JSON()
// 	// 	cid.AddShareUser(nodeStore.NodeSelf.IdInfo.Id)
// 	// 	cid.First = true //第一次同步
// 	// 	//本节点(上传者)参与管理块索引
// 	// 	FD.AddFileChunk(cid)
// 	// 	//同步块索引到相应节点
// 	// 	sendChunkInfoToPeer(cid)
// 	// }
// 	return nil
// }

// // //发送块索引到相应节点
// // func sendChunkInfoToPeer(cid *ChunkInfoData) {
// // 	ids := getQuarterLogicIds(cid.CHash)
// // 	//	auto, err := utils.FromB58String("W1dcuWJQLqU2F9RgnFyP72c2FUUHdZNBgMSXk1JH62Qx3z")
// // 	//	if err != nil {
// // 	//		fmt.Println(err)
// // 	//	}
// // 	//	ids = append(ids, &auto)
// // 	idtmp := new(utils.Multihash)
// // 	for _, id := range ids {
// // 		sd := new(SyncData)
// // 		sd.Type = FileChunkType
// // 		sd.Data = cid.Json()
// // 		//获取最近的节点，包括代理节点
// // 		receiveid := nodeStore.FindNearNodeId(id, nil, false)
// // 		if receiveid == nil {
// // 			receiveid = id
// // 		}
// // 		//防止重复发送
// // 		if idtmp != nil {
// // 			if idtmp.B58String() == receiveid.B58String() {
// // 				continue
// // 			}
// // 		}
// // 		idtmp = receiveid
// // 		sendData(receiveid, sd)
// // 	}
// // }

// //广播数据消息
// func sendData(id *utils.Multihash, data *SyncData) error {
// 	//如果发给自己,则特殊处理
// 	if id.B58String() == nodeStore.NodeSelf.IdInfo.Id.B58String() {
// 		return nil
// 	}
// 	content := data.Json()
// 	mhead := mc.NewMessageHead(id, id, false)
// 	mbody := mc.NewMessageBody(&content, "", nil, 0)
// 	message := mc.NewMessage(mhead, mbody)
// 	if message.Send(MSGID_syncFileInfo) {
// 		//fmt.Println("数据发送成功", id.B58String())
// 		//		bs := mc.WaitRequest(mc.CLASS_syncfileinfo, message.Body.Hash.B58String())
// 		//		if bs == nil {
// 		//			//删除离线的块共享者
// 		//			if data.Type == FileChunkType {
// 		//				cid, err := ParseChunkInfoData(data.Data)
// 		//				if err == nil {
// 		//					FD.DelShareUser(id, cid.CHash)
// 		//				}
// 		//			}
// 		//			return errors.New("同步数据消息失败，可能超时")
// 		//		}
// 		//fmt.Println("有消息返回", string(*bs))
// 		return nil
// 	}
// 	//fmt.Println("数据发送失败", id.B58String())
// 	return errors.New("数据发送失败")
// }

// // //同步块数据到相应节点
// // func syncFileChunkData(cid *ChunkInfoData) {
// // 	chash := cid.CHash
// // 	//优先判断本地是否有文件块的缓存
// // 	ok, err := utils.PathExists(filepath.Join(gconfig.Store_dir, chash.B58String()))
// // 	if err != nil {
// // 		return
// // 	}
// // 	if ok {
// // 		//修改本块的共享者时间(主要处理上传者共享时间更新)
// // 		cid.AddShareUser(nodeStore.NodeSelf.IdInfo.Id)
// // 		FD.AddFileChunk(cid)
// // 		//同步文件索引到相应节点，并且把本节点加入共享者
// // 		syncFileInfoToPeer(cid)
// // 		return
// // 	}

// // 	fileinfo := FD.GetFileInfo(chash)
// // 	if fileinfo == nil {
// // 		return
// // 	}
// // 	chunk := fileinfo.FindChunk(chash.B58String())
// // 	//如果没有找到块，则退出
// // 	if len(*chunk.Hash) == 0 {
// // 		return
// // 	}
// // 	if err := DownloadFilechunkToLocal(fileinfo, chunk.No); err == nil {
// // 		//文件块下载成功
// // 		fmt.Println("块数据同步完成", chunk.Hash.B58String())
// // 		//同步文件索引到相应节点，并且把本节点加入共享者
// // 		syncFileInfoToPeer(cid)
// // 	} else {
// // 		fmt.Println("下载文件块错误", err)
// // 	}
// // }

// // //加入共享块用户为文件共享者
// // func addFileinfoToMyself(shareid *utils.Multihash, fid *FileInfoData) {
// // 	//fmt.Println("加入共享者", shareid.B58String())
// // 	fi, err := ParseFileinfo(fid.FInfo)
// // 	if err != nil {
// // 		return
// // 	}
// // 	finew := new(FileInfo)
// // 	fio, ok := netFileinfo.Load(fid.FHash.B58String())
// // 	if ok {
// // 		finew = fio.(*FileInfo)
// // 	} else {
// // 		finew = fi
// // 	}
// // 	one := finew.FindChunk(fid.CHash.B58String())
// // 	finew.AddShareUser(one.No, shareid)
// // 	AddFileinfoToNet(finew, true)
// // 	fmt.Println("%%%%添加共享用户%%%%")
// // 	fmt.Println(one.Hash.B58String(), shareid.B58String())
// // }

// // //加入1/4节点为块的默认共享用户
// // func addQuarterUser(fi *FileInfo, chunkhash *utils.Multihash) *FileInfo {
// // 	one := fi.FindChunk(chunkhash.B58String())
// // 	ids := getQuarterLogicIds(chunkhash)
// // 	for _, id := range ids {
// // 		fi.AddShareUser(one.No, id)
// // 	}
// // 	return fi
// // }

// // //如果发给自己,则特殊处理
// // func toMyself(id *utils.Multihash, data *SyncData) error {
// // 	if id.B58String() == nodeStore.NodeSelf.IdInfo.Id.B58String() {
// // 		sd := data
// // 		if sd.Type == FileChunkType {
// // 			//fmt.Println("收到块索引同步信息")
// // 			cid, err := ParseChunkInfoData(sd.Data)
// // 			if err != nil {
// // 				fmt.Println(err)
// // 				return err
// // 			}
// // 			cid.AddShareUser(nodeStore.NodeSelf.IdInfo.Id)
// // 			FD.AddFileChunk(cid)
// // 		}
// // 		if sd.Type == FileInfoType {
// // 			//fmt.Println("收到文件索引同步信息", message.Head.Sender.B58String())
// // 			fid, err := ParseFileInfoData(sd.Data)
// // 			if err != nil {
// // 				fmt.Println(err)
// // 				return err
// // 			}
// // 			FD.AddFileInfo(fid)
// // 			//当前文件归自己管理
// // 			addFileinfoToMyself(id, fid)
// // 		}
// // 		return errors.New("发送给自己")
// // 	}
// // 	return nil
// // }
