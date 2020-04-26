package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	_ "mandela/chain_witness_vote/mining/tx_name_in"
	_ "mandela/chain_witness_vote/mining/tx_name_out"
	"mandela/config"
	"mandela/core/engine"
	"mandela/rpc"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
)

func main() {
	// find("D:/workspaces/go/src/mandela/example/peer_super/wallet/data")
	path := filepath.Join("wallet", "data")
	// find(path)
	findNextBlock(path)
	engine.Log.Info("finish!")
}

// var tempBlockHeight = uint64(1000)

// // var maxBlockHeight = uint64(99999999999)

// func findSomeBlock(dir string) {
// 	nums := []uint64{}
// 	for i := uint64(1); i < tempBlockHeight; i++ {
// 		nums = append(nums, i)
// 	}

// 	db.InitDB(dir)
// 	beforBlockHash, err := db.Find(config.Key_block_start)
// 	if err != nil {
// 		engine.Log.Info("111 查询起始块id错误 " + err.Error())
// 		// fmt.Println("111 查询起始块id错误", err)
// 		return
// 	}
// 	maxBlock := uint64(0)
// 	for _, one := range nums {
// 		if one > maxBlock {
// 			maxBlock = one
// 		}
// 	}

// 	for i := uint64(1); i <= maxBlock; i++ {
// 		bs, err := db.Find(*beforBlockHash)
// 		if err != nil {
// 			engine.Log.Info("查询第 %d 个块错误"+err.Error(), i)
// 			// engine.Log.Info("111 查询起始块id错误 " + err.Error())
// 			return
// 		}
// 		bh, err := mining.ParseBlockHead(bs)
// 		if err != nil {
// 			engine.Log.Info("查询第 %d 个块错误"+err.Error(), i)
// 			return
// 		}
// 		beforBlockHash = &bh.Nextblockhash
// 		isPrint := false
// 		for _, one := range nums {
// 			if one == i {
// 				isPrint = true
// 				break
// 			}
// 		}
// 		if isPrint {
// 			engine.Log.Info("第 %d 个块 ----------------------------------\n%s\n", i, hex.EncodeToString(bh.Hash))

// 			txs := make([]string, 0)
// 			for _, one := range bh.Tx {
// 				txs = append(txs, hex.EncodeToString(one))
// 			}
// 			bhvo := rpc.BlockHeadVO{
// 				Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
// 				Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
// 				GroupHeight:       bh.GroupHeight,                           //矿工组高度
// 				Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
// 				Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
// 				NTx:               bh.NTx,                                   //交易数量
// 				MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
// 				Tx:                txs,                                      //本区块包含的交易id
// 				Time:              bh.Time,                                  //出块时间，unix时间戳
// 				Witness:           bh.Witness.B58String(),                   //此块见证人地址
// 				Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
// 			}
// 			bs, _ := json.Marshal(bhvo)
// 			engine.Log.Info(string(bs))

// 			// for _, one := range bh.Nextblockhash {
// 			// 	fmt.Println("下一个块hash", hex.EncodeToString(one))
// 			// }
// 			engine.Log.Info("下一个块hash " + hex.EncodeToString(bh.Nextblockhash))

// 			for _, one := range bh.Tx {
// 				tx, err := db.Find(one)
// 				if err != nil {
// 					engine.Log.Info("查询第 %d 个块的交易错误", err, i)
// 					return
// 				}
// 				txBase, err := mining.ParseTxBase(tx)
// 				if err != nil {
// 					engine.Log.Info("解析第 %d 个块的交易错误", err, i)
// 					return
// 				}

// 				txid := txBase.GetHash()
// 				//				if txBase.Class() == config.Wallet_tx_type_deposit_in {
// 				//					deposit := txBase.(*mining.Tx_deposit_in)
// 				//					txid = deposit.Hash
// 				//				}
// 				engine.Log.Info("%s\n%s\n", string(hex.EncodeToString(*txid)), string(*tx))

// 				switch txBase.Class() {
// 				case config.Wallet_tx_type_vote_in:
// 					tx := txBase.(*mining.Tx_vote_in)
// 					engine.Log.Info("%d %s %s", tx.VoteType, hex.EncodeToString((*tx.GetVout())[0].Address), tx.Vote.B58String())
// 				case config.Wallet_tx_type_vote_out:
// 					tx := txBase.(*mining.Tx_vote_out)
// 					engine.Log.Info("%s", hex.EncodeToString((*tx.GetVin())[0].Txid))
// 				}

// 			}
// 		}
// 	}
// }

