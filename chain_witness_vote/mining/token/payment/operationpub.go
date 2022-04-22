package payment

import (
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/keystore/pubstore"
	"mandela/core/utils/crypto"
	"errors"
)

//创建token交易
func CreateTokenPayPub(pwd, seed string, txid string, items, tokenTxItems []*mining.TxItem,
	addr *crypto.AddressCoin, amount, gas, frozenHeight uint64, comment string) (mining.TxItr, error) {
	keystore, err := pubstore.GetPubStore(pwd, seed)
	if err != nil {
		return nil, err
	}
	// start := time.Now()
	if len(tokenTxItems) == 0 {
		return nil, errors.New("token items empty")
	}

	//---------------------开始构建token的交易----------------------
	//发布token的交易id
	// txidBs, err := hex.DecodeString(txid)
	// if err != nil {
	// 	return nil, config.ERROR_params_fail
	// }

	var commentbs []byte
	if comment != "" {
		commentbs = []byte(comment)
	}

	// srcAddrStr := ""
	// if srcAddr != nil {
	// 	srcAddrStr = srcAddr.B58String()
	// }
	// tokenTotal, tokenTxItems := token.GetReadyPayToken(txid, srcAddrStr, amount)
	// if tokenTotal < amount {
	// 	return nil, config.ERROR_token_not_enough
	// }
	var tokenTotal uint64
	tokenVins := make([]*mining.Vin, 0)
	for _, item := range tokenTxItems {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
		vin := mining.Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		tokenVins = append(tokenVins, &vin)
		tokenTotal = tokenTotal + item.Value
	}
	if tokenTotal < amount {
		return nil, config.ERROR_token_not_enough
	}
	//构建交易输出
	tokenVouts := make([]*mining.Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	tokenVout := mining.Vout{
		Value:        amount,       //输出金额 = 实际金额 * 100000000
		Address:      *addr,        //钱包地址
		FrozenHeight: frozenHeight, //
	}
	tokenVouts = append(tokenVouts, &tokenVout)
	//找零
	if tokenTotal > amount {
		tokenVout := mining.Vout{
			Value:   tokenTotal - amount,   //输出金额 = 实际金额 * 100000000
			Address: *tokenTxItems[0].Addr, //keystore.GetAddr()[0], //钱包地址
		}
		tokenVouts = append(tokenVouts, &tokenVout)
	}

	//---------------------开始构建主链上的交易----------------------
	//查找余额
	vins := make([]*mining.Vin, 0)
	chain := mining.GetLongChain() // forks.GetLongChain()
	// total, items := chain.GetBalance().BuildPayVin("", gas)

	var total uint64
	for _, item := range items {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
		vin := mining.Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)
		total = total + item.Value
	}

	//构建交易输出
	vouts := make([]*mining.Vout, 0)

	//找零
	if total > gas {
		vout := mining.Vout{
			Value:   total - gas,    //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr, //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	_, block := chain.GetLastBlock()
	var txin *TxToken
	for i := uint64(0); i < 10000; i++ {
		//
		base := mining.TxBase{
			Type:       Wallet_tx_class,                                //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                      //
			// CreateTime: time.Now().Unix(),      //创建时间
		}
		txin = &TxToken{
			TxBase:           base,
			Token_Vin_total:  uint64(len(tokenVins)),  //输入交易数量
			Token_Vin:        tokenVins,               //交易输入
			Token_Vout_total: uint64(len(tokenVouts)), //输出交易数量
			Token_Vout:       tokenVouts,              //交易输出
			// Token_publish_txid: txidBs,                  //
		}

		// txin.MergeVout()
		// txin.MergeTokenVout()

		//给token交易签名 给输出签名，防篡改
		// for i, one := range txin.Token_Vin {
		// 	_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	sign := txin.GetTokenSign(&prk, one.Txid, one.Vout, uint64(i))
		// 	txin.Token_Vin[i].Sign = *sign
		// }

		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			txin.Vin[i].Sign = *sign
		}

		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}
	return txin, nil
}

//创建token多个交易
func CreateTokenPayPubs(pwd, seed string, txid string, items, tokenTxItems []*mining.TxItem,
	address []mining.PayNumber, gas uint64, comment string) (mining.TxItr, error) {
	keystore, err := pubstore.GetPubStore(pwd, seed)
	if err != nil {
		return nil, err
	}
	// start := time.Now()
	if len(tokenTxItems) == 0 {
		return nil, errors.New("token items empty")
	}

	//发布token的交易id
	// txidBs, err := hex.DecodeString(txid)
	// if err != nil {
	// 	return nil, config.ERROR_params_fail
	// }

	//---------------------开始构建token的交易----------------------
	var amount uint64
	for _, one := range address {
		amount += one.Amount
	}

	// tokenTotal, tokenTxItems := token.GetReadyPayToken(txid, srcAddrStr, amount)
	// if tokenTotal < amount {
	// 	return nil, config.ERROR_token_not_enough
	// }

	var tokenTotal uint64
	tokenVins := make([]*mining.Vin, 0)
	for _, item := range tokenTxItems {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
		vin := mining.Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		tokenTotal = tokenTotal + item.Value
		tokenVins = append(tokenVins, &vin)
	}
	if tokenTotal < amount {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	//构建交易输出
	tokenVouts := make([]*mining.Vout, 0)
	for _, one := range address {
		vout := mining.Vout{
			Value:        one.Amount,       //输出金额 = 实际金额 * 100000000
			Address:      one.Address,      //钱包地址
			FrozenHeight: one.FrozenHeight, //
		}
		tokenVouts = append(tokenVouts, &vout)
	}
	//找零
	if tokenTotal > amount {
		tokenVout := mining.Vout{
			Value:   tokenTotal - amount,   //输出金额 = 实际金额 * 100000000
			Address: *tokenTxItems[0].Addr, //钱包地址
		}
		tokenVouts = append(tokenVouts, &tokenVout)
	}

	//---------------------开始构建主链上的交易----------------------
	//查找余额
	vins := make([]*mining.Vin, 0)
	chain := mining.GetLongChain() // forks.GetLongChain()
	// total, items := chain.GetBalance().BuildPayVin("", gas)

	// if total < gas {
	// 	//资金不够
	// 	return nil, config.ERROR_not_enough // errors.New("余额不足")
	// }
	var total uint64
	for _, item := range items {
		puk, ok := keystore.GetPukByAddr(*item.Addr)
		if !ok {
			return nil, config.ERROR_public_key_not_exist
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))
		vin := mining.Vin{
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)
		total = total + item.Value
	}
	if total < gas {
		//资金不够
		return nil, config.ERROR_not_enough // errors.New("余额不足")
	}
	//构建交易输出
	vouts := make([]*mining.Vout, 0)

	//找零
	if total > gas {
		vout := mining.Vout{
			Value:   total - gas,    //输出金额 = 实际金额 * 100000000
			Address: *items[0].Addr, //keystore.GetAddr()[0], //钱包地址
		}
		vouts = append(vouts, &vout)
	}

	commentbs := []byte{}
	if comment != "" {
		commentbs = []byte(comment)
	}

	_, block := chain.GetLastBlock()
	var txin *TxToken
	for i := uint64(0); i < 10000; i++ {
		//
		base := mining.TxBase{
			Type:       Wallet_tx_class,                                //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  uint64(len(vins)),                              //输入交易数量
			Vin:        vins,                                           //交易输入
			Vout_total: uint64(len(vouts)),                             //输出交易数量
			Vout:       vouts,                                          //
			Gas:        gas,                                            //交易手续费
			LockHeight: block.Height + config.Wallet_tx_lockHeight + i, //锁定高度
			Payload:    commentbs,                                      //
			// CreateTime: time.Now().Unix(),      //创建时间
		}
		txin = &TxToken{
			TxBase:           base,
			Token_Vin_total:  uint64(len(tokenVins)),  //输入交易数量
			Token_Vin:        tokenVins,               //交易输入
			Token_Vout_total: uint64(len(tokenVouts)), //输出交易数量
			Token_Vout:       tokenVouts,              //交易输出
			// Token_publish_txid: txidBs,                  //
		}

		// txin.MergeVout()
		// txin.MergeTokenVout()

		//给token交易签名 给输出签名，防篡改
		// for i, one := range txin.Token_Vin {
		// 	_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
		// 	if err != nil {
		// 		return nil, err
		// 	}
		// 	sign := txin.GetTokenSign(&prk, one.Txid, one.Vout, uint64(i))
		// 	txin.Token_Vin[i].Sign = *sign
		// }

		//给输出签名，防篡改
		for i, one := range txin.Vin {
			_, prk, err := keystore.GetKeyByPuk(one.Puk, pwd)
			if err != nil {
				return nil, err
			}

			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))

			sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			txin.Vin[i].Sign = *sign
		}

		txin.BuildHash()
		if txin.CheckHashExist() {
			txin = nil
			continue
		} else {
			break
		}
	}

	//把txitem冻结起来
	// token.FrozenToken(txid, tokenTxItems, txin)

	// chain.GetBalance().Frozen(items, txin)

	// mining.AddTx(txin)

	return txin, nil
}
