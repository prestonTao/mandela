package server

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/utils/crypto"
)

/*
	启动这个模块
*/
func Start() {
	// fmt.Println("1111111111111111111111111")
	message_center.Register_p2pHE(config.MSGID_store_online_heartbeat, GetStorePeerAddrCoin)
	// fmt.Println("2222222222222222222222222")

	go RewardTiker()

}

/*
	收到节点收款地址消息
*/
func GetStorePeerAddrCoin(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// fmt.Println("")
	addrNet := crypto.AddressCoin(*message.Body.Content)

	AddStorePeer(*message.Head.Sender, addrNet)

	//回复给发送者
	message_center.SendP2pReplyMsg(message, config.MSGID_store_online_heartbeat_recv, nil)

}
