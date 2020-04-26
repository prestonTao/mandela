package main

import (
	"fmt"
	"mandela/wallet/mining"
)

func main() {
	bt1 := mining.BallotTicket{
		Addr:        "a",    //地址
		Miner:       "hong", //矿工地址
		GroupHeight: 1,      //矿工组高度
	}
	mining.AddBallotTicket(&bt1)

	bt2 := mining.BallotTicket{
		Addr:        "b",    //地址
		Miner:       "hong", //矿工地址
		GroupHeight: 1,      //矿工组高度
	}
	mining.AddBallotTicket(&bt2)

	bt3 := mining.BallotTicket{
		Addr:        "c",    //地址
		Miner:       "hong", //矿工地址
		GroupHeight: 1,      //矿工组高度
	}
	mining.AddBallotTicket(&bt3)

	bt4 := mining.BallotTicket{
		Addr:        "c",    //地址
		Miner:       "hong", //矿工地址
		GroupHeight: 2,      //矿工组高度
	}
	mining.AddBallotTicket(&bt4)
	bt5 := mining.BallotTicket{
		Addr:        "c",    //地址
		Miner:       "hong", //矿工地址
		GroupHeight: 2,      //矿工组高度
	}
	mining.AddBallotTicket(&bt5)

	total := mining.FindTotal(2)

	fmt.Println(total)
}
