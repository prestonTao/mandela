package main

import (
	"mandela/chain_witness_vote/mining"
	"mandela/protos/go_protos"
	"encoding/hex"
	"fmt"
	"time"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	example2()
}

func example2() {
	bhBs := []byte(`{"H":"N6pqivbhI7GfoqjOXpsOh9LteCCPD02OtWFE3leGSlo=","Ht":402581,"GH":378428,"GHG":0,"Pbh":"Ahp3t6NN4nJpg1/noovnh3RSLlyT3BY/JRLUH3kAV74=","Nbh":null,"NTx":0,"M":"","Tx":[],"T":1636040073,"W":"TU1TlwvzU2AN8Tq06iOocRbHMjWm88cjsQh1Aw==","s":"Ip/ZZQ6iZH3F6Ju9tnq6yZ0Qm9kNM6teWSen8VAFCDL9oGZaLDgA7CYQNAAoliyMAXSvgwSVXiw/nec23iSFBA=="}
2021/11/04 23:34:56.303 [1;34m[I][0m [handlers.go:461]  neighbor 6LnDyB3bNQjgiFY9SuT4S3MGZZMEMgtiy99cAB4i59f5 find next block 402581 hash nil. hight:402583
{"H":"N6pqivbhI7GfoqjOXpsOh9LteCCPD02OtWFE3leGSlo=","Ht":402581,"GH":378428,"GHG":0,"Pbh":"Ahp3t6NN4nJpg1/noovnh3RSLlyT3BY/JRLUH3kAV74=","Nbh":null,"NTx":0,"M":"","Tx":[],"T":1636040073,"W":"TU1TlwvzU2AN8Tq06iOocRbHMjWm88cjsQh1Aw==","s":"Ip/ZZQ6iZH3F6Ju9tnq6yZ0Qm9kNM6teWSen8VAFCDL9oGZaLDgA7CYQNAAoliyMAXSvgwSVXiw/nec23iSFBA=="}`)

	total := 1000000

	start := time.Now()
	for i := 0; i < total; i++ {
		bh := new(mining.BlockHead)
		json.Unmarshal(bhBs, bh)
	}
	fmt.Println("json", time.Now().Sub(start))

	start = time.Now()
	for i := 0; i < total; i++ {
		bh := new(go_protos.BlockHead)
		proto.Unmarshal(bhBs, bh)
		bh.Marshal()
	}
	fmt.Println("proto", time.Now().Sub(start))

	start = time.Now()
	for i := 0; i < total; i++ {
		bh := new(go_protos.BlockHead)
		gogoproto.Unmarshal(bhBs, bh)
		bh.Marshal()
		// gogoproto.Marshal(bh)
	}
	fmt.Println("gogo proto", time.Now().Sub(start))

	//序列化后的大小
	bhold := new(mining.BlockHead)
	json.Unmarshal(bhBs, bhold)
	bs, _ := json.Marshal(bhold)
	fmt.Println("json 序列化后大小", len(bs), hex.EncodeToString(bs))

	bh := new(go_protos.BlockHead)
	proto.Unmarshal(bhBs, bh)
	bs, _ = proto.Marshal(bh)
	fmt.Println("proto 序列化后大小", len(bs), hex.EncodeToString(bs))

	bs, _ = gogoproto.Marshal(bh)
	fmt.Println("gogoproto 序列化后大小", len(bs), hex.EncodeToString(bs))
}
