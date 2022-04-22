package config

import (
	"flag"
	"math"
	"math/big"
	"path/filepath"
	"time"
)

const (
	Version_0 = 0 //
	Version_1 = 1 //此版本广播消息机制发生改变，只广播消息hash，节点自己去同步消息到本地
)

// const (
// 	Wallet_txitem_save_mem = 0 //未花费的余额保存到内存
// 	Wallet_txitem_save_db  = 1 //未花费的余额保存到数据库
// )

const (
	Wallet_path          = "wallet"        //钱包目录
	Wallet_path_prkName  = "ec_prk.pem"    //私钥文件名称
	Wallet_path_pukName  = "ec_puk.pem"    //公钥文件名称
	Wallet_seed          = "seed_key.json" //密钥种子文件名称
	Wallet_addr_puk_type = "EC PUBLIC KEY"

	// Wallet_MDL_Total            = 30 * 10000 * 10000                                           //货币发行总量30亿
	// Wallet_MDL_lock             = 1 * 10000 * 10000                                            //预挖量
	// Wallet_MDL_first_mining     = 100                                                          //第一个旷工第一个块分配量
	// Wallet_MDL_mining           = Wallet_MDL_Total - Wallet_MDL_lock - Wallet_MDL_first_mining //剩余
	// Wallet_keystore_AES_CBC_IV  = [32]byte{}                                                   //钱包aes加密向量
)

const (
	Wallet_tx_type_start          = 0 //
	Wallet_tx_type_mining         = 1 //挖矿所得
	Wallet_tx_type_deposit_in     = 2 //备用见证人押金输入，余额锁定
	Wallet_tx_type_deposit_out    = 3 //备用见证人押金输出，余额解锁
	Wallet_tx_type_pay            = 4 //普通支付
	Wallet_tx_type_account        = 5 //申请名称
	Wallet_tx_type_account_cancel = 6 //注销名称
	Wallet_tx_type_vote_in        = 7 //参与见证人投票输入，余额锁定
	Wallet_tx_type_vote_out       = 8 //参与见证人投票输出，余额解锁
	// Wallet_tx_type_deposit_out_force = 9 //见证人3次未出块，强制退还押金

	// Wallet_tx_type_register_store   = 20 //注册成为存储服务提供方
	// Wallet_tx_type_unregister_store = 21 //注册成为存储服务提供方
	// Wallet_tx_type_resources        = 20 //购买存储资源下载权限
	// Wallet_tx_type_resources_upload = 21 //上传资源付费

	Wallet_tx_type_token_publish = 10 //token发布
	Wallet_tx_type_token_payment = 11 //token支付

	Wallet_tx_type_spaces_mining_in  = 12 //存储挖矿押金输入，余额锁定
	Wallet_tx_type_spaces_mining_out = 13 //存储挖矿押金输出，余额解锁
	Wallet_tx_type_spaces_use_in     = 14 //用户存储空间押金输入，余额锁定
	Wallet_tx_type_spaces_use_out    = 15 //用户存储空间押金输出，余额解锁

	Wallet_tx_type_end = 100 //
)

