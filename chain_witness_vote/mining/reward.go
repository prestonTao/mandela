package mining

import (
	"bytes"
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"math/big"
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
	统计本大组已出区块，给他们发放奖励
	@blockhash    *[]byte    本组最后一个区块。从后往前查找连续的区块作为奖励区块
*/
func (this *WitnessBackupGroup) CountRewardWitness(blockhash *[]byte, height uint64) *[]*Witness {
	// engine.Log.Info("preblockhash:%s", hex.EncodeToString(*blockhash))
	rewardWitness := make([]*Witness, 0)
	preBlockHash := blockhash
	if height < config.Reward_witness_height_new {
		for i := len(this.Witnesses); i > 0; i-- {
			witnessOne := this.Witnesses[i-1]
			// engine.Log.Info("check witness:%s", witnessOne.Addr.B58String())
			if witnessOne.Block == nil {
				// engine.Log.Info("this witness block is nil!")
				continue
			}
			if bytes.Equal(*preBlockHash, witnessOne.Block.Id) {
				// engine.Log.Info("add this witness:%s", witnessOne.Addr.B58String())
				rewardWitness = append(rewardWitness, witnessOne)
				preBlockHash = &witnessOne.Block.PreBlockID
			} else {
				// engine.Log.Info("not add this witness:%s", witnessOne.Addr.B58String())
				// engine.Log.Info("not add this witness:%s check block id:%s preBlockhash:%s", witnessOne.Addr.B58String(), hex.EncodeToString(witnessOne.Block.Id), hex.EncodeToString(*preBlockHash))
			}
		}
	} else {
		//新版
		//先根据blockhash找到那个区块
		var lastWitness *Witness
		for _, one := range this.Witnesses {
			if one.Block == nil {
				continue
			}
			if bytes.Equal(*preBlockHash, one.Block.Id) {
				lastWitness = one
				rewardWitness = append(rewardWitness, one)
				preBlockHash = &one.Block.PreBlockID
				// engine.Log.Info("add this witness:%s", one.Addr.B58String())
				break
			}
		}
		if lastWitness != nil {
			for lastWitness = lastWitness.PreWitness; lastWitness != nil && lastWitness.WitnessBackupGroup == this; lastWitness = lastWitness.PreWitness {
				if lastWitness.Block == nil {
					// engine.Log.Info("this witness block is nil!")
					continue
				}
				if bytes.Equal(*preBlockHash, lastWitness.Block.Id) {
					// engine.Log.Info("add this witness:%s", lastWitness.Addr.B58String())
					rewardWitness = append(rewardWitness, lastWitness)
					preBlockHash = &lastWitness.Block.PreBlockID
				} else {
					// engine.Log.Info("not add this witness:%s check block id:%s preBlockhash:%s", lastWitness.Addr.B58String(), hex.EncodeToString(lastWitness.Block.Id), hex.EncodeToString(*preBlockHash))
				}
			}
		}
	}
	return &rewardWitness
}

/*
	构建本组中的见证人出块奖励
	按股权分配
	只有见证人方式出块才统计
	组人数乘以每块奖励，再分配给实际出块的人
*/
// func (this *WitnessBackupGroup) BuildRewardVouts(blocks []Block) []*Vout {
// 	vouts := make([]*Vout, 0)

// 	witneses := make([]*crypto.AddressCoin, 0) //已经出块的见证人
// 	// allWitness := make([]*crypto.AddressCoin, 0) //所有见证人，包括已经出块的，和未选上的候选见证人
// 	allCommiunty := make([]*VoteScore, 0) //保存所有社区节点地址及投票数量

// 	//统计本组股权和交易手续费
// 	allVotePos := uint64(0) //所有投票者票数总和
// 	allGas := uint64(0)     //计算交易手续费
// 	allReward := uint64(0)  //本组奖励数量

// 	//计算出块奖励总和
// 	for _, one := range this.Witnesses {
// 		// allWitness = append(allWitness, one.Addr)
// 		//统计所有社区节点投票
// 		for _, v := range one.CommunityVotes {
// 			v.Scores = 0
// 			allCommiunty = append(allCommiunty, v)
// 			allVotePos = allVotePos + v.Vote
// 			// engine.Log.Info("统计社区节点投票 %s %s %d", v.Addr.B58String(), v.Witness.B58String(), v.Vote)
// 		}

// 		//是否在未确认的区块中
// 		isUnconfirmed := false
// 		//判断是否在正在出块的见证人组里面
// 		nowWitnessGroup := GetLongChain().witnessChain.witnessGroup
// 		if nowWitnessGroup.Height == one.Group.Height {
// 			for _, oneBlock := range blocks {
// 				//高度相同，见证人地址相同
// 				if oneBlock.Group.Height == one.Group.Height && bytes.Equal(*one.Addr, *oneBlock.witness.Addr) {
// 					isUnconfirmed = true
// 					break
// 				}
// 			}
// 			if !isUnconfirmed {
// 				//在未确认的见证人组中，但是没有找到这个见证人出块，则不奖励
// 				continue
// 			}
// 		}

// 		if !isUnconfirmed {

// 			//只计算已经出块的见证人奖励
// 			if one.Block == nil {
// 				continue
// 			}
// 			//不能只简单通过 one.Block == nil 来判断未出块
// 			if one.Block.Group == nil {
// 				continue
// 			}
// 			//还要判断是否在已经确认的组里面
// 			ok, group := one.Group.CheckBlockGroup()
// 			if !ok {
// 				continue
// 			}
// 			//判断这个见证人出的块是否在已经确认的组里面
// 			if one.Block.Group != group {
// 				continue
// 			}

// 		}
// 		//这个见证人的出块已经得到确认
// 		witneses = append(witneses, one.Addr)

// 		//计算交易手续费
// 		// engine.Log.Info("构建本组中的见证人出块奖励 BuildRewardVouts")
// 		_, txs, _ := one.Block.LoadTxs()
// 		for _, one := range *txs {
// 			allGas = allGas + one.GetGas()
// 			// engine.Log.Info("统计区块gas %d, %d", one.GetGas(), allGas)
// 		}

// 		//按照发行总量及减半周期计算出块奖励
// 		oneReward := config.ClacRewardForBlockHeight(one.Block.Height)
// 		allReward = allReward + oneReward
// 		// engine.Log.Info("统计这个区块奖励 %d, %d", oneReward, allReward)

// 	}

// 	//区块奖励
// 	allReward = allReward + allGas

// 	// engine.Log.Info("block reward allReward:%d allVotePos:%d", allReward, allVotePos)

// 	// 10%              给资源节点
// 	// 90% * 10% * 20%  平均分给出块的见证人。
// 	// 90% * 10% * 80%  平均分给所有见证人，包括候选见证人。
// 	// 90% * 90%        给所有社区节点，由社区节点按投票数量分配。

// 	//--------------分给资源节点---------------
// 	//检查资源节点是否存在
// 	resourcesReward := uint64(0)
// 	nameinfo := name.FindNameToNet(config.Name_store)
// 	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
// 		//给资源节点。
// 		resourcesReward = new(big.Int).Div(big.NewInt(int64(allReward)), big.NewInt(int64(10))).Uint64()

// 		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]
// 		vout := Vout{
// 			Value:   resourcesReward,
// 			Address: addrCoin,
// 		}
// 		vouts = append(vouts, &vout)
// 		allReward = allReward - resourcesReward
// 		// engine.Log.Info("resource node reward %s %d", vout.Address.B58String(), vout.Value)
// 	}
// 	// engine.Log.Info("reward shenxia %d", allReward)

// 	//---------------------------------------------------
// 	//平均分给出块的见证人。
// 	temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(2)))
// 	value := new(big.Int).Div(temp, big.NewInt(int64(100)))
// 	witnessReward := value.Uint64()

// 	//平均分给出块的见证人
// 	// engine.Log.Info("Distribute equally to the witnesses of the block %d", witnessReward)

// 	//按股权分给所有见证人，包括候选见证人。
// 	temp = new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(8)))
// 	value = new(big.Int).Div(temp, big.NewInt(int64(100)))
// 	allWitnessReward := value.Uint64()

// 	//平均分给所有见证人，包括候选见证人
// 	// engine.Log.Info("Evenly distributed to all witnesses, including candidate witnesses %d", allWitnessReward)

// 	//给所有社区节点，由社区节点按投票数量分配。
// 	allCommiuntyReward := allReward - allWitnessReward - witnessReward

// 	//给所有社区节点分
// 	// engine.Log.Info("To all community nodes %d", allCommiuntyReward)

// 	//---------------------------------------------------
// 	//给出块的见证人平均分。
// 	use := uint64(0)
// 	temp = new(big.Int).Mul(big.NewInt(int64(witnessReward)), big.NewInt(int64(1)))
// 	value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses))))
// 	oneReward := value.Uint64()
// 	for _, one := range this.Witnesses {
// 		//给所有已经出块的见证人平均分
// 		// engine.Log.Info("Give the witness average score of the block %s %d", one.Addr.B58String(), oneReward)
// 		use = use + oneReward
// 		vout := Vout{
// 			Value:   oneReward,
// 			Address: *one.Addr,
// 		}
// 		vouts = append(vouts, &vout)
// 	}
// 	//平均数不能被整除时候，剩下的给最后一个输出奖励
// 	if len(vouts) > 0 {
// 		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (witnessReward - use)
// 	}

// 	// engine.Log.Info("开始给见证人分配 1111111111")
// 	//---------------------------------------------------
// 	//给所有候选见证人按投票股权分配。
// 	use = uint64(0)
// 	if allVotePos <= 0 {
// 		temp = new(big.Int).Mul(big.NewInt(int64(allWitnessReward)), big.NewInt(1))
// 		value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses)+len(this.WitnessBackup))))
// 		oneReward = value.Uint64()
// 		for _, one := range append(this.Witnesses, this.WitnessBackup...) {
// 			//给所有候选见证人平均分
// 			// engine.Log.Info("Average all candidate witnesses %s %d", one.Addr.B58String(), oneReward)
// 			use = use + oneReward
// 			vout := Vout{
// 				Value:   oneReward,
// 				Address: *one.Addr,
// 			}
// 			vouts = append(vouts, &vout)
// 		}
// 	} else {
// 		for _, one := range append(this.Witnesses, this.WitnessBackup...) {
// 			temp = new(big.Int).Mul(big.NewInt(int64(allWitnessReward)), big.NewInt(int64(one.VoteNum)))
// 			value = new(big.Int).Div(temp, big.NewInt(int64(allVotePos)))
// 			oneReward = value.Uint64()
// 			//给所有候选见证人平均分
// 			// engine.Log.Info("Average all candidate witnesses %s %s", one.Addr.B58String(), oneReward)
// 			use = use + oneReward
// 			vout := Vout{
// 				Value:   oneReward,
// 				Address: *one.Addr,
// 			}
// 			vouts = append(vouts, &vout)
// 		}
// 	}
// 	//平均数不能被整除时候，剩下的给最后一个输出奖励
// 	if len(vouts) > 0 {
// 		// engine.Log.Info("加余数 %d %d", use, allWitnessReward-use)
// 		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (allWitnessReward - use)
// 	}

// 	// engine.Log.Info("开始给见证人分配 2222222222222")

// 	//---------------------------------------------------
// 	//给所有社区节点，由社区节点按投票数量分配。
// 	use = uint64(0)
// 	//如果所有投票数量为0，则将这部分收益分给所有候选见证人。
// 	if allVotePos <= 0 {
// 		// engine.Log.Info("开始给见证人分配 33333333333333")
// 		//给所有候选见证人
// 		temp = new(big.Int).Mul(big.NewInt(int64(allCommiuntyReward)), big.NewInt(int64(1)))
// 		value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses))))
// 		oneReward = value.Uint64()
// 		for i, _ := range this.Witnesses {
// 			use = use + oneReward
// 			vout := Vout{
// 				Value:   oneReward,
// 				Address: *this.Witnesses[i].Addr,
// 			}
// 			vouts = append(vouts, &vout)
// 			//开始给见证人分配
// 			// engine.Log.Info("Start to assign witness %s %d", this.Witnesses[i].Addr.B58String(), oneReward)
// 		}
// 	} else {
// 		// engine.Log.Info("开始给见证人分配 4444444444444")
// 		for i, one := range allCommiunty {
// 			//给所有社区节点参数
// 			// engine.Log.Info("Give all community node parameters %d %d %d", allCommiuntyReward, one.Vote, allVotePos)
// 			if one.Vote == 0 {
// 				continue
// 			}
// 			temp = new(big.Int).Mul(big.NewInt(int64(allCommiuntyReward)), big.NewInt(int64(one.Vote)))
// 			value = new(big.Int).Div(temp, big.NewInt(int64(allVotePos)))
// 			oneReward = value.Uint64()
// 			//给所有社区节点
// 			// engine.Log.Info("To all community nodes %s %d", allCommiunty[i].Addr.B58String(), oneReward)
// 			use = use + oneReward
// 			vout := Vout{
// 				Value:   oneReward,
// 				Address: *allCommiunty[i].Addr,
// 			}
// 			vouts = append(vouts, &vout)
// 		}
// 	}
// 	//平均数不能被整除时候，剩下的给最后一个输出奖励
// 	if len(vouts) > 0 {
// 		// engine.Log.Info("加余数 %d %d", use, allCommiuntyReward-use)
// 		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (allCommiuntyReward - use)
// 	}

// 	return MergeVouts(&vouts)
// }

/*
	构建本组中的见证人出块奖励
	按股权分配
	只有见证人方式出块才统计
	组人数乘以每块奖励，再分配给实际出块的人
	@height    uint64    当前区块高度
*/
func (this *WitnessBackupGroup) BuildRewardVouts(blocks []Block, height uint64, blockhash *[]byte, preBlock *Block) []*Vout {

	// engine.Log.Info("BuildRewardVouts")
	vouts := make([]*Vout, 0)

	witneses := make([]*crypto.AddressCoin, 0) //已经出块的见证人
	// allWitness := make([]*crypto.AddressCoin, 0) //所有见证人，包括已经出块的，和未选上的候选见证人
	allCommiunty := make([]*VoteScore, 0) //保存所有社区节点地址及投票数量

	//统计本组股权和交易手续费
	allVotePos := uint64(0) //所有投票者票数总和
	allGas := uint64(0)     //计算交易手续费
	allReward := uint64(0)  //本组奖励数量

	if height >= config.Reward_witness_height {
		witnessAll := this.CountRewardWitness(&preBlock.Id, height)
		for _, one := range *witnessAll {

			//按照发行总量及减半周期计算出块奖励
			oneReward := config.ClacRewardForBlockHeight(one.Block.Height)
			// engine.Log.Info("reward block heihgt:%d reward:%d", one.Block.Height, oneReward)
			allReward = allReward + oneReward
			//计算交易手续费
			// engine.Log.Info("构建本组中的见证人出块奖励 BuildRewardVouts")
			_, txs, _ := one.Block.LoadTxs()
			for _, one := range *txs {
				allGas = allGas + one.GetGas()
				// engine.Log.Info("统计区块gas %d, %d", one.GetGas(), allGas)
			}
			//这个见证人的出块已经得到确认
			witneses = append(witneses, one.Addr)
		}

		for _, one := range this.Witnesses {
			//统计所有社区节点投票
			for _, v := range one.CommunityVotes {
				v.Scores = 0
				allCommiunty = append(allCommiunty, v)
				allVotePos = allVotePos + v.Vote
				// engine.Log.Info("统计社区节点投票 %s %s %d", v.Addr.B58String(), v.Witness.B58String(), v.Vote)
			}
		}

	} else {

		//计算出块奖励总和
		for _, one := range this.Witnesses {
			// allWitness = append(allWitness, one.Addr)
			//统计所有社区节点投票
			for _, v := range one.CommunityVotes {
				v.Scores = 0
				allCommiunty = append(allCommiunty, v)
				allVotePos = allVotePos + v.Vote
				// if height == 781928 {
				// 	engine.Log.Info("统计社区节点投票 %s %s %d", v.Addr.B58String(), v.Witness.B58String(), v.Vote)
				// }
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
				ok, group := one.Group.CheckBlockGroup(blockhash)
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
			// engine.Log.Info("构建本组中的见证人出块奖励 BuildRewardVouts")
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
	}

	// engine.Log.Info("reward block:%d gas:%d", allReward, allGas)
	//区块奖励
	allReward = allReward + allGas

	// engine.Log.Info("block reward allReward:%d allVotePos:%d", allReward, allVotePos)

	// 基金会6%
	// 投资人8%
	// 团队14%
	// 资源奖励39%
	// 投票奖励33%

	//--------------分给基金会6%---------------
	foundationReward := uint64(0)
	//检查基金会节点是否存在
	nameinfo := name.FindNameToNet(config.Name_Foundation)
	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
		//
		temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(6)))
		value := new(big.Int).Div(temp, big.NewInt(int64(100)))
		foundationReward = value.Uint64()

		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]

		voutsOne := LinearRelease0Day(addrCoin, foundationReward, height)
		vouts = append(vouts, voutsOne...)
		// if height == 781928 {
		// 	engine.Log.Info("allReward:%d foundationReward:%d", allReward, foundationReward)
		// 	for _, one := range voutsOne {
		// 		engine.Log.Info("%s %d", one.Address.B58String(), one.Value)
		// 	}
		// }

		// allReward = allReward - resourcesReward
		// engine.Log.Info("resource node reward %s %d", vout.Address.B58String(), vout.Value)
	}
	// engine.Log.Info("分给基金会0.06 %d", foundationReward)

	//--------------分给投资人8%---------------
	investorReward := uint64(0)
	//检查投资人节点是否存在
	nameinfo = name.FindNameToNet(config.Name_investor)
	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
		//
		temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(8)))
		value := new(big.Int).Div(temp, big.NewInt(int64(100)))
		investorReward = value.Uint64()

		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]

		voutsOne := LinearRelease0Day(addrCoin, investorReward, height)
		vouts = append(vouts, voutsOne...)

		// if height == 781928 {
		// 	engine.Log.Info("allReward:%d investorReward:%d", allReward, investorReward)
		// 	for _, one := range voutsOne {
		// 		engine.Log.Info("%s %d", one.Address.B58String(), one.Value)
		// 	}
		// }

		// allReward = allReward - resourcesReward
		// engine.Log.Info("resource node reward %s %d", vout.Address.B58String(), vout.Value)
	}
	// engine.Log.Info("分给投资人0.08 %d", investorReward)

	//--------------分给团队14%---------------
	teamReward := uint64(0)
	//检查团队节点是否存在
	nameinfo = name.FindNameToNet(config.Name_team)
	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
		//
		temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(14)))
		value := new(big.Int).Div(temp, big.NewInt(int64(100)))
		teamReward = value.Uint64()

		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]

		voutsOne := LinearRelease0Day(addrCoin, teamReward, height)
		vouts = append(vouts, voutsOne...)

		// if height == 781928 {
		// 	engine.Log.Info("allReward:%d teamReward:%d", allReward, teamReward)
		// 	for _, one := range voutsOne {
		// 		engine.Log.Info("%s %d", one.Address.B58String(), one.Value)
		// 	}
		// }

		// allReward = allReward - resourcesReward
		// engine.Log.Info("resource node reward %s %d", vout.Address.B58String(), vout.Value)
	}
	// engine.Log.Info("分给团队0.14 %d", teamReward)

	//--------------分给存储节点39%---------------
	storeReward := uint64(0)
	//检查存储节点是否存在
	nameinfo = name.FindNameToNet(config.Name_store)
	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
		//
		temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(39)))
		value := new(big.Int).Div(temp, big.NewInt(int64(100)))
		storeReward = value.Uint64()

		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]

		vout := &Vout{
			Value:   storeReward, //输出金额 = 实际金额 * 100000000
			Address: addrCoin,    //钱包地址
		}

		// voutsOne := LinearRelease180Day(addrCoin, storeReward, height)
		vouts = append(vouts, vout)

		// if height == 781928 {
		// 	engine.Log.Info("allReward:%d", allReward)
		// 	engine.Log.Info("%s %d", vout.Address.B58String(), vout.Value)

		// }

		// allReward = allReward - resourcesReward
		// engine.Log.Info("resource node reward %s %d", vout.Address.B58String(), vout.Value)
	}
	// engine.Log.Info("分给存储节点0.39 %d", storeReward)

	//---------------------------------------------------
	//0.66%  99个见证人均分
	temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(66)))
	value := new(big.Int).Div(temp, big.NewInt(int64(10000)))
	witnessReward99 := value.Uint64()
	// engine.Log.Info("0.0066  99个见证人均分 %d", witnessReward99)

	//2.64% 按股权分给31个见证人。
	temp = new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(264)))
	value = new(big.Int).Div(temp, big.NewInt(int64(1000)))
	witnessReward31 := value.Uint64()
	// engine.Log.Info("0.0264 按股权分给31个见证人。 %d", witnessReward31)

	//29.7% 社区节点加权分。
	temp = new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(297)))
	value = new(big.Int).Div(temp, big.NewInt(int64(1000)))
	communityReward := value.Uint64()
	// engine.Log.Info("0.297 社区节点加权分。 %d", communityReward)

	//有多余的奖励，分配给最后一项
	surplus := allReward - (foundationReward + investorReward + teamReward + storeReward + witnessReward99 + witnessReward31 + communityReward)
	if surplus > 0 {
		witnessReward31 = witnessReward31 + surplus
	}

	//给所有社区节点分
	// engine.Log.Info("To all community nodes %d", allCommiuntyReward)

	use := uint64(0)
	oneReward := uint64(0)

	//---------------------------------------------------
	//给99个见证人平均分。
	{
		if height < config.Mining_witness_average_height {
			use = uint64(0)
			temp = new(big.Int).Mul(big.NewInt(int64(witnessReward99)), big.NewInt(int64(1)))
			value = new(big.Int).Div(temp, big.NewInt(int64(config.Witness_backup_reward_max)))
			oneReward = value.Uint64()
			for i, one := range append(this.Witnesses, this.WitnessBackup...) {
				if i >= config.Witness_backup_reward_max {
					break
				}
				//给所有已经出块的见证人平均分
				// engine.Log.Info("Give the witness average score of the block %s %d", one.Addr.B58String(), oneReward)
				use = use + oneReward
				voutsOne := LinearRelease0Day(*one.Addr, oneReward, height)
				vouts = append(vouts, voutsOne...)

				// if height == 781928 {
				// 	engine.Log.Info("witnessReward99:%d oneReward:%d", witnessReward99, oneReward)
				// 	for _, one := range voutsOne {
				// 		engine.Log.Info("%s %d", one.Address.B58String(), one.Value)
				// 	}
				// }
			}
			//平均数不能被整除时候，剩下的给最后一个输出奖励
			if len(vouts) > 0 {
				vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (witnessReward99 - use)
			}
		} else {
			//超过这一高度，按新的规则计算奖励
			/*
				1.31个以内，按现有见证人数量平均分配（假如只有5个见证人，则5个见证人平均分）。
				2.31-99个，均分部分按现有见证人数量均分。
				3.大于99个见证人，均分部分给前99个见证人均分。排名99之后的见证人没有奖励。
			*/
			averageBackupTotalMax := config.Witness_backup_reward_max - config.Witness_backup_max
			witnessBackupTotal := len(this.WitnessBackup)
			averageWitness := this.Witnesses
			if witnessBackupTotal >= averageBackupTotalMax {
				averageWitness = append(averageWitness, this.WitnessBackup[:averageBackupTotalMax]...)
			} else if witnessBackupTotal > 0 && witnessBackupTotal < averageBackupTotalMax {
				averageWitness = append(averageWitness, this.WitnessBackup...)
			}

			// engine.Log.Info("%d averageWitness number:%d blockNumber:%d height:%d", witnessReward99, len(averageWitness), len(blocks), height)

			use = uint64(0)
			temp = new(big.Int).Mul(big.NewInt(int64(witnessReward99)), big.NewInt(int64(1)))
			value = new(big.Int).Div(temp, big.NewInt(int64(len(averageWitness))))
			oneReward = value.Uint64()
			for _, one := range averageWitness {
				// engine.Log.Info("averageWitness one:%s %d", one.Addr.B58String(), oneReward)
				// if i >= config.Witness_backup_reward_max {
				// 	break
				// }
				//给所有已经出块的见证人平均分
				// engine.Log.Info("Give the witness average score of the block %s %d", one.Addr.B58String(), oneReward)
				use = use + oneReward
				voutsOne := LinearRelease0Day(*one.Addr, oneReward, height)
				vouts = append(vouts, voutsOne...)

				// if height == 781928 {
				// 	engine.Log.Info("witnessReward99:%d oneReward:%d", witnessReward99, oneReward)
				// 	for _, one := range voutsOne {
				// 		engine.Log.Info("%s %d", one.Address.B58String(), one.Value)
				// 	}
				// }
			}
			//平均数不能被整除时候，剩下的给最后一个输出奖励
			if len(vouts) > 0 {
				vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (witnessReward99 - use)
			}
		}
	}

	// engine.Log.Info("开始给见证人分配 1111111111")
	//---------------------------------------------------
	//给31个证人按投票股权分配。
	{
		use = uint64(0)
		if allVotePos <= 0 {
			//当投票为0时，平均分配
			temp = new(big.Int).Mul(big.NewInt(int64(witnessReward31)), big.NewInt(1))
			value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses))))
			oneReward = value.Uint64()
			for _, one := range this.Witnesses {
				//给所有候选见证人平均分
				// engine.Log.Info("Average all candidate witnesses %s %d", one.Addr.B58String(), oneReward)
				use = use + oneReward
				voutsOne := LinearRelease0Day(*one.Addr, oneReward, height)
				vouts = append(vouts, voutsOne...)
			}
		} else {
			for _, one := range this.Witnesses {
				temp = new(big.Int).Mul(big.NewInt(int64(witnessReward31)), big.NewInt(int64(one.VoteNum)))
				value = new(big.Int).Div(temp, big.NewInt(int64(allVotePos)))
				oneReward = value.Uint64()
				//给所有候选见证人平均分
				// engine.Log.Info("Average all candidate witnesses %s %s", one.Addr.B58String(), oneReward)
				use = use + oneReward
				voutsOne := LinearRelease0Day(*one.Addr, oneReward, height)
				vouts = append(vouts, voutsOne...)
			}
		}
		//平均数不能被整除时候，剩下的给最后一个输出奖励
		if len(vouts) > 0 {
			// engine.Log.Info("加余数 %d %d", use, allWitnessReward-use)
			vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (witnessReward31 - use)
		}
	}
	// engine.Log.Info("开始给见证人分配 2222222222222")

	//---------------------------------------------------
	//给所有社区节点，由社区节点按投票数量分配。
	{

		use = uint64(0)
		//如果所有投票数量为0，则将这部分收益分给所有候选见证人。
		if allVotePos <= 0 {
			// engine.Log.Info("开始给见证人分配 33333333333333")
			//给所有候选见证人
			temp = new(big.Int).Mul(big.NewInt(int64(communityReward)), big.NewInt(int64(1)))
			value = new(big.Int).Div(temp, big.NewInt(int64(len(this.Witnesses))))
			oneReward = value.Uint64()
			for i, _ := range this.Witnesses {
				use = use + oneReward
				vout := Vout{
					Value:   oneReward,
					Address: *this.Witnesses[i].Addr,
				}
				vouts = append(vouts, &vout)

				// if height == 781928 {
				// 	engine.Log.Info("communityReward:%d this.Witnesses:%d oneReward:%d", communityReward, this.Witnesses, oneReward)

				// 	engine.Log.Info("%s %d", vout.Address.B58String(), vout.Value)

				// }
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
				temp = new(big.Int).Mul(big.NewInt(int64(communityReward)), big.NewInt(int64(one.Vote)))
				value = new(big.Int).Div(temp, big.NewInt(int64(allVotePos)))
				oneReward = value.Uint64()
				//给所有社区节点
				// engine.Log.Info("To all community nodes %s %d", allCommiunty[i].Addr.B58String(), oneReward)
				use = use + oneReward
				vout := Vout{
					Value:   oneReward,
					Address: *allCommiunty[i].Addr,
				}
				vouts = append(vouts, &vout)

				// if height == 781928 {
				// 	engine.Log.Info("communityReward:%d one.Vote:%d allVotePos:%d oneReward:%d", communityReward, one.Vote, allVotePos, oneReward)

				// 	engine.Log.Info("%s %d", vout.Address.B58String(), vout.Value)

				// }
			}
		}
		//平均数不能被整除时候，剩下的给最后一个输出奖励
		if len(vouts) > 0 {
			// engine.Log.Info("加余数 %d %d", use, allCommiuntyReward-use)
			vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (communityReward - use)
		}
		// if height == 781928 {
		// 	engine.Log.Info("lastvout value:%d communityReward:%d use:%d", vouts[len(vouts)-1].Value, communityReward, use)
		// 	lastVout := vouts[len(vouts)-1]
		// 	engine.Log.Info("%s %d", lastVout.Address.B58String(), lastVout.Value)

		// }
	}

	vouts = CleanZeroVouts(&vouts)
	// if height == 781928 {
	// 	for _, one := range vouts {
	// 		if bytes.Equal(one.Address, config.SpecialAddrs) {
	// 			engine.Log.Info("%s %d", one.Address.B58String(), one.Value)
	// 			break
	// 		}
	// 	}
	// }
	vouts = MergeVouts(&vouts)
	// if height == 781928 {
	// 	for _, one := range vouts {
	// 		if bytes.Equal(one.Address, config.SpecialAddrs) {
	// 			engine.Log.Info("%s %d", one.Address.B58String(), one.Value)
	// 			break
	// 		}
	// 	}
	// }

	return vouts
}

