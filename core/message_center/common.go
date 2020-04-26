package message_center

import (
	"mandela/core/config"
	"mandela/core/engine"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/virtual_node"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

var sendHash = new(sync.Map) //保存1分钟内的消息sendhash，用于判断重复消息
var sendhashTask = utils.NewTask(sendhashTaskFun)

func sendhashTaskFun(class string, params []byte) {
	sendHash.Delete(hex.EncodeToString(params))
}

/*
	检查这个消息是否发送过
*/
func CheckHash(sendhash []byte) bool {
	_, ok := sendHash.Load(hex.EncodeToString(sendhash))
	if !ok {
		sendHash.Store(hex.EncodeToString(sendhash), nil)
		sendhashTask.Add(time.Now().Unix()+60, "", sendhash)
	}
	return !ok
}

// var (
// 	task        = utils.NewTask(msgTimeOutProsess)
// 	msgHashLock = new(sync.RWMutex)
// 	msgHash     = make(map[string]int64)
// )

// /*
// 	添加一个消息超时
// */
// func addMsgTimeOut(md5 string) {
// 	now := time.Now().Unix()
// 	msgHashLock.Lock()
// 	msgHash[md5] = now
// 	msgHashLock.Unlock()
// 	task.Add(now+60*10, config.TSK_msg_timeout_remove, md5)
// }

// /*
// 	检查一个消息是否超时或者非法
// */
// func checkMsgTimeOut(md5 string) (ok bool) {
// 	msgHashLock.Lock()
// 	_, ok = msgHash[md5]
// 	if ok {
// 		delete(msgHash, md5)
// 	}
// 	msgHashLock.Unlock()
// 	return
// }

type MessageHead struct {
	RecvId        *nodeStore.AddressNet          `json:"r_id"`   //接收者id
	RecvSuperId   *nodeStore.AddressNet          `json:"r_s_id"` //接收者的超级节点id
	RecvVnode     *virtual_node.AddressNetExtend `json:"r_v_id"` //接收者虚拟节点id
	Sender        *nodeStore.AddressNet          `json:"s_id"`   //发送者id
	SenderSuperId *nodeStore.AddressNet          `json:"s_s_id"` //发送者超级节点id
	SenderVnode   *virtual_node.AddressNetExtend `json:"s_v_id"` //发送者虚拟节点id
	Accurate      bool                           `json:"a"`      //是否准确发送给一个节点，如果
	// VnodeAccurate bool                           `json:"v_a"`      //是否准确发送给一个虚拟节点
}

func NewMessageHead(recvid, recvSuperid *nodeStore.AddressNet, accurate bool) *MessageHead {
	if nodeStore.NodeSelf.IsSuper {
		//		head := NewMessageHead(nil, nil, nil, nodeStore.NodeSelf.IdInfo.Id, false)
		return &MessageHead{
			RecvId:        recvid,                        //接收者id
			RecvSuperId:   recvSuperid,                   //接收者的超级节点id
			Sender:        &nodeStore.NodeSelf.IdInfo.Id, //发送者id
			SenderSuperId: &nodeStore.NodeSelf.IdInfo.Id, //发送者超级节点id
			Accurate:      accurate,                      //是否准确发送给一个节点
		}
	} else {
		return &MessageHead{
			RecvId:        recvid,                        //接收者id
			RecvSuperId:   recvSuperid,                   //接收者的超级节点id
			Sender:        &nodeStore.NodeSelf.IdInfo.Id, //发送者id
			SenderSuperId: nodeStore.SuperPeerId,         //发送者超级节点id
			Accurate:      accurate,                      //是否准确发送给一个节点
		}
	}
}

/*
	创建一个虚拟节点消息
*/
func NewMessageHeadVnode(sendVid, recvVid *virtual_node.AddressNetExtend, accurate bool) *MessageHead {
	if nodeStore.NodeSelf.IsSuper {
		//		head := NewMessageHead(nil, nil, nil, nodeStore.NodeSelf.IdInfo.Id, false)
		return &MessageHead{
			// RecvId:        recvid,                        //接收者id
			// RecvSuperId:   recvSuperid,                   //接收者的超级节点id
			RecvVnode:     recvVid,                       //
			Sender:        &nodeStore.NodeSelf.IdInfo.Id, //发送者id
			SenderSuperId: &nodeStore.NodeSelf.IdInfo.Id, //发送者超级节点id
			SenderVnode:   sendVid,                       //
			Accurate:      accurate,                      //是否准确发送给一个节点
		}
	} else {
		return &MessageHead{
			// RecvId:        recvid,                        //接收者id
			// RecvSuperId:   recvSuperid,                   //接收者的超级节点id
			RecvVnode:     recvVid,                       //
			Sender:        &nodeStore.NodeSelf.IdInfo.Id, //发送者id
			SenderSuperId: nodeStore.SuperPeerId,         //发送者超级节点id
			SenderVnode:   sendVid,                       //
			Accurate:      accurate,                      //是否准确发送给一个节点
		}
	}
}

/*
	检查参数是否合法
*/
func (this *MessageHead) Check() bool {
	if this.RecvId == nil {
		return false
	}
	if this.RecvSuperId == nil {
		return false
	}
	if this.Sender == nil {
		return false
	}
	if this.SenderSuperId == nil {
		return false
	}
	return true
}

func (this *MessageHead) JSON() *[]byte {
	//	this.BuildReplyHash()
	bs, _ := json.Marshal(this)
	return &bs
}

type MessageBody struct {
	MessageId  uint64  `json:"m_id"`    //消息协议编号
	CreateTime string  `json:"c_time"`  //消息创建时间unix
	ReplyTime  string  `json:"r_time"`  //消息回复时间unix
	Hash       []byte  `json:"hash"`    //消息的hash值
	ReplyHash  []byte  `json:"r_hash"`  //回复消息的hash
	SendRand   uint64  `json:"s_rand"`  //发送随机数
	RecvRand   uint64  `json:"r_rand"`  //接收随机数
	Content    *[]byte `json:"content"` //发送的内容
}

func NewMessageBody(msgid uint64, content *[]byte, creatTime string, hash []byte, sendRand uint64) *MessageBody {
	return &MessageBody{
		MessageId:  msgid,
		CreateTime: creatTime,
		Hash:       hash,
		SendRand:   sendRand,
		Content:    content, //发送的内容
	}
}

func (this *MessageBody) JSON() *[]byte {
	//	this.BuildReplyHash()
	bs, _ := json.Marshal(this)
	return &bs
}

/*
	发送消息序列化对象
*/
type Message struct {
	msgid    uint64       //
	Head     *MessageHead `json:"head"` //
	Body     *MessageBody `json:"body"` //
	DataPlus *[]byte      `json:"dp"`   //body部分加密数据，消息路由时候不需要解密，临时保存
}

//type Message struct {
//	RecvId        *utils.Multihash `json:"recv_id"`         //接收者id
//	RecvSuperId   *utils.Multihash `json:"recv_super_id"`   //接收者的超级节点id
//	CreateTime    string           `json:"create_time"`     //消息创建时间unix
//	Sender        *utils.Multihash `json:"sender_id"`       //发送者id
//	SenderSuperId *utils.Multihash `json:"sender_super_id"` //发送者超级节点id
//	ReplyTime     string           `json:"reply_time"`      //消息回复时间unix
//	Hash          *utils.Multihash `json:"hash"`            //消息的hash值
//	ReplyHash     *utils.Multihash `json:"reply_hash"`      //回复消息的hash
//	Accurate      bool             `json:"accurate"`        //是否准确发送给一个节点
//	Content       []byte           `json:"content"`         //发送的内容
//	Rand          uint64           `json:"rand"`            //随机数
//}

func (this *Message) BuildHash() {
	this.Body.ReplyHash = nil
	this.Body.Hash = nil
	this.Body.SendRand = utils.GetAccNumber()
	this.Body.RecvRand = 0
	this.Body.CreateTime = utils.TimeFormatToNanosecond()
	bs, _ := json.Marshal(this)

	// hash := sha1.New()
	// hash.Write(bs)
	// mhBs, _ := utils.Encode(utils.Hash_SHA3_256(bs), gconfig.HashCode)
	// mh := utils.Multihash(mhBs)

	this.Body.Hash = utils.Hash_SHA3_256(bs)
	//	this.Hash = hex.EncodeToString(hash.Sum(nil))
}
func (this *Message) BuildReplyHash(createtime string, sendhash []byte, sendrand uint64) {
	this.Body.CreateTime = createtime
	this.Body.Hash = sendhash
	this.Body.SendRand = sendrand
	this.Body.ReplyHash = nil
	this.Body.RecvRand = utils.GetAccNumber()
	this.Body.ReplyTime = utils.TimeFormatToNanosecond()
	bs, _ := json.Marshal(this)
	// hash := sha1.New()
	// hash.Write(bs)
	// mhBs, _ := utils.Encode(utils.Hash_SHA3_256(bs), gconfig.HashCode)
	// mh := utils.Multihash(mhBs)
	this.Body.ReplyHash = utils.Hash_SHA3_256(bs)
	//	this.ReplyHash = hex.EncodeToString(hash.Sum(nil))
}

var debuf_msgid uint64 = 0

//var debuf_msgid uint64 = 1000
//var debuf_msgid uint64 = MSGID_TextMsg
//var debuf_msgid uint64 = gconfig.MSGID_findSuperID

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
	@safe 安全协议使用
*/
func (this *Message) Send(version uint64) (ok bool) {
	// defer fmt.Println("------------", version, this.Body.MessageId, this.Head)

	//虚拟节点之间的路由
	if this.Head.SenderVnode != nil && this.Head.RecvVnode != nil {
		this.Head.Sender = &nodeStore.NodeSelf.IdInfo.Id
		//收消息人就是自己
		if virtual_node.FindInVnodeSelf(*this.Head.RecvVnode) {
			return false
		}

		//Accurate参数是否发送给指定的某个虚拟节点
		//区分查找节点协议和点对点通讯协议
		var targetId virtual_node.AddressNetExtend
		if this.Head.Accurate {
			targetId = virtual_node.FindNearVnode(this.Head.RecvVnode, nil, false)
		} else {
			targetId = virtual_node.FindNearVnode(this.Head.RecvVnode, nil, true)
		}

		//没有可用的邻居节点
		if targetId == nil {
			fmt.Println("没有可用的邻居节点")
			return true
		}

		fmt.Println("打印地址", (targetId).B58String())
		vnodeinfo := virtual_node.FindVnodeinfo(targetId)
		if vnodeinfo == nil {
			fmt.Println("没有可用的邻居节点")
			return false
		}
		this.Head.RecvId = &vnodeinfo.Nid
		this.Head.RecvSuperId = &vnodeinfo.Nid

	}

	// fmt.Println("发送消息1", this.Head)
	return this.sendNormal(version)

}

/*
	发送给普通节点，最原始的消息
*/
func (this *Message) sendNormal(version uint64) bool {
	//安全协议不需buildhash
	this.BuildHash()
	if nodeStore.NodeSelf.IsSuper {
		// if version == debuf_msgid {
		// 	fmt.Println("-=-=- 111111111111")
		// }
		//收消息人就是自己
		// if nodeStore.NodeSelf.IdInfo.Id.B58String() == this.Head.RecvId.B58String() {
		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
			return false
		}
		// if version == debuf_msgid {
		// 	fmt.Println("-=-=- 333333333333333")
		// }
		//查找代理节点
		if _, ok := nodeStore.GetProxyNode(this.Head.RecvId.B58String()); ok {
			//发送给代理节点
			if session, ok := engine.GetSession(this.Head.RecvId.B58String()); ok {
				// if version == debuf_msgid {
				// 	fmt.Println("-=-=- 4444444444444")
				// }
				session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
			} else {
				// fmt.Println("这个代理节点的链接断开了")
			}
			return true
		}
		// if version == debuf_msgid {
		// 	fmt.Println("-=-=- 5555555")
		// }

		//		fmt.Println(string(*this.Head.JSON()))
		var targetId *nodeStore.AddressNet
		if this.Head.Accurate {
			targetId = nodeStore.FindNearInSuper(this.Head.RecvSuperId, nil, false)
		} else {
			targetId = nodeStore.FindNearInSuper(this.Head.RecvSuperId, nil, true)
		}
		// fmt.Println("本节点的其他超级节点", msgId, nodeStore.GetAllNodes(), targetId.B58String())
		if targetId == nil {
			//fmt.Println("没有可用的邻居节点")
			return true
		}
		// if version == debuf_msgid {
		// 	fmt.Println("-=-=- 666666666666")
		// }
		//收消息人就是自己
		// if nodeStore.NodeSelf.IdInfo.Id.B58String() == targetId.B58String() {
		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *targetId) {
			// if version == debuf_msgid {
			// 	fmt.Println("-=-=- 777777777777777")
			// }
			return false
		}

		// if version == debuf_msgid {
		// 	fmt.Println("-=-=- 88888888888888")
		// }

		//转发出去
		if session, ok := engine.GetSession(targetId.B58String()); ok {
			session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
			// if version == debuf_msgid {
			// 	fmt.Println("-=-=- 999999999999")
			// }
		} else {
			// fmt.Println("111 这个超级节点的链接断开了", msgId, targetId.B58String())
		}
		// if version == debuf_msgid {
		// 	fmt.Println("-=-=- 101010101010101010101010")
		// }
		return true

		//		return IsSendToOtherSuperToo(this.Head, this.Body.JSON(), msgId, nil)
	} else {
		// if version == debuf_msgid {
		// 	fmt.Println("-=-=- 22222222222")
		// }
		if nodeStore.SuperPeerId == nil {
			// fmt.Println("没有可用的超级节点")
			return true
		}
		if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
			session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
		} else {
			// fmt.Println("超级节点的session未找到")
		}
		return true
	}
}

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
	@safe 安全协议使用
