package raft

import (
	"time"
	cached "mandela/core/cache/cachedata"
	"mandela/core/utils"
)

func init() {
	go func() {
		for {
			<-time.NewTicker(5 * time.Second).C
			RD.Team.Range(func(k, v interface{}) bool {
				team := v.(RaftTeam)
				if team.Role.Role == Leader {
					syncData(team.TeamId)
				}
				return true
			})
		}
	}()
}

//初始化cachedata
func bindCacheData() {
	cached.Register()
}

//同步数据
func syncData(teamid *utils.Multihash) {
	cached.SyncDataToQuarterLogicIds(teamid)
}

//增加数据
func Add(key, value []byte) {
	cachedata := cached.BuildCacheData(key, value)
	//cachedata.AddOwnId(nodeStore.NodeSelf.IdInfo.Id)
	//保存数据
	cached.Save(cachedata)
	teamid := BuildHash(key)
	//广播出去
	go cached.SyncDataToQuarterLogicIds(teamid)
	FirstTeam(teamid)
}

//修改数据
func Update(key, value []byte) {
	cachedata := cached.BuildCacheData(key, value)
	//保存数据
	cached.Save(cachedata)
	teamid := BuildHash(key)
	//广播出去
	go cached.SyncDataToQuarterLogicIds(teamid)
}

//删除数据
func Del(key []byte) {
	teamid := BuildHash(key)
	cachedata := cached.GetCacheDataByHash(teamid)
	if cachedata == nil {
		return
	}
	cachedata.Del = true
	//广播出去
	go cached.SyncCacheDataToQuarterLogicIds(cachedata)
	cached.Del(key)
}