const (
	Mining_coin_total            = 13 * 10000 * 10000 * 1e8                  //货币发行总量13亿
	Mining_coin_premining        = 100000000 * 1e8                           //1 * 10000 * 10000 * 1e8 //预挖量
	Mining_coin_rest             = Mining_coin_total - Mining_coin_premining //剩余
	Mining_block_cycle           = 180                                       //产出减半周期，每4年出块的数量，单位：天
	Mining_block_time            = 10                                        //出块时间，单位：秒
	Mining_block_start_height    = 1                                         //初始块高度1//144962//250970//370703
	Mining_group_start_height    = Mining_block_start_height                 //初始组高度
	Mining_block_hash_count      = 100                                       //连续n个块的hash连接起来，做一次hash作为随机数
	Mining_group_min             = 1                                         //挖矿组最少成员，少于最少成员不出块
	Mining_group_max             = 3                                         //挖矿组最多成员，最多只有这么多个成员构成一个组
	Mining_deposit               = uint64(100000 * 1e8)                      //见证人押金最少金额
	Mining_vote                  = uint64(10000 * 1e8)                       //社区节点投票押金最少金额
	Mining_light_min             = uint64(10 * 1e8)                          //轻节点押金最少金额
	Mining_name_deposit_min      = uint64(1 * 1e8)                           //注册域名最少押金
	Mining_community_reward_time = 60                                        //60 * 60 * 24 //社区节点奖励间隔时间
	Wallet_community_reward_max  = 500                                       //社区节点一次奖励最大地址数量
	Mining_pay_vin_max           = 100                                       //单笔交易vin最大数量，vin太多给节点带来瞬时很大cpu和内存压力
	Mining_pay_vout_max          = 10000                                     //给多人转账，vout数量最大值
	Witness_backup_min           = Mining_group_min                          //备用见证人数需要满足见证人组最少数量
	Witness_backup_max           = 31                                        //备用见证人排名靠前的最多数量，之后的人依然是选举中的候选见证人。31
	Witness_backup_reward_max    = 99                                        //有奖励的最大见证人数量

	Block_size_max       = 1024 * 1024 * 8 //单个区块容量最大 8M
	Wallet_tx_lockHeight = 30              //交易锁定高度

	Block_confirm                         = 6                 //单位：组。见证人出块共识下区块安全确认数
	Wallet_balance_history                = 10                //历史记录，一次查询n条记录
	Wallet_sync_block_interval_time       = time.Second / 30  //区块同步间隔时间
	Wallet_sync_block_interval_time_relax = time.Second / 300 //区块同步间隔时间，放松
	Witness_token_supply_min              = 1                 //token发行总量的最少数量
	// Wallet_block_FrozenHeight_jump = 10000 //（起始高度+跳过高度）之前是用锁仓的区块高度，之后是表示时间。
	Wallet_frozen_time_min     = 1600000000                                            //余额锁仓功能升级为按时间解锁，以最少时间为标致
	Wallet_vote_start_height   = 0                                                     //从31万个块高度之后，才开放见证人和社区节点质押
	Wallet_not_build_block_max = 24 * 60 * 60 / Mining_block_time / Witness_backup_max //278, 31个见证人就是24小时
	Wallet_sync_block_timeout  = 10                                                    //同步区块超时时间，单位：秒
	Wallet_multicas_block_time = 20                                                    //自己构建区块并广播，广播超时时间

	Wallet_Memory_percentage_max = 90 //内存最大百分比的时候，就要控制内存了

)

const (
	DB_name = "data" //数据库目录名称
	DB_temp = "temp" //临时数据目录
)

var (
	DB_path                     = filepath.Join(Wallet_path, DB_name) //数据库目录路径
	DB_path_temp                = filepath.Join(Wallet_path, DB_temp) //临时数据库目录路径
	Miner                       = false                               //本节点是否是矿工
	InitNode                    = false                               //本节点是否是创世节点
	LoadNode                    = false                               //是否从wallet目录中已有的区块数据拉起链端
	DB_is_null                  = false                               //启动时区块链数据库是否为空
	Wallet_keystore_default_pwd = "123456789"                         //钱包默认密码

	SubmitDepositin            = false    //自己提交见证人押金
	AlreadyMining              = false    //已经出过块了
	EnableCache                = true     //是否启用leveldb数据库缓存
	StartBlockHash             = []byte{} //
	Wallet_print_serialize_hex = false    //

	Witness_backup_group            = 5              //备用见证人组数量
	Witness_backup_group_overheight = uint64(703000) //超过这一高度，使用新的备用见证人组数量
	Witness_backup_group_new        = 60             //新的备用见证人组数量，保证半小时
)

/*
	判断是否有init参数
*/
func ParseInitFlag() bool {
	if InitNode {
		return true
	}
	for _, param := range flag.Args() {
		switch param {
		case "init":
			InitNode = true
			Model = Model_complete
			return true
		case "load":
			LoadNode = true
			Model = Model_complete
			return true
		}
	}
	return false
}

/*
	通过区块高度计算区块奖励
*/
// func ClacRewardForBlockHeight(height uint64) uint64 {
// 	totalBlockForDay := (60 * 60 * 24) / Mining_block_time                                //计算一天出多少块
// 	firstReward := uint64(Mining_coin_rest / 2 / (totalBlockForDay * Mining_block_cycle)) //计算首块奖励
// 	intervalBlockCount := uint64(Mining_block_cycle * totalBlockForDay)                   //计算达到多少块后产出减半
// 	n := height / intervalBlockCount
// 	for i := uint64(0); i < n; i++ {
// 		firstReward = firstReward / 2
// 	}
// 	return firstReward
// }

/*
	通过区块高度计算区块奖励
	第一个区块奖励286枚，
	1到30000高度，每增加一个块加0.0193枚/块；
	30001到60000高度，每增加一个块加0.0238枚/块；
	60001到90000高度，每增加一个块加0.0268枚/块；
	90001到120000高度，每增加一个块加0.0319枚/块；
	120001到184999高度，每增加一个块加0.0349枚/块；
	185000减产10%；
	185000开始每增加10472高度减产10%。
*/
// func ClacRewardForBlockHeight(height uint64) uint64 {
// 	// totalBlockForDay := (60 * 60 * 24) / Mining_block_time //计算一天出多少块
// 	// intervalHeight := 30000                                //间隔高度
// 	// boundary := 185000                                     //高度分界线

