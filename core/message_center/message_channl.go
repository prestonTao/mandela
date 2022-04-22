package message_center

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/protos/go_protos"
	"mandela/sqlite3_db"
	"bytes"
	"errors"
	"runtime"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
)

var msgchannl = make(chan *MsgHolderOne, 1000)

var msgHolderLock = new(sync.RWMutex)
var msgHolder = make(map[string]*MsgHolder) // new(sync.Map)

func init() {
	go func() {
		var mh *MsgHolder
		var ok bool
		var msgOne *MsgHolderOne
		// var mhItr interface{}

		for msgOne = range msgchannl {
			// if nodeStore.FindWhiteList(&msgOne.Addr) {
			// 	engine.Log.Info("recv new multicast message from:%s %s", msgOne.Addr.B58String(), hex.EncodeToString(msgOne.MsgHash))
			// }

			//TODO 查询记录是否存在，换个查询接口
			_, err := new(sqlite3_db.MessageCache).FindByHash(msgOne.MsgHash) //.Add(message.KeyDB(), headBs, bodyBs)
			if err == nil {
				//消息已经存在
				// if nodeStore.FindWhiteList(&msgOne.Addr) {
				// 	engine.Log.Info("message is exist from:%s %s", msgOne.Addr.B58String(), hex.EncodeToString(msgOne.MsgHash))
				// }
				SendNeighborReplyMsg(msgOne.Message, config.MSGID_multicast_return, nil, msgOne.Session)
				continue
			}

			//
			//先查询消息持有者缓存中是否存在
			// mhItr, ok = msgHolder.Load(utils.Bytes2string(msgOne.MsgHash))
			msgHolderLock.Lock()
			mh, ok = msgHolder[utils.Bytes2string(msgOne.MsgHash)]
			if !ok {
				mh = CreateMsgHolder(msgOne)
				// msgHolder.Store(utils.Bytes2string(msgOne.MsgHash), mh)
				msgHolder[utils.Bytes2string(msgOne.MsgHash)] = mh
				msgHolderLock.Unlock()
				// go GetMulticastMsg(mh)
			} else {
				msgHolderLock.Unlock()
				// mh = mhItr.(*MsgHolder)
				//判断这个持有者是否存在
				// engine.Log.Info("添加一个消息持有者:%s", msgOne.Addr.B58String())
				mh.AddHolder(msgOne)
			}
			// if nodeStore.FindWhiteList(&msgOne.Addr) {
			// 	engine.Log.Info("start sync multicast message from:%s %s", msgOne.Addr.B58String(), hex.EncodeToString(msgOne.MsgHash))
			// }

			//判断同步消息协程是否在运行
			isRun := false
			//
			mh.Lock.RLock()
			if !mh.IsRun && !mh.IsFinish {
				mh.IsRun = true
				isRun = true
			}
			mh.Lock.RUnlock()
			if !isRun {
				//同步消息协程在运行，也许在等待新的消息提供者
				select {
				case mh.RunSignal <- false:
				default:
				}
				continue
			}
			go mh.GetMulticastMsg()
			// msgHolderLock.Unlock()
		}
	}()
}

type MsgHolderOne struct {
	MsgHash []byte //消息hash
	Addr    nodeStore.AddressNet
	Message *Message
	Session engine.Session
}

type MsgHolder struct {
	MsgHash   []byte          //消息hash
	Lock      *sync.RWMutex   //
	Holder    []*MsgHolderOne //消息持有者
	HolderTag []bool          //消息持有者同步标志，超时或者同步失败，设置为true
	IsRun     bool            //是否正在同步
	IsFinish  bool            //是否已经同步到了消息
	RunSignal chan bool       //继续运行信号
}

func (this *MsgHolder) AddHolder(holder *MsgHolderOne) {
	this.Lock.Lock()
	have := false
	for _, one := range this.Holder {
		if bytes.Equal(one.Addr, holder.Addr) {
			have = true
			break
		}
	}
	if !have {
		this.Holder = append(this.Holder, holder)
		this.HolderTag = append(this.HolderTag, false)
	}
	this.Lock.Unlock()
}

