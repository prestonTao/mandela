package name

import (
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"sync"

	"github.com/gogo/protobuf/proto"
)

var NameOfValidity = uint64(0) //域名有效高度。有效期为365天，根据出块速度换算为有效高度。
var names = new(sync.Map)      //保存自己注册的域名;key:string=域名;value:Name=域名信息;

func init() {
	//域名有效高度。有效期为365天，根据出块速度换算为有效高度。
	//计算出一年要出多少个块
	NameOfValidity = 60 * 60 * 24 * 365 / config.Mining_block_time
	// NameOfValidity = 20 / config.Mining_block_time
}

/*
	添加一个域名
*/
func AddName(name Nameinfo) {
	names.Store(string(name.Name), name)
}

/*
	删除一个域名
*/
func DelName(name []byte) {
	names.Delete(string(name))
}

/*
	查找域名
*/
func FindName(name string) *Nameinfo {
	itr, ok := names.Load(name)
	if !ok {
		return nil
	}
	txid := itr.(Nameinfo)
	return &txid
}

/*
	查找域名，从域名中随机选择一个地址返回
*/
// func FindNameRandOne(name string) *nodeStore.AddressNet {
// 	nameinfo := FindName(name)
// 	if nameinfo == nil {
// 		return nil
// 	}
// 	// nameinfo.CheckIsOvertime()
// 	addr := nameinfo.NetIds[utils.GetRandNum(len(nameinfo.NetIds))]
// 	return &addr
// }

/*
	获取域名列表
*/
func GetNameList() []Nameinfo {
	name := make([]Nameinfo, 0)
	names.Range(func(k, v interface{}) bool {
		nameOne := v.(Nameinfo)
		nameOne.NameOfValidity = nameOne.Height + NameOfValidity
		name = append(name, nameOne)
		return true
	})
	return name
}

type Nameinfo struct {
	Name           string                 //域名
	Txid           []byte                 //交易id
	NetIds         []nodeStore.AddressNet //节点地址
	AddrCoins      []crypto.AddressCoin   //钱包收款地址
	Height         uint64                 //注册区块高度，通过现有高度计算出有效时间
	NameOfValidity uint64                 //有效块数量
	Deposit        uint64                 //冻结金额
}

func (this *Nameinfo) Proto() ([]byte, error) {
	netids := make([][]byte, 0)
	for _, one := range this.NetIds {
		netids = append(netids, one)
	}
	addrCoins := make([][]byte, 0)
	for _, one := range this.AddrCoins {
		addrCoins = append(addrCoins, one)
	}
	nip := go_protos.Nameinfo{
		Name:           this.Name,
		Txid:           this.Txid,
		NetIds:         netids,
		AddrCoins:      addrCoins,
		Height:         this.Height,
		NameOfValidity: this.NameOfValidity,
		Deposit:        this.Deposit,
	}
	return nip.Marshal()
}

func ParseNameinfo(bs []byte) (*Nameinfo, error) {
	nip := new(go_protos.Nameinfo)
	err := proto.Unmarshal(bs, nip)
	if err != nil {
		return nil, err
	}

	netids := make([]nodeStore.AddressNet, 0)
	for _, one := range nip.NetIds {
		netids = append(netids, one)
	}
	addrCoins := make([]crypto.AddressCoin, 0)
	for _, one := range nip.AddrCoins {
		addrCoins = append(addrCoins, one)
	}
	nameinfo := Nameinfo{
		Name:           nip.Name,
		Txid:           nip.Txid,
		NetIds:         netids,
		AddrCoins:      addrCoins,
		Height:         nip.Height,
		NameOfValidity: nip.NameOfValidity,
		Deposit:        nip.Deposit,
	}
	return &nameinfo, nil
}

/*
	检查是否过期
*/
func (this *Nameinfo) CheckIsOvertime(height uint64) bool {
	if (this.Height + NameOfValidity) > height {
		return false
	}
	return true
}
