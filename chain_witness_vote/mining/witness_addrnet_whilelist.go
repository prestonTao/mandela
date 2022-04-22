/*
	见证人网络地址白名单管理
*/
package mining

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"sync"
)

func init() {
	// go loopAddWitness()
	utils.Go(loopAddWitness)
}

var newWitnessAddrNets = make(chan *[]*crypto.AddressCoin, 100)
var addrnetWhilelist = new(sync.Map) //保存见证人对应网络地址.key:string=见证人地址;value:*nodeStore.AddressNet=见证人网络地址;

func loopAddWitness() {
	for one := range newWitnessAddrNets {
		chain := GetLongChain()
		if !chain.SyncBlockFinish {
			continue
		}
		lookupAddrs := new(go_protos.RepeatedBytes)
		for _, one := range *one {
			if findWitnessAddrNet(one) == nil {
				lookupAddrs.Bss = append(lookupAddrs.Bss, *one)
			}
		}
		continue
		LookupWitnessAddrNet(lookupAddrs)
	}
}

func addWitnessAddrNet(addrCoin *crypto.AddressCoin, addrNet *nodeStore.AddressNet) {
	addrnetWhilelist.Store(utils.Bytes2string(*addrCoin), addrNet)
}

func findWitnessAddrNet(addrCoin *crypto.AddressCoin) *nodeStore.AddressNet {
	value, ok := addrnetWhilelist.Load(utils.Bytes2string(*addrCoin))
	if ok {
		if value != nil {
			addrNet := value.(*nodeStore.AddressNet)
			return addrNet
		}
	}
	return nil
}

func AddWitnessAddrNets(addrs []*crypto.AddressCoin) {
	select {
	case newWitnessAddrNets <- &addrs:
	default:
	}
}

/*
	寻找见证人地址
*/
func LookupWitnessAddrNet(lookupAddrs *go_protos.RepeatedBytes) {
	if lookupAddrs == nil || len(lookupAddrs.Bss) <= 0 {
		return
	}

	bs, err := lookupAddrs.Marshal()
	if err != nil {
		engine.Log.Error("LookupWitnessAddrNet error:%s", err.Error())
		return
	}

	//开始广播要寻找的见证人地址
	ok := message_center.SendMulticastMsg(config.MSGID_multicast_find_witness, &bs)
	if !ok {
		return
	}

}
