package core

//import (
//	"fmt"
//	"time"
//	gconfig "mandela/config"
//	"mandela/core/cache_store"
//	"mandela/core/config"
//	"mandela/core/message_center"
//	"mandela/core/nodeStore"
//	"mandela/core/utils"
//)

//func init() {
//	go readMulticastSyncName()
//	go findRootPeer()
//	go readFlashTempName()
//	go readFlashName()
//	go findNameSelf()
//	go registerNameAddress()
//	go readOutMulticastPKeyName()
//	go readOutFindName()
//	go readOutCloseConnName()
//}

///*
//	读取定时广播需要同步的域名
//*/
//func readMulticastSyncName() {
//	for name := range cache_store.OutMulticastName {
//		if ok, nameOne := cache_store.FindNameInForever(name); ok {
//			//			fmt.Println("广播一个域名", nameOne.Name)
//			message_center.SyncName(nameOne)
//			now := time.Now().Unix()
//			cache_store.AddSyncMulticastName(name, now+config.Time_name_sync_multicast)
//		}

//	}
//}

//func readFlashTempName() {
//	for flashName := range cache_store.OutFlashTempName {
//		fmt.Println("要确认构建的域名", flashName)
//	}

//}

//func readFlashName() {
//	for flashName := range cache_store.OutFlashName {
//		fmt.Println("要更新的域名", flashName)
//	}

//}

///*
//	定时查询根节点是否在线
//*/
//func findRootPeer() {
//	for {
//		time.Sleep(time.Hour)
//		FindRootPeer()
//	}
//}

///*
//	定时查找根节点域名，只在短时间内查询5次
//*/
//func findRootName() {
//	//查到了就不用查了
//	if cache_store.Root.Exist {
//		//		fmt.Println("11查到了，不用查了")
//		return
//	}
//	for i := 0; i < 5; i++ {
//		//查到了就不用查了
//		if cache_store.Root.Exist {
//			//			fmt.Println("22查到了，不用查了")
//			return
//		}
//		FindRootPeer()
//		time.Sleep(time.Second * 10)
//	}
//}

///*
//	查询根节点
//*/
//func FindRootPeer() {
//	//	fmt.Println("开始查找root域名")
//	idInt := utils.GetHashKey(config.C_root_name)

//	bs, _ := utils.Encode(idInt.Bytes(), utils.SHA1)
//	mh := utils.Multihash(bs)

//	ids := utils.GetLogicIds(&mh)
//	//	fmt.Println("1逻辑节点id个数", len(ids))

//	content := []byte(config.C_root_name)
//	for _, one := range ids {
//		mhead := message_center.NewMessageHead(one, one, false)
//		mbody := message_center.NewMessageBody(&content, "", nil, 0)
//		message := message_center.NewMessage(mhead, mbody)
//		if message.Send(gconfig.MSGID_find_name) {
//			continue
//		}

//		message.BuildReplyHash(message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)

//		ok, one := cache_store.FindNameInForever(config.C_root_name)
//		if !ok {
//			continue
//		}
//		if !cache_store.VoteAgree.Vote(config.VOTE_find_name, config.C_root_name, message.Body.ReplyHash.B58String()) {
//			continue
//		}
//		cache_store.AddAddressInName(config.C_root_name, one)
//		cache_store.Root = one
//	}

//	//	if nodeStore.NodeSelf.IsSuper {
//	//		for _, one := range ids {
//	//			message := message_center.Message{
//	//				RecvId:        one,
//	//				RecvSuperId:   one,                          //接收者的超级节点id
//	//				SenderSuperId: nodeStore.NodeSelf.IdInfo.Id, //发送者超级节点id
//	//				Sender:        nodeStore.NodeSelf.IdInfo.Id,
//	//				CreateTime:    utils.TimeFormatToNanosecond(),
//	//				Accurate:      false,
//	//				Content:       []byte(config.C_root_name),
//	//			}
//	//			message.BuildHash()
//	//			//		nearId := nodeStore.FindNearInSuper(recvId, []byte{}, false)
//	//			if !message_center.IsSendToOtherSuper(&message, message_center.MSGID_find_name, nil) {
//	//				//				fmt.Println("是自己的")
//	//				//				cache_store.AgreeTempName(name, message.ReplyHash)
//	//				message.ReplyTime = utils.TimeFormatToNanosecond()
//	//				message.Rand = utils.GetAccNumber()
//	//				message.BuildReplyHash()
//	//				ok, one := cache_store.FindNameInForever(config.C_root_name)
//	//				if !ok {
//	//					continue
//	//				}
//	//				if !cache_store.VoteAgree.Vote(config.VOTE_find_name, config.C_root_name, message.ReplyHash.B58String()) {
//	//					continue
//	//				}
//	//				cache_store.AddAddressInName(config.C_root_name, one)
//	//				cache_store.Root = one
//	//				//				for _, one := range one.Ids {
//	//				//					fmt.Println("111投票通过", one.PeerId.GetIdStr(), one.SuperPeerId.GetIdStr())
//	//				//				}

