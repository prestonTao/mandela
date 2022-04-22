package token

import (
	"encoding/hex"

	//"mandela/core/utils/crypto"
	"mandela/chain_witness_vote/mining"
	"fmt"
	// "time"
)

// func init() {
// 	go func() {
// 		for range time.NewTicker(5 * time.Second).C {
// 			GetTokenTxItem()
// 		}
// 	}()
// }

//获取token item
func GetTokenTxItem() []*mining.TxItem {
	txid := "0a0000000000000025dfdd4de9055749ba209012252fa0437f77904ee1eca5690403769ee08204ee"
	txidBs, _ := hex.DecodeString(txid)
	amount := uint64(100)
	tokenTotal, tokenTxItems := GetReadyPayToken(txidBs, nil, amount)
	fmt.Printf("xxxx tokenNum %v tokenitem %v", tokenTotal, tokenTxItems)
	return tokenTxItems
}
