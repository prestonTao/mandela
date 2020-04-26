package proto

import (
	// "fmt"
	"time"
)

func timing() {
	go func() {
		for {
			for i := 0; i < Que.Length(); i++ {
				quedata := Que.Get(i)
				// fmt.Println("queue:", i, quedata)
			}
			Msg.Range(func(k, v interface{}) bool {
				//fmt.Printf("map:%v %+v %d\n", k, v, v.(*QueueData).Time)
				return true
			})
			doing()
			<-time.NewTicker(SendInterval * time.Second).C
		}
	}()
}
func doing() {
	//fmt.Println("start...")
	if Que.Length() > 0 {
		qd := Que.Peek()
		if qd != nil {
			msgid := qd.(string)
			mess, ok := Msg.Load(msgid)
			if ok {
				quedata := mess.(*QueueData)
				//清理消息
				if (time.Now().UnixNano()-quedata.Time)/1e9 > DelTimeOut {
					Msg.Delete(msgid)
				}
				if quedata.Num > ResendNum {
					Que.Remove()
					return
				}
				// fmt.Printf("*******第%d次重发******\n", quedata.Num)
				err := sendMessage(quedata.Message)
				if err == nil {
					Que.Remove()
				}
			}
		}
	}
}
