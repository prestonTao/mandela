package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/hex"
	"errors"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/go-xorm/xorm"
)

/*
	地址余额管理器
*/
type BalanceManager struct {
	chain         *Chain              //链引用
	syncBlockHead chan *BlockHeadVO   //正在同步的余额，准备导入到余额中
	balance       *sync.Map           //保存各个地址的余额，key:string=收款地址;value:*Balance=收益列表;
	depositin     *TxItem             //保存成为见证人押金交易
	votein        *sync.Map           //保存本节点投票的押金额度，key:string=见证人地址;value:*Balance=押金列表;
	witnessBackup *WitnessBackup      //
	txManager     *TransactionManager //
	otherDeposit  *sync.Map           //其他押金，key:uint64=交易类型;value:*sync.Map=押金列表;
}

func NewBalanceManager(wb *WitnessBackup, tm *TransactionManager, chain *Chain) *BalanceManager {
	bm := &BalanceManager{
		chain:         chain,
		syncBlockHead: make(chan *BlockHeadVO, 1), //正在同步的余额，准备导入到余额中
		balance:       new(sync.Map),              //保存各个地址的余额，key:string=收款地址;value:*Balance=收益列表;
		witnessBackup: wb,                         //
		txManager:     tm,                         //
		votein:        new(sync.Map),              //
		otherDeposit:  new(sync.Map),              //
	}
	go bm.run()
	return bm
}

/*
	保存一个地址的余额列表
	一个地址余额等于多个交易输出相加
*/
type Balance struct {
	Addr *crypto.AddressCoin //
	Txs  *sync.Map           //key:string=交易id;value:*TxItem=交易详细
}

/*
	交易列表
*/
type TxItem struct {
	Addr     *crypto.AddressCoin //收款地址
	Value    uint64              //余额
	Txid     []byte              //交易id
	OutIndex uint64              //交易输出index，从0开始
	Height   uint64              //区块高度，排序用
	VoteType uint16              //投票类型
	// WitnessAddr *crypto.AddressCoin //给谁投票的见证人地址
}

type TxItemSort []*TxItem

func (this *TxItemSort) Len() int {
	return len(*this)
}

/*
	value值大的排在前面
*/
func (this *TxItemSort) Less(i, j int) bool {
	if (*this)[i].Value < (*this)[j].Value {
		return false
	} else {
		return true
	}
}

func (this *TxItemSort) Swap(i, j int) {
	(*this)[i], (*this)[j] = (*this)[j], (*this)[i]
}

/*
	获取一个地址的余额列表
*/
func (this *BalanceManager) FindBalanceOne(addr *crypto.AddressCoin) *Balance {
	bas := this.FindBalance(addr)
	if bas == nil || len(bas) < 1 {
		// fmt.Println("这里错误222")
		return nil
	}
	return bas[0]
}

/*
	获取一个地址的押金列表
*/
func (this *BalanceManager) GetDepositIn() *TxItem {
	return this.depositin
}

/*
	获取一个地址的押金列表
*/
func (this *BalanceManager) GetVoteIn(witnessAddr string) *Balance {
	v, ok := this.votein.Load(witnessAddr)
	if !ok {
		return nil
	}
	b := v.(*Balance)
	return b
}

/*
	获取一个地址的押金列表
*/
func (this *BalanceManager) GetVoteInByTxid(txid string) *TxItem {
	var tx *TxItem
	this.votein.Range(func(k, v interface{}) bool {
		b := v.(*Balance)
		b.Txs.Range(func(txidItr, v interface{}) bool {
			dstTxid := txidItr.(string)
			//0600000000000000b027d84883693a16de4df892c4d856cbf103ed0e28a2d5d98277199ea2d79345_0
			if txid == strings.SplitN(dstTxid, "_", 2)[0] {
				tx = v.(*TxItem)
				return false
			}
			return true
		})
		if tx != nil {
			return false
		}
		return true
	})
	return tx
}

/*
	从最后一个块开始统计多个地址的余额
*/
func (this *BalanceManager) FindBalance(addrs ...*crypto.AddressCoin) []*Balance {
	bas := make([]*Balance, 0)
	for _, one := range addrs {
		v, ok := this.balance.Load(one.B58String())
		if ok {
			b := v.(*Balance)
			bas = append(bas, b)
			continue
		}
	}
	return bas
}

