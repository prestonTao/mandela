package cloud_space

import (
	// "mandela/cloud_reward/client"
	// "mandela/cloud_reward/server"
	"mandela/cloud_space/fs"
	"mandela/config"
	gconfig "mandela/config"
	"mandela/core/message_center"
	"mandela/core/utils"
	"mandela/sqlite3_db"
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
// 	MSGID_syncFileInfo           = 1010 //返回同步文件信息到1/4节点
// 	MSGID_syncFileInfo_recv      = 1012 //返回同步文件信息到1/4节点
// 	MSGID_getfourNodeinfo        = 1013 //返回节点上传地址信息
// 	MSGID_getfourNodeinfo_recv   = 1014 //返回节点上传地址信息 返回

// 	MSGID_getNodeWalletReceiptAddress      = 1015 //查询节点收款地址
// 	MSGID_getNodeWalletReceiptAddress_recv = 1016 //查询节点上传地址信息 返回
// )

func RegsterCloudSpace() error {
	//删除本地所有文件分片及分片索引
	if gconfig.IsRemoveStore {
		err := os.RemoveAll(gconfig.Store_dir)
		if err != nil {
			return err
		}
	}

	//创建保存文件的文件夹
	utils.CheckCreateDir(gconfig.Store_dir)
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(filepath.Join(gconfig.Store_fileinfo_self))
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(gconfig.Store_fileinfo_local)
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(gconfig.Store_fileinfo_net)
	//创建保存文件索引的文件夹
	// utils.CheckCreateDir(gconfig.Store_fileinfo_cache)
	//创建临时文件夹
	utils.CheckCreateDir(gconfig.Store_temp)
	//创建带扩展的完整文件存放文件夹
	utils.CheckCreateDir(gconfig.Store_files)

	//链接数据库
	sqlite3_db.Init()

	//initTask()

	fs.Init()
	fs.LoadAllSpace()
	// TimeSyncChunk()
	initTask()

	//加载自己上传的文件索引，加载本地维护的文件块，定时同步
	loadFileindex()

	//加载自己共享的文件
	//	err := LoadFileInfoSelf()
	//	if err != nil {
	//		return err
	//	}
	//加载本地文件索引
	//	err = LoadFileInfoLocal()
	//	if err != nil {
	//		return err
	//	}
	//加载网络文件索引
	//	err = LoadFileInfoNet()
	//	if err != nil {
	//		return err
	//	}

	message_center.Register_vnode_search(config.MSGID_store_addFileShare, AddFileShare)         //添加一个文件共享
	message_center.Register_p2p(config.MSGID_store_addFileShare_recv, AddFileShare_recv)        //添加一个文件共享 返回
	message_center.Register_vnode_search(config.MSGID_store_findFileinfo, FindFileindexHandler) //网络中查找一个文件信息
	message_center.Register_p2p(config.MSGID_store_findFileinfo_recv, FindFileindex_recv)       //网络中查找一个文件信息 返回
	//	message_center.Register_p2p(config.MSGID_store_getFilesize, FindFilesize)
	//	message_center.Register_p2p(config.MSGID_store_getFilesize_recv, FindFilesize_recv)
	message_center.Register_p2p(config.MSGID_store_downloadFileChunk, DownloadFilechunk)           //网络中下载文件块
	message_center.Register_p2p(config.MSGID_store_downloadFileChunk_recv, DownloadFilechunk_recv) //网络中下载文件块 返回
	//	message_center.Register_p2p(config.MSGID_store_getUploadinfo, Uploadinfo)
	//	message_center.Register_p2p(config.MSGID_store_getUploadinfo_recv, Uploadinfo_recv)
	message_center.Register_vnode_search(config.MSGID_store_syncFileInfo, syncFileInfo)        //返回同步文件信息到1/4节点
	message_center.Register_p2p(config.MSGID_store_syncFileInfo_recv, syncFileInfo_recv)       //返回同步文件信息到1/4节点
	message_center.Register_p2p(config.MSGID_store_getfourNodeinfo, GetfourNodeinfo)           //返回节点上传地址信息
	message_center.Register_p2p(config.MSGID_store_getfourNodeinfo_recv, GetfourNodeinfo_recv) //返回节点上传地址信息 返回
	//验证空间大小
	message_center.Register_p2p(config.MSGID_store_getFileindexList, CheckSpaceInfo)           //获取节点的文件列表，同时可以验证空间大小
	message_center.Register_p2p(config.MSGID_store_getFileindexList_recv, CheckSpaceInfo_recv) //获取节点的文件列表，同时可以验证空间大小 返回
	//engine.RegisterMsg(MSGID_getNodeWalletReceiptAddress, GetWalletAddr)
	//engine.RegisterMsg(MSGID_getNodeWalletReceiptAddress_recv, GetWalletAddr_recv)
	// message_center.Register_vnode_search(config.MSGID_store_addFileOwner, AddFileOwner)  //添加一个文件拥有者
	// message_center.Register_p2p(config.MSGID_store_addFileOwner_recv, AddFileOwner_recv) //添加一个文件拥有者 返回

	// go client.Start()
	// go server.Start()
	return nil
}

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
