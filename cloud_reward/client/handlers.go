package client

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/utils"
)

/*
	启动这个模块
*/
func Start() {
	message_center.Register_p2pHE(config.MSGID_store_online_heartbeat_recv, GetStorePeerAddrCoin_recv)
	hold()
}

/*
	收到节点心跳消息 返回
*/
func GetStorePeerAddrCoin_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// bs := []byte("ok")
	flood.ResponseWait(config.CLASS_reward_hold, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}
