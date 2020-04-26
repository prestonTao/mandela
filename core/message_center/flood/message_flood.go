package flood

// import (
// 	"mandela/core/config"
// 	"mandela/core/utils"
// 	"sync"
// 	"time"
// )

// /*
// 	防止消息泛洪
// */

// var (
// 	task = utils.NewTask(msgTimeOutProsess)
// 	// msgHashLock = new(sync.RWMutex)
// 	// msgHash     = make(map[string]int64)
// 	msgHash = new(sync.Map) //key:string=类型;value:int64=unix时间;
// )

// /*
// 	添加一个消息超时
// */
// func AddMsgTimeOut(md5 string) {
// 	now := time.Now().Unix()
// 	msgHash.Store(md5, now)
// 	// msgHashLock.Lock()
// 	// msgHash[md5] = now
// 	// msgHashLock.Unlock()
// 	task.Add(now+60*10, config.TSK_msg_timeout_remove, md5)
// }

// /*
// 	检查一个消息是否超时或者非法
// */
// func CheckMsgTimeOut(md5 string) (ok bool) {
// 	_, ok = msgHash.Load(md5)
// 	msgHash.Delete(md5)
// 	// msgHashLock.Lock()
// 	// _, ok = msgHash[md5]
// 	// if ok {
// 	// 	delete(msgHash, md5)
// 	// }
// 	// msgHashLock.Unlock()
// 	return
// }

// /*
// 	消息超时删除md5
// */
// func msgTimeOutProsess(class, params string) {
// 	switch class {
// 	case config.TSK_msg_timeout_remove: //删除超时的消息md5
// 		//		fmt.Println("开始删除临时域名", tempName)
// 		//		tempNameLock.Lock()
// 		//		delete(tempName, params)
// 		//		tempNameLock.Unlock()
// 		//		fmt.Println("删除了这个临时域名", params, tempName)
// 	default:
// 		//		//剩下是需要更新的域名
// 		//		flashName := FlashName{
// 		//			Name:  params,
// 		//			Class: class,
// 		//		}
// 		//		OutFlashName <- &flashName
// 	}

// }
