package config

const (
	TSK_build_temp_name        = "TSK_build_temp_name"        //定时查询临时域名
	TSK_build_temp_name_remove = "TSK_build_temp_name_remove" //删除一个临时域名，域名注册失败，过期

	TSK_find_name = "TSK_find_name" //定时查询域名
	//	TSK_build_temp_name_remove = "TSK_build_temp_name_remove" //删除一个临时域名，域名注册失败，过期
	TSK_name_sync_multicast = "TSK_name_sync_multicast" //定时广播需要同步的域名
	TSK_key_sync_multicast  = "TSK_key_sync_multicast"  //定时广播需要同步的公钥

	TSK_msg_timeout_remove = "TSK_msg_timeout_remove" //消息超时删除
)

const (
	Time_name_sync_multicast = 10 //广播需要同步的域名间隔时间 单位：秒
	Time_key_sync_multicast  = 10 //广播需要同步的公钥间隔时间 单位：秒
	Time_Multicast_online    = 60 //广播节点上线间隔时间 单位：秒
	Time_register_addr_name  = 10 //循环注册域名的地址间隔时间 单位：秒
	Time_find_name_self      = 20 //循环查找自己的域名间隔时间 单位：秒
	Time_getNear_super_ip    = 10 //循环获取相邻节点的超级节点ip地址 单位：秒
)

var (
	Time_find_network_peer = []int64{1, 1, 1, 1, 1, 1, 1, 1, 60 * 10} //查询逻辑节点间隔时间 单位：秒
)
