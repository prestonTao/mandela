package proto

import (
	// "fmt"
	mc "mandela/core/message_center"
	"mandela/core/utils"
	"sync"
	"time"

	"gopkg.in/eapache/queue.v1"
)

const (
	ResendNum    = 3   //消息重发次数
	SendInterval = 5   //重复发送时间间隔，秒
	TimeOut      = 90  //消息超时间隔,秒
	DelTimeOut   = 300 //清理消息时间，秒
)

var (
	Que *queue.Queue //队列 msgid
	Msg *sync.Map    //消息 msgid:QueueData
)

type QueueData struct {
	Msgid   *utils.Multihash //消息id
	ToID    *utils.Multihash //接收者id
	Message *mc.Mess         //消息体
	IsSend  bool             //是否发送成功
	Time    int64            //time.Now().UnixNano()
	Num     int              //发送次数
	Result  chan bool        //发送结果
}

func init() {
	Que = queue.New()
	Msg = new(sync.Map)
	timing()
}

//消息写入队列
func addQueue(msgid string) {
	Que.Add(msgid)
}

//消息写入缓存
func addMap(qd *QueueData) {
	qd.Time = time.Now().UnixNano()
	Msg.Store(qd.Msgid.B58String(), qd)
}

//修改消息发送次数
func updateNum(msgid *utils.Multihash, num int) bool {
	qd, ok := Msg.Load(msgid.B58String())
	if ok {
		queuedata := qd.(*QueueData)
		queuedata.Num += num
		Msg.Store(msgid.B58String(), queuedata)
		return true
	}
	return false
}

//修改消息发送成功状态
func updateSucc(msgid *utils.Multihash, suc bool) bool {
	qd, ok := Msg.Load(msgid.B58String())
	if ok {
		queuedata := qd.(*QueueData)
		queuedata.IsSend = suc
		Msg.Store(msgid.B58String(), queuedata)
		return true
	}
	return false
}

//等待消息结果
func waitState(msgid *utils.Multihash) bool {
	qd, ok := Msg.Load(msgid.B58String())
	if ok {
		queuedata := qd.(*QueueData)
		num := queuedata.Num
		if num > ResendNum {
			return false
		}
		result := queuedata.Result
		ticker := time.NewTicker(TimeOut * time.Second)
		select {
		case <-ticker.C:
			// fmt.Println("消息发失败,超时")
			return false
		case <-result:
			ticker.Stop()
			Msg.Delete(msgid.B58String())
			return true
		}
	}
	return false
}

//获取消息发送状态
func getState(msgid *utils.Multihash) bool {
	qd, ok := Msg.Load(msgid.B58String())
	if ok {
		queuedata := qd.(*QueueData)
		if queuedata.IsSend {
			return true
		} else {
			return false
		}
	}
	return true
}

//消息发送成功回调
func callbackSuccess(msgid *utils.Multihash) error {
	qd, ok := Msg.Load(msgid.B58String())
	if ok {
		queuedata := qd.(*QueueData)
		queuedata.IsSend = true
		Msg.Store(msgid.B58String(), queuedata)
		result := queuedata.Result
		result <- true
	}
	return nil
}
