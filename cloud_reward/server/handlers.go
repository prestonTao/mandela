package server

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/utils"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
	启动这个模块
*/
func Start() {
	// fmt.Println("1111111111111111111111111")
	message_center.Register_p2pHE(config.MSGID_store_online_heartbeat, GetStorePeerAddrCoin)
	// fmt.Println("2222222222222222222222222")

	// go RewardTiker()
	utils.Go(RewardTiker)

}

/*
	收到节点收款地址消息
*/
func GetStorePeerAddrCoin(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// fmt.Println("")

	sp, err := ParseStorePeer(message.Body.Content)
	// sp := &StorePeer{}
	// err := json.Unmarshal(*message.Body.Content, sp)
	if err != nil {
		return
	}
	AddStorePeer(*message.Head.Sender, sp)
	spaceTotalAddr, spaceTotal := CountStoreTotal()

	bs := make([]byte, 0, 16)
	bs = append(bs, utils.Uint64ToBytes(spaceTotalAddr)...)
	bs = append(bs, utils.Uint64ToBytes(spaceTotal)...)

	//回复给发送者
	message_center.SendP2pReplyMsg(message, config.MSGID_store_online_heartbeat_recv, &bs)
}
