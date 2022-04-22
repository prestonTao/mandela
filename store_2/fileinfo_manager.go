package store

/*
	保存本地缓存的
*/
var cacheFileinfo = make(map[string]*FileInfo)

/*
	持久化存储文件信息到本地磁盘
*/
func StoreFileinfo() {

}

/*
	通过文件hash值查找文件信息
*/
func FindFileinfo(hs string) *FileInfo {
	//查找本地磁盘
	fileinfo := FindFileinfoToLocal(hs)
	if fileinfo == nil {
		//本地缓存查找
		fileinfo = FindFileinfoToSelf(hs)
	}
	if fileinfo == nil {
		//本地缓存查找
		fileinfo = FindFileinfoToCache(hs)
	}
	if fileinfo == nil {
		//本地网络查找
		fileinfo = FindFileinfoToNet(hs)
	}
	return fileinfo
}
