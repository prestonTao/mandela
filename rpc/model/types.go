package model

import (
	"encoding/json"
)

//统一输出结构
type Result struct {
	Result interface{} `json:"result"`
}

//详情
type Getinfo struct {
	Netid          []byte `json:"netid"`          //网络版本号
	TotalAmount    uint64 `json:"TotalAmount"`    //发行总量
	Balance        uint64 `json:"balance"`        //总余额
	Testnet        bool   `json:"testnet"`        //是否是测试网络
	Blocks         uint64 `json:"blocks"`         //已经同步到的区块高度
	Group          uint64 `json:"group"`          //区块组高度
	StartingBlock  uint64 `json:"StartingBlock"`  //区块开始高度
	HighestBlock   uint64 `json:"HighestBlock"`   //所链接的节点的最高高度
	CurrentBlock   uint64 `json:"CurrentBlock"`   //已经同步到的区块高度
	PulledStates   uint64 `json:"PulledStates"`   //正在同步的区块高度
	BlockTime      uint64 `json:"BlockTime"`      //出块时间
	LightNode      uint64 `json:"LightNode"`      //轻节点押金数量
	CommunityNode  uint64 `json:"CommunityNode"`  //社区节点押金数量
	WitnessNode    uint64 `json:"WitnessNode"`    //见证人押金数量
	NameDepositMin uint64 `json:"NameDepositMin"` //域名押金最少金额
	AddrPre        string `json:"AddrPre"`        //地址前缀
}

//新地址
type GetNewAddress struct {
	Address string `json:"address"`
}

//帐号余额
type GetAccount struct {
	Balance float64 `json:"Balance"`
}

//成功
//type Success struct {
//	Success string `json:"Success"`
//}

func Tojson(data interface{}) ([]byte, error) {
	res, err := json.Marshal(Result{Result: data})
	return res, err
}
