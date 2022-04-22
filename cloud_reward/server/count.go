package server

import (
	"bytes"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/name"
	"mandela/cloud_reward"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"math/big"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
)

// var timeoutInterval = cloud_reward.HoldTimeInterval * 2 //节点超时时间,2个心跳周期内，未收到消息，则判定超时下线，单位：秒
var countStorePeer = new(sync.Map) //key:string=节点地址;value:*StorePeer=存储节点信息;

func init() {
	// go cleanStorePeer()
	utils.Go(cleanStorePeer)
}

func cleanStorePeer() {
	for range time.NewTicker(cloud_reward.TimeoutInterval).C {
		now := time.Now().Unix()
		removeAddr := make([]string, 0)
		countStorePeer.Range(func(k, v interface{}) bool {
			storePeer := v.(*StorePeer)
			// engine.Log.Debug("判断是否超时 %s %s %s", time.Unix(storePeer.FlashTime, 0).Format("2006-01-02 15:04:05"),
			// 	time.Unix(storePeer.FlashTime+int64(timeoutInterval.Seconds()), 0).Format("2006-01-02 15:04:05"),
			// 	time.Now().Format("2006-01-02 15:04:05"))
			if storePeer.FlashTime+int64(cloud_reward.TimeoutInterval.Seconds()) < now {
				//
				key := k.(string)
				removeAddr = append(removeAddr, key)
				// engine.Log.Debug("超时 %s", key)
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
		// str := time.Now().Format("2006-01-02 15:04:05")
		// newTime, _ := time.ParseInLocation("2006-01-02 15:04:05", str[:10]+" 00:00:00", time.Local)
		// zoreHour := newTime.Add(time.Hour * 24) //第二天凌晨
		// engine.Log.Info("现在离明天凌晨间隔时间是：%d", zoreHour.Unix())
		// time.Sleep(time.Second * time.Duration(zoreHour.Unix()))
		time.Sleep(cloud_reward.RewardTimeInterval)
		RewardStorePeers()
	}

}

/*
	奖励给存储节点
*/
func RewardStorePeers() {
	engine.Log.Info("start spaces reward")

	//未同步完成则不分配奖励
	chain := mining.GetLongChain()
	if chain == nil {
		engine.Log.Warn("If the synchronization is not completed, no reward will be allocated")
		return
	}
	if !chain.SyncBlockFinish {
		// engine.Log.Warn("未同步完成则不分配奖励")
		engine.Log.Warn("If the synchronization is not completed, no reward will be allocated")
		return
	}

	//获取存储超级节点地址
	nameinfo := name.FindName(config.Name_store)
	if nameinfo == nil {
		//域名不存在
		engine.Log.Debug("Domain name does not exist")
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
		engine.Log.Debug("You are not in the super node address")
		return
	}

	//先统计今天一天的总奖励
	//直接统计自己的余额
	rewardTotal, _, _ := mining.FindBalanceValue()
	if rewardTotal <= 0 {
		engine.Log.Debug("The account balance for obtaining rewards is 0")
		return
	}
	// engine.Log.Info("start spaces reward 2")
	//计算分配比例
	rewardTotalBig := big.NewInt(int64(rewardTotal))
	num100000 := big.NewInt(100000)
	//0.07692平均分配
	average := new(big.Int).Mul(rewardTotalBig, big.NewInt(7692))
	average = new(big.Int).Div(average, num100000)
	//0.25641质押奖励
	pledgeReward := new(big.Int).Mul(rewardTotalBig, big.NewInt(25641))
	pledgeReward = new(big.Int).Div(pledgeReward, num100000)
	//0.66667存储空间数量加权分
	spacesNum := new(big.Int).Mul(rewardTotalBig, big.NewInt(66667))
	spacesNum = new(big.Int).Div(spacesNum, num100000)

	gas := uint64(0)
	// pwd := "123456789"
	// engine.Log.Info("start spaces reward 3")
	peers := CountStorePeers()
	if len(*peers) <= 0 {
		engine.Log.Debug("The number of storage nodes is 0")
		return
	}
	//计算空间总量
	spaceNumTotal := uint64(0)
	//统计存储节点数量
	peerTotal := 0
	for _, one := range *peers {
		if one.SpaceNum <= 0 {
			continue
		}
		spaceNumTotal = spaceNumTotal + one.SpaceNum
		peerTotal++
	}
	// rewardOne := rewardTotal / uint64(len(peers))
	if spaceNumTotal <= 0 {
		engine.Log.Debug("Storage space is 0")
		return
	}
	// engine.Log.Info("start spaces reward 4")
	height := mining.GetHighestBlock()
	//开始分配奖励
	payNum := make([]mining.PayNumber, 0)
	//平均分
	temp := new(big.Int).Div(average, big.NewInt(int64(peerTotal))).Uint64()
	for _, one := range *peers {
		if one.SpaceNum <= 0 {
			continue
		}
		pns := mining.LinearRelease0DayForLight(one.AddrCoin, temp, height)
		payNum = append(payNum, pns...)
	}
	// engine.Log.Info("start spaces reward 5")
	//质押加权分
	pledgeTotal := 0
	if pledgeTotal == 0 {
		//质押为0，则平均分配
		temp := new(big.Int).Div(pledgeReward, big.NewInt(int64(peerTotal))).Uint64()
		for _, one := range *peers {
			if one.SpaceNum <= 0 {
				continue
			}
			pns := mining.LinearRelease0DayForLight(one.AddrCoin, temp, height)
			payNum = append(payNum, pns...)
		}

	} else {

	}
	// engine.Log.Info("start spaces reward 6")
	//空间加权分
	for i, one := range *peers {
		if one.SpaceNum <= 0 {
			engine.Log.Debug("spaces num reware %d %d %s %d", height, i, one.AddrCoin.B58String(), 0)
			continue
		}
		temp := new(big.Int).Mul(spacesNum, big.NewInt(int64(one.SpaceNum)))
		value := new(big.Int).Div(temp, big.NewInt(int64(spaceNumTotal)))
		rewardOne := value.Uint64()
		engine.Log.Debug("spaces num reware %d %d %s %d", height, i, one.AddrCoin.B58String(), rewardOne)

		pns := mining.LinearRelease0DayForLight(one.AddrCoin, rewardOne, height)
		payNum = append(payNum, pns...)
	}
	// engine.Log.Info("start spaces reward 7")
	// mining.SendToMoreAddress(pns, gas, config.Wallet_keystore_default_pwd, "")
	// engine.Log.Info("开始发放奖励 %+v", payNum)
	tx, err := mining.SendToMoreAddressByPayload(payNum, gas, config.Wallet_keystore_default_pwd, nil)
	if err != nil {
		engine.Log.Error("Failed to create transaction %s", err.Error())
		return
	}
	// engine.Log.Info("start spaces reward 8")
	tx.Check()
	// engine.Log.Info("发放奖励交易大小 %d", len(*tx.Serialize()))
	vo := tx.GetVOJSON()
	bs, _ := json.Marshal(vo)
	engine.Log.Info("spaces transaction content %s transaction size %d", string(bs), len(*tx.Serialize()))

}

/*
	提供存储的节点信息
*/
type StorePeer struct {
	AddrNet   nodeStore.AddressNet //存储节点
	SpaceNum  uint64               //存储空间单位个数
	AddrCoin  crypto.AddressCoin   //收款地址
	FlashTime int64                //刷新时间
}

func (this *StorePeer) Proto() (*[]byte, error) {
	spp := go_protos.StorePeer{
		AddrNet:   this.AddrNet,
		SpaceNum:  this.SpaceNum,
		AddrCoin:  this.AddrCoin,
		FlashTime: this.FlashTime,
	}
	bs, err := spp.Marshal()
	if err != nil {
		return nil, err
	}
	return &bs, nil
	// return spp.Marshal()
}

func ParseStorePeer(bs *[]byte) (*StorePeer, error) {
	if bs == nil {
		return nil, nil
	}
	spp := new(go_protos.StorePeer)
	err := proto.Unmarshal(*bs, spp)
	if err != nil {
		return nil, err
	}
	sp := StorePeer{
		AddrNet:   spp.AddrNet,   //存储节点
		SpaceNum:  spp.SpaceNum,  //存储空间单位个数
		AddrCoin:  spp.AddrCoin,  //收款地址
		FlashTime: spp.FlashTime, //刷新时间
	}
	return &sp, nil
}

/*
	添加一个存储节点信息
*/
func AddStorePeer(addrNet nodeStore.AddressNet, sp *StorePeer) {
	// engine.Log.Info("添加存储节点 1111111111111111")
	value, ok := countStorePeer.Load(utils.Bytes2string(addrNet))
	if ok {
		// engine.Log.Info("添加存储节点 2222 %s", sp.AddrCoin.B58String())
		storePeer := value.(*StorePeer)
		storePeer.AddrCoin = sp.AddrCoin
		storePeer.SpaceNum = sp.SpaceNum
		storePeer.FlashTime = time.Now().Unix()
	} else {
		// engine.Log.Info("添加存储节点 3333 %s", sp.AddrCoin.B58String())
		sp.AddrNet = addrNet
		sp.FlashTime = time.Now().Unix()
		countStorePeer.Store(utils.Bytes2string(addrNet), sp)
	}
}

/*
	统计节点，交给发放奖励用
*/
func CountStorePeers() *[]*StorePeer {
	peers := make([]*StorePeer, 0)
	countStorePeer.Range(func(k, v interface{}) bool {
		storePeer := v.(*StorePeer)
		// fmt.Println("查询存储节点", storePeer.AddrNet.B58String())
		// engine.Log.Info("统计节点，交给发放奖励用 %s %v %d %d %d", storePeer.AddrCoin.B58String(), storePeer.FlashTime+timeoutInterval >= time.Now().Unix(), storePeer.FlashTime, timeoutInterval, time.Now().Unix())
		if storePeer.FlashTime+int64(cloud_reward.TimeoutInterval.Seconds()) >= time.Now().Unix() {
			//
			peers = append(peers, storePeer)
		}
		return true
	})
	return &peers
}

/*
	统计存储节点个数和存储空间总量
	@return    uint64    全网提供的存储节点地址个数
	@return    uint64    全网提供的存储空间总量
*/
func CountStoreTotal() (uint64, uint64) {
	spaceTotalAddr := uint64(0)
	spaceTotal := uint64(0)

	countStorePeer.Range(func(k, v interface{}) bool {
		storePeer := v.(*StorePeer)
		spaceTotalAddr++
		spaceTotal += storePeer.SpaceNum
		return true
	})
	return spaceTotalAddr, spaceTotal
}
