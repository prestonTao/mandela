package cloud_space

// import (
// 	"mandela/cloud_space/fs"
// 	gconfig "mandela/config"
// 	"mandela/core/utils"
// 	"mandela/sqlite3_db"
// 	"os"
// )

// func RegsterCloudSpace() error {
// 	//删除本地所有文件分片及分片索引
// 	if gconfig.IsRemoveStore {
// 		err := os.RemoveAll(gconfig.Store_dir)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	//创建保存文件的文件夹
// 	utils.CheckCreateDir(gconfig.Store_dir)
// 	//创建保存文件索引的文件夹
// 	// utils.CheckCreateDir(filepath.Join(gconfig.Store_fileinfo_self))
// 	//创建保存文件索引的文件夹
// 	// utils.CheckCreateDir(gconfig.Store_fileinfo_local)
// 	//创建保存文件索引的文件夹
// 	// utils.CheckCreateDir(gconfig.Store_fileinfo_net)
// 	//创建保存文件索引的文件夹
// 	// utils.CheckCreateDir(gconfig.Store_fileinfo_cache)
// 	//创建临时文件夹
// 	utils.CheckCreateDir(gconfig.Store_temp)
// 	//创建带扩展的完整文件存放文件夹
// 	utils.CheckCreateDir(gconfig.Store_files)

// 	//链接数据库
// 	sqlite3_db.Init()

// 	//initTask()

// 	fs.Init()
// 	fs.LoadAllSpace()
// 	// TimeSyncChunk()
// 	// initTask()

// 	//加载自己上传的文件索引，加载本地维护的文件块，定时同步
// 	// loadFileindex()

// 	//加载自己共享的文件
// 	//	err := LoadFileInfoSelf()
// 	//	if err != nil {
// 	//		return err
// 	//	}
// 	//加载本地文件索引
// 	//	err = LoadFileInfoLocal()
// 	//	if err != nil {
// 	//		return err
// 	//	}
// 	//加载网络文件索引
// 	//	err = LoadFileInfoNet()
// 	//	if err != nil {
// 	//		return err
// 	//	}

// 	// message_center.Register_vnode_search(config.MSGID_store_addFileShare, AddFileShare)         //添加一个文件共享
// 	// message_center.Register_p2p(config.MSGID_store_addFileShare_recv, AddFileShare_recv)        //添加一个文件共享 返回
// 	// message_center.Register_vnode_search(config.MSGID_store_findFileinfo, FindFileindexHandler) //网络中查找一个文件信息
// 	// message_center.Register_p2p(config.MSGID_store_findFileinfo_recv, FindFileindex_recv)       //网络中查找一个文件信息 返回
// 	// //	message_center.Register_p2p(config.MSGID_store_getFilesize, FindFilesize)
// 	// //	message_center.Register_p2p(config.MSGID_store_getFilesize_recv, FindFilesize_recv)
// 	// message_center.Register_p2p(config.MSGID_store_downloadFileChunk, DownloadFilechunk)           //网络中下载文件块
// 	// message_center.Register_p2p(config.MSGID_store_downloadFileChunk_recv, DownloadFilechunk_recv) //网络中下载文件块 返回
// 	// //	message_center.Register_p2p(config.MSGID_store_getUploadinfo, Uploadinfo)
// 	// //	message_center.Register_p2p(config.MSGID_store_getUploadinfo_recv, Uploadinfo_recv)
// 	// message_center.Register_vnode_search(config.MSGID_store_syncFileInfo, syncFileInfo)        //返回同步文件信息到1/4节点
// 	// message_center.Register_p2p(config.MSGID_store_syncFileInfo_recv, syncFileInfo_recv)       //返回同步文件信息到1/4节点
// 	// message_center.Register_p2p(config.MSGID_store_getfourNodeinfo, GetfourNodeinfo)           //返回节点上传地址信息
// 	// message_center.Register_p2p(config.MSGID_store_getfourNodeinfo_recv, GetfourNodeinfo_recv) //返回节点上传地址信息 返回
// 	// //验证空间大小
// 	// message_center.Register_p2p(config.MSGID_store_getFileindexList, CheckSpaceInfo)           //获取节点的文件列表，同时可以验证空间大小
// 	// message_center.Register_p2p(config.MSGID_store_getFileindexList_recv, CheckSpaceInfo_recv) //获取节点的文件列表，同时可以验证空间大小 返回
// 	// //engine.RegisterMsg(MSGID_getNodeWalletReceiptAddress, GetWalletAddr)
// 	// //engine.RegisterMsg(MSGID_getNodeWalletReceiptAddress_recv, GetWalletAddr_recv)
// 	// message_center.Register_vnode_search(config.MSGID_store_addFileOwner, AddFileOwner)  //添加一个文件拥有者
// 	// message_center.Register_p2p(config.MSGID_store_addFileOwner_recv, AddFileOwner_recv) //添加一个文件拥有者 返回

// 	// go client.Start()
// 	// go server.Start()
// 	return nil
// }
