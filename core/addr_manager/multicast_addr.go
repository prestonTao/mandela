package addr_manager

import (
	// "fmt"
	"mandela/core/config"
	"mandela/core/engine"
	"log"
	"net"
	"strconv"
	"time"
)

const (
	broadcastStartPort  = 8980
	broadcastServerPort = 9981 //广播服务器起始端口号
)

var (
	broadcastClientIsStart = false
	broadcastServerConn    *net.UDPConn //广播服务器
)

func init() {
	//	registerFunc(LoadByBroadcast)
}

/*
	启动一个局域网广播服务器
*/
func startBroadcastServer() {
	engine.Log.Debug("开始启动局域网广播服务器")

	var err error
	count := 10
	for i := 0; i < count; i++ {
		var addr *net.UDPAddr
		addr, err = net.ResolveUDPAddr("udp", config.Init_LocalIP+":"+strconv.Itoa(broadcastServerPort+i))
		if err != nil {
			// log.Panic(err)
			continue
		}
		broadcastServerConn, err = net.ListenUDP("udp", addr)
		if err != nil {
			// log.Panic(err)
			// engine.Log.Debug("开始启动局域网广播服务器")
			continue
		} else {
			break
		}
	}
	if err != nil {
		log.Panic("广播服务器启动失败")
		return
	}

	go func() {
		for {
			time.Sleep(time.Second * 3)
			// if len(Sys_superNodeEntry) == 0 {
			// 	continue
			// }
			if ip, port, err := GetSuperAddrOne(true); err == nil {
				for i := 0; i < 10; i++ {
					// fmt.P "255.255.255.255:" + strconv.Itoa(broadcastStartPort+i)
					udpaddr, err := net.ResolveUDPAddr("udp", "255.255.255.255:"+strconv.Itoa(broadcastStartPort+i))
					if err != nil {
						// fmt.Println("失败")
						continue
					}
					_, err = broadcastServerConn.WriteToUDP([]byte(ip+":"+strconv.Itoa(int(port))), udpaddr)
					if err != nil {
						// fmt.Println("广播失败")
						continue
					} else {
						// fmt.Println("广播成功")
					}
				}
			}
		}
	}()

}

/*
	关闭广播服务器
*/
func CloseBroadcastServer() {
	broadcastServerConn.Close()
}

/*
	通过组播方式获取地址列表
*/
func LoadByMulticast() {
	LoadByBroadcast()
}

/*
	通过广播获取地址
*/
func LoadByBroadcast() {
	if broadcastClientIsStart {
		engine.Log.Debug("局域网广播客户端正在运行")
		return
	}
	engine.Log.Debug("正在启动局域网广播客户端")
	// conns := make([]*net.UDPConn, 0)
	var conn *net.UDPConn
	//开始启动监听
	count := 10
	for i := 0; i < count; i++ {
		addr, err := net.ResolveUDPAddr("udp", config.Init_LocalIP+":"+strconv.Itoa(broadcastStartPort+i))
		if err != nil {
			log.Panic(err)
		}
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			// log.Panic(err)
			count++
			continue
		}
		engine.Log.Debug("局域网广播客户端启动成功，监听端口：%s", conn.LocalAddr().String())
		// conns = append(conns, conn)

		var b [512]byte
		go func() {
			for {
				n, _, err := conn.ReadFromUDP(b[:])
				if err != nil {
					// log.Panic(err)
					return
				}
				if n != 0 {
					// fmt.Printf("---%s\n", b[0:n])
					AddSuperPeerAddr(string(b[:n]))
				}
			}
		}()
		if conn != nil {
			break
		}
	}
	//启动失败
	if conn == nil {
		engine.Log.Debug("启动局域网广播客户端失败")
		return
	}
	// if len(conns) == 0 {
	// 	engine.Log.Debug("启动局域网广播客户端失败")
	// 	return
	// }
	broadcastClientIsStart = true
	go func() {
		c := make(chan string, 1)
		//		AddSubscribe(c)
		<-c
		engine.Log.Debug("开始关闭局域网广播客户端")
		broadcastClientIsStart = false
		conn.Close()
		// for _, one := range conns {
		// 	one.Close()
		// }
	}()

}
