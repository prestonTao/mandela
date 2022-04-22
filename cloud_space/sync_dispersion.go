package cloud_space

import (
	// "mandela/cloud_space/fs"
	// "mandela/config"
	// mc "mandela/core/message_center"
	// "mandela/core/virtual_node"
	"io/ioutil"
	"os"
	"time"
)

const (
	FirstTimeInterval    = 30                     //第一次文件块同步检测，单位秒
	TimeInterval         = 10*Time_sharefile - 30 //文件块同步间隔，单位秒
	TimeIntervalEveryone = 600                    //文件所有者同步间隔，单位秒
)

var (
//	FD *FileData
)

// func init() {
// 	//	FD = NewFileData()
// 	start()
// }

//启动定时同步
func TimeSyncChunk() {

	//第一次同步块数据
	go func() {
		for range time.NewTicker(FirstTimeInterval * time.Second).C {
			//			timingFirst()
		}
	}()
	//定时同步块数据
	go func() {
		for range time.NewTicker(TimeInterval * time.Second).C {
			//			timing()
		}
	}()
	//定时同步文件所有者
	go func() {
		for range time.NewTicker(TimeIntervalEveryone * time.Second).C {
			sendOwerHeart()
		}
	}()
	//测试块共享节点
	//	go func() {
	//		for {
	//			<-time.NewTicker(5 * time.Second).C
	//			i := 0
	//			netFileinfo.Range(func(k, v interface{}) bool {
	//				i++
	//				finew := v.(*FileInfo)
	//				one := finew.FileChunk.GetAll()
	//				for k1, v1 := range one {
	//					chunk := v1.(*FileChunk)
	//					fmt.Println("*** 块:", k1, chunk.Hash.B58String(), "***")
	//					fmt.Println("*******当前共享者*******")
	//					for k2, v2 := range chunk.GetUserAll() {
	//						fmt.Println(k2, v2.Name.B58String(), v2.UpdateTime)
	//					}
	//					fmt.Println("******* end *******")
	//				}
	//				return true
	//			})
	//			if i == 0 {
	//				fmt.Println("当前不是文件共享节点")
	//			}
	//		}
	//	}()
	//测试块共享者
	//	go func() {
	//		for {
	//			<-time.NewTicker(5 * time.Second).C
	//			i := 0
	//			FD.FileChunkInfo.Range(func(k, v interface{}) bool {
	//				fmt.Println("*** start ****", k)
	//				i++
	//				cid := v.(*ChunkInfoData)
	//				fmt.Println("*******当前共享者*******")
	//				j := 0
	//				cid.Shares.Range(func(k1, v1 interface{}) bool {
	//					j++
	//					fmt.Println(j, k1.(string))
	//					return true
	//				})
	//				fmt.Println("******* end *******")
	//				return true
	//			})
	//			if i == 0 {
	//				fmt.Println("当前不是文件块共享节点")
	//			}
	//		}
	//	}()
}

//第一次同步块
//func timingFirst() {
//	FD.FileChunkInfo.Range(func(k, v interface{}) bool {
//		//fmt.Println("同步数据...", k.(string), v.(*ChunkInfoData).CHash.B58String())
//		chunkdata := v.(*ChunkInfoData)
//		//如果文件所有者这空，则删除块索引，块文件
//		finfo, err := ParseFileindex(chunkdata.FInfo)
//		if len(finfo.FileUser) == 0 && err == nil {
//			FD.FileChunkInfo.Delete(k)                //删除块索引
//			DelChunkFile(chunkdata.CHash.B58String()) // 删除块数据
//			return true
//		}
//		//同步本节点管理的块数据
//		if chunkdata.First {
//			//定时器同步，非第一次同步
//			FD.UpdateChunkFirst(chunkdata)
//			syncFileChunkData(chunkdata)
//			//同步块索引到相关节点,块节点收到同步数据时再同步文件索引到相应节点
//			sendChunkInfoToPeer(chunkdata)
//			time.Sleep(TimeIntervalEveryone * time.Millisecond)
//		}
//		return true
//	})
//}

//定时同步
//func timing() {
//	FD.FileChunkInfo.Range(func(k, v interface{}) bool {
//		//fmt.Println("同步数据...", k.(string), v.(*ChunkInfoData).CHash.B58String())
//		chunkdata := v.(*ChunkInfoData)
//		//如果文件所有者这空，则删除块索引，块文件
//		finfo, err := ParseFileindex(chunkdata.FInfo)
//		if len(finfo.FileUser) == 0 && err == nil {
//			FD.FileChunkInfo.Delete(k)                //删除块索引
//			DelChunkFile(chunkdata.CHash.B58String()) //删除块数据
//			return true
//		}
//		//同步本节点管理的块数据
//		syncFileChunkData(chunkdata)
//		//同步块索引到相关节点,块节点收到同步数据时再同步文件索引到相应节点
//		sendChunkInfoToPeer(chunkdata)
//		time.Sleep(TimeIntervalEveryone * time.Millisecond)
//		return true
//	})
//}

