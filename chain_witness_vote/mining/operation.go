package mining

import (
	"bytes"
	"errors"
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/protos/go_protos"
	"mandela/sqlite3_db"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"golang.org/x/crypto/ed25519"
)

/*
	获取账户所有地址的余额
	@return    uint64    可用余额
	@return    uint64    冻结余额
	@return    uint64    锁仓余额
*/
func FindBalanceValue() (uint64, uint64, uint64) {
	chain := forks.GetLongChain()
	if chain == nil {
		return 0, 0, 0
	}
	return chain.GetBalance().FindBalanceValue()
}

/*
	获取账户所有地址的余额
	@return    uint64    可用余额
	@return    uint64    冻结余额
	@return    uint64    锁仓余额
*/
// func GetBalances() (uint64, uint64, uint64) {
// 	count := uint64(0)
// 	countf := uint64(0)
// 	countLockup := uint64(0)

// 	chain := forks.GetLongChain()
// 	if chain == nil {
// 		return 0, 0, 0
// 	}

// 	txitems, bfs, itemsLockup := chain.GetBalance().FindBalanceAll()
// 	for _, one := range txitems {
// 		count = count + one.Value
// 	}
// 	//统计冻结的余额
// 	for _, one := range bfs {
// 		countf = countf + one.Value
// 	}
// 	//统计锁仓的余额
// 	for _, one := range itemsLockup {
// 		countLockup = countLockup + one.Value
// 	}

// 	// engine.Log.Info("打印各种余额 %d %d", count, countf)

// 	return count, countf, countLockup
// }

/*
	获取所有item
*/
// func GetBalanceAllItems() ([]*TxItem, []*TxItem, []*TxItem) {
// 	chain := forks.GetLongChain()
// 	if chain == nil {
// 		return nil, nil, nil
// 	}
// 	return chain.GetBalance().FindBalanceAll()
// }

/*
	获取所有地址余额明细
*/
func GetBalanceAllAddrs() (map[string]uint64, map[string]uint64, map[string]uint64) {
	chain := forks.GetLongChain()
	if chain == nil {
		// basMap := make(map[string]uint64)      //可用余额
		// fbasMap := make(map[string]uint64)     //冻结的余额
		// baLockupMap := make(map[string]uint64) //锁仓的余额
		return make(map[string]uint64), make(map[string]uint64), make(map[string]uint64)
	}
	return chain.GetBalance().FindBalanceAll()
}

/*
	通过地址获取余额
*/
func GetBalanceForAddr(addr crypto.AddressCoin) (uint64, uint64) {
	// count := uint64(0)
	// countf := uint64(0)

	chain := forks.GetLongChain()
	if chain == nil {
		return 0, 0
	}
	return chain.balance.FindBalance(addr)

	// txitems, bfs := chain.balance.FindBalance(addr)
	// for _, item := range txitems {
	// 	count = count + item.Value
	// }

	// for _, item := range bfs {
	// 	// engine.Log.Info("统计冻结 111 %d", item.Value)
	// 	countf = countf + item.Value
	// }
	// // engine.Log.Info("统计冻结 222 %d", countf)

	// return count, countf
}

/*
	获取区块是否同步完成
*/
// func GetSyncFinish() bool {
// 	//判断是否同步完成
// 	if forks.GetHighestBlock() <= 0 {
// 		//区块未同步完成，不能挖矿
// 		// engine.Log.Info("区块最新高度 %d", forks.GetHighestBlock())
// 		return false
// 	}
// 	chain := forks.GetLongChain()
// 	if forks.GetHighestBlock() > chain.GetPulledStates() {
// 		//区块未同步完成，不能挖矿
// 		// engine.Log.Info("区块最新高度 %d 区块同步高度 %d", forks.GetHighestBlock(), chain.GetPulledStates())
// 		return false
// 	}
// 	return true
// }

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
	打开合并交易功能
*/
func SendToAddress(srcAddress, address, change *crypto.AddressCoin, amount, gas, frozenHeight uint64, pwd, comment string) (*Tx_Pay, error) {
	// return nil, errors.New("hahahaha")

	txpay, err := CreateTxPay(srcAddress, address, change, amount, gas, frozenHeight, pwd, comment)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()
	// engine.Log.Info("create tx finish!")
	forks.GetLongChain().transactionManager.AddTx(txpay)

	MulticastTx(txpay)
	// engine.Log.Info("multicast tx finish!")
	// utils.PprofMem()
	return txpay, nil
}

/*
	给一个地址转账
*/
func MergeTx(items []*TxItem, address *crypto.AddressCoin, gas, frozenHeight uint64, pwd, comment string) (*Tx_Pay, error) {
	// return nil, errors.New("hahahaha")
	// engine.Log.Info("开始合并交易")
	txpay, err := MergeTxPay(items, address, gas, frozenHeight, pwd, comment)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()
	// engine.Log.Info("create tx finish!")
	forks.GetLongChain().transactionManager.AddTx(txpay)

	MulticastTx(txpay)
	// engine.Log.Info("multicast tx finish!")
	// utils.PprofMem()
	return txpay, nil
}

/*
	给多个收款地址转账
*/
func SendToMoreAddress(srcAddr crypto.AddressCoin, address []PayNumber, gas uint64, pwd, comment string) (*Tx_Pay, error) {
	txpay, err := CreateTxsPay(srcAddr, address, gas, pwd, comment)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()

	forks.GetLongChain().transactionManager.AddTx(txpay)
	MulticastTx(txpay)

	return txpay, nil
}

