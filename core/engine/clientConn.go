package engine

import (
	crand "crypto/rand"
	"errors"
	"math/big"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//本机向其他服务器的连接
type Client struct {
	sessionBase
	serverName       string
	ip               string
	port             uint32
	conn             net.Conn
	inPack           chan *Packet //接收队列
	outChan          chan *[]byte //发送队列
	outChanCloseLock *sync.Mutex  //是否关闭发送队列
	outChanIsClose   bool         //是否关闭发送队列。1=关闭
	// packet     Packet       //
	//	isClose    bool         //该连接是否被关闭
	isPowerful bool //是否是强连接，强连接有短线重连功能
	engine     *Engine
	controller Controller
}

func (this *Client) Connect(ip string, port uint32) (remoteName string, err error) {

	this.ip = ip
	this.port = port
	this.outChan = make(chan *[]byte, 10000)
	this.outChanCloseLock = new(sync.Mutex)
	this.outChanIsClose = false

	this.conn, err = net.Dial("tcp", ip+":"+strconv.Itoa(int(port)))
	if err != nil {
		return
	}

	//权限验证
	remoteName, err = defaultAuth.SendKey(this.conn, this, this.serverName)
	if err != nil {
		// this.conn.Close()
		this.Close()
		return
	}

	// fmt.Println("Connecting to", ip, ":", strconv.Itoa(int(port)))
	Log.Debug("Connecting to:%s:%s local:%s", ip, strconv.Itoa(int(port)), this.conn.LocalAddr().String())

	this.controller = &ControllerImpl{
		lock:       new(sync.RWMutex),
		engine:     this.engine,
		attributes: make(map[string]interface{}),
	}
	// this.packet.Session = this
	go this.loopSend()
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

		go this.loopSend()
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
			// Log.Warn("网络错误 Client %s", err.Error())
			Log.Warn("network error Client:%s", err.Error())
			break
		} else {
			packet.Session = this
			// Log.Debug("conn recv: %d %s %d\n%s", packet.MsgID, this.conn.RemoteAddr(), len(packet.Data)+len(packet.Dataplus)+16, hex.EncodeToString(append(packet.Data, packet.Dataplus...)))
			// Log.Debug("client conn recv: %d %s %d", packet.MsgID, this.conn.RemoteAddr(), len(packet.Data)+len(packet.Dataplus)+16)
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
				Log.Warn("client The message is not registered, message number:%d", packet.MsgID)
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

	// close(this.outChan)
	this.outChanCloseLock.Lock()
	this.outChanIsClose = true
	close(this.outChan)
	this.outChanCloseLock.Unlock()
	this.Close()
	if this.isPowerful {
		go this.reConnect()
	}
}

func (this *Client) loopSend() {
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

// func (this *Client) Send(msgID uint64, data, dataplus *[]byte, waite bool) (err error) {
// 	defer PrintPanicStack()
// 	// this.packet.IsWait = waite
// 	buff := MarshalPacket(msgID, data, dataplus)

// 	//设置写入超时时间为1秒钟内
// 	// this.conn.SetWriteDeadline(time.Now().Add(time.Second))
// 	var n int
// 	n, err = this.conn.Write(*buff)
// 	if err != nil {
// 		Log.Warn("conn send err: %s", err.Error())
// 	} else {
// 		// Log.Debug("conn send: %d %s %d %d\n%s", msgID, this.conn.RemoteAddr(), len(*buff), n, hex.EncodeToString(*buff))
// 		// Log.Debug("client conn send: %d %s %d", msgID, this.conn.RemoteAddr(), len(*buff))
// 		// Log.Info("clent send %s", hex.EncodeToString(*buff))
// 	}
// 	if n < len(*buff) {
// 		Log.Warn("conn send warn: %d %s length %d %d", msgID, this.conn.RemoteAddr().String(), len(*buff), n)
// 	}
// 	return
// }

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
	Log.Info("Close session ClientConn")
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

/*
	随机获取一个域名
*/
func GetRandomDomain() string {
	str := "abcdefghijklmnopqrstuvwxyz"
	// rand.Seed(int64(time.Now().Nanosecond()))
	result := ""
	r := int64(0)
	for i := 0; i < 8; i++ {
		r = GetRandNum(int64(25))
		result = result + str[r:r+1]
	}
	return result
}

/*
	获得一个随机数(0 - n]，包含0，不包含n
*/
func GetRandNum(n int64) int64 {
	if n <= 0 {
		return 0
	}
	result, _ := crand.Int(crand.Reader, big.NewInt(int64(n)))
	return result.Int64()
}

func TimeFormatToNanosecondStr() string {
	return time.Now().Format("20060102150405999999999")
}
