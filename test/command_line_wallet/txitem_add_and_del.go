package main

import (
	"mandela/core/utils"
	"mandela/sqlite3_db"
	"fmt"
	"strconv"
	"time"
)

func main() {
	start()
}

func start() {
	sqlite3_db.Init()
	txItems := make([]sqlite3_db.TxItem, 0)
	// count := 1
	countEnd := 99
	for i := 0; i < countEnd; i++ {

		key := append(utils.Hash_SHA3_256([]byte(strconv.Itoa(i))), utils.Uint64ToBytes(uint64(i))...)
		txOne := sqlite3_db.TxItem{
			Key:          key,
			Addr:         utils.Hash_SHA3_256([]byte(strconv.Itoa(i))),
			Value:        uint64(i),
			Txid:         utils.Hash_SHA3_256([]byte(strconv.Itoa(i))),
			VoutIndex:    uint64(i),
			Height:       uint64(i),
			VoteType:     uint16(i),
			FrozenHeight: uint64(i),
			LockupHeight: uint64(i),
			// Status      :
		}
		txItems = append(txItems, txOne)
	}
	start := time.Now()
	err := new(sqlite3_db.TxItem).AddTxItems(&txItems)
	fmt.Println(time.Now().Sub(start), err)

	keys := make([][]byte, 0)
	for i := countEnd; i >= 0; i-- {
		key := append(utils.Hash_SHA3_256([]byte(strconv.Itoa(i))), utils.Uint64ToBytes(uint64(i))...)
		keys = append(keys, key)
		// new(sqlite3_db.TxItem).RemoveOne(utils.Hash_SHA3_256([]byte(strconv.Itoa(i))), utils.Hash_SHA3_256([]byte(strconv.Itoa(i))), uint64(i))
		// new(sqlite3_db.TxItem).RemoveOneKey(key)

		// fmt.Println(err)
	}
	start = time.Now()
	new(sqlite3_db.TxItem).RemoveMoreKey(keys)
	fmt.Println(time.Now().Sub(start))
}
