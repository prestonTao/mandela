package sharebox

import (
	"mandela/config"
	"mandela/core/message_center"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	sconfig "mandela/sharebox/config"
	sqldb "mandela/sqlite3_db"
	"errors"
	"os"
	"path/filepath"
	"time"
)

/*
	添加本地共享文件夹
*/
func AddLocalShareFolders(folders ...string) error {
	// fmt.Println("添加共享文件夹", folders)
	for _, one := range folders {
		if one == "" {
			return errors.New("文件夹路径为空")
		}
		fileinfo, err := os.Stat(one)
		if err != nil {
			return err
		}
		if !fileinfo.IsDir() {
			return errors.New("该路径不是一个文件夹")
		}
		//判断是否重复
		for _, two := range rootDir.GetDirs() {
			if two.Path == one {
				//路径相同
				return nil
			}
		}

		//添加文件监听
		err = WatcherFolder(one)
		if err != nil {
			return err
		}
		//遍历所有文件，将文件添加到共享列表
		dir := NewDir(one)
		rootDir.AddDir(dir)
		new(sqldb.ShareFolder).Add(one)

		go listFile(one, dir)

	}
	return nil
}

/*
	删除本地共享文件夹
*/
func DelLocalShareFolders(folders ...string) error {
	for _, one := range folders {
		if one == "" {
			return errors.New("文件夹路径为空")
		}
		fileinfo, err := os.Stat(one)
		if err != nil {
			return err
		}
		if !fileinfo.IsDir() {
			return errors.New("该路径不是一个文件夹")
		}
		//判断是否存在管理目录中
		for _, two := range rootDir.GetDirs() {
			if two.Path == one {

				//删除文件监听
				err = DelWatcherFolder(one)
				if err != nil {
					return err
				}
				rootDir.RemoveFile(one)
				new(sqldb.ShareFolder).Del(one)
				return nil
			}
		}
		return errors.New("该文件夹不在管理目录中")
	}
	return nil
}

/*
	获取共享文件夹根目录列表
*/
// func GetFolderRoots() []string {
// 	for _, one := range rootDir.GetDirs() {
// 		one.Name
// 	}
// 	return GetShareFolderRoots()
// }

/*
	获取一个根目录下的子目录，详细文件列表
*/
func GetShareFolderRootsDetail() DirVO {
	return *rootDir.conversionVO()
}

/*
	查询远端节点共享目录列表
*/
func GetRemoteShareFolderDetail(id string) (*DirVO, error) {
	addrNet := nodeStore.AddressFromB58String(id)

	// mhead := message_center.NewMessageHead(&addrNet, &addrNet, true)
	// mbody := message_center.NewMessageBody(nil, "", nil, 0)
	// message := message_center.NewMessage(mhead, mbody)
	// if message.Send(MSGID_getsharefolderlist) {

	if message, ok, _ := message_center.SendP2pMsg(config.MSGID_sharebox_getsharefolderlist, &addrNet, nil); ok {
		// fmt.Println("发给其他小伙伴了----")
		// bs := flood.WaitRequest(mc.CLASS_getRemoteFolderList, hex.EncodeToString(message.Body.Hash), 0)
		bs, _ := flood.WaitRequest(mc.CLASS_getRemoteFolderList, utils.Bytes2string(message.Body.Hash), 0)
		// fmt.Println("有消息返回了啊")
		if bs == nil {
			// fmt.Println("发送共享文件消息失败，可能超时")
			return nil, errors.New("发送获取远端节点共享文件列表消息失败，可能超时")
		}
		//		fmt.Println("添加文件共享成功")
		// return nil
		dirvo := ParseDirVO(bs)
		return dirvo, nil
	}
	return nil, errors.New("发送获取远端节点共享文件列表消息失败")
}

//同步文件索引到网络中(同时开启定时同步)
func UpNetFileindex(fi *FileIndex) error {
	err := UpNetFileindexAction(fi) //第一次加入到网络中
	if err != nil {
		return err
	}
	//开始定时同步
	go func(fi *FileIndex) {
		for range time.NewTicker(sconfig.Time_sharefile * time.Second).C {
			UpNetFileindexAction(fi)
		}
	}(fi)
	return nil
}

