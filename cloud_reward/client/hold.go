package client

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining/name"
	"mandela/cloud_reward"
	"mandela/cloud_reward/server"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/virtual_node"
	"time"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	holdTimeInterval = cloud_reward.HoldTimeInterval //time.Second * 60 * 60 //发送心跳间隔时间，单位：秒
)

/*
	定时给超级节点发送心跳
*/
func hold() {
	//启动和链接leveldb数据库
	err := db.InitDB(config.DB_path, config.DB_path_temp)
	if err != nil {
		panic(err)
	}
	SendOnlineHeartBeat()
	for range time.NewTicker(holdTimeInterval).C {
		SendOnlineHeartBeat()
	}
}

/*
	发送心跳
*/
func SendOnlineHeartBeat() {
	// engine.Log.Debug("开始发送心跳")
	nets := GetCloudPeer()
	if nets == nil {
		// fmt.Println("没有存储节点网络地址")
		return
	}
	// fmt.Println("有存储节点网络地址")
	for _, one := range nets {
		for i := 0; i < 5; i++ {
			addrCoin := keystore.GetCoinbase()
			// bs := []byte(addrCoin.Addr)

			vnodeinfos := virtual_node.GetVnodeNumber()

			sp := server.StorePeer{
				// AddrNet   nodeStore.AddressNet //存储节点
				SpaceNum: uint64(len(vnodeinfos)), //存储空间单位个数
				AddrCoin: addrCoin.Addr,           //收款地址
				// FlashTime int64                //刷新时间
			}

			bs, err := sp.Proto() //json.Marshal(sp)
			if err != nil {
				engine.Log.Error("发送存储空间心跳失败 %s", err.Error())
				continue
			}

			message, ok, isSelf := message_center.SendP2pMsgHE(config.MSGID_store_online_heartbeat, &one, bs)
			if !ok || isSelf {
				// fmt.Println("重新发送心跳1")
				continue
			}
			recvBS, _ := flood.WaitRequest(config.CLASS_reward_hold, utils.Bytes2string(message.Body.Hash), 0)
			if recvBS == nil {
				// fmt.Println("重新发送心跳2")
				continue
			} else {
				//解析出返回参数
				spaceTotalAddr := utils.BytesToUint64((*recvBS)[:8])
				spaceTotal := utils.BytesToUint64((*recvBS)[8:])
				config.SetSpaceTotalAddr(spaceTotalAddr)
				config.SetSpaceTotal(spaceTotal)
			}
			break
		}
	}

}

/*
	获得云存储空投节点地址
*/
func GetCloudPeer() []nodeStore.AddressNet {
	nameinfo := name.FindNameToNet(config.Name_store)
	if nameinfo != nil {
		return nameinfo.NetIds
	}
	return nil
}