/*
	构建付款输入
*/
func (this *BalanceManager) BuildPayVin(amount uint64) (uint64, []*TxItem) {

	var tis TxItemSort = make([]*TxItem, 0)

	keys := keystore.GetAddrAll()
	for _, one := range keys {
		bas := this.FindBalance(&one)
		for _, two := range bas {
			two.Txs.Range(func(k, v interface{}) bool {
				item := v.(*TxItem)

				tis = append(tis, item)

				return true
			})
		}
	}

	sort.Sort(&tis)

	total := uint64(0)
	items := make([]*TxItem, 0)
	if tis[0].Value >= amount {
		item := tis[0]
		for i, one := range tis {
			if amount >= one.Value {
				item = tis[i]
			} else {
				items = append(items, item)
				total = item.Value
				break
			}
		}
	} else {
		for i, one := range tis {
			items = append(items, tis[i])
			total = total + one.Value
			if total >= amount {
				break
			}
		}
	}
	return total, items
}

/*
	引入最新的块
	将交易计入余额
	使用过的UTXO余额删除
*/
func (this *BalanceManager) CountBalanceForBlock(bhvo *BlockHeadVO) {
	this.countBalance(bhvo)

	//给已经确认的区块建立高度索引
	db.Save([]byte(config.BlockHeight+strconv.Itoa(int(bhvo.BH.Height))), &bhvo.BH.Hash)
}

func (this *BalanceManager) run() {
	for bhvo := range this.syncBlockHead {
		this.countBalance(bhvo)
	}
}

