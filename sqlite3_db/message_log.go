package sqlite3_db

import (
	"errors"
	"sync/atomic"
	"time"

	_ "github.com/go-xorm/xorm"
)

const Self = "self"

var generateMsgLogId int64 = 0

/*
	加载聊天记录表最大的id是多少
*/
func LoadMsgLogGenerateID() error {
	mls := make([]MsgLog, 0)
	err := table_msglog.Desc("id").Limit(1, 0).Find(&mls)
	if err != nil {
		return err
	}
	if len(mls) <= 0 {
		return nil
	}
	generateMsgLogId = mls[0].Id
	return nil
}

/*
	生成自增长id
*/
func GenerateMsgLogId() int64 {
	return atomic.AddInt64(&generateMsgLogId, 1)
}

type MsgLog struct {
	Id         int64     `xorm:"int64 pk notnull unique 'id'"`          //id
	Sender     string    `xorm:"varchar(25) index(recipient) 'sender'"` //发送者
	Recipient  string    `xorm:"varchar(25) 'recipient'"`               //接收者
	Content    string    `xorm:"varchar(25) 'content'"`                 //消息内容
	Read       int       `xorm:"int 'read'"`                            //消息自己是否已读。1=未读;2=已读;
	Successful int       `xorm:"int 'successful'"`                      //消息发送成功。1 默认(未收到结果) 2=成功;3=失败;
	Path       string    `xorm:"varchar(25) 'path'"`                    //文件本地地址
	Class      int       `xorm:"varchar(25) 'class'"`                   //消息类型。1=文本消息;2=表情包;3=图片;4=文件;5=视频;6=链接;
	PayStatus  int       `xorm:"int 'paystatus'"`                       //支付状态 0 未确认 1 成功 2 失败
	Size       int64     `xorm:"int64 'size'"`                          //文件大小
	Rate       int64     `xorm:"int64 'rate'"`                          //传送进度
	Speed      int64     `xorm:"int64 'speed'"`                         //传送速度
	Index      int64     `xorm:"int64 'index'"`                         //传送文件索引
	Toid       int64     `xorm:"int64 'toid'"`                          //对方消息ID,用于断点续传
	CreateTime time.Time `xorm:"created 'createtime'"`                  //创建时间，这个Field将在Insert时自动赋值为当前时间
	UpdateTime uint64    `xorm:"updated 'updated'"`                     //修改后自动更新时间
}

/*
	添加一个消息记录
	@return    int64    数据库id
*/
func (this *MsgLog) Add(sender, recipient, content, path string, class int) (int64, error) {

	ml := MsgLog{
		Id:         GenerateMsgLogId(),
		Sender:     sender,
		Recipient:  recipient,
		Content:    content,
		Path:       path,  //文件地址
		Class:      class, //消息类型
		Read:       1,     //消息已读。1=未读;2=已读;
		Successful: 1,     //消息发送成功。1=未读;2=已读;
	}
	_, err := engineDB.Insert(&ml)
	return ml.Id, err
}

/*
	把消息记录设置为已读
*/
func (this *MsgLog) IsRead(id int64) error {
	ml := MsgLog{
		Read: 2, //消息已读。1=未读;2=已读;
	}
	_, err := engineDB.Id(id).Update(&ml)
	return err
}

/*
	把消息记录设置为发送成功
*/
func (this *MsgLog) IsSuccessful(id int64) error {
	ml := MsgLog{
		Successful: 2, //消息发送成功。1=不成功;2=成功;
	}
	_, err := engineDB.Id(id).Update(&ml)
	return err
}

/*
	把消息记录设置为已确认(收款用)
*/
func (this *MsgLog) IsPaySuccess(id int64, status int) error {
	ml := MsgLog{
		PayStatus: status, //消息已读。0=未确认;1=成功;2=失败;
	}
	_, err := engineDB.Id(id).Update(&ml)
	return err
}

