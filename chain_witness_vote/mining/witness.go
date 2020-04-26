package mining

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining/name"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
	"time"
)

/*
	见证人链
	投票竞选出来的见证人加入链中，但是没有加入组
	当交了押金后，被分配到见证人组
*/
type WitnessChain struct {
	chain           *Chain         //所属链
	witnessBackup   *WitnessBackup //
	witnessGroup    *WitnessGroup  //当前正在出块的见证人组
	witnessNotGroup []*Witness     //未分配组的见证人列表
	// firstWitnessNotGroup *Witness       //首个未分配组的见证人引用
	// lastWitnessNotGroup  *Witness       //最后一个未分配组的见证人引用
}

func NewWitnessChain(wb *WitnessBackup, chain *Chain) *WitnessChain {
	return &WitnessChain{
		chain:         chain,
		witnessBackup: wb,
	}
}

/*
	见证人组
*/
type WitnessGroup struct {
	Task       bool          //是否已经定时出块
	PreGroup   *WitnessGroup //上一个组
	NextGroup  *WitnessGroup //下一个组
	Height     uint64        //见证人组高度
	Witness    []*Witness    //本组见证人列表
	BlockGroup *Group        //构建出来的合法区块组，为空则是没有合法组
	tag        bool          //备用见证人队列标记，一次从备用见证人评选出来的备用见证人为一个队列标记，需要保证备用见证人组中有两个队列
	// check      bool          //这个见证人组是否多半人出块，多半人出块则合法有效
}

/*
	见证人
*/
type Witness struct {
	Group           *WitnessGroup       //
	PreWitness      *Witness            //上一个见证人
	NextWitness     *Witness            //下一个见证人
	Addr            *crypto.AddressCoin //见证人地址
	Puk             []byte              //见证人公钥
	Block           *Block              //见证人生产的块
	Score           uint64              //见证人自己的押金
	CommunityVotes  []*VoteScore        //社区节点投票
	Votes           []*VoteScore        //轻节点投票和押金
	VoteNum         uint64              //投票数量
	StopMining      chan bool           `json:"-"` //停止出块命令
	BlockHeight     uint64              //预计出块高度
	CreateBlockTime int64               //预计出块时间
	//	DepositId   []byte           //押金交易id
	//	ElectionMap *sync.Map        //本块中交易押金投票数 key:string=投票者地址;value:*BallotTicket=选票;
	WitnessBackupGroup *WitnessBackupGroup //候选见证人组
	// GroupStart         bool                //一个见证人组的首个见证人，首个见证人要分配奖励
	CheckIsMining bool       //检查是否已经验证出块标记，用于多次未出块的见证人，踢出列表
	syncBlockOnce *sync.Once //定时同步区块，只执行一次
}

/*
	将见证人组中出的块构建为区块组
*/
func (this *WitnessChain) BuildBlockGroup(bhvo *BlockHeadVO) {
	engine.Log.Info("Height of block group constructed this time %d", bhvo.BH.GroupHeight)

	if bhvo.BH.GroupHeight == config.Mining_group_start_height {
		this.witnessGroup.BuildGroup()
		this.BuildWitnessGroup(false)
		// fmt.Println("下一个见证人是否为空", this.witnessGroup.NextGroup)
		this.witnessGroup = this.witnessGroup.NextGroup
		this.BuildWitnessGroup(false)
		return
	}

	//先判断时间，是否应该构建前面的区块组
	witness, _ := this.FindWitnessForBlock(bhvo)
	if witness == nil {
		return
	}

	//是新的组出块，则构建前面的组
	for i := this.witnessGroup.Height; i < bhvo.BH.GroupHeight; i++ {

		// engine.Log.Info("本次构建的区块组高度   222222222222222 %d", i)

		wg := this.witnessGroup.BuildGroup()

		//找到上一组到本组的见证人组，开始查找没有出块的见证人
		if wg != nil {
			for tempGroup := wg; tempGroup != nil && tempGroup.Height < i; tempGroup = tempGroup.NextGroup {
				for _, one := range wg.Witness {
					if !one.CheckIsMining {
						if one.Block == nil {
							this.chain.witnessBackup.AddBlackList(*one.Addr)
						} else {
							this.chain.witnessBackup.SubBlackList(*one.Addr)
						}
						one.CheckIsMining = true
					}
				}
			}
		}

		this.chain.CountBlock()

		if bhvo.BH.GroupHeight != config.Mining_group_start_height+1 {
			// engine.Log.Info("本次构建的区块组高度   3333333333333333 %d", i)
			this.witnessGroup = this.witnessGroup.NextGroup
		}
		//构建一个新的见证人组
		this.BuildWitnessGroup(false)

	}

}

/*
	则构建这一组见证人的所有出块，保存关系
*/
func (this *WitnessGroup) BuildGroup() *WitnessGroup {
	// fmt.Println("统计并修改组", this.Height)

	//已经构建了则退出，避免重复构建浪费计算资源
	if this.BlockGroup != nil {
		return nil
	}

	//判断这个组是否多人出块
	ok, group := this.CheckBlockGroup()
	if !ok {
		// fmt.Println("本组出块人数太少，不合格")
		return nil
	}

	// engine.Log.Info("本组出块数量 %d", len(group.Blocks))

	this.BlockGroup = group

	// fmt.Println("BuildGroup  1111111111111")

	//找到上一组
	beforeGroup := this.PreGroup
	for beforeGroup = this.PreGroup; beforeGroup != nil; beforeGroup = beforeGroup.PreGroup {
		ok, _ = beforeGroup.CheckBlockGroup()
		if ok {
			break
		}
	}

	if beforeGroup == nil {
		return nil
	}
	//修改组引用
	beforeGroup.BlockGroup.NextGroup = this.BlockGroup
	this.BlockGroup.PreGroup = beforeGroup.BlockGroup

	//修改块引用
	beforeBlock := beforeGroup.BlockGroup.Blocks[len(beforeGroup.BlockGroup.Blocks)-1]
	//这里决定了保存到数据库中的链是已经确认的链，没有保存分叉
	beforeBlock.NextBlock = this.BlockGroup.Blocks[0]
	// beforeBlock.FlashNextblockhash()
	return beforeGroup
}

/*
	判断是否多半人出块，只判断，不作修改和保存
*/
func (this *WitnessGroup) CheckBlockGroup() (bool, *Group) {
	// fmt.Println("统计本组合法的区块组", this.Height)

	//已经统计过了，就不需要再统计了，直接返回
	if this.BlockGroup != nil {
		// fmt.Println("统计过了，就不需要再统计了", len(this.BlockGroup.Blocks))
		return true, this.BlockGroup
	}

	group := this.SelectionChain()
	//出块数量为0
	if group == nil {
		// fmt.Println("评选出来的组为空")
		return false, nil
	}
	totalWitness := len(this.Witness)
	totalHave := len(group.Blocks)
	// //包含只有两个见证人的情况，一人出块既是成功
	// if (totalHave * 2) < totalWitness {
	//包含只有两个见证人的情况，一人出块既是成功
	if (totalHave * 2) <= totalWitness {
		// fmt.Println("出块人太少 不合格")
		return false, group
	}
	// fmt.Println("出块人多数 合格")
	return true, group
}

