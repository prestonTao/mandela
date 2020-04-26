package mining

import (
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"math/big"
	// "mandela/core/engine"
)

/*
	候选见证人组
	保存已经入选的见证人和未选上的见证人
*/
type WitnessBackupGroup struct {
	Witnesses     []*Witness //
	WitnessBackup []*Witness //备用见证人
}

/*
	构建本组中的见证人出块奖励
	按股权分配
	只有见证人方式出块才统计
	组人数乘以每块奖励，再分配给实际出块的人
*/
func (this *WitnessBackupGroup) BuildRewardVouts(blocks []Block) []Vout {
	vouts := make([]Vout, 0)

	witneses := make([]*crypto.AddressCoin, 0) //已经出块的见证人
	// allWitness := make([]*crypto.AddressCoin, 0) //所有见证人，包括已经出块的，和未选上的候选见证人
	allCommiunty := make([]*VoteScore, 0) //保存所有社区节点地址及投票数量

	//统计本组股权和交易手续费
	allVotePos := uint64(0) //所有投票者票数总和
	allGas := uint64(0)     //计算交易手续费
	allReward := uint64(0)  //本组奖励数量

	//计算出块奖励总和
	for _, one := range this.Witnesses {
		// allWitness = append(allWitness, one.Addr)
		//统计所有社区节点投票
		for _, v := range one.CommunityVotes {
			v.Scores = 0
			allCommiunty = append(allCommiunty, v)
			allVotePos = allVotePos + v.Vote
			// engine.Log.Info("统计社区节点投票 %s %s %d", v.Addr.B58String(), v.Witness.B58String(), v.Vote)
		}

		//是否在未确认的区块中
		isUnconfirmed := false
		//判断是否在正在出块的见证人组里面
		nowWitnessGroup := GetLongChain().witnessChain.witnessGroup
		if nowWitnessGroup.Height == one.Group.Height {
			for _, oneBlock := range blocks {
				//高度相同，见证人地址相同
				if oneBlock.Group.Height == one.Group.Height && bytes.Equal(*one.Addr, *oneBlock.witness.Addr) {
					isUnconfirmed = true
					break
				}
			}
			if !isUnconfirmed {
				//在未确认的见证人组中，但是没有找到这个见证人出块，则不奖励
				continue
			}
		}

		if !isUnconfirmed {

			//只计算已经出块的见证人奖励
			if one.Block == nil {
				continue
			}
			//不能只简单通过 one.Block == nil 来判断未出块
			if one.Block.Group == nil {
				continue
			}
			//还要判断是否在已经确认的组里面
			ok, group := one.Group.CheckBlockGroup()
			if !ok {
				continue
			}
			//判断这个见证人出的块是否在已经确认的组里面
			if one.Block.Group != group {
				continue
			}

		}
		//这个见证人的出块已经得到确认
		witneses = append(witneses, one.Addr)

		//计算交易手续费
		_, txs, _ := one.Block.LoadTxs()
		for _, one := range *txs {
			allGas = allGas + one.GetGas()
			// engine.Log.Info("统计区块gas %d, %d", one.GetGas(), allGas)
		}

		//按照发行总量及减半周期计算出块奖励
		oneReward := config.ClacRewardForBlockHeight(one.Block.Height)
		allReward = allReward + oneReward
		// engine.Log.Info("统计这个区块奖励 %d, %d", oneReward, allReward)

	}

	//区块奖励
	allReward = allReward + allGas

	// engine.Log.Info("block reward allReward:%d allVotePos:%d", allReward, allVotePos)

	// 10%              给资源节点
	// 90% * 10% * 20%  平均分给出块的见证人。
	// 90% * 10% * 80%  平均分给所有见证人，包括候选见证人。
	// 90% * 90%        给所有社区节点，由社区节点按投票数量分配。

	//--------------分给资源节点---------------
	//检查资源节点是否存在
	resourcesReward := uint64(0)
	nameinfo := name.FindNameToNet(config.Name_store)
	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
		//给资源节点。
		resourcesReward = new(big.Int).Div(big.NewInt(int64(allReward)), big.NewInt(int64(10))).Uint64()

		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]
		vout := Vout{
			Value:   resourcesReward,
			Address: addrCoin,
		}
		vouts = append(vouts, vout)
		allReward = allReward - resourcesReward
		// engine.Log.Info("resource node reward %s %d", vout.Address.B58String(), vout.Value)
	}
	// engine.Log.Info("reward shenxia %d", allReward)

	//---------------------------------------------------
	//平均分给出块的见证人。
	temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(2)))
	value := new(big.Int).Div(temp, big.NewInt(int64(100)))
	witnessReward := value.Uint64()

	//平均分给出块的见证人
	// engine.Log.Info("Distribute equally to the witnesses of the block %d", witnessReward)

	//按股权分给所有见证人，包括候选见证人。
	temp = new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(8)))
	value = new(big.Int).Div(temp, big.NewInt(int64(100)))
	allWitnessReward := value.Uint64()

	//平均分给所有见证人，包括候选见证人
	// engine.Log.Info("Evenly distributed to all witnesses, including candidate witnesses %d", allWitnessReward)

	//给所有社区节点，由社区节点按投票数量分配。
	allCommiuntyReward := allReward - allWitnessReward - witnessReward

	//给所有社区节点分
	// engine.Log.Info("To all community nodes %d", allCommiuntyReward)

	//---------------------------------------------------
	//给出块的见证人平均分。
	use := uint64(0)
	temp = new(big.Int).Mul(big.NewInt(int64(witnessReward)), big.NewInt(int64(1)))
	value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses))))
	oneReward := value.Uint64()
	for _, one := range this.Witnesses {
		//给所有已经出块的见证人平均分
		// engine.Log.Info("Give the witness average score of the block %s %d", one.Addr.B58String(), oneReward)
		use = use + oneReward
		vout := Vout{
			Value:   oneReward,
			Address: *one.Addr,
		}
		vouts = append(vouts, vout)
	}
	//平均数不能被整除时候，剩下的给最后一个输出奖励
	if len(vouts) > 0 {
		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (witnessReward - use)
	}

	// engine.Log.Info("开始给见证人分配 1111111111")
	//---------------------------------------------------
	//给所有候选见证人按投票股权分配。
	use = uint64(0)
	if allVotePos <= 0 {
		temp = new(big.Int).Mul(big.NewInt(int64(allWitnessReward)), big.NewInt(1))
		value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses)+len(this.WitnessBackup))))
		oneReward = value.Uint64()
		for _, one := range append(this.Witnesses, this.WitnessBackup...) {
			//给所有候选见证人平均分
			// engine.Log.Info("Average all candidate witnesses %s %d", one.Addr.B58String(), oneReward)
			use = use + oneReward
			vout := Vout{
				Value:   oneReward,
				Address: *one.Addr,
			}
			vouts = append(vouts, vout)
		}
	} else {
		for _, one := range append(this.Witnesses, this.WitnessBackup...) {
			temp = new(big.Int).Mul(big.NewInt(int64(allWitnessReward)), big.NewInt(int64(one.VoteNum)))
			value = new(big.Int).Div(temp, big.NewInt(int64(allVotePos)))
			oneReward = value.Uint64()
			//给所有候选见证人平均分
			// engine.Log.Info("Average all candidate witnesses %s %s", one.Addr.B58String(), oneReward)
			use = use + oneReward
			vout := Vout{
				Value:   oneReward,
				Address: *one.Addr,
			}
			vouts = append(vouts, vout)
		}
	}
	//平均数不能被整除时候，剩下的给最后一个输出奖励
	if len(vouts) > 0 {
		// engine.Log.Info("加余数 %d %d", use, allWitnessReward-use)
		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (allWitnessReward - use)
	}

	// engine.Log.Info("开始给见证人分配 2222222222222")

	//---------------------------------------------------
	//给所有社区节点，由社区节点按投票数量分配。
	use = uint64(0)
	//如果所有投票数量为0，则将这部分收益分给所有候选见证人。
	if allVotePos <= 0 {
		// engine.Log.Info("开始给见证人分配 33333333333333")
		//给所有候选见证人
		temp = new(big.Int).Mul(big.NewInt(int64(allCommiuntyReward)), big.NewInt(int64(1)))
		value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses))))
		oneReward = value.Uint64()
		for i, _ := range this.Witnesses {
			use = use + oneReward
			vout := Vout{
				Value:   oneReward,
				Address: *this.Witnesses[i].Addr,
			}
			vouts = append(vouts, vout)
			//开始给见证人分配
			// engine.Log.Info("Start to assign witness %s %d", this.Witnesses[i].Addr.B58String(), oneReward)
		}
	} else {
		// engine.Log.Info("开始给见证人分配 4444444444444")
		for i, one := range allCommiunty {
			//给所有社区节点参数
			// engine.Log.Info("Give all community node parameters %d %d %d", allCommiuntyReward, one.Vote, allVotePos)
			if one.Vote == 0 {
				continue
			}
			temp = new(big.Int).Mul(big.NewInt(int64(allCommiuntyReward)), big.NewInt(int64(one.Vote)))
			value = new(big.Int).Div(temp, big.NewInt(int64(allVotePos)))
			oneReward = value.Uint64()
			//给所有社区节点
			// engine.Log.Info("To all community nodes %s %d", allCommiunty[i].Addr.B58String(), oneReward)
			use = use + oneReward
			vout := Vout{
				Value:   oneReward,
				Address: *allCommiunty[i].Addr,
			}
			vouts = append(vouts, vout)
		}
	}
	//平均数不能被整除时候，剩下的给最后一个输出奖励
	if len(vouts) > 0 {
		// engine.Log.Info("加余数 %d %d", use, allCommiuntyReward-use)
		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (allCommiuntyReward - use)
	}

	return MergeVouts(&vouts)
}

