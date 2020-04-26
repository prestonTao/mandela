package raft

import (
	// "fmt"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"math/rand"
	"strconv"
	"time"
)

var (
	HeartInterval = 15                         //心跳间隔,超过时间没收到，则视为断开
	HeartTime     = make(map[string]time.Time) //心跳时间,key为team leader
	HeartMin      = 1500                       //心跳检查间隔最小值 ms
	HeartMax      = 3000                       //心跳检查最大值 ms
)

func init() {
	go func() {
		for {
			ts := RandInt64(HeartMin, HeartMax)
			//fmt.Println("*********ts***********", ts)
			t, err := time.ParseDuration(ts)
			if err != nil {
				// fmt.Println("ts error", err)
				return
			}
			<-time.NewTicker(t).C
			checkHeartTime()
		}
	}()
}

//随机数
func RandInt64(min, max int) string {
	if min >= max || min == 0 || max == 0 {
		return strconv.Itoa(max)
	}
	num := rand.Intn(max-min) + min
	return strconv.Itoa(num) + "ms"
}

//更新心跳时间
func UpdateHeartTime(teamid *utils.Multihash) error {
	HeartTime[teamid.B58String()] = time.Now()
	return nil
}

//检查心跳时间
func checkHeartTime() {
	for k, v := range HeartTime {
		teamid, err := utils.FromB58String(k)
		if err != nil {
			// fmt.Println(err)
			break
		}
		team := CreateTeam(&teamid)
		//如果已经是leader,则不需检查心跳
		if team.Role.Role == Leader {
			break
		}
		if time.Now().Sub(v).Seconds() > float64(HeartInterval) {
			//只有candidate才能发起投票
			if team.Role.Role == Candidate {
				// fmt.Println("Leader 断开 teamid:", k)
				// fmt.Println("发起新的投票", k)
				team.CreateVote(nodeStore.NodeSelf.IdInfo.Id)
			} else {
				//设置当前节点为备选节点
				team.Role.SetRole(nodeStore.NodeSelf.IdInfo.Id, Candidate)
			}
			// fmt.Println("当前角色：", team.Role.Role)
		} else {
			// fmt.Println("Leader 正常 teamid:", k)
		}
	}
}
