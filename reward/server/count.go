package server

import (
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils/crypto"
	"bytes"
	"fmt"
	"sync"
	"time"
)

var timeoutInterval = int64(60 * 60) //节点超时时间，单位：秒
var countStorePeer = new(sync.Map)   //key:string=节点地址;value:*StorePeer=存储节点信息;

func init() {
	go cleanStorePeer()
}

func cleanStorePeer() {
	for range time.NewTicker(time.Hour).C {
		now := time.Now().Unix()
		removeAddr := make([]string, 0)
		countStorePeer.Range(func(k, v interface{}) bool {
			storePeer := v.(*StorePeer)
			if storePeer.FlashTime+timeoutInterval < now {
				//
				key := k.(string)
				removeAddr = append(removeAddr, key)
			}
			return true
		})
		for _, one := range removeAddr {
			countStorePeer.Delete(one)
		}
	}
}

/*
	存储奖励定时任务
*/
func RewardTiker() {

	for {
		str := time.Now().Format("2006-01-02 15:04:05")
		newTime, _ := time.Parse("2006-01-02 15:04:05", str[:10]+" 00:00:00")
		zoreHour := newTime.Add(time.Hour * 24) //第二天凌晨
		fmt.Println("现在离明天凌晨间隔时间是：", zoreHour.Unix())
		// time.Sleep(time.Second * time.Duration(zoreHour.Unix()))
		time.Sleep(time.Second * 60)
		RewardStorePeers()
	}

}

/*
	奖励给存储节点
*/
func RewardStorePeers() {
	// fmt.Println("==============开发发放奖励")
	//获取存储超级节点地址

	nameinfo := name.FindName(config.Name_store)
	if nameinfo == nil {
		// fmt.Println("域名不存在")
		return
	}
	// nets := client.GetCloudPeer()
	// if nets == nil {
	// 	return
	// }
	//判断自己是否在超级节点地址里
	have := false
	for _, one := range nameinfo.NetIds {
		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, one) {
			have = true
			break
		}
	}
	//没有在列表里，则退出
	if !have {
		// fmt.Println("自己不在超级节点地址里")
		return
	}

	//先统计今天一天的总奖励
	//直接统计自己的余额
	rewardTotal := mining.GetBalance()
	if rewardTotal <= 0 {
		// fmt.Println("获取奖励的账户余额为 0")
		return
	}
	gas := uint64(0)
	pwd := "123456789"

	peers := CountStorePeers()
	if len(peers) <= 0 {
		// fmt.Println("存储节点数量为0")
		return
	}
	rewardOne := rewardTotal / uint64(len(peers))

	pns := make([]mining.PayNumber, 0)
	for _, one := range peers {
		pn := mining.PayNumber{
			Address: one.AddrCoin, //转账地址
			Amount:  rewardOne,    //转账金额
		}
		pns = append(pns, pn)
	}
	mining.SendToMoreAddress(pns, gas, pwd, "")

}

/*
	提供存储的节点信息
*/
type StorePeer struct {
	AddrNet   nodeStore.AddressNet //存储节点
	AddrCoin  crypto.AddressCoin   //收款地址
	FlashTime int64                //刷新时间
}

/*
	添加一个存储节点信息
*/
func AddStorePeer(addrNet nodeStore.AddressNet, addrCoin crypto.AddressCoin) {
	fmt.Println("---------------------\n添加存储节点 1111111111111111")
	value, ok := countStorePeer.Load(addrNet.B58String())
	if ok {
		fmt.Println("添加存储节点 2222222222222222")
		storePeer := value.(*StorePeer)
		storePeer.FlashTime = time.Now().Unix()
	} else {
		fmt.Println("添加存储节点 33333333333333333")
		storePeer := StorePeer{
			AddrNet:   addrNet,
			AddrCoin:  addrCoin,
			FlashTime: time.Now().Unix(),
		}
		countStorePeer.Store(addrNet.B58String(), &storePeer)
	}
}

/*
	统计节点，交给发放奖励用
*/
func CountStorePeers() []*StorePeer {
	peers := make([]*StorePeer, 0)
	countStorePeer.Range(func(k, v interface{}) bool {
		storePeer := v.(*StorePeer)
		// fmt.Println("查询存储节点", storePeer.AddrNet.B58String())
		fmt.Println(storePeer.FlashTime+timeoutInterval >= time.Now().Unix(), storePeer.FlashTime, timeoutInterval, time.Now().Unix())
		if storePeer.FlashTime+timeoutInterval >= time.Now().Unix() {
			//
			peers = append(peers, storePeer)
		}
		return true
	})
	return peers
}
