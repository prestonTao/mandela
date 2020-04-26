package engine

import (
	"sync"
)

type MsgGroup interface {
	RemoveGroup(groupName string)
	AddToGroup(groupName, name string)
	CheckNameInGroup(groupName, name string) bool
	GetNamesByGroup(groupName string) map[string]Session
}

type groupOne struct {
	lock     *sync.RWMutex
	msgGroup map[string]Session
}

//
type msgGroupManager struct {
	lock       *sync.RWMutex
	groups     map[string]*groupOne
	controller Controller
}

//创建一个小组
func (this *msgGroupManager) createGroup(groupName string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.groups[groupName] = new(groupOne)
}

//删除一个小组
func (this *msgGroupManager) RemoveGroup(groupName string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.groups, groupName)
}

//将一个连接添加到组中
func (this *msgGroupManager) AddToGroup(groupName, name string) {
	groupTag, ok := this.groups[groupName]
	if !ok {
		this.createGroup(groupName)
		groupTag, _ = this.groups[groupName]
	}
	session, ok := this.controller.GetSession(name)
	if ok {
		groupTag.msgGroup[name] = session
	}
}

//检查一个name是否在某个组中
func (this *msgGroupManager) CheckNameInGroup(groupName, name string) bool {
	groupTag, ok := this.groups[groupName]
	if ok {
		if _, ok = groupTag.msgGroup[name]; ok {
			return true
		}
	}
	return false
}

//得到一个组的所有成员名称
func (this *msgGroupManager) GetNamesByGroup(groupName string) map[string]Session {
	groupTag, ok := this.groups[groupName]
	if ok {
		return groupTag.msgGroup
	}
	return nil
}

func NewMsgGroupManager() *msgGroupManager {
	manager := new(msgGroupManager)
	manager.lock = new(sync.RWMutex)
	manager.groups = make(map[string]*groupOne)
	return manager
}
