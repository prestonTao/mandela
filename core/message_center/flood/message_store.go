package flood

import (
	"sync"
)

type StoreMemory struct {
	store *sync.Map //
}

func (this *StoreMemory) Load() {

}

//--------------------------------------

// import (
// 	"mandela/core/utils"
// 	"sync"
// 	"time"
// )

// var sendHash = new(sync.Map) //保存1分钟内的消息sendhash，用于判断重复消息
// var sendhashTask = utils.NewTask(sendhashTaskFun)

// func sendhashTaskFun(class, params string) {
// 	sendHash.Delete(params)
// }

// /*
// 	检查这个消息是否发送过
// */
// func CheckHash(sendhash string) bool {
// 	_, ok := sendHash.Load(sendhash)
// 	if !ok {
// 		sendHash.Store(sendhash, nil)
// 		sendhashTask.Add(time.Now().Unix()+60, "", sendhash)
// 	}
// 	return !ok
// }
