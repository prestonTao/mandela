package engine

// import (
// 	"net"
// )

// //其他计算机对本机的连接
// type GnetConn struct {
// 	sessionBase
// 	conn           net.Conn
// 	Ip             string
// 	Connected_time string
// 	CloseTime      string
// 	// packet         Packet
// 	engine     *Engine
// 	controller Controller
// }

// func (this *GnetConn) run() {

// 	// this.packet.Session = this

// 	go this.recv()
// }

// //接收客户端消息协程
// func (this *GnetConn) recv() {
// 	defer PrintPanicStack()
// 	//处理客户端主动断开连接的情况
// 	var err error
// 	var handler MsgHandler
// 	for {
// 		var packet *Packet
// 		packet, err = RecvPackage(this.conn)
// 		if err != nil {
// 			// Log.Warn("网络错误 ServerConn %s", err.Error())
// 			Log.Warn("network error ServerConn %s", err.Error())
// 			break
// 		} else {
// 			packet.Session = this

// 			// Log.Debug("conn recv: %d %s %d\n%s", this.packet.MsgID, this.Ip, len(this.packet.Data)+16, hex.EncodeToString(this.packet.Data))
// 			// Log.Debug("server conn recv: %d %s %d", packet.MsgID, this.Ip, len(packet.Data)+16)

// 			// if packet.IsWait {
// 			// 	Log.Debug("开始等待")
// 			// 	packet.IsWait = false
// 			// 	packet.WaitChan <- true
// 			// 	Log.Debug("开始执行")
// 			// 	<-packet.WaitChan
// 			// 	Log.Debug("执行完成")
// 			// } else {
// 			handler = engine.router.GetHandler(packet.MsgID)
// 			if handler == nil {
// 				Log.Warn("server The message is not registered, message number：%d", packet.MsgID)
// 				//					if this.packet.MsgID == 16 {
// 				//						fmt.Println(string(this.packet.Data))
// 				//					}
// 				//					break
// 			} else {
// 				//这里决定了消息是否异步处理
// 				go this.handlerProcess(handler, packet)
// 			}
// 			// }

// 			//				copy(this.cache, this.cache[this.packet.Size:this.cacheindex])
// 			//				this.cacheindex = this.cacheindex - uint32(n)

// 		}
// 	}

// 	this.Close()

// 	//最后一个包接收了之后关闭chan
// 	//如果有超时包需要等超时了才关闭，目前未做处理
// 	// close(this.outData)
// 	// fmt.Println("关闭连接")
// }

// // func (this *ServerConn) Waite(du time.Duration) *Packet {
// // 	if this.packet.Wait(du) {
// // 		return &this.packet
// // 	}
// // 	return nil
// // }

// // func (this *ServerConn) FinishWaite() {
// // 	this.packet.FinishWait()
// // }

// func (this *GnetConn) handlerProcess(handler MsgHandler, msg *Packet) {
// 	//消息处理模块报错将不会引起宕机
// 	defer PrintPanicStack()
// 	//消息处理前先通过拦截器
// 	itps := this.engine.interceptor.getInterceptors()
// 	itpsLen := len(itps)
// 	for i := 0; i < itpsLen; i++ {
// 		isIntercept := itps[i].In(this.controller, *msg)
// 		//
// 		if isIntercept {
// 			return
// 		}
// 	}
// 	handler(this.controller, *msg)
// 	//消息处理后也要通过拦截器
// 	for i := itpsLen; i > 0; i-- {
// 		itps[i-1].Out(this.controller, *msg)
// 	}
// }

// //给客户端发送数据
// func (this *GnetConn) Send(msgID uint64, data, dataplus []byte, waite bool) (err error) {
// 	defer PrintPanicStack()
// 	// this.packet.IsWait = waite
// 	buff := MarshalPacket(msgID, data, dataplus)
// 	var n int
// 	n, err = this.conn.Write(buff)
// 	if err != nil {
// 		Log.Warn("conn send err: %s", err.Error())
// 	} else {
// 		// Log.Info("server send %s", hex.EncodeToString(*buff))
// 		// Log.Debug("server conn send: %d %s %d", msgID, this.conn.RemoteAddr(), len(*buff))

// 		// Log.Debug("conn send: %d %s %d %d\n%s", msgID, this.Ip, len(*buff), n, hex.EncodeToString(*buff))
// 	}
// 	if n < len(buff) {
// 		Log.Warn("conn send warn: %d %s length %d %d", msgID, this.conn.RemoteAddr().String(), len(buff), n)
// 	}
// 	return
// }

// //给客户端发送数据
// func (this *GnetConn) SendJSON(msgID uint64, data interface{}, waite bool) (err error) {
// 	defer PrintPanicStack()
// 	// this.packet.IsWait = waite
// 	var f []byte
// 	f, err = json.Marshal(data)
// 	if err != nil {
// 		return
// 	}
// 	buff := MarshalPacket(msgID, f, nil)
// 	_, err = this.conn.Write(buff)
// 	return
// }

// //关闭这个连接
// func (this *GnetConn) Close() {
// 	if this.engine.closecallback != nil {
// 		this.engine.closecallback(this.GetName())
// 	}
// 	this.engine.sessionStore.removeSession(this.GetName())
// 	err := this.conn.Close()
// 	if err != nil {
// 	}
// }

// func (this *GnetConn) SetName(name string) {
// 	this.engine.sessionStore.renameSession(this.name, name)
// 	this.name = name
// }

// //获取远程ip地址和端口
// func (this *GnetConn) GetRemoteHost() string {
// 	return this.conn.RemoteAddr().String()
// }
