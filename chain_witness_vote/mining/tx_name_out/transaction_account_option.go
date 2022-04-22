package tx_name_out

import (
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/utils/crypto"
)

/*
	注销域名，退还押金
*/
func NameOut(srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string, name string) (mining.TxItr, error) {

	//缴纳押金注册一个名称
	txItr, err := mining.GetLongChain().GetBalance().BuildOtherTx(config.Wallet_tx_type_account_cancel,
		srcAddr, addr, 0, gas, frozenHeight, pwd, comment, name)
	if err != nil {
		// fmt.Println("退还押金失败", err)
	} else {
		// fmt.Println("退还押金完成")
	}
	return txItr, err
}