/*
	选出这个见证人组中出块最多块的链
*/
func (this *WitnessGroup) SelectionChain() *Group {
	this.CollationRelationship()

	// engine.Log.Info("开始评选出块最多的组 group height %d", this.Height)

	//开始评选出块最多的组
	groupMap := make(map[string]*Group) //key:string=每个组的首个快hash;value:*Group=分叉组;
	for _, one := range this.Witness {
		// fmt.Println("开始评选出块最多的组 111111111111111")
		// totalWitness++
		if one.Block == nil {
			// fmt.Println("开始评选出块最多的组 2222222222222222222222")
			continue
		}
		if one.Block.Group == nil {
			continue
		}
		// engine.Log.Info("开始评选出块最多的组 3333333333333333333 %v", one.Block.Group)
		// engine.Log.Info("开始评选出块最多的组 3333333333333333333 group:%d total:%d", this.Height, len(one.Block.Group.Blocks))
		groupMap[hex.EncodeToString(one.Block.Group.Blocks[0].Id)] = one.Block.Group
	}
	// fmt.Println("开始评选出块最多的组 444444444444444444444")
	//评选出块多的组
	var group *Group
	for _, v := range groupMap {
		// fmt.Println("开始评选出块最多的组 555555555555555555")
		if group == nil {
			// fmt.Println("开始评选出块最多的组 66666666666666666")
			group = v
			continue
		}
		// fmt.Println("开始评选出块最多的组 77777777777777777777")
		if len(v.Blocks) > len(group.Blocks) {
			// fmt.Println("开始评选出块最多的组 88888888888888888888")
			group = v
		}
		// fmt.Println("开始评选出块最多的组 99999999999999999")
	}
	// fmt.Println("开始评选出块最多的组 99999", group)
	return group
}

/*
	整理本组中已出区块之间的前后关系
*/
func (this *WitnessGroup) CollationRelationship() {
	// engine.Log.Info("整理本组中区块的前后关系 group height %d", this.Height)
	//先将这个组的块分成组
	for _, witness := range this.Witness {
		// engine.Log.Info("关系   1111111111")
		if witness.Block == nil {
			// engine.Log.Info("关系   2222222222")
			continue
		}
		// engine.Log.Info("关系   3333333333333 %d", witness.Block.Height)
		newBlock := witness.Block
		//寻找前置区块的见证人
		var beforeWitness *Witness
		for witnessTemp := witness.PreWitness; witnessTemp != nil; witnessTemp = witnessTemp.PreWitness {
			// beforeWitness = nil
			// engine.Log.Info("关系   44444444444 %d %s", witnessTemp.Group.Height, witnessTemp.Addr.B58String())
			if witnessTemp.Block == nil {
				// engine.Log.Info("关系   55555555555555")
				continue
			}
			//往前查找n个组，查不到跳出循环
			if witnessTemp.Group.Height+config.Witness_backup_group <= witness.Group.Height {
				//刚好本见证人的前面的一个见证人没有出块，则同步这个块
				for syncWitness := witness.PreWitness; syncWitness != nil &&
					witness.Group.Height == syncWitness.Group.Height+1; syncWitness =
					syncWitness.PreWitness {
					if syncWitness.Block != nil {
						continue
					}

					//					bfw := BlockForWitness{
					//						GroupHeight: syncWitness.Group.Height, //见证人组高度
					//						Addr:        *syncWitness.Addr,        //见证人地址
					//					}
					//					//开始同步
					//					go syncWitness.syncBlock(5, time.Second/5, &bfw)

				}
				break
			}

			// fmt.Println("--", newBlock)
			// engine.Log.Info("查看上一个区块 %d hash %s %s ", witnessTemp.Block.Height,
			// hex.EncodeToString(witnessTemp.Block.Id), hex.EncodeToString(newBlock.PreBlockID))
			if bytes.Equal(witnessTemp.Block.Id, newBlock.PreBlockID) {
				// engine.Log.Info("关系   66666666666666")
				// newBlock.PreBlock = witnessTemp.Block
				// if witnessTemp.Group.Height == witness.Group.Height {
				// 	fmt.Println("关系   777777777777777")
				// 	witnessTemp.Block.NextBlock = newBlock
				// }
				beforeWitness = witnessTemp
				break
			}
		}
		// engine.Log.Info("关系   7777777777777777")
		//创始区块
		if witness.Block.Height == config.Mining_block_start_height {
			// engine.Log.Info("关系   888888888888888")
			newGroup := new(Group)
			newGroup.Height = witness.Group.Height // bhvo.BH.GroupHeight
			newGroup.Blocks = []*Block{newBlock}
			// newGroup.NextGroup = make([]*Group, 0)
			newBlock.Group = newGroup
			continue
		}
		// engine.Log.Info("关系   9999999999999999")
		//没有前置区块的，直接不分成组了
		if beforeWitness == nil {
			// engine.Log.Info("关系   ----------------")
			continue
		}
		// engine.Log.Info("关系   1111  111111111111")
		//有前置区块，判断前置区块组高度是否相同
		if beforeWitness.Group.Height == witness.Group.Height {
			// engine.Log.Info("关系   1111  22222222222")
			//去重复
			have := false
			for _, one := range beforeWitness.Block.Group.Blocks {
				if newBlock.witness == one.witness {
					have = true
					break
				}
			}
			if !have {
				//组高度相同
				beforeWitness.Block.Group.Blocks = append(beforeWitness.Block.Group.Blocks, newBlock)
				newBlock.Group = beforeWitness.Block.Group
				newBlock.PreBlock = beforeWitness.Block
				beforeWitness.Block.NextBlock = newBlock
			}
		} else {
			// engine.Log.Info("关系   1111  3333333333333")
			//组高度不同
			if newBlock.Group == nil {
				// engine.Log.Info("关系   1111  444444444444")
				newBlock.Group = new(Group)
				newBlock.Group.Height = witness.Group.Height // bhvo.BH.GroupHeight
				newBlock.Group.Blocks = []*Block{newBlock}
				// newGroup.NextGroup = make([]*Group, 0)
			}
			newBlock.PreBlock = beforeWitness.Block
		}
		// beforeWitness.Block.NextBlock = append(beforeWitness.Block.NextBlock, newBlock)
		// engine.Log.Info("关系   1111  5555555555555555 %d", len(newBlock.Group.Blocks))
		// for _, one := range newBlock.Group.Blocks {
		// 	engine.Log.Info("关系   1111  666666666 %d", one.Height)
		// }

	}
}

