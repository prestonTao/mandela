package sharebox

import (
	"mandela/core/utils"
	"bytes"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Task_class_share_self_fileinfo      = "Task_class_share_self_fileinfo"      //定时发送自己上传的文件索引信息
	Task_class_share_local_fileinfo     = "Task_class_share_local_fileinfo"     //定时发送自己上传的文件索引信息
	Task_class_net_fileinfo_remove_user = "Task_class_net_fileinfo_remove_user" //定时删除掉线的用户
//	Task_class_share_local_fileinfo_remove = "Task_class_sharefileinfo_remove" //定时删除掉线的节点
)

var task *utils.Task

func initTask() {
	task = utils.NewTask(taskFunc)

}

func taskFunc(class string, params []byte) {
	switch class {
	case Task_class_share_self_fileinfo:
		// //		fmt.Println("该更新该索引", params)
		// go func() {
		// 	fi := FindFileinfoToSelf(params)
		// 	if fi == nil {
		// 		fmt.Println("这个文件不存在了")
		// 		return
		// 	}
		// 	UpNetFileinfo(fi)
		// 	//			addShareFileinfo_task(params)
		// 	//task.Add(time.Now().Unix()+Time_sharefile, Task_class_share_self_fileinfo, params)
		// }()

	case Task_class_share_local_fileinfo:
		// go func() {
		// 	fi := FindFileinfoToLocal(params)
		// 	//			fi := FindFileinfoToLocal(params)
		// 	if fi == nil {
		// 		fmt.Println("这个文件不存在了")
		// 		return
		// 	}
		// 	UpNetFileinfo(fi)
		// 	//task.Add(time.Now().Unix()+Time_sharefile, Task_class_share_local_fileinfo, params)
		// }()
	case Task_class_net_fileinfo_remove_user: //定时删除掉线的用户
		go func() {
			cofs := ParseCheckOnlineFileShare(params)
			if cofs == nil {
				// fmt.Println("定时删除掉线的节点，解析JSON失败")
				return
			}
			//			fmt.Println("触发清理下线用户方法=====", cofs.FileHash, string(cofs.JSON()))
			fi := FindFileindexToNet(cofs.FileHash)
			if fi == nil {
				return
			}
			fi.Clear()
			// chunk := fi.FindChunk(cofs.ChunkHash)
			// if chunk == nil {
			// 	return
			// }
			// chunk.Clear()
			//			fmt.Println("触发清理下线用户方法=====完成")
		}()
		//	case Task_class_sharefileinfo_remove: //定时删除掉线的节点
		//		go func() {
		//			cofs := ParseCheckOnlineFileShare(params)
		//			if cofs == nil {
		//				fmt.Println("定时删除掉线的节点，解析JSON失败")
		//				return
		//			}
		//			fi := FindFileinfo(cofs.FileHash)
		//			chunk := fi.FindChunk(cofs.ChunkHash)
		//			if chunk == nil {
		//				return
		//			}

		//		}()

	default:
		// fmt.Println("未注册的定时器类型", class)

	}
}

type CheckOnlineFileShare struct {
	//	Chunk    int    //块
	FileHash  string //文件hash值
	ChunkHash string //块hash值
	User      string //共享用户
}

func (this *CheckOnlineFileShare) JSON() []byte {
	bs, err := json.Marshal(this)
	if err != nil {
		return bs
	}
	return bs
}

func NewCheckOnlineFileShare(fhash, chash, user string) *CheckOnlineFileShare {
	return &CheckOnlineFileShare{
		FileHash:  fhash,
		ChunkHash: chash,
		User:      user,
	}
}

func ParseCheckOnlineFileShare(bs []byte) *CheckOnlineFileShare {
	share := new(CheckOnlineFileShare)
	// err := json.Unmarshal([]byte(bs), share)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(share)
	if err != nil {
		return nil
	}
	return share
}

//func addShareFileinfo_task(hash string) error {
//	if nodeStore.NodeSelf.IdInfo.Id == nil {
//		return errors.New("本节点idinfo为空")
//	}

//	fi := FindFileinfoToLocal(hash)
//	if fi == nil {
//		return errors.New("本地未找到文件索引")
//	}

//	//每个块添加自己的共享
//	for _, one := range fi.FileChunk {
//		one.AddUpdateUser(nodeStore.NodeSelf.IdInfo.Id)
//	}

//	//	fi := NewFileInfo(hash)

//	//
//	mh, err := utils.FromB58String(hash)
//	if err != nil {
//		return err
//	}

//	//	fileinfoHash, err := hex.DecodeString(hash)
//	//	if err != nil {
//	//		return err
//	//	}

//	mhead := mc.NewMessageHead(&mh, &mh, false)
//	content := fi.JSON()
//	mbody := mc.NewMessageBody(&content, "", nil, 0)
//	message := mc.NewMessage(mhead, mbody)
//	if message.Send(MSGID_addFileShare) {
//		fmt.Println("发给其他小伙伴了")
//		bs := mc.WaitRequest(mc.CLASS_sharefile, message.Body.Hash.B58String())
//		if bs == nil {
//			fmt.Println("发送共享文件消息失败，可能超时")
//			return errors.New("发送共享文件消息失败，可能超时")
//		}
//		return nil
//	}

//	//	message := mc.Message{
//	//		RecvId:        &mh,
//	//		RecvSuperId:   &mh,                          //接收者的超级节点id
//	//		SenderSuperId: nodeStore.NodeSelf.IdInfo.Id, //发送者超级节点id
//	//		Sender:        nodeStore.NodeSelf.IdInfo.Id,
//	//		CreateTime:    utils.TimeFormatToNanosecond(),
//	//		Accurate:      false,
//	//		Content:       fi.JSON(),
//	//	}
//	//	if !nodeStore.NodeSelf.IsSuper {
//	//		message.SenderSuperId = nodeStore.SuperPeerId
//	//	}
//	//	message.BuildHash()
//	//	//		nearId := nodeStore.FindNearInSuper(recvId, []byte{}, false)
//	//	if mc.IsSendToOtherSuper(&message, MSGID_addFileShare, nil) {
//	//		fmt.Println("发给其他小伙伴了")
//	//		bs := mc.WaitRequest(mc.CLASS_sharefile, message.Hash.B58String())
//	//		if bs == nil {
//	//			fmt.Println("发送共享文件消息失败，可能超时")
//	//			return errors.New("发送共享文件消息失败，可能超时")
//	//		}
//	//		return nil
//	//	}
//	fmt.Println("自己保存")

//	//判断本地是否存在文件，若不存在则添加
//	filocal := FindFileinfoToNet(fi.Hash.B58String())
//	if filocal == nil {
//		//添加文件
//		AddFileinfoToNet(fi, true)
//	} else {
//		//文件中添加共享用户
//		for _, one := range fi.FileChunk {
//			filocal.AddShareUser(one.No, message.Head.Sender)
//		}
//	}
//	return nil
//}
