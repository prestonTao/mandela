package mining

import (
	"errors"
)

var (
	ERROR_repeat_import_block                = errors.New("Repeat import block")                                                 //导入重复的区块
	ERROR_fork_import_block                  = errors.New("The front block cannot be found, and the new block is discontinuous") //导入的区块分叉了
	ERROR_import_block_height_not_continuity = errors.New("Import block height discontinuity")                                   //导入的区块高度不连续
)
