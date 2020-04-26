package sqlite3_db

import (
	_ "github.com/go-xorm/xorm"
)

type Friends struct {
	Id           uint64 `xorm:"pk autoincr unique 'id'"`     //id
	NodeId       string `xorm:"varchar(25) 'nodeid'"`        //网络节点id
	Nickname     string `xorm:"varchar(25) 'nick_name'"`     //昵称
	Notename     string `xorm:"varchar(25) 'note_name'"`     //备注昵称
	Basecoinaddr string `xorm:"varchar(100) 'basecoinaddr'"` //好友收款地址
	Note         string `xorm:"varchar(25) 'note'"`          //备注信息
	Status       int    `xorm:"int 'status'"`                //好友状态.1=添加好友时，用户不在线;2=申请添加好友状态;3=同意添加;4=;5=;6=;
	IsAdd        int    `xorm:"int 'isadd'"`                 //是否自己主动添加的好友.1=别人添加的自己;2=自己主动添加的别人;
	Hello        string `xorm:"varchar(25) 'hello'"`         //打招呼内容
	Read         int    `xorm:"int 'read'"`                  //添加好友消息自己是否已读。1=未读;2=已读;
}

func (this *Friends) Add(f *Friends) error {
	_, err := table_friends.Insert(f)
	return err
}

func (this *Friends) Del(id string) error {
	_, err := engineDB.Where("nodeid = ?", id).Unscoped().Delete(this)
	return err
}

func (this *Friends) Update() error {
	_, err := engineDB.Where("nodeid = ?", this.NodeId).Update(this)
	return err
}

//修改用户备注，可以为空
func (this *Friends) UpdateNoteName() error {
	_, err := engineDB.Nullable("note_name").Where("nodeid = ?", this.NodeId).Update(this)
	return err
}
func (this *Friends) Getall() ([]Friends, error) {

	fs := make([]Friends, 0)
	err := engineDB.Find(&fs)
	return fs, err
}

/*
	检查用户id是否存在
*/
func (this *Friends) FindById(id string) (*Friends, error) {
	fs := make([]Friends, 0)
	err := engineDB.Where("nodeid = ?", id).Find(&fs)
	if err != nil {
		return nil, err
	}
	if len(fs) <= 0 {
		return nil, nil
	}
	return &fs[0], nil
}
