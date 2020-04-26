package kstore

// import (
// 	"bytes"
// 	"errors"
// 	"mandela/chain_witness_vote/mining"
// 	"mandela/config"
// 	"mandela/core/keystore"
// 	"mandela/core/utils/crypto"
// 	"strconv"
// )

// /*
// 	创建一个转款交易
// 	@params height 当前区块高度 item 交易item address 接受者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
// */
// func CreateTxPay(height uint64, items []*mining.TxItem, address *crypto.AddressCoin, amount, gas uint64, pwd, comment string) (*mining.Tx_Pay, error) {
// 	if len(items) == 0 {
// 		return nil, errors.New("余额不足")
// 	}
// 	// chain := forks.GetLongChain()
// 	// _, block := chain.GetLastBlock()
// 	// //查找余额
// 	vins := make([]mining.Vin, 0)
// 	total := uint64(0)
// 	// keys := keystore.GetAddrAll()
// 	// for _, one := range keys {

// 	// 	bas, err := chain.balance.FindBalance(&one)
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}

// 	// 	for _, two := range bas {
// 	// 		two.Txs.Range(func(k, v interface{}) bool {
// 	// 			item := v.(*TxItem)
// 	for _, item := range items {
// 		puk, ok := keystore.GetPukByAddr(*item.Addr)
// 		if !ok {
// 			continue
// 		}
// 		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

// 		vin := mining.Vin{
// 			Txid: item.Txid,     //UTXO 前一个交易的id
// 			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 			Puk:  puk,           //公钥
// 			//					Sign: *sign,           //签名
// 		}
// 		vins = append(vins, vin)

// 		total = total + item.Value
// 		if total >= amount+gas {
// 			//return false
// 			break
// 		}

// 	}
// 	// if total >= amount+gas {
// 	// 	break
// 	// }
// 	//}
// 	//}

// 	if total < amount+gas {
// 		//资金不够
// 		return nil, errors.New("余额不足")
// 	}

// 	//构建交易输出
// 	vouts := make([]mining.Vout, 0)
// 	vout := mining.Vout{
// 		Value:   amount,   //输出金额 = 实际金额 * 100000000
// 		Address: *address, //钱包地址
// 	}
// 	vouts = append(vouts, vout)
// 	//检查押金是否刚刚好，多了的转账给自己
// 	//TODO 将剩余款项转入新的地址，保证资金安全
// 	if total > amount+gas {
// 		vout := mining.Vout{
// 			Value:   total - amount - gas,  //输出金额 = 实际金额 * 100000000
// 			Address: keystore.GetAddr()[0], //钱包地址
// 		}
// 		vouts = append(vouts, vout)
// 	}
// 	var pay *mining.Tx_Pay
// 	//for i := uint64(0); i < 10000; i++ {
// 	//没有输出
// 	base := mining.TxBase{
// 		Type:       config.Wallet_tx_type_pay, //交易类型
// 		Vin_total:  uint64(len(vins)),         //输入交易数量
// 		Vin:        vins,                      //交易输入
// 		Vout_total: uint64(len(vouts)),        //输出交易数量
// 		Vout:       vouts,                     //交易输出
// 		Gas:        gas,                       //交易手续费
// 		//LockHeight: block.Height + 100 + i,    //锁定高度
// 		LockHeight: height + 100, //锁定高度
// 		//		CreateTime: time.Now().Unix(),         //创建时间
// 	}
// 	pay = &mining.Tx_Pay{
// 		TxBase: base,
// 	}

// 	//给输出签名，防篡改
// 	for i, one := range pay.Vin {
// 		for _, key := range keystore.GetAddrAll() {

// 			puk, ok := keystore.GetPukByAddr(key)
// 			if !ok {
// 				return nil, errors.New("未找到公钥")
// 			}

// 			if bytes.Equal(puk, one.Puk) {
// 				_, prk, _, err := keystore.GetKeyByAddr(key, pwd)
// 				// prk, err := key.GetPriKey(pwd)
// 				if err != nil {
// 					return nil, err
// 				}
// 				sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
// 				//				sign := pay.GetVoutsSign(prk, uint64(i))
// 				pay.Vin[i].Sign = *sign
// 			}
// 		}
// 	}
// 	pay.BuildHash()
// 	// if pay.CheckHashExist() {
// 	// 	pay = nil
// 	// 	continue
// 	// } else {
// 	// 	break
// 	// }
// 	//}
// 	return pay, nil
// }

