package mining

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"strconv"
	"sync"
	"time"
)

//合并txitem开关信号
// var txitemMerge = make(chan bool, 1)
var txitemMergeIsOpenLock = new(sync.Mutex)
var txitemMergeIsOpen = false
var txitemMergePwd = ""
var txitemMergeUnifieAddr *crypto.AddressCoin
var txitemMergeTotalMax = uint64(0)
var txitemMergeGas = uint64(0)

func init() {

	// go LoopTxItemMerge()
	utils.Go(LoopTxItemMerge)
}

/*
	打开合并功能
*/
func SwitchOpenMergeTx(pwd string, gas uint64, unifieAddr *crypto.AddressCoin, totalMax uint64) {
	txitemMergeIsOpenLock.Lock()
	txitemMergeIsOpen = true
	txitemMergePwd = pwd
	txitemMergeGas = gas
	if unifieAddr == nil {
		txitemMergeUnifieAddr = nil
		txitemMergeTotalMax = 0
	} else {
		txitemMergeUnifieAddr = unifieAddr
		if totalMax <= 0 || totalMax > config.Mining_pay_vin_max {
			txitemMergeTotalMax = config.Mining_pay_vin_max
		} else {
			txitemMergeTotalMax = totalMax
		}
	}
	txitemMergeIsOpenLock.Unlock()
}

/*
	关闭合并功能
*/
func SwitchCloseMergeTx(pwd string) {
	txitemMergeIsOpenLock.Lock()
	txitemMergeIsOpen = false
	txitemMergePwd = pwd
	txitemMergeIsOpenLock.Unlock()
}

