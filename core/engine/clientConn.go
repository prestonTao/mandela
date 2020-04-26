package engine

import (
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"time"
)

//本机向其他服务器的连接
type Client struct {
	sessionBase
	serverName string
	ip         string
	port       uint32
	conn       net.Conn
	inPack     chan *Packet //接收队列
	// packet     Packet       //
	//	isClose    bool         //该连接是否被关闭
	isPowerful bool //是否是强连接，强连接有短线重连功能
	engine     *Engine
	controller Controller
}

func (this *Client) Connect(ip string, port uint32) (remoteName string, err error) {

	this.ip = ip
	this.port = port

	this.conn, err = net.Dial("tcp", ip+":"+strconv.Itoa(int(port)))
	if err != nil {
		return
	}

	//权限验证
	remoteName, err = defaultAuth.SendKey(this.conn, this, this.serverName)
	if err != nil {
		this.conn.Close()
		return
	}

	// fmt.Println("Connecting to", ip, ":", strconv.Itoa(int(port)))
	Log.Debug("Connecting to %s:%s", ip, strconv.Itoa(int(port)))

	this.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		engine:     this.engine,
		attributes: make(map[string]interface{}),
	}
	// this.packet.Session = this
	go this.recv()
	// go this.hold()
	return
}
func (this *Client) reConnect() {
	for {
		//十秒钟后重新连接
		time.Sleep(time.Second * 10)
		var err error
		this.conn, err = net.Dial("tcp", this.ip+":"+strconv.Itoa(int(this.port)))
		if err != nil {
			continue
		}

		Log.Debug("Connecting to %s:%s", this.ip, strconv.Itoa(int(this.port)))

		go this.recv()
		// go this.hold()
		return
	}
}

func (this *Client) recv() {
	defer PrintPanicStack()
	var err error
	var handler MsgHandler
	for {
		//		Log.Debug("engine client 111111111111")
		var packet *Packet
		packet, err = RecvPackage(this.conn)
		if err != nil {
			//			Log.Debug("engine client 222222222222")
			break
		} else {
			packet.Session = this
			// Log.Debug("conn recv: %d %s %d\n%s", this.packet.MsgID, this.conn.RemoteAddr(), len(this.packet.Data)+16, hex.EncodeToString(this.packet.Data))
			//			Log.Debug("conn recv: %d %s %d", packet.MsgID, this.conn.RemoteAddr(), len(packet.Data)+16)
			// if packet.IsWait {
			// 	Log.Debug("开始等待")
			// 	packet.IsWait = false
			// 	packet.WaitChan <- true
			// 	Log.Debug("开始执行")
			// 	<-packet.WaitChan
			// 	Log.Debug("执行完成")
			// } else {
			//				Log.Debug("engine client 4444444444444")
			handler = this.engine.router.GetHandler(packet.MsgID)
			if handler == nil {
				Log.Warn("client该消息未注册，消息编号：%d", packet.MsgID)
				//					if this.packet.MsgID == 16 {
				//						fmt.Println(string(this.packet.Data))
				//					}
				//					break
			} else {
				//					Log.Debug("engine client 55555555555555")
				//这里决定了消息是否异步处理
				go this.handlerProcess(handler, packet)
				//					Log.Debug("engine client 6666666666666")
			}
			// }
			//			Log.Debug("engine client 777777777777777777")
			//				copy(this.cache, this.cache[this.packet.Size:this.cacheindex])
			//				this.cacheindex = this.cacheindex - uint32(n)

		}
	}

	this.Close()
	if this.isPowerful {
		go this.reConnect()
	}

}

// func (this *Client) Waite(du time.Duration) *Packet {
// 	if this.packet.Wait(du) {
// 		return &this.packet
// 	}
// 	return nil
// }

// func (this *Client) FinishWaite() {
// 	this.packet.FinishWait()
// }

func (this *Client) handlerProcess(handler MsgHandler, msg *Packet) {
	//消息处理模块报错将不会引起宕机
	defer PrintPanicStack()
	//消息处理前先通过拦截器
	itps := this.engine.interceptor.getInterceptors()
	itpsLen := len(itps)
	for i := 0; i < itpsLen; i++ {
		isIntercept := itps[i].In(this.controller, *msg)
		//
		if isIntercept {
			return
		}
	}
	//	Log.Debug("engine client 888888888888")
	handler(this.controller, *msg)
	//	Log.Debug("engine client 99999999999999")
	//消息处理后也要通过拦截器
	for i := itpsLen; i > 0; i-- {
		itps[i-1].Out(this.controller, *msg)
	}
}

//发送序列化后的数据
func (this *Client) Send(msgID uint64, data, dataplus *[]byte, waite bool) (err error) {
	defer PrintPanicStack()
	// this.packet.IsWait = waite
	buff := MarshalPacket(msgID, data, dataplus)

	//	var n int
	_, err = this.conn.Write(*buff)
	if err != nil {
		Log.Warn("conn send err: %s", err.Error())
	} else {
		// Log.Debug("conn send: %d %s %d %d\n%s", msgID, this.conn.RemoteAddr(), len(*buff), n, hex.EncodeToString(*buff))
		//		Log.Debug("conn send: %d %s %d %d", msgID, this.conn.RemoteAddr(), len(*buff), n)
	}
	return
}

//发送序列化后的数据
func (this *Client) SendJSON(msgID uint64, data interface{}, waite bool) (err error) {
	defer PrintPanicStack()
	// this.packet.IsWait = waite
	var f []byte
	f, err = json.Marshal(data)
	if err != nil {
		return
	}
	buff := MarshalPacket(msgID, &f, nil)
	_, err = this.conn.Write(*buff)
	return
}

//客户端关闭时,退出recv,send
func (this *Client) Close() {
	if this.engine.closecallback != nil {
		this.engine.closecallback(this.GetName())
	}
	this.engine.sessionStore.removeSession(this.GetName())
	err := this.conn.Close()
	if err != nil {
	}
}

//获取远程ip地址和端口
func (this *Client) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}

func (this *Client) SetName(name string) {
	this.engine.sessionStore.renameSession(this.name, name)
	this.name = name
}

func NewClient(name, ip string, port uint32) *Client {
	client := new(Client)
	client.name = name
	client.inPack = make(chan *Packet, 1000)
	// client.outData = make(chan *[]byte, 1000)
	client.Connect(ip, port)
	return client
}
