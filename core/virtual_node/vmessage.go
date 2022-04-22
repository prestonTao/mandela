package virtual_node

// import (
// 	"mandela/config"
// 	"mandela/core/nodeStore"
// 	"mandela/core/utils"

// 	jsoniter "github.com/json-iterator/go"
// )

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

// type VMessage struct {
// 	msgid    uint64        //
// 	Head     *VMessageHead `json:"head"` //
// 	Body     *VMessageBody `json:"body"` //
// 	DataPlus *[]byte       `json:"dp"`   //body部分加密数据，消息路由时候不需要解密，临时保存
// }

// /*

//  */
// type VMessageHead struct {
// 	RecvId        *nodeStore.AddressNet `json:"r_id"`     //接收者id
// 	RecvSuperId   *nodeStore.AddressNet `json:"r_s_id"`   //接收者的超级节点id
// 	RecvVnode     *AddressNetExtend     `json:"r_s_id"`   //接收者虚拟节点id
// 	Sender        *nodeStore.AddressNet `json:"s_id"`     //发送者id
// 	SenderSuperId *nodeStore.AddressNet `json:"s_s_id"`   //发送者超级节点id
// 	SenderVnode   *AddressNetExtend     `json:"s_v_id"`   //
// 	Accurate      bool                  `json:"accurate"` //是否准确发送给一个节点
// }

// type VMessageBody struct {
// 	MessageId  uint64           `json:"m_id"`    //消息协议编号
// 	CreateTime string           `json:"c_time"`  //消息创建时间unix
// 	ReplyTime  string           `json:"r_time"`  //消息回复时间unix
// 	Hash       *utils.Multihash `json:"hash"`    //消息的hash值
// 	ReplyHash  *utils.Multihash `json:"r_hash"`  //回复消息的hash
// 	SendRand   uint64           `json:"s_rand"`  //发送随机数
// 	RecvRand   uint64           `json:"r_rand"`  //接收随机数
// 	Content    *[]byte          `json:"content"` //发送的内容
// }

// // /*
// // 	检查该消息是否是自己的
// // 	不是自己的则自动转发出去
// // 	@return    bool     是否是自己的
// // 	@return    error    是否发送成功
// // */
// // func (this *VMessage) SendMsg(vid AddressNetExtend) (bool, error) {
// // 	//收消息人就是自己
// // 	if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *this.Head.RecvId) {
// // 		return false, nil
// // 	}
// // 	//
// // 	return false, nil
// // }

// /*
// 	检查该消息是否是自己的
// 	不是自己的则自动转发出去
// 	@return    bool     是否是自己的
// 	@return    error    是否发送成功
// */
// func (this *VMessage) Send(version uint64) (bool, error) {
// 	// //安全协议不需buildhash
// 	// this.BuildHash()
// 	// // if nodeStore.NodeSelf.IsSuper {
// 	// // 	// if version == debuf_msgid {
// 	// // 	// 	fmt.Println("-=-=- 111111111111")
// 	// // 	// }
// 	// //收消息人就是自己
// 	// if FindInVnodeSelf(this.Head.RecvVnode) {
// 	// 	return false, nil
// 	// }

// 	// // targetId := FindNearVnode(this.Head.RecvVnode, nil)

// 	// var targetId *AddressNetExtend
// 	// if this.Head.Accurate {
// 	// 	targetId = FindNearVnode(this.Head.RecvVnode, nil, false)
// 	// } else {
// 	// 	targetId = FindNearVnode(this.Head.RecvVnode, nil, true)
// 	// }
// 	// //没有可用的邻居节点
// 	// if targetId == nil {
// 	// 	//fmt.Println("没有可用的邻居节点")
// 	// 	return true
// 	// }

// 	// //收消息人就是自己
// 	// if FindInVnodeSelf(targetId) {
// 	// 	return false, nil
// 	// }

// 	// vnodeinfo := FindVnodeinfo(*targetId)

// 	// // if this.Head.Accurate {
// 	// // 	message_center.SendP2pMsgHE(this.msgid, this.)

// 	// // } else {

// 	// // }

// 	// //转发出去
// 	// if session, ok := engine.GetSession(vnodeinfo.Nid.B58String()); ok {
// 	// 	session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
// 	// 	// if version == debuf_msgid {
// 	// 	// 	fmt.Println("-=-=- 999999999999")
// 	// 	// }
// 	// } else {
// 	// 	// fmt.Println("111 这个超级节点的链接断开了", msgId, targetId.B58String())
// 	// }
// 	// // if version == debuf_msgid {
// 	// // 	fmt.Println("-=-=- 101010101010101010101010")
// 	// // }
// 	return true, nil

