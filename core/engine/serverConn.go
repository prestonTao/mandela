package engine

import (
	"errors"
	"net"
	"runtime"
	"sync"
	"time"
)

//其他计算机对本机的连接
type ServerConn struct {
	sessionBase
	conn             net.Conn
	Ip               string
	Connected_time   string
	CloseTime        string
	outChan          chan *[]byte //发送队列
	outChanCloseLock *sync.Mutex  //是否关闭发送队列
	outChanIsClose   bool         //是否关闭发送队列。1=关闭
	// packet         Packet
	engine     *Engine
	controller Controller
}

func (this *ServerConn) run() {

	// this.packet.Session = this
	go this.loopSend()
	go this.recv()

}

//接收客户端消息协程
func (this *ServerConn) recv() {
	defer PrintPanicStack()
	//处理客户端主动断开连接的情况
	var err error
	var handler MsgHandler
	for {
		var packet *Packet
		packet, err = RecvPackage(this.conn)
		if err != nil {
			// Log.Warn("网络错误 ServerConn %s", err.Error())
			Log.Warn("network error ServerConn %s", err.Error())
			break
		} else {
			packet.Session = this

			// Log.Debug("conn recv: %d %s %d\n%s", packet.MsgID, this.Ip, len(packet.Data)+len(packet.Dataplus)+16, hex.EncodeToString(append(packet.Data, packet.Dataplus...)))
			// Log.Debug("server conn recv: %d %s %d", packet.MsgID, this.Ip, len(packet.Data)+len(packet.Dataplus)+16)

			// if packet.IsWait {
			// 	Log.Debug("开始等待")
			// 	packet.IsWait = false
			// 	packet.WaitChan <- true
			// 	Log.Debug("开始执行")
			// 	<-packet.WaitChan
			// 	Log.Debug("执行完成")
			// } else {
			handler = engine.router.GetHandler(packet.MsgID)
			if handler == nil {
				Log.Warn("server The message is not registered, message number：%d", packet.MsgID)
				//					if this.packet.MsgID == 16 {
				//						fmt.Println(string(this.packet.Data))
				//					}
				//					break
			} else {
				//这里决定了消息是否异步处理
				go this.handlerProcess(handler, packet)
			}
			// }

			//				copy(this.cache, this.cache[this.packet.Size:this.cacheindex])
			//				this.cacheindex = this.cacheindex - uint32(n)

		}
	}

	this.Close()

	//最后一个包接收了之后关闭chan
	//如果有超时包需要等超时了才关闭，目前未做处理
	this.outChanCloseLock.Lock()
	this.outChanIsClose = true
	close(this.outChan)
	this.outChanCloseLock.Unlock()
	// fmt.Println("关闭连接")
}

func (this *ServerConn) loopSend() {
	var n int
	var err error
	var buff *[]byte
	var isClose = false
	var count = 5
	var total = 0
	for {
		buff, isClose = <-this.outChan
		if !isClose {
			Log.Warn("out chan is close")
			break
		}
		this.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
		n, err = this.conn.Write(*buff)
		if err != nil {
			total++
			Log.Warn("conn send err: %s", err.Error())
			if total > count {
				this.conn.Close()
			}
		} else {
			total = 0
			// Log.Debug("conn send: %d %s %d %d\n%s", msgID, this.conn.RemoteAddr(), len(*buff), n, hex.EncodeToString(*buff))
			// Log.Debug("client conn send: %d %s %d", msgID, this.conn.RemoteAddr(), len(*buff))
			// Log.Info("clent send %s", hex.EncodeToString(*buff))
		}
		if n < len(*buff) {
			// Log.Warn("conn send warn: %d %s length %d %d", msgID, this.conn.RemoteAddr().String(), len(*buff), n)
		}
	}
}

// func (this *ServerConn) Waite(du time.Duration) *Packet {
// 	if this.packet.Wait(du) {
// 		return &this.packet
// 	}
// 	return nil
// }

// func (this *ServerConn) FinishWaite() {
// 	this.packet.FinishWait()
// }

func (this *ServerConn) handlerProcess(handler MsgHandler, msg *Packet) {
	goroutineId := GetRandomDomain() + TimeFormatToNanosecondStr()
	_, file, line, _ := runtime.Caller(0)
	AddRuntime(file, line, goroutineId)
	defer DelRuntime(file, line, goroutineId)

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
	handler(this.controller, *msg)
	//消息处理后也要通过拦截器
	for i := itpsLen; i > 0; i-- {
		itps[i-1].Out(this.controller, *msg)
	}
}

//给客户端发送数据
func (this *ServerConn) Send(msgID uint64, data, dataplus *[]byte, waite bool) (err error) {
	this.outChanCloseLock.Lock()
	if this.outChanIsClose {
		this.outChanCloseLock.Unlock()
		return errors.New("send channel is close")
	}
	buff := MarshalPacket(msgID, data, dataplus)
	select {
	case this.outChan <- buff:
	default:
		addr := AddressNet([]byte(this.GetName()))
		Log.Warn("conn send err chan is full :%s", addr.B58String())
		this.conn.Close()
	}
	this.outChanCloseLock.Unlock()
	return
}

//给客户端发送数据
// func (this *ServerConn) Send(msgID uint64, data, dataplus *[]byte, waite bool) (err error) {
// 	defer PrintPanicStack()
// 	// this.packet.IsWait = waite
// 	buff := MarshalPacket(msgID, data, dataplus)
// 	var n int
// 	n, err = this.conn.Write(*buff)
// 	if err != nil {
// 		Log.Warn("conn send err: %s", err.Error())
// 	} else {
// 		// Log.Debug("conn send: %d %s %d %d\n%s", msgID, this.Ip, len(buff), n, hex.EncodeToString(buff))
// 		// Log.Debug("server conn send: %d %s %d", msgID, this.conn.RemoteAddr(), len(*buff))
// 	}
// 	if n < len(*buff) {
// 		Log.Warn("conn send warn: %d %s length %d %d", msgID, this.conn.RemoteAddr().String(), len(*buff), n)
// 	}
// 	return
// }

//给客户端发送数据
func (this *ServerConn) SendJSON(msgID uint64, data interface{}, waite bool) (err error) {
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

//关闭这个连接
func (this *ServerConn) Close() {
	Log.Info("Close session ServerConn")
	if this.engine.closecallback != nil {
		this.engine.closecallback(this.GetName())
	}
	this.engine.sessionStore.removeSession(this.GetName())
	err := this.conn.Close()
	if err != nil {
	}
}

func (this *ServerConn) SetName(name string) {
	this.engine.sessionStore.renameSession(this.name, name)
	this.name = name
}

//获取远程ip地址和端口
func (this *ServerConn) GetRemoteHost() string {
	return this.conn.RemoteAddr().String()
}
