package role

import (
	"../dao"
	"code.google.com/p/goprotobuf/proto"
	acc "common/message/m_acc"
	"common/net"
	server "gameServerEngine"
	// "common/net/net_server"
	"fmt"
)

type Role struct {
	DataModule *dao.DataModule
}

func (this *Role) Init() {
	this.DataModule = new(dao.DataModule)
	// this.DataModule.InitDB()
}

func (this *Role) Register(c server.Controller, msg net.Packet) {

}
func (this *Role) Login(c server.Controller, msg net.Packet) {
	//反序列化
	recvMsg := &acc.LoginPlayer_MA{}
	proto.Unmarshal(msg.Date, recvMsg)
	// param := this.DataModule.Accountlogin(recvMsg)
	//计算数据库id和表id
	idbid, itbid := GetDatabaseId(*recvMsg.SzAccount)
	fmt.Println("biaoge id ", itbid)
	fmt.Println("dbid ", idbid)
	//得到数据库连接
	conn := c.GetDBConn(int32(idbid))
	fmt.Println("得到数据库的连接")
	param := this.DataModule.AccountloginToo(recvMsg, int32(itbid), conn)
	fmt.Println("执行完存储过程")
	fmt.Println(param)
	repMsg := new(acc.LoginPlayer_AM)
	repMsg.RetValue = &param.Iretval
	repMsg.Accid = &param.Accid
	repMsg.ServerID = &param.Serverid
	repMsg.GmLevel = &param.Gmlevel
	repMsg.GamePoints = &param.Points
	isFcm := true
	repMsg.BIsUnderAge = &isFcm
	repMsg.EnterTimes = &param.Ifcmtime

	// repMsg.SzAccount = &param.SzAccount

	// repMsg.

	// 	// msg.iRetValue = iretval
	//  //   msg.qwAccID = params.uiaccid
	//  //   msg.wGMLevel = params.uigmlevel
	//  //   msg.dwGamePoints = params.uipoints
	//  //   msg.ucIsUnderAge = params.ucisfcm
	//  //   msg.dwEntherTime = params.ifcmtime
	//  //   msg.wOnlineCnt = 1

	// SzAccountEx      *string `protobuf:"bytes,2,opt,name=szAccountEx" json:"szAccountEx,omitempty"`
	// Accid            *uint64 `protobuf:"varint,4,opt,name=accid" json:"accid,omitempty"`
	// ServerID         *uint64 `protobuf:"varint,5,opt,name=serverID" json:"serverID,omitempty"`
	// GmLevel          *uint32 `protobuf:"varint,6,opt,name=gmLevel" json:"gmLevel,omitempty"`
	// BIsUnderAge      *bool   `protobuf:"varint,8,opt,name=bIsUnderAge" json:"bIsUnderAge,omitempty"`
	// EnterTimes       *uint64 `protobuf:"varint,9,opt,name=enterTimes" json:"enterTimes,omitempty"`
	// OnLineCnt        *uint32 `protobuf:"varint,10,opt,name=onLineCnt" json:"onLineCnt,omitempty"`
	// Conn             *uint64 `protobuf:"varint,11,opt,name=conn" json:"conn,omitempty"`
}
func (this *Role) LoginOut(c server.Controller, msg net.Packet) {

}
