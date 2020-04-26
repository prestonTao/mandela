package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/sqlite3_db"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"golang.org/x/crypto/ed25519"
)

/*
	获取账户所有地址的余额
*/
func GetBalance() uint64 {
	count := uint64(0)

	keys := keystore.GetAddr()
	addrs := make([]*crypto.AddressCoin, 0)
	for i, _ := range keys {
		addrs = append(addrs, &keys[i])
	}
	chain := forks.GetLongChain()
	if chain == nil {
		return 0
	}
	bs := chain.balance.FindBalance(addrs...)
	for _, one := range bs {
		one.Txs.Range(func(k, v interface{}) bool {
			item := v.(*TxItem)
			count = count + item.Value
			return true
		})
	}
	return count
}

/*
	通过地址获取余额
*/
func GetBalanceForAddr(addr *crypto.AddressCoin) uint64 {
	count := uint64(0)
	chain := forks.GetLongChain()
	// fmt.Println(chain)
	// fmt.Println(chain.balance)
	bs := chain.balance.FindBalance(addr)
	for _, one := range bs {
		one.Txs.Range(func(k, v interface{}) bool {
			item := v.(*TxItem)
			// txidStr := k.(string)
			// fmt.Println("统计余额", txidStr, item.Value)
			count = count + item.Value
			return true
		})
	}
	return count
}

/*
	获取区块是否同步完成
*/
func GetSyncFinish() bool {
	chain := forks.GetLongChain()
	//判断是否同步完成
	if forks.GetHighestBlock() <= 0 {
		//区块未同步完成，不能挖矿
		return false
	}
	if forks.GetHighestBlock() > chain.GetCurrentBlock() {
		//区块未同步完成，不能挖矿
		return false
	}
	return true
}

// var oldCountTotal = uint64(0)
// var countTotal = uint64(0)
// var onece sync.Once

// func Onece() {
// 	onece.Do(func() {
// 		for {
// 			time.Sleep(time.Second)
// 			newTotal := atomic.LoadUint64(&countTotal)
// 			fmt.Println("====================\n每秒钟处理交易笔数", newTotal-oldCountTotal)
// 			atomic.StoreUint64(&oldCountTotal, countTotal)
// 		}
// 	})
// }

/*
	给一个地址转账
*/
func SendToAddress(address *crypto.AddressCoin, amount, gas uint64, pwd, comment string) (*Tx_Pay, error) {
	txpay, err := CreateTxPay(address, amount, gas, pwd, comment)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()

	// fmt.Println("---------------------查看一个交易的大小", len(*txpay.Serialize()))

	// fmt.Println("txid 11111111", hex.EncodeToString(*txpay.GetHash()))
	bs, err := txpay.Json()
	if err != nil {
		// fmt.Println("33333333333333 33333")
		return nil, err
	}
	// fmt.Println("4444444444444444")

	txbase, err := ParseTxBase(bs)
	if err != nil {
		return nil, err
	}
	txbase.BuildHash()
	// fmt.Println("txid 22222222", hex.EncodeToString(*txbase.GetHash()))
	// fmt.Println("66666666666666")
	//验证交易
	if err := txbase.Check(); err != nil {
		//交易不合法，则不发送出去
		engine.Log.Info("If the transaction is illegal, it will not be sent out. error:%s", err.Error())
		return nil, err
	}
	forks.GetLongChain().transactionManager.AddTx(txbase)

	MulticastTx(bs)
	// atomic.AddUint64(&countTotal, 1)
	// fmt.Println("增加一笔交易")
	// go Onece()
	//	unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	return txpay, nil
}

/*
	给多个收款地址转账
*/
func SendToMoreAddress(address []PayNumber, gas uint64, pwd, comment string) (*Tx_Pay, error) {
	txpay, err := CreateTxsPay(address, gas, pwd, comment)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()
	bs, err := txpay.Json()
	if err != nil {
		// fmt.Println("33333333333333 33333")
		return nil, err
	}
	// fmt.Println("4444444444444444")

	txbase, err := ParseTxBase(bs)
	if err != nil {
		return nil, err
	}
	txbase.BuildHash()
	// fmt.Println("66666666666666")
	//验证交易
	if err := txbase.Check(); err != nil {
		//交易不合法，则不发送出去
		// fmt.Println("交易不合法，则不发送出去")
		return nil, err
	}
	forks.GetLongChain().transactionManager.AddTx(txbase)
	MulticastTx(bs)

	//	unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	return txpay, nil
}

/*
	给多个地址转账,带Payload签名
*/
func SendToMoreAddressByPayload(address []PayNumber, gas uint64, pwd string, cs CommunitySign) (*Tx_Pay, error) {
	txpay, err := CreateTxsPayByPayload(address, gas, pwd, cs)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()
	bs, err := txpay.Json()
	if err != nil {
		// fmt.Println("33333333333333 33333")
		return nil, err
	}
	// fmt.Println("4444444444444444")

	txbase, err := ParseTxBase(bs)
	if err != nil {
		return nil, err
	}
	txbase.BuildHash()
	// fmt.Println("66666666666666")
	//验证交易
	if err := txbase.Check(); err != nil {
		//交易不合法，则不发送出去
		// fmt.Println("交易不合法，则不发送出去")
		return nil, err
	}
	forks.GetLongChain().transactionManager.AddTx(txbase)
	MulticastTx(bs)

	//	unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	return txpay, nil
}

