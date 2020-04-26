package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	_ "mandela/chain_witness_vote/mining/tx_name_in"
	_ "mandela/chain_witness_vote/mining/tx_name_out"
	"mandela/config"
	"encoding/hex"
	"fmt"
	"path/filepath"

	"gopkg.in/gookit/color.v1"
)

func main() {
	// find("D:/workspaces/go/src/mandela/example/peer_super/wallet/data")
	path := filepath.Join("../example/peer_root/wallet", "data")
	// find(path)
	findNextBlock(path)
	fmt.Println("finish!")
}

var tempBlockHeight = uint64(1000)

// var maxBlockHeight = uint64(99999999999)

func findSomeBlock(dir string) {
	nums := []uint64{}
	for i := uint64(1); i < tempBlockHeight; i++ {
		nums = append(nums, i)
	}

	db.InitDB(dir)
	beforBlockHash, err := db.Find(config.Key_block_start)
	if err != nil {
		fmt.Println("111 查询起始块id错误", err)
		return
	}
	maxBlock := uint64(0)
	for _, one := range nums {
		if one > maxBlock {
			maxBlock = one
		}
	}

	for i := uint64(1); i <= maxBlock; i++ {
		bs, err := db.Find(*beforBlockHash)
		if err != nil {
			fmt.Println("查询第", i, "个块错误", err)
			return
		}
		bh, err := mining.ParseBlockHead(bs)
		if err != nil {
			fmt.Println("解析第", i, "个块错误", err)
			return
		}
		beforBlockHash = &bh.Nextblockhash
		isPrint := false
		for _, one := range nums {
			if one == i {
				isPrint = true
				break
			}
		}
		if isPrint {
			fmt.Println("第", i, "个块 ----------------------------------\n",
				hex.EncodeToString(bh.Hash), "\n", string(*bs), "\n")
			// for _, one := range bh.Nextblockhash {
			// 	fmt.Println("下一个块hash", hex.EncodeToString(one))
			// }
			fmt.Println("下一个块hash", hex.EncodeToString(bh.Nextblockhash))

			for _, one := range bh.Tx {
				tx, err := db.Find(one)
				if err != nil {
					fmt.Println("查询第", i, "个块的交易错误", err)
					return
				}
				txBase, err := mining.ParseTxBase(tx)
				if err != nil {
					fmt.Println("解析第", i, "个块的交易错误", err)
					return
				}

				txid := txBase.GetHash()
				//				if txBase.Class() == config.Wallet_tx_type_deposit_in {
				//					deposit := txBase.(*mining.Tx_deposit_in)
				//					txid = deposit.Hash
				//				}
				fmt.Println(string(hex.EncodeToString(*txid)), "\n", string(*tx), "\n")
			}
		}
	}
}

func findNextBlock(dir string) {

	db.InitDB(dir)
	beforBlockHash, err := db.Find(config.Key_block_start)
	if err != nil {
		fmt.Println("111 查询起始块id错误", err)
		return
	}

	for beforBlockHash != nil {
		bs, err := db.Find(*beforBlockHash)
		if err != nil {
			fmt.Println("查询第", "个块错误", err)
			return
		}
		bh, err := mining.ParseBlockHead(bs)
		if err != nil {

			fmt.Println("解析第", "个块错误", err)
			fmt.Println(string(*bs))
			return
		}
		if bh.Nextblockhash == nil {
			fmt.Println("第", bh.Height, "个块 -----------------------------------\n",
				hex.EncodeToString(bh.Hash), "\n", string(*bs), "\nnext区块个数")
		} else {
			fmt.Println("第", bh.Height, "个块 -----------------------------------\n",
				hex.EncodeToString(bh.Hash), "\n", string(*bs), "\nnext区块个数", len(bh.Nextblockhash))
		}
		// for _, one := range bh.Nextblockhash {
		// 	fmt.Println("下一个块hash", hex.EncodeToString(one))
		// }
		fmt.Println("下一个块hash", hex.EncodeToString(bh.Nextblockhash))
		for _, one := range bh.Tx {
			tx, err := db.Find(one)
			if err != nil {
				fmt.Println("查询第", bh.Height, "个块的交易错误", err)
				panic("error:查询交易错误")
				return
			}
			txBase, err := mining.ParseTxBase(tx)
			if err != nil {
				fmt.Println("解析第", bh.Height, "个块的交易错误", err)
				panic("error:解析交易错误")
				return
			}
			fmt.Println("##########start#############")
			color.Green.Printf("txBase:%+v\n", txBase)
			vt := txBase.GetVout()
			for voutIndex, vout := range *vt {
				txItem := mining.TxItem{
					Addr:     &vout.Address,
					Value:    vout.Value,        //余额
					Txid:     *txBase.GetHash(), //交易id
					OutIndex: uint64(voutIndex), //交易输出index，从0开始
					Height:   bh.Height,         //
				}
				color.Red.Printf("txItem:%+v\n", txItem)
			}
			fmt.Println("##########end#############")
			//panic("end")
			txid := txBase.GetHash()
			//				if txBase.Class() == config.Wallet_tx_type_deposit_in {
			//					deposit := txBase.(*mining.Tx_deposit_in)
			//					txid = deposit.Hash
			//				}
			fmt.Println(string(hex.EncodeToString(*txid)), "\n", string(*tx), "\n")
		}

		if bh.Nextblockhash != nil {
			beforBlockHash = &bh.Nextblockhash
		} else {
			beforBlockHash = nil
		}
	}
}