func findNextBlock(dir string) {

	db.InitDB(dir)
	beforBlockHash, err := db.Find(config.Key_block_start)
	if err != nil {
		engine.Log.Info("111 查询起始块id错误 " + err.Error())
		return
	}

	beforGroupHeight := uint64(0)

	for beforBlockHash != nil {
		bs, err := db.Find(*beforBlockHash)
		if err != nil {
			engine.Log.Info("查询第 个块错误" + err.Error())
			return
		}
		bh, err := mining.ParseBlockHead(bs)
		if err != nil {

			engine.Log.Info("查询第 个块错误" + err.Error())
			// fmt.Println(string(*bs))
			engine.Log.Info(string(*bs))
			return
		}
		if bh.Nextblockhash == nil {
			engine.Log.Info("第%d个块 -----------------------------------\n%s\nnext区块个数", bh.Height,
				hex.EncodeToString(bh.Hash))
		} else {
			engine.Log.Info("第%d个块 -----------------------------------\n%s\nnext区块个数%d", bh.Height,
				hex.EncodeToString(bh.Hash), len(bh.Nextblockhash))
			// fmt.Println("第", bh.Height, "个块 -----------------------------------\n",
			// 	hex.EncodeToString(bh.Hash), "\n", string(*bs), "\nnext区块个数", len(bh.Nextblockhash))
		}

		txs := make([]string, 0)
		for _, one := range bh.Tx {
			txs = append(txs, hex.EncodeToString(one))
		}
		bhvo := rpc.BlockHeadVO{
			Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
			Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       bh.GroupHeight,                           //矿工组高度
			Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
			Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
			NTx:               bh.NTx,                                   //交易数量
			MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
			Tx:                txs,                                      //本区块包含的交易id
			Time:              bh.Time,                                  //出块时间，unix时间戳
			Witness:           bh.Witness.B58String(),                   //此块见证人地址
			Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
		}
		*bs, _ = json.Marshal(bhvo)
		engine.Log.Info(string(*bs))

		//计算是否跳过了组
		intervalGroup := bhvo.GroupHeight - beforGroupHeight
		if intervalGroup > 1 {
			engine.Log.Warn("跳过了组高度 %d", intervalGroup)
		}
		beforGroupHeight = bhvo.GroupHeight

		// for _, one := range bh.Nextblockhash {
		// 	fmt.Println("下一个块hash", hex.EncodeToString(one))
		// }

		for _, one := range bh.Tx {
			tx, err := db.Find(one)
			if err != nil {
				engine.Log.Info("查询第 %d 个块的交易错误"+err.Error(), bh.Height)
				panic("error:查询交易错误")
				return
			}
			txBase, err := mining.ParseTxBase(tx)
			if err != nil {
				engine.Log.Info("解析第 %d 个块的交易错误"+err.Error(), bh.Height)
				// fmt.Println("解析第", bh.Height, "个块的交易错误", err)
				panic("error:解析交易错误")
				return
			}

			txid := txBase.GetHash()
			//				if txBase.Class() == config.Wallet_tx_type_deposit_in {
			//					deposit := txBase.(*mining.Tx_deposit_in)
			//					txid = deposit.Hash
			//				}
			engine.Log.Info("交易id " + string(hex.EncodeToString(*txid)))

			itr := txBase.GetVOJSON()
			bs, _ := json.Marshal(itr)
			engine.Log.Info(string(bs))

			//如果是区块奖励，则计算奖励总和
			if txBase.Class() == config.Wallet_tx_type_mining {
				rewardTotal := uint64(0)
				for _, one := range *txBase.GetVout() {
					rewardTotal += one.Value
				}
				engine.Log.Info("区块奖励 %d", rewardTotal)
			}

			// switch txBase.Class() {
			// case config.Wallet_tx_type_vote_in:
			// 	tx := txBase.(*mining.Tx_vote_in)
			// 	engine.Log.Info("%d %s %s", tx.VoteType, hex.EncodeToString((*tx.GetVout())[0].Address), tx.Vote.B58String())
			// case config.Wallet_tx_type_vote_out:
			// 	tx := txBase.(*mining.Tx_vote_out)
			// 	engine.Log.Info("%s", hex.EncodeToString((*tx.GetVin())[0].Txid))
			// }
		}

		engine.Log.Info("下一个块hash %s \n", hex.EncodeToString(bh.Nextblockhash))

		if bh.Nextblockhash != nil {
			beforBlockHash = &bh.Nextblockhash
		} else {
			beforBlockHash = nil
		}
	}
}
