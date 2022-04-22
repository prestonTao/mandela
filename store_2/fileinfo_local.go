/*
	保存本地磁盘上已经存在的文件信息
	提供网络查找
*/
package store

import (
	"mandela/config"
	"mandela/core/message_center/flood"
	"bytes"
	"errors"

	gconfig "mandela/config"
	mc "mandela/core/message_center"
	"mandela/core/utils"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var localFileinfo = new(sync.Map)

/*
	保存文件索引到本地内存和磁盘
	@cover    bool    是否保存（覆盖）到本地磁盘
*/
func AddFileinfoToLocal(fi *FileInfo, cover bool) error {
	localFileinfo.Store(fi.Hash.B58String(), fi)
	//添加定时任务，定时更新文件索引
	//task.Add(time.Now().Unix(), Task_class_share_local_fileinfo, fi.Hash.B58String())
	if cover {
		return saveFileinfoToLocal(filepath.Join(gconfig.Store_fileinfo_local, fi.Hash.B58String()), fi)
	} else {
		return nil
	}
}

func FindFileinfoToLocal(name string) *FileInfo {
	if value, ok := localFileinfo.Load(name); ok {
		return value.(*FileInfo)
	}
	return nil
}

/*
	本地查找一个文件的块
*/
func FindFileChunk(filehash, chunkhash string) bool {
	//先查找是否有这个文件
	fi := FindFileinfoToLocal(filehash)
	if fi == nil {
		return false
	}
	return fi.Have(chunkhash)
}

/*
	保存文件索引到本地磁盘
*/
func saveFileinfoToLocal(path string, fi *FileInfo) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		f.Close()
		return err
	}
	//fmt.Println("xxx%+v\n", fi, fi.JSON())
	_, err = f.Write(fi.JSON())
	if err != nil {
		f.Close()
		return err
	}
	f.Close()
	return nil
}

/*
	下载一个文件块到本地磁盘
*/
func DownloadFilechunkToLocal(fileinfo *FileInfo, no uint64, dps ...*DownProc) error {
	filehash := fileinfo.Hash
	var fc *FileChunk
	for _, v := range fileinfo.FileChunk.GetAll() {
		one := v.(*FileChunk)
		if one.No == no {
			fc = one
			break
		}
	}
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

			recvid := one.Name
			fcVO.Index = uint64(stat.Size())
			content := fcVO.JSON()
			fmt.Println("***********请求下载***********")
			fmt.Println("块:", fc.Hash.B58String())
			fmt.Println("发送给:", recvid.B58String())
			fmt.Println("发送起止:", fcVO.Index, fcVO.Index+fcVO.Length)
			fmt.Println("当前大小:", stat.Size(), contentlength)
			fmt.Println("*************end****************")

			//TODO 共享者可能是普通节点，需要查找到他的超级节点地址
			message, ok, _ := mc.SendP2pMsg(config.MSGID_store_downloadFileChunk, recvid, &content)
			if ok {

				bs := flood.WaitRequest(mc.CLASS_downloadfile, message.Body.Hash.B58String())
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

/*
	程序启动时加载本地磁盘缓存的文件信息
*/
func LoadFileInfoLocal() error {
	return filepath.Walk(gconfig.Store_fileinfo_local, func(path string, f os.FileInfo, err error) error {
		// fmt.Println("path", path, gconfig.Store_fileinfo_local)

		//		fmt.Println(path, f.Name(), f)
		if path == gconfig.Store_fileinfo_local {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			// fmt.Println("-1-1", err)
			return err
		}
		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, file)
		file.Close()
		if err != nil {
			// fmt.Println("-2-2", err)
			// fmt.Println(err)
			return err
		}

		fileinfo, err := ParseFileinfo(buf.Bytes())

		//		fileinfo := new(FileInfo)
		//		err = json.Unmarshal(buf.Bytes(), fileinfo)
		if err != nil {
			// fmt.Println("-3-3", err)
			return err
		}
		//		fileinfo.lock = new(sync.RWMutex)
		//		fmt.Println("0000", string(fileinfo.JSON()))
		AddFileinfoToLocal(fileinfo, false)
		return nil
	})

}