/*
	开始统计余额
*/
func (this *BalanceManager) countBalance(bhvo *BlockHeadVO) {
	//		fmt.Println("开始解析余额 111111")
	//	atomic.StoreUint64(&this.syncHeight, bhvo.BH.Height)

	//统计社区奖励
	this.countCommunityReward(bhvo)

	for txIndex, txItr := range bhvo.Txs {
		txItr.BuildHash()

		//将之前的UTXO标记为已经使用，余额中减去。
		for _, vin := range *txItr.GetVin() {

			//是区块奖励
			if txItr.Class() == config.Wallet_tx_type_mining {
				continue
			}

			//将之前的UTXO标记为已经使用
			txbs, err := db.Find(vin.Txid)
			if err != nil {
				engine.Log.Error("Mark this transaction as used, find the transaction error %s %s %d", err.Error(), hex.EncodeToString(vin.Txid), vin.Vout)
			} else {
				preTxItr, err := ParseTxBase(txbs)
				if err != nil {
					engine.Log.Error("Mark this transaction as used, parsing error %s %s %d", err.Error(), hex.EncodeToString(vin.Txid), vin.Vout)
				} else {
					err = preTxItr.SetTxid(txbs, vin.Vout, txItr.GetHash())
					if err != nil {
						engine.Log.Error("Mark this transaction as used error %s %s %d", err.Error(), hex.EncodeToString(vin.Txid), vin.Vout)
					}
				}
			}

			//检查地址是否是自己的
			addr, isSelf := vin.ValidateAddr()
			if !isSelf {
				continue
			}
			//查找这个地址的余额列表，没有则创建一个
			v, ok := this.balance.Load(addr.B58String())
			var ba *Balance
			if ok {
				ba = v.(*Balance)
			} else {
				ba = new(Balance)
				ba.Txs = new(sync.Map)
			}
			// fmt.Println("vout 数量", len(*txItr.GetVout()), vin.Vout)

			// fmt.Println("删除掉的交易余额", hex.EncodeToString(vin.Txid)+"_"+strconv.Itoa(int(vin.Vout)))
			ba.Txs.Delete(hex.EncodeToString(vin.Txid) + "_" + strconv.Itoa(int(vin.Vout)))
			this.balance.Store(addr.B58String(), ba)

		}

		txCtrl := GetTransactionCtrl(txItr.Class())
		if txCtrl != nil {

			txCtrl.CountBalance(this.balance, this.otherDeposit, bhvo, uint64(txIndex))

			continue
		}
		//其他类型交易，自己节点不支持，直接当做普通交易处理

		voutAddrs := make([]*crypto.AddressCoin, 0)
		for i, one := range *txItr.GetVout() {
			//如果地址是自己的，就可以不用显示
			if keystore.FindAddress(one.Address) {
				continue
			}
			voutAddrs = append(voutAddrs, &(*txItr.GetVout())[i].Address)
		}

		vinAddrs := make([]*crypto.AddressCoin, 0)

		hasPayOut := false //是否有支付类型转出记录
		//将之前的UTXO标记为已经使用，余额中减去。
		for _, vin := range *txItr.GetVin() {

			//检查地址是否是自己的
			addr, isSelf := vin.ValidateAddr()
			// fmt.Println(addr)
			vinAddrs = append(vinAddrs, addr)
			if !isSelf {
				continue
			}

			//是区块奖励
			if txItr.Class() == config.Wallet_tx_type_mining {
				continue
			}

			bs, err := db.Find(vin.Txid)
			if err != nil {
				//TODO 不能找到上一个交易，程序出错退出
				continue
			}
			preTxItr, err := ParseTxBase(bs)
			if err != nil {
				//TODO 不能解析上一个交易，程序出错退出
				continue
			}

			switch txItr.Class() {
			case config.Wallet_tx_type_mining:
			case config.Wallet_tx_type_deposit_in:
			case config.Wallet_tx_type_deposit_out:

				if preTxItr.Class() == config.Wallet_tx_type_deposit_in {
					if this.depositin != nil {
						if bytes.Equal(*addr, *this.depositin.Addr) {
							this.depositin = nil
						}
					}
				}
			case config.Wallet_tx_type_pay:
				if !hasPayOut {
					//和自己相关的输入地址
					vinAddrsSelf := make([]*crypto.AddressCoin, 0)
					for _, vin := range *txItr.GetVin() {
						//检查地址是否是自己的
						addr, isSelf := vin.ValidateAddr()
						// fmt.Println(addr)
						if isSelf {
							vinAddrsSelf = append(vinAddrsSelf, addr)
						}
					}

					temp := make([]*crypto.AddressCoin, 0) //转出地址
					amount := uint64(0)                    //转出金额
					for i, one := range *txItr.GetVout() {
						//如果地址是自己的，就可以不用显示
						if keystore.FindAddress(one.Address) {
							continue
						}
						temp = append(temp, &(*txItr.GetVout())[i].Address)
						amount = amount + one.Value
					}
					if amount > 0 {
						//将转出保存历史记录
						hi := HistoryItem{
							IsIn:    false,         //资金转入转出方向，true=转入;false=转出;
							Type:    txItr.Class(), //交易类型
							InAddr:  temp,          //输入地址
							OutAddr: vinAddrsSelf,  //输出地址
							// Value:   (*preTxItr.GetVout())[vin.Vout].Value, //交易金额
							Value:  amount,           //交易金额
							Txid:   *txItr.GetHash(), //交易id
							Height: bhvo.BH.Height,   //
							// OutIndex: uint64(voutIndex),           //交易输出index，从0开始
						}
						this.chain.history.Add(hi)
						// engine.Log.Info("转出记录", bhvo.BH.Height, hi, (*preTxItr.GetVout())[vin.Vout].Value)
					}
					hasPayOut = true
				}
			case config.Wallet_tx_type_vote_in:
			case config.Wallet_tx_type_vote_out:

				if preTxItr.Class() == config.Wallet_tx_type_vote_in {
					votein := preTxItr.(*Tx_vote_in)
					b, ok := this.votein.Load(votein.Vote.B58String())
					if ok {
						ba := b.(*Balance)
						ba.Txs.Delete(hex.EncodeToString(*preTxItr.GetHash()) + "_" + strconv.Itoa(int(vin.Vout)))
						this.votein.Store(votein.Vote.B58String(), ba)
					}
				}

			}
		}
		//生成新的UTXO收益，保存到列表中
		for voutIndex, vout := range *txItr.GetVout() {
			//找出需要统计余额的地址

			//和自己无关的地址
			if !keystore.FindAddress(vout.Address) {
				continue
			}

			txItem := TxItem{
				Addr:     &(*txItr.GetVout())[voutIndex].Address, //  &vout.Address,
				Value:    vout.Value,                             //余额
				Txid:     *txItr.GetHash(),                       //交易id
				OutIndex: uint64(voutIndex),                      //交易输出index，从0开始
				Height:   bhvo.BH.Height,                         //
			}

			// fmt.Println("放入内存的txid为", base64.StdEncoding.EncodeToString(*txItr.GetHash()))

			switch txItr.Class() {
			case config.Wallet_tx_type_mining:
				//保存历史记录
				//*如果是找零的记录不用保存历史记录
				hi := HistoryItem{
					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
					Type:    txItr.Class(),                        //交易类型
					InAddr:  []*crypto.AddressCoin{&vout.Address}, //输入地址
					OutAddr: nil,                                  //输出地址
					Value:   vout.Value,                           //交易金额
					Txid:    *txItr.GetHash(),                     //交易id
					Height:  bhvo.BH.Height,                       //
					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
				}
				this.chain.history.Add(hi)
			case config.Wallet_tx_type_deposit_in:
				if voutIndex == 0 {
					this.depositin = &txItem
					//创始节点直接打开挖矿
					if config.InitNode {
						config.AlreadyMining = true
					}
					//自己提交见证人押金后，再打开出块的开关
					if config.SubmitDepositin {
						config.AlreadyMining = true
					}
					continue
				}
			case config.Wallet_tx_type_deposit_out:
				//保存历史记录
				//*如果是找零的记录不用保存历史记录
				hi := HistoryItem{
					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
					Type:    txItr.Class(),                        //交易类型
					InAddr:  vinAddrs,                             //输入地址
					OutAddr: []*crypto.AddressCoin{&vout.Address}, //输出地址
					Value:   vout.Value,                           //交易金额
					Txid:    *txItr.GetHash(),                     //交易id
					Height:  bhvo.BH.Height,                       //
					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
				}
				this.chain.history.Add(hi)
			case config.Wallet_tx_type_pay:
				//判断是否是找零地址，判断依据是输入地址是否有自己钱包的地址
				have := false
				for _, one := range vinAddrs {
					if keystore.FindAddress(*one) {
						have = true
						break
					}
				}

				//保存历史记录
				//*如果是找零的记录不用保存历史记录
				hi := HistoryItem{
					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
					Type:    txItr.Class(),                        //交易类型
					InAddr:  []*crypto.AddressCoin{&vout.Address}, //
					OutAddr: vinAddrs,                             //输出地址
					Value:   vout.Value,                           //交易金额
					Txid:    *txItr.GetHash(),                     //交易id
					Height:  bhvo.BH.Height,                       //
					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
				}
				if !have {
					hi.InAddr = []*crypto.AddressCoin{&vout.Address}
					this.chain.history.Add(hi)
				}

			case config.Wallet_tx_type_vote_in:
				if voutIndex == 0 {
					voteIn := txItr.(*Tx_vote_in)

					witnessAddr := voteIn.Vote.B58String()

					// engine.Log.Info("----------------------------------保存投票" + witnessAddr + "end")

					v, ok := this.votein.Load(witnessAddr)
					var ba *Balance
					if ok {
						ba = v.(*Balance)
					} else {
						ba = new(Balance)
						addr := voteIn.Vote.GetAddress()
						ba.Addr = &addr
						ba.Txs = new(sync.Map)
					}
					txItem.VoteType = voteIn.VoteType
					ba.Txs.Store(hex.EncodeToString(*txItr.GetHash())+"_"+strconv.Itoa(voutIndex), &txItem)
					this.votein.Store(witnessAddr, ba)
					continue
				}
			case config.Wallet_tx_type_vote_out:
				//保存历史记录
				//*如果是找零的记录不用保存历史记录
				hi := HistoryItem{
					IsIn:    true,                                 //资金转入转出方向，true=转入;false=转出;
					Type:    txItr.Class(),                        //交易类型
					InAddr:  []*crypto.AddressCoin{&vout.Address}, //输入地址
					OutAddr: vinAddrs,                             //输出地址
					Value:   vout.Value,                           //交易金额
					Txid:    *txItr.GetHash(),                     //交易id
					Height:  bhvo.BH.Height,                       //
					// OutIndex: uint64(voutIndex),                    //交易输出index，从0开始
				}
				this.chain.history.Add(hi)
			}

			//计入余额列表
			v, ok := this.balance.Load(vout.Address.B58String())
			var ba *Balance
			if ok {
				ba = v.(*Balance)
			} else {
				ba = new(Balance)
				ba.Txs = new(sync.Map)
			}
			ba.Txs.Store(hex.EncodeToString(*txItr.GetHash())+"_"+strconv.Itoa(voutIndex), &txItem)
			this.balance.Store(vout.Address.B58String(), ba)
			// fmt.Println("保存转入历史记录")
			// fmt.Println(vinAddrs)

		}

	}

	//TODO 纯粹的统计，发布版本去掉
	// total := uint64(0)
	// key := keystore.GetCoinbase()
	// bas := this.FindBalance(&key)
	// for _, one := range bas {
	// 	one.Txs.Range(func(k, v interface{}) bool {
	// 		ba := v.(*TxItem)
	// 		// fmt.Println("余额+", hex.EncodeToString(ba.Txid), ba.Value)
	// 		total += ba.Value
	// 		return true
	// 	})
	// }
	// fmt.Println("引入新的交易后 余额", total, "高度", bhvo.BH.Height)
}

