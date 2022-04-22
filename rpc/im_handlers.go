package rpc

// import (
// 	"mandela/config"
// 	"mandela/core"
// 	"mandela/core/message_center"
// 	"mandela/core/message_center/flood"
// 	"mandela/core/nodeStore"
// 	"mandela/core/utils/crypto"

// 	// "mandela/im"
// 	"mandela/rpc/model"
// 	"mandela/sqlite3_db"
// 	"bytes"
// 	"encoding/hex"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"
// )

// /*
// 	更多的其他接口
// */
// func More(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	return
// }

// /*
// 	查询联系人列表
// */
// func GetContactsList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	fs := im.GetContactsList()
// 	res, err = model.Tojson(fs)
// 	return
// }

// /*
// 	添加联系人
// 	联系人有各种状态，新添加的联系人状态为1。添加之后给对方发送消息时候，对方会收到添加好友请求。
// */
// func AddContacts(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	if !rj.VerifyType("id", "string") {
// 		res, err = model.Errcode(model.TypeWrong, "id")
// 		return
// 	}
// 	idItr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "id")
// 		return
// 	}

// 	hello := ""
// 	helloItr, ok := rj.Get("hello")
// 	if ok {
// 		if !rj.VerifyType("hello", "string") {
// 			res, err = model.Errcode(model.TypeWrong, "hello")
// 			return
// 		}
// 		hello = helloItr.(string)
// 	}

// 	notename := ""
// 	notenameItr, ok := rj.Get("notename")
// 	if ok {
// 		if !rj.VerifyType("notename", "string") {
// 			res, err = model.Errcode(model.TypeWrong, "notename")
// 			return
// 		}
// 		notename = notenameItr.(string)
// 	}

// 	fdb := sqlite3_db.Friends{}
// 	f, err := fdb.FindById(idItr.(string))
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl)
// 		return
// 	}
// 	if f != nil {
// 		//好友已经存在
// 		res, err = model.Errcode(model.Exist)
// 		return
// 	}
// 	f = &sqlite3_db.Friends{
// 		NodeId:   idItr.(string), //网络节点id
// 		Nickname: "",             //昵称
// 		Notename: notename,       //备注昵称
// 		Note:     "",             //备注信息
// 		Status:   1,              //好友状态.1=添加好友时，用户不在线;2=申请添加好友状态;3=同意添加;4=拒绝添加;5=;6=;
// 		IsAdd:    2,              //是否自己主动添加的好友.1=别人添加的自己;2=自己主动添加的别人;
// 		Hello:    hello,          //打招呼内容
// 	}
// 	err = fdb.Add(f)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl)
// 		return
// 	}

// 	//发送打招呼消息
// 	go func() {

// 		addrNet := nodeStore.AddressFromB58String(idItr.(string))
// 		if hello == "" {
// 			hello = "请求添加你为好友"
// 		}
// 		helloBs := []byte(hello)

// 		message_center.SendP2pMsgHE(config.MSGID_TextMsg, &addrNet, &helloBs)
// 		// if ok {
// 		// 	bs := flood.WaitRequest(config.CLASS_im_msg_come, message.Body.Hash.B58String())
// 		// 	if bs != nil {
// 		// 		//发送成功，对方已经接收到消息
// 		// 		new(sqlite3_db.MsgLog).IsSuccessful(id)
// 		// 		res, err = model.Tojson(resultMap)
// 		// 		return
// 		// 	} else {
// 		// 		//发送失败，接收返回消息超时
// 		// 		new(sqlite3_db.MsgLog).IsFalse(id)
// 		// 		res, err = model.Errcode(model.TypeWrong, "send error!")
// 		// 		return
// 		// 	}
// 		// }
// 	}()

// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	同意添加好友
// */
// func AgreeToAdd(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	if !rj.VerifyType("id", "string") {
// 		res, err = model.Errcode(model.TypeWrong, "id")
// 		return
// 	}
// 	idItr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "id")
// 		return
// 	}