*/
// func (this *Message) sendVnode(version uint64) bool {

// }

// /*
// 	强制发送消息给邻居节点
// 	用于节点发现，实现网络自治
// */
// func (this *Message) SendForce(msgId uint64) bool {
// 	this.BuildHash()
// 	//TODO 这里对消息加密

// 	// fmt.Println("本节点是否是超级节点", nodeStore.NodeSelf.IsSuper)

// 	if nodeStore.NodeSelf.IsSuper {
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("-=-=- 111111111111")
// 		// }
// 		//收消息人就是自己

// 		// if nodeStore.NodeSelf.IdInfo.Id.B58String() == this.Head.RecvId.B58String() {
// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
// 			return false
// 		}
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("-=-=- 333333333333333")
// 		// }
// 		//查找代理节点
// 		if _, ok := nodeStore.GetProxyNode(this.Head.RecvId.B58String()); ok {
// 			//发送给代理节点
// 			if session, ok := engine.GetSession(this.Head.RecvId.B58String()); ok {
// 				// if msgId == debuf_msgid {
// 				// 	fmt.Println("-=-=- 4444444444444")
// 				// }
// 				session.Send(msgId, this.Head.JSON(), this.Body.JSON(), false)
// 			} else {
// 				// fmt.Println("这个代理节点的链接断开了")
// 			}
// 			return true
// 		}
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("-=-=- 5555555")
// 		// }
// 		targetId := nodeStore.FindNearInSuper(this.Head.RecvSuperId, nil, false)
// 		if targetId == nil {
// 			// fmt.Println("没有可用的邻居节点")
// 			return true
// 		}
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("-=-=- 666666666666")
// 		// }
// 		//收消息人就是自己
// 		// if nodeStore.NodeSelf.IdInfo.Id.B58String() == targetId.B58String() {
// 		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *targetId) {
// 			// if msgId == debuf_msgid {
// 			// 	fmt.Println("-=-=- 777777777777777")
// 			// }
// 			return false
// 		}

// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("-=-=- 88888888888888")
// 		// }

// 		//转发出去
// 		if session, ok := engine.GetSession(targetId.B58String()); ok {
// 			session.Send(msgId, this.Head.JSON(), this.Body.JSON(), false)
// 			// if msgId == debuf_msgid {
// 			// 	fmt.Println("-=-=- 999999999999")
// 			// }
// 		} else {
// 			// fmt.Println("222 这个超级节点的链接断开了")
// 		}
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("-=-=- 101010101010101010101010")
// 		// }
// 		return true

// 		//		return IsSendToOtherSuperToo(this.Head, this.Body.JSON(), msgId, nil)
// 	} else {
// 		// if msgId == debuf_msgid {
// 		// 	fmt.Println("-=-=- 22222222222")
// 		// }
// 		if nodeStore.SuperPeerId == nil {
// 			// fmt.Println("没有可用的超级节点")
// 			return true
// 		}
// 		if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
// 			session.Send(msgId, this.Head.JSON(), this.Body.JSON(), false)
// 		} else {
// 			// fmt.Println("超级节点的session未找到")
// 		}
// 		return true
// 	}
// }

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
*/
func (this *Message) IsSendOther(form *nodeStore.AddressNet) bool {
	engine.Log.Info("打印消息1 %v", this.Body)
	//如果是虚拟节点之间的消息，则一定是指定某节点的
	// oldAccurate := this.Head.Accurate
	// if this.Head.SenderVnode != nil && this.Head.RecvVnode != nil {
	// 	this.Head.Accurate = true
	// }
	ok := IsSendToOtherSuperToo(this.Head, this.DataPlus, this.msgid, form)
	engine.Log.Info("打印消息2 %v", this.Body)
	//将messageHead.Accurate参数恢复
	// messageHead.Accurate = oldAccurate

	//发送给自己并且是虚拟节点之间的消息
	if !ok && this.Head.SenderVnode != nil && this.Head.RecvVnode != nil {
		if len(virtual_node.GetVnodeSelf()) <= 0 {
			return true
		}

		// this.Head.Sender = &nodeStore.NodeSelf.IdInfo.Id
		//收消息人就是自己
		if virtual_node.FindInVnodeSelf(*this.Head.RecvVnode) {
			return ok
		}

		//Accurate参数是否发送给指定的某个虚拟节点
		//区分查找节点协议和点对点通讯协议
		var targetId virtual_node.AddressNetExtend
		if this.Head.Accurate {
			targetId = virtual_node.FindNearVnode(this.Head.RecvVnode, nil, false)
		} else {
			targetId = virtual_node.FindNearVnode(this.Head.RecvVnode, nil, true)
		}
		//没有可用的邻居节点
		if targetId == nil {
			return true
		}
		engine.Log.Info("打印消息3 %v", this.Body)
		vnodeinfo := virtual_node.FindVnodeinfo(targetId)
		this.Head.RecvId = &vnodeinfo.Nid
		this.Head.RecvSuperId = &vnodeinfo.Nid
		// bs :=
		if this.DataPlus == nil {
			SendVnodeP2pMsgHE(this.Body.MessageId, this.Head.SenderVnode, this.Head.RecvVnode, nil)
		} else {

			fmt.Println(this.Body.MessageId)
			fmt.Println(this.Head.SenderVnode)
			fmt.Println(this.Head.RecvVnode)
			fmt.Println(this.DataPlus)
			SendVnodeP2pMsgHE(this.Body.MessageId, this.Head.SenderVnode, this.Head.RecvVnode, this.DataPlus)
		}
		return true
	}
	return ok
}

