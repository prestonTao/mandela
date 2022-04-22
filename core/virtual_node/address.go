package virtual_node

import (
	"mandela/core/nodeStore"
	"mandela/core/utils/base58"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math/big"
)

//扩展地址
type AddressNetExtend nodeStore.AddressNet

func (this *AddressNetExtend) B58String() string {
	if this == nil || len(*this) <= 0 {
		return ""
	}
	return string(base58.Encode(*this))
}

func AddressFromB58String(str string) AddressNetExtend {
	if str == "" {
		return nil
	}
	return AddressNetExtend(base58.Decode(str))
}

/*
	通过公钥生成网络节点地址，将公钥两次hash得到网络节点地址
	@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddrNetExtend(addr nodeStore.AddressNet, index uint64) AddressNetExtend {
	bs := []byte(addr)
	binary.LittleEndian.PutUint64(bs, index)

	//计算SHA-256哈希值
	hashAddr := sha256.Sum256(bs)
	addrEx := AddressNetExtend(hashAddr[:])
	return addrEx
}

/*
	检查公钥生成的地址是否一样
	@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddr(addr nodeStore.AddressNet, index uint64, addrEx AddressNetExtend) bool {
	if addr == nil {
		return false
	}
	tagAddr := BuildAddrNetExtend(addr, index)
	return bytes.Equal(tagAddr, addrEx)
}

/*
	得到保存数据的逻辑节点
	@idStr  id十六进制字符串
	@return 4分之一节点
*/
func GetQuarterLogicAddrNetByAddrNetExtend(id *AddressNetExtend) (logicIds []*AddressNetExtend) {

	logicIds = make([]*AddressNetExtend, 0)
	logicIds = append(logicIds, id)
	idInt := new(big.Int).SetBytes(*id)
	for _, one := range nodeStore.Number_quarter {
		bs := new(big.Int).Xor(idInt, one).Bytes()
		// mhbs, _ := utils.Encode(bs, config.HashCode)
		// mh := utils.Multihash(mhbs)
		mh := AddressNetExtend(bs)
		logicIds = append(logicIds, &mh)
	}
	return
}