/*
	构建首个见证人组
*/
// func (this *WitnessChain) BuildFirstGroup(block *Block) {
// 	this.BuildWitnessGroup(true, block)
// 	group := this.backupGroupLast
// 	for group.PreGroup != nil {
// 		group = group.PreGroup
// 	}
// 	// this.witness = group.Witness[0]
// 	this.witnessGroup = group

// }

/*
	依次获取n个未分配组的见证人，构建一个新的见证人组
*/
func (this *WitnessChain) BuildWitnessGroup(first bool) {
	backupGroup := uint64(config.Witness_backup_group)

	// //判断是否有多半人出块
	// if !this.witnessChain.witnessGroup.BuildWitnessGroup() {
	// 	//没有多半人出块
	// 	return
	// }

	//
	// fmt.Println("-------构建见证人组")

	//判断备用见证人组数量是否足够，不够则创建
	totalBackupGroup := 0
	tag := false
	lastGroupHeight := uint64(config.Mining_group_start_height)
	var lastGroup *WitnessGroup
	for lastGroup = this.witnessGroup; lastGroup != nil; lastGroup = lastGroup.NextGroup {
		// fmt.Println("-------构建见证人组  222222222222222222222")
		// engine.Log.Info("-------构建见证人组  222222222222222222222")
		totalBackupGroup++
		tag = lastGroup.tag
		lastGroupHeight = lastGroup.Height
		if lastGroup.NextGroup == nil {
			break
		}

	}

	// engine.Log.Info("-------构建见证人组  3333333333333333333 %d %d", totalBackupGroup, backupGroup)

	if first {
		backupGroup = 0
	} else {

		total := this.witnessBackup.GetBackupWitnessTotal()
		groupNum := total / config.Mining_group_max
		if groupNum > backupGroup {
			backupGroup = groupNum
		}
	}

	// engine.Log.Info("-------构建见证人组  444444444444444 %d %d", totalBackupGroup, backupGroup)
	for i := uint64(totalBackupGroup); i <= backupGroup; i++ {
		//从候选见证人中评选出备用见证人来
		// witness := this.witnessBackup.CreateWitnessGroup(block)
		// if witness == nil {
		// 	return
		// }
		// this.AddWitness(witness, block)
		//把评选出来的所有备用见证人分组成为见证人组。
		// for {
		witnessGroup := this.GetOneGroupWitness()
		// engine.Log.Info("获取的见证人组中，见证人数量 %d", len(witnessGroup))
		if witnessGroup == nil || len(witnessGroup) < config.Mining_group_min {
			// fmt.Println("见证人数量不够 11111")
			// engine.Log.Info("见证人数量不够 %d %d", len(witnessGroup), config.Mining_group_min)
			break
		}
		// fmt.Println("111111111111")
		//找到上一个见证人的出块时间
		startTime := int64(0)
		if lastGroup != nil {
			w := lastGroup.Witness[len(lastGroup.Witness)-1]
			startTime = w.CreateBlockTime
		}
		//给见证人计算出块时间
		for _, one := range witnessGroup {
			startTime = startTime + config.Mining_block_time
			one.CreateBlockTime = startTime
		}

		// groupHeight := uint64(0)
		// if this.backupGroupLast != nil {
		// 	groupHeight = this.backupGroupLast.Height + 1
		// }
		if lastGroup != nil {
			lastGroupHeight++
		}
		newGroup := &WitnessGroup{
			PreGroup: lastGroup,       //备用见证人最后一个组
			Height:   lastGroupHeight, //
			Witness:  witnessGroup,    //本组见证人列表
			tag:      !tag,            //
			// check:    false,                //这个见证人组是否多半人出块，多半人出块则合法有效
		}

		tag = !tag
		for i, _ := range newGroup.Witness {
			newGroup.Witness[i].Group = newGroup
		}

		if lastGroup != nil {
			// fmt.Println("设置next为什么不生效", newGroup.Witness[0])
			lastGroup.Witness[len(lastGroup.Witness)-1].NextWitness = newGroup.Witness[0]
			newGroup.Witness[0].PreWitness = lastGroup.Witness[len(lastGroup.Witness)-1]
			lastGroup.NextGroup = newGroup
			newGroup.PreGroup = lastGroup
		}

		// fmt.Println("本次创建见证人组1", &newGroup, newGroup)

		// if this.backupGroupLast != nil {

		// 	this.backupGroupLast.NextGroup = newGroup
		// 	this.backupGroupLast.Witness[len(this.backupGroupLast.Witness)-1].NextWitness = newGroup.Witness[0]
		// 	newGroup.Witness[0].PreWitness = this.backupGroupLast.Witness[len(this.backupGroupLast.Witness)-1]
		// }
		// this.backupGroupLast = newGroup
		lastGroup = newGroup

		// fmt.Println("本次创建见证人组2", &lastGroup, lastGroup)
		// for i, one := range newGroup.Witness {
		// 	fmt.Println("跟踪见证人---", i, *one)
		// }
		// fmt.Println("555555555555")
		if this.witnessGroup == nil {
			// fmt.Println("666666666666666")
			this.witnessGroup = newGroup
		}
		// fmt.Println("本次创建见证人组3", this.witnessGroup)
		// }

	}
	this.PrintWitnessList()

}

/*
	暂停所有出块
*/
func (this *WitnessChain) StopAllMining() {
	// engine.Log.Info("-----开始暂停所有出块-----")
	//判断自己出块顺序的时间
	addr := keystore.GetCoinbase()
	// witnessTemp := this.witness
	witnessTemp := this.witnessGroup.Witness[0]
	for {
		witnessTemp = witnessTemp.NextWitness
		if witnessTemp == nil || witnessTemp.Group == nil {
			break
		}
		if !witnessTemp.Group.Task {
			continue
		}

		if !bytes.Equal(*witnessTemp.Addr, addr) {
			continue
		}

		select {
		case witnessTemp.StopMining <- false:
		default:
		}
		witnessTemp.Group.Task = false
	}
}

