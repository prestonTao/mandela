package tx_name_in

import (
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils/crypto"
)

/*
	注册域名，缴押金
*/
func NameIn(srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string,
	name string, netIds []nodeStore.AddressNet, addrCoins []crypto.AddressCoin) (mining.TxItr, error) {

	//缴纳押金注册一个名称
	txItr, err := mining.GetLongChain().GetBalance().BuildOtherTx(config.Wallet_tx_type_account,
		srcAddr, addr, amount, gas, frozenHeight, pwd, comment, name, netIds, addrCoins)
	if err != nil {
		// fmt.Println("缴纳域名押金失败", err)
	} else {
		// fmt.Println("缴纳域名押金完成")
	}
	return txItr, err
}