// /*
// 	创建多个转款交易
// 	@params height 当前区块高度 item 交易item address 接受者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
// */
// func CreateTxsPay(height uint64, items []*mining.TxItem, address []mining.PayNumber, gas uint64, pwd, comment string) (*mining.Tx_Pay, error) {
// 	if len(items) == 0 {
// 		return nil, errors.New("余额不足")
// 	}
// 	// chain := forks.GetLongChain()
// 	// _, block := chain.GetLastBlock()
// 	// //查找余额
// 	vins := make([]mining.Vin, 0)
// 	total := uint64(0)
// 	amount := uint64(0)
// 	for _, one := range address {
// 		amount += one.Amount
// 	}
// 	// keys := keystore.GetAddrAll()
// 	// for _, one := range keys {

// 	// 	bas, err := chain.balance.FindBalance(&one)
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}

// 	// 	for _, two := range bas {
// 	// 		two.Txs.Range(func(k, v interface{}) bool {
// 	// 			item := v.(*TxItem)
// 	for _, item := range items {
// 		puk, ok := keystore.GetPukByAddr(*item.Addr)
// 		if !ok {
// 			continue
// 		}
// 		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

// 		vin := mining.Vin{
// 			Txid: item.Txid,     //UTXO 前一个交易的id
// 			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 			Puk:  puk,           //公钥
// 			//					Sign: *sign,           //签名
// 		}
// 		vins = append(vins, vin)

// 		total = total + item.Value
// 		if total >= amount+gas {
// 			//return false
// 			break
// 		}

// 	}
// 	// if total >= amount+gas {
// 	// 	break
// 	// }
// 	//}
// 	//}

// 	if total < amount+gas {
// 		//资金不够
// 		return nil, errors.New("余额不足")
// 	}
// 	//构建交易输出
// 	vouts := make([]mining.Vout, 0)
// 	for _, one := range address {
// 		vout := mining.Vout{
// 			Value:   one.Amount,  //输出金额 = 实际金额 * 100000000
// 			Address: one.Address, //钱包地址
// 		}
// 		vouts = append(vouts, vout)
// 	}
// 	//检查押金是否刚刚好，多了的转账给自己
// 	//TODO 将剩余款项转入新的地址，保证资金安全
// 	if total > amount+gas {
// 		vout := mining.Vout{
// 			Value:   total - amount - gas,  //输出金额 = 实际金额 * 100000000
// 			Address: keystore.GetAddr()[0], //钱包地址
// 		}
// 		vouts = append(vouts, vout)
// 	}
// 	var pay *mining.Tx_Pay
// 	//for i := uint64(0); i < 10000; i++ {
// 	//没有输出
// 	base := mining.TxBase{
// 		Type:       config.Wallet_tx_type_pay, //交易类型
// 		Vin_total:  uint64(len(vins)),         //输入交易数量
// 		Vin:        vins,                      //交易输入
// 		Vout_total: uint64(len(vouts)),        //输出交易数量
// 		Vout:       vouts,                     //交易输出
// 		Gas:        gas,                       //交易手续费
// 		//LockHeight: block.Height + 100 + i,    //锁定高度
// 		LockHeight: height + 100, //锁定高度
// 		//		CreateTime: time.Now().Unix(),         //创建时间
// 	}
// 	pay = &mining.Tx_Pay{
// 		TxBase: base,
// 	}

// 	//给输出签名，防篡改
// 	for i, one := range pay.Vin {
// 		for _, key := range keystore.GetAddrAll() {

// 			puk, ok := keystore.GetPukByAddr(key)
// 			if !ok {
// 				return nil, errors.New("未找到公钥")
// 			}

// 			if bytes.Equal(puk, one.Puk) {
// 				_, prk, _, err := keystore.GetKeyByAddr(key, pwd)
// 				// prk, err := key.GetPriKey(pwd)
// 				if err != nil {
// 					return nil, err
// 				}
// 				sign := pay.GetSign(&prk, one.Txid, one.Vout, uint64(i))
// 				//				sign := pay.GetVoutsSign(prk, uint64(i))
// 				pay.Vin[i].Sign = *sign
// 			}
// 		}
// 	}
// 	pay.BuildHash()
// 	// if pay.CheckHashExist() {
// 	// 	pay = nil
// 	// 	continue
// 	// } else {
// 	// 	break
// 	// }
// 	//}
// 	return pay, nil
// }

// /*
// 	创建一个投票交易
// 	@params height 当前区块高度 item 交易item voteType 投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；  witnessAddr 接受者地址 addr 投票者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
// */
// func CreateTxVoteIn(height uint64, items []*mining.TxItem, voteType uint16, witnessAddr crypto.AddressCoin, addr string, amount, gas uint64, pwd, comment string) (*mining.Tx_vote_in, error) {
// 	if len(items) == 0 {
// 		return nil, errors.New("余额不足")
// 	}
// 	if amount < config.Mining_vote {
// 		// fmt.Println("投票交押金数量最少", config.Mining_vote)
// 		//return nil, errors.New("投票交押金数量最少" + strconv.Itoa(config.Mining_vote))
// 		return nil, errors.New("投票交押金数量最少" + strconv.FormatUint(config.Mining_vote, 10))
// 	}
// 	// chain := forks.GetLongChain()
// 	// _, block := chain.GetLastBlock()
// 	// //查找余额
// 	vins := make([]mining.Vin, 0)
// 	total := uint64(0)
// 	// keys := keystore.GetAddrAll()
// 	// for _, one := range keys {

