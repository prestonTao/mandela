package mining

// import (
// 	//"mandela/core/keystore"
// 	// "fmt"
// 	"time"
// )

// func init() {
// 	go func() {
// 		for range time.NewTicker(5 * time.Second).C {
// 			GetIteam()
// 		}
// 	}()
// }
// func GetIteam() []*TxItem {
// 	// txitem := make([]*TxItem, 0)
// 	chain := forks.GetLongChain()
// 	if chain == nil {
// 		return nil
// 	}
// 	// fmt.Println(chain.balance)
// 	//addr := keystore.GetCoinbase()
// 	chain.balance.FindBalance()
// 	bs, _, _ := chain.balance.FindBalanceAll()
// 	// //余额
// 	// for _, one := range bs {
// 	// 	one.Txs.Range(func(k, v interface{}) bool {
// 	// 		item := v.(*TxItem)
// 	// 		fmt.Printf("xxx %+v", item)
// 	// 		txitem = append(txitem, item)
// 	// 		return true
// 	// 	})
// 	// }
// 	// //冻结
// 	// for _, one := range bfs {
// 	// 	one.Txs.Range(func(k, v interface{}) bool {
// 	// 		item := v.(*TxItem)
// 	// 		fmt.Println("xxx", item)
// 	// 		return true
// 	// 	})
// 	// }
// 	// fmt.Printf("xxx item %+v", bs)
// 	return bs
// }
