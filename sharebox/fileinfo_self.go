/*
	保存自己上传的文件索引
*/
package sharebox

// import (
// 	sconfig "mandela/sharebox/config"
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"os"
// 	"path/filepath"
// 	"sync"
// )

// var selfFileinfo = new(sync.Map)

// /*
// 	保存文件索引到本地内存和磁盘
// 	@cover    bool    是否保存（覆盖）到本地磁盘
// */
// func AddFileinfoToSelf(fi *FileInfo, cover bool) error {
// 	selfFileinfo.Store(fi.Hash.B58String(), fi)
// 	//添加定时任务，定时更新文件索引
// 	//task.Add(time.Now().Unix(), Task_class_share_self_fileinfo, fi.Hash.B58String())
// 	if cover {
// 		return saveFileinfoToLocal(filepath.Join(sconfig.Store_fileinfo_self, fi.Hash.B58String()), fi)
// 	} else {
// 		return nil
// 	}
// }

// func FindFileinfoToSelf(name string) *FileInfo {
// 	if value, ok := selfFileinfo.Load(name); ok {
// 		return value.(*FileInfo)
// 	}
// 	return nil
// }

// //func GetFileinfoToNetAll() ([]string, []*FileInfo) {
// //	names := make([]string, 0)
// //	fis := make([]*FileInfo, 0)
// //	netFileinfo.Range(func(key, value interface{}) bool {
// //		fmt.Println(key, value)
// //		names = append(names, key.(string))
// //		fis = append(fis, value.(*FileInfo))
// //		return true
// //	})
// //	return names, fis
// //}

// func GetFileinfoToSelfAll() ([]string, []*FileInfo) {
// 	names := make([]string, 0)
// 	fis := make([]*FileInfo, 0)
// 	selfFileinfo.Range(func(key, value interface{}) bool {
// 		names = append(names, key.(string))
// 		fis = append(fis, value.(*FileInfo))
// 		return true
// 	})
// 	return names, fis
// }

// /*
// 	程序启动时加载本地磁盘缓存的文件信息
// */
// func LoadFileInfoSelf() error {
// 	return filepath.Walk(sconfig.Store_fileinfo_self, func(path string, f os.FileInfo, err error) error {

// 		//		fmt.Println(path, f.Name(), f)
// 		if path == sconfig.Store_fileinfo_self {
// 			return nil
// 		}
// 		file, err := os.Open(path)
// 		if err != nil {
// 			fmt.Println(err)
// 			return err
// 		}
// 		buf := bytes.NewBuffer(nil)
// 		_, err = io.Copy(buf, file)
// 		file.Close()
// 		if err != nil {
// 			fmt.Println(err)
// 			return err
// 		}

// 		fileinfo, err := ParseFileinfo(buf.Bytes())

// 		//		fileinfo := new(FileInfo)
// 		//		err = json.Unmarshal(buf.Bytes(), fileinfo)
// 		if err != nil {
// 			fmt.Println(err)
// 			return err
// 		}
// 		//		fileinfo.lock = new(sync.RWMutex)
// 		//		fmt.Println("0000", string(fileinfo.JSON()))
// 		AddFileinfoToSelf(fileinfo, false)
// 		return nil
// 	})

// }
