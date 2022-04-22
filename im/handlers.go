package im

import (
	"mandela/config"
	"mandela/core"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/sqlite3_db"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

/*
	接收发送的图文消息
*/
func FileMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	var id int64
	//	fmt.Println("这个文本消息是自己的")
	//发送给自己的，自己处理
	content := *message.Body.Content
	m, err := ParseMsg(content)
	if err == nil {
		//发送者
		sendId := message.Head.Sender.B58String()
		m.File.Path = filepath.Join(imfilepath, m.File.Name)
		num := 0
	Rename:
		//如果文件存在，则重命名为新的文件名
		if ok, _ := utils.PathExists(m.File.Path); ok {
			num++
			filenamebase := filepath.Base(m.File.Name)
			fileext := filepath.Ext(m.File.Name)
			filename := strings.TrimSuffix(filenamebase, fileext)
			newname := filename + "_" + strconv.Itoa(num) + fileext
			m.File.Path = filepath.Join(imfilepath, newname)
			if ok1, _ := utils.PathExists(m.File.Path); ok1 {
				goto Rename
			}
			m.File.Name = newname
		}
		//临时文件，传输完成后改为原来文件名
		tmpPath := filepath.Join(imfilepath, m.File.Name+"_"+sendId+"_tmp")
		fi, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		start := m.File.Index - Lenth
		if start >= m.File.Size {
			start = m.File.Size
		}
		fi.Seek(start, 0)
		fi.Write(m.File.Data)
		defer fi.Close()
		fmt.Println(sendId, start)
		ml := sqlite3_db.MsgLog{}
		if m.Toid == 0 { //第一次传，写入
			id, err = ml.Add(sendId, sqlite3_db.Self, string(m.Text), m.File.Path, m.Class)
			if err != nil {
				fmt.Println(err)
				return
			}
		} else { //第二次传，则更新
			id = m.Toid
		}
		rate := int64(float64(m.File.Index) / float64(m.File.Size) * float64(100))
		m.SetSpeed(time.Now().Unix(), len(content))
		speed := m.GetSpeed()
		err = ml.UpRate(id, m.Nowid, m.File.Index, rate, speed, m.File.Size)
		if err != nil {
			fmt.Println("update transimission rate fail", err)
		}
		//传输完成，则更新状态
		if rate >= 100 {
			//传输完成，则重命名文件名
			fi.Close()
			os.Rename(tmpPath, m.File.Path)
			//标记状态为已完成
			ml.IsSuccessful(id)
			now := time.Now()
			msgVO := core.MessageVO{
				DBID:     id,
				Id:       sendId,
				Index:    now.Unix(),
				Time:     utils.FormatTimeToSecond(now),
				Content:  string(m.Text),
				Path:     m.File.Path,
				FileName: m.File.Name,
				Size:     m.File.Size,
				Cate:     m.Class,
			}
			msgVO.DBID = id
			select {
			case core.MsgChannl <- &msgVO:
			default:
			}
		}
	}
	//回复发送者，自己已经收到消息ID
	bs := utils.Int64ToBytes(id)
	message_center.SendP2pReplyMsgHE(message, config.MSGID_im_file_recv, &bs)
}

/*
	接收发送的图文消息  返回
*/
func FileMsg_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_im_file_msg, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(config.CLASS_im_file_msg, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	接收用户属性消息
*/
func PropertyMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//发送给自己的，自己处理
	p := new(sqlite3_db.Property).Get(nodeStore.NodeSelf.IdInfo.Id.B58String())
	//回复发送者，自己已经收到消息ID
	bs := p.Json()
	message_center.SendP2pReplyMsgHE(message, config.MSGID_im_property_recv, &bs)
}

/*
	接收用户属性消息  返回
*/
func PropertyMsg_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_im_property_msg, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(config.CLASS_im_property_msg, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	获取用户收款地址消息
*/
func BaseCoinAddrMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	//获取收款地址
	addr := []byte(keystore.GetCoinbase().Addr)
	message_center.SendP2pReplyMsgHE(message, config.MSGID_im_addr_recv, &addr)
}

/*
	获取用户收款地址消息  返回
*/
func BaseCoinAddrMsg_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_im_addr_msg, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(config.CLASS_im_addr_msg, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	文本消息
*/
func PayMsg(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	content := string(*message.Body.Content)
	//发送者
	sendId := message.Head.Sender.B58String()
	ml := sqlite3_db.MsgLog{}
	id, err := ml.Add(sendId, sqlite3_db.Self, content, "", core.MsgPayId)
	if err == nil {
		ml.IsSuccessful(id)
	}

	now := time.Now()
	msgVO := core.MessageVO{
		DBID:    id,
		Id:      sendId,
		Index:   now.Unix(),
		Time:    utils.FormatTimeToSecond(now),
		Content: content,
		Cate:    core.MsgPayId,
	}
	msgVO.DBID = id
	select {
	case core.MsgChannl <- &msgVO:
	default:
	}
	bs := []byte("ok")
	message_center.SendP2pReplyMsgHE(message, config.MSGID_im_pay_recv, &bs)
}

/*
	文本消息  返回
*/
func PayMsg_recv(c engine.Controller, msg engine.Packet, message *message_center.Message) {
	// flood.ResponseWait(config.CLASS_im_pay_msg, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(config.CLASS_im_pay_msg, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}
