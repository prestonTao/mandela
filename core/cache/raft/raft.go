package raft

import (
	"mandela/core/utils"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	Leader    = "leader"    //管理者
	Follower  = "follower"  //追随者
	Candidate = "candidate" //候选者
)

type RaftData struct {
	Team sync.Map //根据KEY的HASH值生成TEAM TeamId:RaftTeam
}

func NewRaftData() *RaftData {
	return &RaftData{}
}

//创建team
func (rd *RaftData) CreateTeam(teamid *utils.Multihash) *RaftTeam {
	teami, ok := rd.Team.Load(utils.Bytes2string(teamid))
	if !ok {
		nodeids := getQuarterLogicIds(teamid)
		teamnew := RaftTeam{TeamId: teamid, Nodeids: nodeids, Role: &RaftRole{}, Vote: make(map[string]int)}
		rd.Team.Store(utils.Bytes2string(teamid), teamnew)
		return &teamnew
	}
	team := teami.(RaftTeam)
	return &team
}

//获取team
func (rd *RaftData) GetTeam(teamid *utils.Multihash) *RaftTeam {
	teami, ok := rd.Team.Load(utils.Bytes2string(teamid))
	if ok {
		team := teami.(RaftTeam)
		return &team
	}
	return nil
}

//获删除team
func (rd *RaftData) DelTeam(teamid *utils.Multihash) error {
	rd.Team.Delete(utils.Bytes2string(teamid))
	return nil
}

//保存team
func (rd *RaftData) SaveTeam(team *RaftTeam) error {
	rd.Team.Store(utils.Bytes2string(team.TeamId), *team)
	return nil
}

//角色
type RaftRole struct {
	Nodeid *utils.Multihash //节点id
	Role   string
}

//设置角色
func (rr *RaftRole) SetRole(nodeid *utils.Multihash, role string) {
	rr.Nodeid = nodeid
	rr.Role = role
}

//team
//每个team有唯一的leader
type RaftTeam struct {
	TeamId  *utils.Multihash   //key HASH
	Nodeids []*utils.Multihash //成员节点
	Role    *RaftRole          //角色 NodeId:RaftRole
	Vote    map[string]int     //投票结果
	Index   uint64             //投票计数，发起新一轮投票,则计算+1
}

//发起投票
func (rt *RaftTeam) CreateVote(nodeid *utils.Multihash) error {
	rt.Vote[utils.Bytes2string(nodeid)] = 1
	//广播投票信息,只有备选者才能发起投票
	if rt.Role.Role == Candidate {
		MulitiVote(rt)
	}
	return nil
}

//投票
func (rt *RaftTeam) DoVote(nodeid *utils.Multihash) *RaftTeam {
	rt.Vote[utils.Bytes2string(nodeid)] = 1
	result := float32(len(rt.Vote)) / float32(len(rt.Nodeids))
	// fmt.Println("投票结果:", len(rt.Vote), len(rt.Nodeids), result)
	if result >= 0.5 {
		//设置为leader
		rt.Role.SetRole(rt.Role.Nodeid, Leader)
		//计数+1
		rt.Index++
		//发送心跳，通知已当选
		MulitiHeart(rt)
		//加入心跳定时器
		AddHeartTask(utils.Bytes2string(rt.TeamId))
	}
	return rt
}
func (rt *RaftTeam) Json() []byte {
	res, err := json.Marshal(rt)
	if err != nil {
		// fmt.Println(err)
		return nil
	}
	return res
}

//解析字节为raftdata
func Parse(data []byte) (*RaftTeam, error) {
	cd := new(RaftTeam)
	// err := json.Unmarshal(data, cd)
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err := decoder.Decode(cd)
	if err != nil {
		// fmt.Println("RaftTeam:", err)
		return cd, err
	}
	return cd, nil
}

//心跳数据
type HeartData struct {
	Teamid *utils.Multihash
	Index  uint64 //当前team 投票计数
}

func (hd *HeartData) Json() []byte {
	res, err := json.Marshal(hd)
	if err != nil {
		// fmt.Println(err)
		return nil
	}
	return res
}

//解析字节为heartdata
func ParseHeartData(data []byte) (*HeartData, error) {
	ht := new(HeartData)
	// err := json.Unmarshal(data, ht)
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	err := decoder.Decode(ht)
	if err != nil {
		// fmt.Println("RaftTeam:", err)
		return ht, err
	}
	return ht, nil
}