/*
	给备用见证人添加定时任务，定时出块
*/
func (this *WitnessChain) BuildMiningTime(force bool) error {

	// engine.Log.Info("=========开始构建所有出块=======")

	//判断自己出块顺序的时间
	addr := keystore.GetCoinbase()

	for witnessGroup := this.witnessGroup; witnessGroup != nil; witnessGroup = witnessGroup.NextGroup {
		// fmt.Println("构建出块 11111111111")
		if witnessGroup.Task {
			// fmt.Println("构建出块 2222222222222")
			continue
		}
		for _, witnessTemp := range witnessGroup.Witness {
			// fmt.Println("构建出块 3333333333333333333")
			if witnessTemp.Block != nil {
				continue
			}
			// fmt.Println("构建出块 555555555555555")

			// fmt.Println("这个见证人是否是自己")
			if bytes.Equal(*witnessTemp.Addr, addr) {
				// fmt.Println("构建出块 66666666666666666")
				future := int64(0) //出块时间，也可以记录上一个块的出块时间
				// fmt.Println("5555555555555")
				now := utils.GetNow() //time.Now().Unix()

				if witnessTemp.CreateBlockTime > now {
					future = witnessTemp.CreateBlockTime - now
				} else if witnessTemp.CreateBlockTime == now {
					future = 0
				} else {
					difference := now - witnessTemp.CreateBlockTime
					if difference < config.Mining_block_time {
						future = 0
					} else {
						// //如果参数加了init，可以继续出块
						// if force {
						// 	// fmt.Println("----------- 强制出块 -----------")
						// 	future = (totalBlock - 1) * config.Mining_block_time
						// 	hasOvertimeBlock++
						// } else {
						// 	//来不及出块了，算了
						// 	continue
						// }
						engine.Log.Warn("It's too late for a block. Forget it %s %s", time.Unix(witnessTemp.CreateBlockTime, 0), time.Unix(now, 0))

						//来不及出块了，算了
						continue
					}

				}

				//时间太少，来不及出块
				if future <= 0 {
					engine.Log.Info("There's too little time for a block %s %s", time.Unix(witnessTemp.CreateBlockTime, 0), time.Unix(now, 0))
					continue
				}

				//开关打开才能出块
				if !config.AlreadyMining {
					continue
				}

				// //NOTE 测试某个节点提前出块
				// engine.Log.Info("当前区块高度 %d", GetHighestBlock())
				// if GetHighestBlock() > 80 && i == 2 {
				// 	engine.Log.Info("当前正常时间 %d", future)
				// 	future = future - config.Mining_block_time - 1
				// 	engine.Log.Info("当前时间减少到 %d", future)
				// }

				// fmt.Println("是自己，自己多少秒钟后出块", config.Mining_block_time*(totalBlock))
				engine.Log.Info("Build blocks in %d seconds", future)
				// utils.AddTimetask(time.Now().Unix()+int64(config.Mining_block_time*(totalBlock)),
				// 	witnessTemp.TaskBuildBlock, Task_class_buildBlock, "")
				witnessTemp.Group.Task = true
				go witnessTemp.SyncBuildBlock(int64(future))
			} else {
				//给不是自己的见证人设置一个定时同步区块的方法
				witnessTemp.syncBlockTiming()
			}
			// witnessTemp = witnessTemp.NextWitness
		}

	}

	return nil
}

/*
	获取一组新的见证人组
	从未分配组的见证人中按顺序获取一个组的见证人
*/
func (this *WitnessChain) GetOneGroupWitness() []*Witness {
	//当前组见证人数量由候选见证人数量来定
	groupNum := config.Mining_group_min
	total := this.witnessBackup.GetBackupWitnessTotal()
	if total > config.Mining_group_max {
		groupNum = config.Mining_group_max
	} else if total < config.Mining_group_min {
		groupNum = config.Mining_group_min
	} else {
		groupNum = int(total)
	}

	// engine.Log.Info("本次计算出见证人数量 %d %d", groupNum, total)

	//备用见证人数量太少，则从候选见证人中选一批新的备用见证人
	if len(this.witnessNotGroup) < groupNum {
		// engine.Log.Info("数量太少，则评选出新的备用见证人 %d", len(this.witnessNotGroup))
		witness := this.witnessBackup.CreateWitnessGroup()
		if witness == nil {
			return nil
		}
		this.witnessNotGroup = append(this.witnessNotGroup, witness...)
		for _, temp := range this.witnessNotGroup {
			engine.Log.Info("print backup witness list %s", temp.Addr.B58String())
		}
		engine.Log.Info("print end")

	}

	//保存重复的见证人，需要向后面移动
	index := 0
	moveWitness := make([]*Witness, 0)
	witnessGroup := make([]*Witness, 0)
	for i, tempWitness := range this.witnessNotGroup {
		index = i
		//判断组中是否有重复的见证人
		isHave := false
		for _, one := range witnessGroup {
			if bytes.Equal(*tempWitness.Addr, *one.Addr) {
				// engine.Log.Info("有重复 %s", one.Addr.B58String())
				//把重复的见证人保存下来
				// moveOne := tempWitness
				moveWitness = append(moveWitness, tempWitness)
				// tempWitness = tempWitness.NextWitness
				// moveOne.NextWitness = nil
				// moveOne.PreWitness = nil
				isHave = true
				break
			}
		}
		//有重复的，跳出循环
		if isHave {
			continue
		}
		witnessGroup = append(witnessGroup, tempWitness)
		if len(witnessGroup) >= groupNum {
			break
		}
	}

	newWitnessNotGroup := this.witnessNotGroup[index+1:]
	this.witnessNotGroup = make([]*Witness, 0)
	//有重复的向后移动，从新排序
	for _, one := range moveWitness {
		this.witnessNotGroup = append(this.witnessNotGroup, one)
	}
	for _, one := range newWitnessNotGroup {
		this.witnessNotGroup = append(this.witnessNotGroup, one)
	}

	//从新建立引用关系
	tempWitness := witnessGroup[0]
	for i, _ := range witnessGroup {
		if i == 0 {
			continue
		}
		tempWitness.NextWitness = witnessGroup[i]
		witnessGroup[i].PreWitness = tempWitness
		tempWitness = witnessGroup[i]
	}

	// engine.Log.Info("最后的见证人数量 %d", len(witnessGroup))

	return witnessGroup

	// //----------------------------------------

	// witnessGroup := make([]*Witness, 0)
	// tempWitness := this.firstWitnessNotGroup
	// if tempWitness != nil {
	// 	for i := 0; i < config.Mining_group_max; i++ {
	// 		witnessGroup = append(witnessGroup, tempWitness)
	// 		tempWitness = tempWitness.NextWitness
	// 		if tempWitness == nil {
	// 			break
	// 		}
	// 	}
	// }

	// if len(witnessGroup) == groupNum {
	// 	this.firstWitnessNotGroup = tempWitness
	// 	return witnessGroup
	// }
	// //数量太少，则评选出新的备用见证人
	// engine.Log.Info("数量太少，则评选出新的备用见证人 %d", groupNum)
	// witness := this.witnessBackup.CreateWitnessGroup()
	// if witness == nil {
	// 	return witnessGroup
	// }
	// // this.AddWitness(witness)
	// this.witnessNotGroup = append(this.witnessNotGroup, witness...)

	// //
	// // for temp := this.firstWitnessNotGroup; temp != nil; temp = temp.NextWitness {
	// for _, temp := range this.witnessNotGroup {
	// 	engine.Log.Info("打印新的备用见证人列表 %s", temp.Addr.B58String())
	// }
	// engine.Log.Info("打印完毕")

	// //保存重复的见证人，需要向后面移动
	// moveWitness := make([]*Witness, 0)
	// witnessGroup = make([]*Witness, 0)
	// for tempWitness = this.firstWitnessNotGroup; tempWitness != nil &&
	// 	len(witnessGroup) < groupNum; tempWitness = tempWitness.NextWitness {
	// 	//判断组中是否有重复的见证人
	// 	isHave := false
	// 	for _, one := range witnessGroup {
	// 		if bytes.Equal(*tempWitness.Addr, *one.Addr) {
	// 			engine.Log.Info("有重复 %s", one.Addr.B58String())
	// 			//把重复的见证人保存下来
	// 			moveOne := tempWitness
	// 			moveWitness = append(moveWitness, tempWitness)
	// 			tempWitness = tempWitness.NextWitness
	// 			moveOne.NextWitness = nil
	// 			moveOne.PreWitness = nil
	// 			isHave = true
	// 			break
	// 		}
	// 	}
	// 	//有重复的，跳出循环
	// 	if isHave {
	// 		continue
	// 	}
	// 	witnessGroup = append(witnessGroup, tempWitness)

	// 	// tempWitness = tempWitness.NextWitness
	// 	// if tempWitness == nil {
	// 	// 	break
	// 	// }
	// }
	// if len(witnessGroup) != groupNum {
	// 	return nil
	// }

	// //有向后移动的，从新排序
	// if len(moveWitness) > 0 {
	// 	this.firstWitnessNotGroup = moveWitness[0]
	// 	newWitnessChainLast := this.firstWitnessNotGroup
	// 	for i, _ := range moveWitness {
	// 		if i == 0 {
	// 			continue
	// 		}
	// 		newWitnessChainLast.NextWitness = moveWitness[i]
	// 		moveWitness[i].PreWitness = newWitnessChainLast
	// 		newWitnessChainLast = moveWitness[i]
	// 	}
	// 	if tempWitness != nil {
	// 		newWitnessChainLast.NextWitness = tempWitness
	// 		tempWitness.PreWitness = newWitnessChainLast
	// 	}

	// } else {
	// 	this.firstWitnessNotGroup = tempWitness
	// }
	// //从新建立引用关系
	// tempWitness = witnessGroup[0]
	// for i, _ := range witnessGroup {
	// 	if i == 0 {
	// 		continue
	// 	}
	// 	tempWitness.NextWitness = witnessGroup[i]
	// 	witnessGroup[i].PreWitness = tempWitness
	// 	tempWitness = witnessGroup[i]
	// }

	// engine.Log.Info("最后的见证人数量 %d", len(witnessGroup))

	// return witnessGroup

}

