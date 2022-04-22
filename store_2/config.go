package store

const (
	dbpath                     = "db/space"
	spacenum                   = 1024 * perspace        //总占用空间 10G  单位byte
	perspace                   = 10 * 1024 * 1024       //每块10M 单位byte
	prefix                     = "KEY_"                 //默认数据前缀
	Time_sharefile             = 600                    //共享文件添加索引间隔时间 单位：秒
	Time_shareUserOfflineClear = Time_sharefile*120 + 1 //共享的用户超过间隔时间（Time_sharefile）20小时后删除这个共享用户
	Time_loopClearUser         = 60 * 60 * 24           //定时清理文件索引，文件索引中超过60天没有用户共享的块删除掉

	//	Chunk_size   = 1024 * 1024 * 8  //1024 * 1024 * 8 //文件分块大小，默认8M
	Chunk_size    = perspace         //文件分块大小，默认10Mb
	UploadScheme  = "http"           //文件上传协议
	UploadPath    = "/store/addfile" //文件HTTP上传地址
	UploadField   = "files[]"        //文件HTTP上传表单名
	Time_fileuser = 60 * 24 * 3600   //文件所有者最大离线时间，超过则删除当前所有者 单位：秒
	Renum         = 5                //下载失败后重试次数
	Suffix        = ".crypt"         //加密文件后缀名
	fileprikname  = "filecrypt.key"  //文件加解密私钥
	DepositMin    = 100000000        //最少押金 1个币
)
