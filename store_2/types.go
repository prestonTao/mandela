package store

import (
	"time"
)

//文件所有者
type FileUser struct {
	Hash       string //用户hash
	UpdateTime int64  //最后在线时间
}

//更新在线时间
func (fu *FileUser) Update() error {
	fu.UpdateTime = time.Now().Unix()
	return nil
}