/* 添加见证人，依次添加*/
// func (this *WitnessChain) AddWitness(newwitness []*Witness) {

// 	if this.firstWitnessNotGroup == nil {
// 		this.firstWitnessNotGroup = newwitness
// 	} else {
// 		//查找到最后一个见证人
// 		lastWitnessNotGroup := this.firstWitnessNotGroup
// 		for lastWitnessNotGroup.NextWitness != nil {
// 			lastWitnessNotGroup = lastWitnessNotGroup.NextWitness
// 		}
// 		newwitness.PreWitness = lastWitnessNotGroup
// 		lastWitnessNotGroup.NextWitness = newwitness
// 	}

// }

/*
	打印见证人列表
*/
func (this *WitnessChain) PrintWitnessList() {
	//打印未分组的见证人列表
	// this.witnessBackup.PrintWitnessBackup()

	group := this.witnessGroup
	for group != nil {
		engine.Log.Info("--------------")
		for _, one := range group.Witness {

			// gp, _ := fmt.Printf("%p", group)
			engine.Log.Info("witness tag %s %t %d %s %s %d", fmt.Sprintf("%p", one.WitnessBackupGroup), group.tag, group.Height, one.Addr.B58String(),
				time.Unix(one.CreateBlockTime, 0).Format("2006-01-02 15:04:05"), one.VoteNum)
		}
		group = group.NextGroup
	}

}

/*
	通过新区块，在未出块的见证人组中找到这个见证人
	@return    *Witness    找到的见证人
	@return    bool        是否需要同步
*/
func (this *WitnessChain) FindWitnessForBlock(bhvo *BlockHeadVO) (*Witness, bool) {
	var witness *Witness
	for group := this.witnessGroup; group != nil; group = group.NextGroup {
		// engine.Log.Info("设置已出的块 2222222222222222222 %d", group.Height)
		if group.Height < bhvo.BH.GroupHeight {
			// engine.Log.Info("设置已出的块 333333333333333333 %d", group.Height)
			continue
		}
		if group.Height > bhvo.BH.GroupHeight {
			// engine.Log.Info("设置已出的块 4444444444444444444 %d", group.Height)
			// engine.Log.Warn("不能导入之前已经确认的块")
			return nil, false
		}
		// engine.Log.Info("设置已出的块 55555555555555555 %d", group.Height)
		for _, one := range group.Witness {
			// engine.Log.Info("设置已出的块，对比一下 %s %s", bhvo.BH.Witness.B58String(), one.Addr.B58String())
			if !bytes.Equal(bhvo.BH.Witness, *one.Addr) {
				// fmt.Println("-=-=-=-=-=对比下一个1", one.Block, one.BlockHeight, one.Group.Height, bhvo.BH.Witness.B58String(), one.Addr.B58String())
				continue
			}
			now := utils.GetNow() //time.Now().Unix()
			//是未来的一个时间，直接退出
			if one.CreateBlockTime > now+config.Mining_block_time {

				// engine.Log.Info("是未来的一个时间，直接退出 %s %s %s", time.Unix(one.CreateBlockTime, 0).Format("2006-01-02 15:04:05"),
				// 	time.Unix(bhvo.BH.Time, 0).Format("2006-01-02 15:04:05"), time.Unix(now, 0).Format("2006-01-02 15:04:05"))

				break
			}

			// engine.Log.Info("设置已出的块 666666666666666 %d", group.Height)

			//找到这个见证人了
			witness = one
			// engine.Log.Info("找到这个见证人了")
			break
		}
		if witness != nil {
			// engine.Log.Info("找到这个见证人了 退出")
			break
		}
	}
	return witness, false
}

