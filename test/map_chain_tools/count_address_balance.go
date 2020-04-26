package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/utils/crypto"
	"mandela/rpc"
	"mandela/test/map_chain_tools/sqlite"
	"encoding/hex"
	"path/filepath"
	"time"
)

func main() {
	Start()

}

/*
	统计链上所有地址余额
	保存在数据库里面
*/
func Start() {
	sqlite.Init()
	path := filepath.Join("oldwallet", "data")
	CountAddrBalance(path)

	//等待3秒钟再关闭，让sqlite数据库处理完
	time.Sleep(time.Second * 3)
}

/*
	统计链上所有地址余额
*/
func CountAddrBalance(dir string) {

	db.InitDB(dir)
	beforBlockHash, err := db.Find(config.Key_block_start)
	if err != nil {
		engine.Log.Info("111 查询起始块id错误 " + err.Error())
		return
	}

	balanceTotal := uint64(0)

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
		// *bs, _ = json.Marshal(bhvo)
		// engine.Log.Info(string(*bs))

		//计算是否跳过了组
		intervalGroup := bhvo.GroupHeight - beforGroupHeight
		if intervalGroup > 1 {
			// engine.Log.Warn("跳过了组高度 %d", intervalGroup)
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
				panic("error:解析交易错误")
				return
			}

			// txid := txBase.GetHash()
			// engine.Log.Info("交易id " + string(hex.EncodeToString(*txid)))

			// itr := txBase.GetVOJSON()
			// bs, _ := json.Marshal(itr)
			// engine.Log.Info(string(bs))

			//如果是区块奖励，则计算奖励总和
			if txBase.Class() == config.Wallet_tx_type_mining {
				rewardTotal := uint64(0)
				for _, one := range *txBase.GetVout() {
					rewardTotal += one.Value
				}
				engine.Log.Info("区块奖励 %d", rewardTotal)
			}

			//计算所有交易中的地址余额
			if txBase.Class() != config.Wallet_tx_type_mining {
				for _, one := range *txBase.GetVin() {
					addrCoin := crypto.BuildAddr(config.AddrPre, one.Puk)
					addr := addrCoin.B58String()
					balance, _ := new(sqlite.Balance).FindByAddr(addr)
					if balance == nil {
						panic("找不到这个地址的记录" + addr)
						return
					}
					tx, err := db.Find(one.Txid)
					if err != nil {
						// engine.Log.Info("查询第 %d 个块的交易错误"+err.Error(), bh.Height)
						panic("error:查询交易错误2")
						return
					}
					txBase, err := mining.ParseTxBase(tx)
					if err != nil {
						// engine.Log.Info("解析第 %d 个块的交易错误"+err.Error(), bh.Height)
						// fmt.Println("解析第", bh.Height, "个块的交易错误", err)
						panic("error:解析交易错误2")
						return
					}
					vout := (*txBase.GetVout())[one.Vout]
					engine.Log.Info("修改余额 %s %d %d", addr, balance.Balance, vout.Value)
					balanceTotal = balanceTotal - vout.Value
					err = new(sqlite.Balance).UpdateBalance(addr, balance.Balance-vout.Value)
					if err != nil {
						engine.Log.Info("修改余额错误 " + err.Error())
						panic(err)
						return
					}
				}
			}

			//计算输出中的余额
			for _, one := range *txBase.GetVout() {
				balanceTotal = balanceTotal + one.Value
				addr := one.Address.B58String()
				balance, _ := new(sqlite.Balance).FindByAddr(addr)
				if balance == nil {
					//新地址，添加记录
					err := new(sqlite.Balance).Add(addr, one.Value)
					if err != nil {
						engine.Log.Info("添加新地址错误 " + err.Error())
						panic(err)
						return
					}
					continue
				}
				err := new(sqlite.Balance).UpdateBalance(addr, one.Value+balance.Balance)
				if err != nil {
					engine.Log.Info("2修改余额错误 " + err.Error())
					panic(err)
					return
				}
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

		// engine.Log.Info("下一个块hash %s \n", hex.EncodeToString(bh.Nextblockhash))

		if bh.Nextblockhash != nil {
			beforBlockHash = &bh.Nextblockhash
		} else {
			beforBlockHash = nil
		}
	}
	engine.Log.Info("链上共流通币数量：%d", balanceTotal)

}
