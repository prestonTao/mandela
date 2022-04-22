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
	return LevelDB.Save([]byte(config.Block_Highest), &bs)
}

/*
	查询网络中区块最新高度
*/
func GetHighstBlock() uint64 {
	bs, err := LevelDB.Find([]byte(config.Block_Highest))
	if err != nil {
		return 0
	}
	return utils.BytesToUint64(*bs)
}

/*
	保存一笔交易所属的区块hash
*/
func SaveTxToBlockHash(txid, blockhash *[]byte) error {
	key := config.BuildTxToBlockHash(*txid)
	// engine.Log.Info("SaveTxToBlockHash :%s", hex.EncodeToString(*txid))
	return LevelTempDB.Save(key, blockhash)
}

/*
	获取一笔交易所属的区块hash
*/
func GetTxToBlockHash(txid *[]byte) (*[]byte, error) {
	// engine.Log.Info("GetTxToBlockHash :%s", hex.EncodeToString(*txid))
	key := config.BuildTxToBlockHash(*txid)
	value, err := LevelTempDB.Find(key)
	if err != nil {
		return nil, err
	}
	return value, nil
}