func (this *MsgHolder) GetHolder() (holder *MsgHolderOne) {
	this.Lock.RLock()
	for i, one := range this.HolderTag {
		if one {
			continue
		}
		holder = this.Holder[i]
	}
	this.Lock.RUnlock()
	return
}
func (this *MsgHolder) SetHolder(holder *MsgHolderOne) {
	this.Lock.RLock()
	for i, one := range this.Holder {
		if bytes.Equal(one.Addr, holder.Addr) {
			this.HolderTag[i] = true
			break
		}
	}
	this.Lock.RUnlock()
	return
}

func CreateMsgHolder(holder *MsgHolderOne) *MsgHolder {
	msgHolder := MsgHolder{
		MsgHash:   holder.MsgHash,
		Lock:      new(sync.RWMutex),
		Holder:    make([]*MsgHolderOne, 0),
		HolderTag: make([]bool, 0),
	}
	msgHolder.Holder = append(msgHolder.Holder, holder)
	msgHolder.HolderTag = append(msgHolder.HolderTag, false)
	return &msgHolder
}

/*
	同步新消息并广播给其他节点
*/
func (this *MsgHolder) GetMulticastMsg() {
	goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	engine.AddRuntime(file, line, goroutineId)
	defer engine.DelRuntime(file, line, goroutineId)

	//循环从消息持有者同步消息
	var messageProto *go_protos.MessageMulticast
	var message *Message
	var holder *MsgHolderOne
	var err error
	for {
		holder = this.GetHolder()
		if holder == nil || len(holder.Addr) <= 0 {
			// engine.Log.Info("not holder")
			//没有可用的消息提供者，设置超时删除这个协程
			timeout := time.NewTimer(time.Second * 10)
			select {
			case <-this.RunSignal:
				timeout.Stop()
				continue
			case <-timeout.C:
				break
			}
			break
		}
		this.SetHolder(holder)
		// engine.Log.Info("555555555555")
		messageProto, err = SyncMulticastMsg(holder.Addr, this.MsgHash)
		if err != nil {
			// engine.Log.Info("666666666666")
			engine.Log.Info(err.Error())
			continue
		}
		//解析获取到的交易
		message, err = ParserMessageProto(messageProto.Head, messageProto.Body, 0)
		if err != nil {
			engine.Log.Error("proto unmarshal error %s", err.Error())
			continue
		}
		err = message.ParserContentProto()
		if err != nil {
			engine.Log.Error("proto unmarshal error %s", err.Error())
			continue
		}
		// engine.Log.Info("保存这个消息到数据库 hash:%s", hex.EncodeToString(message.Body.Hash))
		//先保存这个消息到数据库
		err = new(sqlite3_db.MessageCache).Add(message.Body.Hash, messageProto.Head, messageProto.Body)
		if err != nil {
			engine.Log.Error(err.Error())
			continue
		}
		break
	}
	// engine.Log.Info("77777777777")
	this.Lock.Lock()
	this.IsRun = false
	if message != nil && err == nil {
		this.IsFinish = true
	} else {
		this.IsRun = false
	}
	this.Lock.Unlock()
	// engine.Log.Info("888888888888888888")
	if message == nil || err != nil {
		//未同步到消息，又超时了
		msgHolderLock.Lock()
		delete(msgHolder, utils.Bytes2string(this.MsgHash))
		msgHolderLock.Unlock()
		return
	}
	// fromAddr := nodeStore.AddressNet([]byte(holder.Session.GetName()))
	// if nodeStore.FindWhiteList(&fromAddr) {
	// 	engine.Log.Info("sync multicast message success from:%s %s", fromAddr.B58String(), hex.EncodeToString(this.MsgHash))
	// }

	// msgHolder.Delete(utils.Bytes2string(this.MsgHash))
	msgHolderLock.Lock()
	delete(msgHolder, utils.Bytes2string(this.MsgHash))
	msgHolderLock.Unlock()
	//同步到消息，给每个发送者都回复收到
	for _, one := range this.Holder {
		SendNeighborReplyMsg(one.Message, config.MSGID_multicast_return, nil, one.Session)
	}

	//继续广播给其他节点
	// engine.Log.Info("同步到了广播消息")
	if nodeStore.NodeSelf.IsSuper {
		//广播给其他超级节点
		utils.Go(func() {
			goroutineId := utils.GetRandomDomain() + utils.TimeFormatToNanosecondStr()
			_, file, line, _ := runtime.Caller(0)
			engine.AddRuntime(file, line, goroutineId)
			defer engine.DelRuntime(file, line, goroutineId)
			//先发送给超级节点
			// superNodes := nodeStore.GetIdsForFar(message.Head.SenderSuperId)
			// whiltlistNodes := nodeStore.GetWhiltListNodes()

			superNodes := append(nodeStore.GetLogicNodes(), nodeStore.GetNodesClient()...)
			//广播给代理对象
			proxyNodes := nodeStore.GetProxyAll()
			BroadcastsAll(version_multicast, 0, nil, superNodes, proxyNodes, &this.MsgHash)
			return
		})
	}
	// engine.Log.Info("99999999999999")
	//自己处理
	h := router.GetHandler(message.Body.MessageId)
	if h == nil {
		engine.Log.Info("This broadcast message is not registered:", message.Body.MessageId)
		return
	}

	//

	// engine.Log.Info("有广播消息，消息编号 %d", message.Body.MessageId)
	//TODO 这里获取不到控制器了,因此传空
	h(nil, engine.Packet{}, message)
}