/*
	把消息记录设置为发送成功
*/
func (this *MsgLog) IsDefault(id int64) error {
	ml := MsgLog{
		Successful: 1, //消息发送成功。1=不成功;2=成功;
	}
	_, err := engineDB.Id(id).Update(&ml)
	return err
}

//修改传送进度
func (this *MsgLog) UpRate(id, toid, index, rate, speed, size int64) error {
	if index > size {
		index = size
	}
	if rate > 100 {
		rate = 100
		speed = 0
	}
	ml := MsgLog{
		Toid:  toid,  //对方消息ID
		Index: index, //传送偏移量
		Rate:  rate,  //已发送百分比
		Size:  size,  //文件大小
		Speed: speed, //速率 单位字节
	}
	_, err := engineDB.Id(id).Update(&ml)
	return err
}

/*
	把消息记录设置为发送成功
*/
func (this *MsgLog) IsFalse(id int64) error {
	ml := MsgLog{
		Successful: 3, //消息发送成功。1=默认(未接收到状态);2=成功;3=失败
	}
	_, err := engineDB.Id(id).Update(&ml)
	return err
}
func (this *MsgLog) GetPage(recipient string, startId int64) ([]MsgLog, error) {
	mls := make([]MsgLog, 0)
	var err error
	if startId == 0 {
		//(t.sender = self and t.recipient = recipient) or (t.sender = recipient and t.recipient = self)
		err = engineDB.Alias("t").Where("t.sender = ? or t.recipient = ?",
			recipient, recipient).Limit(10, 0).Desc("createtime").Find(&mls)
	} else {
		err = engineDB.Alias("t").Where("t.id < ? and (t.sender = ? or t.recipient = ?)",
			startId, recipient, recipient).Limit(10, 0).Desc("createtime").Find(&mls)
	}

	return mls, err
}

/*
	查询一条记录
*/
func (this *MsgLog) FindById(id int64) (*MsgLog, error) {
	mls := make([]MsgLog, 0)
	//err := table_msglog.Id(id).Find(&mls)
	err := engineDB.Where("id=?", id).Find(&mls)
	if err != nil {
		return nil, err
	}
	if len(mls) <= 0 {
		return nil, errors.New("未查找到记录")
	}
	return &mls[0], nil
}

/*
	删除多个消息记录
*/
func (this *MsgLog) Remove(ids ...int64) error {
	_, err := engineDB.In("id", ids).Unscoped().Delete(this)
	return err
}

/*
	删除和某个好友的所有聊天记录
*/
func (this *MsgLog) RemoveAllForFriend(recipient string) error {
	_, err := engineDB.Where("sender = ? or recipient = ?",
		recipient, recipient).Unscoped().Delete(this)
	return err
}

//查询消息发送状态
func (this *MsgLog) FindState(ids []int64) ([]MsgLog, error) {
	res := []MsgLog{}
	err := engineDB.In("id", ids).Unscoped().Find(&res)
	return res, err
}

//查询给某个好友传输的文件总大小
func (this *MsgLog) FindSize(toid string) (int64, error) {
	size, err := engineDB.Where("Recipient=? and successful=?", toid, 2).SumInt(this, "size")
	return size, err
}

// /*
// 	保存消息日志
// */
// func SaveMsgLog(name, sendId, content string) {
// 	tracefile(name, sendId, content)
// }

// /*
// 	打印内容到文件中
// */
// func tracefile(name, sendId, content string) error {
// 	fd, err := os.OpenFile(filepath.Join(config.Path_configDir, name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
// 	if err != nil {
// 		return err
// 	}
// 	fd_time := time.Now().Format("2006-01-02 15:04:05")
// 	//	fd_content := strings.Join([]string{"======", fd_time, "=====", str_content, "\n"}, "")
// 	buf := []byte(sendId + " " + fd_time + " " + content + "\r\n")
// 	fd.Write(buf)
// 	fd.Close()
// 	return nil
// }
