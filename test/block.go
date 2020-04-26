package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	//	BuildMerkleRoot()
	BuildFirstBlock()
}

/*
	计算所有交易的默克尔树根
*/
func BuildMerkleRoot() {
	ids := [][]byte{
		[]byte("W1aLWC4unTJZhSFc4VNLFsazAJ1PyTocV7agmteQDL3J3N"),
		[]byte("W1gfVGa52yUJ4Gws4TiA9YbwGP8qCGgaYeeT8APjSiNk6U"),
		[]byte("W1j9RJ1xYHaoAuRk2HGBrVA82njoxFAoctYKQMH43k8hXu"),
		[]byte("W1atFt7bJ5Ubk4MXuV5GfsEYE7srWXR51exDgUEJcVr5fZ"),
		[]byte("W1n9XtbLAjRsh9sr2kbwfkfy3VGenyhazbHJwrEYsnDZ8M"),
	}

	root := mining.BuildMerkleRoot(ids)

	fmt.Println(root)
}

/*
	构建创世块
	创世块生成两个组，一个组一个块
	第一个块给3个见证者账户分配初始额度，下一个组见证者投票结果
	第二个块
*/
func BuildFirstBlock() {

	witness := make([]*keystore.KeyStore, 0)

	seed1 := keystore.NewKeyStore("wallet1", config.Wallet_seed)
	if n, _ := seed1.Load(); n <= 0 {
		seed1.NewLoad("wallet1_seed", "123456")
	}
	witness = append(witness, seed1)

	seed2 := keystore.NewKeyStore("wallet2", config.Wallet_seed)
	if n, _ := seed2.Load(); n <= 0 {
		seed2.NewLoad("wallet2_seed", "123456")
	}
	witness = append(witness, seed2)

	seed3 := keystore.NewKeyStore("wallet3", config.Wallet_seed)
	if n, _ := seed3.Load(); n <= 0 {
		seed3.NewLoad("wallet3_seed", "123456")
	}
	witness = append(witness, seed3)
	fmt.Println("-------------")

	db.InitDB("data")

	//初始3个矿工地址
	//	miners := []string{
	//		"12FF9pXG5v69Zn7Ft8Wchqf9YKwu4Y",
	//		"12F2qNrh2c15ShtshvR2RxrtnhUMZb",
	//		"12GgYwqvq5wjbnfnfF8B8effSxxFcz",
	//	}

	//----------------生成第一个区块-----------------
	balanceTotal := uint64(3 * mining.Unit)

	now := time.Now().Unix()

	//构建交易
	txs := make([][]byte, 0)

	//给每个见证人地址分配一点余额，用以支付见证人押金
	txReward := make([]mining.Tx_reward, 0)
	for _, one := range witness {
		vout := mining.Vout{
			Value:   balanceTotal,           //输出金额 = 实际金额 * 100000000
			Address: *one.GetAddr()[0].Hash, //钱包地址
		}
		base := mining.TxBase{
			Type: config.Wallet_tx_type_mining,
		}
		tx := mining.Tx_reward{
			TxBase:     base,
			CreateTime: now,
		}
		tx.TxBase.Vout = append(tx.TxBase.Vout, vout)
		tx.BuildHash()
		txs = append(txs, *tx.GetHash())
		txReward = append(txReward, tx)
	}
	//每个地址有两个入账交易
	//	for _, one := range witness {
	//		vout := mining.Vout{
	//			Value:   2 * mining.Unit,        //输出金额 = 实际金额 * 100000000，和第一笔入账金额不一样，是为了保证入账hash值不一样
	//			Address: *one.GetAddr()[0].Hash, //钱包地址
	//		}
	//		base := mining.TxBase{
	//			Type: config.Wallet_tx_type_mining,
	//		}
	//		tx := mining.Tx_reward{
	//			TxBase:     base,
	//			CreateTime: now + 1,
	//		}
	//		tx.TxBase.Vout = append(tx.TxBase.Vout, vout)
	//		tx.BuildHash()
	//		txs = append(txs, *tx.GetHash())
	//		txReward = append(txReward, tx)
	//	}

	//构建备用见证人押金交易
	deposits := make([]mining.Tx_deposit_in, 0)
	//第二组区块的见证人押金
	vouts := make([]mining.Vout, 0)
	vout := mining.Vout{
		Value:   config.Mining_deposit,         //输出金额 = 实际金额 * 100000000
		Address: *witness[0].GetAddr()[0].Hash, //钱包地址
	}
	vouts = append(vouts, vout)
	//余额打入自己的地址
	vout = mining.Vout{
		Value:   balanceTotal - config.Mining_deposit, //输出金额 = 实际金额 * 100000000
		Address: *witness[0].GetAddr()[0].Hash,        //钱包地址
	}
	vouts = append(vouts, vout)
	bs, err := json.Marshal(vouts)
	if err != nil {
		fmt.Println("json格式化错误", err)
		return
	}
	voutSign, err := witness[0].GetAddr()[0].Sign(bs, "123456")
	if err != nil {
		fmt.Println("签名错误111", err)
		return
	}
	sign, err := txReward[0].Sign(&witness[0].GetAddr()[0], "123456")
	if err != nil {
		fmt.Println("签名错误222", err)
		return
	}
	vins := make([]mining.Vin, 0)
	vin := mining.Vin{
		Txid:     *txReward[0].GetHash(),              //UTXO 前一个交易的id
		Vout:     0,                                   //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
		Puk:      witness[0].GetAddr()[0].GetPubKey(), //公钥
		Sign:     *sign,                               //对上一个交易的输出签名
		VoutSign: *voutSign,                           //对本交易的输出签名
	}
	vins = append(vins, vin)
	base := mining.TxBase{
		Type:       config.Wallet_tx_type_deposit_in, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  uint64(len(vins)),                //输入交易数量
		Vin:        vins,                             //交易输入
		Vout_total: uint64(len(vouts)),               //输出交易数量
		Vout:       vouts,                            //
	}
	tx := mining.Tx_deposit_in{
		TxBase: base,
	}
	//	tx := mining.CreateTxDepositIn(&witness[0].GetAddr()[0], txs[0], 0, &txReward[0].Vout[0])
	tx.BuildHash()
	txs = append(txs, *tx.GetHash())
	deposits = append(deposits, tx)

	//第三组区块的见证人押金
	//从第3组块开始，必须要有3个见证人同时出块才是有效块
	for i, one := range witness {
		key := one.GetAddr()[0]
		newvouts := make([]mining.Vout, 0)
		vout := mining.Vout{
			Value:   config.Mining_deposit,  //输出金额 = 实际金额 * 100000000
			Address: *one.GetAddr()[0].Hash, //钱包地址
		}
		newvouts = append(newvouts, vout)
		if i == 0 {
			//余额打入自己的地址
			vout = mining.Vout{
				Value:   1 * mining.Unit,        //输出金额 = 实际金额 * 100000000
				Address: *one.GetAddr()[0].Hash, //钱包地址
			}
			newvouts = append(newvouts, vout)
		} else {
			//余额打入自己的地址
			vout = mining.Vout{
				Value:   2 * mining.Unit,        //输出金额 = 实际金额 * 100000000
				Address: *one.GetAddr()[0].Hash, //钱包地址
			}
			newvouts = append(newvouts, vout)
		}
		bs, err := json.Marshal(newvouts)
		if err != nil {
			fmt.Println("json格式化错误", err)
			return
		}
		voutSign, err := one.GetAddr()[0].Sign(bs, "123456")
		if err != nil {
			fmt.Println("签名错误111", err)
			return
		}

		vins := make([]mining.Vin, 0)
		var sign *[]byte
		if i == 0 {
			sign, err = deposits[0].Sign(&one.GetAddr()[0], "123456")
			if err != nil {
				fmt.Println("签名错误222", err)
				return
			}
			vin := mining.Vin{
				Txid:     *deposits[0].GetHash(),       //UTXO 前一个交易的id
				Vout:     1,                            //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
				Puk:      one.GetAddr()[0].GetPubKey(), //公钥
				Sign:     *sign,                        //对上一个交易的输出签名
				VoutSign: *voutSign,                    //对本交易的输出签名
			}
			vins = append(vins, vin)
		} else {
			sign, err = txReward[i].Sign(&key, "123456")
			if err != nil {
				fmt.Println("签名错误222", err)
				return
			}
			vin := mining.Vin{
				Txid:     *txReward[i].GetHash(), //UTXO 前一个交易的id
				Vout:     0,                      //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
				Puk:      key.GetPubKey(),        //公钥
				Sign:     *sign,                  //对上一个交易的输出签名
				VoutSign: *voutSign,              //对本交易的输出签名
			}
			vins = append(vins, vin)
		}

		base := mining.TxBase{
			Type:       config.Wallet_tx_type_deposit_in, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                //输入交易数量
			Vin:        vins,                             //交易输入
			Vout_total: uint64(len(newvouts)),            //输出交易数量
			Vout:       newvouts,                         //

		}
		tx := mining.Tx_deposit_in{
			TxBase: base,
		}

		//		tx := mining.CreateTxDepositIn(&one.GetAddr()[0], txs[i+3], 0, &txReward[i+3].Vout[0])
		tx.BuildHash()
		txs = append(txs, *tx.GetHash())
		deposits = append(deposits, tx)
	}

	//构建投票结果
	//构建第二组见证人投票结果
	backupMiner1 := mining.BackupMiners{
		Time:   time.Now().Unix(),             //统计时间
		Miners: make([]mining.BackupMiner, 0), //
	}
	miner := mining.BackupMiner{
		Miner: *witness[0].GetAddr()[0].Hash, //矿工地址
		Count: 1,                             //票数
	}
	backupMiner1.Miners = append(backupMiner1.Miners, miner)
	//	//构建第三组见证人投票结果
	//	for _, one := range witness {
	//		miner := mining.BackupMiner{
	//			Miner: *one.GetAddr()[0].Hash, //矿工地址
	//			Count: 1,                      //票数
	//		}
	//		backupMiner1.Miners = append(backupMiner1.Miners, miner)
	//	}

	//区块头
	blockHead1 := mining.BlockHead{
		//				Hash              string   //区块头hash
		Height:      1, //区块高度(每秒产生一个块高度，也足够使用上千亿年)
		GroupHeight: 1, //
		//	MerkleRoot        string   //交易默克尔树根hash
		//	Previousblockhash string   //上一个区块头hash
		//	Nextblockhash     string   //下一个区块头hash
		NTx:         uint64(len(txs)),  //交易数量
		Tx:          txs,               //本区块包含的交易id
		Time:        time.Now().Unix(), //unix时间戳
		Witness:     *witness[0].GetAddr()[0].Hash,
		BackupMiner: utils.Hash_SHA3_256(*backupMiner1.JSON()), //备用矿工选举结果hash
	}

	blockHead1.BuildMerkleRoot()
	blockHead1.BuildHash()
	db.Save(blockHead1.BackupMiner, backupMiner1.JSON())

	fmt.Println(hex.EncodeToString(blockHead1.Hash))
	bhbs, _ := blockHead1.Json()
	//	fmt.Println(err)
	fmt.Println(string(*bhbs), "\n")
	db.Save(blockHead1.Hash, bhbs)

	db.Save(config.Key_block_start, &blockHead1.Hash)

	//保存到数据库
	for _, one := range txReward {
		one.TxBase.BlockHash = blockHead1.Hash
		bs, err := one.Json()
		if err != nil {
			fmt.Println("2 json格式化错误", err)
			return
		}
		db.Save(*one.GetHash(), bs)
		fmt.Println(string(*bs), "\n")
	}

	//保存交易到数据库
	for _, one := range deposits {
		one.TxBase.BlockHash = blockHead1.Hash
		bs, err := one.Json()
		if err != nil {
			fmt.Println("3 json格式化错误", err)
			return
		}
		db.Save(*one.GetHash(), bs)
		fmt.Println(string(*bs), "\n")
	}
	db.SaveBlockHeight(blockHead1.Height, &blockHead1.Hash)

}