// 	result := uint64(28600000000 / 31)
// 	if height <= 30000*31 {
// 		//
// 		for i := uint64(1); i < height; i++ {
// 			result = result + (1930000 / 31)
// 		}

// 	} else if height <= 60000*31 {
// 		result = uint64(2790258387)
// 		for i := uint64(0); i < height; i++ {
// 			result = result + (2380000 / 31)
// 		}
// 	} else if height <= 90000*31 {
// 		result = uint64(4657998387)
// 		for i := uint64(0); i < height; i++ {
// 			result = result + (2680000 / 31)
// 		}
// 	} else if height <= 120000*31 {
// 		result = uint64(6525738387)
// 		for i := uint64(0); i < height; i++ {
// 			result = result + (3190000 / 31)
// 		}
// 	} else if height <= 184999*31 {
// 		result = uint64(8393478387)
// 		for i := uint64(0); i < height; i++ {
// 			result = result + (3490000 / 31)
// 		}
// 	} else {
// 		//185000开始每增加10472高度减产10%
// 		//TODO 性能有待提升
// 		result = uint64(12440186129)
// 		for i := uint64(0); i < (height / uint64(10472)); i++ {
// 			result = result - (result / 10)
// 			if result <= 100 {
// 				result = 0
// 				break
// 			}
// 		}
// 	}
// 	return result
// }

/*
	通过区块高度计算区块奖励
	第一个区块奖励286枚，
	1到30000高度，每增加一个块加0.0193枚/块；
	30001到60000高度，每增加一个块加0.0238枚/块；
	60001到90000高度，每增加一个块加0.0268枚/块；
	90001到120000高度，每增加一个块加0.0319枚/块；
	120001到184999高度，每增加一个块加0.0349枚/块；
	185000减产10%；
	185000开始每增加10472高度减产10%。
*/
// func ClacRewardForBlockHeight(height uint64) uint64 {
// 	// def calcr_new3(height):
// 	//    res = 286.0
// 	//    if height < 490000*31+1:
// 	//        res = (res+(int(height/31))*0.00347)/31
// 	//    else:
// 	//        base = res + 490000 * 0.00347
// 	//        res = 0.95*base
// 	//        res = res*math.pow( 0.95, int((height - 490000*31)/(20000*31)) )/31
// 	//        res = max(res, 0)

// 	//    return res

// 	// sum = 0
// 	// for i in range(2110001*31):
// 	//    sum = sum + calcr_new3(i)

// 	// #sum = 1299664262
// 	// print("sum=", sum)

// 	//-----------------
// 	ClacRewardForBlockHeightBase(height)

// 	engine.Log.Info("====================================")

// 	heightBig := big.NewInt(int64(height))
// 	num31 := big.NewInt(31)
// 	num347 := big.NewInt(347000)
// 	res := big.NewInt(28600000000)
// 	if height < 490000*31+1 {
// 		temp := new(big.Int).Div(heightBig, num31)
// 		temp = new(big.Int).Mul(temp, num347)
// 		temp = new(big.Int).Add(res, temp)
// 		temp = new(big.Int).Div(temp, num31)
// 		return temp.Uint64()
// 	} else {
// 		num49 := big.NewInt(490000)
// 		num95 := big.NewInt(95)
// 		num100 := big.NewInt(100)
// 		num2 := big.NewInt(20000)
// 		base := new(big.Int).Mul(num49, num347)
// 		base = new(big.Int).Add(res, base)
// 		res = new(big.Int).Mul(num95, base)
// 		res = new(big.Int).Div(res, num100)

// 		engine.Log.Info("2222222222 base = %v res = %v", base.Uint64(), res.Uint64())

// 		temp1 := new(big.Int).Mul(num49, num31)
// 		temp1 = new(big.Int).Sub(heightBig, temp1)
// 		temp2 := new(big.Int).Mul(num2, num31)
// 		temp := new(big.Int).Div(temp1, temp2)
// 		engine.Log.Info("222222222 temp = %v", temp)

// 		div1 := new(big.Int).Exp(num95, temp, nil)
// 		div2 := new(big.Int).Exp(num100, temp, nil)

// 		engine.Log.Info("div1 = %v div2 = %v", div1, div2)

// 		div1 = new(big.Int).Mul(res, div1)
// 		div2 = new(big.Int).Mul(res, div2)
// 		engine.Log.Info("div1 = %v div2 = %v", div1, div2)
// 		temp = new(big.Int).Div(div1, div2)
// 		engine.Log.Info("temp = %v", temp.Uint64())
// 		temp = new(big.Int).Div(temp, num31)

