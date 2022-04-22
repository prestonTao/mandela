package nodeStore

import (
	"mandela/core/utils"
	"mandela/core/utils/base58"
	"bytes"
	"crypto/sha256"
)

//节点地址
type AddressNet []byte

func (this *AddressNet) B58String() string {
	return string(base58.Encode(*this))
}

func AddressFromB58String(str string) AddressNet {
	return AddressNet(base58.Decode(str))
}

/*
	通过公钥生成网络节点地址，将公钥两次hash得到网络节点地址
	@version    []byte    版本号（如比特币主网版本号“0x00"）
*/
func BuildAddr(pubKey []byte) AddressNet {
	//第一步，计算SHA-256哈希值
	publicSHA256 := sha256.Sum256(pubKey)
	//第二步，计算上一步结果的SHA-256哈希值
	temp := sha256.Sum256(publicSHA256[:])
	return temp[:]
}

/*
	检查公钥生成的地址是否一样
	@return    bool    是否一样 true=相同;false=不相同;
*/
func CheckPukAddr(pubKey []byte, addr AddressNet) bool {
	tagAddr := BuildAddr(pubKey)
	return bytes.Equal(tagAddr, addr)
}

/*
	去除重复地址
*/
func RemoveDuplicateAddress(addrs []*AddressNet) []*AddressNet {
	m := make(map[string]*AddressNet)
	for i, one := range addrs {
		// m[hex.EncodeToString(*one)] = addrs[i]
		m[utils.Bytes2string(*one)] = addrs[i]
	}
	dstAddrs := make([]*AddressNet, 0)
	for _, v := range m {
		dstAddrs = append(dstAddrs, v)
	}
	return dstAddrs
}
