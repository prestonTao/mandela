package sharebox

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/utils"
)

/*
	获取本节点钱包中收款地址
*/
func GetWalletAddr(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// fmt.Println("获取本节点钱包中收款地址")

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err = message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }

	addr := keystore.GetCoinbase()
	// if err != nil {
	// 	// fmt.Println("获取矿工地址错误", err)
	// 	return
	// }
	// bs := make([]byte, 0)
	// bs = append(bs, addr)
	bs := []byte(addr.Addr)

	//回复给发送者
	mc.SendP2pReplyMsg(message, config.MSGID_sharebox_getNodeWalletReceiptAddress_recv, &bs)
	// mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	// mbody := mc.NewMessageBody(&bs, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	// message = mc.NewMessage(mhead, mbody)
	// message.Reply(MSGID_getNodeWalletReceiptAddress_recv)
}

/*
	获取本节点钱包中收款地址 返回
*/
func GetWalletAddr_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
	// fmt.Println("获取本节点钱包中收款地址 返回")

	// message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	// if err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// // form, _ := utils.FromB58String(msg.Session.GetName())
	// form := nodeStore.AddressFromB58String(msg.Session.GetName())
	// if message.IsSendOther(&form) {
	// 	return
	// }
	// //发送给自己的，自己处理
	// if err := message.ParserContent(); err != nil {
	// 	// fmt.Println(err)
	// 	return
	// }
	// flood.ResponseWait(mc.CLASS_getWalletAddr, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(mc.CLASS_getWalletAddr, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}
