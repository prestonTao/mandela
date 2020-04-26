package message_center

import (
	"mandela/core/engine"
	"mandela/core/nodeStore"
	"bytes"
	"fmt"
)

/*
	检查该消息是否是自己的
	不是自己的则自动转发出去
	@return    sendOk    bool    是否发送给其他人。true=发送给其他人了;false=自己的消息;
*/
func IsSendToOtherSuperToo(messageHead *MessageHead, dataplus *[]byte, version uint64, form *nodeStore.AddressNet) (sendOk bool) {
	// fmt.Println("接收到的消息头1", messageHead, string(*dataplus))

	//如果是虚拟节点之间的消息，则一定是指定某节点的
	oldAccurate := messageHead.Accurate
	if messageHead.SenderVnode != nil && messageHead.RecvVnode != nil {
		messageHead.Accurate = true
	}

	updateAccurate := func() {
		//将messageHead.Accurate参数恢复
		messageHead.Accurate = oldAccurate
	}

	//------------------

	recvSuperId := messageHead.RecvSuperId
	recvId := messageHead.RecvId

	if !nodeStore.NodeSelf.IsSuper {
		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *recvId) {
			sendOk = false
			return
		} else {
			if messageHead.Accurate {
				//发错节点了
				fmt.Println("发错节点了", nodeStore.NodeSelf.IdInfo.Id.B58String(), recvSuperId.B58String(), recvId.B58String())
				return true
			} else {
				if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
					updateAccurate()
					session.Send(version, messageHead.JSON(), dataplus, false)
				}
				if version == debuf_msgid {
					fmt.Println("发送给超级节点")
				}
				return true
			}
		}
	}

	if version == debuf_msgid {
		//fmt.Println("----000000000", recvId.B58String(), recvSuperId.B58String(), nodeStore.NodeSelf.IdInfo.Id.B58String())
		// fmt.Println("----000000000", form.B58String())
	}
	if bytes.Equal(*recvId, *recvSuperId) {
		if bytes.Equal(*recvId, nodeStore.NodeSelf.IdInfo.Id) {
			sendOk = false
			return
		} else {
			if version == debuf_msgid {
				//fmt.Println("----1111111")
			}
			targetId := nodeStore.FindNearInSuper(messageHead.RecvSuperId, form, true)
			if version == debuf_msgid {
				//fmt.Println("----222222222", targetId.B58String(), nodeStore.NodeSelf.IdInfo.Id.B58String())
				// fmt.Println("----222222222", form.B58String())
			}
			if bytes.Equal(*targetId, nodeStore.NodeSelf.IdInfo.Id) {
				//查找代理节点
				_, ok := nodeStore.GetProxyNode(recvId.B58String())
				if version == debuf_msgid {
					//fmt.Println("----333333333")
				}
				if ok {
					//发送给代理节点
					if session, ok := engine.GetSession(recvId.B58String()); ok {
						if version == debuf_msgid {
							fmt.Println("发送出去了111")
						}
						updateAccurate()
						session.Send(version, messageHead.JSON(), dataplus, false)
					} else {
						//这个链接断开了
						if version == debuf_msgid {
							fmt.Println("这个链接断开了")
						}
					}
				} else {
					if !messageHead.Accurate {
						sendOk = false
						return
					}

					if version == debuf_msgid {
						fmt.Println("该节点不在线")
					}
					// //该节点不在线了
					// if msgId == debuf_msgid {
					// 	fmt.Println("111111111", recvId, recvSuperId)
					// }
				}
				return true
			}

			session, ok := engine.GetSession(targetId.B58String())
			if ok {
				if version == debuf_msgid {
					//fmt.Println("发送出去了222")
				}
				updateAccurate()
				session.Send(version, messageHead.JSON(), dataplus, false)
			} else {
				if version == debuf_msgid {
					fmt.Println("这个链接断开了222")
				}
				// if msgId == debuf_msgid {
				// 	fmt.Println("-=-=-=-= 这个session已经断开")
				// }
			}
		}
		if version == debuf_msgid {
			//fmt.Println("4444444444", recvId, recvSuperId)
		}
		return true

	} else {

		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, *recvSuperId) {
			if recvId == nil {
				sendOk = false
				return
			} else {
				if version == debuf_msgid {
					//fmt.Println("----444444444")
				}
				_, ok := nodeStore.GetProxyNode(recvId.B58String())
				if version == debuf_msgid {
					//fmt.Println("----555555555")
				}
				if ok {
					if session, ok := engine.GetSession(recvId.B58String()); ok {
						if version == debuf_msgid {
							//fmt.Println("发送出去了")
						}
						updateAccurate()
						session.Send(version, messageHead.JSON(), dataplus, false)
					} else {
						if version == debuf_msgid {
							fmt.Println("这个session不存在")
						}
					}
				}
				//代理节点转移或下线，忽略这个消息
				if version == debuf_msgid {
					//fmt.Println("22222222")
				}
				return true
			}
		}
		if version == debuf_msgid {
			//fmt.Println("----6666666666")
		}
		targetId := nodeStore.FindNearInSuper(messageHead.RecvSuperId, form, true)
		if version == debuf_msgid {
			//fmt.Println("----777777777777")
		}
		// hex.EncodeToString(targetId) == nodeStore.NodeSelf.IdInfo.Id.GetIdStr()
		if bytes.Equal(*targetId, nodeStore.NodeSelf.IdInfo.Id) {
			if messageHead.Accurate {
				//该节点不在线
				// fmt.Println("该节点不在线，这个包会被丢弃", msgId, targetId.B58String(),
				// 	messageHead.RecvSuperId.B58String(), string(*messageHead.JSON()))
				if version == debuf_msgid {
					//fmt.Println("33333333")
				}
				return true
			} else {
				sendOk = false
				return
			}
		}

		session, ok := engine.GetSession(targetId.B58String())
		if ok {
			updateAccurate()
			session.Send(version, messageHead.JSON(), dataplus, false)
		}
		if version == debuf_msgid {
			//fmt.Println("5555555555", recvId, recvSuperId)
		}
		return true
	}

}