/*
	180天线性释放
	25%及时到账
	75%按180天线性释放
*/
func LinearRelease180Day(addr crypto.AddressCoin, total uint64, height uint64) []*Vout {
	//TODO 处理好不能被180整除的情况
	vouts := make([]*Vout, 0)
	//25%直接到账
	first25 := new(big.Int).Div(big.NewInt(int64(total)), big.NewInt(int64(4)))
	//剩下的75%
	surplus := new(big.Int).Sub(big.NewInt(int64(total)), first25)

	vout := &Vout{
		Value:   first25.Uint64(), //输出金额 = 实际金额 * 100000000
		Address: addr,             //钱包地址
		// FrozenHeight: height + uint64(i*intervalHeight), //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
	}
	vouts = append(vouts, vout)

	dayOne := new(big.Int).Div(surplus, big.NewInt(int64(18))).Uint64()
	intervalHeight := 60 * 60 * 24 * 10 / 10

	totalUse := uint64(0)
	for i := 0; i < 18; i++ {
		vout := &Vout{
			Value:        dayOne,                                //输出金额 = 实际金额 * 100000000
			Address:      addr,                                  //钱包地址
			FrozenHeight: height + uint64((i+1)*intervalHeight), //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
		}
		vouts = append(vouts, vout)
		totalUse = totalUse + dayOne
	}
	//平均数不能被整除时候，剩下的给最后一个输出奖励
	if totalUse < surplus.Uint64() {
		// engine.Log.Info("加余数 %d %d", use, allCommiuntyReward-use)
		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (surplus.Uint64() - totalUse)
	}
	return vouts
}

