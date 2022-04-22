/*
	保存本地磁盘上已经存在的文件信息
	提供网络查找
*/
package sharebox

import (
	"mandela/config"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/utils"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// var localFileinfo = new(sync.Map)

// /*
// 	保存文件索引到本地内存和磁盘
// 	@cover    bool    是否保存（覆盖）到本地磁盘
// */
// func AddFileinfoToLocal(fi *FileInfo, cover bool) error {
// 	localFileinfo.Store(fi.Hash.B58String(), fi)
// 	//添加定时任务，定时更新文件索引
// 	//task.Add(time.Now().Unix(), Task_class_share_local_fileinfo, fi.Hash.B58String())
// 	if cover {
// 		return saveFileinfoToLocal(filepath.Join(sconfig.Store_fileinfo_local, fi.Hash.B58String()), fi)
// 	} else {
// 		return nil
// 	}
// }

// func FindFileinfoToLocal(name string) *FileInfo {
// 	if value, ok := localFileinfo.Load(name); ok {
// 		return value.(*FileInfo)
// 	}
// 	return nil

// 	//	if value, ok := localFileinfo.Load(name); ok {
// 	//		fi := value.(*FileInfo)
// 	//		//内存中有记录还要验证磁盘中记录是否存在
// 	//		if _, err := os.Stat(filepath.Join(gconfig.Store_dir, name)); err != nil {
// 	//			if os.IsNotExist(err) {
// 	//				//文件不存在了，从内存中删除记录
// 	//				localFileinfo.Delete(name)
// 	//				//并且删除磁盘上的索引文件
// 	//				os.Remove(filepath.Join(gconfig.Store_fileinfo_local, name))
// 	//			}
// 	//			return nil
// 	//		}
// 	//		if _, err := os.Stat(filepath.Join(gconfig.Store_fileinfo_local, name)); err != nil {
// 	//			if os.IsNotExist(err) {
// 	//				//索引文件丢失没关系，重新保存到磁盘上
// 	//				saveFileinfoToLocal(filepath.Join(gconfig.Store_fileinfo_local, fi.Hash.B58String()), fi)
// 	//				return fi
// 	//			}
// 	//			return nil
// 	//		}
// 	//		return fi
// 	//	}
// 	//	return nil
// }

// /*
// 	本地查找一个文件的块
// */
// // func FindFileChunk(filehash, chunkhash string) bool {
// // 	//先查找是否有这个文件
// // 	fi := FindFileinfoToLocal(filehash)
// // 	if fi == nil {
// // 		return false
// // 	}
// // 	return fi.Have(chunkhash)
// // }

// /*
// 	保存文件索引到本地磁盘
// */
// func saveFileinfoToLocal(path string, fi *FileInfo) error {
// 	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
// 	if err != nil {
// 		f.Close()
// 		return err
// 	}
// 	//fmt.Println("xxx%+v\n", fi, fi.JSON())
// 	_, err = f.Write(fi.JSON())
// 	if err != nil {
// 		f.Close()
// 		return err
// 	}
// 	f.Close()
// 	return nil
// }

/*
	下载一个文件到本地磁盘
*/
func DownloadFilechunkToLocal(fileinfo *FileIndex, dps ...*DownProc) error {
	fileAbsPath := filepath.Join(config.Store_files, fileinfo.Hash.B58String())

	//优先判断本地是否有文件缓存
	ok, err := utils.PathExists(fileAbsPath)
	if err != nil {
		// fmt.Println("err 11111")
		return err
	}
	if ok {
		//更新下载进速与速度
		if len(dps) > 0 && dps[0] != nil {
			sz := dps[0].GetSize(fileAbsPath)
			dps[0].UpdateDownProc(fileinfo.Hash.B58String(), sz)
		}
		return nil
	}

	//不在本地，去网络上下载
	chunknamepath := filepath.Join(fileAbsPath)
	newfile, err := os.OpenFile(chunknamepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer newfile.Close()
	if err != nil {
		// fmt.Println(err)
		return err
	}

	fcVO := FileChunkVO{
		FileHash: fileinfo.Hash, //完整文件hash
		Index:    0,             //下载块起始位置
		Length:   1024 * 1024,   //下载块长度
	}

	//对在线共享者排序，更新时间越近，优先
	users := fileinfo.GetUserOnline()
	fmt.Println("获取10分钟内在线用户", len(users))
	usersort := SortSU(users)
	for _, one := range usersort {
		//one.Name = fc.RandUser()
		//如果块已经下载完成，则退出
		// ok, err := utils.PathExists(filepath.Join(gconfig.Store_dir, fileinfo.Hash.B58String()))
		// if err != nil {
		// 	return err
		// }
		// if ok {
		// 	return nil
		// }

		// newfile, err := os.OpenFile(chunkcachenamepath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		// defer newfile.Close()
		// if err != nil {
		// 	fmt.Println(err)
		// 	return err
		// }
		//fmt.Println(fc.Hash.B58String(), "共享的用户", one.Name.B58String())
		for {
			stat, err := newfile.Stat() //获取文件状态
			if err != nil {
				// newfile.Close()
				//读取块缓存文件状态失败
				return errors.New("Failed to read block cache file state")
			}
			//更新下载进度与速度
			if len(dps) > 0 && dps[0] != nil {
				size := uint64(stat.Size())
				dps[0].UpdateDownProc(fileinfo.Hash.B58String(), size)
				if stop, ok := StopDown[fileinfo.Hash.B58String()]; ok {
					if stop {
						//暂停下载...
						return errors.New("Pause download...")
					}
				}
			}
			if uint64(stat.Size()) >= fileinfo.Size {
				// fmt.Println("文件已下载完")
				// newfile.Close()
				// os.Rename(chunkcachenamepath, chunknamepath)
				return nil
			}

			recvid := one.Name
			fcVO.Index = uint64(stat.Size())
			// fcVO.Length += fcVO.Index
			if fcVO.Length > fileinfo.Size-fcVO.Index {
				fcVO.Length = fileinfo.Size - fcVO.Index
			}
			content := fcVO.JSON()
			// fmt.Println("***********请求下载***********")
			// // fmt.Println("块:", fc.Hash.B58String())
			// fmt.Println("发送给:", recvid.B58String())
			// fmt.Println("发送起止:", fcVO.Index, fcVO.Length)
			// fmt.Println("当前大小:", stat.Size())
			// fmt.Println("*************end****************")

			//TODO 共享者可能是普通节点，需要查找到他的超级节点地址
			message, ok, _ := mc.SendP2pMsg(config.MSGID_sharebox_downloadFileChunk, recvid, &content)
			if ok {

				// }

				// mhead := mc.NewMessageHead(recvid, recvid, true)
				// mbody := mc.NewMessageBody(&content, "", nil, 0)
				// message := mc.NewMessage(mhead, mbody)
				// if message.Send(MSGID_downloadFileChunk) {
				// bs := flood.WaitRequest(mc.CLASS_downloadfile, hex.EncodeToString(message.Body.Hash), 0)
				bs, _ := flood.WaitRequest(mc.CLASS_downloadfile, utils.Bytes2string(message.Body.Hash), 0)
				if bs == nil {
					// fmt.Println("返回的下载结果没数据，可能超时")
					break
				}
				downloadfilechunk := ParseFileChunkVO(*bs)
				// if downloadfilechunk.ContentLength >= contentlength {
				// 	contentlength = downloadfilechunk.ContentLength
				// }
				// if uint64(stat.Size()) == contentlength && contentlength != 0 {
				// 	fmt.Println("文件已下载完")
				// 	newfile.Close()
				// 	os.Rename(chunkcachenamepath, chunknamepath)
				// 	return nil
				// }
				if downloadfilechunk.Content == nil || len(downloadfilechunk.Content) <= 0 {
					// fmt.Println("下载的文件大小为0")
					break
				}
				if len(downloadfilechunk.Content) != int(fcVO.Length) {
					// fmt.Println("收到的数据小于当前请求的大小，丢弃")
					// } else if len(downloadfilechunk.Content) > int(fcVO.Length-fcVO.Index) { //兼容老版本
					// 	newfile.Write(downloadfilechunk.Content)
				} else {
					newfile.Seek(stat.Size(), 0)
					newfile.Write(downloadfilechunk.Content)
				}

			} else {
				break
			}
		}
		newfile.Close()
	}
	//没有共享用户
	return errors.New("not share user")

}

// /*
// 	程序启动时加载本地磁盘缓存的文件信息
// */
// func LoadFileInfoLocal() error {
// 	return filepath.Walk(sconfig.Store_fileinfo_local, func(path string, f os.FileInfo, err error) error {
// 		fmt.Println("path", path, sconfig.Store_fileinfo_local)

// 		//		fmt.Println(path, f.Name(), f)
// 		if path == sconfig.Store_fileinfo_local {
// 			return nil
// 		}
// 		file, err := os.Open(path)
// 		if err != nil {
// 			fmt.Println("-1-1", err)
// 			return err
// 		}
// 		buf := bytes.NewBuffer(nil)
// 		_, err = io.Copy(buf, file)
// 		file.Close()
// 		if err != nil {
// 			fmt.Println("-2-2", err)
// 			fmt.Println(err)
// 			return err
// 		}

// 		fileinfo, err := ParseFileinfo(buf.Bytes())

// 		//		fileinfo := new(FileInfo)
// 		//		err = json.Unmarshal(buf.Bytes(), fileinfo)
// 		if err != nil {
// 			fmt.Println("-3-3", err)
// 			return err
// 		}
// 		//		fileinfo.lock = new(sync.RWMutex)
// 		//		fmt.Println("0000", string(fileinfo.JSON()))
// 		AddFileinfoToLocal(fileinfo, false)
// 		return nil
// 	})

// }
