package fs

import (
	"mandela/core/virtual_node"
)

var fsm = NewFsManager()

func LoadAllSpace() {
	fsm.LoadAllSpace()
}

/*
	保存一个文件块
	@filepath    string    空间文件的保存路径
*/
func SaveFileChunk(vid virtual_node.AddressNetExtend,
	key virtual_node.AddressNetExtend, value *[]byte) {
	fsm.SaveFileChunk(vid, key, value)
}

/*
	添加n个空间
	@filepath    string    空间文件的保存路径
*/
func AddSpace(absPath string, n uint64) {
	fsm.AddSpace(absPath, n)
}

/*
	删除n个空间
*/
func DelSpace(n uint64) {
	fsm.DelSpace(n)
}

/*
	删除指定的一个空间
*/
func DelSpaceForDbPath(dbpath string) {
	fsm.DelSpaceOne(dbpath)
}

/*
	获取空间列表信息
*/
func GetSpaceList() []StorageVO {
	return fsm.GetSpaceList()
}

/*
	获取总空间大小
*/
func GetSpaceSize() uint64 {
	return fsm.GetSpaceNum()
}

/*
	获取已经使用空间大小
*/
func GetUseSpaceSize() uint64 {
	return fsm.GetUseSpaceSize()
}

/*
	获取一个能放下指定大小文件的空间
*/
func GetNotUseSpace(size uint64) *virtual_node.Vnodeinfo {
	return fsm.GetNotUseSpace(size)
}
