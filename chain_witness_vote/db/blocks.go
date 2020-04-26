package db

import (
	"mandela/config"
	"mandela/core/utils"
)

/*
	保存网络中区块最新高度
*/
func SaveHighstBlock(number uint64) error {
	bs := utils.Uint64ToBytes(number)
	return Save([]byte(config.Block_Highest), &bs)
}

/*
	查询网络中区块最新高度
*/
func GetHighstBlock() uint64 {
	bs, err := Find([]byte(config.Block_Highest))
	if err != nil {
		return 0
	}
	return utils.BytesToUint64(*bs)
}
