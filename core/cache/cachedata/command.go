package cachedata

import (
	"bytes"
	"crypto/sha256"
	"errors"

	// "fmt"
	mc "mandela/core/message_center"
	"mandela/core/nodeStore"
	"mandela/core/utils"
)

//获取1/4节点id
func getQuarterLogicIds(id *utils.Multihash) []*utils.Multihash {
	return nodeStore.GetQuarterLogicIds(id)
}

//同步数据
func SyncDataToQuarterLogicIds(khash *utils.Multihash) error {
	cachedata := GetCacheDataByHash(khash)
	if cachedata == nil {
		return errors.New("no data")
	}
	ids := getQuarterLogicIds(cachedata.Key)
	for _, idpt := range ids {
		//fmt.Println(idpt.B58String())
		sendMsg(idpt, cachedata.Json())
	}
	return nil
}

//同步数据(主要删除用)
func SyncCacheDataToQuarterLogicIds(cachedata *CacheData) error {
	if cachedata == nil {
		return errors.New("no data")
	}
	ids := getQuarterLogicIds(cachedata.Key)
	for _, idpt := range ids {
		//fmt.Println(idpt.B58String())
		sendMsg(idpt, cachedata.Json())
	}
	return nil
}

//广播数据消息
func sendMsg(id *utils.Multihash, data []byte) error {
	mhead := mc.NewMessageHead(id, id, false)
	mbody := mc.NewMessageBody(&data, "", nil, 0)
	message := mc.NewMessage(mhead, mbody)
	if message.Send(MSGID_syncData) {
		//fmt.Println("数据发送成功", id.B58String())
		//		bs := mc.WaitRequest(mc.CLASS_syncdata, message.Body.Hash.B58String())
		//		fmt.Println("有消息返回", string(*bs))
		//		if bs == nil {
		//			fmt.Println("同步数据消息失败，可能超时")
		//			return errors.New("同步数据消息失败，可能超时")
		//		}
		return nil
	}
	//fmt.Println("数据发送失败", id.B58String())
	return nil
}

//生成hash
func buildHash(key []byte) *utils.Multihash {
	hash := sha256.Sum256(key)
	bs, err := utils.Encode(hash[:], Version)
	if err != nil {
		// fmt.Println("buildhash:", err)
		return nil
	}
	has := utils.Multihash(bs)
	return &has
}

//检查当前节点是否是共享节点
func checkOwn(id *utils.Multihash, ownid []*utils.Multihash) bool {
	for _, v := range ownid {
		if bytes.Equal(id, v) {
			return true
		}
	}
	return false
}
