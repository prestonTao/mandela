package engine

import (
	"sync"
)

type sessionBase struct {
	name      string
	attrbutes *sync.Map // map[string]interface{}
	//	cache      []byte
	//	cacheindex uint32
	//	tempcache  []byte
	// lock *sync.RWMutex
}

func (this *sessionBase) Set(name string, value interface{}) {
	this.attrbutes.Store(name, value)
	// this.lock.Lock()
	// this.attrbutes[name] = value
	// this.lock.Unlock()
}
func (this *sessionBase) Get(name string) interface{} {
	v, _ := this.attrbutes.Load(name)
	return v
	// this.lock.RLock()
	// itr := this.attrbutes[name]
	// this.lock.RUnlock()
	// return itr
}
func (this *sessionBase) GetName() string {
	return this.name
}

func (this *sessionBase) SetName(name string) {
	//	this.sessionStore.renameSession(this.name, name)
	this.name = name
}

//func (this *sessionBase) Send(msgID, opt, errcode uint32, cryKey []byte, data *[]byte) (err error) {
//	return
//}
func (this *sessionBase) Close() {}

// func (this *sessionBase) GetRemoteHost() string {
// 	return "127.0.0.1:0"
// }

type Session interface {
	Send(msgID uint64, data, datapuls *[]byte, waite bool) error
	SendJSON(msgID uint64, data interface{}, waite bool) error
	// Waite(du time.Duration) *Packet
	// FinishWaite()
	Close()
	Set(name string, value interface{})
	Get(name string) interface{}
	GetName() string
	SetName(name string)
	GetRemoteHost() string
}

type sessionStore struct {
	//	lock              *sync.RWMutex
	//	nameStore         map[string]Session
	//	sessionPool       chan *ServerConn //待使用的连接池
	//	sessionClientPool chan *Client     //客户端待使用的连接池
	nameStore *sync.Map //key:string=sessionID;value:Session=Session对象;
}

func (this *sessionStore) addSession(name string, session Session) {
	//	this.lock.Lock()
	//	this.nameStore[session.GetName()] = session
	//	this.lock.Unlock()

	// netaddr := AddressNet([]byte(name))

	// Log.Info("add sessionid %s", netaddr.B58String())
	this.nameStore.Store(name, session)
}

func (this *sessionStore) getSession(name string) (Session, bool) {
	//	this.lock.RLock()
	//	s, ok := this.nameStore[name]
	//	this.lock.RUnlock()
	// this.nameStore.Range(func(k, v interface{}) bool {
	// nameOne := k.(string)
	// netaddr := AddressNet([]byte(nameOne))
	// Log.Info("get sessionid one: %s", netaddr.B58String())
	// return true
	// })

	// netaddr := AddressNet([]byte(name))
	// Log.Info("get sessionid %s", netaddr.B58String())
	value, ok := this.nameStore.Load(name)
	if !ok {
		return nil, false
	}
	ss := value.(Session)
	return ss, true
}

func (this *sessionStore) getSessionByHost(host string) Session {
	var session Session
	this.nameStore.Range(func(k, v interface{}) bool {
		one := v.(Session)
		// Log.Info("getSessionByHost %s %s", host, one.GetRemoteHost())
		if host == one.GetRemoteHost() {
			session = one
			return false
		}
		return true
	})
	return session
}

func (this *sessionStore) removeSession(name string) {
	netaddr := AddressNet([]byte(name))
	Log.Info("del sessionid %s", netaddr.B58String())
	this.nameStore.Delete(name)
}

func (this *sessionStore) renameSession(oldName, newName string) {
	value, ok := this.nameStore.Load(oldName)
	if !ok {
		return
	}
	// Log.Info("update sessionid %s %s", oldName, newName)
	this.nameStore.Store(newName, value)
	this.nameStore.Delete(oldName)
}

func (this *sessionStore) getAllSession() []Session {
	ss := make([]Session, 0)
	this.nameStore.Range(func(k, v interface{}) bool {
		s := v.(Session)
		ss = append(ss, s)
		return true
	})
	return ss

	//	this.lock.RLock()
	//	for _, s := range this.nameStore {
	//		ss = append(ss, s)
	//	}
	//	this.lock.RUnlock()
	//	return ss
}

func (this *sessionStore) getAllSessionName() []string {
	names := make([]string, 0)
	this.nameStore.Range(func(k, v interface{}) bool {
		name := k.(string)
		names = append(names, name)
		return true
	})
	return names

	//	this.lock.RLock()
	//	for key, _ := range this.nameStore {
	//		names = append(names, key)
	//	}
	//	this.lock.RUnlock()
	//	return names
}

/*
	获得一个未使用的服务器连接
*/
func (this *sessionStore) getClientConn(engine *Engine) *Client {
	//	select {
	//	case ss := <-this.sessionClientPool:
	//		//		ss.cacheindex = 0
	//		ss.attrbutes = make(map[string]interface{})
	//		select {
	//		case <-ss.packet.WaitChan:
	//		default:
	//		}
	//		return ss
	//	default:
	//	}

	sessionBase := sessionBase{
		attrbutes: new(sync.Map),
		//		cache:      make([]byte, 1024, 16*1024*1024),
		//		cacheindex: 0,
		//		tempcache:  make([]byte, 1024, 1024),
		// lock: new(sync.RWMutex),
	}
	clientConn := &Client{
		sessionBase: sessionBase,
		//		serverName:  this.name,
		//		inPack: make(chan *Packet, 5000),
		// packet: Packet{WaitChan: make(chan bool, 1)},
		engine: engine,
		//		isPowerful:  powerful,
	}
	// clientConn.packet.temp = make([]byte, 0)
	return clientConn

}

/*
	获得一个未使用的服务器连接
*/
func (this *sessionStore) getServerConn(engine *Engine) *ServerConn {
	//	select {
	//	case ss := <-this.sessionPool:
	//		//		ss.cacheindex = 0
	//		ss.attrbutes = make(map[string]interface{})
	//		select {
	//		case <-ss.packet.WaitChan:
	//		default:
	//		}
	//		return ss
	//	default:
	//	}
	//创建一个新的session
	sessionBase := sessionBase{
		attrbutes: new(sync.Map),
		//		name:       "",
		//		cache:      make([]byte, 16*1024*1024, 16*1024*1024),
		//		cacheindex: 0,
		//		tempcache:  make([]byte, 1024, 1024),
		// lock:      new(sync.RWMutex),
		// attrbutes: make(map[string]interface{}),
	}

	serverConn := &ServerConn{
		sessionBase: sessionBase,
		//		conn:           nil,
		//		Ip:             conn.RemoteAddr().String(),
		//		Connected_time: time.Now().String(),
		// packet: Packet{WaitChan: make(chan bool, 1)},
		engine: engine,
	}
	serverConn.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		engine:     engine,
		attributes: make(map[string]interface{}),
	}
	// serverConn.packet.temp = make([]byte, 0)
	return serverConn
}

func NewSessionStore() *sessionStore {
	sessionStore := new(sessionStore)
	sessionStore.nameStore = new(sync.Map)
	//	sessionStore.lock = new(sync.RWMutex)
	//	sessionStore.nameStore = make(map[string]Session)
	//	n := 1000
	//	sessionStore.sessionPool = make(chan *ServerConn, n)
	//	sessionStore.sessionClientPool = make(chan *Client, n)
	//	for i:=0;i<
	return sessionStore
}
