package raft

import (
	"mandela/core/nodeStore"
	"mandela/core/utils"
)

var (
	RD *RaftData
)

func RegisterRaft() {
	RD = NewRaftData()
}

//生成multihash
func BuildHash(key []byte) *utils.Multihash {
	return buildHash(key)
}

//创建一个team
//一个key对应一个team
func CreateTeam(teamid *utils.Multihash) *RaftTeam {
	return RD.CreateTeam(teamid)
}

//处理投票
func DoVote(teamid, nodeid *utils.Multihash) {
	team := GetTeam(teamid)
	if team != nil {
		tn := team.DoVote(nodeid)
		SaveTeam(tn)
	}
}

//获取TEAM
func GetTeam(teamid *utils.Multihash) *RaftTeam {
	return RD.GetTeam(teamid)
}

//保存team
func SaveTeam(team *RaftTeam) error {
	return RD.SaveTeam(team)
}
func DelTeam(teamid *utils.Multihash) error {
	return RD.DelTeam(teamid)
}

//获取当前角色
func GetRole(teamid, nodeid *utils.Multihash) string {
	raftteam := GetTeam(teamid)
	if raftteam.Role.Nodeid.B58String() == nodeid.B58String() {
		return raftteam.Role.Role
	} else {
		return Follower
	}
}

//首次加入网络
func FirstTeam(teamid *utils.Multihash) *RaftTeam {
	teamo := GetTeam(teamid)
	if teamo == nil {
		team := CreateTeam(teamid)
		team.Role.SetRole(nodeStore.NodeSelf.IdInfo.Id, Candidate)
		team.CreateVote(nodeStore.NodeSelf.IdInfo.Id)
		return team
	}
	return teamo
}
