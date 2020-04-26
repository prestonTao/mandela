package im

import (
	"bytes"
	"encoding/json"
	"fmt"
)

//添加好友时结构体，附带把收款地址带过去
type FriendInfo struct {
	// ID       string `json:"id"`       //节点网络id
	Nickname string `json:"nickname"` //昵称
	Hello    string `json:"hello"`    //打招呼内容
}

func (friendinfo *FriendInfo) Json() []byte {
	res, err := json.Marshal(friendinfo)
	if err != nil {
		fmt.Println(err)
	}
	return res
}
func ParseFriendInfo(bs []byte) *FriendInfo {
	friendinfo := new(FriendInfo)
	// err := json.Unmarshal(bs, friendinfo)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(friendinfo)
	if err != nil {
		fmt.Println(err)
	}
	return friendinfo
}