// 		engine.Log.Info("22222222222 temp = %v", temp)

// 		if temp.Cmp(big.NewInt(0)) == -1 {
// 			return 0
// 		}

// 		return temp.Uint64()
// 	}

// }

/*
	通过区块高度计算区块奖励
	第一个区块奖励286枚，
	1到30000高度，每增加一个块加0.0193枚/块；
	30001到60000高度，每增加一个块加0.0238枚/块；
	60001到90000高度，每增加一个块加0.0268枚/块；
	90001到120000高度，每增加一个块加0.0319枚/块；
	120001到184999高度，每增加一个块加0.0349枚/块；
	185000减产10%；
	185000开始每增加10472高度减产10%。
*/
func ClacRewardForBlockHeight(height uint64) uint64 {
	// def calcr_new3(height):
	//    res = 286.0
	//    if height < 490000*31+1:
	//        res = (res+(int(height/31))*0.00347)/31
	//    else:
	//        base = res + 490000 * 0.00347
	//        res = 0.95*base
	//        res = res*math.pow( 0.95, int((height - 490000*31)/(20000*31)) )/31
	//        res = max(res, 0)

	//    return res

	// sum = 0
	// for i in range(2110001*31):
	//    sum = sum + calcr_new3(i)

	// #sum = 1299664262
	// print("sum=", sum)

	//-----------------
	// ClacRewardForBlockHeightBase(height)

	// engine.Log.Info("====================================")

	heightBig := big.NewInt(int64(height))
	num31 := big.NewInt(31)
	num347 := big.NewInt(347000)
	res := big.NewInt(28600000000)
	if height < 490000*31+1 {
		temp := new(big.Int).Div(heightBig, num31)
		temp = new(big.Int).Mul(temp, num347)
		temp = new(big.Int).Add(res, temp)
		temp = new(big.Int).Div(temp, num31)
		return temp.Uint64()
	} else {
		num49 := big.NewInt(490000)
		num95 := big.NewInt(95)
		num100 := big.NewInt(100)
		num2 := big.NewInt(20000)
		base := new(big.Int).Mul(num49, num347)
		base = new(big.Int).Add(res, base)
		res = new(big.Int).Mul(num95, base)

		// engine.Log.Info("2222222222 base = %v res = %v", base.Uint64(), res.Uint64())
		//res*math.pow( 0.95, int((height - 490000*31)/(20000*31)) )/31
		temp1 := new(big.Int).Mul(num49, num31)
		temp1 = new(big.Int).Sub(heightBig, temp1)
		temp2 := new(big.Int).Mul(num2, num31)
		temp := new(big.Int).Div(temp1, temp2)
		// engine.Log.Info("222222222 temp = %v", temp)

		div1 := new(big.Int).Exp(num95, temp, nil)
		div2 := new(big.Int).Exp(num100, temp, nil)

		// engine.Log.Info("div1 = %v div2 = %v", div1, div2)

		res = new(big.Int).Mul(res, div1)
		res = new(big.Int).Div(res, num100)
		temp = new(big.Int).Div(res, div2)

		// engine.Log.Info("div1 = %v div2 = %v", div1, div2)

		// engine.Log.Info("temp = %v", temp.Uint64())
		temp = new(big.Int).Div(temp, num31)

		// engine.Log.Info("22222222222 temp =

		// engine.Log.Info("22222222222 temp = %v", temp)

		if temp.Cmp(big.NewInt(0)) == -1 {
			return 0
		}

		return temp.Uint64()
	}

}

/*
	通过区块高度计算区块奖励
	第一个区块奖励286枚，
	1到30000高度，每增加一个块加0.0193枚/块；
	30001到60000高度，每增加一个块加0.0238枚/块；
	60001到90000高度，每增加一个块加0.0268枚/块；
	90001到120000高度，每增加一个块加0.0319枚/块；
	120001到184999高度，每增加一个块加0.0349枚/块；
	185000减产10%；
	185000开始每增加10472高度减产10%。
*/
func ClacRewardForBlockHeightBase(height uint64) uint64 {
	res := 286.0
	if height < 490000*31+1 {
		res = (res + (float64(height/31))*0.00347) / 31
	} else {
		base := res + 490000*0.00347
		res = 0.95 * base

		res = res * math.Pow(0.95, float64((height-(490000*31))/(20000*31))) / 31
	}
	if res < 0 {
		res = 0
	}
	res = res * 100000000
	// return uint64(res)
	// engine.Log.Info("--- %d", uint64(res))
	return uint64(res)
}
