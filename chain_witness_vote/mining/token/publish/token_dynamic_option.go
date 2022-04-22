package publish

import (
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/utils/crypto"
)

const (
	Wallet_tx_class = config.Wallet_tx_type_token_publish
)

/*
	发布一种Token
	@addr    *crypto.AddressCoin    收款地址
	@amount    uint64    转账金额
	@gas    uint64    手续费
	@pwd    string    支付密码
	@name    string    Token名称全称
	@symbol    string    Token单位，符号
	@supply    uint64    发行总量
    @owner    crypto.AddressCoin    所有者
*/
func PublishToken(srcAddr, addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string,
	name, symbol string, supply uint64, owner crypto.AddressCoin) (mining.TxItr, error) {

	//缴纳押金注册一个名称
	txItr, err := mining.GetLongChain().GetBalance().BuildOtherTx(Wallet_tx_class,
		srcAddr, addr, amount, gas, frozenHeight, pwd, comment, name, symbol, supply, owner)
	if err != nil {
		// fmt.Println("缴纳域名押金失败", err)
	} else {
		// fmt.Println("缴纳域名押金完成")
	}
	return txItr, err
}