/*
	统计社区奖励
	@bhvo    *BlockHeadVO    区块头和所有交易
*/
func (this *BalanceManager) countCommunityReward(bhvo *BlockHeadVO) {
	for _, txItr := range bhvo.Txs {
		//判断交易类型
		if txItr.Class() != config.Wallet_tx_type_pay {
			continue
		}
		//检查签名
		addr, ok, cs := CheckPayload(txItr)
		if !ok {
			//签名不正确
			continue
		}
		//判断地址是否属于自己
		if !keystore.FindAddress(addr) {
			//签名者地址不属于自己
			continue
		}

		//判断有没有这个快照
		sn, _, err := FindNotSendReward(&addr)
		if err != nil && err.Error() != xorm.ErrNotExist.Error() {
			engine.Log.Error("querying database Error %s", err.Error())
			return
		}
		//同步快照
		if sn == nil || sn.EndHeight < cs.EndHeight {
			//创建快照
			rt, r, err := GetRewardCount(&addr, cs.StartHeight, cs.EndHeight)
			if err != nil {
				return
			}
			err = CreateRewardCount(addr.B58String(), rt, *r)
			if err != nil {
				return
			}
		}
		//

		//
		// for _, vout := range *txItr.GetVout() {

		// }
	}

}

/*
	缴纳押金，并广播
*/
func (this *BalanceManager) DepositIn(amount, gas uint64, pwd, payload string) error {
	key := keystore.GetCoinbase()

	//不能重复提交押金
	if this.depositin != nil {
		return errors.New("Deposit cannot be paid repeatedly")
	}
	if this.txManager.FindDeposit(hex.EncodeToString(key)) {
		return errors.New("Deposit cannot be paid repeatedly")
	}
	if amount < config.Mining_deposit {
		return errors.New("Deposit not less than" + strconv.Itoa(int(uint64(config.Mining_deposit)/Unit)))
	}

	deposiIn, err := CreateTxDepositIn(amount, gas, pwd, payload)
	if err != nil {
		return err
	}
	if deposiIn == nil {
		//		fmt.Println("33333333333333 22222")
		return errors.New("Failure to pay deposit")
	}
	deposiIn.BuildHash()
	bs, err := deposiIn.Json()
	if err != nil {
		//		fmt.Println("33333333333333 33333")
		return err
	}
	//	fmt.Println("4444444444444444")
	MulticastTx(bs)
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
	this.txManager.AddTx(txbase)
	// fmt.Println("添加押金是否成功", ok)
	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	//	fmt.Println("7777777777777777")
	return nil
}

