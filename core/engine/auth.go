package engine

import (
	"net"
	"strconv"
)

var defaultAuth Auth = new(NoneAuth)

type Auth interface {
	SendKey(conn net.Conn, session Session, name string) (remoteName string, err error)
	RecvKey(conn net.Conn, name string) (remoteName string, err error)
}

type NoneAuth struct {
	session int64
}

//发送
//@name                 本机服务器的名称
//@return  remoteName   对方服务器的名称
func (this *NoneAuth) SendKey(conn net.Conn, session Session, name string) (remoteName string, err error) {
	remoteName = name
	return
}

//接收
//@name                 本机服务器的名称
//@return  remoteName   对方服务器的名称
func (this *NoneAuth) RecvKey(conn net.Conn, name string) (remoteName string, err error) {
	this.session++
	// name = strconv.ParseInt(this.session, 10, )
	// name = strconv.Itoa(this.session)
	remoteName = strconv.FormatInt(this.session, 10)
	return
}
