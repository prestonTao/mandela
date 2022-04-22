package raft

import (
	"crypto/sha256"
	"errors"

	// "fmt"
	mc "mandela/core/message_center"
	"mandela/core/nodeStore"
	"mandela/core/utils"
)

//生成hash
func buildHash(key []byte) *utils.Multihash {
	hash := sha256.Sum256(key)
	bs, err := utils.Encode(hash[:], Version)
	if err != nil {
		// fmt.Println("buildhash:", err)
		return nil
	}
	has := utils.Multihash(bs)
	return &has
}

//获取1/4节点id
func getQuarterLogicIds(id *utils.Multihash) []*utils.Multihash {
	return nodeStore.GetQuarterLogicIds(id)
}

//广播投票消息
func MulitiVote(rt *RaftTeam) {
	ids := getQuarterLogicIds(rt.TeamId)
	for _, idpt := range ids {
		//fmt.Println(idpt.B58String())
		sendVote(idpt, rt.Json())
	}
}

//发送投票消息
func sendVote(id *utils.Multihash, data []byte) error {
	mhead := mc.NewMessageHead(id, id, false)
	voteid := mhead.RecvId
	mbody := mc.NewMessageBody(&data, "", nil, 0)
	message := mc.NewMessage(mhead, mbody)
	if message.Send(MSGID_RaftVote) {
		//fmt.Println("数据发送成功", id.B58String())
		bs := mc.WaitRequest(mc.CLASS_raftvote, utils.Bytes2string(message.Body.Hash))
		//fmt.Println("有消息返回", string(*bs))
		if bs == nil {
			// fmt.Println("发送投票数据消息失败，可能超时")
			return errors.New("发送投票数据消息失败，可能超时")
		}
		raftteam, err := Parse(*bs)
		if err != nil {
			// fmt.Println("raftteam parse error")
		}
		// fmt.Println("****************************")
		// fmt.Println("发起投票", voteid.B58String())
		// fmt.Println("当前角色：", raftteam.Role.Role)
		DoVote(raftteam.TeamId, voteid)
		return nil
	}
	//fmt.Println("数据发送失败", id.B58String())
	return nil
}

//发送心跳
func MulitiHeart(team *RaftTeam) {
	heartdata := HeartData{}
	heartdata.Teamid = team.TeamId
	heartdata.Index = team.Index
	ids := getQuarterLogicIds(team.TeamId)
	for _, idpt := range ids {
		sendHeart(idpt, heartdata.Json())
	}
}

//发送心跳消息
func sendHeart(id *utils.Multihash, data []byte) error {
	mhead := mc.NewMessageHead(id, id, false)
	mbody := mc.NewMessageBody(&data, "", nil, 0)
	message := mc.NewMessage(mhead, mbody)
	if message.Send(MSGID_RaftVoteHeart) {
		// fmt.Println("Leader ", string(data))
		// fmt.Println("发送心跳", id.B58String())
		return nil
	}
	//fmt.Println("数据发送失败", id.B58String())
	return nil
}
