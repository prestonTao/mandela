package store

import (
	"mandela/core/virtual_node"
)

/*
	保存本地缓存的
*/
var cacheFileindex = make(map[string]*FileIndex)

/*
	持久化存储文件信息到本地磁盘
*/
func StoreFileinfo() {

}

/*
	通过文件hash值查找文件信息
*/
func FindFileindex(hs virtual_node.AddressNetExtend) *FileIndex {
	//查找本地磁盘
	fileinfo, _ := FindFileindexToLocal(hs)
	if fileinfo == nil {
		//本地缓存查找
		fileinfo, _, _ = FindFileindexToSelf(hs)
	}
	if fileinfo == nil {
		//本地缓存查找
		fileinfo = FindFileindexToCache(hs)
	}
	if fileinfo == nil {
		//本地网络查找
		fileinfo, _, _ = FindFileindexToNet(hs)
	}
	return fileinfo
}
