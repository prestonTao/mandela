package config

import (
	// "mandela/core/utils"
	"path/filepath"
)

const (
	Time_sharefile             = 120                   //共享文件添加索引间隔时间 单位：秒
	Time_shareUserOfflineClear = Time_sharefile*15 + 1 //共享的用户超过间隔时间（Time_sharefile）3倍后删除这个共享用户
	Time_loopClearUser         = 60 * 60 * 24          //定时清理文件索引，文件索引中超过60天没有用户共享的块删除掉

	Chunk_size   = 1024 * 1024 * 8  //1024 * 1024 * 8 //文件分块大小，默认8M
	UploadScheme = "http"           //文件上传协议
	UploadPath   = "/store/addfile" //文件HTTP上传地址
	UploadField  = "files[]"        //文件HTTP上传表单名
	Suffix       = ".crypt"         //加密文件后缀名
	// StoreDBPath = "" //
)

const (
	Store_path_dir      = "store"    //本地共享文件存储目录名称
	Store_path_fileinfo = "fileinfo" //保存本节点维护的文件索引
	Store_path_cache    = "cache"    //保存
	Store_path_temp     = "temp"     //临时文件夹，本地上传存放目录，存放未切片的完整文件

	Store_path_fileinfo_self  = "self"  //自己上传的文件索引存储目录名称
	Store_path_fileinfo_local = "local" //本地下载过的文件索引存储目录名称
	Store_path_fileinfo_net   = "net"   //网络需要保存的文件索引存储目录名称
	Store_path_fileinfo_cache = "cache" //缓存中保存的文件索引存储目录名称
	IsRemoveStore             = false   //启动时删除本地所有文件分片及分片索引
	//	IsCreateId                = true    //启动时是否要创建新的id

)

var (
	Store_dir            string = filepath.Join(Store_path_dir)                            //本地共享文件存储目录路径
	Store_fileinfo_self  string = filepath.Join(Store_path_dir, Store_path_fileinfo_self)  //自己上传的文件索引存储目录路径
	Store_fileinfo_local string = filepath.Join(Store_path_dir, Store_path_fileinfo_local) //本地下载过的文件索引存储目录路径
	Store_fileinfo_net   string = filepath.Join(Store_path_dir, Store_path_fileinfo_net)   //网络需要保存的文件索引存储目录路径
	Store_fileinfo_cache string = filepath.Join(Store_path_dir, Store_path_fileinfo_cache) //缓存中保存的文件索引存储目录路径
	Store_cache          string = filepath.Join(Store_path_dir, Store_path_temp)           //临时文件夹，本地上传存放目录，存放未切片的完整文件
	Store_temp           string = filepath.Join(Store_path_dir, Store_path_temp)           //临时文件夹，本地上传存放目录，存放未切片的完整文件
)

var (
	DB_Path         = filepath.Join(Store_path_dir, "data") //
	DB_ShareFolders = []byte("SHAREFOLDERS")                //共享文件夹列表
)
