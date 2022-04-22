/*
	保存网络中共享的文件索引
	只有索引，没有文件块
*/
package store

import (
	"bytes"
	// "fmt"
	gconfig "mandela/config"
	"mandela/core/utils"

	//"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var netFileinfo = new(sync.Map)

/*
	保存文件索引到本地内存和磁盘
	@cover    bool    是否保存（覆盖）到本地磁盘
*/
func AddFileinfoToNet(fi *FileInfo, cover bool) error {
	//合并FileUser
	fiold, ok := netFileinfo.Load(fi.Hash.B58String())
	if ok {
		fi = MergeFileUser(fiold.(*FileInfo), fi)
	}
	netFileinfo.Store(fi.Hash.B58String(), fi)

	if cover {
		return saveFileinfoToLocal(filepath.Join(gconfig.Store_fileinfo_net, fi.Hash.B58String()), fi)
	} else {
		return nil
	}
}

func FindFileinfoToNet(name string) *FileInfo {
	if value, ok := netFileinfo.Load(name); ok {
		finfo := value.(*FileInfo)
		//如果所有者为空，删除文件索引
		if len(finfo.FileUser) == 0 {
			netFileinfo.Delete(name)
			os.Remove(filepath.Join(gconfig.Store_fileinfo_net, name))
			return nil
		}
		return finfo
	}
	return nil
}

func GetFileinfoToNetAll() ([]string, []*FileInfo) {
	names := make([]string, 0)
	fis := make([]*FileInfo, 0)
	netFileinfo.Range(func(key, value interface{}) bool {
		names = append(names, key.(string))
		fis = append(fis, value.(*FileInfo))
		return true
	})
	return names, fis
}

/*
	程序启动时加载本地磁盘缓存的文件信息
*/
func LoadFileInfoNet() error {
	err := filepath.Walk(gconfig.Store_fileinfo_net, func(path string, f os.FileInfo, err error) error {

		//fmt.Println(path, f.Name(), f)
		if path == gconfig.Store_fileinfo_net {
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

		fileinfo, err := ParseFileinfo(buf.Bytes())

		if err != nil {
			// fmt.Println(err)
			return err
		}
		AddFileinfoToNet(fileinfo, false)
		//如果本地有文件，则同步块索引到1/4节点
		ok, _ := utils.PathExists(filepath.Join(gconfig.Store_temp, fileinfo.Name))
		if ok {
			go SyncFileChunkToPeer(fileinfo)
		}
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
	for range time.NewTicker(Time_loopClearUser * time.Second).C {
		remove := make([]string, 0)
		//		fmt.Println("haha")
		//		names := make([]string, 0)
		//		fis := make([]*FileInfo, 0)
		netFileinfo.Range(func(key, value interface{}) bool {
			have := false
			v := value.(*FileInfo)
			for _, v := range v.FileChunk.GetAll() {
				one := v.(*FileChunk)
				one.Clear()
				if len(one.GetUserAll()) > 0 {
					have = true
				}
			}

			//如果文件索引没有共享用户，则删除这个文件索引
			if !have {
				remove = append(remove, v.Name)
			}
			return true
		})

		for _, one := range remove {
			netFileinfo.Delete(one)
		}
	}
}