/*
	给多个地址转账,带Payload签名
*/
func SendToMoreAddressByPayload(address []PayNumber, gas uint64, pwd string, cs *CommunitySign) (*Tx_Pay, error) {
	txpay, err := CreateTxsPayByPayload(address, gas, pwd, cs)
	if err != nil {
		// fmt.Println("创建交易失败", err)
		return nil, err
	}
	txpay.BuildHash()

	forks.GetLongChain().transactionManager.AddTx(txpay)
	MulticastTx(txpay)

	//	unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	return txpay, nil
}

/*
	从邻居节点查询起始区块hash
*/
func FindStartBlockForNeighbor() *ChainInfo {
	for _, key := range nodeStore.GetLogicNodes() {

		message, _ := message_center.SendNeighborMsg(config.MSGID_getBlockHead, &key, nil)

		// bs := flood.WaitRequest(mc.CLASS_getBlockHead, hex.EncodeToString(message.Body.Hash), 0)
		bs, _ := flood.WaitRequest(mc.CLASS_getBlockHead, utils.Bytes2string(message.Body.Hash), 0)
		// fmt.Println("有消息返回了啊")
		if bs == nil {
			// fmt.Println("从邻居节点查询起始区块hash 发送共享文件消息失败，可能超时")
			continue
		}
		chainInfo, err := ParseChainInfo(bs)
		// chainInfo := new(ChainInfo)
		// // var jso = jsoniter.ConfigCompatibleWithStandardLibrary
		// // err := json.Unmarshal(*bs, chainInfo)
		// decoder := json.NewDecoder(bytes.NewBuffer(*bs))
		// decoder.UseNumber()
		// err := decoder.Decode(chainInfo)
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
func FindBlockForNeighbor(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
	var bhvo *BlockHeadVO
	var bs *[]byte
	var err error
	var newBhvo *BlockHeadVO

	//根据超时时间排序
	// logicNodes := nodeStore.GetLogicNodes()
	// logicNodes = append(logicNodes, nodeStore.GetNodesClient()...)
	// logicNodesInfo := core.SortNetAddrForSpeed(logicNodes)
	//工作模式，false=询问未超时节点;true=询问超时节点;
	//如果不切换工作模式，会导致节点同步永远落后几个区块高度

	// mode := false
	// for i := 0; i < 2; i++ {

	// TAG:
	peers := peerBlockInfo.Sort()
	addrs := make([]nodeStore.AddressNet, 0)
	for _, one := range peers {
		addrs = append(addrs, *one.Addr)
	}
	logicNodesInfo := core.SortNetAddrForSpeed(addrs)

	for i, _ := range logicNodesInfo {
		//询问未超时节点工作模式下：遇到超时的节点，则退出
		// if !mode && one.Speed >= int64(time.Second*config.Wallet_sync_block_timeout) {
		// 	continue
		// }
		//询问超时节点工作模式下：遇到未超时的节点，则退出
		// if mode && one.Speed < int64(time.Second*config.Wallet_sync_block_timeout) {
		// 	continue
		// }
		key := &logicNodesInfo[i].AddrNet
		// engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
		engine.Log.Info("Send query message to node %s", key.B58String())
		bs, err = getValueForNeighbor(*key, bhash)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
			continue
		}
		if bs == nil {
			engine.Log.Info("Send query message to node from:%s bs is nil", key.B58String())
			continue
		}
		newBhvo, err = ParseBlockHeadVOProto(bs)
		// newBhvo, err = ParseBlockHeadVO(bs)
		if err != nil {
			engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
			continue
		}
		bhvo = newBhvo
		//检查本区块是否有nextHash
		if newBhvo.BH.Nextblockhash != nil && len(newBhvo.BH.Nextblockhash) > 0 {
			// engine.Log.Info("this block nextblock not nil")
			return newBhvo, err
		}
		//为空也返回
		// return newBhvo, nil
	}
	//如果从，未超时的节点同步到区块，则不继续同步了
	if bhvo != nil {
		return bhvo, nil
	}
	//TODO 超时的节点会永久打入冷宫，给超时的节点一个翻身的机会
	//如果从，未超时的节点，都没同步到区块，则尝试已超时的节点。
	// if !mode {
	// 	engine.Log.Info("switch sync timeout nodes mode")
	// 	mode = true
	// 	// goto TAG
	// 	continue
	// }
	// }
	// if bs != nil {
	// 	engine.Log.Info("this block nextblock nil %s", string(*bs))
	// }
	return bhvo, err
}

/*
	查询邻居节点已经导入的最高区块
*/
func FindLastBlockForNeighbor(peerBlockInfo *PeerBlockInfoDESC) (*BlockHead, error) {
	var err error
	var bh *BlockHead

	peers := peerBlockInfo.Sort()
	addrs := make([]nodeStore.AddressNet, 0)
	for _, one := range peers {
		addrs = append(addrs, *one.Addr)
	}
	logicNodesInfo := core.SortNetAddrForSpeed(addrs)

	for i, _ := range logicNodesInfo {
		key := &logicNodesInfo[i].AddrNet

		message, _ := message_center.SendNeighborMsg(config.MSGID_getBlockLastCurrent, key, nil)
		bs, _ := flood.WaitRequest(mc.CLASS_getBlockLastCurrent, utils.Bytes2string(message.Body.Hash), config.Wallet_sync_block_timeout)
		if bs == nil {
			continue
		}
		bh, err = ParseBlockHeadProto(bs)
		if err != nil {
			continue
		}
		return bh, nil
	}
	return nil, err
}

// func FindBlockForNeighbor(bhash *[]byte, peerBlockInfo *PeerBlockInfoDESC) (*BlockHeadVO, error) {
// 	var bhvo *BlockHeadVO
// 	var bs *[]byte
// 	var err error
// 	var newBhvo *BlockHeadVO

// 	//根据超时时间排序
// 	logicNodes := nodeStore.GetLogicNodes()
// 	logicNodes = append(logicNodes, nodeStore.GetNodesClient()...)
// 	logicNodesInfo := core.SortNetAddrForSpeed(logicNodes)
// 	//工作模式，false=询问未超时节点;true=询问超时节点;
// 	//如果不切换工作模式，会导致节点同步永远落后几个区块高度
// 	mode := false
// 	for i := 0; i < 2; i++ {

// 		// TAG:
// 		for _, one := range logicNodesInfo {
// 			//询问未超时节点工作模式下：遇到超时的节点，则退出
// 			if !mode && one.Speed >= int64(time.Second*config.Wallet_sync_block_timeout) {
// 				continue
// 			}
// 			//询问超时节点工作模式下：遇到未超时的节点，则退出
// 			if mode && one.Speed < int64(time.Second*config.Wallet_sync_block_timeout) {
// 				continue
// 			}
// 			key := one.AddrNet
// 			// engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
// 			engine.Log.Info("Send query message to node %s", key.B58String())
// 			bs, err = getValueForNeighbor(key, bhash)
// 			if err != nil {
// 				engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
// 				continue
// 			}
// 			if bs == nil {
// 				engine.Log.Info("Send query message to node from:%s bs is nil", key.B58String())
// 				continue
// 			}
// 			newBhvo, err = ParseBlockHeadVOProto(bs)
// 			// newBhvo, err = ParseBlockHeadVO(bs)
// 			if err != nil {
// 				engine.Log.Info("Send query message to node from:%s error:%s", key.B58String(), err.Error())
// 				continue
// 			}
// 			bhvo = newBhvo
// 			//检查本区块是否有nextHash
// 			if newBhvo.BH.Nextblockhash != nil && len(newBhvo.BH.Nextblockhash) > 0 {
// 				// engine.Log.Info("this block nextblock not nil")
// 				return newBhvo, err
// 			}
// 			//为空也返回
// 			// return newBhvo, nil
// 		}
// 		//如果从，未超时的节点同步到区块，则不继续同步了
// 		if bhvo != nil {
// 			return bhvo, nil
// 		}
// 		//TODO 超时的节点会永久打入冷宫，给超时的节点一个翻身的机会
// 		//如果从，未超时的节点，都没同步到区块，则尝试已超时的节点。
// 		if !mode {
// 			engine.Log.Info("switch sync timeout nodes mode")
// 			mode = true
// 			// goto TAG
// 			continue
// 		}
// 	}
// 	if bs != nil {
// 		engine.Log.Info("this block nextblock nil %s", string(*bs))
// 	}
// 	return bhvo, err
// }

// func FindBlockForNeighbor(bhash *[]byte) *BlockHeadVO {
// 	bhvo := new(BlockHeadVO)
// 	bs := getValueForNeighbor(bhash)
// 	if bs == nil {
// 		engine.Log.Info("Error synchronizing chunk from neighbor node")
// 		return nil
// 	}
// 	bhvo, err := ParseBlockHeadVO(bs)
// 	if err != nil {
// 		return nil
// 	}
// 	return bhvo
// }

/*
	查询邻居节点数据库，key：value查询
*/
func getValueForNeighbor(key nodeStore.AddressNet, bhash *[]byte) (*[]byte, error) {

	start := time.Now()

	message, _ := message_center.SendNeighborMsg(config.MSGID_getTransaction, &key, bhash)
	// engine.Log.Info("44444444444 %s", key.B58String())
	// bs := flood.WaitRequest(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), config.Mining_block_time)
	bs, _ := flood.WaitRequest(mc.CLASS_getTransaction, utils.Bytes2string(message.Body.Hash), config.Wallet_sync_block_timeout)
	if bs == nil {
		endTime := time.Now()
		// engine.Log.Info("5555555555555555 %s", key.B58String())
		//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
		engine.Log.Error("Receive %s message timeout %s", key.B58String(), time.Now().Sub(start))
		//有可能是对方没有查询到区块，返回空，则判定它超时
		if (endTime.Unix() - start.Unix()) < config.Wallet_sync_block_timeout {
			core.AddNodeAddrSpeed(key, time.Second*(config.Wallet_sync_block_timeout+1))
		} else {
			//TODO 应该取平均数
			//保存上一次同步超时时间
			// config.NetSpeedMap.Store(utils.Bytes2string(key), time.Now().Sub(start))
			core.AddNodeAddrSpeed(key, time.Now().Sub(start))
		}
		// err = errors.New("Failed to send shared file message, may timeout")

		return nil, config.ERROR_chain_sync_block_timeout
	}
	core.AddNodeAddrSpeed(key, time.Now().Sub(start))
	// engine.Log.Info("Receive message %s", key.B58String())
	return bs, nil
}

// func getValueForNeighbor(bhash *[]byte) *[]byte {
// 	// fmt.Println("1查询区块或交易", hex.EncodeToString(*bhash))
// 	var bs *[]byte
// 	var err error
// 	for {
// 		logicNodes := nodeStore.GetLogicNodes()
// 		logicNodes = OrderNodeAddr(logicNodes)
// 		for _, key := range logicNodes {
// 			engine.Log.Info("Find a neighbor node and start synchronizing block data \n" + hex.EncodeToString(*bhash))
// 			engine.Log.Info("Send query message to node %s", key.B58String())

// 			message, _ := message_center.SendNeighborMsg(config.MSGID_getTransaction, key, bhash)
// 			// engine.Log.Info("44444444444 %s", key.B58String())
// 			bs = flood.WaitRequest(mc.CLASS_getTransaction, hex.EncodeToString(message.Body.Hash), config.Mining_block_time)
// 			if bs == nil {
// 				// engine.Log.Info("5555555555555555 %s", key.B58String())
// 				//查询邻居节点数据库，key：value查询 发送共享文件消息失败，可能超时
// 				engine.Log.Info("Receive message timeout %s", key.B58String())
// 				err = errors.New("Failed to send shared file message, may timeout")
// 				continue
// 			}
// 			engine.Log.Info("Receive message %s", key.B58String())
// 			// engine.Log.Info("66666666666666 %s", key.B58String())
// 			err = nil
// 			break
// 		}
// 		// engine.Log.Info("7777777777777777777")
// 		if err == nil {
// 			// engine.Log.Info("888888888888888")
// 			break
// 		}
// 		// engine.Log.Info("99999999999999999999")
// 	}
// 	if err != nil {
// 		// engine.Log.Info("10101010101001010101")
// 		engine.Log.Warn("Failed to query block transaction", hex.EncodeToString(*bhash))
// 	}
// 	// engine.Log.Info("11 11 11 11 11")
// 	// if bs == nil {
// 	// 	engine.Log.Info("查询区块或交易结果 %s", bs)
// 	// } else {
// 	// 	engine.Log.Info("查询区块或交易结果 %s", string(*bs))
// 	// }
// 	return bs
// }

/*
	从邻居节点获取未确认的区块
*/
func GetUnconfirmedBlockForNeighbor(height uint64, peerBlockInfo *PeerBlockInfoDESC) ([]*BlockHeadVO, error) {
	engine.Log.Info("Synchronize unacknowledged chunks from neighbor nodes")

	heightBs := utils.Uint64ToBytes(height)

	var bs *[]byte
	var err error
	// for i := 0; i < 10; i++ {
	// engine.Log.Info("Synchronize unacknowledged total: %d", i)
	// logicNodes := nodeStore.GetLogicNodes()
	logicNodes := peerBlockInfo.Sort()
	for j, _ := range logicNodes {
		engine.Log.Info("Synchronize unacknowledged from:%s height:%d", logicNodes[j].Addr.B58String(), height)
		// <<<<<<< HEAD
		// 		message, ok := message_center.SendNeighborMsg(config.MSGID_getUnconfirmedBlock, logicNodes[j].Addr, &heightBs)
		// 		if !ok {
		// =======
		message, err := message_center.SendNeighborMsg(config.MSGID_getUnconfirmedBlock, logicNodes[j].Addr, &heightBs)
		if err != nil {
			// >>>>>>> dev
			//消息未发送成功
			continue
		}

		// bs = flood.WaitRequest(mc.CLASS_getUnconfirmedBlock, hex.EncodeToString(message.Body.Hash), 0)
		bs, _ = flood.WaitRequest(mc.CLASS_getUnconfirmedBlock, utils.Bytes2string(message.Body.Hash), 0)
		if bs == nil {
			engine.Log.Info("Failed to get unconfirmed block from neighbor node, sending shared file message, maybe timeout")
			err = errors.New("Failed to get unconfirmed block from neighbor node, sending shared file message, maybe timeout")
			continue
		} else {
			err = nil
		}
		engine.Log.Info("Synchronize unacknowledged ok")
		break
	}
	// if err == nil {
	// engine.Log.Info("Synchronize unacknowledged exist")
	// break
	// }
	// }
	// engine.Log.Info("获取的未确认区块 bs", string(*bs))
	if bs == nil {
		engine.Log.Warn("Get unacknowledged block BS error")
		return nil, err
	}

	rbsp := new(go_protos.RepeatedBytes)
	err = proto.Unmarshal(*bs, rbsp)
	if err != nil {
		engine.Log.Warn("Get unacknowledged block BS error:", err.Error())
		return nil, err
	}

	blockHeadVOs := make([]*BlockHeadVO, 0)
	for _, one := range rbsp.Bss {
		bhvo, err := ParseBlockHeadVOProto(&one)
		if err != nil {
			engine.Log.Warn("Get unacknowledged block BS error:", err.Error())
			return nil, err
		}
		blockHeadVOs = append(blockHeadVOs, bhvo)
	}

	// temp := make([]interface{}, 0)
	// // err = json.Unmarshal(*bs, &temp)
	// decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	// decoder.UseNumber()
	// err = decoder.Decode(&temp)

	// blockHeadVOs := make([]*BlockHeadVO, 0)
	// for _, one := range temp {
	// 	bs, err := json.Marshal(one)
	// 	blockVOone, err := ParseBlockHeadVO(&bs)
	// 	if err != nil {
	// 		engine.Log.Warn("Get unacknowledged block BS error:%s", err.Error())
	// 		continue
	// 	}
	// 	blockHeadVOs = append(blockHeadVOs, blockVOone)
	// }

	engine.Log.Info("Get unacknowledged block Success")
	return blockHeadVOs, nil

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
func VoteIn(t uint16, witnessAddr crypto.AddressCoin, addr crypto.AddressCoin, amount, gas uint64, pwd, payload string) error {
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
func VoteOut(txid []byte, addr crypto.AddressCoin, amount, gas uint64, pwd string) error {
	//缴纳备用见证人押金交易
	return forks.GetLongChain().balance.VoteOut(txid, addr, amount, gas, pwd)

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
func GetWitnessStatus() (IsCandidate bool, IsBackup bool, IsKickOut bool, Addr crypto.AddressCoin, value uint64) {
	addrInfo := keystore.GetCoinbase()
	Addr = addrInfo.Addr // .GetAddrStr()
	IsCandidate = forks.GetLongChain().witnessBackup.FindWitness(Addr)
	IsBackup = forks.GetLongChain().witnessChain.FindWitness(addrInfo.Addr)
	IsKickOut = forks.GetLongChain().witnessBackup.FindWitnessInBlackList(Addr)
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
	查询一个交易是否上链成功，以及交易详情
	@return    uint64    交易详情
	@return    uint64    1=未上链;2=成功上链;3=上链失败;
*/
func FindTx(txid []byte) (TxItr, uint64) {

	txItr, err := LoadTxBase(txid)
	if err != nil {
		return nil, 1
	}

	_, err = db.GetTxToBlockHash(&txid)
	if err != nil {
		lockheight := txItr.GetLockHeight()
		//查询当前确认的区块高度
		height := GetLongChain().GetCurrentBlock()
		if height > lockheight {
			//超过了锁定高度还没有上链，则失败了
			return txItr, 3
		}
		return txItr, 1
	}
	return txItr, 2
}

func FindTxJsonVo(txid []byte) (interface{}, uint64) {
	txItr, code := FindTx(txid)
	return txItr.GetVOJSON(), code
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

	ok := forks.GetLongChain().transactionManager.AddTx(txItr)
	if !ok {
		//等待上链,请稍后重试.
		return errors.New("Waiting for the chain, please try again later")
	}
	MulticastTx(txItr)
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
	vins := make([]*Vin, 0)
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
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)

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
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:   amount,   //输出金额 = 实际金额 * 100000000
		Address: *address, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr,
		}
		vouts = append(vouts, &vout)
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
		LockHeight: height + 100,    //锁定高度
		Payload:    []byte(comment), //备注
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
	vins := make([]*Vin, 0)
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
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)

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
	vouts := make([]*Vout, 0)
	for _, one := range address {
		vout := Vout{
			Value:   one.Amount,  //输出金额 = 实际金额 * 100000000
			Address: one.Address, //钱包地址
		}
		vouts = append(vouts, &vout)
	}
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr, //钱包地址
		}
		vouts = append(vouts, &vout)
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
		LockHeight: height + 100,    //锁定高度
		Payload:    []byte(comment), //备注
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
	vins := make([]*Vin, 0)
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
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//					Sign: *sign,           //签名
		}
		vins = append(vins, &vin)

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
	vouts := make([]*Vout, 0)
	vout := Vout{
		Value:   amount,  //输出金额 = 实际金额 * 100000000
		Address: dstAddr, //钱包地址
	}
	vouts = append(vouts, &vout)
	//检查押金是否刚刚好，多了的转账给自己
	//TODO 将剩余款项转入新的地址，保证资金安全
	if total > amount+gas {
		vout := Vout{
			Value: total - amount - gas, //输出金额 = 实际金额 * 100000000
			//Address: keystore.GetAddr()[0], //钱包地址
			Address: returnaddr,
		}
		vouts = append(vouts, &vout)
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
	vins := make([]*Vin, 0)
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
			Txid: item.Txid,      //UTXO 前一个交易的id
			Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
			Puk:  puk,            //公钥
			//			Sign: *sign,         //签名
		}
		vins = append(vins, &vin)

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
				Txid: item.Txid,      //UTXO 前一个交易的id
				Vout: item.VoutIndex, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（第一个为零）
				Puk:  puk,            //公钥
				//						Sign: *sign,           //签名
			}
			vins = append(vins, &vin)

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
	vouts := make([]*Vout, 0)
	//下标为0的交易输出是见证人押金，大于0的输出是多余的钱退还。
	vout := Vout{
		Value:   total - gas, //输出金额 = 实际金额 * 100000000
		Address: dstAddr,     //钱包地址
	}
	vouts = append(vouts, &vout)

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
	// engine.Log.Info("FindLightVote 111111 %d %d", startHeight, endHeight)

	bvVO := &BlockVotesVO{
		EndHeight: endHeight,
		Group:     make([]GroupVO, 0),
	}

	//查找上一个已经确认的组
	var preBlock *Block
	preGroup := forks.LongChain.witnessChain.witnessGroup
	for {
		// engine.Log.Info("FindLightVote 111111")
		preGroup = preGroup.PreGroup
		ok, preGroupBlock := preGroup.CheckBlockGroup(nil)
		if ok {
			preBlock = preGroupBlock.Blocks[len(preGroupBlock.Blocks)-1]
			break
		}
	}
	// engine.Log.Info("FindLightVote 111111")
	//找到查询的起始区块
	for {
		// engine.Log.Info("FindLightVote 111111 height:%d", preBlock.Height)
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
	// engine.Log.Info("FindLightVote 111111")
	//找到这个见证人组的第一个见证人
	for {
		// engine.Log.Info("FindLightVote 111111")
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
	// engine.Log.Info("FindLightVote 111111")
	//找到从start高度开始，到最新高度的所有见证人组的首块
	isFind := false
	for ; preBlock != nil && preBlock.NextBlock != nil && preBlock.Height <= endHeight; preBlock = preBlock.NextBlock {
		// engine.Log.Info("FindLightVote 111111 height:%d", preBlock.Height)
		// if preBlock.NextBlock == nil {
		// 	engine.Log.Info("FindLightVote 111111")
		// 	break
		// }
		if isFind && preBlock.NextBlock.witness.WitnessBackupGroup == preBlock.witness.WitnessBackupGroup {
			// preBlock = preBlock.NextBlock
			// engine.Log.Info("FindLightVote 111111")
			continue
		} else {
			isFind = false
		}
		_, txs, err := preBlock.NextBlock.LoadTxs()
		if err != nil {
			// engine.Log.Info("FindLightVote 111111")
			return nil, err
		}
		if txs == nil || len(*txs) <= 0 {
			// preBlock = preBlock.NextBlock
			// if preBlock.Height > endHeight {
			// 	engine.Log.Info("FindLightVote 111111")
			// 	break
			// }
			continue
		}
		//查找这组见证人中的奖励
		reward, ok := (*txs)[0].(*Tx_reward)
		if !ok {
			continue
		}
		isFind = true

		groupVO := new(GroupVO)
		groupVO.CommunityVotes = make([]VoteScoreRewadrVO, 0)
		// groupVO.StartHeight = preBlock.Height
		groupVO.EndHeight = preBlock.Height
		//组装下一个见证人组的投票
		for _, one := range preBlock.witness.WitnessBackupGroup.Witnesses {
			// engine.Log.Info("FindLightVote 111111")
			m := make(map[string]*[]VoteScore)
			for i, two := range one.Votes {
				vo := VoteScore{
					Witness: one.Votes[i].Witness, //见证人地址。当自己是轻节点的时候，此字段是社区节点地址
					Addr:    one.Votes[i].Addr,    //投票人地址
					Scores:  one.Votes[i].Scores,  //押金
					Vote:    one.Votes[i].Vote,    //获得票数
				}
				v, ok := m[utils.Bytes2string(*two.Witness)]
				if ok {

				} else {
					temp := make([]VoteScore, 0)
					v = &temp
				}
				*v = append(*v, vo)
				m[utils.Bytes2string(*two.Witness)] = v
			}

			for _, two := range one.CommunityVotes {
				// two.
				// engine.Log.Info("FindLightVote 111111")
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

				v, ok := m[utils.Bytes2string(*two.Addr)]
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

		// preBlock = preBlock.NextBlock
		// if preBlock.Height > endHeight {
		// 	engine.Log.Info("FindLightVote 111111")
		// 	break
		// }
	}
	// engine.Log.Info("FindLightVote 111111")
	return bvVO, nil

}

/*
	通过区块高度，查询一个区块头信息
*/
// func FindBlockHead(height uint64) *BlockHead {
// 	bhash, err := db.Find([]byte(config.BlockHeight + strconv.Itoa(int(height))))
// 	if err != nil {
// 		return nil
// 	}
// 	bs, err := db.Find(*bhash)
// 	if err != nil {
// 		return nil
// 	}
// 	bh, err := ParseBlockHead(bs)
// 	if err != nil {
// 		return nil
// 	}
// 	return bh
// }

/*
	从数据库加载一整个区块，包括区块中的所有交易
*/
// func LoadBlockHeadVOByHash(hash []byte) (*BlockHeadVO, error) {
// 	bhvo := new(BlockHeadVO)
// 	bhvo.Txs = make([]TxItr, 0)
// 	//通过区块hash查找区块头
// 	bs, err := db.Find(hash)
// 	if err != nil {
// 		return nil, err
// 	} else {
// 		bh, err := ParseBlockHead(bs)
// 		if err != nil {
// 			return nil, err
// 		}
// 		bhvo.BH = bh
// 		for _, one := range bh.Tx {
// 			txOne, err := FindTxBase(one)
// 			if err != nil {
// 				return nil, err
// 			}
// 			bhvo.Txs = append(bhvo.Txs, txOne)
// 		}
// 	}
// 	return bhvo, nil
// }

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
	奖励结果明细
*/
type RewardTotalDetail struct {
	startHeight uint64
	endHeight   uint64
	RT          *RewardTotal
	RL          *[]sqlite3_db.RewardLight
}

var rewardCountProcessMapLock = new(sync.Mutex)
var rewardCountProcessMap = make(map[string]*RewardTotalDetail) //new(sync.Map) //保存统计社区奖励结果。key:string=社区地址;value:*RewardTotalDetail=奖励结果明细;

/*
	奖励统计
*/
func GetRewardCount(addr *crypto.AddressCoin, startHeight, endHeight uint64) (*RewardTotal, *[]sqlite3_db.RewardLight, error) {
	currentHeight := forks.LongChain.GetCurrentBlock()
	if endHeight <= 0 || endHeight > currentHeight {
		endHeight = currentHeight
	}

	var rt *RewardTotal
	var rl *[]sqlite3_db.RewardLight
	var err error
	//先查询是否有缓存结果
	have := false
	rewardCountProcessMapLock.Lock()
	rtd, ok := rewardCountProcessMap[utils.Bytes2string(*addr)]
	if ok {
		have = true
		if rtd != nil {
			//有缓存，对比结束区块高度是否相差太远，24小时产生86400个块，也就是24小时只有才可以重新创建新的缓存
			if endHeight > rtd.endHeight && endHeight-rtd.endHeight > 86400 {
				have = false
			} else {
				rt = rtd.RT
				rl = rtd.RL
			}
		}
	}
	if !have {
		err = config.ERROR_get_reward_count_sync
		//没有缓存和程序，则启动一个统计程序
		rewardCountProcessMap[utils.Bytes2string(*addr)] = nil
		utils.Go(func() {
			rt, rl, err := RewardCountProcess(addr, startHeight, endHeight)
			if err != nil {
				engine.Log.Info("RewardCountProcess error:%s", err.Error())
				return
			}
			rtd := RewardTotalDetail{
				startHeight: startHeight,
				endHeight:   endHeight,
				RT:          rt,
				RL:          rl,
			}
			rewardCountProcessMapLock.Lock()
			rewardCountProcessMap[utils.Bytes2string(*addr)] = &rtd
			rewardCountProcessMapLock.Unlock()
		})
	} else {
		// err = config.ERROR_get_reward_count_sync
	}
	rewardCountProcessMapLock.Unlock()
	return rt, rl, err
}

/*
	删除缓存
*/
func CleanRewardCountProcessMap(addr *crypto.AddressCoin) {
	rewardCountProcessMapLock.Lock()
	delete(rewardCountProcessMap, utils.Bytes2string(*addr))
	rewardCountProcessMapLock.Unlock()
}

/*
	奖励统计明细，消耗时间长
*/
func RewardCountProcess(addr *crypto.AddressCoin, startHeight, endHeight uint64) (*RewardTotal, *[]sqlite3_db.RewardLight, error) {
	// engine.Log.Info("555555555555555555555")
	rewardDBs := make([]sqlite3_db.RewardLight, 0)

	//查询新的统计
	bvvo, err := FindLightVote(startHeight, endHeight)
	if err != nil {
		// engine.Log.Info("555555555555555555555")
		return nil, nil, err
	}
	// engine.Log.Info("555555555555555555555")
	// lightVout := make([]mining.Vout, 0)
	allReward := uint64(0)
	for _, one := range bvvo.Group {
		for _, one := range one.CommunityVotes {
			//找到这个社区
			if bytes.Equal(*one.Addr, *addr) {
				allReward += one.Reward

				for _, two := range one.LightVotes {
					temp := new(big.Int).Mul(big.NewInt(int64(one.Reward)), big.NewInt(int64(two.Vote)))
					value := new(big.Int).Div(temp, big.NewInt(int64(one.Vote)))
					reward := value.Uint64()
					r := sqlite3_db.RewardLight{
						Addr:         *two.Addr, //轻节点地址
						Reward:       reward,    //自己奖励多少
						Distribution: 0,         //已经分配的奖励
					}
					// engine.Log.Info("统计一个轻节点投票 %s %d %d %d %d %d %d", two.Addr.B58String(), one.Reward, one.Scores, one.Vote, two.Scores, two.Vote, reward)
					rewardDBs = append(rewardDBs, r)
				}

				break
			}
		}
	}
	// engine.Log.Info("555555555555555555555")
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

	//合并奖励，把相同轻节点地址的奖励合并了
	voutMap := make(map[string]*sqlite3_db.RewardLight)
	for i, _ := range rewardDBs {
		one := rewardDBs[i]
		// engine.Log.Info("统计一个轻节点投票 222 %s", one.Addr)
		if one.Reward == 0 {
			continue
		}
		v, ok := voutMap[utils.Bytes2string(one.Addr)]
		if ok {
			v.Reward = v.Reward + one.Reward
			continue
		}
		voutMap[utils.Bytes2string(one.Addr)] = &(rewardDBs)[i]
	}
	vouts := make([]sqlite3_db.RewardLight, 0)
	for _, v := range voutMap {
		// engine.Log.Info("统计一个轻节点投票 333 %s", v.Addr)
		vouts = append(vouts, *v)
	}
	//统计所有轻节点的数量
	reward.AllLight = uint64(len(vouts))
	// engine.Log.Info("统计一个轻节点投票 333 %s", v.Addr)
	// engine.Log.Info("555555555555555555555")
	return &reward, &vouts, nil
}

/*
	从数据库获取未分配完的奖励和统计
*/
func FindNotSendReward(addr *crypto.AddressCoin) (*sqlite3_db.SnapshotReward, *[]sqlite3_db.RewardLight, error) {
	// startHeight := uint64(0)
	//查询最新的快照
	s, err := new(sqlite3_db.SnapshotReward).Find(*addr)
	if err != nil {
		return nil, nil, err
	}
	if s == nil {
		return nil, nil, nil
	} else {
		rewardNotSend, err := new(sqlite3_db.RewardLight).FindNotSend(s.Id)
		if err != nil {
			return nil, nil, err
		}
		return s, rewardNotSend, nil

	}
}

/*
	创建轻节点奖励快照
*/
func CreateRewardCount(addr crypto.AddressCoin, rt *RewardTotal, rs []sqlite3_db.RewardLight) error {
	ss := &sqlite3_db.SnapshotReward{
		Addr:        addr,            //社区节点地址
		StartHeight: rt.StartHeight,  //快照开始高度
		EndHeight:   rt.Height,       //快照结束高度
		Reward:      rt.LightReward,  //此快照的总共奖励
		LightNum:    uint64(len(rs)), //
	}

	err := new(sqlite3_db.SnapshotReward).Add(ss)
	if err != nil {
		return err
	}

	ss, err = new(sqlite3_db.SnapshotReward).Find(addr)
	if err != nil {
		return err
	}

	count := uint64(0)
	for _, one := range rs {
		count++
		one.Sort = count
		one.SnapshotId = ss.Id
		err := new(sqlite3_db.RewardLight).Add(&one)
		if err != nil {
			//TODO 事务回滚
			return err
		}
	}
	return nil
}

/*
	分配奖励
	@height    uint64    当前区块高度，方便计算180天线性释放时间
*/
func DistributionReward(notSend *[]sqlite3_db.RewardLight, gas uint64, pwd string, cs *CommunitySign, height uint64) error {
	// engine.Log.Info("DistributionReward %+v", notSend)
	if notSend == nil || len(*notSend) <= 0 {
		return nil
	}
	max := len(*notSend)

	if len(*notSend) > config.Wallet_community_reward_max {
		max = config.Wallet_community_reward_max
	}

	var tx TxItr
	var err error
	for {
		//计算平摊的手续费
		value := new(big.Int).Div(big.NewInt(int64(gas)), big.NewInt(int64(max))).Uint64()

		payNum := make([]PayNumber, 0)
		for i := 0; i < max; i++ {
			one := (*notSend)[i]
			addr := crypto.AddressCoin(one.Addr)

			pns := LinearRelease0DayForLight(addr, one.Reward-value, height)
			payNum = append(payNum, pns...)

		}
		tx, err = SendToMoreAddressByPayload(payNum, gas, pwd, cs)
		if err != nil {
			if err.Error() == config.ERROR_pay_vin_too_much.Error() {
				if max <= 1 {
					engine.Log.Error(err.Error())
					return err
				}
				max = max / 2
				continue
			}
			engine.Log.Error(err.Error())
			return err
		} else {
			break
		}
	}
	//修改数据库，分配奖励修改为上链中
	for i := 0; i < max; i++ {
		one := (*notSend)[i]
		one.Txid = *tx.GetHash()
		one.LockHeight = tx.GetLockHeight() // LockHeight
		err := one.UpdateTxid(one.Id)
		if err != nil {
			engine.Log.Error(err.Error())
		}
	}
	return err
}

/*
	180天线性释放给轻节点
*/
func LinearRelease180DayForLight(addr crypto.AddressCoin, total uint64, height uint64) []PayNumber {
	//TODO 处理好不能被180整除的情况

	pns := make([]PayNumber, 0)
	//25%直接到账
	first25 := new(big.Int).Div(big.NewInt(int64(total)), big.NewInt(int64(4)))
	//剩下的75%
	surplus := new(big.Int).Sub(big.NewInt(int64(total)), first25)

	// engine.Log.Error("180天线性释放 %d %d %d", total, first25.Uint64(), surplus.Uint64())

	pnOne := PayNumber{
		Address: addr,             //转账地址
		Amount:  first25.Uint64(), //转账金额
		// FrozenHeight: height + uint64(i*intervalHeight), //冻结高度
	}

	pns = append(pns, pnOne)

	dayOne := new(big.Int).Div(surplus, big.NewInt(int64(18))).Uint64()

	// dayOne := new(big.Int).Div(big.NewInt(int64(total)), big.NewInt(int64(180))).Uint64()
	intervalHeight := 60 * 60 * 24 * 10 / 10

	totalUse := uint64(0)
	for i := 0; i < 18; i++ {
		pnOne := PayNumber{
			Address:      addr,                                  //转账地址
			Amount:       dayOne,                                //转账金额
			FrozenHeight: height + uint64((i+1)*intervalHeight), //冻结高度
		}
		pns = append(pns, pnOne)
		totalUse = totalUse + dayOne
	}
	//平均数不能被整除时候，剩下的给最后一个输出奖励
	if totalUse < surplus.Uint64() {
		// engine.Log.Info("加余数 %d %d", use, allCommiuntyReward-use)
		pns[len(pns)-1].Amount = pns[len(pns)-1].Amount + (surplus.Uint64() - totalUse)
	}
	return pns
}

/*
	0天线性释放给轻节点
*/
func LinearRelease0DayForLight(addr crypto.AddressCoin, total uint64, height uint64) []PayNumber {
	pns := make([]PayNumber, 0)
	pnOne := PayNumber{
		Address: addr,  //转账地址
		Amount:  total, //转账金额
	}
	pns = append(pns, pnOne)
	return pns
}
