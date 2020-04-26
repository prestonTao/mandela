package im

import (
	"mandela/config"
	"mandela/core/message_center"
	"mandela/core/utils"
	"path/filepath"
)

func RegisterIM() {
	message_center.Register_p2pHE(config.MSGID_im_file, FileMsg)                   //接收图片消息
	message_center.Register_p2pHE(config.MSGID_im_file_recv, FileMsg_recv)         //接收图片消息返回
	message_center.Register_p2pHE(config.MSGID_im_property, PropertyMsg)           //接收用户属性消息
	message_center.Register_p2pHE(config.MSGID_im_property_recv, PropertyMsg_recv) //接收用户属性消息返回
	message_center.Register_p2pHE(config.MSGID_im_addr, BaseCoinAddrMsg)           //获取用户收款地址消息
	message_center.Register_p2pHE(config.MSGID_im_addr_recv, BaseCoinAddrMsg_recv) //获取用户收款地址消息返回
	message_center.Register_p2pHE(config.MSGID_im_pay, PayMsg)                     //付款消息
	message_center.Register_p2pHE(config.MSGID_im_pay_recv, PayMsg_recv)           //付款消息返回
	//创建保存文件索引的文件夹
	utils.CheckCreateDir(filepath.Join(imfilepath))
}