/*
	从邻居节点查询起始区块hash
*/
func FindStartBlockForNeighbor() *ChainInfo {
	for _, key := range nodeStore.GetLogicNodes() {

		message, _ := message_center.SendNeighborMsg(config.MSGID_getBlockHead, key, nil)

		bs := flood.WaitRequest(mc.CLASS_getBlockHead, hex.EncodeToString(message.Body.Hash), 0)
		// fmt.Println("有消息返回了啊")
		if bs == nil {
			// fmt.Println("从邻居节点查询起始区块hash 发送共享文件消息失败，可能超时")
			continue
		}

		chainInfo := new(ChainInfo)
		decoder := json.NewDecoder(bytes.NewBuffer(*bs))
		decoder.UseNumber()
		err := decoder.Decode(chainInfo)
		// err := json.Unmarshal(*bs, chainInfo)
		if err != nil {
			return nil
		}

		return chainInfo
		// }
	}
	return nil
}

/*
	从邻居节点查询区块头和区块中的交易
*/
func FindBlockForNeighbor(bhash *[]byte) *BlockHeadVO {
	bhvo := new(BlockHeadVO)
	bs := getValueForNeighbor(bhash)
	if bs == nil {
		engine.Log.Info("Error synchronizing chunk from neighbor node")
		return nil
	}
	bhvo, err := ParseBlockHeadVO(bs)
	if err != nil {
		return nil
	}
	return bhvo

	//	bh, err := ParseBlockHead(bs)
	//	if err != nil {
	//		return nil
	//	}
	//	// if bh.Nextblockhash == nil || len(bh.Nextblockhash) <= 0 {
	//	// 	engine.Log.Warn("收到的区块头，没有 Nextblockhash")
	//	// }
	//	bhvo.BH = bh
	//	bhvo.Txs = make([]TxItr, 0)
	//	for i, _ := range bh.Tx {
	//		txbs := getValueForNeighbor(&bh.Tx[i])
	//		if txbs == nil {
	//			engine.Log.Info("Error synchronizing transactions in block from neighbor node")
	//			return nil
	//		}
	//		//TODO 验证交易是否合法
	//		txItr, err := ParseTxBase(txbs)
	//		if err != nil {
	//			//TODO 这里一个节点错误，应该从另一个邻居节点拉取交易
	//			return nil
	//		}
	//		bhvo.Txs = append(bhvo.Txs, txItr)
	//	}
	//	return bhvo
}

/*
	查询邻居节点数据库，key：value查询
*/
func getValueForNeighbor(bhash *[]byte) *[]byte {
	// fmt.Println("1查询区块或交易", hex.EncodeToString(*bhash))
	var bs *[]byte
	var err error
	for {
		logicNodes := nodeStore.GetLogicNodes()
		logicNodes = OrderNodeAddr(logicNodes)
		for _, key := range logicNodes {
			engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
			engine.Log.Info("Send query message to node %s", key.B58String())

			message, _ := message_center.SendNeighborMsg(config.MSGID_getTransaction, key, bhash)
			// engine.Log.Info("44444444444 %s", key.B58String())
			bs = flood.WaitRequest(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), config.Mining_block_time)
			if bs == nil {
				// engine.Log.Info("5555555555555555 %s", key.B58String())
				//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
				engine.Log.Info("Receive message timeout %s", key.B58String())
				err = errors.New("Failed to send shared file message, may timeout")
				continue
			}
			engine.Log.Info("Receive message %s", key.B58String())
			// engine.Log.Info("66666666666666 %s", key.B58String())
			err = nil
			break
		}
		// engine.Log.Info("7777777777777777777")
		if err == nil {
			// engine.Log.Info("888888888888888")
			break
		}
		// engine.Log.Info("99999999999999999999")
	}
	if err != nil {
		// engine.Log.Info("10101010101001010101")
		engine.Log.Warn("Failed to query block transaction", hex.EncodeToString(*bhash))
	}
	// engine.Log.Info("11 11 11 11 11")
	// if bs == nil {
	// 	fmt.Println("查询区块或交易结果", bs)
	// } else {
	// 	fmt.Println("查询区块或交易结果", len(*bs))
	// }
	return bs
}