func LoopTxItemMerge() {
	total := 0
	unfinished := make(chan bool, 1)
	randMinute := utils.GetRandNum(10)
	intervalChan := time.NewTicker(time.Minute * time.Duration(5+randMinute)).C
	// intervalChan := time.NewTicker(time.Minute * 1).C
	for {
		// engine.Log.Info("1111111111111111")
		select {
		// case isOpen := <-txitemMerge:
		case <-intervalChan:
			total = 0
		case <-unfinished:
			total++
			time.Sleep(time.Minute * 2)
		}
		if total >= 3 { //每轮只能提交3比交易
			// engine.Log.Info("22222222222222222")
			continue
		}
		time.Sleep(time.Second * 2)

		txitemMergeIsOpenLock.Lock()
		isOpen := txitemMergeIsOpen
		unifieAddr := txitemMergeUnifieAddr
		totalMax := txitemMergeTotalMax
		gas := txitemMergeGas
		pwd := txitemMergePwd
		engine.Log.Info("check tx merge is open:%v", isOpen)
		if !isOpen {
			engine.Log.Info("check tx merge is open:%v", isOpen)
			//见证人默认打开此功能
			chain := GetLongChain()
			if chain != nil && chain.witnessChain.FindWitness(keystore.GetCoinbase().Addr) {
				isOpen = true
				txitemMergePwd = config.Wallet_keystore_default_pwd
				engine.Log.Info("check tx merge is open:%v", isOpen)
			}

			// if chain != nil {
			// 	addr := keystore.GetCoinbase().Addr
			// 	if chain.witnessBackup.FindWitness(addr) && chain.witnessChain.FindWitness(addr) && chain.witnessBackup.FindWitnessInBlackList(addr) {
			// 		isOpen = true
			// 		engine.Log.Info("check tx merge is open:%v", isOpen)
			// 	}
			// }
		}
		txitemMergeIsOpenLock.Unlock()
		if !isOpen {
			engine.Log.Info("333333333333333333")
			continue
		}

		chain := GetLongChain()
		if chain == nil {
			engine.Log.Info("4444444444444444")
			continue
		}
		if !chain.SyncBlockFinish {
			engine.Log.Info("5555555555555555")
			continue
		}
		engine.Log.Info("666666666666666666666")

		if unifieAddr != nil {
			MergeOneceUnifie(unifieAddr, totalMax, gas, pwd)
			continue
		}

		txItemsTotal := 0 //所有地址items总个数
		addrTotal := 0    //地址个数
		txitems := chain.GetBalance().notspentBalance.FindBalanceNotSpentAll()
		notSpentBalance := make(map[string]map[string]*TxItem)
		for _, one := range txitems {
			txItemsTotal++
			key := utils.Bytes2string(one.Txid) + "_" + strconv.Itoa(int(one.VoutIndex))

			items, ok := notSpentBalance[utils.Bytes2string(*one.Addr)]
			if !ok {
				items = make(map[string]*TxItem)
				notSpentBalance[utils.Bytes2string(*one.Addr)] = items
				addrTotal++
			}
			items[key] = one
		}

		if txItemsTotal < config.Mining_pay_vin_max {
			// engine.Log.Info("7777777777777777 txItemsTotal:%d", txItemsTotal)
			continue
		}
		if txItemsTotal <= 0 || addrTotal <= 0 {
			// engine.Log.Info("8888888888888888")
			continue
		}

		//地址数量大于10个，每个地址都要合并
		if addrTotal >= 10 {
			// engine.Log.Info("99999999999999999999")
			for k, items := range notSpentBalance {
				if len(items) <= 2 {
					continue
				}
				time.Sleep(time.Second * 2)
				// amountTotal := uint64(0)
				vinTotal := 0
				newitems := make([]*TxItem, 0)
				// if len(items) > config.Mining_pay_vin_max {
				// 	newitems = items[:config.Mining_pay_vin_max]
				// }
				for _, item := range items {
					newitems = append(newitems, item)
					// amountTotal += item.Value
					vinTotal++
					if vinTotal >= config.Mining_pay_vin_max {
						break
					}
				}
				// engine.Log.Info("99999 1111111111111111")
				addrCoin := crypto.AddressCoin([]byte(k))
				_, err := MergeTx(newitems, &addrCoin, gas, 0, txitemMergePwd, "")
				if err != nil {
					engine.Log.Info("merge tx error:%s", err.Error())
					continue
				}
				if len(items) > config.Mining_pay_vin_max {
					select {
					case unfinished <- false:
					default:
					}
				}
			}
		} else {
			// engine.Log.Info("10101010101010101010101010101010")
			averageNumber := config.Mining_pay_vin_max / addrTotal
			for k, items := range notSpentBalance {
				if len(items) < averageNumber {
					// engine.Log.Info("1111111111111111")
					continue
				}
				//大于平均数，就要合并了
				time.Sleep(time.Second * 2)
				// amountTotal := uint64(0)
				vinTotal := 0
				newitems := make([]*TxItem, 0)
				// if len(items) > config.Mining_pay_vin_max {
				// 	newitems = items[:config.Mining_pay_vin_max]
				// }
				for _, item := range items {
					newitems = append(newitems, item)
					// amountTotal += item.Value
					vinTotal++
					if vinTotal >= config.Mining_pay_vin_max {
						break
					}
				}
				// engine.Log.Info("2222222222222222222222222")
				addrCoin := crypto.AddressCoin([]byte(k))
				_, err := MergeTx(newitems, &addrCoin, gas, 0, txitemMergePwd, "")
				if err != nil {
					engine.Log.Info("merge tx error:%s", err.Error())
					continue
				}
				if len(items) > config.Mining_pay_vin_max {
					select {
					case unfinished <- false:
					default:
					}
				}
			}
		}

	}
}

/*
	所有地址的余额归集到指定地址
	@unifieAddr    *crypto.AddressCoin    归集到统一的地址
	@totalMax      uint64                 utxo超过阈值则开始归集操作
*/
func MergeOneceUnifie(unifieAddr *crypto.AddressCoin, totalMax, gas uint64, pwd string) {
	chain := GetLongChain()
	isRun := false
	txitems := chain.GetBalance().notspentBalance.FindBalanceNotSpentAll()
	newitems := make([]*TxItem, 0, config.Mining_pay_vin_max)
	for _, one := range txitems {
		newitems = append(newitems, one)
		if len(newitems) >= int(totalMax) {
			isRun = true
		}
		if len(newitems) == config.Mining_pay_vin_max {
			break
		}
	}
	if !isRun {
		return
	}
	_, err := MergeTx(newitems, unifieAddr, gas, 0, pwd, "")
	if err != nil {
		engine.Log.Info("merge tx error:%s", err.Error())
	}
}