//发送所有者心跳
func sendOwerHeart() error {
	// fiTable := fs.FileindexSelf{}
	// fis, err := fiTable.Getall()
	// if err != nil {
	// 	return err
	// }
	// for _, one := range fis {
	// 	vid := virtual_node.AddressNetExtend(one.Vid)
	// 	vnodeinfo := virtual_node.FindInVnodeinfoSelf(vid)
	// 	if vnodeinfo == nil {
	// 		continue
	// 	}

	// 	fiOne, err := ParseFileindex(one.Value)
	// 	if err != nil {
	// 		continue
	// 	}

	// 	owner := FileOwner{
	// 		Vnodeinfo: *vnodeinfo,
	// 	}

	// 	fiOne.FileOwner[owner.Nid.B58String()] = owner

	// 	bs := fiOne.JSON()

	// 	ids := virtual_node.GetQuarterLogicAddrNetByAddrNetExtend(fiOne.Hash)
	// 	for _, idOne := range ids {
	// 		if _, ok := mc.SendVnodeSearchMsg(config.MSGID_store_addFileOwner, &vid, idOne, &bs); ok {
	// 			//				fmt.Println("发给其他小伙伴了----")
	// 			//				bs := flood.WaitRequest(mc.CLASS_sharefile, message.Body.Hash.B58String())
	// 			//				//		fmt.Println("有消息返回了啊")
	// 			//				if bs == nil {
	// 			//					// fmt.Println("发送共享文件消息失败，可能超时")
	// 			//					return errors.New("发送共享文件消息失败，可能超时")
	// 			//				}
	// 			//				return nil
	// 		}
	// 	}
	// }

	// //	myfolder := config.Store_fileinfo_self
	// //	fileInfos, _ := ioutil.ReadDir(myfolder)
	// //	for _, fileinfo := range fileInfos {
	// //		if !fileinfo.IsDir() {
	// //			fi := parseFile(filepath.Join(myfolder, fileinfo.Name()))
	// //			fmt.Printf("同步索引，更新所有者时间")
	// //			fi.AddFileUser(nodeStore.NodeSelf.IdInfo.Id.B58String())
	// //			UpNetFileindex(fi) //更新网络上的所有者时间
	// //		}
	// //	}
	return nil
}

//解析索引文件内容
func parseFile(name string) (fi *FileIndex) {
	f, err := os.Open(name)
	if err != nil {
		return
	}
	res, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}
	fi, err = ParseFileindex(res)
	if err != nil {
		return
	}
	return
}

////存储格式
//type FileData struct {
//	FileInfo      *sync.Map //索引 filehash:FileInfoData
//	FileChunkInfo *sync.Map //文件块 chunkhash:ChunkInfoData
//	Lock          *sync.Mutex
//}

//func NewFileData() *FileData {
//	return &FileData{FileInfo: new(sync.Map), FileChunkInfo: new(sync.Map), Lock: new(sync.Mutex)}
//}

////加入文件索引
//func (fd *FileData) AddFileInfo(fid *FileInfoData) {
//	fd.FileInfo.Store(fid.FHash.B58String(), fid)
//}

////修改第一次同步为定时同步
//func (fd *FileData) UpdateChunkFirst(cid *ChunkInfoData) {
//	cid.First = false
//	fd.UpdateFileChunk(cid)
//}

////加入块索引
//func (fd *FileData) AddFileChunk(cid *ChunkInfoData) {
//	fd.Lock.Lock()
//	defer fd.Lock.Unlock()
//	cids, ok := fd.FileChunkInfo.Load(cid.CHash.B58String())
//	if !ok {
//		fd.FileChunkInfo.Store(cid.CHash.B58String(), cid)
//	} else {
//		//如果已存在块索引，则加入共享者
//		chunkinfodata := cids.(*ChunkInfoData)
//		cid.Shares.Range(func(k, v interface{}) bool {
//			chunkinfodata.Shares.Store(k.(string), v.(*nodeStore.AddressNet))
//			return true
//		})
//		//合并原始fileinfo的所有者
//		chunkinfodata.FInfo = cid.MergeFileUser(chunkinfodata.FInfo, cid.FInfo)
//		//更新时间
//		chunkinfodata.Time = time.Now()
//		fd.FileChunkInfo.Store(chunkinfodata.CHash.B58String(), chunkinfodata)
//	}
//}

