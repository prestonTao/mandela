package persistence

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
	gconfig "mandela/config"
)

var friendIdsLock = new(sync.RWMutex)
var friendIds = make(map[string]*Friends)
var friends = make([]*Friends, 0)

/*
	加载用户
*/
func loadFriends() error {
	raw, err := ioutil.ReadFile(filepath.Join(gconfig.Path_configDir, folderName_db, friendsFileName))
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buffer)
	friendIdsLock.Lock()
	err = dec.Decode(friends)
	if err == nil {
		for i, one := range friends {
			friendIds[one.Id] = friends[i]
		}
	}
	friendIdsLock.Unlock()
	if err != nil {
		return err
	}
	return nil
}

/*
	保存用户
*/
func saveFriends() error {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(friends)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(gconfig.Path_configDir, folderName_db, friendsFileName), buf.Bytes(), 0600)
	if err != nil {
		return err
	}
	return nil
}

/*
	消息中心的联系人
*/
type Friends struct {
	Id     string //id
	Name   string //昵称
	Unread int32  //未读消息数量
	Order  int64  //最新消息unix时间
	Type   int    //id类型  1=普通id; 2=群id;
}

func Friends_add(id string) error {
	f := Friends{
		Id:    id,
		Order: time.Now().Unix(),
	}

	friendIdsLock.Lock()
	friendIds[id] = &f
	friends = append(friends, &f)
	friendIdsLock.Unlock()

	return saveFriends()
}

/*
	添加未读消息数量
*/
func Friends_addMsgNum(id string) {
	friendIdsLock.RLock()
	f := friendIds[id]
	friendIdsLock.RUnlock()
	atomic.AddInt32(&f.Unread, 1)
	atomic.StoreInt64(&f.Order, time.Now().Unix())
	saveFriends()
}

/*
	将未读消息数量清零
*/
func Friends_msgNumClear(id string) {
	friendIdsLock.RLock()
	f := friendIds[id]
	friendIdsLock.RUnlock()
	atomic.StoreInt32(&f.Unread, 0)
	//	atomic.StoreInt64(&f.Order, time.Now().Unix())
	saveFriends()
}

func Friends_getall() (fs []*Friends) {
	friendIdsLock.RLock()
	fs = friends
	friendIdsLock.RUnlock()
	return
}

/*
	检查用户id是否存在
*/
func Friends_findIdExist(id string) (ok bool) {
	friendIdsLock.Lock()
	_, ok = friendIds[id]
	friendIdsLock.Unlock()
	return
}
