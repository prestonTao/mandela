package store

import (
	"bytes"
	// "fmt"
	gconfig "mandela/config"
	"mandela/core/utils"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var cacheLruLock = new(sync.RWMutex)
var cacheLru = utils.NewCache(10000)

/*
	在缓存中查找文件信息
	此缓存没有把文件保存到
*/
func FindFileinfoToCache(name string) *FileInfo {
	cacheLruLock.RLock()
	value, ok := cacheLru.Get(name)
	if ok {
		cacheLruLock.RUnlock()
		return value.(*FileInfo)
	}
	cacheLruLock.RUnlock()
	return nil
}
func addFileinfoToCache(fi *FileInfo) {
	cacheLruLock.Lock()
	cacheLru.Add(fi.Hash, fi)
	cacheLruLock.Unlock()
}

/*
	程序启动时加载本地磁盘缓存的文件信息
*/
func LoadFileInfoCache() error {
	return filepath.Walk(gconfig.Store_fileinfo_cache, func(path string, f os.FileInfo, err error) error {

		//		fmt.Println(path, f.Name(), f)
		if path == gconfig.Store_fileinfo_cache {
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

		//		fileinfo := new(FileInfo)
		//		err = json.Unmarshal(buf.Bytes(), fileinfo)
		if err != nil {
			// fmt.Println(err)
			return err
		}
		//		fileinfo.lock = new(sync.RWMutex)
		//		fmt.Println("0000", string(fileinfo.JSON()))
		addFileinfoToCache(fileinfo)
		return nil
	})

}
