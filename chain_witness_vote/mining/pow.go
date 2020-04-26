package mining

// import (
// 	"bytes"
// 	"sync"
// )

// var powBlockHeadLock = new(sync.RWMutex)
// var pow *POW

// // var powBlockHead *BlockHead //正在挖矿寻找幸运数字的块
// // var stopSignalChan = make(chan bool, 1)

// type POW struct {
// 	bh             *BlockHead
// 	stopSignalChan chan bool
// }

// /*
// 	其他节点已经出块，停止寻找幸运数字
// */
// func stopFindNonce(bh *BlockHead) {
// 	//	fmt.Println("调用中断方法")
// 	powBlockHeadLock.Lock()
// 	if pow != nil && bytes.Equal(pow.bh.Previousblockhash[0], bh.Previousblockhash[0]) {
// 		// fmt.Println("其他矿工率先出块，中断挖矿，下次继续努力！")
// 		select {
// 		case pow.stopSignalChan <- true:
// 		default:
// 		}
// 	}
// 	powBlockHeadLock.Unlock()
// 	//	fmt.Println("调用中断方法完成")
// }

// /*
// 	开始寻找幸运数字
// */
// func findNonce(bh *BlockHead) (ok bool) {
// 	powBlockHeadLock.Lock()
// 	//若还有其他协程在挖矿，则先暂停其他挖矿程序，再创建新的挖矿程序
// 	if pow != nil {
// 		select {
// 		case pow.stopSignalChan <- true:
// 		default:
// 		}
// 	}
// 	pow = new(POW)
// 	pow.bh = bh
// 	pow.stopSignalChan = make(chan bool, 1)

// 	// powBlockHead = bh
// 	powBlockHeadLock.Unlock()
// 	ok = <-bh.FindNonce(config.Mining_difficulty, pow.stopSignalChan)
// 	// powBlockHeadLock.Lock()
// 	// powBlockHead = nil
// 	// powBlockHeadLock.Unlock()
// 	return
// }
