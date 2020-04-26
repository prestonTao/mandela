/*
	保存自己上传的文件索引
*/
package store

import (
	"mandela/core/engine"
	"mandela/core/virtual_node"
	"mandela/store/fs"
	"fmt"
	"time"
)

//var selfFileinfo = new(sync.Map)

/*
	保存文件索引到本地内存和磁盘
	@cover    bool    是否保存（覆盖）到本地磁盘
*/
func AddFileindexToSelf(fi *FileIndex, vid virtual_node.AddressNetExtend) error {
	//	//合并FileUser
	//	fiold, ok := selfFileinfo.Load(fi.Hash.B58String())
	//	if ok {
	//		fi = MergeFileUser(fiold.(*FileIndex), fi)
	//	}
	//	selfFileinfo.Store(fi.Hash.B58String(), fi)
	//	//添加定时任务，定时更新文件索引
	//	//task.Add(time.Now().Unix(), Task_class_share_self_fileinfo, fi.Hash.B58String())
	//	if cover {
	//		return saveFileinfoToLocal(filepath.Join(gconfig.Store_fileinfo_self, fi.Hash.B58String()), fi)
	//	} else {
	//		return nil
	//	}

	engine.Log.Info("保存文件索引到本地 %s", fi.Hash.B58String())

	vidSelf := virtual_node.FindInVnodeinfoSelf(vid)
	engine.Log.Info("保存文件索引到本地 %s", vidSelf.Vid.B58String())

	fiTable := fs.FileindexSelf{}
	fiLocal, _, err := FindFileindexToSelf(*fi.Hash)
	if fiLocal != nil {
		engine.Log.Info("保存文件索引到本地 修改")
		//
		// fiLocal.MergeFileUser(fi)
		// fiLocal.AddFileOwner()
		//修改
		fiTable.Name = fi.Name
		fiTable.FileId = *fi.Hash
		fiTable.Vid = vidSelf.Vid
		//		if bytes.Equal(vidSelf, vidLocal) {
		//		}
		fiTable.Value = fiLocal.JSON()
		err = fiTable.Update()
	} else {
		engine.Log.Info("保存文件索引到本地 新增")
		//不存在，则新增
		fiTable = fs.FileindexSelf{
			Name:   fi.Name,     //真实文件名称
			Vid:    vidSelf.Vid, //虚拟节点id
			FileId: *fi.Hash,    //索引哈市值
			Value:  fi.JSON(),   //内容
		}
		err = fiTable.Add(&fiTable)
	}
	fmt.Println("打印错误", err)
	//添加定时任务，定时同步到其它节点
	task.Add(time.Now().Unix(), Task_class_ower_self_fileinfo, *fi.Hash)
	return err
	//	fiSelf := fs.FileindexSelf{}
	//	existFi, err := fiSelf.FindByFileid(*fi.Hash)
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
	//		err = fiSelf.UpdateValue(*fi.Hash, fi.JSON())
	//	} else {
	//		//不存在，则新增
	//		fiSelf = fs.FileindexSelf{
	//			Vid:    vid,       //虚拟节点id
	//			Name:   fi.Name,   //文件真实名称
	//			FileId: *fi.Hash,  //索引哈市值
	//			Value:  fi.JSON(), //内容
	//		}
	//		err = fiSelf.Add(&fiSelf)
	//	}
	//	return err

}

/*
	查询自己上传的文件索引
	@name    virtual_node.AddressNetExtend    文件索引hash值
*/
func FindFileindexToSelf(name virtual_node.AddressNetExtend) (*FileIndex, virtual_node.AddressNetExtend, error) {

	fiTable := fs.FileindexSelf{}
	fiDb, err := fiTable.FindByFileid(name)
	if err != nil {
		return nil, nil, err
	}
	if fiDb == nil {
		return nil, nil, nil
	}
	fi, err := ParseFileindex(fiDb.Value)
	if err != nil {
		return nil, nil, err
	}
	//如果所有者为空，删除文件索引
	if len(fi.FileOwner) == 0 {
		err = fiTable.Del(name)
		return nil, nil, err
	}
	vid := virtual_node.AddressNetExtend(fiDb.Vid)
	return fi, vid, nil

	//	if value, ok := selfFileinfo.Load(name); ok {
	//		return value.(*FileIndex)
	//	}
	//	return nil
}

//func GetFileinfoToNetAll() ([]string, []*FileInfo) {
//	names := make([]string, 0)
//	fis := make([]*FileInfo, 0)
//	netFileinfo.Range(func(key, value interface{}) bool {
//		fmt.Println(key, value)
//		names = append(names, key.(string))
//		fis = append(fis, value.(*FileInfo))
//		return true
//	})
//	return names, fis
//}

func GetFileindexToSelfAll() ([]virtual_node.AddressNetExtend, []*FileIndex) {
	fiTable := fs.FileindexSelf{}
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

	//	names := make([]string, 0)
	//	fis := make([]*FileIndex, 0)
	//	selfFileinfo.Range(func(key, value interface{}) bool {
	//		names = append(names, key.(string))
	//		fis = append(fis, value.(*FileIndex))
	//		return true
	//	})
	return names, fis
}

///*
//	程序启动时加载本地磁盘缓存的文件信息
//*/
//func LoadFileInfoSelf() error {
//	return filepath.Walk(gconfig.Store_fileinfo_self, func(path string, f os.FileInfo, err error) error {

//		//		fmt.Println(path, f.Name(), f)
//		if path == gconfig.Store_fileinfo_self {
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

//		//		fileinfo := new(FileInfo)
//		//		err = json.Unmarshal(buf.Bytes(), fileinfo)
//		if err != nil {
//			// fmt.Println(err)
//			return err
//		}
//		//		fileinfo.lock = new(sync.RWMutex)
//		//		fmt.Println("0000", string(fileinfo.JSON()))
//		AddFileindexToSelf(fileinfo, false)
//		return nil
//	})

//}

//删除文件索引
func DelFileInfoFromSelf(hash virtual_node.AddressNetExtend) error {
	fiTable := fs.FileindexSelf{}
	return fiTable.Del(hash)
	//	selfFileinfo.Delete(hash)
	//	err := os.Remove(gconfig.Store_path_dir + "/" + gconfig.Store_path_fileinfo_self + "/" + hash)
	//	return err
}
