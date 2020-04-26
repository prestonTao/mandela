package raft

import (
	// "fmt"
	"mandela/core/engine"
	mc "mandela/core/message_center"
	"mandela/core/nodeStore"
	"mandela/core/utils"
)

func RaftVote(c engine.Controller, msg engine.Packet) {
	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		// fmt.Println(err)
		return
	}
	form, _ := utils.FromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//发送给自己的，自己处理
	if err := message.ParserContent(); err != nil {
		// fmt.Println(err)
		return
	}
	content := message.Body.Content
	raftdata, err := Parse(*content)
	if err != nil {
		// fmt.Println("数据解析错误", err)
		return
	}
	//follower收到投票信息，则初始化team
	team := CreateTeam(raftdata.TeamId)
	team.Role.SetRole(nodeStore.NodeSelf.IdInfo.Id, Follower)
	if GetRole(team.TeamId, nodeStore.NodeSelf.IdInfo.Id) == Follower {
		// fmt.Println("投票成功：当前角色 ", team.Role.Role)
	} else {
		// fmt.Println("投票失败：当前角色 ", team.Role.Role)
		content = nil
	}
	//当新发起的计数低于当前计数，则略过此次投票
	if raftdata.Index < team.Index {
		// fmt.Println("投票失败(计数低)：当前角色 ", team.Role.Role)
		content = nil
	}
	//回复给发送者
	mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	mbody := mc.NewMessageBody(content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	message = mc.NewMessage(mhead, mbody)
	message.Reply(MSGID_RaftVote_recv)
}
func RaftVote_recv(c engine.Controller, msg engine.Packet) {
	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		// fmt.Println("error  1", err)
		return
	}
	form, _ := utils.FromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//发送给自己的，自己处理
	if err := message.ParserContent(); err != nil {
		engine.NLog.Error(engine.LOG_file, "%s", err.Error())
		engine.NLog.Error(engine.LOG_file, "%s", string(msg.Dataplus))
		return
	}
	mc.ResponseWait(mc.CLASS_raftvote, message.Body.Hash.B58String(), message.Body.Content)
}

func RaftVoteHeart(c engine.Controller, msg engine.Packet) {
	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	if err != nil {
		// fmt.Println(err)
		return
	}
	form, _ := utils.FromB58String(msg.Session.GetName())
	if message.IsSendOther(&form) {
		return
	}
	//发送给自己的，自己处理
	if err := message.ParserContent(); err != nil {
		// fmt.Println(err)
		return
	}
	content := message.Body.Content
	heartdata, err := ParseHeartData(*content)
	if err != nil {
		// fmt.Println("心跳数据解析失败", err)
	}
	teamid := heartdata.Teamid
	// fmt.Println("收到心跳：teamid ", teamid.B58String())
	//更新心跳时间
	UpdateHeartTime(teamid)
	// fmt.Println("心跳时间", HeartTime)
	//收到心跳，则把状态设置为follower
	team := CreateTeam(teamid)
	team.Role.SetRole(nodeStore.NodeSelf.IdInfo.Id, Follower)
	//更新计数
	team.Index = heartdata.Index
	SaveTeam(team)
	// fmt.Println("计数:", team.Index)
	//回复给发送者
	//	mhead := mc.NewMessageHead(message.Head.Sender, message.Head.SenderSuperId, true)
	//	mbody := mc.NewMessageBody(content, message.Body.CreateTime, message.Body.Hash, message.Body.SendRand)
	//	message = mc.NewMessage(mhead, mbody)
	//	message.Reply(MSGID_RaftVote_recv)
}

//心跳返回
func RaftVoteHeart_recv(c engine.Controller, msg engine.Packet) {
	//	message, err := mc.ParserMessage(&msg.Data, &msg.Dataplus, msg.MsgID)
	//	if err != nil {
	//		fmt.Println("error  1", err)
	//		return
	//	}
	//	form, _ := utils.FromB58String(msg.Session.GetName())
	//	if message.IsSendOther(&form) {
	//		return
	//	}
	//	//发送给自己的，自己处理
	//	if err := message.ParserContent(); err != nil {
	//		engine.NLog.Error(engine.LOG_file, "%s", err.Error())
	//		engine.NLog.Error(engine.LOG_file, "%s", string(msg.Dataplus))
	//		return
	//	}
	//	mc.ResponseWait(mc.CLASS_raftvoteheart, message.Body.Hash.B58String(), message.Body.Content)
}
