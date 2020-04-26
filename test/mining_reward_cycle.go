package main

import (
	"mandela/config"
	"fmt"
	"strconv"
)

const (
	// unit = 100000000
	year       = 60 * 60 * 24 * 365 //一年秒数
	year20     = year * 20          //20年时间秒数
	total      = 3000000000         //奖励总量
	blockTime  = 60                 //区块间隔时间
	totalBlack = year20 / blockTime //所有周期共产出块数量
	sysle      = 40                 //产出周期数量
	reduce     = 0.9                //每个周期减少比例
)

func main() {
	height := uint64(1)
	interval := uint64(12614400)
	for {
		reward := config.ClacRewardForBlockHeight(height)
		fmt.Println("区块高度", height, "奖励数量", reward)
		if reward == 0 {
			break
		}
		height += interval
	}

	fmt.Println("奖励数量", config.ClacRewardForBlockHeight(1))
	fmt.Println("奖励数量", config.ClacRewardForBlockHeight(12614400))

	clacCycleForTotalAndInterval(1170000000*1e8, 4)
	//	jisuan()
}

/*
	给定总量和减半时间，计算周期数量
*/
func clacCycle(first, total, intervalYear uint64) {
	count := uint64(0)
	cycle := 0
	for count < total {
		cycle++
		oneCycle := 2880 * 365 * intervalYear * first
		fmt.Printf("第 %d 个周期,共产出 %d \n", cycle, oneCycle)
		count += oneCycle
		first = first / 2
		if first == 0 {
			break
		}
	}
	fmt.Printf("共产出 %d", count)
}

/*
	通过总量和减半周期计算
*/
func clacCycleForTotalAndInterval(total, intervalYear uint64) {
	first := total / 2 / (2880 * 365 * intervalYear)
	fmt.Printf("首块奖励 %d\n", first)
	clacCycle(first, total, intervalYear)

}

func jisuan() {
	//第20个周期总共奖励多少钱
	ARewardTotal := float64(total) / float64(sysle)
	//第20个周期产出多少个块
	//每个周期产出多少块
	tb := float64(totalBlack) / float64(sysle)

	A := float64(ARewardTotal*2) / 1.9

	a := strconv.FormatFloat(A, 'f', -1, 64)
	fmt.Println(a)

	tr := float64(0)

	fmt.Println("总量", strconv.FormatFloat(total, 'f', -1, 64))
	fmt.Println("20年总产出区块", strconv.FormatFloat(totalBlack, 'f', -1, 64))
	fmt.Println("40个周期，每个周期产出区块", strconv.FormatFloat(tb, 'f', -1, 64))

	for i := 19; i > 0; i-- {
		total := A
		for j := 0; j < i; j++ {
			total = total / 0.9
		}
		x := total / tb
		fmt.Println("第", 20-i, "个周期 共产出区块", tb, " 每个块奖励", x, " 共产出Token", strconv.FormatFloat(total, 'f', -1, 64))
		tr = tr + total
	}
	fmt.Println("------")

	for i := 0; i <= 20; i++ {
		total := A
		for j := 0; j < i; j++ {
			total = total * 0.9
		}
		x := total / tb
		fmt.Println("第", i+20, "个周期 共产出区块", tb, " 每个块奖励", x, " 共产出Token", strconv.FormatFloat(total, 'f', -1, 64))
		tr = tr + total
	}
	fmt.Println("共产出", strconv.FormatFloat(tr, 'f', -1, 64))

}