/*
	共享本节点
	把文件索引上传到网络中去，并且增加本节点共享
*/
func UpNetFileindexAction(fi *FileIndex) error {
	//fmt.Println("1111111111111")
	//添加自己为共享节点
	fi.AddShareUser(&nodeStore.NodeSelf.IdInfo.Id)
	//fmt.Println("222222222222222")
	recvId := fi.Hash
	content := fi.JSON()
	message, ok := mc.SendSearchAllMsg(config.MSGID_sharebox_addFileShare, recvId, &content)
	if ok {

		// }

		// mhead := mc.NewMessageHead(recvId, recvId, false)
		// // fmt.Println("3333333333333333")
		// mbody := mc.NewMessageBody(&content, "", nil, 0)
		// message := mc.NewMessage(mhead, mbody)
		// // fmt.Println("44444444444444444444")
		// if message.Send(MSGID_addFileShare) {
		//fmt.Println("发给其他小伙伴了----")
		// bs := flood.WaitRequest(mc.CLASS_sharefile, hex.EncodeToString(message.Body.Hash), 0)
		bs, _ := flood.WaitRequest(mc.CLASS_sharefile, utils.Bytes2string(message.Body.Hash), 0)
		//fmt.Println("有消息返回了啊")
		if bs == nil {
			// fmt.Println("发送共享文件消息失败，可能超时")
			return errors.New("发送共享文件消息失败，可能超时")
		}
		//		fmt.Println("添加文件共享成功")
		return nil
	}
	//这个文件索引归自己管理
	// fmt.Println("这个文件索引归自己管理")
	AddFileindexToNet(fi, true)
	//同步文件到最近的邻居节点
	//go SyncFiletoNearId(fi)
	// fmt.Println("===== 9999999999999")
	return nil
}

/*
	网络中查找一个文件信息
*/
func DownloadFileindexOpt(hash string) (fi *FileIndex, err error) {
	err = ErrNotFindCode
	// mh, errs := utils.FromB58String(hash)
	// if errs != nil {
	// 	err = errs
	// 	return
	// }
	mh := nodeStore.AddressFromB58String(hash)
	ids := getQuarterLogicIds(&mh)
	for _, id := range ids {
		content := []byte(mh)

		message, ok := mc.SendSearchAllMsg(config.MSGID_sharebox_findFileinfo, id, &content)
		if ok {

			// }

			// mhead := mc.NewMessageHead(id, id, false)
			// mbody := mc.NewMessageBody(&content, "", nil, 0)
			// message := mc.NewMessage(mhead, mbody)
			// if message.Send(MSGID_findFileinfo) {
			// fmt.Println("开始等待查找返回", message)
			// fmt.Println("开始等待查找返回", message.Body)
			// fmt.Println("开始等待查找返回", message.Body.Hash)
			// fmt.Println("开始等待查找返回", message.Body.Hash.B58String())
			// bs := flood.WaitRequest(mc.CLASS_findfileinfo, hex.EncodeToString(message.Body.Hash), 0)
			bs, _ := flood.WaitRequest(mc.CLASS_findfileinfo, utils.Bytes2string(message.Body.Hash), 0)

			// fmt.Println("开始等待查找返回000", message.Body.Hash.B58String())
			if bs != nil {
				// fmt.Println("等待查找已经返回", string(*bs))
				fi, err = ParseFileindex(*bs)
				return
			}

		}
	}
	//	fmt.Println("这里直接就返回错误了")
	return

}

/*
	网络中下载一个文件到本地
*/
func DownloadFileOpt(fileindex *FileIndex) error {
	var dp *DownProc
	//下载时才加入下载列表
	dp = NewDownProc(fileindex) //添加到下载列表
	//如果缓存文件已存在，则直接返回
	ok, errs := utils.PathExists(filepath.Join(sconfig.Store_temp, fileindex.Hash.B58String()))
	if errs != nil {
		// fmt.Println("err 2222222")
		return errs
	}
	if ok {
		return nil
	}
	err := DownloadFilechunkToLocal(fileindex, dp)
	return err

}

// //同步文件索引到邻居节点
// func SyncFiletoNearId(fi *FileIndex) error {
// 	recvId := nodeStore.FindNearInSuper(&nodeStore.NodeSelf.IdInfo.Id, nil, false)
// 	if recvId == nil {
// 		return errors.New("没有附近节点")
// 	}
// 	//加入自己为共享用户
// 	fi.AddShareUser(recvId)

// 	// for _, v := range fi.FileChunk.GetAll() {
// 	// 	one := v.(*FileChunk)
// 	// 	fi.AddShareUser(one.No, recvId)
// 	// }
// 	content := fi.JSON()

// 	mhead := mc.NewMessageHead(recvId, recvId, false)
// 	mbody := mc.NewMessageBody(&content, "", nil, 0)
// 	message := mc.NewMessage(mhead, mbody)
// 	if message.Send(MSGID_addFileShare) {
// 		bs := flood.WaitRequest(mc.CLASS_sharefile, message.Body.Hash.B58String())
// 		if bs == nil {
// 			// fmt.Println("发送共享文件消息失败，可能超时")
// 			return errors.New("发送共享文件消息失败，可能超时")
// 		}
// 		return nil
// 	}
// 	return nil
// }
