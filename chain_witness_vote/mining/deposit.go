/*
	管理缴纳押金
*/
package mining

//import (
//	"fmt"
//	"mandela/config"
//	"mandela/chain_witness_vote/keystore"
//)

/*
	交押金
*/
//func Deposit() {
//	if !config.Miner {
//		return
//	}

//	addr, err := keystore.GetCoinbase()
//	if err != nil {
//		fmt.Println("111获取矿工地址失败", err)
//		return
//	}

//	//交了押金就不再交了
//	if forks.GetLongChain().witnessBackup.hashWitness(addr.Hash) {
//		fmt.Println("交了押金就不用再交了")
//		return
//	}

//	//缴纳备用见证人押金交易
//	err = forks.GetLongChain().witnessChain.DepositIn(config.Mining_deposit)
//	if err != nil {
//		fmt.Println("缴纳押金失败", err)
//	}
//	fmt.Println("缴纳押金完成")

//	//判断自己的押金交易是否被打包到块中，已打包则给自己投票
//	//	chain.CheckVote(addr)
//	//	fmt.Println("投票完成")
//}
