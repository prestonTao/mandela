package engine

import (
	"net"
	"strconv"
	"sync"
	"time"
)

type Engine struct {
	name          string //本机名称
	status        int    //服务器状态
	auth          Auth
	onceRead      *sync.Once
	interceptor   *InterceptorProvider
	sessionStore  *sessionStore
	closecallback CloseCallback
	lis           *net.TCPListener
	router        *Router
	isSuspend     bool //暂停服务器
}

/*
	注册一个普通消息
*/
func (this *Engine) RegisterMsg(msgId uint64, handler MsgHandler) {
	//打印注册消息id
	Log.Debug("register message id %d", msgId)
	this.router.AddRouter(msgId, handler)
}

func (this *Engine) Listen(ip string, port uint32, async bool) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip+":"+strconv.Itoa(int(port)))
	if err != nil {
		Log.Error("%v", err)
		return err
	}

	this.lis, err = net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		Log.Error("%v", err)
		return err
	}
	//监听一个地址和端口
	Log.Debug("Listen to an IP：%s", ip+":"+strconv.Itoa(int(port)))
	if async {
		go this.listener(this.lis)
	} else {
		this.listener(this.lis)
	}
	return nil
	//	return this.net.Listen(ip, port)
}

func (this *Engine) ListenByListener(listener *net.TCPListener, async bool) error {
	if async {
		go this.listener(this.lis)
	} else {
		this.listener(this.lis)
	}
	return nil
}

func (this *Engine) listener(listener *net.TCPListener) {
	this.lis = listener
	//	this.ipPort = listener.Addr().String()
	var conn net.Conn
	var err error
	for !this.isSuspend {
		conn, err = this.lis.Accept()
		if err != nil {
			continue
		}
		if this.isSuspend {
			conn.Close()
			continue
		}
		go this.newConnect(conn)
	}
}

//创建一个新的连接
func (this *Engine) newConnect(conn net.Conn) {
	defer PrintPanicStack()
	remoteName, err := defaultAuth.RecvKey(conn, this.name)
	if err != nil {
		//接受连接错误
		Log.Warn("Accept connection error %s", err.Error())
		conn.Close()
		return
	}

	serverConn := this.sessionStore.getServerConn(this)
	serverConn.name = remoteName
	serverConn.conn = conn
	serverConn.Ip = conn.RemoteAddr().String()
	serverConn.Connected_time = time.Now().String()
	serverConn.outChan = make(chan *[]byte, 10000)
	serverConn.outChanCloseLock = new(sync.Mutex)
	serverConn.outChanIsClose = false
	serverConn.run()
	this.sessionStore.addSession(remoteName, serverConn)

	// fmt.Println(time.Now().String(), "建立连接", conn.RemoteAddr().String())
	Log.Debug("Accept remote addr:%s", conn.RemoteAddr().String())

}

/*
	添加一个连接，给这个连接取一个名字，连接名字可以在自定义权限验证方法里面修改
	@powerful      是否是强连接
	@return  name  对方的名称
*/
func (this *Engine) AddClientConn(ip string, port uint32, powerful bool) (ss Session, err error) {
	// Log.Info("1111 %v", this.sessionStore)

	clientConn := this.sessionStore.getClientConn(this)
	clientConn.name = this.name
	//	clientConn.attrbutes = make(map[string]interface{})
	remoteName, err := clientConn.Connect(ip, port)
	if err == nil {
		clientConn.name = remoteName
		this.sessionStore.addSession(remoteName, clientConn)
		return clientConn, nil
	}
	return nil, err
}

//添加一个拦截器，所有消息到达业务方法之前都要经过拦截器处理
func (this *Engine) AddInterceptor(itpr Interceptor) {
	this.interceptor.addInterceptor(itpr)
}

//获得session
func (this *Engine) GetSession(name string) (Session, bool) {
	return this.sessionStore.getSession(name)
}

//通过ip地址和端口获得session,可以用于是否有重复连接
func (this *Engine) GetSessionByHost(host string) Session {
	return this.sessionStore.getSessionByHost(host)
}

func (this *Engine) GetAllSession() []Session {
	return this.sessionStore.getAllSession()
}

//设置自定义权限验证
func (this *Engine) SetAuth(auth Auth) {
	if auth == nil {
		return
	}
	defaultAuth = auth
}

//设置关闭连接回调方法
func (this *Engine) SetCloseCallback(call CloseCallback) {
	this.closecallback = call
}

//@name   本服务器名称
func NewEngine(name string) *Engine {
	engine := new(Engine)
	engine.name = name
	engine.interceptor = NewInterceptor()
	engine.onceRead = new(sync.Once)
	engine.sessionStore = NewSessionStore()
	//	net.inPacket = RecvPackage
	//	net.outPacket = MarshalPacket
	engine.router = NewRouter()
	return engine
}

/*
	暂停服务器
*/
func (this *Engine) Suspend(names ...string) {
	// Log.Debug("暂停服务器")
	// this.lis.Close()
	this.isSuspend = true
	for _, one := range this.sessionStore.getAllSessionName() {
		done := false
		for _, nameOne := range names {
			if nameOne == one {
				done = true
				break
			}
		}
		if done {
			continue
		}
		if session, ok := this.GetSession(one); ok {
			session.Close()
		}
	}
}

/*
	恢复服务器
*/
func (this *Engine) Recovery() {
	Log.Debug("恢复服务器")
	this.isSuspend = false
	go this.listener(this.lis)
}
