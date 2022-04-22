package cloud_reward

import (
	"time"
)

const (
	HoldTimeInterval   = time.Second * 60 * 10 //发送心跳间隔时间，单位：秒
	TimeoutInterval    = time.Second * 60 * 60 //清理缓存时间
	RewardTimeInterval = time.Second * 60 * 10 //奖励间隔时间
)
