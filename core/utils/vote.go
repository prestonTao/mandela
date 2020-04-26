/*
	投票系统
*/
package utils

import (
	"encoding/hex"
	"sync"
	"time"
)

type pollManager struct {
	class    string
	lock     *sync.RWMutex
	data     map[string]*poll
	timetask *Task
}

func NewPollManager() *pollManager {
	pm := &pollManager{
		//		class: "remove",
		lock: new(sync.RWMutex),
		data: make(map[string]*poll),
	}
	pm.timetask = NewTask(pm.removePross)
	return pm
}

type poll struct {
	lock  *sync.RWMutex
	votes map[string]*vote
}

type vote struct {
	lock      *sync.RWMutex
	replyHash map[string]int
}

///*
//	投一票赞成
//	10分钟内收集8票，没收集够10分钟内会删除
//	@success    bool    是否通过投票
//*/
//func (this *pollManager) InitClass(class string) {
//	this.lock.Lock()
//	p := poll{
//		lock:  new(sync.RWMutex),
//		votes: make(map[string]*vote),
//	}
//	this.data[class] = &p
//	this.lock.Unlock()
//}

/*
	投一票赞成
	10分钟内收集8票，没收集够10分钟内会删除
	@success    bool    是否通过投票
*/
func (this *pollManager) Vote(class string, params []byte, replyHash string) (success bool) {
	//	fmt.Println("投票 11111111111111")
	this.lock.RLock()
	one, ok := this.data[class]
	this.lock.RUnlock()
	if ok {
		//		fmt.Println("投票 22222222222222")
		one.lock.Lock()
		vOne, ok := one.votes[hex.EncodeToString(params)]
		if ok {
			//			fmt.Println("投票 333333333333")
			//			fmt.Println("hash", replyHash)
			vOne.replyHash[replyHash] = 0
			if len(vOne.replyHash) >= 8 {
				success = true
			}
			//			fmt.Println("投票 44444444444444")
			//			fmt.Println("all hash", vOne.replyHash)
		} else {
			//			fmt.Println("投票 555555555555555555")
			//			fmt.Println("添加一个key")
			v := vote{
				lock:      new(sync.RWMutex),
				replyHash: make(map[string]int),
			}
			v.replyHash[replyHash] = 0
			one.votes[hex.EncodeToString(params)] = &v
			//			fmt.Println("添加一个定时任务")
			//			fmt.Println("投票 66666666666")
			this.timetask.Add(time.Now().Unix()+3, class, params)
			//			fmt.Println("投票 7777777777")
		}
		one.lock.Unlock()
	} else {
		//		fmt.Println("添加一个class")
		p := poll{
			lock:  new(sync.RWMutex),
			votes: make(map[string]*vote),
		}
		this.lock.Lock()
		this.data[class] = &p
		this.lock.Unlock()
		this.Vote(class, params, replyHash)
	}
	//	fmt.Println(this.data)
	//	fmt.Println("投票 99999999999999")
	return
}

func (this *pollManager) removePross(class string, param []byte) {
	//	fmt.Println("删除 ", class, param)
	this.lock.RLock()
	pOne, ok := this.data[class]
	this.lock.RUnlock()
	//	fmt.Println(pOne, ok)
	if !ok {
		//		fmt.Println("错误")
		return
	}
	pOne.lock.Lock()
	delete(pOne.votes, hex.EncodeToString(param))
	pOne.lock.Unlock()
	//	fmt.Println(this.data)

}
