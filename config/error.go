package config

import (
	"errors"
	"strconv"
)

const (
	ERROR_fail = 5005
)

var (
	ERROR_chain_sysn_block_fail = errors.New("sync fail,not find block") //同步失败，未找到区块

	ERROR_deposit_witness   = errors.New("deposit shoud be:" + strconv.Itoa(int(Mining_deposit))) //见证人押金数量
	ERROR_deposit_not_exist = errors.New("deposit not exist")                                     //没有缴纳押金

	ERROR_password_fail        = errors.New("password fail")                                                                             //密码错误
	ERROR_not_enough           = errors.New("balance is not enough")                                                                     //余额不足
	ERROR_public_key_not_exist = errors.New("not find public key")                                                                       //未找到公钥
	ERROR_amount_zero          = errors.New("Transfer amount cannot be 0")                                                               //转账金额不能为0
	ERROR_tx_not_exist         = errors.New("Transaction not found")                                                                     //未找到交易
	ERROR_tx_format_fail       = errors.New("Error parsing transaction")                                                                 //解析交易错误
	ERROR_tx_is_use            = errors.New("Transaction has been used")                                                                 //交易已经被使用
	ERROR_tx_fail              = errors.New("Transaction error")                                                                         //交易错误
	ERROR_sign_fail            = errors.New("Signature error")                                                                           //签名错误
	ERROR_vote_exist           = errors.New("Vote already exists")                                                                       //投票已经存在
	ERROR_name_deposit         = errors.New("Domain name deposit is required at least" + strconv.Itoa(int(Mining_name_deposit_min/1e8))) //域名押金最少需要 n
	ERROR_name_not_self        = errors.New("Domain name does not belong to itself")                                                     //域名不属于自己
	ERROR_name_exist           = errors.New("Domain name already exists")                                                                //域名已经存在
	ERROR_name_not_exist       = errors.New("Domain name does not exist")                                                                //域名不存在
	ERROR_get_sign_data_fail   = errors.New("Error getting signed data")                                                                 //获取签名的数据时出错
	ERROR_params_not_enough    = errors.New("params not enough")                                                                         //参数不够
	ERROR_params_fail          = errors.New("params fail")                                                                               //参数错误
)
