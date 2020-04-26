package nodeStore

import (
	"math/big"
	"math/rand"
	"time"
	gconfig "mandela/config"
	"mandela/core/utils"
)

//得到指定长度的节点id
//@return 10进制字符串
func RandNodeId() *big.Int {
	min := rand.New(rand.NewSource(99))
	timens := int64(time.Now().Nanosecond())
	min.Seed(timens)
	maxId := new(big.Int).Lsh(big.NewInt(1), uint(gconfig.NodeIDLevel))
	randInt := new(big.Int).Rand(min, maxId)
	return randInt
}

/*
	得到一个节点最近的邻居节点
*/
func GetNearId(id *utils.Multihash) *utils.Multihash {
	networkNum := new(big.Int).Xor(new(big.Int).SetBytes(id.Data()), big.NewInt(1))
	mhbs, _ := utils.Encode(networkNum.Bytes(), gconfig.HashCode)
	mh := utils.Multihash(mhbs)
	return &mh
}
