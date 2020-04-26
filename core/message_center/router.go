package message_center

import (
	"sync"
)

type Router struct {
	handlers *sync.Map //key:uint64=消息版本号;value:MsgHandler=;
}

func (this *Router) Register(version uint64, handler MsgHandler) {
	this.handlers.Store(version, handler)
}

func (this *Router) GetHandler(msgid uint64) MsgHandler {
	value, ok := this.handlers.Load(msgid)
	if !ok {
		return nil
	}
	h := value.(MsgHandler)
	return h
}

func NewRouter() *Router {
	return &Router{
		handlers: new(sync.Map),
	}
}