////修改块信息
//func (fd *FileData) UpdateFileChunk(cid *ChunkInfoData) {
//	fd.FileChunkInfo.Store(cid.CHash.B58String(), cid)
//}

////根据文件块获取fileinfo
//func (fd *FileData) GetFileIndex(chunkid *nodeStore.AddressNet) *FileIndex {
//	fds, ok := fd.FileChunkInfo.Load(chunkid.B58String())
//	if ok {
//		cid := fds.(*ChunkInfoData)
//		fileinfo, _ := ParseFileindex(cid.FInfo)
//		//加入共享者到fileinfo
//		one := fileinfo.FindChunk(chunkid.B58String())
//		cid.Shares.Range(func(k, v interface{}) bool {
//			fileinfo.AddShareUser(one.No, v.(*nodeStore.AddressNet))
//			return true
//		})
//		//加入默认点为共享节点
//		//fileinfo = addQuarterUser(fileinfo, chunkid)
//		return fileinfo
//	}
//	return nil
//}

////删除离线的块共享者
//func (fd *FileData) DelShareUser(id, chunkid *utils.Multihash) error {
//	fds, ok := fd.FileChunkInfo.Load(chunkid.B58String())
//	if ok {
//		cid := fds.(*ChunkInfoData)
//		cid.Shares.Range(func(k, v interface{}) bool {
//			if k.(string) == id.B58String() {
//				cid.Shares.Delete(id.B58String())
//				fd.UpdateFileChunk(cid)
//			}
//			return true
//		})
//		return nil
//	}
//	return nil
//}

////文件索引存储结构
//type FileInfoData struct {
//	FHash *virtual_node.AddressNetExtend //文件hash
//	CHash *virtual_node.AddressNetExtend //块hash
//	FInfo []byte                         //文件索引 fileinfo
//	Time  time.Time                      //共享者同步时间
//}

//func (fid *FileInfoData) Json() []byte {
//	res, err := json.Marshal(fid)
//	if err != nil {
//		// fmt.Println(err)
//		return nil
//	}
//	return res
//}

////解析字节为FileInfoData
//func ParseFileInfoData(data []byte) (*FileInfoData, error) {
//	fid := new(FileInfoData)
//	// err := json.Unmarshal(data, fid)
//	decoder := json.NewDecoder(bytes.NewBuffer(data))
//	decoder.UseNumber()
//	err := decoder.Decode(fid)
//	if err != nil {
//		// fmt.Println("ParseFileInfoData:", err)
//		return fid, err
//	}
//	return fid, nil
//}

////块存储结构
//type ChunkInfoData struct {
//	FHash  *virtual_node.AddressNetExtend //文件hash
//	CHash  *virtual_node.AddressNetExtend //块hash
//	Shares sync.Map                       //共享者hashstring:*crypto.Address
//	FInfo  []byte                         //文件索引 fileinfo 原始索引
//	Time   time.Time                      //共享者同步时间
//	First  bool                           //是否是第一次同步
//}

////增加块共享者
//func (cid *ChunkInfoData) AddShareUser(shareid *nodeStore.AddressNet) {
//	cid.Shares.Store(shareid.B58String(), shareid)
//}

////删除共享者
//func (cid *ChunkInfoData) DelShareUser(shareid *nodeStore.AddressNet) {
//	cid.Shares.Delete(shareid.B58String())
//}
//func (cid *ChunkInfoData) Json() []byte {
//	res, err := json.Marshal(cid)
//	if err != nil {
//		// fmt.Println(err)
//		return nil
//	}
//	return res
//}

////合并原始fileinfo的所有者
//func (cid *ChunkInfoData) MergeFileUser(fold, fnew []byte) []byte {
//	f1, err := ParseFileindex(fold)
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	f2, err := ParseFileindex(fnew)
//	if err != nil {
//		fmt.Println(err)
//		return nil
//	}
//	fn := MergeFileUser(f1, f2)
//	return fn.JSON()
//}

////解析字节为ChunkInfoData
//func ParseChunkInfoData(data []byte) (*ChunkInfoData, error) {
//	cid := new(ChunkInfoData)
//	// err := json.Unmarshal(data, cid)
//	decoder := json.NewDecoder(bytes.NewBuffer(data))
//	decoder.UseNumber()
//	err := decoder.Decode(cid)
//	if err != nil {
//		// fmt.Println("ParseFileInfoData:", err)
//		return cid, err
//	}
//	return cid, nil
//}