/*
	解析内容
*/
func (this *Message) ParserContent() error {
	//TODO 解密内容

	this.Body = new(MessageBody)
	// err := json.Unmarshal(*this.DataPlus, this.Body)
	decoder := json.NewDecoder(bytes.NewBuffer(*this.DataPlus))
	decoder.UseNumber()
	err := decoder.Decode(this.Body)
	if err != nil {
		return err
	}
	return nil
}

/*
	验证hash
*/
func (this *Message) CheckSendhash() bool {
	//TODO 验证sendhash是否正确
	//TODO 验证时间不能相差太远

	//验证sendhash是否已经接受过此消息
	return CheckHash(this.Body.Hash)
}

/*
	验证hash
*/
func (this *Message) CheckReplyhash() bool {
	//TODO 验证replyhash是否正确
	//TODO 验证时间不能相差太远

	//验证replyhash是否已经接受过此消息
	return CheckHash(this.Body.ReplyHash)
}

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
*/
func (this *Message) Reply(version uint64) bool {
	this.BuildReplyHash(this.Body.CreateTime, this.Body.Hash, this.Body.SendRand)
	//TODO 这里对消息加密

	if nodeStore.NodeSelf.IsSuper {
		return IsSendToOtherSuperToo(this.Head, this.Body.JSON(), version, nil)
	} else {
		if nodeStore.SuperPeerId == nil {
			// fmt.Println("没有可用的超级节点")
			return true
		}
		if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
			session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
		} else {
			// fmt.Println("超级节点的session未找到")
		}
		return true
	}
}

