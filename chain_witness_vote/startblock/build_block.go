package startblock

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"bytes"
)

const (
// firstReward   =config.Mining_reward // config.Wallet_MDL_first_mining // 100000000 * mining.Unit //创始节点奖励
// depositInCoin = 1 * mining.Unit //押金

)

/*
	构建创世块
	创世块生成两个组，一个组一个块
	第一个块给3个见证者账户分配初始额度，下一个组见证者投票结果
	第二个块
*/
func BuildFirstBlock() (*mining.BlockHeadVO, error) {
	//如果模块未启用，则不执行

	if !config.InitNode || !config.DB_is_null {
		// fmt.Println("不是创始节点")
		return nil, nil
	}
	// config.InitNode = true

	// fmt.Println("开始创建创世区块")

	db.InitDB(config.DB_path)

	//----------------生成第一个区块-----------------

	balanceTotal := uint64(config.Mining_coin_premining)
	//构建交易
	txHashes := make([][]byte, 0)
	txs := make([]mining.TxItr, 0)
	//创建首块预挖奖励
	reward := BuildReward(balanceTotal)
	txs = append(txs, reward)
	txHashes = append(txHashes, *reward.GetHash())
	//创建见证人押金交易
	depositIn := BuildDepositIn()
	txs = append(txs, depositIn)
	txHashes = append(txHashes, *depositIn.GetHash())

	//区块头
	blockHead1 := mining.BlockHead{
		Height:      config.Mining_block_start_height, //区块高度(每秒产生一个块高度，也足够使用上千亿年)
		GroupHeight: config.Mining_group_start_height, //
		NTx:         uint64(len(txHashes)),            //交易数量
		Tx:          txHashes,                         //本区块包含的交易id
		Time:        utils.GetNow(),                   //  time.Now().Unix(),                //unix时间戳
		Witness:     keystore.GetCoinbase(),           //
	}
	blockHead1.BuildMerkleRoot()
	blockHead1.BuildSign(keystore.GetCoinbase())
	blockHead1.BuildHash()

	// blockHead1.FindNonce(1, make(chan bool, 1))
	bhbs, _ := blockHead1.Json()
	err := db.Save(blockHead1.Hash, bhbs)
	if err != nil {
		return nil, err
	}
	// fmt.Println("key", "blockHead", hex.EncodeToString(blockHead1.Hash))
	// fmt.Println("value", "blockHead", string(*bhbs), "\n")

	db.Save(config.Key_block_start, &blockHead1.Hash)

	hashExist := false

	//保存到数据库
	for _, one := range txs {
		one.SetBlockHash(blockHead1.Hash)
		// one.TxBase.BlockHash = blockHead1.Hash
		bs, err := one.Json()
		if err != nil {
			// fmt.Println("2 json格式化错误", err)
			return nil, err
		}

		if one.CheckHashExist() {
			hashExist = true
		}

		err = db.Save(*one.GetHash(), bs)
		if err != nil {
			return nil, err
		}

		// fmt.Println("key", "tx", hex.EncodeToString(*one.GetHash()))
		// fmt.Println("value", "tx", string(*bs))
	}

	//保存见证人押金交易
	depositIn.SetBlockHash(blockHead1.Hash)
	bs, err := depositIn.Json()
	if err != nil {
		return nil, err
	}
	err = db.Save(*depositIn.GetHash(), bs)
	if err != nil {
		return nil, err
	}
	// fmt.Println("key", "tx", hex.EncodeToString(*depositIn.GetHash()))
	// fmt.Println("value", "tx", string(*bs))

	// db.SaveBlockHeight(blockHead1.Height, &blockHead1.Hash)
	// fmt.Println("创建初始块完成")

	bhvo := mining.BlockHeadVO{
		BH:  &blockHead1, //区块
		Txs: txs,         //交易明细
	}

	if hashExist {
		return BuildFirstBlock()
	}

	return &bhvo, nil
}

