package mining

// "mandela/config"
// "mandela/core/engine"
// "mandela/core/utils"

func FinishFirstLoadBlockChain() {
	// engine.Log.Info("开始拉起链端")

	//拉起链端之前确认之前未确认的块
	// GetLongChain().witnessChain.BuildBlockGroup(bhvo)
	//先统计之前的区块
	// for buildGroup := group; buildGroup != nil && buildGroup.BlockGroup == nil; buildGroup = buildGroup.PreGroup {
	// 	buildGroup.BuildGroup()
	// }

	GetLongChain().witnessChain.CompensateWitnessGroup()

	// engine.Log.Info("拉起链端完成")
}
