package core

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	version = 1
)

type Auth struct {
}

/*
+++++++++++++++++++++++++++++++++++++++++++++++++++++++
| version   | ctp        | size      | name           |
+++++++++++++++++++++++++++++++++++++++++++++++++++++++
| 版本       | 连接类型    | 数据长度    | 连接名称         |
+++++++++++++++++++++++++++++++++++++++++++++++++++++++
| 2 byte    | 2 byte     | 4 byte    |                |
+++++++++++++++++++++++++++++++++++++++++++++++++++++++

version：版本
	1：第一个版本

ctp：连接类型
	1：带name的连接
	2：不带name的连接

name：连接名称
	区分每一个客户端的名称

*/

//发送
//@name                 本机服务器的名称
//@return  remoteName   对方服务器的名称
func (this *Auth) SendKey(conn net.Conn, session engine.Session, name string) (remoteName string, err error) {
	// fmt.Println("主动连接")
	//向对方发送网络id
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, uint32(engine.Netid))
	//	var n int
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		//		fmt.Println("主动连接错误 11111", n, buf.Bytes(), err)
		return "", err
	}
	// fmt.Println("发送了网络id 成功", n, engine.Netid)

	//第一次连接，向对方发送自己的Node
	node := &nodeStore.Node{
		IdInfo:  nodeStore.NodeSelf.IdInfo,
		IsSuper: false, //自己是否是超级节点，对方会判断，这里只需要虚心的说自己不是超级节点
		Addr:    nodeStore.NodeSelf.Addr,
		TcpPort: nodeStore.NodeSelf.TcpPort,
	}
	bs := node.Marshal()
	buf = bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, uint16(len(bs)))
	_, err = buf.Write(bs)
	if err != nil {
		//		fmt.Println("写入node size错误 22222", n, err)
		return "", err
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		//		fmt.Println("写入node错误 22222.5", n, err)
		return "", err
	}
	//接收对方的Node
	sizebs := make([]byte, 2)
	_, err = io.ReadFull(conn, sizebs)
	if err != nil {
		//		fmt.Println(time.Now().Format("2006-01-02 15:04:05.999999999"), "接收对方node size错误 33333", n, err)
		engine.Log.Error("netid different,self node:%d", engine.Netid)
		return "", Error_different_netid
	}
	size := binary.BigEndian.Uint16(sizebs)
	nodeBs := make([]byte, size)
	_, err = io.ReadFull(conn, nodeBs)
	if err != nil {
		// fmt.Println("接收对方node错误 44444", err)
		return "", err
	}
	node, err = nodeStore.ParseNode(nodeBs)
	if err != nil {
		// fmt.Println("解析对方node错误 55555", err)
		return "", err
	}
	if !nodeStore.CheckIdInfo(node.IdInfo) {
		//非法的 idinfo
		return "", errors.New("illegal idinfo")
	}

	//检查这个链接是否已经存在
	remoteName = node.IdInfo.Id.B58String()
	if _, ok := engine.GetSession(remoteName); ok {
		//send 这个链接已经存在
		err = errors.New("send This link already exists")
		return
	}

	//获取对方ip地址
	node.Addr = strings.Split(conn.RemoteAddr().String(), ":")[0]
	// fmt.Println("SendKey", strings.Split(conn.RemoteAddr().String(), ":")[0], conn.RemoteAddr().Network())

	// fmt.Println("添加一个node", node.IdInfo.Id.B58String())
	//能直接通过ip地址访问的节点，一定是超级节点。
	node.IsSuper = true
	nodeStore.AddNode(node)
	// fmt.Println("添加一个node之后", node.IdInfo.Id.B58String())
	//接收对方判断自己是否是超级节点
	isSuperBs := make([]byte, 2)
	_, err = io.ReadFull(conn, isSuperBs)
	if err != nil {
		// fmt.Println("接收对方判断自己是否是超级节点错误", err)
		return "", err
	}
	// fmt.Println("是否卡在这里", node.IdInfo.Id.B58String())
	isSuperInt := binary.BigEndian.Uint16(isSuperBs)
	//	isSuperInt = 1
	if isSuperInt == 1 {
		nodeStore.NodeSelf.IsSuper = true
	} else {
		nodeStore.NodeSelf.IsSuper = false
	}

	err = nil
	// fmt.Println("连接到新的节点", remoteName)
	return
}

