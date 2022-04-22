package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/token/payment"
	"mandela/config"
	"mandela/core/utils"
	"encoding/json"
	"fmt"
	"sync"
)

var TagTxSubmiting string = "TxSub_"

func init() {
	tpc := new(payment.TokenPublishController)
	tpc.ActiveVoutIndex = new(sync.Map)
	mining.RegisterTransaction(config.Wallet_tx_type_account, tpc)

	err := db.InitDB(config.DB_path)
	if err != nil {
		fmt.Println("init db error:", err.Error())
		panic(err)
	}

}
func main() {
	db.PrintAll()

	// findAll()
}

func findAll() {
	keys, values, _ := db.FindPrefixKeyAll([]byte(TagTxSubmiting))
	fmt.Println(len(keys), len(values))

	outBs := make([]byte, 0)
	for _, one := range values {
		// fmt.Println(string(one))
		txItr, err := mining.ParseTxBase(0, &one)
		panicError(err)
		txVO := txItr.GetVOJSON()
		// str, _ := json.MarshalToString(txVO)
		txBs, err := json.Marshal(txVO)
		panicError(err)

		outBs = append(outBs, txBs...)
		outBs = append(outBs, []byte("\n")...)

		// outStr = outStr + str + "\n"
	}

	// outBs := []byte(outStr)
	utils.SaveFile("tx.txt", &outBs)
}
func panicError(err error) {
	if err != nil {
		panic(err)
	}
}
