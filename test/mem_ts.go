package main

import (
	"fmt"
	"time"
	"mandela/wallet/mining"
)

func main() {
	example2()
}

func example2() {
	bss := make([][]byte, 0)
	for i := 0; i < 9500000; i++ {
		id := []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks")
		bss = append(bss, id)
	}
	fmt.Println("end")
	time.Sleep(time.Minute)
}

func example1() {
	bhs := make([]mining.BlockHead, 0)
	for i := 0; i < 9500000; i++ {
		bh := mining.BlockHead{
			Hash:              []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"), //区块头hash
			Height:            1,                                        //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
			GroupHeight:       1,                                        //矿工组高度
			Previousblockhash: []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"), //上一个区块头hash
			Nextblockhash:     []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"), //下一个区块头hash
			NTx:               1,                                        //交易数量
			MerkleRoot:        []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"), //交易默克尔树根hash
			Tx: [][]byte{[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"),
				[]byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks")}, //本区块包含的交易id
			Time:        1,                                        //unix时间戳
			BackupMiner: []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"), //备用矿工选举结果hash
			DepositId:   []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"), //押金交易id
			Witness:     []byte("12FRzz2xrVtEm9cwzgFArrLE7VA7ks"), //此块见证人地址
		}
		bhs = append(bhs, bh)
	}
	fmt.Println("ok")
	time.Sleep(time.Minute)
	fmt.Println("end")
}
