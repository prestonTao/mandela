package proto

import (
	"mandela/core/utils"
)

//发送消息
func SendMsg(uid *utils.Multihash, content []byte) (msgid *utils.Multihash) {
	mess := getMessage(uid, content)
	msgid = mess.Body.Hash
	qd := new(QueueData)
	qd.Num = 1
	qd.Msgid = msgid
	qd.ToID = uid
	qd.Result = make(chan bool)
	qd.Message = mess
	addMap(qd)
	go func() {
		err := sendMessage(mess)
		if err != nil {
			addQueue(msgid.B58String())
		}
	}()
	return
}

//等待消息结果
func WaitState(msgid *utils.Multihash) bool {
	return waitState(msgid)
}

//获取消息状态
func GetState(msgid *utils.Multihash) bool {
	return getState(msgid)
}
