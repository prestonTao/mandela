package message_center

const (
	MSG_WAIT_http_request  = "MSG_WAIT_http_request"  //
	CLASS_findfileinfo     = "CLASS_findfileinfo"     //查找文件信息
	CLASS_downloadfile     = "CLASS_downloadfile"     //下载文件块
	CLASS_sharefile        = "CLASS_sharefile"        //共享文件
	CLASS_syncfileinfo     = "CLASS_syncfileinfo"     //文件块分布式存储
	CLASS_getfourNodeinfo  = "CLASS_getfourNodeinfo"  //获取1/4节点信息(APP用)
	CLASS_safemsginfo      = "CLASS_safemsginfo"      //消息安全协议
	CLASS_getfileindexlist = "CLASS_getfileindexlist" //验证空间大小

	CLASS_uploadinfo    = "CLASS_uploadinfo"   //文件http上传信息
	CLASS_syncdata      = "CLASS_syncdata"     //分布式数据同步
	CLASS_raftvote      = "CLASS_raftvote"     //raft发起投票
	CLASS_raftvoteheart = "CLASS_raftvotehear" //raft发起心跳

	CLASS_findHeightBlock     = "CLASS_findHeightBlock"     //查询区块高度
	CLASS_getBlockHead        = "CLASS_getBlockHead"        //获取区块头
	CLASS_getTransaction      = "CLASS_getTransaction"      //获取交易
	CLASS_getUnconfirmedBlock = "CLASS_getUnconfirmedBlock" //获取未确认区块
	CLASS_getBlockLastCurrent = "CLASS_getBlockLastCurrent" //获取已确认高度最高区块

	CLASS_getWalletAddr = "CLASS_getWalletAddr" //获取节点钱包地址

	CLASS_getRemoteFolderList = "CLASS_getRemoteFolderList" //获取远端节点共享文件夹列表
	// CLASS_http_getwebinfo     = "CLASS_http_getwebinfo"     //获取远端节点web端口信息

	CLASS_vnode_getstate = "CLASS_vnode_getstate" //获取节点的虚拟节点开通状态

)