/*
	设置见证人生成的块
	只能设置当前组，不能设置其他组
	当本组所有见证人都出块了，将当前组见证人的变量指针修改为下一组见证人
	@return    bool    是否设置成功
*/
func (this *WitnessChain) SetWitnessBlock(bhvo *BlockHeadVO) bool {

	// engine.Log.Info("设置已出的块 1111111111111111111 %d %d %s", bhvo.BH.GroupHeight, bhvo.BH.Height, time.Unix(bhvo.BH.Time, 0).Format("2006-01-02 15:04:05"))

	//找到这个出块的见证人
	witness, needSync := this.FindWitnessForBlock(bhvo)
	if witness != nil && witness.Block != nil {
		//已经设置了就不需要重复设置了
		engine.Log.Warn("You don't need to set it again if it's already set")
		return false
	}

	//	var witness *Witness
	//	for group := this.witnessGroup; group != nil; group = group.NextGroup {
	//		// engine.Log.Info("设置已出的块 2222222222222222222 %d", group.Height)
	//		if group.Height < bhvo.BH.GroupHeight {
	//			// engine.Log.Info("设置已出的块 333333333333333333 %d", group.Height)
	//			continue
	//		}
	//		if group.Height > bhvo.BH.GroupHeight {
	//			// engine.Log.Info("设置已出的块 4444444444444444444 %d", group.Height)
	//			engine.Log.Warn("不能导入之前已经确认的块")
	//			return false
	//		}
	//		// engine.Log.Info("设置已出的块 55555555555555555 %d", group.Height)
	//		for _, one := range group.Witness {
	//			// engine.Log.Info("设置已出的块，对比一下 %s %s", bhvo.BH.Witness.B58String(), one.Addr.B58String())
	//			if !bytes.Equal(bhvo.BH.Witness, *one.Addr) {
	//				// fmt.Println("-=-=-=-=-=对比下一个1", one.Block, one.BlockHeight, one.Group.Height, bhvo.BH.Witness.B58String(), one.Addr.B58String())
	//				continue
	//			}
	//			now := utils.GetNow() //time.Now().Unix()
	//			//是未来的一个时间，直接退出
	//			if one.CreateBlockTime > now+config.Mining_block_time {

	//				// engine.Log.Info("是未来的一个时间，直接退出 %s %s %s", time.Unix(one.CreateBlockTime, 0).Format("2006-01-02 15:04:05"),
	//				// 	time.Unix(bhvo.BH.Time, 0).Format("2006-01-02 15:04:05"), time.Unix(now, 0).Format("2006-01-02 15:04:05"))

	//				break
	//			}

	//			// engine.Log.Info("设置已出的块 666666666666666 %d", group.Height)
	//			if one.Block != nil {
	//				engine.Log.Warn("已经设置了就不需要重复设置了")
	//				//已经设置了就不需要重复设置了
	//				return false
	//			}

	//			//找到这个见证人了
	//			witness = one
	//			// engine.Log.Info("找到这个见证人了")
	//			break
	//		}
	//		if witness != nil {
	//			// engine.Log.Info("找到这个见证人了 退出")
	//			break
	//		}
	//	}

	// engine.Log.Info("开始设置见证人出块 33333333333")
	if witness == nil {
		//没有找到这个见证人
		engine.Log.Warn("No witness found")

		if needSync {
			//从邻居节点同步
			this.chain.NoticeLoadBlockForDB(false)
		}
		return false
	}
	// engine.Log.Info("开始设置见证人出块 444444444444")

	if !bhvo.BH.CheckBlockHead(witness.Puk) {
		//区块验证不通过，区块不合法
		engine.Log.Warn("Block verification failed, block is illegal group:%d block:%d", bhvo.BH.GroupHeight, bhvo.BH.Height)
		// this.chain.NoticeLoadBlockForDB(false)
		return false
	}

	if bhvo.BH.Height != config.Mining_block_start_height {

		//查找出未确认的块
		preBlock, blocks := witness.FindUnconfirmedBlock()

		//检查区块中的交易是否正确
		for _, one := range bhvo.Txs {
			//自己的未打包交易，是已经验证过的合法交易，已经验证过就不需要重复验证了
			_, ok := this.chain.transactionManager.unpacked.Load(hex.EncodeToString(*one.GetHash()))
			if ok {
				continue
			}
			err := one.Check()
			if err != nil {
				engine.Log.Error("Illegal transaction %s %s", hex.EncodeToString(*one.GetHash()), err.Error())
				//交易不合法
				return false
			}
		}

		//检查未确认的块中的交易是否正确
		//未确认的交易
		unacknowledgedTxs := make([]TxItr, 0)
		//排除已经打包的交易
		exclude := make(map[string]string)
		for _, one := range blocks {
			_, txs, err := one.LoadTxs()
			if err != nil {
				engine.Log.Warn("not find transaction %s", err.Error())
				//找不到这个交易
				return false
			}
			for _, txOne := range *txs {
				exclude[hex.EncodeToString(*txOne.GetHash())] = ""
				unacknowledgedTxs = append(unacknowledgedTxs, txOne)
			}
		}

		sizeTotal := uint64(0) //保存区块所有交易大小
		for i, one := range bhvo.Txs {
			//判断重复的交易
			if !one.CheckRepeatedTx(unacknowledgedTxs...) {
				engine.Log.Warn("Transaction verification failed")
				//交易验证不通过
				return false
			}
			unacknowledgedTxs = append(unacknowledgedTxs, bhvo.Txs[i])
			sizeTotal = sizeTotal + uint64(len(*one.Serialize()))
		}
		//判断交易总大小
		if sizeTotal > config.Block_size_max {
			engine.Log.Warn("Transaction over total size %d", sizeTotal)
			//这里用continue是因为，排在后面的交易有可能占用空间小，可以打包到区块中去，使交易费最大化。
			return false
		}

		//检查区块奖励是否正确
		if witness.WitnessBackupGroup != preBlock.witness.WitnessBackupGroup {
			engine.Log.Info("check reward")
			vouts := preBlock.witness.WitnessBackupGroup.BuildRewardVouts(blocks)
			//对比vouts中分配比例是否正确
			haveReward := false //标记是否有区块奖励
			for _, one := range bhvo.Txs {
				if one.Class() != config.Wallet_tx_type_mining {
					continue
				}
				if haveReward {
					//如果一个块里有多个奖励交易，则不合法
					engine.Log.Warn("Illegal if there are multiple reward transactions in a block")
					return false
				}
				haveReward = true

				//对比奖励是否正确
				m := make(map[string]uint64) //key:string=奖励地址;value:uint64=奖励金额;
				for _, one := range *one.GetVout() {
					m[one.Address.B58String()] = one.Value
				}
				for _, one := range vouts {
					value, ok := m[one.Address.B58String()]
					if !ok {
						//没有这个人的奖励，则验证不通过
						engine.Log.Warn("Without this person's reward, the verification fails %s", one.Address.B58String())
						return false
					}
					if value != one.Value {
						//奖励数额不正确，则验证不通过
						engine.Log.Warn("If the reward amount is incorrect, the verification fails %d %d", value, one.Value)
						return false
					}
				}
			}
			if !haveReward {
				//如果没有区块奖励，则区块不合法
				engine.Log.Warn("If there is no block reward, the block is illegal")
				return false
			}
		} else {
			//判断每组不能有多个区块奖励交易
			for _, one := range bhvo.Txs {
				if one.Class() == config.Wallet_tx_type_mining {
					engine.Log.Warn("每组不能有多个区块奖励交易")
					return false
				}
			}

		}

	}
	//找到见证人了，不管这个见证人有没有开始出块，给他发送一个停止出块的信号
	select {
	case witness.StopMining <- false:
	default:
	}
	// engine.Log.Info("开始设置见证人出块 55555555555")
	//创建新的块
	newBlock := new(Block)
	newBlock.Id = bhvo.BH.Hash
	newBlock.Height = bhvo.BH.Height
	newBlock.PreBlockID = bhvo.BH.Previousblockhash
	// newBlock.PreBlock = make([]*Block, 0)
	// newBlock.NextBlock = make([]*Block, 0)

	// newBlock.PreBlock = append()

	//找到了见证人，将见证人标记为已经出块
	witness.Block = newBlock
	witness.Block.witness = witness

	// engine.Log.Info("找到这个见证人了 111 %d %s %v", witness.Group.Height, witness.Addr.B58String(), witness.Block)
	// if witness.PreWitness != nil {
	// 	engine.Log.Info("找到上一个见证人了 111 %d %s %v", witness.Group.Height, witness.Addr.B58String(), witness.PreWitness.Block)
	// }

	//整理已出区块之间的前后关系
	witness.Group.CollationRelationship()

	// engine.Log.Info("找到这个见证人了 222 %d %s %v", witness.Group.Height, witness.Addr.B58String(), witness.Block)

	// fmt.Println("开始设置见证人出块 2", witness)

	// fmt.Println("111", this.witnessGroup.Witness[0])

	//如果是创始区块，则设置见证人的出块时间
	if bhvo.BH.Height == config.Mining_block_start_height {
		witness.CreateBlockTime = bhvo.BH.Time
	}

	this.chain.SetPulledStates(bhvo.BH.Height)

	// engine.Log.Info("将新导入进来的交易UTXO输出标记为未使用")

	//将新导入进来的交易UTXO输出标记为未使用
	for i, one := range bhvo.Txs {
		have := false
		for j, vout := range *one.GetVout() {
			// engine.Log.Info("将新导入进来的交易UTXO输出标记为未使用 111111 %v", vout.Tx)
			if vout.Tx != nil || len(vout.Tx) <= 0 {
				// engine.Log.Info("将新导入进来的交易UTXO输出标记为未使用 22222222 %v", vout.Tx)
				// vout.Tx = nil
				(*one.GetVout())[j].Tx = nil
				have = true
			}
			// engine.Log.Info("将新导入进来的交易UTXO输出标记为未使用 333333 %v", vout.Tx)
		}
		if !have {
			continue
		}
		// engine.Log.Info("将新导入进来的交易UTXO输出标记为未使用 4444444")
		bs, err := bhvo.Txs[i].Json()
		if err != nil {
			engine.Log.Error("Save transaction JSON format error %s", err.Error())
			continue
		}
		// engine.Log.Info("将新导入进来的交易UTXO输出标记为未使用 55555555 %s", string(*bs))
		err = db.Save(*bhvo.Txs[i].GetHash(), bs)
		if err != nil {
			engine.Log.Error("Error saving transaction to database %s", err.Error())
			continue
		}
		// engine.Log.Info("将新导入进来的交易UTXO输出标记为未使用 66666666")
	}

	// bs, err := db.Find(bhvo.BH.Hash)
	// if err != nil {
	// 	return true
	// }
	// bh, err := ParseBlockHead(bs)
	// if err != nil {
	// 	return true
	// }
	// for _, one := range bh.Tx {
	// 	bs, _ := db.Find(one)
	// 	engine.Log.Info("打印首块交易 %s", string(*bs))
	// }

	return true

}