// 	// 	bas, err := chain.balance.FindBalance(&one)
// 	// 	if err != nil {
// 	// 		return nil, err
// 	// 	}

// 	// 	for _, two := range bas {
// 	// 		two.Txs.Range(func(k, v interface{}) bool {
// 	// 			item := v.(*TxItem)
// 	for _, item := range items {
// 		puk, ok := keystore.GetPukByAddr(*item.Addr)
// 		if !ok {
// 			continue
// 		}
// 		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

// 		vin := mining.Vin{
// 			Txid: item.Txid,     //UTXO 前一个交易的id
// 			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 			Puk:  puk,           //公钥
// 			//					Sign: *sign,           //签名
// 		}
// 		vins = append(vins, vin)

// 		total = total + item.Value
// 		if total >= amount+gas {
// 			//return false
// 			break
// 		}

// 	}
// 	// if total >= amount+gas {
// 	// 	break
// 	// }
// 	//}
// 	//}

// 	if total < amount+gas {
// 		//资金不够
// 		return nil, errors.New("余额不足")
// 	}
// 	//解析转账目标账户地址
// 	var dstAddr crypto.AddressCoin
// 	if addr == "" {
// 		// fmt.Println("自己地址数量", len(keystore.GetAddr()))
// 		//为空则转给自己
// 		dstAddr = keystore.GetAddr()[0]
// 	} else {
// 		// var err error
// 		// *dstAddr, err = utils.FromB58String(addr)
// 		// if err != nil {
// 		// 	// fmt.Println("解析地址失败")
// 		// 	return nil
// 		// }
// 		dstAddr = crypto.AddressFromB58String(addr)
// 	}
// 	//构建交易输出
// 	vouts := make([]mining.Vout, 0)
// 	vout := mining.Vout{
// 		Value:   amount,  //输出金额 = 实际金额 * 100000000
// 		Address: dstAddr, //钱包地址
// 	}
// 	vouts = append(vouts, vout)
// 	//检查押金是否刚刚好，多了的转账给自己
// 	//TODO 将剩余款项转入新的地址，保证资金安全
// 	if total > amount+gas {
// 		vout := mining.Vout{
// 			Value:   total - amount - gas,  //输出金额 = 实际金额 * 100000000
// 			Address: keystore.GetAddr()[0], //钱包地址
// 		}
// 		vouts = append(vouts, vout)
// 	}
// 	var txin *mining.Tx_vote_in
// 	//for i := uint64(0); i < 10000; i++ {
// 	//没有输出
// 	base := mining.TxBase{
// 		Type:       config.Wallet_tx_type_vote_in, //交易类型
// 		Vin_total:  uint64(len(vins)),             //输入交易数量
// 		Vin:        vins,                          //交易输入
// 		Vout_total: uint64(len(vouts)),            //输出交易数量
// 		Vout:       vouts,                         //交易输出
// 		Gas:        gas,                           //交易手续费
// 		//LockHeight: block.Height + 100 + i,    //锁定高度
// 		LockHeight: height + 100, //锁定高度
// 		//		CreateTime: time.Now().Unix(),         //创建时间
// 	}
// 	txin = &mining.Tx_vote_in{
// 		TxBase:   base,
// 		Vote:     witnessAddr,
// 		VoteType: voteType,
// 	}

// 	//给输出签名，防篡改
// 	for i, one := range txin.Vin {
// 		for _, key := range keystore.GetAddrAll() {

// 			puk, ok := keystore.GetPukByAddr(key)
// 			if !ok {
// 				return nil, errors.New("未找到公钥")
// 			}

// 			if bytes.Equal(puk, one.Puk) {
// 				_, prk, _, err := keystore.GetKeyByAddr(key, pwd)
// 				// prk, err := key.GetPriKey(pwd)
// 				if err != nil {
// 					return nil, err
// 				}
// 				sign := txin.GetSign(&prk, one.Txid, one.Vout, uint64(i))
// 				//				sign := pay.GetVoutsSign(prk, uint64(i))
// 				txin.Vin[i].Sign = *sign
// 			}
// 		}
// 	}
// 	txin.BuildHash()
// 	// if pay.CheckHashExist() {
// 	// 	pay = nil
// 	// 	continue
// 	// } else {
// 	// 	break
// 	// }
// 	//}
// 	return txin, nil
// }

