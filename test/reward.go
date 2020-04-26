package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func main() {
	example()

}

var max int64 = 100000

func example() {
	rewer := 100 //每块奖励100个币
	total := 30  //见证人总数
	group := 3   //每组见证人数量

	//准备n个见证人及投票
	a := make([]Witness, 0)
	for i := 0; i < total; i++ {
		a = append(a, NewWitness(i+1))
	}

	//计算预期收益
	allRewer := rewer * 1e8 * total
	allPos := big.NewInt(0)
	for _, w := range a {
		allPos = new(big.Int).Add(allPos, big.NewInt(int64(w.deposit)))
		for _, one := range w.vote {
			allPos = new(big.Int).Add(allPos, big.NewInt(int64(one)))
		}
	}
	for i, w := range a {
		temp := new(big.Int).Mul(big.NewInt(int64(allRewer)), big.NewInt(int64(w.deposit)))
		a[i].expect = new(big.Int).Div(temp, allPos)

		for i, one := range w.vote {
			temp := new(big.Int).Mul(big.NewInt(int64(allRewer)), big.NewInt(int64(one)))
			w.expects[i] = new(big.Int).Div(temp, allPos)
		}
	}

	//打印新创建的见证人及投票者
	for _, one := range a {
		fmt.Println("见证人押金-----", one.deposit, "预期收益", one.expect, "见证人收益", one.total)
		for i, voreOne := range one.vote {
			fmt.Println("投票者押金", voreOne, "预期收益", one.expects[i], "收益", one.balance[i])
		}
	}

	fmt.Println("11111111111 ---------------------------------------")

	//计算经过多少论出块后的收益
	cishu := 1000
	for x := 0; x < cishu; x++ {
		//随机排序
		newList := make([]Witness, 0)
		lastList := a
		for i := 0; i < total; i++ {
			index := GetRandNum(int64(len(lastList)))
			// fmt.Println("随机数", index, len(lastList))
			newList = append(newList, lastList[index])
			temp := lastList[:index]
			lastList = append(temp, lastList[index+1:]...)
		}

		for i := 0; i < total/group; i++ {
			//n个块所有奖励
			balanceTotal := rewer * 1e8 * group
			//计算本组所有股权
			guquanTotal := big.NewInt(0)
			for j := 0; j < group; j++ {
				w := newList[i*3+j]
				guquanTotal = new(big.Int).Add(guquanTotal, big.NewInt(int64(w.deposit)))
				for _, one := range w.vote {
					guquanTotal = new(big.Int).Add(guquanTotal, big.NewInt(int64(one)))
				}
			}
			//计算收益
			for j := 0; j < group; j++ {
				w := newList[i*3+j]
				temp := new(big.Int).Mul(big.NewInt(int64(balanceTotal)), big.NewInt(int64(w.deposit)))
				value := new(big.Int).Div(temp, guquanTotal)
				newList[i*3+j].total = new(big.Int).Add(w.total, value)

				for i, one := range w.vote {
					temp := new(big.Int).Mul(big.NewInt(int64(balanceTotal)), big.NewInt(int64(one)))
					value := new(big.Int).Div(temp, guquanTotal)
					w.balance[i] = new(big.Int).Add(w.balance[i], value)
				}

			}

		}
		a = newList
	}
	fmt.Println("22222222222 ---------------------------------------")
	for _, one := range a {
		fmt.Println("见证人押金-----", one.deposit, "预期收益", one.expect, "见证人收益", one.total)
		for i, voreOne := range one.vote {
			fmt.Println("投票者押金", voreOne, "预期收益", one.expects[i], "收益", one.balance[i])
		}
	}

	fmt.Println("end")
}

type Witness struct {
	deposit uint64     //见证人投票数量
	expect  *big.Int   //见证人预期收益
	total   *big.Int   //见证人收益
	vote    []uint64   //投票数量
	balance []*big.Int //收益
	expects []*big.Int //预期收益
}

func NewWitness(n int) Witness {
	w := Witness{
		deposit: 250000,
		expect:  new(big.Int),
		total:   new(big.Int),
		vote:    make([]uint64, 0),
		balance: make([]*big.Int, 0),
		expects: make([]*big.Int, 0),
	}
	for i := 0; i < n; i++ {
		value := GetRandNum(max)
		w.vote = append(w.vote, uint64(value+1))
		w.balance = append(w.balance, big.NewInt(0))
		w.expects = append(w.expects, big.NewInt(0))
	}
	return w
}

/*
	获得一个随机数(0 - n]，包含0，不包含n
*/
func GetRandNum(n int64) int {
	if n == 0 {
		return 0
	}

	result, _ := rand.Int(rand.Reader, big.NewInt(n))
	return int(result.Uint64())
	// fmt.Println(result)

	// rand.Seed(int64(time.Now().Nanosecond()))
	// return rand.Intn(n)
}
