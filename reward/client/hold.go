package client

import (
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"time"
)

var (
	holdTimeInterval = time.Second * 5 //发送心跳间隔时间，单位：秒
)

/*
	定时给超级节点发送心跳
*/
func hold() {
	sendOnlineHeartBeat()
	for range time.NewTicker(holdTimeInterval).C {
		sendOnlineHeartBeat()
	}
}

/*
	发送心跳
*/
func sendOnlineHeartBeat() {
	// fmt.Println("开始发送心跳")
	nets := GetCloudPeer()
	if nets == nil {
		// fmt.Println("没有存储节点网络地址")
		return
	}
	// fmt.Println("有存储节点网络地址")
	for _, one := range nets {
		for i := 0; i < 5; i++ {
			addrCoin := keystore.GetCoinbase()
			bs := []byte(addrCoin)
			message, ok := message_center.SendP2pMsgHE(config.MSGID_store_online_heartbeat, &one, &bs)
			if !ok {
				// fmt.Println("重新发送心跳1")
				continue
			}
			recvBS := flood.WaitRequest(config.CLASS_reward_hold, message.Body.Hash.B58String())
			if recvBS == nil {
				// fmt.Println("重新发送心跳2")
				continue
			}
			break
		}
	}

}

/*
	获得云存储空投节点地址
*/
func GetCloudPeer() []nodeStore.AddressNet {
	// nameinfo := name.FindName(config.Name_store)
	// if nameinfo != nil {
	// 	return nameinfo.NetIds
	// }

	nameinfo := name.FindNameToNet(config.Name_store)
	if nameinfo != nil {
		return nameinfo.NetIds
	}
	return nil
}
