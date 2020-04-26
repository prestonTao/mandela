package raft

import (
	// "fmt"
	"mandela/core/utils"
	"time"
)

var (
	TimeInterval = 5 //心跳间隔,单位秒
)

const (
	Task_class_raft_heart = "Task_class_raft_heart" //定时发送心跳信息
)

var task *utils.Task

func Init() {
	initTask()
}
func initTask() {
	task = utils.NewTask(taskFunc)
}

//加入定时器
func AddHeartTask(params string) {
	//fmt.Println("加入定时器...")
	task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_raft_heart, params)
}
func taskFunc(class, params string) {
	switch class {
	case Task_class_raft_heart:
		go func() {
			//fmt.Println(class, params)
			sendHeartTask(params)
			task.Add(time.Now().Unix()+int64(TimeInterval), Task_class_raft_heart, params)
		}()
	default:
		// fmt.Println("未注册的定时器类型", class)

	}
}
func sendHeartTask(teamidstr string) {
	teamid, err := utils.FromB58String(teamidstr)
	if err != nil {
		// fmt.Println("解析错误", err)
	}
	//只有leader能发送心跳信息
	team := CreateTeam(&teamid)
	if team.Role.Role == Leader {
		MulitiHeart(team)
	}
}
