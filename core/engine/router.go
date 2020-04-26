package engine

import (
	"sync"
)

type MsgHandler func(c Controller, msg Packet)

type Router struct {
	handlersMapping map[uint64]MsgHandler
	lock            *sync.RWMutex
}

func (this *Router) AddRouter(msgId uint64, handler MsgHandler) {
	this.lock.Lock()
	if _, ok := this.handlersMapping[msgId]; ok {
		Log.Warn("协议编号 [%d] 被覆盖", msgId)
	}
	this.handlersMapping[msgId] = handler
	this.lock.Unlock()
}

func (this *Router) GetHandler(msgId uint64) MsgHandler {
	this.lock.RLock()
	handler := this.handlersMapping[msgId]
	this.lock.RUnlock()
	return handler
}

func NewRouter() *Router {
	return &Router{
		handlersMapping: make(map[uint64]MsgHandler),
		lock:            new(sync.RWMutex),
	}
}