/*
	去邻居节点同步广播消息
*/
func SyncMulticastMsg(id nodeStore.AddressNet, hash []byte) (*go_protos.MessageMulticast, error) {
	// if nodeStore.FindWhiteList(&id) {
	// 	engine.Log.Info("sync multicast message from:%s %s", id.B58String(), hex.EncodeToString(hash))
	// }
	session, ok := engine.GetSession(utils.Bytes2string(id))
	if !ok {
		engine.Log.Error("get node conn fail %s", id.B58String())
		// nodeStore.DelNode(recvid)
		return nil, errors.New("get node conn fail")
	}
	head := NewMessageHead(nil, nil, false)
	body := NewMessageBody(0, &hash, 0, nil, 0) //广播采用同步机制后，不需要真实msgid，所以设置为0
	message := NewMessage(head, body)
	message.BuildHash()
	// fmt.Println("给这个session发送消息", recvid.B58String())
	mheadBs := head.Proto()
	mbodyBs := body.Proto()
	err := session.Send(version_multicast_sync, &mheadBs, &mbodyBs, false)
	if err != nil {
		engine.Log.Error("msg send error to:%s %s", id.B58String(), err.Error())
		return nil, err
	}

	bs, err := flood.WaitRequest(config.CLASS_engine_multicast_sync, utils.Bytes2string(message.Body.Hash), config.Mining_block_time) //香港节点网络不稳定
	if err != nil {
		// if nodeStore.FindWhiteList(&id) {
		// 	engine.Log.Info("Timeout receiving broadcast reply message:%s %s", id.B58String(), hex.EncodeToString(hash))
		// }
		return nil, errors.New("Timeout receiving broadcast reply message")
	}
	if bs == nil {
		// if nodeStore.FindWhiteList(&id) {
		// 	engine.Log.Info("Timeout receiving broadcast reply message:%s %s", id.B58String(), hex.EncodeToString(hash))
		// }
		// engine.Log.Warn("Timeout receiving broadcast reply message %s %s", id.B58String(), hex.EncodeToString(message.Body.Hash))
		// failNode = append(failNode, broadcasts[j])
		// continue
		return nil, errors.New("Timeout receiving broadcast reply message")
	}

	//验证同步到的消息
	mmp := new(go_protos.MessageMulticast)
	err = proto.Unmarshal(*bs, mmp)
	if err != nil {
		engine.Log.Error("proto unmarshal error %s", err.Error())
		return nil, err
	}
	// if nodeStore.FindWhiteList(&id) {
	// 	engine.Log.Info("sync multicast message success from:%s %s", id.B58String(), hex.EncodeToString(hash))
	// }
	return mmp, nil

	// head := MessageHead{
	// 		RecvId       :mmp
	// RecvSuperId   nodeStore.AddressNet          `json:"r_s_id"` //接收者的超级节点id
	// RecvVnode     virtual_node.AddressNetExtend `json:"r_v_id"` //接收者虚拟节点id
	// Sender        nodeStore.AddressNet          `json:"s_id"`   //发送者id
	// SenderSuperId nodeStore.AddressNet          `json:"s_s_id"` //发送者超级节点id
	// SenderVnode   virtual_node.AddressNetExtend `json:"s_v_id"` //发送者虚拟节点id
	// Accurate      bool                          `json:"a"`      //是否准确发送给一个节点，如果
	// }

	// return
	//将消息放数据库

}
