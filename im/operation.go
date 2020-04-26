package im

import (
	"mandela/sqlite3_db"
)

// var NewFriend = make(chan sqlite3_db.Friends, 1000)

/*
	发送添加好友的消息
*/
func AddFriend_opt() {}

/*
	获取联系人列表
*/
func GetContactsList() []sqlite3_db.Friends {
	fs, err := new(sqlite3_db.Friends).Getall()
	if err != nil {
		return nil
	}
	return fs
}

/*
	添加联系人
*/
// func AddContacts(id, notename string) {
// 	err := new(sqlite3_db.Friends).Add(id, notename, "", 1, 2)
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// }

/*
	删除联系人
*/
func DelContacts(id string) {
	new(sqlite3_db.Friends).Del(id)
	new(sqlite3_db.MsgLog).RemoveAllForFriend(id)
}
