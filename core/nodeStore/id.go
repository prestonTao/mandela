package nodeStore

import (
	"mandela/config"
	"mandela/core/utils"
	"math/big"
)

const (
	Str_zaro      = "0000000000000000000000000000000000000000000000000000000000000000" //字符串0
	Str_maxNumber = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" //256位的最大数十六进制表示id
//	Str_halfNumber    = "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" //最大id的二分之一
//	Str_quarterNumber = "3fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" //最大id的四分之一
)

var (
	Number_interval []*big.Int = make([]*big.Int, 0) //相隔距离16分之一

	//	Number_max       *big.Int //最大id数
	//	Number_half      *big.Int //最大id的二分之一
	Number_quarter []*big.Int = make([]*big.Int, 0) //最大id的四分之一
//	Number_eighth    *big.Int //最大id的八分之一
//	Number_sixteenth *big.Int //最大id的16分之一
)

func init() {
	//间隔16分之一
	Number_interval = initData(16)
	//间隔4分之一
	Number_quarter = initData(4)
}

//初始节点数据
//@num 几分之一 值为16，则为十六分之一
func initData(num int) []*big.Int {
	number_interval := make([]*big.Int, 0)
	Number_max, ok := new(big.Int).SetString(Str_maxNumber, 16)
	if !ok {
		panic("id string format error")
	}

	one_sixteenth := new(big.Int).Div(Number_max, big.NewInt(int64(num)))
	for i := 1; i < num; i++ {
		number_interval = append(number_interval, new(big.Int).Mul(one_sixteenth, big.NewInt(int64(i))))
	}
	return number_interval

}

/*
	得到保存数据的逻辑节点
	@idStr  id十六进制字符串
	@return 16分之一节点
*/
func GetLogicIds(id *utils.Multihash) (logicIds []*utils.Multihash) {

	logicIds = make([]*utils.Multihash, 0)
	idInt := new(big.Int).SetBytes(id.Data())
	for _, one := range Number_interval {
		bs := new(big.Int).Xor(idInt, one).Bytes()
		mhbs, _ := utils.Encode(bs, config.HashCode)
		mh := utils.Multihash(mhbs)
		logicIds = append(logicIds, &mh)
	}

	return
}

/*
	得到保存数据的逻辑节点
	@idStr  id十六进制字符串
	@return 4分之一节点
*/
// func GetQuarterLogicAddrNetByAddrCoin(id *crypto.AddressCoin) (logicIds []*AddressNet) {

// 	logicIds = make([]*AddressNet, 0)
// 	logicIds = append(logicIds, id)
// 	idInt := new(big.Int).SetBytes(*id)
// 	for _, one := range Number_quarter {
// 		bs := new(big.Int).Xor(idInt, one).Bytes()
// 		mh := AddressNet(bs)
// 		logicIds = append(logicIds, &mh)
// 	}
// 	return
// }

/*
	得到保存数据的逻辑节点
	@idStr  id十六进制字符串
	@return 4分之一节点
*/
func GetQuarterLogicAddrNetByAddrNet(id *AddressNet) (logicIds []*AddressNet) {

	logicIds = make([]*AddressNet, 0)
	logicIds = append(logicIds, id)
	idInt := new(big.Int).SetBytes(*id)
	for _, one := range Number_quarter {
		bs := new(big.Int).Xor(idInt, one).Bytes()
		// mhbs, _ := utils.Encode(bs, config.HashCode)
		// mh := utils.Multihash(mhbs)
		mh := AddressNet(bs)
		logicIds = append(logicIds, &mh)
	}
	return
}
