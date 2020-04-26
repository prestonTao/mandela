package utils

import (
	"log"
	"net"
	"strconv"
	"strings"
)

/*
	获得一个TCP监听
*/
func GetTCPListener(ip string, port int) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ip+":"+strconv.Itoa(int(port)))
	if err != nil {
		// Log.Error("这个地址不符合规范：%s", ip+":"+strconv.Itoa(int(port)))
		return nil, err
	}
	var listener *net.TCPListener
	listener, err = net.ListenTCP("tcp4", tcpAddr)
	if err != nil {
		// Log.Error("监听一个地址失败：%s", ip+":"+strconv.Itoa(int(port)))
		// Log.Error("%v", err)
		return nil, err
	}
	// Log.Debug("监听一个地址：%s", ip+":"+strconv.Itoa(int(port)))
	// fmt.Println("监听一个地址：", ip+":"+strconv.Itoa(int(port)))
	// fmt.Println(ip + ":" + strconv.Itoa(int(port)) + "成功启动服务器")
	return listener, nil
}

/*
	获取本机能联网的ip地址
	@return    string    获得的ip地址
	@return    bool      是否能联网
*/
func GetLocalIntenetIp() (string, bool) {
	/*
	  获得所有本机地址
	  判断能联网的ip地址
	*/
	//TODO 修改测试地址
	conn, err := net.Dial("udp", "baidu.com:80")
	if err != nil {
		log.Println(err.Error())
		return "", false
	}
	ip := strings.Split(conn.LocalAddr().String(), ":")[0]
	conn.Close()
	return ip, true
}

/*
	不联网的情况下，得到本机ip地址
*/
func GetLocalHost() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	// for i, one := range addrs {
	// 	fmt.Println(i, one)
	// }
	return addrs[0].String()
}

/*
	是全球唯一ip
*/
func IsOnlyIp(ip string) bool {
	if ip == "0.0.0.0" || ip == "255.255.255.255" || ip == "127.0.0.1" || ip == "localhost" || ip == "" {
		return false
	}
	ips := strings.Split(ip, ".")
	//Class C 192.168.0.0-192.168.255.255
	if ips[0] == "192" && ips[1] == "168" {
		return false
	}
	//Class A 10.0.0.0-10.255.255.255
	if ips[0] == "10" {
		return false
	}
	//Class B 172.16.0.0-172.31.255.255
	if ips[0] == "172" {
		ipTwo, err := strconv.Atoi(ips[1])
		if err != nil {
			return false
		}
		start, end := 16, 31
		if ipTwo >= start && ipTwo <= end {
			return false
		}
	}
	return true
}

/*
	获得一个可用的UDP端口
*/
func GetAvailablePortForUDP() int {
	startPort := 9981
	for i := 0; i < 1000; i++ {
		_, err := net.ListenPacket("udp", "127.0.0.1:"+strconv.Itoa(startPort))
		if err != nil {
			startPort = startPort + 1
		} else {
			return startPort
		}
	}
	return 0
}

/*
	获得一个可用的TCP端口
*/
func GetAvailablePortForTCP(addr string) net.Listener {
	startPort := 9981
	for i := 0; i < 1000; i++ {
		lnr, err := net.Listen("tcp", addr+":"+strconv.Itoa(startPort))
		if err != nil {
			startPort = startPort + 1
		} else {
			return lnr
		}
	}
	return nil
}