/*
	构建本组中的见证人出块奖励
	按股权分配
	只有见证人方式出块才统计
	组人数乘以每块奖励，再分配给实际出块的人
*/
func (this *WitnessGroup) CountReward(blockHeight uint64) *Tx_reward {

	vouts := make([]Vout, 0)

	//统计本组股权和交易手续费
	witnessPos := uint64(0) //见证人押金
	votePos := uint64(0)    //投票者押金
	allPos := uint64(0)     //股权数量
	allGas := uint64(0)     //计算交易手续费
	allReward := uint64(0)  //本组奖励数量
	// txs := make([]TxItr, 0)
	for _, one := range this.Witness {
		if one.Block == nil {
			continue
		}

		//计算交易手续费
		_, txs, _ := one.Block.LoadTxs()
		for _, one := range *txs {
			allGas = allGas + one.GetGas()
		}

		//计算股权
		allPos = allPos + (one.Score * 2) //计算股权的时候，见证人的股权要乘以2
		witnessPos = witnessPos + one.Score
		for _, vote := range one.Votes {
			allPos = allPos + vote.Scores
			votePos = votePos + vote.Scores
		}

		//计算区块奖励，第一个块产出80个币
		//每增加一定块数，产出减半，直到为0
		//最多减半9次，第10次减半后产出为0
		//		oneReward := uint64(config.Mining_reward)
		//		if one.Block.Height <= config.Mining_lastblock_reward {
		//			n := one.Block.Height / config.Mining_block_cycle
		//			for i := uint64(0); i < n; i++ {
		//				oneReward = oneReward / 2
		//			}
		//		} else {
		//			oneReward = 0
		//		}

		//按照发行总量及减半周期计算出块奖励
		oneReward := config.ClacRewardForBlockHeight(one.Block.Height)
		allReward = allReward + oneReward
	}

	//--------------所有交易手续费分给云存储节点---------------
	// cloudReward := uint64(0)
	nameinfo := name.FindNameToNet(config.Name_store)
	if nameinfo != nil && len(nameinfo.AddrCoins) != 0 {
		addrCoin := nameinfo.AddrCoins[utils.GetRandNum(int64(len(nameinfo.AddrCoins)))]
		// cloudReward = uint64(float64(allReward) * 0.8)
		vout := Vout{
			Value:   allGas,
			Address: addrCoin,
		}
		vouts = append(vouts, vout)
		// allReward = allReward - cloudReward
	}

	//---------------------------------------------------

	// allReward = allReward + allGas

	//计算见证人奖励
	// witnessRatio := int64(config.Mining_reward_witness_ratio * 100)
	// witnessReward := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(witnessRatio))
	// witnessReward = new(big.Int).Div(witnessReward, big.NewInt(100))
	countReward := uint64(0)
	for _, one := range this.Witness {
		//分配奖励是所有见证人组成员都要分配
		temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(one.Score*2)))
		value := new(big.Int).Div(temp, big.NewInt(int64(allPos)))
		//奖励为0的矿工交易不写入区块
		if value.Uint64() <= 0 {
			continue
		}
		vout := Vout{
			Value:   value.Uint64(),
			Address: *one.Addr,
		}
		vouts = append(vouts, vout)
		countReward = countReward + value.Uint64()
		//给投票者分配奖励
		for _, two := range one.Votes {
			temp := new(big.Int).Mul(big.NewInt(int64(allReward)), big.NewInt(int64(two.Scores)))
			value := new(big.Int).Div(temp, big.NewInt(int64(allPos)))
			//奖励为0的矿工交易不写入区块
			if value.Uint64() <= 0 {
				continue
			}
			vout := Vout{
				Value:   value.Uint64(),
				Address: *two.Addr,
			}
			vouts = append(vouts, vout)
			countReward = countReward + value.Uint64()
		}
	}
	//平均数不能被整除时候，剩下的给最后一个出块的见证人
	if len(vouts) > 0 {
		vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (allReward - countReward)
	}

	// //计算投票者的奖励
	// voteReward := allReward - witnessReward.Uint64()
	// countReward = uint64(0)
	// for _, one := range this.Witness {
	// 	for _, two := range one.Votes {
	// 		temp := new(big.Int).Mul(big.NewInt(int64(voteReward)), big.NewInt(int64(two.Score)))
	// 		value := new(big.Int).Div(temp, big.NewInt(int64(votePos)))
	// 		//奖励为0的矿工交易不写入区块
	// 		if value.Uint64() <= 0 {
	// 			continue
	// 		}
	// 		vout := Vout{
	// 			Value:   value.Uint64(),
	// 			Address: *two.Addr,
	// 		}
	// 		vouts = append(vouts, vout)
	// 		countReward = countReward + value.Uint64()
	// 	}
	// }
	// //平均数不能被整除时候，剩下的给最后一个投票者
	// vouts[len(vouts)-1].Value = vouts[len(vouts)-1].Value + (voteReward - countReward)

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