/*
	不冻结，完全到账
*/
func LinearRelease0Day(addr crypto.AddressCoin, total uint64, height uint64) []*Vout {
	vouts := make([]*Vout, 0)
	vout := &Vout{
		Value:   total, //输出金额 = 实际金额 * 100000000
		Address: addr,  //钱包地址
	}
	vouts = append(vouts, vout)
	return vouts
}

/*
	构建本组中的见证人出块奖励
	按股权分配
	只有见证人方式出块才统计
	组人数乘以每块奖励，再分配给实际出块的人
*/
func (this *WitnessBackupGroup) CountRewardToWitnessGroup(blockHeight uint64, blocks []Block, preBlock *Block) *Tx_reward {

	//构建区块奖励输出
	// engine.Log.Info("build reward")
	// start := time.Now()
	vouts := this.BuildRewardVouts(blocks, blockHeight, nil, preBlock)
	// engine.Log.Info("构建输出消耗时间 %s", time.Now().Sub(start))
	// for _, one := range vouts {
	// 	engine.Log.Info("%+v", one)
	// }

	//构建输入
	baseCoinAddr := keystore.GetCoinbase()
	// puk, ok := keystore.GetPukByAddr(baseCoinAddr)
	// if !ok {
	// 	return nil
	// }
	vins := make([]*Vin, 0)
	vin := Vin{
		Puk:  baseCoinAddr.Puk, //公钥
		Sign: nil,              //对上一个交易签名，是对整个交易签名（若只对输出签名，当地址和金额一样时，签名输出相同）。
	}
	vins = append(vins, &vin)

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

			_, prk, err := keystore.GetKeyByPuk(one.Puk, config.Wallet_keystore_default_pwd)
			if err != nil {
				return nil
			}
			// engine.Log.Info("查找公钥key 耗时 %d %s", i, time.Now().Sub(startTime))
			sign := txReward.GetSign(&prk, one.Txid, one.Vout, uint64(i))
			txReward.Vin[i].Sign = *sign

		}

		txReward.BuildHash()
		if txReward.CheckHashExist() {
			txReward = nil
			continue
		} else {
			break
		}
	}
	// engine.Log.Info("构建见证人奖励消耗时间 %s", time.Now().Sub(start))
	return txReward
}
