package core

import (
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"sort"
	"sync"
	"time"
)

var netSpeedMap = new(sync.Map) //保存节点同步超时时间

/*
	添加保存一个节点超时时间
*/
func AddNodeAddrSpeed(addr nodeStore.AddressNet, speed time.Duration) {
	netSpeedMap.Store(utils.Bytes2string(addr), speed)
}

func DelNodeAddrSpeed(addr nodeStore.AddressNet) {
	netSpeedMap.Delete(utils.Bytes2string(addr))
}

/*
	通过同步节点速度排序节点地址
*/
func SortNetAddrForSpeed(netAddrs []nodeStore.AddressNet) []AddrNetSpeedInfo {
	addrNetSpeedASC := NewAddrNetSpeedASC(netAddrs)
	return addrNetSpeedASC.Sort()
}

/*
	网络地址同步速度排序算法
	超时时间从小到大排序
*/
type AddrNetSpeedASC struct {
	addrs []AddrNetSpeedInfo
}

func (this AddrNetSpeedASC) Len() int {
	return len(this.addrs)
}

func (this AddrNetSpeedASC) Less(i, j int) bool {

	// a := new(big.Int).Xor(this.findNode, this.nodes[i])
	// b := new(big.Int).Xor(this.findNode, this.nodes[j])
	// if a.Cmp(b) > 0 {
	if this.addrs[i].Speed > this.addrs[j].Speed {
		return false
	} else {
		return true
	}
}

func (this AddrNetSpeedASC) Swap(i, j int) {
	this.addrs[i], this.addrs[j] = this.addrs[j], this.addrs[i]
}

func (this AddrNetSpeedASC) Sort() []AddrNetSpeedInfo {
	// sort.Sort(this)
	sort.Stable(this)
	// result := make([]nodeStore.AddressNet, 0)
	// for i, _ := range this.addrs {
	// 	// mhash := this.addrMap[hex.EncodeToString(one.Bytes())]
	// 	// mhash := this.addrMap[utils.Bytes2string(one.Bytes())]
	// 	// engine.Log.Info("timeout sort: %s %d", this.addrs[i].AddrNet.B58String(), this.addrs[i].Speed)
	// 	result = append(result, this.addrs[i].AddrNet)
	// }
	return this.addrs
}

/*
	创建一个网络地址同步速度排序算法
	不能有重复地址
*/
func NewAddrNetSpeedASC(addrs []nodeStore.AddressNet) *AddrNetSpeedASC {
	addrNetSpeedASC := new(AddrNetSpeedASC)

	addrMap := make(map[string]nodeStore.AddressNet)
	for i, one := range addrs {
		key := utils.Bytes2string(one)
		if _, ok := addrMap[key]; ok {
			continue
		}
		info := AddrNetSpeedInfo{
			AddrNet: addrs[i],
			Speed:   0,
		}
		value, ok := netSpeedMap.Load(key)
		if ok {
			info.Speed = int64(value.(time.Duration))
		}
		addrMap[key] = addrs[i]
		addrNetSpeedASC.addrs = append(addrNetSpeedASC.addrs, info)
	}

	return addrNetSpeedASC
}

type AddrNetSpeedInfo struct {
	AddrNet nodeStore.AddressNet
	Speed   int64
}
