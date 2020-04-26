package config

const (
	CLASS_get_MachineID           = "CLASS_get_MachineID"           //获取节点的机器id
	CLASS_security_searchAddr     = "CLASS_security_searchAddr"     //加密通信搜索节点
	CLASS_reward_hold             = "CLASS_reward_hold"             //存储节点发送收款地址领取奖励消息
	CLASS_im_addfriend            = "CLASS_im_addfriend"            //添加好友
	CLASS_im_msg_come             = "CLASS_im_msg_come"             //消息到达
	CLASS_im_file_msg             = "CLASS_im_file_msg"             // 图片文件消息到达
	CLASS_im_property_msg         = "CLASS_im_property_msg"         // 用户属性消息到达
	CLASS_im_addr_msg             = "CLASS_im_addr_msg"             // 获取用户收款地址消息到达
	CLASS_im_pay_msg              = "CLASS_im_pay_msg"              // 文本消息(付款)到达
	CLASS_im_security_create_pipe = "CLASS_im_security_create_pipe" //创建加密通道消息

	CLASS_wallet_broadcast_return   = "CLASS_wallet_broadcast_return"   //广播消息回复
	CLASS_wallet_getblockforwitness = "CLASS_wallet_getblockforwitness" //通过见证人同步区块
)
