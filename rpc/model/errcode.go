package model

import (
	"fmt"
	"strconv"
)

const (
	Success                = 2000 //成功
	NoMethod               = 4001 //没有这个方法
	TypeWrong              = 5001 //参数类型错误
	NoField                = 5002 //缺少参数
	Nomarl                 = 5003 //一般错误
	Timeout                = 5004 //超时
	Exist                  = 5005 //已经存在
	FailPwd                = 5006 //密码错误
	NotExist               = 5007 //不存在
	NotEnough              = 5008 //余额不足
	ContentIncorrectFormat = 5009 //参数格式不正确
	AmountIsZero           = 5010 //转账不能为0
	RuleField              = 5011 //地址角色不正确
	BalanceNotEnough       = 5012 //余额不足
	VoteExist              = 5013 //投票已经存在
)

var codes = map[int]string{
	NoMethod:               "no method",
	TypeWrong:              "type wrong",
	NoField:                "no field",
	Nomarl:                 "",
	Timeout:                "timeout",
	Exist:                  "exist",
	FailPwd:                "fail password",
	NotExist:               "not exist",
	NotEnough:              "not enough",
	ContentIncorrectFormat: "",
	AmountIsZero:           "",
	RuleField:              "",
	BalanceNotEnough:       "BalanceNotEnough",
	VoteExist:              "VoteExist",
}

func Errcode(code int, p ...string) (res []byte, err error) {
	res = []byte(strconv.Itoa(code))
	c, ok := codes[code]
	if ok {
		if len(p) > 0 {
			if c == "" {
				err = fmt.Errorf("%s", p[0])
			} else {
				err = fmt.Errorf("%s: %s", p[0], c)
			}
		} else {
			err = fmt.Errorf("%s", c)
		}
	}
	return
}
