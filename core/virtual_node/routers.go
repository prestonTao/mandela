package virtual_node

// import (
// 	"sync"
// )

// var router = NewRouter()

// // func RegisterProtocol() {

// // 	// message_center

// // }

// // /*
// // 	注册搜索节点消息
// // 	从所有节点中搜索节点，虚拟节点中没有超级节点和普通节点的区别
// // */
// // func Register_search_all(msgid uint64, handler message_center.MsgHandler) {
// // 	router.Register(msgid, handler)
// // }

// // /*
// // 	从所有节点中搜索目标节点消息控制器
// // */
// // func searchAllHandler(c engine.Controller, msg engine.Packet) {
// // 	message, err := message_center.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
// // 	if err != nil {
// // 		return
// // 	}
// // 	form := nodeStore.AddressFromB58String(msg.Session.GetName())
// // 	if message.IsSendOther(&form) {
// // 		return
// // 	}
// // 	//解析包体内容
// // 	if err := message.ParserContent(); err != nil {
// // 		// fmt.Println(err)
// // 		return
// // 	}

// // 	//自己处理
// // 	h := router.GetHandler(message.Body.MessageId)
// // 	if h == nil {
// // 		fmt.Println("这个searchAll消息未注册:", message.Body.MessageId)
// // 		return
// // 	}
// // 	h(c, msg, message)
// // }

// // /*
// // 	注册虚拟节点之间的点对点通信消息
// // */
// // func Register_p2pHE(msgid uint64, handler message_center.MsgHandler) {
// // 	router.Register(msgid, handler)
// // }

// //---------------------

// type Router struct {
// 	handlers *sync.Map //key:uint64=消息版本号;value:MsgHandler=;
// }

// type RouterClass struct {
// 	class      uint64 //路由类型
// 	msgHandler message_center.MsgHandler
// }

// func (this *Router) Register(version uint64, handler message_center.MsgHandler) {

// 	this.handlers.Store(version, handler)
// }

// func (this *Router) GetHandler(msgid uint64) message_center.MsgHandler {
// 	value, ok := this.handlers.Load(msgid)
// 	if !ok {
// 		return nil
// 	}
// 	h := value.(message_center.MsgHandler)
// 	return h
// }

// func NewRouter() *Router {
// 	return &Router{
// 		handlers: new(sync.Map),
// 	}
// }
