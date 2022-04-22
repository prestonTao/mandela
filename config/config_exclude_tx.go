package config

import (
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"encoding/hex"
	"sync"
)

const Mining_block_start_height_jump = 0 //跳过区块高度不验证签名和交易锁定上链高度//631751//632060/634818/661575/700000//719745//814565/940000/981720
const WitnessOrderCorrectStart = 0       //599802/0
const WitnessOrderCorrectEnd = 0         //719745/981720
const RandomHashHeightMin = 0            //用一个无限高度控制随机数//
const NextHashHeightMax = 0

/*
1169333    5c91873af3f9b6d9b18c649d2c0a4c49b2540a31fee26546e259bc5fce88c2e9  +

1169334    ef57d947d2f7a10d5ade6024e4bef1c74743f1d79807914ee9e268e6074d71c6  +
1169335    77ce8deff0c43893e7fba2312d934f32a715063aa7ade53ae539bdcfa012b8a1  -
1169335    b89ace0f46b39cbddec13b6cd3bc3cba99cff244477e207928d0cd7d625d9907  +

1169336    e00171f80fb6dddaeb1aec410169c2348e0514a373f03e189380393cf244da36  -
1169336    9910e28c5868fbfa56e4cd052c01725b2b28f7462827c2e0eccba6bad1506fed  +
1169337    09f5d7c1211d190a7d93afd734412277b48fdc8ef2179552e0ca2ea81c2b44a4  +

1169338    2727a5e5361425d58a59275a1346bab2470310b387a45c3784dcdcae79b86bbd  +
*/
const FixBuildGroupBUGHeightMax = 0 //1367480

var RandomHashFixed = []byte{}

const RandomHashFixedStr = "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"

var SpecialAddrs = []byte{}

var SpecialBlockHash = []byte{}

var NextHash = new(sync.Map)

// var BlockNextHash = new(sync.Map) //临时修改区块的nexthash

var RandomMap = new(sync.Map) //make(map[uint64]*[]byte)

var Exclude_Tx = []ExcludeTx{
	// ExcludeTx{960, "040000000000000042ad26665e29ab78806b4t2a9a6e69f9fa4e5y8341c295834a114cb11e3bad62be", nil},
	// ExcludeTx{961, "040000000000000042ad26665e29atb78806b42a9a6ye69f9fa4e58341c295834a114cb11e3bad62be", nil},
}

var BlockHashs = make([][]byte, 0)

/*
	1.31个以内，按现有见证人数量平均分配（假如只有5个见证人，则5个见证人平均分）。
	2.31-99个，均分部分按现有见证人数量均分。
	3.大于99个见证人，均分部分给前99个见证人均分。排名99之后的见证人没有奖励。
*/
const Mining_witness_average_height = 0 //超过这一高度，区块奖励按新的规则算//624827
const Reward_witness_height = 0         //新版奖励起始高度//745333
/*
	上一版本的奖励还有一处bug，就是保存的见证人排序是按投票数量排序，
	然而寻找出块的时候，prehash导致不连续，导致寻找的出块见证人数量比预计的少，奖励也变少。
	新版将解决这个问题。
*/
const Reward_witness_height_new = 0 //新版奖励起始高度//800000

func init() {
	BuildRandom()
	CustomOrderBlockHash()

	for i, one := range Exclude_Tx {
		bs, err := hex.DecodeString(one.TxStr)
		if err != nil {
			panic("交易hash不规范:" + err.Error())
		}
		Exclude_Tx[i].TxByte = bs
		Exclude_Tx[i].TxStr = ""
		// one.TxByte = bs
		// one.TxStr = ""
	}
	// engine.Log.Info("打印需要排除的交易hash %+v", Exclude_Tx)
}

type ExcludeTx struct {
	Height uint64
	TxStr  string
	TxByte []byte
}

