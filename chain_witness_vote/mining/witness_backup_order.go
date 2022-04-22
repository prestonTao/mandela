package mining

import (
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"math/big"
	"sort"
)

func OrderWitness(ws []*Witness, random *[]byte) []*Witness {
	witnesses := make(map[string]*Witness)
	addrs := make([]*[]byte, 0)
	for i, one := range ws {
		addrBs := []byte(*one.Addr)
		addrs = append(addrs, &addrBs)
		witnesses[utils.Bytes2string(*one.Addr)] = ws[i]
	}

	idasc := NewAddrASC(random, addrs)
	idsOrder := idasc.Sort()

	newWitness := make([]*Witness, 0)

	// var start *Witness
	// var last *Witness
	for i, _ := range idsOrder {
		addrOne := idsOrder[i]
		addr := crypto.AddressCoin(*addrOne)
		witness := witnesses[utils.Bytes2string(addr)]
		newWitness = append(newWitness, witness)
	}
	// return start
	return newWitness
}

/*
	将节点地址随机排序
*/
func OrderNodeAddr(addrs []nodeStore.AddressNet) []nodeStore.AddressNet {
	witnesses := make(map[string]nodeStore.AddressNet)
	newaddrBs := make([]*[]byte, 0)
	for i, one := range addrs {
		addrBs := []byte(addrs[i])
		newaddrBs = append(newaddrBs, &addrBs)
		witnesses[utils.Bytes2string(one)] = addrs[i]
	}

	random := utils.GetHashForDomain(utils.GetRandomDomain())

	idasc := NewAddrASC(&random, newaddrBs)
	idsOrder := idasc.Sort()

	newaddrs := make([]nodeStore.AddressNet, 0)
	for i, _ := range idsOrder {
		addrOne := idsOrder[i]
		addr := nodeStore.AddressNet(*addrOne)
		witness := witnesses[utils.Bytes2string(addr)]
		newaddrs = append(newaddrs, witness)
	}
	return newaddrs
}

/*
	收款地址排序算法
	从小到大排序
*/
type AddrASC struct {
	findNode *big.Int
	nodes    []*big.Int
	addrMap  map[string]*[]byte
}

func (this AddrASC) Len() int {
	return len(this.nodes)
}

func (this AddrASC) Less(i, j int) bool {
	a := new(big.Int).Xor(this.findNode, this.nodes[i])
	b := new(big.Int).Xor(this.findNode, this.nodes[j])
	if a.Cmp(b) > 0 {
		return false
	} else {
		return true
	}
}

func (this AddrASC) Swap(i, j int) {
	this.nodes[i], this.nodes[j] = this.nodes[j], this.nodes[i]
}

func (this AddrASC) Sort() []*[]byte {
	// sort.Sort(this)
	sort.Stable(this)
	result := make([]*[]byte, 0)
	for _, one := range this.nodes {
		// mhash := this.addrMap[hex.EncodeToString(one.Bytes())]
		mhash := this.addrMap[utils.Bytes2string(one.Bytes())]
		result = append(result, mhash)
	}
	return result
}

/*
	创建一个收款地址排序算法
	不能有重复地址
*/
func NewAddrASC(random *[]byte, addrs []*[]byte) *AddrASC {
	addrMap := make(map[string]*[]byte)
	addrArray := make([]*big.Int, 0)
	for i, one := range addrs {
		oneBig := new(big.Int).SetBytes(*one)
		// addrMap[hex.EncodeToString(oneBig.Bytes())] = addrs[i]
		addrMap[utils.Bytes2string(oneBig.Bytes())] = addrs[i]
		addrArray = append(addrArray, oneBig)
	}
	findNode := new(big.Int).SetBytes(*random)

	return &AddrASC{
		findNode: findNode,
		nodes:    addrArray,
		addrMap:  addrMap,
	}
}