/*
	从邻居节点获取未确认的区块
*/
func GetUnconfirmedBlockForNeighbor(height uint64) []*BlockHeadVO {
	engine.Log.Info("Synchronize unacknowledged chunks from neighbor nodes")

	heightBs := utils.Uint64ToBytes(height)

	var bs *[]byte
	var err error
	for i := 0; i < 10; i++ {

		for _, key := range nodeStore.GetLogicNodes() {
			// for _, key := range nodeStore.GetAllNodes() {
			// fmt.Println("找到一个邻居节点，开始同步区块数据\n", hex.EncodeToString(*bhash))

			message, _ := message_center.SendNeighborMsg(config.MSGID_getUnconfirmedBlock, key, &heightBs)

			bs = flood.WaitRequest(mc.CLASS_getUnconfirmedBlock, hex.EncodeToString(message.Body.Hash), 0)
			if bs == nil {
				engine.Log.Info("Failed to get unconfirmed block from neighbor node, sending shared file message, maybe timeout")
				err = errors.New("Failed to get unconfirmed block from neighbor node, sending shared file message, maybe timeout")
				continue
			}
			break
		}
		if err == nil {
			break
		}
	}
	// engine.Log.Info("获取的未确认区块 bs", string(*bs))
	if bs == nil {
		engine.Log.Warn("Get unacknowledged block BS error")
		return nil
	}

	temp := make([]interface{}, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err = decoder.Decode(&temp)

	blockHeadVOs := make([]*BlockHeadVO, 0)
	for _, one := range temp {
		bs, err := json.Marshal(one)
		blockVOone, err := ParseBlockHeadVO(&bs)
		if err != nil {
			engine.Log.Warn("Get unacknowledged block BS error:%s", err.Error())
			continue
		}
		blockHeadVOs = append(blockHeadVOs, blockVOone)
	}

	engine.Log.Info("Get unacknowledged block Success")
	return blockHeadVOs

}

/*
	缴纳押金，成为备用见证人
*/
func DepositIn(amount, gas uint64, pwd, payload string) error {
	//缴纳备用见证人押金交易
	err := forks.GetLongChain().balance.DepositIn(amount, gas, pwd, payload)
	if err != nil {
		// fmt.Println("缴纳押金失败", err)
	}
	// fmt.Println("缴纳押金完成")
	return err
}

/*
	退还押金
	@addr    string    可选（默认退回到原地址）。押金赎回到的账户地址
	@amount  uint64    可选（默认退还全部押金）。押金金额
*/
func DepositOut(addr string, amount, gas uint64, pwd string) error {
	//缴纳备用见证人押金交易
	err := forks.GetLongChain().balance.DepositOut(addr, amount, gas, pwd)

	return err
}

/*
	给见证人投票
*/
func VoteIn(t uint16, witnessAddr crypto.AddressCoin, addr string, amount, gas uint64, pwd, payload string) error {
	//缴纳备用见证人押金交易
	err := forks.GetLongChain().balance.VoteIn(t, witnessAddr, addr, amount, gas, pwd, payload)
	if err != nil {
		// fmt.Println("缴纳押金失败", err)
	}
	// fmt.Println("缴纳押金完成")
	return err
}

/*
	退还见证人投票押金
*/
func VoteOut(witnessAddr *crypto.AddressCoin, txid, addr string, amount, gas uint64, pwd string) error {
	//缴纳备用见证人押金交易
	return forks.GetLongChain().balance.VoteOut(witnessAddr, txid, addr, amount, gas, pwd)

}

// /*
// 	给社区节点投票
// */
// func VoteInLight(witnessAddr crypto.AddressCoin, addr string, amount, gas uint64, pwd string) error {
// 	//缴纳备用见证人押金交易
// 	err := forks.GetLongChain().balance.VoteIn(witnessAddr, addr, amount, gas, pwd)
// 	if err != nil {
// 		// fmt.Println("缴纳押金失败", err)
// 	}
// 	// fmt.Println("缴纳押金完成")
// 	return err
// }

// /*
// 	退还给社区节点投票
// */
// func VoteOutLight(witnessAddr *crypto.AddressCoin, txid, addr string, amount, gas uint64, pwd string) error {
// 	//缴纳备用见证人押金交易
// 	return forks.GetLongChain().balance.VoteOut(witnessAddr, txid, addr, amount, gas, pwd)

// }

/*
	获取见证人状态
	@return    bool    是否是候选见证人
	@return    bool    是否是备用见证人
	@return    bool    是否是没有按时出块，已经被踢出局，只有退还押金，重新缴纳押金成为候选见证人
	@return    crypto.AddressCoin    见证人地址
*/
func GetWitnessStatus() (IsCandidate bool, IsBackup bool, IsKickOut bool, Addr string, value uint64) {
	addr := keystore.GetCoinbase()
	Addr = addr.B58String()
	IsCandidate = forks.GetLongChain().witnessBackup.FindWitness(&addr)
	IsBackup = forks.GetLongChain().witnessChain.FindWitness(addr)
	IsKickOut = forks.GetLongChain().witnessBackup.FindWitnessInBlackList(addr)
	txItem := forks.GetLongChain().balance.GetDepositIn()
	if txItem == nil {
		value = 0
	} else {
		value = txItem.Value
	}
	return
}

/*
	获取候选见证人列表
*/
func GetWitnessListSort() *WitnessBackupGroup {
	return forks.GetLongChain().witnessBackup.GetWitnessListSort()
}

/*
	获取社区节点列表
*/
func GetCommunityListSort() []*VoteScoreVO {
	return forks.GetLongChain().witnessBackup.GetCommunityListSort()
}

/*
	获得自己给哪些见证人投过票的列表
*/
func GetVoteList() []*Balance {
	return forks.GetLongChain().balance.GetVoteList()
}

/*
	查询一个交易是否上链成功
	@return    uint64    1=未确认;2=成功;3=失败;
*/
func FindTx(txid []byte, lockheight uint64) uint64 {
	//查询当前确认的区块高度
	height := GetLongChain().GetCurrentBlock()
	//查询交易上链的区块高度
	bs, err := db.Find(txid)
	if err != nil {
		if height > lockheight {
			//超过了锁定高度还没有上链，则失败了
			return 3
		}
		//没有上链，未确认
		return 1
	}
	txItr, err := ParseTxBase(bs)
	if err != nil {
		return 1
	}
	// bhash := txItr.GetBlockHash()

	blockHeadBs, err := db.Find(*txItr.GetBlockHash())
	if err != nil {
		return 1
	}

	bh, err := ParseBlockHead(blockHeadBs)
	if err != nil {
		return 1
	}
	if bh.Height < height {
		return 2
	}
	return 1
}

/*
	查询地址角色状态
	@return    int    1=见证人;2=社区节点;3=轻节点;4=什么也不是;
*/
func GetAddrState(addr crypto.AddressCoin) int {
	witnessBackup := forks.GetLongChain().witnessBackup
	//是否是轻节点
	_, isLight := witnessBackup.haveLight(&addr)
	if isLight {
		return 3
	}
	//是否是社区节点
	_, isCommunity := witnessBackup.haveCommunityList(&addr)
	if isCommunity {
		return 2
	}
	//是否是见证人
	isWitness := witnessBackup.haveWitness(&addr)
	if isWitness {
		return 1
	}
	return 4
}

/*
	添加一个自定义交易
	验证交易并广播
*/
func AddTx(txItr TxItr) error {
	if txItr == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	txItr.BuildHash()
	bs, err := txItr.Json()
	if err != nil {
		//		fmt.Println("33333333333333 33333")
		return err
	}
	//	fmt.Println("4444444444444444")
	//	fmt.Println("5555555555555555")
	txbase, err := ParseTxBase(bs)
	if err != nil {
		return err
	}
	txbase.BuildHash()
	//	fmt.Println("66666666666666")
	//验证交易
	if err := txbase.Check(); err != nil {
		//交易不合法，则不发送出去
		// fmt.Println("交易不合法，则不发送出去")
		return err
	}

	ok := forks.GetLongChain().transactionManager.AddTx(txbase)
	if !ok {
		//等待上链,请稍后重试.
		return errors.New("Waiting for the chain, please try again later")
	}
	MulticastTx(bs)
	return nil
}

/*
	创建一个转款交易
	@params height 当前区块高度 item 交易item  pubs 地址公钥对 address 接受者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
*/
func CreateTxPayM(height uint64, items []*TxItem, pubs map[string]ed25519.PublicKey, address *crypto.AddressCoin, amount, gas uint64, comment string, returnaddr crypto.AddressCoin) (*Tx_Pay, error) {
	if len(items) == 0 {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	// chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	// //查找余额
	vins := make([]Vin, 0)
	total := uint64(0)
	// keys := keystore.GetAddrAll()
	// for _, one := range keys {

	// 	bas, err := chain.balance.FindBalance(&one)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			item := v.(*TxItem)
	//var returnaddr crypto.AddressCoin //找零退回地址
	for _, item := range items {
		// if k == 0 {
		// 	returnaddr = *item.Addr
		// }
		addrstr := *item.Addr
		puk, ok := pubs[addrstr.B58String()]
		if !ok {
			continue
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

		vin := Vin{
			Txid: item.Txid,     //UTXO 前一个交易的id
			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,           //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, vin)

		total = total + item.Value
		if total >= amount+gas {
			//return false
			break
		}

	}
	// if total >= amount+gas {
	// 	break
	// }
	//}
	//}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough
	}

	//构建交易输出
	vouts := make([]Vout, 0)
	vout := Vout{
		Value:   amount,   //输出金额 = 实际金额 * 100000000
		Address: *address, //钱包地址
	}
	vouts = append(vouts, vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr,
		}
		vouts = append(vouts, vout)
	}
	var pay *Tx_Pay
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_pay, //交易类型
		Vin_total:  uint64(len(vins)),         //输入交易数量
		Vin:        vins,                      //交易输入
		Vout_total: uint64(len(vouts)),        //输出交易数量
		Vout:       vouts,                     //交易输出
		Gas:        gas,                       //交易手续费
		//LockHeight: block.Height + 100 + i,    //锁定高度
		LockHeight: height + 100, //锁定高度
		//		CreateTime: time.Now().Unix(),         //创建时间
	}
	pay = &Tx_Pay{
		TxBase: base,
	}
	//给输出签名，防篡改
	for i, one := range pay.Vin {
		sign := pay.GetWaitSign(one.Txid, one.Vout, uint64(i))
		if sign == nil {
			//预签名时数据错误
			return nil, errors.New("Data error while pre signing")
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}
	//pay.BuildHash()
	// if pay.CheckHashExist() {
	// 	pay = nil
	// 	continue
	// } else {
	// 	break
	// }
	//}
	return pay, nil
}

/*
	创建多个转款交易
	@params height 当前区块高度 item 交易item  pubs 地址公钥对 address 接受者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
*/
func CreateTxsPayM(height uint64, items []*TxItem, pubs map[string]ed25519.PublicKey, address []PayNumber, gas uint64, comment string, returnaddr crypto.AddressCoin) (*Tx_Pay, error) {
	if len(items) == 0 {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	// chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	// //查找余额
	vins := make([]Vin, 0)
	total := uint64(0)
	amount := uint64(0)
	for _, one := range address {
		amount += one.Amount
	}
	// keys := keystore.GetAddrAll()
	// for _, one := range keys {

	// 	bas, err := chain.balance.FindBalance(&one)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			item := v.(*TxItem)
	//var returnaddr crypto.AddressCoin //找零退回地址
	for _, item := range items {
		// if k == 0 {
		// 	returnaddr = *item.Addr
		// }
		addrstr := *item.Addr
		puk, ok := pubs[addrstr.B58String()]
		if !ok {
			continue
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

		vin := Vin{
			Txid: item.Txid,     //UTXO 前一个交易的id
			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,           //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, vin)

		total = total + item.Value
		if total >= amount+gas {
			//return false
			break
		}

	}
	// if total >= amount+gas {
	// 	break
	// }
	//}
	//}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	//构建交易输出
	vouts := make([]Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:   one.Amount,  //输出金额 = 实际金额 * 100000000
			Address: one.Address, //钱包地址
		}
		vouts = append(vouts, vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr, //钱包地址
		}
		vouts = append(vouts, vout)
	}
	var pay *Tx_Pay
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_pay, //交易类型
		Vin_total:  uint64(len(vins)),         //输入交易数量
		Vin:        vins,                      //交易输入
		Vout_total: uint64(len(vouts)),        //输出交易数量
		Vout:       vouts,                     //交易输出
		Gas:        gas,                       //交易手续费
		//LockHeight: block.Height + 100 + i,    //锁定高度
		LockHeight: height + 100, //锁定高度
		//		CreateTime: time.Now().Unix(),         //创建时间
	}
	pay = &Tx_Pay{
		TxBase: base,
	}

	//给输出签名，防篡改
	for i, one := range pay.Vin {
		sign := pay.GetWaitSign(one.Txid, one.Vout, uint64(i))
		if sign == nil {
			//预签名时数据错误
			return nil, errors.New("Data error while pre signing")
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		pay.Vin[i].Sign = *sign
	}
	//pay.BuildHash()
	// if pay.CheckHashExist() {
	// 	pay = nil
	// 	continue
	// } else {
	// 	break
	// }
	//}
	return pay, nil
}

/*
	创建一个投票交易
	@params height 当前区块高度 item 交易item  pubs 地址公钥对 voteType 投票类型 1=给见证人投票；2=给社区节点投票；3=轻节点押金；  witnessAddr 接受者地址 addr 投票者地址 amount 金额 gas 手续费 pwd 密码 commnet 说明
*/
func CreateTxVoteInM(height uint64, items []*TxItem, pubs map[string]ed25519.PublicKey, voteType uint16, witnessAddr crypto.AddressCoin, addr string, amount, gas uint64, comment string, returnaddr crypto.AddressCoin) (*Tx_vote_in, error) {
	if len(items) == 0 {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	if voteType == 1 && amount < config.Mining_vote {
		//交押金数量最少需要
		return nil, errors.New("Minimum deposit required" + strconv.FormatUint(config.Mining_vote, 10))
	}
	if voteType == 3 && amount < config.Mining_light_min {
		//交押金数量最少需要
		return nil, errors.New("Minimum deposit required" + strconv.FormatUint(config.Mining_light_min, 10))
	}
	// chain := forks.GetLongChain()
	// _, block := chain.GetLastBlock()
	// //查找余额
	vins := make([]Vin, 0)
	total := uint64(0)
	// keys := keystore.GetAddrAll()
	// for _, one := range keys {

	// 	bas, err := chain.balance.FindBalance(&one)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	for _, two := range bas {
	// 		two.Txs.Range(func(k, v interface{}) bool {
	// 			item := v.(*TxItem)
	//var returnaddr crypto.AddressCoin //找零退回地址
	for _, item := range items {
		// if k == 0 {
		// 	returnaddr = *item.Addr
		// }
		addrstr := *item.Addr
		puk, ok := pubs[addrstr.B58String()]
		if !ok {
			continue
		}
		// fmt.Println("创建交易时候公钥", hex.EncodeToString(puk))

		vin := Vin{
			Txid: item.Txid,     //UTXO 前一个交易的id
			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,           //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, vin)

		total = total + item.Value
		if total >= amount+gas {
			//return false
			break
		}

	}
	// if total >= amount+gas {
	// 	break
	// }
	//}
	//}

	if total < amount+gas {
		//余额不足
		return nil, config.ERROR_not_enough
	}
	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin
	if addr == "" {
		// fmt.Println("自己地址数量", len(keystore.GetAddr()))
		//为空则转给自己
		dstAddr = returnaddr //keystore.GetAddr()[0]
	} else {
		// var err error
		// *dstAddr, err = utils.FromB58String(addr)
		// if err != nil {
		// 	// fmt.Println("解析地址失败")
		// 	return nil
		// }
		dstAddr = crypto.AddressFromB58String(addr)
	}
	//构建交易输出
	vouts := make([]Vout, 0)
	vout := Vout{
		Value:   amount,  //输出金额 = 实际金额 * 100000000
		Address: dstAddr, //钱包地址
	}
	vouts = append(vouts, vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr,
		}
		vouts = append(vouts, vout)
	}
	var txin *Tx_vote_in
	//for i := uint64(0); i < 10000; i++ {
	//没有输出
	base := TxBase{
		Type:       config.Wallet_tx_type_vote_in, //交易类型
		Vin_total:  uint64(len(vins)),             //输入交易数量
		Vin:        vins,                          //交易输入
		Vout_total: uint64(len(vouts)),            //输出交易数量
		Vout:       vouts,                         //交易输出
		Gas:        gas,                           //交易手续费
		//LockHeight: block.Height + 100 + i,    //锁定高度
		LockHeight: height + 100,    //锁定高度
		Payload:    []byte(comment), //备注
		//		CreateTime: time.Now().Unix(),         //创建时间
	}

	voteAddr := NewVoteAddressByAddr(witnessAddr)

	txin = &Tx_vote_in{
		TxBase:   base,
		Vote:     voteAddr,
		VoteType: voteType,
	}

	//给输出签名，防篡改
	for i, one := range txin.Vin {
		sign := txin.GetWaitSign(one.Txid, one.Vout, uint64(i))
		if sign == nil {
			return nil, config.ERROR_get_sign_data_fail
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		// fmt.Printf("sign前:puk:%x signdst:%x", md5.Sum(one.Puk), md5.Sum(*sign))
		txin.Vin[i].Sign = *sign
	}
	//txin.BuildHash()
	// if pay.CheckHashExist() {
	// 	pay = nil
	// 	continue
	// } else {
	// 	break
	// }
	//}
	return txin, nil
}

/*
	创建一个投票押金退还交易
	退还按交易为单位，交易的押金全退
	@param height 区块高度 voteitems 投票item items 余额 pubs 地址公钥对 witness 见证人 addr 投票地址
*/
func CreateTxVoteOutM(height uint64, voteitems, items []*TxItem, pubs map[string]ed25519.PublicKey, witness *crypto.AddressCoin, addr string, amount, gas uint64, returnaddr crypto.AddressCoin) (*Tx_vote_out, error) {
	//查找余额
	vins := make([]Vin, 0)
	total := uint64(0)
	//TODO 此处item为投票
	for _, item := range voteitems {
		//TODO txid对应的vout addr. 即上一个输出的out addr
		voutaddr := *item.Addr
		puk, ok := pubs[voutaddr.B58String()]
		if !ok {
			continue
		}

		vin := Vin{
			Txid: item.Txid,     //UTXO 前一个交易的id
			Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,           //公钥
			//			Sign: *sign,         //签名
		}
		vins = append(vins, vin)

		total = total + item.Value
		if total >= amount+gas {
			break
		}
	}

	// fmt.Println("==============3")
	//资金不够
	//TODO 此处items为余额
	//var returnaddr crypto.AddressCoin //找零退回地址
	if total < amount+gas {
		for _, item := range items {
			// if k == 0 {
			// 	returnaddr = *item.Addr
			// }
			addrstr := *item.Addr
			puk, ok := pubs[addrstr.B58String()]
			if !ok {
				continue
			}

			vin := Vin{
				Txid: item.Txid,     //UTXO 前一个交易的id
				Vout: item.OutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
				Puk:  puk,           //公钥
				//						Sign: *sign,           //签名
			}
			vins = append(vins, vin)

			total = total + item.Value
			if total >= amount+gas {
				break
			}
		}
	}
	// fmt.Println("==============4")
	//余额不够给手续费
	if total < (amount + gas) {
		// fmt.Println("押金不够")
		//押金不够
		return nil, config.ERROR_not_enough
	}
	// fmt.Println("==============5")

	//解析转账目标账户地址
	var dstAddr crypto.AddressCoin
	if addr == "" {
		//为空则转给自己
		dstAddr = returnaddr
	} else {
		// var err error
		// *dstAddr, err = utils.FromB58String(addr)
		// if err != nil {
		// 	// fmt.Println("解析地址失败")
		// 	return nil
		// }
		dstAddr = crypto.AddressFromB58String(addr)
	}
	// fmt.Println("==============6")

	//构建交易输出
	vouts := make([]Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   total - gas, //输出金额 = 实际金额 * 100000000
		Address: dstAddr,     //钱包地址
	}
	vouts = append(vouts, vout)

	//	crateTime := time.Now().Unix()

	var txout *Tx_vote_out
	//
	base := TxBase{
		Type:       config.Wallet_tx_type_vote_out, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  uint64(len(vins)),              //输入交易数量
		Vin:        vins,                           //交易输入
		Vout_total: uint64(len(vouts)),             //输出交易数量
		Vout:       vouts,                          //
		Gas:        gas,                            //交易手续费
		LockHeight: height + 100,                   //锁定高度
		//		CreateTime: crateTime,                      //创建时间
	}
	txout = &Tx_vote_out{
		TxBase: base,
	}
	// fmt.Println("==============7")

	//给输出签名，防篡改
	for i, one := range txout.Vin {
		sign := txout.GetWaitSign(one.Txid, one.Vout, uint64(i))
		if sign == nil {
			return nil, config.ERROR_get_sign_data_fail
		}
		//				sign := pay.GetVoutsSign(prk, uint64(i))
		txout.Vin[i].Sign = *sign
	}
	//txout.BuildHash()
	return txout, nil
}

type BlockVotesVO struct {
	EndHeight uint64
	Group     []GroupVO
}

type GroupVO struct {
	StartHeight    uint64
	EndHeight      uint64
	CommunityVotes []VoteScoreRewadrVO
}

type VoteScoreRewadrVO struct {
	VoteScore              //
	LightVotes []VoteScore //轻节点投票列表
	Reward     uint64      //这个见证人获得的奖励
}

/*
	查询历史轻节点投票
*/
func FindLightVote(startHeight, endHeight uint64) (*BlockVotesVO, error) {
	if endHeight == 0 {
		endHeight = forks.LongChain.GetCurrentBlock()
	}
	bvVO := &BlockVotesVO{
		EndHeight: endHeight,
		Group:     make([]GroupVO, 0),
	}

	//查找上一个已经确认的组
	var preBlock *Block
	preGroup := forks.LongChain.witnessChain.witnessGroup
	for {
		preGroup = preGroup.PreGroup
		ok, preGroupBlock := preGroup.CheckBlockGroup()
		if ok {
			preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
			break
		}
	}

	//找到查询的起始区块
	for {
		if preBlock.Height > startHeight {
			if preBlock.PreBlock == nil {
				break
			}
			preBlock = preBlock.PreBlock
		}
		if preBlock.Height < startHeight {
			if preBlock.NextBlock == nil {
				break
			}
			preBlock = preBlock.NextBlock
		}
		if preBlock.Height == startHeight {
			break
		}
	}

	//找到这个见证人组的第一个见证人
	for {
		if preBlock.PreBlock == nil {
			//已经是创始区块了
			break
		}
		temp := preBlock.witness.WitnessBackupGroup
		if preBlock.PreBlock.witness.WitnessBackupGroup == temp {
			preBlock = preBlock.PreBlock
		} else {
			break
		}
	}

	//找到从start高度开始，到最新高度的所有见证人组的首块
	for {

		if preBlock.NextBlock == nil {
			break
		}
		if preBlock.NextBlock.witness.WitnessBackupGroup == preBlock.witness.WitnessBackupGroup {
			preBlock = preBlock.NextBlock
			continue
		}
		_, txs, err := preBlock.NextBlock.LoadTxs()
		if err != nil {
			return nil, err
		}
		reward := (*txs)[0].(*Tx_reward)

		groupVO := new(GroupVO)
		groupVO.CommunityVotes = make([]VoteScoreRewadrVO, 0)
		// groupVO.StartHeight = preBlock.Height
		groupVO.EndHeight = preBlock.Height
		//组装下一个见证人组的投票
		for _, one := range preBlock.witness.WitnessBackupGroup.Witnesses {

			m := make(map[string]*[]VoteScore)
			for i, two := range one.Votes {
				vo := VoteScore{
					Witness: one.Votes[i].Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
					Addr:    one.Votes[i].Addr,    //投票人地址
					Scores:  one.Votes[i].Scores,  //押金
					Vote:    one.Votes[i].Vote,    //获得票数
				}
				v, ok := m[two.Witness.B58String()]
				if ok {

				} else {
					temp := make([]VoteScore, 0)
					v = &temp
				}
				*v = append(*v, vo)
				m[two.Witness.B58String()] = v
			}

			for _, two := range one.CommunityVotes {
				// two.

				vo := VoteScore{
					Witness: two.Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
					Addr:    two.Addr,    //投票人地址
					Scores:  two.Scores,  //押金
					Vote:    two.Vote,    //获得票数
				}
				vsVOone := VoteScoreRewadrVO{
					VoteScore:  vo,
					LightVotes: make([]VoteScore, 0),
					Reward:     0,
				}
				//查找奖励
				for _, one := range *reward.GetVout() {
					if bytes.Equal(one.Address, *vsVOone.VoteScore.Addr) {
						vsVOone.Reward = one.Value
						break
					}
				}

				v, ok := m[two.Addr.B58String()]
				if ok {
					vsVOone.LightVotes = *v
				}

				// for i, one := range one.Votes {
				// 	vs := VoteScore{
				// 		Witness: one.Votes[i].Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
				// 		Addr:    one.Votes[i].Addr,    //投票人地址
				// 		Scores:  one.Votes[i].Scores,  //押金
				// 		Vote:    one.Votes[i].Vote,    //获得票数
				// 	}
				// 	vsVOone.LightVotes = append(vsVOone.LightVotes, vs)
				// }
				groupVO.CommunityVotes = append(groupVO.CommunityVotes, vsVOone)
			}
		}
		bvVO.Group = append(bvVO.Group, *groupVO)
		// blocks = append(blocks, preBlock)

		preBlock = preBlock.NextBlock
		if preBlock.Height > endHeight {
			break
		}
	}

	return bvVO, nil

}

/*
	通过区块高度，查询一个区块头信息
*/
func FindBlockHead(height uint64) *BlockHead {
	bhash, err := db.Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
	if err != nil {
		return nil
	}
	bs, err := db.Find(*bhash)
	if err != nil {
		return nil
	}
	bh, err := ParseBlockHead(bs)
	if err != nil {
		return nil
	}
	return bh
}

type RewardTotal struct {
	CommunityReward uint64 //社区节点奖励
	LightReward     uint64 //轻节点奖励
	StartHeight     uint64 //统计的开始区块高度
	Height          uint64 //最新区块高度
	IsGrant         bool   //是否可以分发奖励，24小时后才可以分发奖励
	AllLight        uint64 //所有轻节点数量
	RewardLight     uint64 //已经奖励的轻节点数量
}

/*
	奖励统计
*/
func GetRewardCount(addr *crypto.AddressCoin, startHeight, endHeight uint64) (*RewardTotal, *[]sqlite3_db.Reward, error) {

	rewardDBs := make([]sqlite3_db.Reward, 0)

	//查询新的统计
	bvvo, err := FindLightVote(startHeight, endHeight)
	if err != nil {
		return nil, nil, err
	}

	// lightVout := make([]mining.Vout, 0)
	allReward := uint64(0)
	for _, one := range bvvo.Group {
		for _, one := range one.CommunityVotes {
			if bytes.Equal(*one.Addr, *addr) {
				//找到这个社区
				allReward += one.Reward

				for _, two := range one.LightVotes {
					temp := new(big.Int).Mul(big.NewInt(int64(one.Reward)), big.NewInt(int64(two.Scores)))
					value := new(big.Int).Div(temp, big.NewInt(int64(one.Vote)))
					reward := value.Uint64()
					r := sqlite3_db.Reward{
						Addr:         two.Addr.B58String(), //轻节点地址
						Reward:       reward,               //自己奖励多少
						Distribution: 0,                    //已经分配的奖励
					}
					rewardDBs = append(rewardDBs, r)
				}

				break
			}
		}
	}

	community := allReward / 10
	light := allReward - community

	reward := RewardTotal{
		CommunityReward: community,                                                                                         //社区节点奖励
		LightReward:     light,                                                                                             //轻节点奖励
		StartHeight:     startHeight,                                                                                       //
		Height:          bvvo.EndHeight,                                                                                    //最新区块高度
		IsGrant:         (bvvo.EndHeight - startHeight) > (config.Mining_community_reward_time / config.Mining_block_time), //是否可以分发奖励，24小时后才可以分发奖励
		AllLight:        0,                                                                                                 //所有轻节点数量
		RewardLight:     0,                                                                                                 //已经奖励的轻节点数量
	}

	//合并奖励
	voutMap := make(map[string]*sqlite3_db.Reward)
	for i, one := range rewardDBs {
		if one.Reward == 0 {
			continue
		}
		v, ok := voutMap[one.Addr]
		if ok {
			v.Reward = v.Reward + one.Reward
			continue
		}
		voutMap[one.Addr] = &(rewardDBs)[i]
	}
	vouts := make([]sqlite3_db.Reward, 0)
	for _, v := range voutMap {
		vouts = append(vouts, *v)
	}
	return &reward, &vouts, nil
}

/*
	从数据库获取未分配完的奖励和统计
*/
func FindNotSendReward(addr *crypto.AddressCoin) (*sqlite3_db.Snapshot, *[]sqlite3_db.Reward, error) {
	// startHeight := uint64(0)
	//查询最新的快照
	s, err := new(sqlite3_db.Snapshot).Find(addr.B58String())
	if err != nil {
		return nil, nil, err
	}
	if s == nil {
		return nil, nil, nil
	} else {
		rewardNotSend, err := new(sqlite3_db.Reward).FindNotSend(s.Id)
		if err != nil {
			return nil, nil, err
		}
		return s, rewardNotSend, nil

	}
}

/*
	创建轻节点奖励快照
*/
func CreateRewardCount(addr string, rt *RewardTotal, rs []sqlite3_db.Reward) error {
	ss := &sqlite3_db.Snapshot{
		Addr:        addr,            //社区节点地址
		StartHeight: rt.StartHeight,  //快照开始高度
		EndHeight:   rt.Height,       //快照结束高度
		Reward:      rt.LightReward,  //此快照的总共奖励
		LightNum:    uint64(len(rs)), //
	}

	err := new(sqlite3_db.Snapshot).Add(ss)
	if err != nil {
		return err
	}

	ss, err = new(sqlite3_db.Snapshot).Find(addr)
	if err != nil {
		return err
	}

	count := uint64(0)
	for _, one := range rs {
		count++
		one.Sort = count
		one.SnapshotId = ss.Id
		err := new(sqlite3_db.Reward).Add(&one)
		if err != nil {
			//TODO 事务回滚
			return err
		}
	}
	return nil
}

/*
	分配奖励
*/
func DistributionReward(notSend *[]sqlite3_db.Reward, gas uint64, pwd string, cs CommunitySign) error {
	max := len(*notSend)
	if len(*notSend) > config.Mining_pay_vout_max {
		max = config.Mining_pay_vout_max
	}
	//计算平摊的手续费
	value := new(big.Int).Div(big.NewInt(int64(gas)), big.NewInt(int64(max))).Uint64()

	payNum := make([]PayNumber, 0)
	for i := 0; i < max; i++ {
		one := (*notSend)[i]
		addr := crypto.AddressFromB58String(one.Addr)
		payOne := PayNumber{
			Address: addr,               //转账地址
			Amount:  one.Reward - value, //转账金额
		}
		payNum = append(payNum, payOne)
	}

	_, err := SendToMoreAddressByPayload(payNum, gas, pwd, cs)
	return err
}