func BuildRandom() {
	// addrs := "TESTMSLAW5cNYDKDHAggEQtF7GmvSHZ42ocRB4"
	addrs := "TEST792Dg5nQQvpYoajJtcrsCR4vFG4xeyU9z4"
	SpecialAddrs = crypto.AddressFromB58String(addrs)

	blockhashStr := "bd775d8fd597816b791407f0f706b6d34d9e2883aa97f8eb50595c4faccbd63a"
	SpecialBlockHash, _ = hex.DecodeString(blockhashStr)

	RandomHashFixed, _ = hex.DecodeString(RandomHashFixedStr) //599803

	blockhash, _ := hex.DecodeString("327ed7c01ac82c31f3454f4a652921e4c14045c1220b566d08e8241358b93a03") //599803
	random, _ := hex.DecodeString("536c6049a7310afb957553ef9643a6efff586951cddb9832a556fbca912ced34")
	RandomMap.Store(utils.Bytes2string(blockhash), &random)

	blockhash, _ = hex.DecodeString("c7e52c1922ff9e7d9de87eca78e25ccf378c281b6340f982f5837c72fafad618") //630456
	random, _ = hex.DecodeString("8b4a13c54087342bd20979ef8e0a10045897cf5a0473b2a95f192f6fb004b031")
	RandomMap.Store(utils.Bytes2string(blockhash), &random)

	blockhash, _ = hex.DecodeString("a65a7bf2f888876c10b2d3ec7827d59cc8f4f22cd7538d02cde186724bdfa036") //599800
	random, _ = hex.DecodeString("903d5e2e20aba4464bbfc929419d7eb5b63abb8227ac56b1feec63b0c15fc0fd")
	RandomMap.Store(utils.Bytes2string(blockhash), &random)

	//------------------
	blockhash, _ = hex.DecodeString("ed74bccbd2f61c61eec8fbc33883e347c96429e94dc64f59faf677c715ae865d")
	random, _ = hex.DecodeString("a65a7bf2f888876c10b2d3ec7827d59cc8f4f22cd7538d02cde186724bdfa036")
	NextHash.Store(utils.Bytes2string(blockhash), &random)

	blockhash, _ = hex.DecodeString("a65a7bf2f888876c10b2d3ec7827d59cc8f4f22cd7538d02cde186724bdfa036")
	random, _ = hex.DecodeString("2d53b5e4e21ba7ff9508418e450dc17aeafc0c4e007316a45e19199c92d34b85")
	NextHash.Store(utils.Bytes2string(blockhash), &random)

	// blockhash, _ = hex.DecodeString("2d53b5e4e21ba7ff9508418e450dc17aeafc0c4e007316a45e19199c92d34b85")
	// random, _ = hex.DecodeString("e91a2b7f57fca199b274d0aa63d911c545f922f9ac19887ec093bc914e5a25fa")
	// NextHash.Store(utils.Bytes2string(blockhash), &random)

	//
	// blockhash, _ = hex.DecodeString("723419e31599f3edd29c8ce3e85d8d9b33f74fd137910539f398b89a958e854d")
	// nextHash, _ := hex.DecodeString("bd775d8fd597816b791407f0f706b6d34d9e2883aa97f8eb50595c4faccbd63a")
	// BlockNextHash.Store(utils.Bytes2string(blockhash), &nextHash)

	// blockhash, _ = hex.DecodeString("bd775d8fd597816b791407f0f706b6d34d9e2883aa97f8eb50595c4faccbd63a")
	// nextHash, _ = hex.DecodeString("be1b85a0e410dbb4248b2ae425c1d551b07d6a398ea7e9bb5bf0c74faf59ceb9")
	// BlockNextHash.Store(utils.Bytes2string(blockhash), &nextHash)

	// blockhash, _ = hex.DecodeString("be1b85a0e410dbb4248b2ae425c1d551b07d6a398ea7e9bb5bf0c74faf59ceb9")
	// nextHash, _ = hex.DecodeString("d6c53b2a05fbcee8dc12a6eff595c2a8dd51f2b68ac8bc4368056a9f69e05679")
	// BlockNextHash.Store(utils.Bytes2string(blockhash), &nextHash)

}

/*
	按照指定的顺序加载区块
	一个测试方法，加载到这里的第一个区块后，就按照这里的顺序加载指定的区块
*/
func CustomOrderBlockHash() {
	if false {
		hash, _ := hex.DecodeString("d6ab9a1e6b58d06f7c763cf0fb474f06fe4f88e9b50424eb852eb13e3140e7bf")
		BlockHashs = append(BlockHashs, hash)
		hash, _ = hex.DecodeString("ee480d940a729d2b6600bd92ed01ce9c448b1229a9dbc940ad970314de3ac9e7")
		BlockHashs = append(BlockHashs, hash)
		hash, _ = hex.DecodeString("bf40583a532b95f440926f4d17eecd61468581a7cacd659a42ef90c2d0fad54f")
		BlockHashs = append(BlockHashs, hash)
		hash, _ = hex.DecodeString("0b7d94bc4d497e59eb97c863a5798f9595d1f601d66d960700865f6138af7343")
		BlockHashs = append(BlockHashs, hash)
		hash, _ = hex.DecodeString("acbc1cacc247152762d65a13f5898f704b26556d950d54ef1af26b13a14dd749")
		BlockHashs = append(BlockHashs, hash)
		hash, _ = hex.DecodeString("ec9aaaa8505786b3a1a736c5505ff9d1b32d4fec271dae01158bbb194e250b9e")
		BlockHashs = append(BlockHashs, hash)
	}
}
