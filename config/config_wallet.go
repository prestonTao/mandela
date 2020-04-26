package config

import (
	"flag"
	"path/filepath"
	"time"
)

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
	Wallet_tx_type_mining         = 0 //挖矿所得
	Wallet_tx_type_deposit_in     = 1 //备用见证人押金输入，余额锁定
	Wallet_tx_type_deposit_out    = 2 //备用见证人押金输出，余额解锁
	Wallet_tx_type_pay            = 3 //普通支付
	Wallet_tx_type_account        = 4 //申请名称
	Wallet_tx_type_account_cancel = 5 //注销名称
	Wallet_tx_type_vote_in        = 6 //参与见证人投票输入，余额锁定
	Wallet_tx_type_vote_out       = 7 //参与见证人投票输出，余额解锁
	// Wallet_tx_type_deposit_out_force = 8 //见证人3次未出块，强制退还押金

	// Wallet_tx_type_register_store   = 20 //注册成为存储服务提供方
	// Wallet_tx_type_unregister_store = 21 //注册成为存储服务提供方
	// Wallet_tx_type_resources        = 20 //购买存储资源下载权限
	// Wallet_tx_type_resources_upload = 21 //上传资源付费

	Wallet_tx_type_end = 100 //
)

const (
	Mining_coin_total            = 13 * 10000 * 10000 * 1e8                  //货币发行总量13亿
	Mining_coin_premining        = 13000 * 10000 * 1e8                       //预挖量
	Mining_coin_rest             = Mining_coin_total - Mining_coin_premining //剩余
	Mining_block_cycle           = 4 * 365                                   //产出减半周期，每4年出块的数量，单位：天
	Mining_block_time            = 10                                        //出块时间，单位：秒
	Mining_block_start_height    = 1                                         //初始块高度
	Mining_group_start_height    = Mining_block_start_height                 //初始组高度
	Mining_block_hash_count      = 100                                       //连续n个块的hash连接起来，做一次hash作为随机数
	Mining_group_min             = 1                                         //挖矿组最少成员，少于最少成员不出块
	Mining_group_max             = 3                                         //挖矿组最多成员，最多只有这么多个成员构成一个组
	Mining_deposit               = uint64(10000 * 1e8)                       //见证人押金最少金额
	Mining_vote                  = uint64(1000 * 1e8)                        //社区节点投票押金最少金额
	Mining_light_min             = uint64(10 * 1e8)                          //轻节点押金最少金额
	Mining_name_deposit_min      = uint64(1 * 1e8)                           //注册域名最少押金
	Mining_community_reward_time = 60 * 60 * 24                              //社区节点奖励间隔时间
	Mining_pay_vout_max          = 20000                                     //给多人转账，vout数量最大值
	Witness_backup_min           = Mining_group_min                          //备用见证人数需要满足见证人组最少数量
	Witness_backup_max           = 31                                        //备用见证人排名靠前的最多数量，之后的人依然是选举中的候选见证人。31
	Witness_backup_group         = 5                                         //备用见证人组数量
	Block_size_max               = 1024 * 1024 * 8                           //单个区块容量最大 8M

	Block_confirm = 6 //单位：组。见证人出块共识下区块安全确认数

	Wallet_balance_history = 10 //历史记录，一次查询n条记录

	Wallet_sync_block_interval_time = time.Second / 3 //区块同步间隔时间

)

const (
	DB_name = "data" //数据库目录名称
)

var (
	DB_path                     = filepath.Join(Wallet_path, DB_name) //数据库目录路径
	Miner                       = false                               //本节点是否是矿工
	InitNode                    = false                               //本节点是否是创世节点
	DB_is_null                  = false                               //启动时区块链数据库是否为空
	Wallet_keystore_default_pwd = "123456789"                         //钱包默认密码

	SubmitDepositin = false //自己提交见证人押金
	AlreadyMining   = false //已经出过块了
)

/*
	判断是否有init参数
*/
func ParseInitFlag() bool {
	for _, param := range flag.Args() {
		switch param {
		case "init":
			InitNode = true
			return true
		}
	}
	return false
}

/*
	通过区块高度计算区块奖励
*/
func ClacRewardForBlockHeight(height uint64) uint64 {
	totalBlockForDay := (60 * 60 * 24) / Mining_block_time                                //计算一天出多少块
	firstReward := uint64(Mining_coin_rest / 2 / (totalBlockForDay * Mining_block_cycle)) //计算首块奖励
	intervalBlockCount := uint64(Mining_block_cycle * totalBlockForDay)                   //计算达到多少块后产出减半
	n := height / intervalBlockCount
	for i := uint64(0); i < n; i++ {
		firstReward = firstReward / 2
	}
	return firstReward
}
