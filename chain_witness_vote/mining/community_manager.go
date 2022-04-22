package mining

import (
	"mandela/config"
	"mandela/core/utils/crypto"
	"time"
)

const communityRewardCountUnconfirmedHeight = uint64(20) //设置区块未确认高度
var communityRewardHeightStart = uint64(0)               //保存本节点多少高度之前的奖励是分配了的

func init() {
	// go loopSelectTx()
}
func loopSelectTx() {
	for {
		//判断区块是否同步完成，如果没有同步完成则不要操作
		chain := GetLongChain()
		if chain == nil {
			time.Sleep(time.Minute * 10)
			continue
		}
		if !chain.SyncBlockFinish {
			time.Sleep(time.Minute * 10)
			continue
		}
		time.Sleep(time.Minute * 10)
	}
}

/*

 */
func BuildCommunityAddrStartHeight(addr crypto.AddressCoin) []byte {
	return append([]byte(config.DB_community_addr), addr...)
}
