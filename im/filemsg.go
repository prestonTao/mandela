/**
	图文消息发送
**/
package im

import (
	"mandela/config"
	"bytes"
	"encoding/hex"

	//"mandela/core"
	"mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/sqlite3_db"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const (
	Count  int64 = 20 * 1024 * 1024 * 1024 //单个好友传输总量不超过20G
	Lenth  int64 = 200 * 1024              //每次传输大小（200kb）
	ErrNum int   = 5                       //传输失败重试次数 5次
	Second int64 = 10                      //传输速度统计时间间隔 10秒
)

//消息体
type Msg struct {
	From     nodeStore.AddressNet //消息发送者
	To       nodeStore.AddressNet //消息接收者
	Nowid    int64                //自己消息ID
	Toid     int64                //对方消息ID
	Text     []byte               //文本内容
	FilePath string               //图片文件路径
	File     *FileInfo            //文件信息
	Class    int                  //消息类型 1 //文本消息 2 //表情包消息 3 //图片消息 4 //文件消息 5 //视频消息 6 //链接消息 7 //转帐消息
	Speed    map[string]int64     //传输速度统计
}

//采集速度参数
func (msg *Msg) SetSpeed(stime int64, size int) error {

	if _, ok := msg.Speed["time"]; !ok {
		msg.Speed["time"] = stime
		msg.Speed["size"] = int64(size)
	}
	if time.Now().Unix()-msg.Speed["time"] > Second {
		msg.Speed["time"] = stime
		msg.Speed["size"] = 0
	} else {
		msg.Speed["size"] += int64(size)
	}
	return nil
}

//获取速度
func (msg *Msg) GetSpeed() int64 {
	t := time.Now().Unix() - msg.Speed["time"]
	if t < 1 {
		t = 1
	}
	return msg.Speed["size"] / t * 100
}
func (msg *Msg) Json() []byte {
	d, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	return d
}

//解析消息
func ParseMsg(d []byte) (*Msg, error) {
	msg := &Msg{}
	// err := json.Unmarshal(d, msg)
	decoder := json.NewDecoder(bytes.NewBuffer(d))
	decoder.UseNumber()
	err := decoder.Decode(msg)
	if err != nil {
		fmt.Println(err)
	}
	return msg, err
}

//读取文件(一次性传输，弃用)
func (msg *Msg) ReadFile() []byte {
	stat, err := os.Stat(msg.FilePath)
	if err != nil {
		fmt.Println(err)
	}
	d, err := ioutil.ReadFile(msg.FilePath)
	if err != nil {
		fmt.Println(err)
	}
	fi := FileInfo{Name: stat.Name(), Size: stat.Size(), Data: d}
	msg.File = &fi
	return msg.Json()
}

//分段传输，续传
/**
@param id 消息ID ，i 分段序号
@return fd 段数据 fileinfo 文件属性ok 是否传送完 err 错误
*/
func (msg *Msg) ReadFileSlice(id int64) (fd []byte, fileinfo FileInfo, ok bool, errs error) {
	//msglog := sqlite3_db.MsgLog{}
	msginfo, err := new(sqlite3_db.MsgLog).FindById(id)
	if err != nil {
		fmt.Println(err)
		errs = err
		return
	}
	//已经传完
	if msginfo.Index >= msginfo.Size && msginfo.Size > 0 {
		return
	}
	//重发时重新组装msg
	msg.To = nodeStore.AddressFromB58String(msginfo.Recipient)
	msg.Text = []byte(msginfo.Content)
	msg.FilePath = msginfo.Path
	msg.Class = msginfo.Class
	msg.Nowid = id
	msg.Toid = msginfo.Toid
	path := msginfo.Path
	index := msginfo.Index //当前已传偏移量
	fi, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		errs = err
		return
	}
	stat, err := fi.Stat()
	if err != nil {
		fmt.Println(err)
		errs = err
		return
	}
	size := stat.Size()
	start := index
	length := Lenth
	//如果偏移量小于文件大小，并且剩余大小小于长度，则长度为剩余大小(即最后一段)
	if start < size && size-index < Lenth {
		length = size - index
	}
	buf := make([]byte, length)
	_, err = fi.ReadAt(buf, start)
	if err != nil {
		fmt.Println(err)
		errs = err
		return
	}
	//下一次start位置
	nextstart := start + Lenth
	if nextstart >= size {
		ok = true
	}
	fmt.Println("文件发送中...", size, nextstart)
	finfo := FileInfo{Name: stat.Name(), Size: size, Index: nextstart, Data: buf}
	//fmt.Printf("xxx%+v", finfo.Index)
	msg.File = &finfo
	fd = msg.Json()
	fileinfo = finfo
	return
}

//发送图文消息
func (msg *Msg) SendFile(id int64) (bl bool) {
	//bs:=msg.ReadFile()
	var errnum int
	for i := int64(0); ; i++ {
	BEGIN:
		bs, fi, okf, err := msg.ReadFileSlice(id)
		if bs == nil { //已传输完，则退出
			break
		}
		if err != nil {
			//开始重传
			errnum++
			if errnum <= ErrNum {
				fmt.Println("resend slice...")
				goto BEGIN
			}
			break
		}
		message, ok, _ := message_center.SendP2pMsgHE(config.MSGID_im_file, &msg.To, &bs)
		if ok {
			rbs := flood.WaitRequest(config.CLASS_im_file_msg, hex.EncodeToString(message.Body.Hash), 0)
			if rbs != nil {
				toid := utils.BytesToInt64(*rbs) //返回对方的消息ID
				//发送成功，对方已经接收到消息
				rate := int64(float64(fi.Index) / float64(fi.Size) * float64(100))
				msg.SetSpeed(time.Now().Unix(), len(bs))
				speed := msg.GetSpeed()
				//fmt.Println(id, fi.Index, rate, fi.Size)
				//fmt.Printf("update:%+v", fi.Index)
				new(sqlite3_db.MsgLog).UpRate(id, toid, fi.Index, rate, speed, fi.Size)
				if okf {
					bl = true
					break
				}
			} else {
				//发送失败，接收返回消息超时
				fmt.Println("fail")
				errnum++
				if errnum <= ErrNum {
					//开始重传
					fmt.Println("resend...")
					goto BEGIN
				}
				bl = false
				break
			}
		} else {
			bl = false
			break
		}
	}
	return
}

// func Test() {
// 	address := nodeStore.AddressFromB58String("2rkiUi1YfMmxx69TPfMSxQkF6bardtWZcPNYh8iMmA3a")
// 	msg := &Msg{
// 		To:       address,
// 		Text:     []byte("消息内容"),
// 		FilePath: "C:\\Users\\Administrator\\Pictures\\Saved Pictures\\2.jpg",
// 		Class:    2,
// 	}
// 	msglog := sqlite3_db.MsgLog{}
// 	id, err := msglog.Add("self", address.B58String(), string(msg.Text)+msg.FilePath, msg.FilePath, core.MsgPicId)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	fmt.Println(id)
// 	msg.SendFile(id)
// }