//接收
//name   自己的名称
//@return  remoteName   对方服务器的名称
func (this *Auth) RecvKey(conn net.Conn, name string) (remoteName string, err error) {
	// fmt.Println("接受连接", conn.RemoteAddr(), name)
	//接受连接
	// engine.Log.Error("Accept connection %s", conn.RemoteAddr())
	//接收对方网络id
	netIdBs := make([]byte, 4)
	for i := 0; i < 5; i++ {
		_, err = io.ReadFull(conn, netIdBs)
		if err != nil {

			// engine.Log.Error("连接错误 %s", err.Error())
			//			fmt.Println(err.Error())
			//判断连接是否有被远程断开
			if _, ok := err.(*net.OpError); ok {
				// engine.Log.Error("对方已经断开网络 %+v", e)

				return "", err
			}
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	if err != nil {
		// engine.Log.Error("2连接错误 %s", err.Error())
		//		fmt.Println("222", err.Error())
		//接收对方netid错误
		return "", err
	}
	netId := binary.BigEndian.Uint32(netIdBs)
	if netId != engine.Netid {
		// engine.Log.Error("网络id不相同 %d %d", netId, engine.Netid)
		//网络id不相同，本节点:%d 远程节点:%d
		return "", errors.New(fmt.Sprintf("netid different,self node:%d remote node:%d", engine.Netid, netId))
	}

	//接收对方的Node
	sizebs := make([]byte, 2)
	_, err = io.ReadFull(conn, sizebs)
	if err != nil {
		// engine.Log.Error("接收对方node size错误 22222")
		return "", err
	}
	size := binary.BigEndian.Uint16(sizebs)
	nodeBs := make([]byte, size)
	_, err = io.ReadFull(conn, nodeBs)
	if err != nil {
		// engine.Log.Error("接收对方node错误 33333")
		return "", err
	}
	// fmt.Println(string(nodeBs))
	node, err := nodeStore.ParseNode(nodeBs)
	if err != nil {
		// engine.Log.Error("111 %s", err.Error())
		return "", err
	}
	//检查地址是不是安全地址
	//	if !nodeStore.CheckSafeAddr(node.IdInfo.Puk) {
	//		fmt.Println("000", errors.New("idinfo非安全地址"))
	//		return "", errors.New("idinfo非安全地址")
	//	}
	//验证s256生成的地址
	if !nodeStore.CheckIdInfo(node.IdInfo) {
		//		engine.Log.Error("222 %s", "非法的 idinfo")
		//非法的 idinfo
		return "", errors.New("Illegal idinfo")
	}
	//若对方网络地址和自己的一样，那么断开连接
	if bytes.Equal(node.IdInfo.Id, nodeStore.NodeSelf.IdInfo.Id) {
		//		engine.Log.Error("333 自己连接自己，断开连接")
		//自己连接自己，断开连接
		return "", errors.New("Connect yourself, disconnect yourself")
	}

	//检查这个链接是否已经存在
	remoteName = node.IdInfo.Id.B58String()
	if _, ok := engine.GetSession(remoteName); ok {
		//		engine.Log.Error("444 这个链接已经存在 %s", remoteName)
		//这个链接已经存在
		err = errors.New("recv This link already exists")
		//return
	}

	//	engine.Log.Info("检查这个地址是否在网络中已经存在 111")

	// //检查这个地址是否在网络中已经存在
	// mid := GetNodeMachineID(&node.IdInfo.Id)
	// engine.Log.Info("检查这个地址是否在网络中已经存在 222 %d", mid)

	// if mid != 0 && node.MachineID != mid {
	// 	engine.Log.Info("不能用相同的节点地址连接到网络")
	// 	return "", errors.New("不能用相同的节点地址连接到网络")
	// }
	// engine.Log.Info("这个地址在网络中不存在 %d", mid)

	//给对方发送自己的Node
	bs := nodeStore.NodeSelf.Marshal()
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, uint16(len(bs)))
	_, err = buf.Write(bs)
	if err != nil {
		// fmt.Println("连接错误 44444")
		return "", err
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		// fmt.Println("连接错误 55555")
		return "", err
	}

	//获取对方ip地址
	node.Addr, _, err = net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		// fmt.Println("获取对方ip地址错误", err.Error())
		return "", err
	}
	// fmt.Println("解析到对方ip地址为", node.Addr)

	// node.Addr = strings.Split(conn.RemoteAddr().String(), ":")[0]
	// fmt.Println("RecvKey", strings.Split(conn.RemoteAddr().String(), ":")[0], conn.RemoteAddr().Network())

	//连接自己，又说自己是超级节点的，直接断开连接
	if node.IsSuper {
		// fmt.Println("连接错误 66666")
		//这是一个验证是否有公网ip地址的超级节点的连接
		err = errors.New("This is a super node connection to verify whether there is a public IP address")
		return
	}
	//如果是局域网地址，尝试局域网连接
	// if !utils.IsOnlyIp(node.Addr) && TryConn(node) {

	// 	nodeStore.AddLanNode(node)

	// 	buf = bytes.NewBuffer(nil)
	// 	binary.Write(buf, binary.BigEndian, uint16(0))
	// 	nodeStore.AddProxyNode(node)
	// 	_, err = conn.Write(buf.Bytes())
	// 	if err != nil {
	// 		fmt.Println("连接错误 77777")
	// 		return "", err
	// 	}
	// 	select {
	// 	case isOnline <- true:
	// 	default:
	// 	}
	// 	return

	// }

	//判断对方是否是超级节点
	if config.NetType == config.NetType_release {
		// fmt.Println("网络类型是release")
		if !utils.IsOnlyIp(node.Addr) {
			// fmt.Println("不是公网ip地址")
			buf = bytes.NewBuffer(nil)
			binary.Write(buf, binary.BigEndian, uint16(0))
			nodeStore.AddProxyNode(node)
			_, err = conn.Write(buf.Bytes())
			if err != nil {
				// fmt.Println("连接错误 77777")
				return "", err
			}

			//不是公网ip地址
			err = errors.New("Not a public IP address")
			select {
			case isOnline <- true:
			default:
			}
			return
		}
	}

	// fmt.Println("判断对方是不是超级节点")
	//判断对方是否能链接上
	isSuper := TryConn(node)
	node.IsSuper = isSuper
	// fmt.Println("对方是不是超级节点", isSuper)

	buf = bytes.NewBuffer(nil)
	if isSuper {
		binary.Write(buf, binary.BigEndian, uint16(1))
		nodeStore.AddNode(node)
	} else {
		binary.Write(buf, binary.BigEndian, uint16(0))
		nodeStore.AddProxyNode(node)
	}
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		// fmt.Println("连接错误 77777")
		return "", err
	}

	err = nil
	// fmt.Println("有新的连接", remoteName)
	//	fmt.Println("auth end")
	//发送节点上线信号
	select {
	case isOnline <- true:
	default:
	}
	return
}