func NewMessage(head *MessageHead, body *MessageBody) *Message {
	return &Message{
		Head: head,
		Body: body,
	}
}

func ParserMessage(data, dataplus *[]byte, msgId uint64) (*Message, error) {
	head := new(MessageHead)
	// err := json.Unmarshal(*data, head)
	decoder := json.NewDecoder(bytes.NewBuffer(*data))
	decoder.UseNumber()
	err := decoder.Decode(head)
	if err != nil {
		return nil, err
	}

	msg := Message{
		msgid:    msgId,
		Head:     head,
		DataPlus: dataplus,
	}
	return &msg, nil
}

/*
	得到一条消息的hash值
*/
//func GetHash(msg *Message) string {
//	hash := sha256.New()
//	hash.Write([]byte(msg.RecvId))
//	//	binary.Write(hash, binary.BigEndian, uint64(msg.ProtoId))
//	binary.Write(hash, binary.BigEndian, msg.CreateTime)
//	// hash.Write([]byte(int64(msg.ProtoId)))
//	// hash.Write([]byte(msg.CreateTime))
//	hash.Write([]byte(msg.Sender))
//	// hash.Write([]byte(msg.RecvTime))
//	binary.Write(hash, binary.BigEndian, msg.ReplyTime)
//	hash.Write(msg.Content)
//	hash.Write([]byte(msg.ReplyHash))
//	return hex.EncodeToString(hash.Sum(nil))
//}

/*
	消息超时删除md5
*/
func msgTimeOutProsess(class, params string) {
	switch class {
	case config.TSK_msg_timeout_remove: //删除超时的消息md5
		//		fmt.Println("开始删除临时域名", tempName)
		//		tempNameLock.Lock()
		//		delete(tempName, params)
		//		tempNameLock.Unlock()
		//		fmt.Println("删除了这个临时域名", params, tempName)
	default:
		//		//剩下是需要更新的域名
		//		flashName := FlashName{
		//			Name:  params,
		//			Class: class,
		//		}
		//		OutFlashName <- &flashName
	}

}