/*
	退还押金，并广播
*/
func (this *BalanceManager) DepositOut(addr string, amount, gas uint64, pwd string) error {
	//	key, err := keystore.GetCoinbase()
	//	if err != nil {
	//		return err
	//	}
	if this.depositin == nil {
		return errors.New("I didn't pay the deposit")
	}

	deposiOut, err := CreateTxDepositOut(addr, amount, gas, pwd)
	if err != nil {
		return err
	}
	if deposiOut == nil {
		//		fmt.Println("33333333333333 22222")
		return errors.New("Failure to pay deposit")
	}
	deposiOut.BuildHash()
	bs, err := deposiOut.Json()
	if err != nil {
		//		fmt.Println("33333333333333 33333")
		return err
	}
	//	fmt.Println("4444444444444444")
	MulticastTx(bs)
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
	this.txManager.AddTx(txbase)
	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	//	fmt.Println("7777777777777777")
	return nil
}

/*
	投票押金，并广播
	不能自己给自己投票，统计票的时候会造成循环引用
	给见证人投票的都是社区节点，一个社区节点只能给一个见证人投票。
	给社区节点投票的都是轻节点，轻节点投票前先缴押金。
	轻节点可以给轻节点投票，相当于一个轻节点尾随另一个轻节点投票。
	引用关系不能出现死循环
	@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
*/
func (this *BalanceManager) VoteIn(voteType uint16, witnessAddr crypto.AddressCoin, addr string, amount, gas uint64, pwd, payload string) error {

	//不能自己给自己投票
	if witnessAddr.B58String() == addr {
		return errors.New("You can't vote for yourself")
	}
	dstAddr := crypto.AddressFromB58String(addr)

	isWitness := this.witnessBackup.haveWitness(&dstAddr)
	_, isCommunity := this.witnessBackup.haveCommunityList(&dstAddr)
	_, isLight := this.witnessBackup.haveLight(&dstAddr)
	// fmt.Println("查看自己的角色", addr, isWitness, isCommunity, isLight)
	switch voteType {
	case 1: //1=给见证人投票
		if isLight || isWitness {
			return errors.New("The voting address is already another role")
		}
		vs, ok := this.witnessBackup.haveCommunityList(&dstAddr)
		if ok {
			if bytes.Equal(*vs.Witness, witnessAddr) {
				return errors.New("Can't vote again")
			}
			return errors.New("Cannot vote for multiple witnesses")
		}
		//检查押金
		if amount != config.Mining_vote {
			return errors.New("Community node deposit is " + strconv.Itoa(int(config.Mining_vote/1e8)))
		}

	case 2: //2=给社区节点投票

		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		//检查是否成为轻节点
		if !isLight {
			//先成为轻节点
			return errors.New("Become a light node first")
		}

		vs, ok := this.witnessBackup.haveVoteList(&dstAddr)
		if ok {
			if !bytes.Equal(*vs.Witness, witnessAddr) {
				//不能给多个社区节点投票
				return errors.New("Cannot vote for multiple community nodes")
			}
		}

	case 3: //3=轻节点押金
		if isCommunity || isWitness {
			//投票地址已经是其他角色了
			return errors.New("The voting address is already another role")
		}
		if isLight {
			//已经是轻节点了
			return errors.New("It's already a light node")
		}
		// engine.Log.Info("轻节点押金是 %d %d", amount, config.Mining_light_min)

		if amount != config.Mining_light_min {
			//轻节点押金是
			return errors.New("Light node deposit is " + strconv.Itoa(int(config.Mining_light_min/1e8)))
		}
		witnessAddr = nil
	default:
		//不能识别的投票类型
		return errors.New("Unrecognized voting type")

	}

	voetIn, err := CreateTxVoteIn(voteType, witnessAddr, addr, amount, gas, pwd, payload)
	if err != nil {
		return err
	}
	if voetIn == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	voetIn.BuildHash()
	bs, err := voetIn.Json()
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
	this.txManager.AddTx(txbase)
	MulticastTx(bs)
	// fmt.Println("添加投票押金是否成功", ok)
	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	//	fmt.Println("7777777777777777")
	return nil
}

