package virtual_node

import (
	"math/big"
)

/*
	得到每个节点网络的网络号，不包括本节点
	@id        *utils.Multihash    要计算的id
	@level     int                 深度
*/
func GetNodeNetworkNum(level uint, id AddressNetExtend) []*AddressNetExtend {

	root := new(big.Int).SetBytes(id)

	ids := make([]*AddressNetExtend, 0)
	for i := 0; i < int(level); i++ {
		//---------------------------------
		//将后面的i位置零
		//---------------------------------
		//		startInt := new(big.Int).Lsh(new(big.Int).Rsh(root, uint(i)), uint(i))
		//---------------------------------
		//第i位取反
		//---------------------------------
		networkNum := new(big.Int).Xor(root, new(big.Int).Lsh(big.NewInt(1), uint(i)))

		mhbs := AddressNetExtend(networkNum.Bytes())

		ids = append(ids, &mhbs)
	}

	return ids
}
