package client

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
)

/*
	启动这个模块
*/
func Start() {
	message_center.Register_p2pHE(config.MSGID_store_online_heartbeat_recv, GetStorePeerAddrCoin_recv)
	hold()
}

/*
	收到节点收款地址消息 返回
*/
func GetStorePeerAddrCoin_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	bs := []byte("ok")
	flood.ResponseWait(config.CLASS_reward_hold, message.Body.Hash.B58String(), &bs)
}