// /*
// 	退还一笔投票押金，并广播
// */
// func (this *BalanceManager) VoteOutOne(txid, addr string, amount, gas uint64, pwd string) error {
// 	tx := this.GetVoteInByTxid(txid)
// 	if tx == nil {
// 		return errors.New("没有找到这个交易")
// 	}
// 	deposiOut, err := CreateTxVoteOutOne(tx, addr, amount, gas, pwd)
// 	if err != nil {
// 		return err
// 	}
// 	if deposiOut == nil {
// 		//		fmt.Println("33333333333333 22222")
// 		return errors.New("退押金失败")
// 	}
// 	deposiOut.BuildHash()
// 	bs, err := deposiOut.Json()
// 	if err != nil {
// 		//		fmt.Println("33333333333333 33333")
// 		return err
// 	}
// 	//	fmt.Println("4444444444444444")
// 	MulticastTx(bs)
// 	//	fmt.Println("5555555555555555")
// 	txbase, err := ParseTxBase(bs)
// 	if err != nil {
// 		return err
// 	}
// 	txbase.BuildHash()
// 	//	fmt.Println("66666666666666")
// 	//验证交易
// 	if !txbase.Check() {
// 		//交易不合法，则不发送出去
// 		// fmt.Println("交易不合法，则不发送出去")
// 		return errors.New("交易不合法，则不发送出去")
// 	}
// 	this.txManager.AddTx(txbase)
// 	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
// 	//	fmt.Println("7777777777777777")
// 	return nil
// 	return nil