/*
	构建本组中的见证人出块奖励
	按股权分配
	只有见证人方式出块才统计
	组人数乘以每块奖励，再分配给实际出块的人
*/
func (this *WitnessBackupGroup) CountRewardToWitnessGroup(blockHeight uint64, blocks []Block) *Tx_reward {

	//构建区块奖励输出
	engine.Log.Info("build reward")
	vouts := this.BuildRewardVouts(blocks)

	//构建输入
	baseCoinAddr := keystore.GetCoinbase()
	puk, ok := keystore.GetPukByAddr(baseCoinAddr)
	if !ok {
		return nil
	}
	vins := make([]Vin, 0)
	vin := Vin{
		Puk:  puk, //公钥
		Sign: nil, //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
	vins = append(vins, vin)

	var txReward *Tx_reward
	for i := uint64(0); i < 10000; i++ {
		base := TxBase{
			Type:       config.Wallet_tx_type_mining, //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
			Vin_total:  1,
			Vin:        vins,
			Vout_total: uint64(len(vouts)), //输出交易数量
			Vout:       vouts,              //交易输出
			LockHeight: blockHeight + i,    //锁定高度
			//		CreateTime: time.Now().Unix(),            //创建时间
		}
		txReward = &Tx_reward{
			TxBase: base,
		}

		//合并交易输出
		txReward.MergeVout()

		//给输出签名，防篡改
		for i, one := range txReward.Vin {
			for _, key := range keystore.GetAddrAll() {

				puk, ok := keystore.GetPukByAddr(key)
				if !ok {
					return nil
				}

				if bytes.Equal(puk, one.Puk) {
					_, prk, _, err := keystore.GetKeyByAddr(key, config.Wallet_keystore_default_pwd)
					// prk, err := key.GetPriKey(pwd)
					if err != nil {
						return nil
					}
					sign := txReward.GetSign(&prk, one.Txid, one.Vout, uint64(i))
					//				sign := pay.GetVoutsSign(prk, uint64(i))
					txReward.Vin[i].Sign = *sign
				}
			}
		}

		txReward.BuildHash()
		if txReward.CheckHashExist() {
			txReward = nil
			continue
		} else {
			break
		}
	}
	return txReward
}
