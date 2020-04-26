package engine

import (
	"sync"
)

type Controller interface {
	GetSession(name string) (Session, bool)      //通过accId得到客户端的连接Id
	GetNet() *Engine                             //获得连接到本地的计算机连接
	SetAttribute(name string, value interface{}) //设置共享数据，实现业务模块之间通信
	GetAttribute(name string) interface{}        //得到共享数据，实现业务模块之间通信
	GetGroupManager() MsgGroup                   //获得消息组管理器
}

type ControllerImpl struct {
	lock       *sync.RWMutex
	engine     *Engine
	attributes map[string]interface{}
	msgGroup   *msgGroupManager
}

//得到net模块，用于给用户发送消息
func (this *ControllerImpl) GetNet() *Engine {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.engine
}

func (this *ControllerImpl) SetAttribute(name string, value interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.attributes[name] = value
}
func (this *ControllerImpl) GetAttribute(name string) interface{} {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.attributes[name]
}

//
func (this *ControllerImpl) GetSession(name string) (Session, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.engine.GetSession(name)
}

func (this *ControllerImpl) GetGroupManager() MsgGroup {
	return this.msgGroup
}
