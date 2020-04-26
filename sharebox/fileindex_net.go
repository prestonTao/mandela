/*
	保存网络中共享的文件索引
	只有索引，没有文件块
*/
package sharebox

import (
	sconfig "mandela/sharebox/config"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var netFileinfo = new(sync.Map)

//func AddFileinfoToNet(fi *FileInfo) {
//	netFileinfo.Store(fi.Hash, fi)
//}

/*
	保存文件索引到本地内存和磁盘
	@cover    bool    是否保存（覆盖）到本地磁盘
*/
func AddFileindexToNet(fi *FileIndex, cover bool) error {

	netFileinfo.Store(fi.Hash.B58String(), fi)
	return nil

	//	for _, v := range fi.FileChunk.GetAll() {
	//		one := v.(*FileChunk)
	//		for _, two := range one.GetUserAll() {
	//			cofs := NewCheckOnlineFileShare(fi.Hash.B58String(), one.Hash.B58String(), two.Name.B58String())
	//			//添加定时任务，定时删除共享块的用户
	//			//task.Add(two.UpdateTime+Time_sharefile, Task_class_net_fileinfo_remove_user, string(cofs.JSON()))
	//		}
	//	}

	//	for _, one := range fi.FileChunk {
	//		for _, two := range one.GetUserAll() {
	//			cofs := NewCheckOnlineFileShare(one.Hash.B58String(), one.Hash.B58String(), two.Name.B58String())
	//			//添加定时任务，定时删除共享块的用户
	//			task.Add(two.UpdateTime+Time_shareUserOfflineClear, Task_class_net_fileinfo_remove_user, string(cofs.JSON()))
	//		}
	//	}

	// if cover {
	// 	return saveFileinfoToLocal(filepath.Join(sconfig.Store_fileinfo_net, fi.Hash.B58String()), fi)
	// } else {
	// 	return nil
	// }
}

func FindFileindexToNet(name string) *FileIndex {
	if value, ok := netFileinfo.Load(name); ok {
		return value.(*FileIndex)
	}
	return nil
}

func GetFileinfoToNetAll() ([]string, []*FileIndex) {
	names := make([]string, 0)
	fis := make([]*FileIndex, 0)
	netFileinfo.Range(func(key, value interface{}) bool {
		names = append(names, key.(string))
		fis = append(fis, value.(*FileIndex))
		return true
	})
	return names, fis
}

/*
	程序启动时加载本地磁盘缓存的文件信息
*/
func LoadFileInfoNet() error {
	err := filepath.Walk(sconfig.Store_fileinfo_net, func(path string, f os.FileInfo, err error) error {

		//fmt.Println(path, f.Name(), f)
		if path == sconfig.Store_fileinfo_net {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			// fmt.Println(err)
			return err
		}
		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, file)
		file.Close()
		if err != nil {
			// fmt.Println(err)
			return err
		}

		fileinfo, err := ParseFileindex(buf.Bytes())
		if err != nil {
			// fmt.Println(err)
			return err
		}
		AddFileindexToNet(fileinfo, false)
		//如果本地有文件，则同步块索引到1/4节点
		// ok, _ := utils.PathExists(filepath.Join(gconfig.Store_temp, fileinfo.Name))
		// if ok {
		// 	go SyncFileChunkToPeer(fileinfo)
		// }
		return nil
	})
	if err != nil {
		return err
	}

	go LoopClearFileinfoToNet()

	return nil

}

/*
	定时清理文件索引，文件索引中超过60天没有用户共享的块删除掉
*/
func LoopClearFileinfoToNet() {
	for range time.NewTicker(sconfig.Time_loopClearUser * time.Second).C {
		remove := make([]string, 0)
		netFileinfo.Range(func(key, value interface{}) bool {
			// have := false
			v := value.(*FileIndex)
			total := v.Clear()

			//如果文件索引没有共享用户，则删除这个文件索引
			if total <= 0 {
				remove = append(remove, v.Name)
			}

			return true
		})

		for _, one := range remove {
			netFileinfo.Delete(one)
		}
	}
}
