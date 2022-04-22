package config

import (
	"errors"
	"strconv"
)

const (
	ERROR_fail = 5005
)

var (
	ERROR_chain_sysn_block_fail    = errors.New("sync fail,not find block") //同步失败，未找到区块
	ERROR_chain_sync_block_timeout = errors.New("sync block timeout")       //同步区块超时
	ERROR_wait_msg_timeout         = errors.New("wait message timeout")     //等待消息返回超时

	ERROR_deposit_witness   = errors.New("deposit shoud be:" + strconv.Itoa(int(Mining_deposit))) //见证人押金数量
	ERROR_deposit_not_exist = errors.New("deposit not exist")                                     //没有缴纳押金
	ERROR_deposit_exist     = errors.New("deposit exist")                                         //押金已经存在

	ERROR_password_fail            = errors.New("password fail")                           //密码错误
	ERROR_not_enough               = errors.New("balance is not enough")                   //余额不足
	ERROR_token_not_enough         = errors.New("token balance is not enough")             //token余额不足
	ERROR_public_key_not_exist     = errors.New("not find public key")                     //未找到公钥
	ERROR_amount_zero              = errors.New("Transfer amount cannot be 0")             //转账金额不能为0
	ERROR_tx_not_exist             = errors.New("Transaction not found")                   //未找到交易
	ERROR_tx_format_fail           = errors.New("Error parsing transaction")               //解析交易错误
	ERROR_tx_Repetitive_vin        = errors.New("Duplicate VIN in transaction")            //交易中有重复的vin
	ERROR_tx_is_use                = errors.New("Transaction has been used")               //交易已经被使用
	ERROR_tx_fail                  = errors.New("Transaction error")                       //交易错误
	ERROR_tx_lockheight            = errors.New("Lock height error")                       //交易锁定高度错误
	ERROR_tx_frozenheight          = errors.New("frozen height error")                     //交易冻结高度错误
	ERROR_public_and_addr_notMatch = errors.New("The public key and address do not match") //公钥和地址不匹配
	ERROR_sign_fail                = errors.New("Signature error")                         //签名错误
	ERROR_vote_exist               = errors.New("Vote already exists")                     //投票已经存在
	ERROR_pay_vin_too_much         = errors.New("vin too much")                            //交易中vin数量过多

	ERROR_name_deposit = errors.New("Domain name deposit is required at least" + strconv.Itoa(int(Mining_name_deposit_min/1e8))) //域名押金最少需要 n

	ERROR_name_not_self      = errors.New("Domain name does not belong to itself") //域名不属于自己
	ERROR_name_exist         = errors.New("Domain name already exists")            //域名已经存在
	ERROR_name_not_exist     = errors.New("Domain name does not exist")            //域名不存在
	ERROR_get_sign_data_fail = errors.New("Error getting signed data")             //获取签名的数据时出错
	ERROR_params_not_enough  = errors.New("params not enough")                     //参数不够
	ERROR_params_fail        = errors.New("params fail")                           //参数错误
	ERROR_token_min_fail     = errors.New("params token min fail")                 //小于token发行最少数量

	ERROR_get_node_conn_fail = errors.New("get node conn fail") //获取连接失败

	ERROR_get_reward_count_sync = errors.New("get reward count sync") //正在异步统计社区节点奖励中
)