// 	f := new(sqlite3_db.Friends)
// 	f.NodeId = idItr.(string)
// 	f.Status = 3
// 	err = f.Update()
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 	} else {
// 		res, err = model.Tojson("success")

// 		//添加一个打招呼日志
// 		friend, err := f.FindById(idItr.(string))
// 		if err == nil {
// 			msglog := new(sqlite3_db.MsgLog)
// 			msglog.Add(friend.NodeId, "self", friend.Hello, "", core.MsgTextId)
// 		}
// 	}
// 	return
// }

// /*
// 	删除联系人
// */
// func DelContacts(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idItr, ok := rj.Get("id") //朋友id
// 	if !ok {
// 		res, err = model.Errcode(5002, "id")
// 		return
// 	}
// 	id := idItr.(string)

// 	im.DelContacts(id)

// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	修改好友信息
// */
// func UpdateContacts(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idItr, ok := rj.Get("id") //朋友id
// 	if !ok {
// 		res, err = model.Errcode(5002, "id")
// 		return
// 	}
// 	id := idItr.(string)

// 	nameItr, ok := rj.Get("notename") //好友昵称
// 	if !ok {
// 		res, err = model.Errcode(5002, "notename")
// 		return
// 	}
// 	notename := nameItr.(string)

// 	f := sqlite3_db.Friends{
// 		NodeId:   id,
// 		Notename: notename,
// 	}
// 	fmt.Println("%+v", f)
// 	f.UpdateNoteName()

// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	修改昵称
// */
// func UpdateNickName(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	nameItr, ok := rj.Get("nickname") //好友昵称
// 	if !ok {
// 		res, err = model.Errcode(5002, "nickname")
// 		return
// 	}
// 	nikename := nameItr.(string)

// 	p := sqlite3_db.Property{
// 		Hash:     nodeStore.NodeSelf.IdInfo.Id.B58String(),
// 		Nickname: nikename,
// 	}

// 	p.Update()

// 	res, err = model.Tojson("success")
// 	return
// }

// // /*
// // 	新的朋友异步推送
// // */
// // func GetNewFriend(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// // 	overtime := time.NewTicker(time.Second * 120)
// // 	select {
// // 	case <-overtime.C: //超时
// // 		res, err = model.Errcode(model.Timeout)
// // 		// res, err = model.Tojson("success")
// // 	case <-w.(http.CloseNotifier).CloseNotify(): //断开连接
// // 		overtime.Stop()
// // 		res, err = model.Errcode(model.Timeout)
// // 	case msg := <-im.NewFriend: //有新消息返回
// // 		overtime.Stop()
// // 		res, err = model.Tojson(msg)
// // 	}
// // 	return
// // }

// /*
// 	获取消息历史记录
// */
// func GetMsgHistory(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	if !rj.VerifyType("recipient", "string") {
// 		res, err = model.Errcode(model.TypeWrong, "recipient")
// 		return
// 	}
// 	recipientItr, ok := rj.Get("recipient")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "recipient")
// 		return
// 	}

// 	idItr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(5002, "id")
// 		return
// 	}
// 	id := int64(idItr.(float64))

// 	ml := sqlite3_db.MsgLog{}
// 	mls, err := ml.GetPage(recipientItr.(string), id)

// 	bs, err := json.Marshal(mls)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	mlsVO := make([]MsgLogVO, 0)
// 	// err = json.Unmarshal(bs, &mlsVO)
// 	decoder := json.NewDecoder(bytes.NewBuffer(bs))
// 	decoder.UseNumber()
// 	err = decoder.Decode(&mlsVO)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	for i, one := range mlsVO {
// 		mlsVO[i].Unix = one.CreateTime.Unix()
// 	}
// 	//拉取对方属性(昵称等)
// 	go UpdateProperty(recipientItr.(string))
// 	res, err = model.Tojson(mlsVO)
// 	return
// }

// type MsgLogVO struct {
// 	sqlite3_db.MsgLog
// 	Unix int64
// }

// func SendMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	if !rj.VerifyType("address", "string") {
// 		res, err = model.Errcode(model.TypeWrong, "address")
// 		return
// 	}
// 	addr, ok := rj.Get("address")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "address")
// 		return
// 	}
// 	if !rj.VerifyType("content", "string") {
// 		res, err = model.Errcode(model.TypeWrong, "content")
// 		return
// 	}
// 	content, ok := rj.Get("content")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "content")
// 		return
// 	}
// 	address := nodeStore.AddressFromB58String(addr.(string))
// 	bs := []byte(content.(string))

// 	msglog := sqlite3_db.MsgLog{}
// 	id, err := msglog.Add("self", addr.(string), content.(string), "", core.MsgTextId)
// 	if err != nil {
// 		res, err = model.Tojson(model.Nomarl)
// 		return
// 	}

// 	// fmt.Println("=================开始发送加密文本消息=====================")

// 	resultMap := make(map[string]interface{})
// 	resultMap["id"] = id
// 	resultMap["successful"] = 2

// 	go func() {
// 		message, ok, _ := message_center.SendP2pMsgHE(config.MSGID_TextMsg, &address, &bs)
// 		if ok {
// 			bs := flood.WaitRequest(config.CLASS_im_msg_come, hex.EncodeToString(message.Body.Hash), 0)
// 			if bs != nil {
// 				//发送成功，对方已经接收到消息
// 				new(sqlite3_db.MsgLog).IsSuccessful(id)
// 				res, err = model.Tojson(resultMap)
// 				return
// 			} else {
// 				//发送失败，接收返回消息超时
// 				new(sqlite3_db.MsgLog).IsFalse(id)
// 				res, err = model.Errcode(model.TypeWrong, "send error!")
// 				return
// 			}
// 		} else {
// 			new(sqlite3_db.MsgLog).IsFalse(id)
// 		}
// 	}()
// 	//发送消息失败
// 	// new(sqlite3_db.MsgLog).IsFalse(id)
// 	// res, err = model.Errcode(model.Nomarl, strconv.Itoa(int(id)))

// 	res, err = model.Tojson(resultMap)
// 	return
// }

// // type MsgLogVO struct {
// // 	Id int64
// // }

// /*
// 	重新发送消息
// */
// func ResendMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	idItr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(5002, "id")
// 		return
// 	}
// 	id := int64(idItr.(float64))
// 	msglog := sqlite3_db.MsgLog{}
// 	ml, e := msglog.FindById(id)
// 	if e != nil {
// 		err = e
// 		return
// 	}
// 	new(sqlite3_db.MsgLog).IsDefault(id)
// 	go func() {
// 		address := nodeStore.AddressFromB58String(ml.Recipient)
// 		bs := []byte(ml.Content)
// 		_, ok, _ = message_center.SendP2pMsgHE(config.MSGID_TextMsg, &address, &bs)
// 		if !ok {
// 			res, err = model.Tojson(model.Timeout)
// 			return
// 		}
// 		msglog.IsSuccessful(id)
// 	}()
// 	res, err = model.Tojson(id)
// 	return
// }

// /*
// 	设置消息已读
// */
// func IsReadMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idItr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(5002, "id")
// 		return
// 	}
// 	id := int64(idItr.(float64))

// 	msglog := sqlite3_db.MsgLog{}

// 	err = msglog.IsRead(id)
// 	if err != nil {
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	设置消息已读
// */
// func IsReadAddFirend(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idItr, ok := rj.Get("netid")
// 	if !ok {
// 		res, err = model.Errcode(5002, "netid")
// 		return
// 	}
// 	id := idItr.(string)

// 	friend := sqlite3_db.Friends{}
// 	friend.NodeId = id
// 	friend.Read = 2
// 	err = friend.Update()
// 	if err != nil {
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	获取消息队列中的消息
// */
// func GetNewMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	// out := make(map[string]interface{})

// 	overtime := time.NewTicker(time.Second * 120)
// 	select {
// 	case <-overtime.C:
// 		res, err = model.Errcode(model.Timeout)
// 		// res, err = model.Tojson("success")
// 	case <-w.(http.CloseNotifier).CloseNotify():
// 		overtime.Stop()
// 		res, err = model.Errcode(model.Timeout)
// 	case msg := <-core.MsgChannl:
// 		overtime.Stop()
// 		res, err = model.Tojson(msg)
// 	}
// 	return
// }

// /*
// 	删除消息
// */
// func RemoveMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	//要删除的消息id
// 	ids := make([]int64, 0)
// 	netIdsItr, ok := rj.Get("ids")
// 	if ok {
// 		netIds := netIdsItr.([]interface{})
// 		for _, one := range netIds {
// 			idOne := int64(one.(float64))
// 			ids = append(ids, idOne)
// 		}
// 	}

// 	log := new(sqlite3_db.MsgLog)
// 	err = log.Remove(ids...)
// 	if err == nil {
// 		res, err = model.Tojson("success")
// 		return
// 	}
// 	res, err = model.Errcode(model.Nomarl)
// 	return
// }

// /*
// 	删除某个好友的所有消息
// */
// func RemoveMsgAll(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	//要删除历史消息的好友id
// 	idItr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(5002, "id")
// 		return
// 	}
// 	id := idItr.(string)

// 	log := new(sqlite3_db.MsgLog)
// 	err = log.RemoveAllForFriend(id)
// 	if err == nil {
// 		res, err = model.Tojson("success")
// 		return
// 	}
// 	res, err = model.Errcode(model.Nomarl)
// 	return
// }

// //发送图文信息
// func SendPicMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	return SendFileAction(core.MsgPicId, rj, w, r)
// }

// //发送文件信息
// func SendFileMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	return SendFileAction(core.MsgFileId, rj, w, r)
// }

// //发送文件消息通用方法
// // class 消息类型
// func SendFileAction(class int, rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	if !rj.VerifyType("address", "string") {
// 		res, err = model.Errcode(model.TypeWrong, "address")
// 		return
// 	}
// 	addr, ok := rj.Get("address")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "address")
// 		return
// 	}
// 	//好友才能发送消息
// 	friend := sqlite3_db.Friends{}
// 	fri, _ := friend.FindById(addr.(string))
// 	if fri == nil {
// 		res, err = model.Errcode(model.Nomarl, "not friend")
// 		return
// 	}
// 	//不能超过最大传输大小
// 	mlog := sqlite3_db.MsgLog{}
// 	size, _ := mlog.FindSize(addr.(string))
// 	if size > im.Count {
// 		res, err = model.Errcode(model.Nomarl, "too large size")
// 		return
// 	}
// 	address := nodeStore.AddressFromB58String(addr.(string))
// 	if !rj.VerifyType("content", "string") {
// 		res, err = model.Errcode(model.TypeWrong, "content")
// 		return
// 	}
// 	content, ok := rj.Get("content")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "content")
// 		return
// 	}
// 	bs := []byte(content.(string))
// 	fp, ok := rj.Get("filepath")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "filepath")
// 		return
// 	}
// 	msg := &im.Msg{
// 		To:       address,
// 		Text:     bs,
// 		FilePath: fp.(string),
// 		Class:    class,
// 		Speed:    make(map[string]int64),
// 	}
// 	msglog := sqlite3_db.MsgLog{}
// 	id, err := msglog.Add("self", address.B58String(), string(msg.Text), msg.FilePath, class)
// 	if err != nil {
// 		fmt.Println(err)
// 		res, err = model.Tojson(model.Nomarl)
// 		return
// 	}
// 	resultMap := make(map[string]interface{})
// 	resultMap["id"] = id
// 	go func() {
// 		ok = msg.SendFile(id)
// 		if ok {
// 			new(sqlite3_db.MsgLog).IsSuccessful(id)
// 		} else {
// 			//发送消息失败
// 			new(sqlite3_db.MsgLog).IsFalse(id)
// 		}
// 	}()
// 	resultMap["successful"] = 1
// 	res, err = model.Tojson(resultMap)
// 	return
// }
// func ReSendFileAction(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idstr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "id")
// 		return
// 	}
// 	id := int64(idstr.(float64))
// 	resultMap := make(map[string]interface{})
// 	resultMap["id"] = id
// 	msg := &im.Msg{Speed: make(map[string]int64)}
// 	//状态设为默认
// 	new(sqlite3_db.MsgLog).IsDefault(id)
// 	go func() {
// 		ok = msg.SendFile(id)
// 		if ok {
// 			new(sqlite3_db.MsgLog).IsSuccessful(id)
// 		} else {
// 			//发送消息失败
// 			new(sqlite3_db.MsgLog).IsFalse(id)
// 		}
// 	}()
// 	resultMap["successful"] = 1
// 	res, err = model.Tojson(resultMap)
// 	return
// }

// //获取消息发送状态
// func GetMsgState(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	ids := make([]int64, 0)
// 	netIdsItr, ok := rj.Get("ids")
// 	if ok {
// 		netIds := netIdsItr.([]interface{})
// 		for _, one := range netIds {
// 			idOne := int64(one.(float64))
// 			ids = append(ids, idOne)
// 		}
// 	}
// 	//fmt.Printf("%+v", ids)
// 	msgs, err := new(sqlite3_db.MsgLog).FindState(ids)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson(msgs)
// 	return
// }

// //修改用户属性
// func UpdateProperty(hash string) {
// 	address := nodeStore.AddressFromB58String(hash)
// 	bs := []byte("ok")
// 	message, ok, _ := message_center.SendP2pMsgHE(config.MSGID_im_property, &address, &bs)
// 	if ok {
// 		bs := flood.WaitRequest(config.CLASS_im_property_msg, hex.EncodeToString(message.Body.Hash), 0)
// 		if bs != nil {
// 			//发送成功，对方已经接收到消息
// 			p := sqlite3_db.ParseProperty(*bs)
// 			f := sqlite3_db.Friends{NodeId: p.Hash, Nickname: p.Nickname}
// 			f.Update()
// 			return
// 		}
// 	}
// 	return
// }

// //获取消息发送状态
// func GetMyInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	msgs := new(sqlite3_db.Property).Get(nodeStore.NodeSelf.IdInfo.Id.B58String())
// 	res, err = model.Tojson(msgs)
// 	return
// }

// //获取好友昵称等
// func GetFriendInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idstr, ok := rj.Get("nodeid")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "nodeid")
// 		return
// 	}
// 	id := idstr.(string)
// 	msgs, _ := new(sqlite3_db.Friends).FindById(id)
// 	res, err = model.Tojson(msgs)
// 	return
// }

// //获取好友收款地址
// func GetFriendBaseCoinAddr(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	addr, ok := rj.Get("address")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "address")
// 		return
// 	}
// 	address := nodeStore.AddressFromB58String(addr.(string))
// 	if address == nil {
// 		res, err = model.Errcode(model.Nomarl, "address error")
// 		return
// 	}
// 	message, ok, _ := message_center.SendP2pMsgHE(config.MSGID_im_addr, &address, nil)
// 	if ok {
// 		bs := flood.WaitRequest(config.CLASS_im_addr_msg, hex.EncodeToString(message.Body.Hash), 0)
// 		if bs != nil {
// 			//发送成功，对方已经接收到消息
// 			basecoin := crypto.AddressCoin(*bs)
// 			res, err = model.Tojson(basecoin.B58String())
// 			return
// 		}
// 	}
// 	res, err = model.Errcode(model.Nomarl, "mess send error")
// 	return
// }

// //付款消息
// func SendPayMsg(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	addr, ok := rj.Get("address")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "address")
// 		return
// 	}
// 	address := nodeStore.AddressFromB58String(addr.(string))
// 	if address == nil {
// 		res, err = model.Errcode(model.Nomarl, "address error")
// 		return
// 	}
// 	content, ok := rj.Get("content")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "content")
// 		return
// 	}

// 	msglog := sqlite3_db.MsgLog{}
// 	id, err := msglog.Add("self", address.B58String(), content.(string), "", core.MsgPayId)
// 	if err != nil {
// 		fmt.Println(err)
// 		res, err = model.Errcode(model.Nomarl, "database error")
// 		return
// 	}
// 	resultMap := make(map[string]interface{})
// 	resultMap["id"] = id
// 	go func() {
// 		text := []byte(content.(string))
// 		message, ok, _ := message_center.SendP2pMsgHE(config.MSGID_im_pay, &address, &text)
// 		if ok {
// 			bs := flood.WaitRequest(config.CLASS_im_pay_msg, hex.EncodeToString(message.Body.Hash), 0)
// 			if bs != nil {
// 				//发送成功，对方已经接收到消息
// 				new(sqlite3_db.MsgLog).IsSuccessful(id)
// 				res, err = model.Tojson(resultMap)
// 				return
// 			} else {
// 				//发送失败，接收返回消息超时
// 				new(sqlite3_db.MsgLog).IsFalse(id)
// 				res, err = model.Errcode(model.TypeWrong, "send error!")
// 				return
// 			}
// 		} else {
// 			new(sqlite3_db.MsgLog).IsFalse(id)
// 		}
// 	}()
// 	resultMap["successful"] = 1
// 	res, err = model.Tojson(resultMap)
// 	return
// }

// //修改转帐确认状态
// func UpdatePayStatus(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idstr, ok := rj.Get("id")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "id")
// 		return
// 	}
// 	id := int64(idstr.(float64))
// 	statusstr, ok := rj.Get("status")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "status")
// 		return
// 	}
// 	status := int(statusstr.(float64))
// 	if status > 3 {
// 		res, err = model.Errcode(model.Nomarl, "status: 1 success/2 fail")
// 		return
// 	}
// 	new(sqlite3_db.MsgLog).IsPaySuccess(id, status)
// 	res, err = model.Tojson("successful")
// 	return
// }
