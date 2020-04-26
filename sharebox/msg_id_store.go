package sharebox

import (
	"os"
)

// const (
// 	MSGID_addFileShare           = 1000 //添加一个文件共享
// 	MSGID_addFileShare_recv      = 1001 //添加一个文件共享 返回
// 	MSGID_findFileinfo           = 1002 //网络中查找一个文件信息
// 	MSGID_findFileinfo_recv      = 1003 //网络中查找一个文件信息 返回
// 	MSGID_getFilesize            = 1004 //网络中查找一个文件信息
// 	MSGID_getFilesize_recv       = 1005 //网络中查找一个文件信息 返回
// 	MSGID_downloadFileChunk      = 1006 //网络中下载文件块
// 	MSGID_downloadFileChunk_recv = 1007 //网络中下载文件块 返回
// 	MSGID_getUploadinfo          = 1008 //返回节点上传地址信息
// 	MSGID_getUploadinfo_recv     = 1009 //返回节点上传地址信息 返回
// 	// MSGID_syncFileInfo           = 1010 //返回同步文件信息到1/4节点
// 	// MSGID_syncFileInfo_recv      = 1012 //返回同步文件信息到1/4节点
// 	// MSGID_getfourNodeinfo      = 1013 //返回节点上传地址信息
// 	// MSGID_getfourNodeinfo_recv = 1014 //返回节点上传地址信息 返回

// 	MSGID_getNodeWalletReceiptAddress      = 1015 //查询节点收款地址
// 	MSGID_getNodeWalletReceiptAddress_recv = 1016 //查询节点上传地址信息 返回

// 	MSGID_getsharefolderlist      = 1017 //获取节点共享文件列表
// 	MSGID_getsharefolderlist_recv = 1018 //获取节点共享文件列表 返回

// )

/*
	判断一个路径的文件是否存在
*/
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Mkdir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	//	err := os.Mkdir(path, os.ModeDir)
	if err != nil {
		// fmt.Println("创建文件夹失败", path, err)
		return err
	}
	return nil
}
