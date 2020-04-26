package main

import (
	"fmt"
	"mandela/core/cache/raft"
)

func main() {
	raft.RegisterRaft()
	teamkey := []byte("fileinfo")
	teamid := raft.BuildHash(teamkey)
	fmt.Println("tm:", teamid.B58String())
	nodeid1 := raft.BuildHash([]byte("node1"))
	nodeid2 := raft.BuildHash([]byte("node2"))
	nodeid3 := raft.BuildHash([]byte("node3"))
	nodeid4 := raft.BuildHash([]byte("node4"))
	fmt.Println(nodeid1)
	team := raft.CreateTeam(teamid)
	raft.RD.Team.Range(func(k, v interface{}) bool {
		fmt.Println(k, v)
		return true
	})
	team.CreateVote(nodeid1)
	fmt.Printf("%+v\n", team)
	team.DoVote(nodeid1)
	team.DoVote(nodeid2)
	team.DoVote(nodeid3)
	team.DoVote(nodeid4)
	fmt.Printf("%+v\n", team.Vote)
	fmt.Printf("%+v\n", team.Role)
}