// 	// 	// if version == debuf_msgid {
// 	// 	// 	fmt.Println("-=-=- 333333333333333")
// 	// 	// }
// 	// 	//查找代理节点
// 	// 	if _, ok := nodeStore.GetProxyNode(this.Head.RecvId.B58String()); ok {
// 	// 		//发送给代理节点
// 	// 		if session, ok := engine.GetSession(this.Head.RecvId.B58String()); ok {
// 	// 			// if version == debuf_msgid {
// 	// 			// 	fmt.Println("-=-=- 4444444444444")
// 	// 			// }
// 	// 			session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
// 	// 		} else {
// 	// 			// fmt.Println("这个代理节点的链接断开了")
// 	// 		}
// 	// 		return true
// 	// 	}
// 	// 	// if version == debuf_msgid {
// 	// 	// 	fmt.Println("-=-=- 5555555")
// 	// 	// }

// 	// 	//		fmt.Println(string(*this.Head.JSON()))
// 	// 	var targetId *nodeStore.AddressNet
// 	// 	if this.Head.Accurate {
// 	// 		targetId = nodeStore.FindNearInSuper(this.Head.RecvSuperId, nil, false)
// 	// 	} else {
// 	// 		targetId = nodeStore.FindNearInSuper(this.Head.RecvSuperId, nil, true)
// 	// 	}
// 	// 	// fmt.Println("本节点的其他超级节点", msgId, nodeStore.GetAllNodes(), targetId.B58String())
// 	// 	if targetId == nil {
// 	// 		//fmt.Println("没有可用的邻居节点")
// 	// 		return true
// 	// 	}
// 	// 	// if version == debuf_msgid {
// 	// 	// 	fmt.Println("-=-=- 666666666666")
// 	// 	// }
// 	// 	//收消息人就是自己
// 	// 	// if nodeStore.NodeSelf.IdInfo.Id.B58String() == targetId.B58String() {
// 	// 	if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *targetId) {
// 	// 		// if version == debuf_msgid {
// 	// 		// 	fmt.Println("-=-=- 777777777777777")
// 	// 		// }
// 	// 		return false
// 	// 	}

// 	// 	// if version == debuf_msgid {
// 	// 	// 	fmt.Println("-=-=- 88888888888888")
// 	// 	// }

// 	// 	//转发出去
// 	// 	if session, ok := engine.GetSession(targetId.B58String()); ok {
// 	// 		session.Send(version, this.Head.JSON(), this.Body.JSON(), false)
// 	// 		// if version == debuf_msgid {
// 	// 		// 	fmt.Println("-=-=- 999999999999")
// 	// 		// }
// 	// 	} else {
// 	// 		// fmt.Println("111 这个超级节点的链接断开了", msgId, targetId.B58String())
// 	// 	}
// 	// 	// if version == debuf_msgid {
// 	// 	// 	fmt.Println("-=-=- 101010101010101010101010")
// 	// 	// }
// 	// 	return true

// 	// 	//		return IsSendToOtherSuperToo(this.Head, this.Body.JSON(), msgId, nil)
// 	// }

// 	return true, nil
// }

// func (this *VMessage) BuildHash() {
// 	this.Body.ReplyHash = nil
// 	this.Body.Hash = nil
// 	this.Body.SendRand = utils.GetAccNumber()
// 	this.Body.RecvRand = 0
// 	this.Body.CreateTime = utils.TimeFormatToNanosecond()
// 	bs, _ := json.Marshal(this)
// 	// hash := sha1.New()
// 	// hash.Write(bs)
// 	mhBs, _ := utils.Encode(utils.Hash_SHA3_256(bs), config.HashCode)
// 	mh := utils.Multihash(mhBs)
// 	this.Body.Hash = &mh
// 	//	this.Hash = hex.EncodeToString(hash.Sum(nil))
// }
// func (this *VMessage) BuildReplyHash(createtime string, sendhash *utils.Multihash, sendrand uint64) {
// 	this.Body.CreateTime = createtime
// 	this.Body.Hash = sendhash
// 	this.Body.SendRand = sendrand
// 	this.Body.ReplyHash = nil
// 	this.Body.RecvRand = utils.GetAccNumber()
// 	this.Body.ReplyTime = utils.TimeFormatToNanosecond()
// 	bs, _ := json.Marshal(this)
// 	// hash := sha1.New()
// 	// hash.Write(bs)
// 	mhBs, _ := utils.Encode(utils.Hash_SHA3_256(bs), config.HashCode)
// 	mh := utils.Multihash(mhBs)
// 	this.Body.ReplyHash = &mh
// 	//	this.ReplyHash = hex.EncodeToString(hash.Sum(nil))
// }
