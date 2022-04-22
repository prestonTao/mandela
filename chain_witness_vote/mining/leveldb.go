package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"strconv"
)

/*
	通过区块hash从数据库加载区块头
*/
func LoadBlockHeadByHash(hash *[]byte) (*BlockHead, error) {
	bh, err := db.LevelDB.Find(*hash)
	if err != nil {
		return nil, err
	}
	return ParseBlockHeadProto(bh)
}

/*
	通过区块高度，查询一个区块头信息
*/
func LoadBlockHeadByHeight(height uint64) *BlockHead {
	bhash, err := db.LevelDB.Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
	if err != nil {
		return nil
	}

	bh, err := LoadBlockHeadByHash(bhash)
	// bs, err := db.Find(*bhash)
	// if err != nil {
	// 	return nil
	// }
	// bh, err := ParseBlockHead(bs)
	if err != nil {
		return nil
	}
	return bh
}

/*
	查询数据库和解析交易
*/
func LoadTxBase(txid []byte) (TxItr, error) {
	var err error
	var txItr TxItr
	ok := false
	//是否启用缓存
	if config.EnableCache {
		//先判断缓存中是否存在
		// txItr, ok = TxCache.FindTxInCache(hex.EncodeToString(txid))
		txItr, ok = TxCache.FindTxInCache(txid)
	}
	if !ok {
		// engine.Log.Error("未命中缓存 FindTxInCache")
		var bs *[]byte
		bs, err = db.LevelDB.Find(txid)
		if err != nil {
			return nil, err
		}

		txItr, err = ParseTxBaseProto(ParseTxClass(txid), bs)
		if err != nil {
			return nil, err
		}
	}
	return txItr, err
}

/*
	从数据库中加载一个区块
*/
func loadBlockForDB(bhash *[]byte) (*BlockHead, []TxItr, error) {

	hB, err := LoadBlockHeadByHash(bhash)
	if err != nil {
		return nil, nil, err
	}
	txItrs := make([]TxItr, 0)
	for _, one := range hB.Tx {
		// txItr, err := FindTxBase(one, hex.EncodeToString(one))
		txItr, err := LoadTxBase(one)

		// txBs, err := db.Find(one)
		if err != nil {
			// fmt.Println("3333", err)
			return nil, nil, err
		}
		// txItr, err := ParseTxBase(ParseTxClass(one), txBs)
		txItrs = append(txItrs, txItr)
	}

	return hB, txItrs, nil
}

/*
	从数据库加载一整个区块，包括区块中的所有交易
*/
func LoadBlockHeadVOByHash(hash *[]byte) (*BlockHeadVO, error) {
	bh, txs, err := loadBlockForDB(hash)
	if err != nil {
		return nil, err
	}

	bhvo := new(BlockHeadVO)
	bhvo.Txs = make([]TxItr, 0)

	bhvo.BH = bh
	bhvo.Txs = txs
	return bhvo, nil

	// //通过区块hash查找区块头
	// bh, err := LoadBlockHeadByHash(hash)

	// // bs, err := db.Find(hash)
	// if err != nil {
	// 	return nil, err
	// } else {
	// 	// bh, err := ParseBlockHead(bs)
	// 	// if err != nil {
	// 	// 	return nil, err
	// 	// }
	// 	bhvo.BH = bh
	// 	for _, one := range bh.Tx {
	// 		txOne, err := LoadTxBase(one)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		bhvo.Txs = append(bhvo.Txs, txOne)
	// 	}
	// }
	// return bhvo, nil
}