// /*
// 	创建一个投票押金退还交易
// 	退还按交易为单位，交易的押金全退
// 	@param height 区块高度 voteitems 投票item items 余额 witness 见证人 addr 投票地址
// */
// func CreateTxVoteOut(height uint64, voteitems, items []*mining.TxItem, witness *crypto.AddressCoin, addr string, amount, gas uint64, pwd string) *mining.Tx_vote_out {
// 	//查找余额
// 	vins := make([]mining.Vin, 0)
// 	total := uint64(0)
// 	//TODO 此处item为投票
// 	for _, item := range voteitems {
// 		//TODO txid对应的vout addr. 即上一个输出的out addr
// 		voutaddr := crypto.AddressCoin{}
// 		puk, ok := keystore.GetPukByAddr(voutaddr)
// 		if !ok {
// 			continue
// 		}

// 		vin := mining.Vin{
// 			Txid: item.Txid,     //UTXO 前一个交易的id
// 			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 			Puk:  puk,           //公钥
// 			//			Sign: *sign,         //签名
// 		}
// 		vins = append(vins, vin)

// 		total = total + item.Value
// 		if total >= amount+gas {
// 			break
// 		}
// 	}

// 	// fmt.Println("==============3")
// 	//资金不够
// 	//TODO 此处items为余额
// 	for _, item := range items {
// 		puk, ok := keystore.GetPukByAddr(*item.Addr)
// 		if !ok {
// 			continue
// 		}

// 		vin := mining.Vin{
// 			Txid: item.Txid,     //UTXO 前一个交易的id
// 			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
// 			Puk:  puk,           //公钥
// 			//						Sign: *sign,           //签名
// 		}
// 		vins = append(vins, vin)

// 		total = total + item.Value
// 		if total >= amount+gas {
// 			break
// 		}
// 	}
// 	// fmt.Println("==============4")
// 	//余额不够给手续费
// 	if total < (amount + gas) {
// 		// fmt.Println("押金不够")
// 		//押金不够
// 		return nil
// 	}
// 	// fmt.Println("==============5")

// 	//解析转账目标账户地址
// 	var dstAddr crypto.AddressCoin
// 	if addr == "" {
// 		//为空则转给自己
// 		dstAddr = keystore.GetAddr()[0]
// 	} else {
// 		// var err error
// 		// *dstAddr, err = utils.FromB58String(addr)
// 		// if err != nil {
// 		// 	// fmt.Println("解析地址失败")
// 		// 	return nil
// 		// }
// 		dstAddr = crypto.AddressFromB58String(addr)
// 	}
// 	// fmt.Println("==============6")

// 	//构建交易输出
// 	vouts := make([]mining.Vout, 0)
// 	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
// 	vout := mining.Vout{
// 		Value:   total - gas, //输出金额 = 实际金额 * 100000000
// 		Address: dstAddr,     //钱包地址
// 	}
// 	vouts = append(vouts, vout)

// 	//	crateTime := time.Now().Unix()

// 	var txout *mining.Tx_vote_out
// 	//
// 	base := mining.TxBase{
// 		Type:       config.Wallet_tx_type_vote_out, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
// 		Vin_total:  uint64(len(vins)),              //输入交易数量
// 		Vin:        vins,                           //交易输入
// 		Vout_total: uint64(len(vouts)),             //输出交易数量
// 		Vout:       vouts,                          //
// 		Gas:        gas,                            //交易手续费
// 		LockHeight: height + 100,                   //锁定高度
// 		//		CreateTime: crateTime,                      //创建时间
// 	}
// 	txout = &mining.Tx_vote_out{
// 		TxBase: base,
// 	}
// 	// fmt.Println("==============7")

// 	//给输出签名，防篡改
// 	for i, one := range txout.Vin {
// 		for _, key := range keystore.GetAddr() {
// 			puk, ok := keystore.GetPukByAddr(key)
// 			if !ok {
// 				return nil
// 			}

// 			if bytes.Equal(puk, one.Puk) {
// 				_, prk, _, err := keystore.GetKeyByAddr(key, pwd)

// 				// prk, err := key.GetPriKey(pwd)
// 				if err != nil {
// 					// fmt.Println("获取key错误")
// 					return nil
// 				}
// 				sign := txout.GetSign(&prk, one.Txid, one.Vout, uint64(i))
// 				//				sign := txout.GetVoutsSign(prk, uint64(i))
// 				txout.Vin[i].Sign = *sign
// 			}
// 		}
// 	}
// 	txout.BuildHash()
// 	return txout
// }
