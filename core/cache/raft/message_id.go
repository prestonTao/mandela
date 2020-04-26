package raft

import (
	"mandela/core/engine"
)

var (
	Version = uint64(0x00)
)

const (
	MSGID_RaftVote           = 3001 //投票同步
	MSGID_RaftVote_recv      = 3002 //投票同步 返回
	MSGID_RaftVoteHeart      = 3003 //心跳同步
	MSGID_RaftVoteHeart_recv = 3004 //心跳同步 返回
)

func Register() {
	//初始化Cache
	RegisterRaft()
	//注册消息ID
	engine.RegisterMsg(MSGID_RaftVote, RaftVote)
	engine.RegisterMsg(MSGID_RaftVote_recv, RaftVote_recv)
	engine.RegisterMsg(MSGID_RaftVoteHeart, RaftVoteHeart)
	engine.RegisterMsg(MSGID_RaftVoteHeart_recv, RaftVoteHeart_recv)
	//定时同步数据
	initTask()
	//绑定CacheData模块
	bindCacheData()
}
