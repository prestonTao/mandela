package persistence

import (
	// "fmt"
	gconfig "mandela/config"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

var msgTableKey uint64 = 0 //消息表主键，id自增长

/*
	加载表格最大id
*/
func loadMsgTableKey() error {
	rows, err := db.Query("select * from message order by id desc limit ?", 1)
	if err != nil {
		return err
	}

	for rows.Next() {
		err = rows.Scan(&msgTableKey)
		if err != nil {
			break
		}
	}
	// fmt.Println("加载的日志id最大为：", msgTableKey)
	return err
}

/*
	按页倒序读取日志
	@count   int    一页显示数量
	@order   int64  unix时间
*/
func findMsgLogPage(count int, order int64) ([]Message, error) {
	//select * from message where updatetime > 0 order by updatetime desc limit 10

	msgs := make([]Message, 0)
	rows, err := db.Query("select * from message where updatetime > ? order by updatetime desc limit ?", order, count)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		msg := Message{}
		err = rows.Scan(&msg.Id, &msg.Sender, &msg.Recver, &msg.Content, &msg.UpdateTime)
		if err != nil {
			break
		}
		msgs = append(msgs, msg)
	}

	return msgs, err
}

/*
	保存消息日志
*/
func SaveMsgLog(sendId, recver, content string) error {
	//	return tracefile(name, sendId, content)

	stmt, err := db.Prepare("insert into message values(?,?,?,?,?)")
	if err != nil {
		return err
	}
	stmt.Exec(atomic.AddUint64(&msgTableKey, 1), sendId, recver, content, time.Now().Unix())
	//	friendIdsLock.Lock()
	//	friendIds[id] = id
	//	friendIdsLock.Unlock()
	return nil

}

/*
	打印内容到文件中
*/
func tracefile(name, sendId, content string) error {
	fd, err := os.OpenFile(filepath.Join(gconfig.Path_configDir, folderName_msg, name), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	fd_time := time.Now().Format("2006-01-02 15:04:05")
	//	fd_content := strings.Join([]string{"======", fd_time, "=====", str_content, "\n"}, "")
	buf := []byte(sendId + " " + fd_time + " " + content + "\r\n")
	fd.Write(buf)
	fd.Close()
	return nil
}

type Message struct {
	Id         int
	Sender     string
	Recver     string
	Content    string
	UpdateTime int64
	UpdateNano int64
}