//	//			}
//	//		}
//	//	} else {
//	//		for _, one := range ids {
//	//			message := message_center.Message{
//	//				RecvId:        one,
//	//				RecvSuperId:   one,                          //接收者的超级节点id
//	//				SenderSuperId: nodeStore.SuperPeerId,        //发送者超级节点id
//	//				Sender:        nodeStore.NodeSelf.IdInfo.Id, //发送者节点id
//	//				CreateTime:    utils.TimeFormatToNanosecond(),
//	//				Accurate:      false,
//	//				Content:       []byte(config.C_root_name),
//	//			}
//	//			message.BuildHash()
//	//			if session, ok := engine.GetSession(nodeStore.SuperPeerId.B58String()); ok {
//	//				session.Send(message_center.MSGID_find_name, message.JSON(), false)
//	//			}
//	//		}
//	//	}
//}

///*
//	给域名添加一个地址
//*/
////func AddNameAdress() {
////	idInt := utils.GetHashKey(config.C_root_name)

////	ids := utils.GetLogicIds(idInt.Bytes())
////	fmt.Println("1逻辑节点id个数", len(ids))

////	if nodeStore.NodeSelf.IsSuper {
////		for _, one := range ids {
////			message := message_center.Message{
////				ReceSuperId:   one,                          //接收者的超级节点id
////				SenderSuperId: nodeStore.NodeSelf.IdInfo.Id, //发送者超级节点id
////				Accurate:      false,
////				Content:       []byte(config.C_root_name),
////			}
////			//		nearId := nodeStore.FindNearInSuper(recvId, []byte{}, false)
////			if !message_center.IsSendToOtherSuper(&message, message_center.MSGID_find_name, []byte{}) {
////				//				fmt.Println("是自己的")
////				//				cache_store.AgreeTempName(name, message.ReplyHash)
////			}
////		}
////	}
////}

///*
//	循环查找自己的域名
//*/
//func findNameSelf() {
//	for {
//		time.Sleep(time.Second * config.Time_find_name_self)

//		if cache_store.NameSelf != "." {
//			continue
//		}
//		if cache_store.NameSelf != "" {
//			break
//		}
//		if nodeStore.NodeSelf.IdInfo.Id == nil {
//			continue
//		}
//		//		fmt.Println("自己的域名", cache_store.NameSelf)
//		message_center.FindNameSelf()

//	}
//}

///*
//	循环注册自己域名的地址id
//*/
//func registerNameAddress() {
//	for {
//		time.Sleep(time.Second * config.Time_register_addr_name)
//		if cache_store.NameSelf == "" || cache_store.NameSelf == "." {
//			continue
//		}
//		message_center.AddNameAddress()
//	}
//}

///*
//	循环读取需要广播同步的公钥
//*/
//func readOutMulticastPKeyName() {
//	for key := range cache_store.OutMulticastPKeyName {
//		name, ok := cache_store.FindKeyName(key)
//		if !ok {
//			continue
//		}
//		message_center.SyncKey(key, name)
//		cache_store.AddSyncMulticastKey(key, time.Now().Unix()+config.Time_key_sync_multicast)
//	}
//}

///*
//	读取要查询的域名
//*/
//func readOutFindName() {
//	for name := range cache_store.OutFindName {
//		//		fmt.Println("查询域名前")
//		message_center.FindName(name)
//		//		fmt.Println("查询域名后")
//	}
//}

///*
//	读取需要询问关闭的连接名称
//*/
//func readOutCloseConnName() {
//	for name := range nodeStore.OutCloseConnName {
//		message_center.AskCloseConn(name.B58String())
//	}
//}