/*
	通过名称字符串获得bytes
	@name   要序列化的name字符串
*/
func GetBytesForName(name string) []byte {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.BigEndian, int32(len(name)))
	buf.Write([]byte(name))
	return buf.Bytes()
}

/*
	通过读连接中的bytes获取name字符串
*/
func GetNameForConn(conn net.Conn) (name string, err error) {
	lenghtByte := make([]byte, 4)
	io.ReadFull(conn, lenghtByte)
	nameLenght := binary.BigEndian.Uint32(lenghtByte)
	nameByte := make([]byte, nameLenght)
	if n, e := conn.Read(nameByte); e != nil {
		err = e
		return
	} else {
		//得到对方名称
		name = string(nameByte[:n])
		return
	}
}

/*
	尝试去连接一个ip地址，判断对方是否是超级节点
*/
func TryConn(srcNode *nodeStore.Node) bool {
	//设置3秒钟超时
	conn, err := net.DialTimeout("tcp", srcNode.Addr+":"+strconv.Itoa(int(srcNode.TcpPort)), time.Second*3)
	if err != nil {
		return false
	}

	//向对方发送网络id
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, uint32(engine.Netid))
	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return false
	}

	//第一次连接，向对方发送自己的Node
	node := &nodeStore.Node{
		IdInfo:  nodeStore.NodeSelf.IdInfo,
		IsSuper: true,
		Addr:    nodeStore.NodeSelf.Addr,
		TcpPort: nodeStore.NodeSelf.TcpPort,
	}
	bs := node.Marshal()
	buf = bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, uint16(len(bs)))
	_, err = buf.Write(bs)
	if err != nil {
		return false
	}
	_, err = conn.Write(buf.Bytes())

	//接收对方的Node
	sizebs := make([]byte, 2)
	_, err = io.ReadFull(conn, sizebs)
	if err != nil {
		return false
	}
	size := binary.BigEndian.Uint16(sizebs)
	nodeBs := make([]byte, size)
	_, err = io.ReadFull(conn, nodeBs)
	if err != nil {
		return false
	}
	node, err = nodeStore.ParseNode(nodeBs)
	if err != nil {
		return false
	}
	if !nodeStore.CheckIdInfo(node.IdInfo) {
		return false
	}

	//检查这个链接是否已经存在
	//	remoteName = node.IdInfo.Id.B58String()
	//	if _, ok := engine.GetSession(remoteName); ok {
	//		err = errors.New("这个链接已经存在")
	//		return
	//	}

	//获取对方ip地址
	node.Addr = strings.Split(conn.RemoteAddr().String(), ":")[0]
	//	fmt.Println("SendKey", strings.Split(conn.RemoteAddr().String(), ":")[0], conn.RemoteAddr().Network())
	return true
}
