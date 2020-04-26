/*
	保存网络中共享的文件索引
	只有索引，没有文件块
*/
package store

import (
	"mandela/core/virtual_node"
	"mandela/store/fs"
	"time"
)

//var netFileinfo = new(sync.Map)

/*
	保存文件索引到本地内存和磁盘
	@cover    bool    是否保存（覆盖）到本地磁盘
*/
func AddFileindexToNet(fi *FileIndex, vid virtual_node.AddressNetExtend) error {
	//合并FileUser
	//	fiold, ok := netFileinfo.Load(fi.Hash.B58String())
	//	if ok {
	//		fi = MergeFileUser(fiold.(*FileIndex), fi)
	//	}
	//	netFileinfo.Store(fi.Hash.B58String(), fi)

	vidSelf := virtual_node.FindNearVnodeInSelf(&vid)

	fiTable := fs.FileindexNet{}
	fiLocal, _, _ := FindFileindexToNet(*fi.Hash)
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
		fiTable = fs.FileindexNet{
			Vid:    vidSelf.B58String(), //虚拟节点id
			FileId: fi.Hash.B58String(), //索引哈市值
			Value:  fi.JSON(),           //内容
		}
		return fiTable.Add(&fiTable)
	}

	//	fiTable := fs.FileindexNet{}
	//	existFi, err := fiTable.FindByFileid(*fi.Hash)
	//	if err != nil {
	//		return err
	//	}
	//	if existFi != nil {
	//		fileinfo, err := ParseFileindex(existFi.Value)
	//		if err != nil {
	//			// fmt.Println(err)
	//			return err
	//		}
	//		//合并FileUser
	//		fi = MergeFileUser(fileinfo, fi)
	//		//已经存在，则修改
	//		err = fiTable.UpdateValue(*fi.Hash, fi.JSON())
	//	} else {
	//		//不存在，则新增
	//		fiTable = fs.FileindexNet{
	//			Vid:    vid,       //虚拟节点id
	//			FileId: *fi.Hash,  //索引哈市值
	//			Value:  fi.JSON(), //内容
	//		}
	//		err = fiTable.Add(&fiTable)
	//	}
	//	return err
	//		return saveFileinfoToLocal(fi, vid)

}

func FindFileindexToNet(fileIndexId virtual_node.AddressNetExtend) (*FileIndex, virtual_node.AddressNetExtend, error) {
	fiTable := fs.FileindexNet{}
	fiDb, err := fiTable.FindByFileid(fileIndexId.B58String())
	if err != nil {
		return nil, nil, err
	}
	fi, err := ParseFileindex(fiDb.Value)
	if err != nil {
		return nil, nil, err
	}
	//如果所有者为空，删除文件索引
	if len(fi.FileOwner) == 0 {
		err = fiTable.Del(fileIndexId.B58String())
		return nil, nil, err
	}
	vid := virtual_node.AddressFromB58String(fiDb.Vid)
	return fi, vid, nil

	//	if value, ok := netFileinfo.Load(name); ok {
	//		finfo := value.(*FileIndex)
	//		//如果所有者为空，删除文件索引
	//		if len(finfo.FileOwner) == 0 {
	//			netFileinfo.Delete(name)
	//			os.Remove(filepath.Join(gconfig.Store_fileinfo_net, name))
	//			return nil
	//		}
	//		return finfo
	//	}
	//	return nil
}

func GetFileindexToNetAll() ([]virtual_node.AddressNetExtend, []*FileIndex) {
	fiTable := fs.FileindexNet{}
	fins, err := fiTable.Getall()
	if err != nil {
		return nil, nil
	}
	names := make([]virtual_node.AddressNetExtend, 0)
	fis := make([]*FileIndex, 0)
	for i, one := range fins {
		fid := virtual_node.AddressNetExtend(fins[i].FileId)
		names = append(names, fid)

		fi, err := ParseFileindex(one.Value)
		if err != nil {
			continue
		}
		fis = append(fis, fi)
	}

	//	netFileinfo.Range(func(key, value interface{}) bool {
	//		names = append(names, key.(string))
	//		fis = append(fis, value.(*FileIndex))
	//		return true
	//	})
	return names, fis
}

///*
//	程序启动时加载本地磁盘缓存的文件信息
//*/
//func LoadFileInfoNet() error {
//	err := filepath.Walk(gconfig.Store_fileinfo_net, func(path string, f os.FileInfo, err error) error {

//		//fmt.Println(path, f.Name(), f)
//		if path == gconfig.Store_fileinfo_net {
//			return nil
//		}
//		file, err := os.Open(path)
//		if err != nil {
//			// fmt.Println(err)
//			return err
//		}
//		buf := bytes.NewBuffer(nil)
//		_, err = io.Copy(buf, file)
//		file.Close()
//		if err != nil {
//			// fmt.Println(err)
//			return err
//		}

//		fileinfo, err := ParseFileindex(buf.Bytes())

//		if err != nil {
//			// fmt.Println(err)
//			return err
//		}
//		AddFileindexToNet(fileinfo, false)
//		//如果本地有文件，则同步块索引到1/4节点
//		ok, _ := utils.PathExists(filepath.Join(gconfig.Store_temp, fileinfo.Name))
//		if ok {
//			go SyncFileChunkToPeer(fileinfo)
//		}
//		return nil
//	})
//	if err != nil {
//		return err
//	}

//	go LoopClearFileindexToNet()

//	return nil

//}

/*
	定时清理文件索引，文件索引中超过60天没有用户共享的块删除掉
*/
func LoopClearFileindexToNet() {
	fiTable := fs.FileindexNet{}
	for range time.NewTicker(Time_loopClearUser * time.Second).C {
		remove := make([]*virtual_node.AddressNetExtend, 0)

		_, fis := GetFileindexToNetAll()

		for i, one := range fis {
			have := false
			for _, two := range one.FileChunk {
				two.Clear()
				if len(two.GetShareUserAll()) > 0 {
					have = true
				}
			}
			// for _, v := range one.FileChunk.GetAll() {
			// 	one := v.(*FileChunk)
			// 	one.Clear()
			// 	if len(one.GetUserAll()) > 0 {
			// 		have = true
			// 	}
			// }

			//如果文件索引没有共享用户，则删除这个文件索引
			if !have {
				remove = append(remove, fis[i].Hash)
			}
		}

		for _, one := range remove {
			fiTable.Del(one.B58String())
		}

		//		netFileinfo.Range(func(key, value interface{}) bool {
		//			have := false
		//			v := value.(*FileIndex)
		//			for _, v := range v.FileChunk.GetAll() {
		//				one := v.(*FileChunk)
		//				one.Clear()
		//				if len(one.GetUserAll()) > 0 {
		//					have = true
		//				}
		//			}

		//			//如果文件索引没有共享用户，则删除这个文件索引
		//			if !have {
		//				remove = append(remove, v.Name)
		//			}
		//			return true
		//		})

		//		for _, one := range remove {
		//			netFileinfo.Delete(one)
		//		}
	}
}