// }

/*
	退还投票押金，并广播
*/
func (this *BalanceManager) VoteOut(witnessAddr *crypto.AddressCoin, txid, addr string, amount, gas uint64, pwd string) error {

	waddr := ""

	if witnessAddr != nil && witnessAddr.B58String() != "" {
		waddr = witnessAddr.B58String()
	}
	// engine.Log.Info("---------------------------查询这个见证人" + waddr + "end")
	balance := this.GetVoteIn(waddr)
	if balance == nil {
		//没有对这个见证人投票
		return errors.New("No vote for this witness")
	}

	deposiOut, err := CreateTxVoteOut(witnessAddr, txid, addr, amount, gas, pwd)
	if err != nil {
		return err
	}
	if deposiOut == nil {
		//交押金失败
		return errors.New("Failure to pay deposit")
	}
	deposiOut.BuildHash()
	bs, err := deposiOut.Json()
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
	this.txManager.AddTx(txbase)
	MulticastTx(bs)
	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	//	fmt.Println("7777777777777777")
	return nil
}

/*
	构建一个其他交易，并广播
*/
func (this *BalanceManager) BuildOtherTx(class uint64, addr *crypto.AddressCoin, amount, gas uint64, pwd string, params ...interface{}) error {

	ctrl := GetTransactionCtrl(class)
	txItr, err := ctrl.BuildTx(this.balance, this.otherDeposit, addr, amount, gas, pwd, params...)
	if err != nil {
		//		fmt.Println("33333333333333 22222")
		return err
	}
	txItr.BuildHash()
	bs, err := txItr.Json()
	if err != nil {
		//		fmt.Println("33333333333333 33333")
		return err
	}
	//	fmt.Println("4444444444444444")
	MulticastTx(bs)
	//	fmt.Println("5555555555555555")
	txbase, err := ParseTxBase(bs)
	if err != nil {
		return err
	}
	txbase.BuildHash()
	// fmt.Println("66666666666666", txbase.Class())
	//验证交易
	if err := txbase.Check(); err != nil {
		//交易不合法，则不发送出去
		// fmt.Println("交易不合法，则不发送出去")
		return err
	}

	// if !ctrl.Check(txbase) {
	// 	//交易不合法，则不发送出去
	// 	fmt.Println("交易不合法，则不发送出去")
	// 	return errors.New("交易不合法，则不发送出去")
	// }
	this.txManager.AddTx(txbase)
	// fmt.Println("添加投票押金是否成功", ok)
	//		unpackedTransactions.Store(hex.EncodeToString(*txbase.GetHash()), txbase)
	//	fmt.Println("7777777777777777")
	return nil
}

/*
	获得自己轻节点押金列表
*/
func (this *BalanceManager) GetVoteList() []*Balance {
	balances := make([]*Balance, 0)
	this.votein.Range(func(k, v interface{}) bool {
		b := v.(*Balance)
		balances = append(balances, b)
		return true
	})
	return balances
}
