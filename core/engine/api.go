package engine

import (
	"net"
)

// "fmt"

//实例化
var engine *Engine

/*
	启动一个消息引擎
*/
func InitEngine(name string) {
	// Log.Info("创建Engine NewEngine")
	engine = NewEngine(name)
}

/*
	注册一个普通消息
*/
func RegisterMsg(msgId uint64, handler MsgHandler) {
	// if msgId <= 20 {
	// 	fmt.Println("该消息不能注册，消息编号0-20被系统占用。")
	// 	return
	// }
	engine.RegisterMsg(msgId, handler)
}

func Listen(ip string, port uint32, async bool) error {
	//	engine.run()
	return engine.Listen(ip, port, async)
}

func ListenByListener(listener *net.TCPListener, async bool) error {
	return engine.ListenByListener(listener, async)
}

/*
	添加一个连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
	@powerful      是否是强连接
	@return  name  对方的名称
*/
func AddClientConn(ip string, port uint32, powerful bool) (ss Session, err error) {
	//	engine.run()
	// Log.Info("检查engine %v", engine)
	session, err := engine.AddClientConn(ip, port, powerful)
	if err != nil {
		return nil, err
	}
	return session, err
}

//给一个session绑定另一个名称
func LinkName(name string, session Session) {

}

//添加一个拦截器，所有消息到达业务方法之前都要经过拦截器处理
func AddInterceptor(itpr Interceptor) {
	engine.interceptor.addInterceptor(itpr)
}

//得到session
func GetSession(name string) (Session, bool) {
	return engine.GetSession(name)
}

//通过ip地址和端口获得session,可以用于是否有重复连接
func GetSessionByHost(host string) Session {
	return engine.GetSessionByHost(host)
}

/*
	获取所有session
*/
func GetAllSession() []Session {
	return engine.GetAllSession()
}

//断开后清除session里的连接信息
func RemoveSession(name string) {
	engine.sessionStore.removeSession(name)
	return
}

//设置自定义权限验证
func SetAuth(auth Auth) {
	if auth == nil {
		return
	}
	defaultAuth = auth
}

//设置关闭连接回调方法
func SetCloseCallback(call CloseCallback) {
	engine.closecallback = call
}

/*
	暂停服务器
*/
func Suspend(names ...string) {
	engine.Suspend(names...)
}

/*
	恢复服务器
*/
func Recovery() {
	engine.Recovery()
}
