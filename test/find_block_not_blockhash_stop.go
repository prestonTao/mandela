package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	_ "mandela/chain_witness_vote/mining/token/payment"
	_ "mandela/chain_witness_vote/mining/token/publish"
	_ "mandela/chain_witness_vote/mining/tx_name_in"
	_ "mandela/chain_witness_vote/mining/tx_name_out"
	"mandela/config"
	"mandela/core/engine"
	"mandela/rpc"
	"encoding/hex"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	// find("D:/workspaces/go/src/mandela/example/peer_super/wallet/data")
	path := filepath.Join("wallet", "data")
	// find(path)
	findNextBlock(path)
	engine.Log.Info("finish!")
}

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
			engine.Log.Info("第%d个块 -----------------------------------\n%s\n", bh.Height)
		} else {
			engine.Log.Info("第%d个块 -----------------------------------\n%s\n", bh.Height,
				hex.EncodeToString(bh.Hash))
			// fmt.Println("第", bh.Height, "个块 -----------------------------------\n",
			// 	hex.EncodeToString(bh.Hash), "\n", string(*bs), "\nnext区块个数", len(bh.Nextblockhash))
		}
		engine.Log.Info("交易数量 %d", len(bh.Tx))

		txs := make([]string, 0)
		for _, one := range bh.Tx {
			txs = append(txs, hex.EncodeToString(one))
		}
		bhvo := rpc.BlockHeadVO{
			Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
			Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       bh.GroupHeight,                           //矿工组高度
			GroupHeightGrowth: bh.GroupHeightGrowth,                     //
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
			txBase, err := mining.ParseTxBase(mining.ParseTxClass(one), tx)
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

			// engine.Log.Info("blockhash %v", txBase.GetBlockHash())

			//
			if txBase.GetBlockHash() == nil || len(*txBase.GetBlockHash()) <= 0 {
				engine.Log.Info("blockhash为空")
				panic("blockhash为空")
				return
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