/*
	构建创始块奖励
*/
func BuildReward(balanceTotal uint64) mining.TxItr {
	//创世块矿工奖励
	baseCoinAddr := keystore.GetCoinbase()
	puk, ok := keystore.GetPukByAddr(baseCoinAddr)
	if !ok {
		return nil
	}

	//构建输入

	vins := make([]mining.Vin, 0)
	vin := mining.Vin{
		Puk:  puk, //公钥
		Sign: nil, //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
	vins = append(vins, vin)

	vouts := make([]mining.Vout, 0)
	vouts = append(vouts, mining.Vout{
		Value:   balanceTotal,           //输出金额 = 实际金额 * 100000000
		Address: keystore.GetCoinbase(), //钱包地址
	})
	base := mining.TxBase{
		Type:       config.Wallet_tx_type_mining, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  1,
		Vin:        vins,
		Vout_total: 1,     //输出交易数量
		Vout:       vouts, //交易输出
		LockHeight: 1,     //
		//		CreateTime: time.Now().Unix(),            //创建时间
	}
	reward := &mining.Tx_reward{
		TxBase: base,
	}

	//给输出签名，防篡改
	for i, one := range reward.Vin {
		for _, key := range keystore.GetAddrAll() {

			puk, ok := keystore.GetPukByAddr(key)
			if !ok {
				return nil
			}

			if bytes.Equal(puk, one.Puk) {
				_, prk, _, err := keystore.GetKeyByAddr(key, config.Wallet_keystore_default_pwd)
				// prk, err := key.GetPriKey(pwd)
				if err != nil {
					return nil
				}
				sign := reward.GetSign(&prk, one.Txid, one.Vout, uint64(i))
				//				sign := pay.GetVoutsSign(prk, uint64(i))
				reward.Vin[i].Sign = *sign
			}
		}
	}
	reward.BuildHash()
	return reward
}

func BuildDepositIn() mining.TxItr {
	coinbase := keystore.GetCoinbase()

	puk, _ := keystore.GetPukByAddr(coinbase)

	//首个见证人押金
	// mining.CreateTxDepositIn()
	txin := mining.Tx_deposit_in{
		Puk: puk,
	}
	//创世块矿工奖励
	// vins := make([]mining.Vin, 0)
	// vins = append(vins, mining.Vin{
	// 	Txid: txid,        //UTXO 前一个交易的id
	// 	Vout: 0,           //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
	// 	Puk:  coinbasepuk, //公钥
	// 	// Sign :, //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	// })
	vouts := make([]mining.Vout, 0)
	vouts = append(vouts, mining.Vout{
		Value:   config.Mining_deposit, //  config.Wallet_MDL_mining, //输出金额 = 实际金额 * 100000000
		Address: coinbase,              //钱包地址
	})
	depositInBase := mining.TxBase{
		Type: config.Wallet_tx_type_deposit_in, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		// Vin_total:  uint64(len(vins)),                //输入交易数量
		// Vin:        vins,                             //交易输入
		Vout_total: uint64(len(vouts)), //
		Vout:       vouts,              //
		LockHeight: 1,                  //锁定高度
		//		CreateTime: time.Now().Unix(),                //创建时间
		Payload: []byte("first_witness"),
	}
	txin.TxBase = depositInBase
	txin.BuildHash()
	return &txin
}

/*
	修改上一个区块的Nextblockhash
*/
// func updateNextblockhash(blockOne, blockTwo []byte) error {
// 	bs, err := db.Find(blockOne)
// 	if err != nil {
// 		//TODO 区块未同步完整可以查找不到之前的区块
// 		return err
// 	}
// 	bh, err := mining.ParseBlockHead(bs)
// 	if err != nil {
// 		// fmt.Println("严重错误5", err)
// 		return err
// 	}
// 	if bh.Nextblockhash == nil {
// 		bh.Nextblockhash = make([][]byte, 0)
// 	}
// 	bh.Nextblockhash = append(bh.Nextblockhash, blockTwo)
// 	bs, err = bh.Json()
// 	if err != nil {
// 		// fmt.Println("严重错误6", err)
// 		return err
// 	}
// 	return db.Save(bh.Hash, bs)
// }

// /*
// 	构建创世块
// 	创世块生成两个组，一个组一个块
// 	第一个块给3个见证者账户分配初始额度，下一个组见证者投票结果
// 	第二个块
// */
// func BuildFirstBlock() error {
// 	//如果模块未启用，则不执行

// 	if !parseInitFlag() {
// 		// fmt.Println("不是创始节点")
// 		return nil
// 	}
// 	config.InitNode = true

// 	fmt.Println("开始创建创世区块")

// 	err := keystore.CreateKeystore(filepath.Join(config.Wallet_path, config.Wallet_seed), config.Wallet_keystore_default_pwd)
// 	if err != nil {
// 		return err
// 	}
// 	// witness := make([]*keystore.KeyStore, 0)
// 	// seed1 := keystore.NewKeyStore("wallet", config.Wallet_seed)
// 	// if n, _ := seed1.Load(); n <= 0 {
// 	// 	seed1.NewLoad("wallet1_seed", "123456")
// 	// }
// 	// witness = append(witness, seed1)

// 	db.InitDB("wallet/data")

// 	//----------------生成第一个区块-----------------

// 	balanceTotal := uint64(config.Mining_reward)
// 	//构建交易
// 	txHashes := make([][]byte, 0)
// 	txs := make([]mining.Tx_reward, 0)
// 	//创世块矿工奖励
// 	vouts := make([]mining.Vout, 0)
// 	vouts = append(vouts, mining.Vout{
// 		Value:   balanceTotal,           //输出金额 = 实际金额 * 100000000
// 		Address: keystore.GetCoinbase(), //钱包地址
// 	})
// 	base := mining.TxBase{
// 		Type:       config.Wallet_tx_type_mining, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
// 		Vout_total: 1,                            //输出交易数量
// 		Vout:       vouts,                        //交易输出
// 		LockHeight: 1,                            //
// 		//		CreateTime: time.Now().Unix(),            //创建时间
// 	}
// 	reward := mining.Tx_reward{
// 		TxBase: base,
// 	}
// 	txs = append(txs, reward)
// 	reward.BuildHash()
// 	txHashes = append(txHashes, reward.Hash)

// 	//区块头
// 	blockHead1 := mining.BlockHead{
// 		//				Hash              string   //区块头hash
// 		Height:      1, //区块高度(每秒产生一个块高度，也足够使用上千亿年)
// 		GroupHeight: 1, //
// 		//	MerkleRoot        string   //交易默克尔树根hash
// 		//	Previousblockhash string   //上一个区块头hash
// 		//	Nextblockhash     string   //下一个区块头hash
// 		NTx:     uint64(len(txHashes)),  //交易数量
// 		Tx:      txHashes,               //本区块包含的交易id
// 		Time:    time.Now().Unix(),      //unix时间戳
// 		Witness: keystore.GetCoinbase(), //
// 		//				BackupMiner: , //备用矿工选举结果hash
// 	}
// 	blockHead1.BuildMerkleRoot()

// 	// blockHead1.FindNonce(1, make(chan bool, 1))
// 	//	db.Save(blockHead1.BackupMiner, backupMiner1.JSON())
// 	bhbs, _ := blockHead1.Json()
// 	err = db.Save(blockHead1.Hash, bhbs)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("key", "blockHead", hex.EncodeToString(blockHead1.Hash))
// 	fmt.Println("value", "blockHead", string(*bhbs), "\n")

// 	db.Save(config.Key_block_start, &blockHead1.Hash)
// 	//保存到数据库
// 	for _, one := range txs {
// 		one.TxBase.BlockHash = blockHead1.Hash
// 		bs, err := one.Json()
// 		if err != nil {
// 			// fmt.Println("2 json格式化错误", err)
// 			return err
// 		}
// 		err = db.Save(*one.GetHash(), bs)
// 		if err != nil {
// 			return err
// 		}

// 		fmt.Println("key", "tx", hex.EncodeToString(*one.GetHash()))
// 		fmt.Println("value", "tx", string(*bs))
// 	}
// 	// db.SaveBlockHeight(blockHead1.Height, &blockHead1.Hash)
// 	// fmt.Println("创建初始块完成")

// 	//构建第二个区块
// 	BuildSecoundBlock(blockHead1.Hash, reward.Hash)
// 	return nil
// }

// /*
// 	构建第二个区块
// */
// func BuildSecoundBlock(bhash, txid []byte) error {
// 	//构建交易
// 	txHashes := make([][]byte, 0)
// 	txs := make([]mining.Tx_deposit_in, 0)

// 	coinbase := keystore.GetCoinbase()
// 	coinbasepuk, ok := keystore.GetPukByAddr(coinbase)
// 	if !ok {
// 		return errors.New("未找到coinbase地址的公钥")
// 	}
// 	//首个见证人押金
// 	// mining.CreateTxDepositIn()
// 	txin := mining.Tx_deposit_in{}
// 	//创世块矿工奖励
// 	vins := make([]mining.Vin, 0)
// 	vins = append(vins, mining.Vin{
// 		Txid: txid,        //UTXO 前一个交易的id
// 		Vout: 0,           //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
// 		Puk:  coinbasepuk, //公钥
// 		// Sign :, //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
// 	})
// 	vouts := make([]mining.Vout, 0)
// 	vouts = append(vouts, mining.Vout{
// 		Value:   depositInCoin, //输出金额 = 实际金额 * 100000000
// 		Address: coinbase,      //钱包地址
// 	})
// 	vouts = append(vouts, mining.Vout{
// 		Value:   depositInCoin, //输出金额 = 实际金额 * 100000000
// 		Address: coinbase,      //钱包地址
// 	})
// 	depositInBase := mining.TxBase{
// 		Type:       config.Wallet_tx_type_deposit_in, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
// 		Vin_total:  uint64(len(vins)),                //输入交易数量
// 		Vin:        vins,                             //交易输入
// 		Vout_total: uint64(len(vouts)),               //
// 		Vout:       vouts,                            //
// 		LockHeight: 100,                              //锁定高度
// 		//		CreateTime: time.Now().Unix(),                //创建时间
// 	}
// 	txin.TxBase = depositInBase
// 	txs = append(txs, txin)
// 	txin.BuildHash()
// 	txHashes = append(txHashes, txin.Hash)

// 	//区块头
// 	blockHead := mining.BlockHead{
// 		//				Hash              string   //区块头hash
// 		Height:      2, //区块高度(每秒产生一个块高度，也足够使用上千亿年)
// 		GroupHeight: 2, //
// 		//	MerkleRoot        string   //交易默克尔树根hash
// 		Previousblockhash: [][]byte{bhash}, //上一个区块头hash
// 		// Nextblockhash:     [][]byte{bhash},        //下一个区块头hash
// 		NTx:     uint64(len(txHashes)),  //交易数量
// 		Tx:      txHashes,               //本区块包含的交易id
// 		Time:    time.Now().Unix(),      //unix时间戳
// 		Witness: keystore.GetCoinbase(), //
// 		//				BackupMiner: , //备用矿工选举结果hash
// 	}
// 	blockHead.BuildMerkleRoot()

// 	// blockHead.FindNonce(1, make(chan bool, 1))
// 	//	db.Save(blockHead1.BackupMiner, backupMiner1.JSON())
// 	bhbs, _ := blockHead.Json()
// 	err := db.Save(blockHead.Hash, bhbs)
// 	if err != nil {
// 		return err
// 	}
// 	fmt.Println("key", "blockHead", hex.EncodeToString(blockHead.Hash))
// 	fmt.Println("value", "blockHead", string(*bhbs), "\n")

// 	// db.Save(config.Key_block_start, &blockHead1.Hash)

// 	//保存到数据库
// 	for _, one := range txs {
// 		one.TxBase.BlockHash = blockHead.Hash
// 		bs, err := one.Json()
// 		if err != nil {
// 			// fmt.Println("2 json格式化错误", err)
// 			return err
// 		}
// 		err = db.Save(*one.GetHash(), bs)
// 		if err != nil {
// 			return err
// 		}

// 		fmt.Println("key", "tx", hex.EncodeToString(*one.GetHash()))
// 		fmt.Println("value", "tx", string(*bs))
// 	}

// 	//修改上一个区块的Nextblockhash
// 	bs, err := db.Find(config.Key_block_start)
// 	if err != nil {
// 		//TODO 区块未同步完整可以查找不到之前的区块
// 		return err
// 	}
// 	return updateNextblockhash(*bs, blockHead.Hash)

// }
