package main

import (
	"mandela/config"
	"mandela/core/engine"
	"fmt"
	"strconv"
	"time"
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

	cycleRange(790617, 790620)
	cycleRange(790621, 790624)
	return

	fmt.Println(float64(32 / 31))
	// height := uint64(30001)
	// interval := uint64(12614400)
	start := time.Now()
	// reward := config.ClacRewardForBlockHeight(height)
	// fmt.Println("区块高度", 1, "奖励数量", config.ClacRewardForBlockHeight(1))
	// fmt.Println("区块高度", 30000, "奖励数量", config.ClacRewardForBlockHeight(30000))
	// fmt.Println("区块高度", 30001, "奖励数量", config.ClacRewardForBlockHeight(30001))
	// fmt.Println("区块高度", 60000, "奖励数量", config.ClacRewardForBlockHeight(60000))
	// fmt.Println("区块高度", 60001, "奖励数量", config.ClacRewardForBlockHeight(60001))
	// fmt.Println("区块高度", 90000, "奖励数量", config.ClacRewardForBlockHeight(90000))
	// fmt.Println("区块高度", 90001, "奖励数量", config.ClacRewardForBlockHeight(90001))
	// fmt.Println("区块高度", 120000, "奖励数量", config.ClacRewardForBlockHeight(120000))
	// fmt.Println("区块高度", 120001, "奖励数量", config.ClacRewardForBlockHeight(120001))
	// fmt.Println("区块高度", 184999, "奖励数量", config.ClacRewardForBlockHeight(184999))
	// fmt.Println("区块高度", 185000, "奖励数量", config.ClacRewardForBlockHeight(185000))

	// fmt.Println("区块高度", 1409475, "奖励数量", config.ClacRewardForBlockHeight(1409475))

	// fmt.Println("区块高度", 15190000, "奖励数量", config.ClacRewardForBlockHeight(15190000))
	fmt.Println("区块高度", 15190001, "奖励数量", config.ClacRewardForBlockHeight(15190002))
	// fmt.Println("区块高度", 100000000, "奖励数量", config.ClacRewardForBlockHeight(100000000))

	jiange := time.Now().Sub(start)
	fmt.Println("耗时", jiange.Seconds(), jiange.Milliseconds(), jiange.Microseconds(), jiange.Nanoseconds())

	// return

	height := uint64(1)
	interval := uint64(1)
	total := uint64(0)
	for {
		reward := config.ClacRewardForBlockHeight(height)
		//整数亿块高度打印一次日志
		if height%3153600 == 0 {
			engine.Log.Info("第%d年 区块高度 %d 奖励数量 %d 奖励积累量 %d", height/3153600, height, reward, total)
			// jiange = time.Now().Sub(start)
			// engine.Log.Info("耗时 %d %d %d %d", jiange.Seconds(), jiange.Milliseconds(), jiange.Microseconds(), jiange.Nanoseconds())
		}
		// fmt.Println("区块高度", height, "奖励数量", reward)
		// engine.Log.Info("区块高度 %d 奖励数量 %d", height, reward)
		total += reward

		// if height == 31536000 {
		// 	break
		// }
		if reward <= 0 {
			break
		}
		height += interval
	}

	// fmt.Println("奖励数量", config.ClacRewardForBlockHeight(1))
	// fmt.Println("奖励数量", config.ClacRewardForBlockHeight(12614400))
	// fmt.Println("高度", height, "总共发放", total)
	engine.Log.Info("区块高度 %d 总共发放 %d", height, total)
	jiange = time.Now().Sub(start)
	engine.Log.Info("耗时 %d %d %d %d", jiange.Seconds(), jiange.Milliseconds(), jiange.Microseconds(), jiange.Nanoseconds())

	// 区块高度 5734933 奖励数量 654032235527
	// 区块高度 5734934 奖励数量 654032348107
	// 区块高度 5734935 奖励数量 654032460687
	// 区块高度 5734936 奖励数量 654032573267
	// 区块高度 5734937 奖励数量 654032685847
	// 区块高度 5734938 奖励数量 654032798427
	// 区块高度 5734939 奖励数量 654032911007
	// 区块高度 5734940 奖励数量 654033023587
	// 区块高度 5734941 奖励数量 654033136167
	// 区块高度 5734942 奖励数量 654033248747
	// 区块高度 5734943 奖励数量 654033361327
	// 区块高度 5734944 奖励数量 654033473907
	// 区块高度 5734945 奖励数量 654033586487
	// 区块高度 5734946 奖励数量 654033699067
	// 区块高度 5734947 奖励数量 654033811647
	// 区块高度 5734948 奖励数量 654033924227
	// 区块高度 5734949 奖励数量 654034036807
	// 区块高度 5734950 奖励数量 654034149387
	// 区块高度 5734951 奖励数量 654034261967
	// 区块高度 5734952 奖励数量 654034374547
	// 区块高度 5734953 奖励数量 654034487127
	// 区块高度 5734954 奖励数量 654034599707
	// 区块高度 5734955 奖励数量 654034712287
	// 区块高度 5734956 奖励数量 654034824867
	// 区块高度 5734957 奖励数量 654034937447
	// 区块高度 5734958 奖励数量 654035050027
	// 区块高度 5734959 奖励数量 654035162607
	// 区块高度 5734960 奖励数量 654035275187
	// 区块高度 5734961 奖励数量 654035387767
	// 区块高度 5734962 奖励数量 654035500347
	// 区块高度 5734963 奖励数量 654035612927
	// 区块高度 5734964 奖励数量 654035725507
	// 区块高度 5734965 奖励数量 654035838087
	// 区块高度 5734966 奖励数量 654035950667
	// 区块高度 5734967 奖励数量 654036063247
	// 区块高度 5734968 奖励数量 654036175827
	// 区块高度 5734969 奖励数量 654036288407
	// 区块高度 5734970 奖励数量 0
	// 总共发放 1728131548942504703

}

func exampleold() {
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

func cycleRange(start, end uint64) {
	total := uint64(0)
	for i := start; i < end+1; i++ {
		rewardOne := config.ClacRewardForBlockHeight(i)
		fmt.Println("奖励数量", rewardOne)
		total += rewardOne
	}
	fmt.Println("总共奖励", total)
}
