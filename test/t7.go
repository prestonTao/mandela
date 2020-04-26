package main

import (
	"fmt"
)

const (
	year  = 20 //多少年挖完
	cycle = 40 //减产总周期数
)

func pow(x float64, n int) float64 {
	if x == 0 {
		return 0
	}
	result := calPow(x, n)
	if n < 0 {
		result = 1 / result
	}
	return result
}
func calPow(x float64, n int) float64 {
	if n == 0 {
		return 1
	}
	if n == 1 {
		return x
	}
	// 向右移动一位
	result := calPow(x, n>>1)
	result *= result

	// 如果n是奇数
	if n&1 == 1 {
		result *= x
	}
	return result
}
func main() {
	num := float64(3000000000)
	//计算X值 X*0.9`1+X*0.9`2...+X*0.9`40=3000000000  X*(0.9+0.9`2...+0.9`40)=3000000000
	p := float64(0)
	for i := 1; i <= cycle; i++ {
		p += pow(0.9, i)
	}
	X := num / p
	fmt.Println(X)
	perkuai := 60
	t := year * 365 * 24 * 60 * 60
	kuai := t / perkuai //总块数
	allnum := float64(0)
	for i := 1; i <= cycle; i++ {
		bi := X * pow(0.9, i) //每个周期产量,递减10%
		perbi := bi / float64(kuai)
		fmt.Printf("第%d阶段产出区块数 %d 每块奖励 %.8f 共产出 %.8f \n", i, kuai/cycle, perbi, bi)
		allnum += bi
	}
	fmt.Printf("%.8f", allnum)
	// fmt.Println(allnum)
}
