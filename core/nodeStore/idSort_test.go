package nodeStore

import (
	"fmt"
	"math/big"
	"sort"
	"testing"
)

func TestIdSort(t *testing.T) {
	// idStoreSimple1()
}

func idStoreSimple1() {

	desc := new(IdDESC)

	node1, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	*desc = append(*desc, node1)
	node2, _ := new(big.Int).SetString("31622036050853307757176718873676335712993063093791913422933189278586653352673", 10)
	*desc = append(*desc, node2)
	node3, _ := new(big.Int).SetString("38879061860890225964363770808076149471375052911854164467748691902681942298885", 10)
	*desc = append(*desc, node3)
	node4, _ := new(big.Int).SetString("59422813065590763321187925186011450884940934337897117431794152839561407098597", 10)
	*desc = append(*desc, node4)
	// node5, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)
	// node6, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)
	// node7, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)
	// node8, _ := new(big.Int).SetString("67491569314988856926507052272791838610626096514906525411496620109834031904600", 10)
	// desc = append(desc, node1)

	sort.Sort(desc)

	for _, id := range *desc {
		fmt.Println(id.String())
	}
}