/*
	判断是否是本组首个见证人出块
*/
func (this *WitnessGroup) FirstWitness() bool {
	for _, one := range this.Witness {
		if one.Block != nil {
			return false
		}
	}
	return true
}

/*
	分配出块奖励

*/
func (this *WitnessGroup) DistributionRewards() {

}

/*
	查询见证人是否在备用见证人列表中
*/
func (this *WitnessChain) FindWitness(addr crypto.AddressCoin) bool {
	witnessTemp := this.witnessGroup.Witness[0]
	for {
		witnessTemp = witnessTemp.NextWitness
		if witnessTemp == nil || witnessTemp.Group == nil {
			break
		}
		if bytes.Equal(*witnessTemp.Addr, addr) {
			return true
		}
	}
	return false
}

/*
	见证人定时同步区块
*/
func (this *Witness) syncBlockTiming() {
	//同步区块没有完成则不定时同步
	if !forks.GetLongChain().SyncBlockFinish {
		return
	}
	if this.syncBlockOnce != nil {
		return
	}
	this.syncBlockOnce = new(sync.Once)
	var syncBlock = func() {
		go func() {
			bfw := BlockForWitness{
				GroupHeight: this.Group.Height, //见证人组高度
				Addr:        *this.Addr,        //见证人地址
			}

			now := utils.GetNow()
			// engine.Log.Info("查看时间 %d %d %d", now, this.CreateBlockTime, now-this.CreateBlockTime)

			if this.CreateBlockTime < now {
				intervalTime := now - this.CreateBlockTime
				if intervalTime > config.Mining_block_time*2 {
					// engine.Log.Info("时间太久远了，就不需要添加定时同步区块了")
					return
				}
			}

			waitTime := this.CreateBlockTime - utils.GetNow()
			//
			if waitTime < 0 {
				waitTime = 0
			}

			//给等待同步时间设置一个随机，不要所有节点同一时间开始同步
			delayTime_min := time.Duration(config.Mining_block_time * time.Second / 3) //延迟同步最小时间为出块时间的1/3
			delayTime_max := time.Duration(config.Mining_block_time * time.Second / 2) //延迟同步最长时间为出块时间的一半
			delayTime := delayTime_min + time.Duration(utils.GetRandNum(int64(delayTime_max-delayTime_min)))
			delayTime = (time.Duration(waitTime) * time.Second) + delayTime

			engine.Log.Info("Groups %d Time to wait for synchronization %d %s", this.Group.Height, waitTime, delayTime)
			// time.Sleep((time.Duration(waitTime) * time.Second) + (time.Second * 4)) //加n秒，n秒钟后再同步
			time.Sleep(delayTime) //加n秒，n秒钟后再同步

			intervalTime := time.Second                                                            //同步时间间隔
			intervalTotal := ((config.Mining_block_time * time.Second) - delayTime) / intervalTime //间隔次数
			// engine.Log.Info("%d 间隔时间 %d   间隔次数 %d", this.Group.Height, intervalTime, intervalTotal)

			this.syncBlock(int(intervalTotal), intervalTime, &bfw)

			// engine.Log.Info("%d 同步 end", this.Group.Height)

		}()
	}
	this.syncBlockOnce.Do(syncBlock)
}

/*
	见证人同步区块
	@total           uint64           总共同步多少次
	@intervalTime    time.Duration    同步失败间隔时间
*/
func (this *Witness) syncBlock(total int, intervalTime time.Duration, bfw *BlockForWitness) {
	bs, err := json.Marshal(bfw)
	if err != nil {
		return
	}
	for i := int64(0); i < int64(total); i++ {
		if this.Block != nil {
			// engine.Log.Info("%d 这个区块已经有了，不需要同步了", this.Group.Height)
			return
		}
		//开始从邻居节点同步区块
		broadcasts := append(nodeStore.GetLogicNodes(), nodeStore.GetProxyAll()...)
		// engine.Log.Info("%d 邻居节点个数 %d", this.Group.Height, len(broadcasts))
		for j, _ := range broadcasts {
			if this.Block != nil {
				// engine.Log.Info("%d 这个区块已经有了，不需要同步了", this.Group.Height)
				return
			}
			engine.Log.Info("%d Synchronize blocks from neighbor nodes %s", this.Group.Height, broadcasts[j].B58String())
			// engine.Log.Info("%d 555555555555555555555", this.Group.Height)
			message, ok := message_center.SendNeighborMsg(config.MSGID_getblockforwitness, broadcasts[j], &bs)
			if !ok {
				// engine.Log.Info("这个邻居节点消息发送不成功")
				continue
			}
			// engine.Log.Info("%d 66666666666666666666", this.Group.Height)
			bs := flood.WaitRequest(config.CLASS_wallet_getblockforwitness, hex.EncodeToString(message.Body.Hash), 1)
			if bs == nil {
				// engine.Log.Warn("%d 收到查询区块回复消息超时  %s", this.Group.Height, broadcasts[j].B58String())
				// lock.Unlock()
				continue
			}
			// engine.Log.Info("%d 收到查询区块回复消息  %s", this.Group.Height, broadcasts[j].B58String())
			//导入区块
			bhVO, err := ParseBlockHeadVO(bs)
			if err != nil {
				// engine.Log.Warn("%d 收到查询区块回复消息 error: %s", this.Group.Height, err.Error())
				continue
			}
			forks.AddBlockHead(bhVO)
			// engine.Log.Info("%d 8888888888888888888", this.Group.Height)
			//等待导入
			//			time.Sleep(intervalTime)
			//			break
		}

		time.Sleep(intervalTime)
	}
}
