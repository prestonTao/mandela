package main

import (
	gconfig "mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"encoding/hex"
	"fmt"
	"math/big"
)

func main() {
	BuildIds()

	distance()

}

/*
	测试查询逻辑节点
*/
func BuildIds() {

	fmt.Println("---------- BuildIds ----------")

	ids := []string{

		"HCzFqZSNAZysyBazDQRtpFzy6eq7JigKVbQSqp3Zw1m",

		"BL9ZL1qX8vGRNUaZEq1E4svhjX1fiFm8iYJrMLHV7NTG",
		"8VmyBM5XmJRzPmQu7TC82TxdkX5BNc87sMaSfyft9fXW",
		"9eQ642R1zttsGw4ikdn5vEtG2f18GGiV2NBXL9gsszmM",
		"CsZqHcCVTukktiBDv9Cc1ELrhN7KVKwGnwzPNG2a9obV",

		"3yk5sZhd7o7PGT6i5muUXSCHDHxjoDAWuaMbqX1K1NXT",
		"EuNqmo354mxgizaZUgNkd59M2ho8nPWMs4QEN13W37Xo",
		"GKr849QTWzmhfMkpT4iatrvQNmREXBZrAkwHNmCDZx5m",
		"AxwnZMGyNsLSLHRB4nN2AvuFodtLZwYkCdC5L49beyEE",
	}

	for n := 0; n < len(ids); n++ {
		fmt.Println("本节点为", ids[n])
		index := n

		idMH := nodeStore.AddressFromB58String(ids[index])
		idsm := nodeStore.NewIds(idMH, gconfig.NodeIDLevel)
		for i, one := range ids {
			if i == index {
				continue
			}

			idMH := nodeStore.AddressFromB58String(one)
			idsm.AddId(idMH)
			//		ok, remove := idsm.AddId(&idMH)
			//		if ok {
			//			fmt.Println(one, remove)
			//		}
		}

		is := idsm.GetIds()
		for _, one := range is {

			idOne := nodeStore.AddressNet(one)

			fmt.Println("--逻辑节点", idOne.B58String())
		}

	}

}

/*
	计算节点距离
*/
func distance() {

	fmt.Println("---------- distance ----------")

	ids := []string{
		"W1aLWC4unTJZhSFc4VNLFsazAJ1PyTocV7agmteQDL3J3N",
		"W1gfVGa52yUJ4Gws4TiA9YbwGP8qCGgaYeeT8APjSiNk6U",
		"W1j9RJ1xYHaoAuRk2HGBrVA82njoxFAoctYKQMH43k8hXu",
		"W1atFt7bJ5Ubk4MXuV5GfsEYE7srWXR51exDgUEJcVr5fZ",
		"W1n9XtbLAjRsh9sr2kbwfkfy3VGenyhazbHJwrEYsnDZ8M",
	}

	index := 4

	kl := nodeStore.NewKademlia()
	for i, one := range ids {
		if i == index {
			continue
		}

		idMH, _ := utils.FromB58String(one)
		kl.Add(new(big.Int).SetBytes(idMH.Data()))

	}

	idMH, _ := utils.FromB58String(ids[index])
	is := kl.Get(new(big.Int).SetBytes(idMH.Data()))
	src := new(big.Int).SetBytes(idMH.Data())

	//	is := idsm.GetIds()
	for _, one := range is {
		tag := new(big.Int).SetBytes(one.Bytes())
		juli := tag.Xor(tag, src)

		bs, err := utils.Encode(one.Bytes(), gconfig.HashCode)
		if err != nil {
			fmt.Println("编码失败")
			continue
		}
		mh := utils.Multihash(bs)

		fmt.Println("排序结果", mh.B58String(), "距离", hex.EncodeToString(juli.Bytes()))
	}

}
