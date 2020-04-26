/*
	保存本地磁盘上已经存在的文件信息
	提供网络查找
*/
package store

import (
	"mandela/config"
	gconfig "mandela/config"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/utils"
	"mandela/core/virtual_node"
	"mandela/store/fs"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

//var localFileinfo = new(sync.Map)

/*
	保存文件索引到本地内存和磁盘
	@cover    bool    是否保存（覆盖）到本地磁盘
*/
func AddFileindexToLocal(fi *FileIndex, vid virtual_node.AddressNetExtend) error {
	//	localFileinfo.Store(fi.Hash.B58String(), fi)
	//添加定时任务，定时更新文件索引
	//task.Add(time.Now().Unix(), Task_class_share_local_fileinfo, fi.Hash.B58String())

	vidSelf := virtual_node.FindNearVnodeInSelf(&vid)

	fiTable := fs.FileindexLocal{}
	fiLocal, _ := FindFileindexToLocal(*fi.Hash)
	if fiLocal != nil {
		//修改
		fiTable.FileId = fi.Hash.B58String()
		fiTable.Vid = vidSelf.B58String()
		//		if bytes.Equal(vidSelf, vidLocal) {
		//		}
		fiTable.Value = fi.JSON()
		return fiTable.Update()
	} else {
		//不存在，则新增
		fiTable = fs.FileindexLocal{
			Vid:    vidSelf.B58String(), //虚拟节点id
			FileId: fi.Hash.B58String(), //索引哈市值
			Value:  fi.JSON(),           //内容
		}
		return fiTable.Add(&fiTable)
	}

	//	fiSelf := fs.FileindexLocal{}
	//	existFi, err := fiSelf.FindByFileid(*fi.Hash)
	//	if err != nil {
	//		return err
	//	}
	//	if existFi != nil {
	//		//已经存在，则修改
	//		err = fiSelf.UpdateValue(*fi.Hash, fi.JSON())
	//	} else {
	//		//不存在，则新增
	//		fiSelf = fs.FileindexLocal{
	//			Vid:    vid,       //虚拟节点id
	//			FileId: *fi.Hash,  //索引哈市值
	//			Value:  fi.JSON(), //内容
	//		}
	//		err = fiSelf.Add(&fiSelf)
	//	}
	//	return err
	//		return saveFileinfoToLocal(fi, vid)

}

func FindFileindexToLocal(name virtual_node.AddressNetExtend) (*FileIndex, error) {
	fiTable := fs.FileindexLocal{}
	fiDb, err := fiTable.FindByFileid(name.B58String())
	if err != nil {
		return nil, err
	}
	return ParseFileindex(fiDb.Value)
}

/*
	本地查找一个文件的块
*/
func FindFileChunk(filehash, chunkhash virtual_node.AddressNetExtend) bool {
	//先查找是否有这个文件
	fi, _ := FindFileindexToLocal(filehash)
	if fi == nil {
		return false
	}
	return fi.FindChunk(chunkhash)
}

///*
//	保存文件索引到本地磁盘
//*/
//func saveFileinfoToLocals(fi *FileIndex, vid string) error {
//	//先看是否存在，不存在则保存

//	fiSelf := fs.FileindexSelf{}
//	existFi, err := fiSelf.FindByFileid(fi.Hash.B58String())
//	if err != nil {
//		return err
//	}
//	if existFi != nil {
//		//已经存在，则修改
//		err = fiSelf.UpdateValue(fi.Hash.B58String(), fi.JSON())
//	} else {
//		//不存在，则新增
//		fiSelf = fs.FileindexSelf{
//			Vid:    vid,                 //虚拟节点id
//			FileId: fi.Hash.B58String(), //索引哈市值
//			Value:  fi.JSON(),           //内容
//		}
//		err = fiSelf.Add(&fiSelf)
//	}
//	return err

//	//	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
//	//	if err != nil {
//	//		f.Close()
//	//		return err
//	//	}
//	//	//fmt.Println("xxx%+v\n", fi, fi.JSON())
//	//	_, err = f.Write(fi.JSON())
//	//	if err != nil {
//	//		f.Close()
//	//		return err
//	//	}
//	//	f.Close()
//	return nil
//}

/*
	下载一个文件块到本地磁盘
*/
func DownloadFilechunkToLocal(fileinfo *FileIndex, no uint64, dps ...*DownProc) error {
	filehash := fileinfo.Hash
	var fc *FileChunk

	for _, one := range fileinfo.FileChunk {
		if one.No == no {
			fc = one
			break
		}
	}

	// for _, v := range fileinfo.FileChunk.GetAll() {
	// 	one := v.(*FileChunk)
	// 	if one.No == no {
	// 		fc = one
	// 		break
	// 	}
	// }
	//优先判断本地是否有文件块的缓存
	chunkpath := filepath.Join(gconfig.Store_dir, fc.Hash.B58String())
	ok, err := utils.PathExists(chunkpath)
	if err != nil {
		return err
	}
	if ok {
		//更新下载进速与速度
		if len(dps) > 0 && dps[0] != nil {
			sz := dps[0].GetSize(chunkpath)
			dps[0].UpdateDownProc(fc.Hash.B58String(), sz)
		}
		return nil
	}

	//不在本地，去网络上下载
	fcVO := FileChunkVO{
		FileHash:  filehash, //完整文件hash
		No:        fc.No,    //文件块编号
		ChunkHash: fc.Hash,  //块 hash
		Index:     0,        //下载块起始位置
		Length:    204800,   //下载块长度
	}

	contentlength := fc.Size
	chunknamepath := filepath.Join(gconfig.Store_dir, fc.Hash.B58String())
	chunkcachenamepath := chunknamepath + "_tmp"
	//对在线共享者排序，更新时间越近，优先
	users := fc.GetUserOnline()
	usersort := SortSU(users)
	for _, one := range usersort {
		//one.Name = fc.RandUser()
		//如果块已经下载完成，则退出
		ok, err := utils.PathExists(filepath.Join(gconfig.Store_dir, fc.Hash.B58String()))
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		newfile, err := os.OpenFile(chunkcachenamepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		defer newfile.Close()
		if err != nil {
			// fmt.Println(err)
			return err
		}
		//fmt.Println(fc.Hash.B58String(), "共享的用户", one.Name.B58String())
		for {
			stat, err := newfile.Stat() //获取文件状态
			if err != nil {
				newfile.Close()
				return errors.New("读取块缓存文件状态失败")
			}
			//更新下载进度与速度
			if len(dps) > 0 && dps[0] != nil {
				size := uint64(stat.Size())
				dps[0].UpdateDownProc(fc.Hash.B58String(), size)
				if stop, ok := StopDown[filehash.B58String()]; ok {
					if stop {
						return errors.New("暂停下载...")
					}
				}
			}
			if uint64(stat.Size()) == contentlength && contentlength != 0 {
				// 文件已下载完
				newfile.Close()
				os.Rename(chunkcachenamepath, chunknamepath)
				return nil
			}

			recvid := one.Nid
			fcVO.Index = uint64(stat.Size())
			content := fcVO.JSON()
			fmt.Println("***********请求下载***********")
			fmt.Println("块:", fc.Hash.B58String())
			fmt.Println("发送给:", recvid.B58String())
			fmt.Println("发送起止:", fcVO.Index, fcVO.Index+fcVO.Length)
			fmt.Println("当前大小:", stat.Size(), contentlength)
			fmt.Println("*************end****************")

			//TODO 共享者可能是普通节点，需要查找到他的超级节点地址
			message, ok, _ := mc.SendP2pMsg(config.MSGID_store_downloadFileChunk, &recvid, &content)
			if ok {

				bs := flood.WaitRequest(mc.CLASS_downloadfile, hex.EncodeToString(message.Body.Hash), 0)
				if bs == nil {
					fmt.Println("返回的下载结果没数据，可能超时")
					break
				}
				downloadfilechunk := ParseFileChunkVO(*bs)
				if uint64(stat.Size()) == contentlength && contentlength != 0 {
					//文件已下载完
					newfile.Close()
					os.Rename(chunkcachenamepath, chunknamepath)
					return nil
				}
				if downloadfilechunk.Content == nil || len(downloadfilechunk.Content) <= 0 {
					// fmt.Println("下载的文件大小为0")
					break
				}
				fmt.Printf("写入位置 %d 写入大小 %d", stat.Size(), len(downloadfilechunk.Content))
				newfile.Seek(stat.Size(), 0)
				newfile.Write(downloadfilechunk.Content)
				newfile.Sync()

			} else {
				break
			}
		}
		newfile.Close()
	}
	return errors.New("没有共享用户")
}

///*
//	程序启动时加载本地磁盘缓存的文件信息
//*/
//func LoadFileInfoLocal() error {
//	fiLocal := fs.FileindexLocal{}
//	fis, err := fiLocal.Getall()
//	if err != nil {
//		return err
//	}

//	for _, one := range fis {
//		fileinfo, err := ParseFileindex(one.Value)
//		if err != nil {
//			return err
//		}
//		AddFileindexToLocal(fileinfo, false, one.Vid)
//	}

//	return filepath.Walk(gconfig.Store_fileinfo_local, func(path string, f os.FileInfo, err error) error {
//		// fmt.Println("path", path, gconfig.Store_fileinfo_local)

//		//		fmt.Println(path, f.Name(), f)
//		if path == gconfig.Store_fileinfo_local {
//			return nil
//		}
//		file, err := os.Open(path)
//		if err != nil {
//			// fmt.Println("-1-1", err)
//			return err
//		}
//		buf := bytes.NewBuffer(nil)
//		_, err = io.Copy(buf, file)
//		file.Close()
//		if err != nil {
//			// fmt.Println("-2-2", err)
//			// fmt.Println(err)
//			return err
//		}

//		fileinfo, err := ParseFileindex(buf.Bytes())

//		//		fileinfo := new(FileInfo)
//		//		err = json.Unmarshal(buf.Bytes(), fileinfo)
//		if err != nil {
//			// fmt.Println("-3-3", err)
//			return err
//		}
//		//		fileinfo.lock = new(sync.RWMutex)
//		//		fmt.Println("0000", string(fileinfo.JSON()))
//		AddFileindexToLocal(fileinfo, false)
//		return nil
//	})

//}
