package config

const (
	//---------------- base --------------------------
	//	MSGID_Text = 101 //显示文本消息

	MSGID_search_node           = 135 //搜索一个节点地址是否在线
	MSGID_search_node_recv      = 136 //搜索一个节点地址是否在线_返回
	MSGID_checkNodeOnline       = 110 //检查节点是否在线
	MSGID_checkNodeOnline_recv  = 111 //检查节点是否在线_返回
	MSGID_TextMsg               = 112 //接收文本消息
	MSGID_getNearSuperIP        = 113 //从邻居节点得到自己的逻辑节点
	MSGID_getNearSuperIP_recv   = 114 //从邻居节点得到自己的逻辑节点_返回
	MSGID_multicast_online_recv = 122 //接收节点上线广播
	MSGID_ask_close_conn_recv   = 128 //询问关闭连接
	MSGID_TextMsg_recv          = 129 //接收消息返回消息

	//---------------- 可靠传输加密通道协议 --------------------------
	MSGID_SearchAddr                = 130 //搜索一个节点，获取节点地址和身份公钥
	MSGID_SearchAddr_recv           = 131 //搜索一个节点，获取节点地址和身份公钥_返回
	MSGID_security_create_pipe      = 132 //发送密钥消息，对方建立通道
	MSGID_security_create_pipe_recv = 133 //发送密钥消息，对方建立通道_返回
	MSGID_security_pipe_error       = 134 //解密错误

	//---------------- name 域名模块 --------------------------
	MSGID_register_name        = 102 //注册一个域名
	MSGID_register_name_recv   = 103 //注册一个域名_返回
	MSGID_build_name           = 104 //构建一个域名
	MSGID_build_name_recv      = 105 //构建一个域名_返回
	MSGID_check_temp_name      = 106 //检查刚构建的域名是否成功
	MSGID_check_temp_name_recv = 107 //检查刚构建的域名是否成功 返回
	//	MSGID_findDomain               = 108 //查找这个域名是否存在
	//	MSGID_recv_domain              = 109 //返回这个域名是否存在
	MSGID_find_name                = 115 //查找一个域名的地址
	MSGID_find_name_recv           = 116 //查找一个域名的地址 返回
	MSGID_name_add_address_recv    = 117 //收到域名添加新地址
	MSGID_name_sync_multicast_recv = 118 //接收需要同步的域名广播

	MSGID_ROOT_register_name    = 123 //申请注册一个域名
	MSGID_ROOT_RECV_create_name = 124 //创建一个域名

	//---------------- 公钥 模块 --------------------------
	MSGID_key_sync_multicast_recv = 119 //接收需要同步的公钥广播
	MSGID_key_find_keyname        = 120 //查找公钥对应的域名
	MSGID_key_find_keyname_recv   = 121 //查找公钥对应的域名 返回
	MSGID_ROOT_RECV_save_key_name = 125 //接收保存的公钥key对应的域名

	//---------------- web 模块 --------------------------
	MSGID_http_request  = 126 //http请求
	MSGID_http_response = 127 //http返回
	// MSGID_http_getwebinfo      = 128 //获取节点web信息
	// MSGID_http_getwebinfo_recv = 129 //获取节点web信息 返回

	//---------------- wallet 模块 --------------------------
	MSGID_multicast_vote_recv      = 200 //接收见证人投票广播
	MSGID_multicast_blockhead      = 201 //接收区块头广播
	MSGID_heightBlock              = 202 //查询邻居节点区块高度
	MSGID_heightBlock_recv         = 203 //查询邻居节点区块高度_返回
	MSGID_getBlockHead             = 204 //查询邻居节点的起始区块头
	MSGID_getBlockHead_recv        = 205 //查询邻居节点的起始区块头_返回
	MSGID_getTransaction           = 206 //查询交易
	MSGID_getTransaction_recv      = 207 //查询交易_返回
	MSGID_multicast_transaction    = 208 //接收交易广播
	MSGID_getUnconfirmedBlock      = 209 //从邻居节点获取未确认的区块
	MSGID_getUnconfirmedBlock_recv = 210 //从邻居节点获取未确认的区块_返回

	MSGID_multicast_return        = 211 //收到广播消息回复
	MSGID_getblockforwitness      = 212 //从邻居节点获取指定见证人的区块
	MSGID_getblockforwitness_recv = 213 //从邻居节点获取指定见证人的区块_返回

	//---------------- sharebox 模块 --------------------------
	MSGID_sharebox_addFileShare                     = 300 //添加一个文件共享
	MSGID_sharebox_addFileShare_recv                = 301 //添加一个文件共享 返回
	MSGID_sharebox_findFileinfo                     = 302 //网络中查找一个文件信息
	MSGID_sharebox_findFileinfo_recv                = 303 //网络中查找一个文件信息 返回
	MSGID_sharebox_getFilesize                      = 304 //网络中查找一个文件信息
	MSGID_sharebox_getFilesize_recv                 = 305 //网络中查找一个文件信息 返回
	MSGID_sharebox_downloadFileChunk                = 306 //网络中下载文件块
	MSGID_sharebox_downloadFileChunk_recv           = 307 //网络中下载文件块 返回
	MSGID_sharebox_getUploadinfo                    = 308 //返回节点上传地址信息
	MSGID_sharebox_getUploadinfo_recv               = 309 //返回节点上传地址信息 返回
	MSGID_sharebox_getNodeWalletReceiptAddress      = 315 //查询节点收款地址
	MSGID_sharebox_getNodeWalletReceiptAddress_recv = 316 //查询节点上传地址信息 返回
	MSGID_sharebox_getsharefolderlist               = 317 //获取节点共享文件列表
	MSGID_sharebox_getsharefolderlist_recv          = 318 //获取节点共享文件列表 返回

	//---------------- store 模块 --------------------------
	MSGID_store_addFileShare      = 400 //添加一个文件共享
	MSGID_store_addFileShare_recv = 401 //添加一个文件共享 返回
	MSGID_store_findFileinfo      = 402 //网络中查找一个文件信息
	MSGID_store_findFileinfo_recv = 403 //网络中查找一个文件信息 返回
	//	MSGID_store_getFilesize            = 404 //网络中查找一个文件信息
	//	MSGID_store_getFilesize_recv       = 405 //网络中查找一个文件信息 返回
	MSGID_store_downloadFileChunk      = 406 //网络中下载文件块
	MSGID_store_downloadFileChunk_recv = 407 //网络中下载文件块 返回
	//	MSGID_store_getUploadinfo                    = 408 //返回节点上传地址信息
	//	MSGID_store_getUploadinfo_recv               = 409 //返回节点上传地址信息 返回
	MSGID_store_syncFileInfo                     = 410 //返回同步文件信息到1/4节点
	MSGID_store_syncFileInfo_recv                = 412 //返回同步文件信息到1/4节点
	MSGID_store_getfourNodeinfo                  = 413 //返回节点上传地址信息
	MSGID_store_getfourNodeinfo_recv             = 414 //返回节点上传地址信息 返回
	MSGID_store_getNodeWalletReceiptAddress      = 415 //查询节点收款地址
	MSGID_store_getNodeWalletReceiptAddress_recv = 416 //查询节点上传地址信息 返回

	MSGID_store_online_heartbeat      = 417 //在线心跳，发送自己的收款地址
	MSGID_store_online_heartbeat_recv = 418 //在线心跳，发送自己的收款地址 返回
	MSGID_store_getFileindexList      = 419 //获取节点的文件列表，同时可以验证空间大小
	MSGID_store_getFileindexList_recv = 420 //获取节点的文件列表，同时可以验证空间大小 返回
	MSGID_store_addFileOwner          = 421 //添加文件拥有者
	MSGID_store_addFileOwner_recv     = 422 //添加文件拥有者 返回

	//---------------- IM 模块 --------------------------
	MSGID_im_addfriend        = 500  //添加好友
	MSGID_im_addfriend_recv   = 501  //添加好友 返回
	MSGID_im_agreefriend      = 502  //同意或拒绝添加好友
	MSGID_im_agreefriend_recv = 503  //同意或拒绝添加好友 返回
	MSGID_im_file             = 504  //文件传输（图片、文件）
	MSGID_im_file_recv        = 505  //文件传输（图片、文件）返回
	MSGID_im_property         = 506  //修改用户属性同步
	MSGID_im_property_recv    = 507  //修改用户属性同步 返回
	MSGID_im_addr             = 508  //获取用户的收款地址
	MSGID_im_addr_recv        = 509  //获取用户的收款地址 返回
	MSGID_im_pay              = 5010 //文本消息(付款)
	MSGID_im_pay_recv         = 5011 //付文本消息(付款) 返回

	//---------------- Vnode 模块 --------------------------

	MSGID_vnode_getstate            = 600 //查询一个节点是否开通了虚拟节点服务
	MSGID_vnode_getstate_recv       = 601 //查询一个节点是否开通了虚拟节点服务_返回
	MSGID_vnode_getNearSuperIP      = 602 //从邻居节点得到自己的逻辑节点
	MSGID_vnode_getNearSuperIP_recv = 603 //从邻居节点得到自己的逻辑节点_返回

)

// const (
// 	//------- 本模块编号范围 2000 - 2999 ------------
// 	// MSGID_SearchAddr            = 2000 //获取节点地址和身份公钥
// 	// MSGID_SearchAddr_recv       = 2001 //获取节点地址和身份公钥_返回
// 	MSGID_TextMsg               = 112 //接收文本消息
// 	MSGID_getNearSuperIP        = 113 //从邻居节点得到自己的逻辑节点
// 	MSGID_getNearSuperIP_recv   = 114 //从邻居节点得到自己的逻辑节点_返回
// 	MSGID_multicast_online_recv = 122 //接收节点上线广播
// 	MSGID_ask_close_conn_recv   = 128 //询问关闭连接
// )
