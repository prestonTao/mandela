package main

import (
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/token/payment"
	"mandela/chain_witness_vote/mining/token/publish"
	"mandela/chain_witness_vote/mining/tx_name_in"
	"mandela/chain_witness_vote/mining/tx_name_out"

	// "mandela/chain_witness_vote/mining/tx_spaces_mining_in"
	// "mandela/chain_witness_vote/mining/tx_spaces_mining_out"
	"mandela/config"
	"mandela/core/engine"
	"mandela/rpc"
	"bytes"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	splitBlockHeight = 1

	// addressMax = 3000
)

func init() {
	tpc := new(payment.TokenPublishController)
	tpc.ActiveVoutIndex = new(sync.Map)
	mining.RegisterTransaction(config.Wallet_tx_type_token_payment, tpc)
	mining.RegisterTransaction(config.Wallet_tx_type_token_publish, new(publish.TokenPublishController))

	mining.RegisterTransaction(config.Wallet_tx_type_account, new(tx_name_in.AccountController))
	mining.RegisterTransaction(config.Wallet_tx_type_account_cancel, new(tx_name_out.AccountController))

	// mining.RegisterTransaction(config.Wallet_tx_type_spaces_mining_in, new(spaces_mining_in.AccountController))
	// mining.RegisterTransaction(config.Wallet_tx_type_spaces_mining_out, new(spaces_mining_out.AccountController))

}

// var payChan = make(chan *Address, 1000)

func main() {

	start()
}

func start() {
	peer1 := Peer{
		// Addr: "47.114.76.113:2080", //节点地址及端口，杭州
		// Addr: "47.106.221.172:2080", //节点地址及端口，深圳
		// Addr: "39.108.82.129:2080", //节点地址及端口，区块链1
		// Addr: "47.108.14.67:2080", //节点地址及端口，成都
		// Addr: "120.25.255.142:2080", //节点地址及端口，本地
		// Addr: "120.24.182.236:2080", //节点地址及端口，本地
		// Addr:       "127.0.0.1:5081",        //节点地址及端口，本地
		Addr:       "192.168.28.36:5081",    //节点地址及端口，本地
		AddressMax: 50000,                   //收款地址总数
		RPCUser:    "test",                  //rpc用户名
		RPCPwd:     "testp",                 //rpc密码
		WalletPwd:  "witness2",              //
		PayChan:    make(chan *Address, 10), //
	}

	info := peer1.GetInfo()
	fmt.Printf("%+v\n", info)
	startHeight := info.StartingBlock
	endHeight := info.CurrentBlock

	for i := startHeight; i <= endHeight; i++ {
		bhvos, _ := peer1.GetBlockAndTxProto(i, i)
		// fmt.Println(err)
		// bs, _ := bhvos[0].Json()
		// fmt.Println(string(*bs))
		// break
		for _, one := range bhvos {
			PrintBlock(one)
		}
	}

	return

}

type Fission struct {
	SrcAddress string //
	SrcIndex   int    //
	DstAddress string //
	DstIndex   int    //
	Fission    int    //是否已经裂变
}

type Peer struct {
	Addr       string        //节点地址及端口
	AddressMax uint64        //收款地址总数
	RPCUser    string        //rpc用户名
	RPCPwd     string        //rpc密码
	WalletPwd  string        //钱包支付密码
	PayChan    chan *Address //
}

func (this *Peer) Fission() {

}

func (this *Peer) StartPay(addr *Address) {
	go func() {
		this.PayChan <- addr
		this.TxPay(addr.AddrCoin, addr.Value-1, 1, addr.AddrCoin)
		<-this.PayChan
	}()
}

/*
	查询区块高度
*/
func (this *Peer) GetInfo() *Info {
	//{"method":"getinfo"}
	params := map[string]interface{}{
		"method": "getinfo",
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return nil
	}

	info := new(Info)

	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(info)
	return info
}

type Info struct {
	Netid          []byte `json:"netid"`          //网络版本号
	TotalAmount    uint64 `json:"TotalAmount"`    //发行总量
	Balance        uint64 `json:"balance"`        //可用余额
	BalanceFrozen  uint64 `json:"BalanceFrozen"`  //冻结的余额
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

/*
	判断地址数量，解析地址及余额
*/
func GetAddressList(peer Peer) (uint64, []Address) {
	fmt.Println("判断地址数量，解析地址及余额")
	//{"method":"listaccounts"}
	params := map[string]interface{}{
		"method": "listaccounts",
	}
	result := Post(peer.Addr, peer.RPCUser, peer.RPCPwd, params)
	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return 0, nil
	}

	addrs := make([]Address, 0)

	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(&addrs)

	// fmt.Sprintf("+v%", addrs)

	return uint64(len(addrs)), addrs
}

type Address struct {
	Index       int
	AddrCoin    string
	Value       uint64
	ValueFrozen uint64
	Type        int
}

/*
	创建新地址
*/
func (this *Peer) CreatNewAddr() bool {
	fmt.Println("创建新地址")
	//{"method":"getnewaddress","params":{"password":"123456789"}}
	paramsChild := map[string]interface{}{
		"password": this.WalletPwd,
	}
	params := map[string]interface{}{
		"method": "getnewaddress",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	_, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return false
	}
	return true
}

/*
	交易
*/
func (this *Peer) TxPay(srcAddress string, amount, gas uint64, address string) bool {
	//{"method":"sendtoaddress","params":{"address":"1Hy62rv8BDypQgLpGeUPwGQzVkkkBBU8v22",
	//"amount":1000000000000000,"gas":100000000,"pwd":"123456789","comment":"test"}}

	// fmt.Println("转账", srcAddress, amount, gas, address)

	paramsChild := map[string]interface{}{
		"srcaddress": srcAddress,
		"amount":     amount,
		"address":    address,
		"gas":        gas,
		"pwd":        this.WalletPwd,
	}
	params := map[string]interface{}{
		"method": "sendtoaddress",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	if result == nil {
		// fmt.Println("请求错误")
		return false
	}
	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return false
	}
	if result == nil {
		fmt.Println("转账出错", srcAddress, amount, gas, address)
	} else {
		if result.Code != 2000 {
			fmt.Println(result.Code, result.Message, string(bs))
		}
	}
	return true
}

/*
	给多人转账
*/
func (this *Peer) TxPayMore(payNumber []PayNumber, gas uint64) bool {
	//{"method":"sendtoaddress","params":{"address":"1Hy62rv8BDypQgLpGeUPwGQzVkkkBBU8v22",
	//"amount":1000000000000000,"gas":100000000,"pwd":"123456789","comment":"test"}}

	fmt.Println("多人转账", payNumber, gas)

	paramsChild := map[string]interface{}{
		"addresses": payNumber,
		"gas":       gas,
		"pwd":       this.WalletPwd,
	}
	params := map[string]interface{}{
		"method": "sendtoaddressmore",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return false
	}
	fmt.Println(result.Code, result.Message, string(bs))
	return true
}

/*
	多人转账
*/
type PayNumber struct {
	Address string `json:"address"` //转账地址
	Amount  uint64 `json:"amount"`  //转账金额
}

/*
	查询指定账户余额
*/
func (this *Peer) GetAccount(address string) *Account {
	//{"method":"sendtoaddress","params":{"address":"1Hy62rv8BDypQgLpGeUPwGQzVkkkBBU8v22",
	//"amount":1000000000000000,"gas":100000000,"pwd":"123456789","comment":"test"}}
	paramsChild := map[string]interface{}{
		"address": address,
	}
	params := map[string]interface{}{
		"method": "getaccount",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return nil
	}

	account := new(Account)
	buf := bytes.NewBuffer(bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(&account)

	return account
}

/*
	获取指定高度的区块及交易
*/
func (this *Peer) GetBlockAndTx(start, end uint64) ([]mining.BlockHeadVO, error) {
	//{"method":"sendtoaddress","params":{"address":"1Hy62rv8BDypQgLpGeUPwGQzVkkkBBU8v22",
	//"amount":1000000000000000,"gas":100000000,"pwd":"123456789","comment":"test"}}

	fmt.Println("获取指定高度的区块及交易", start, end)

	paramsChild := map[string]interface{}{
		"startHeight": start,
		"endHeight":   end,
	}
	params := map[string]interface{}{
		"method": "getblocksrange",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)

	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误0", err.Error())
		return nil, err
	}

	// fmt.Println(string(bs))

	bhvos := make([]mining.BlockHeadVO, 0)

	//
	resultMap := make([]map[string]interface{}, 0)
	err = json.Unmarshal(bs, &resultMap)
	if err != nil {
		fmt.Println("序列化错误1111", err.Error())
		return nil, err
	}
	for _, one := range resultMap {
		bhItr := one["bh"]
		bs, err := json.Marshal(bhItr)
		if err != nil {
			fmt.Println("序列化错误1", err.Error())
			return nil, err
		}
		bh := new(mining.BlockHead)
		err = json.Unmarshal(bs, bh)
		if err != nil {
			fmt.Println("序列化错误2", err.Error())
			return nil, err
		}

		bhvo := new(mining.BlockHeadVO)
		bhvo.BH = bh
		bhvo.Txs = make([]mining.TxItr, 0)

		txsItr := one["txs"]
		txsMap := make([]map[string]interface{}, 0)
		bs, err = json.Marshal(txsItr)
		if err != nil {
			fmt.Println("序列化错误3", err.Error())
			return nil, err
		}
		err = json.Unmarshal(bs, &txsMap)
		if err != nil {
			fmt.Println("序列化错误4", err.Error())
			return nil, err
		}
		for _, two := range txsMap {
			bs, _ = json.Marshal(two)

			// txOne, _ := json.Unmarshal()

			txOne, _ := mining.ParseTxBaseProto(0, &bs) //ParseTxBase(0, &bs)
			bhvo.Txs = append(bhvo.Txs, txOne)
		}
		bhvos = append(bhvos, *bhvo)
	}

	// fmt.Println(result.Code, result.Message, string(bs))
	return bhvos, nil
}

/*
	获取指定高度的区块及交易
*/
func (this *Peer) GetBlockAndTxProto(start, end uint64) ([]mining.BlockHeadVO, error) {
	//{"method":"sendtoaddress","params":{"address":"1Hy62rv8BDypQgLpGeUPwGQzVkkkBBU8v22",
	//"amount":1000000000000000,"gas":100000000,"pwd":"123456789","comment":"test"}}

	fmt.Println("获取指定高度的区块及交易", start, end)

	paramsChild := map[string]interface{}{
		"startHeight": start,
		"endHeight":   end,
	}
	params := map[string]interface{}{
		"method": "getblocksrangeproto",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)

	bs, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误0", err.Error())
		return nil, err
	}

	// fmt.Println(string(bs))

	bhvos := make([]mining.BlockHeadVO, 0, end-start+1)

	// bs, err = json.Marshal(result.Result)
	resultArray := make([]*[]byte, 0)
	json.Unmarshal(bs, &resultArray)

	// result.Result.([][]interface{})
	for _, one := range resultArray {
		// bsOne := one.([]byte)
		// fmt.Println(one)
		bhvo, _ := mining.ParseBlockHeadVOProto(one)
		bhvos = append(bhvos, *bhvo)
	}
	// fmt.Println(resultArray)

	// //
	// resultMap := make([]map[string]interface{}, 0)
	// err = json.Unmarshal(bs, &resultMap)
	// if err != nil {
	// 	fmt.Println("序列化错误1111", err.Error())
	// 	return nil, err
	// }
	// for _, one := range resultMap {
	// 	bhItr := one["bh"]
	// 	bs, err := json.Marshal(bhItr)
	// 	if err != nil {
	// 		fmt.Println("序列化错误1", err.Error())
	// 		return nil, err
	// 	}
	// 	bh := new(mining.BlockHead)
	// 	err = json.Unmarshal(bs, bh)
	// 	if err != nil {
	// 		fmt.Println("序列化错误2", err.Error())
	// 		return nil, err
	// 	}

	// 	bhvo := new(mining.BlockHeadVO)
	// 	bhvo.BH = bh
	// 	bhvo.Txs = make([]mining.TxItr, 0)

	// 	txsItr := one["txs"]
	// 	txsMap := make([]map[string]interface{}, 0)
	// 	bs, err = json.Marshal(txsItr)
	// 	if err != nil {
	// 		fmt.Println("序列化错误3", err.Error())
	// 		return nil, err
	// 	}
	// 	err = json.Unmarshal(bs, &txsMap)
	// 	if err != nil {
	// 		fmt.Println("序列化错误4", err.Error())
	// 		return nil, err
	// 	}
	// 	for _, two := range txsMap {
	// 		bs, _ = json.Marshal(two)

	// 		// txOne, _ := json.Unmarshal()

	// 		txOne, _ := mining.ParseTxBaseProto(0, &bs) //ParseTxBase(0, &bs)
	// 		bhvo.Txs = append(bhvo.Txs, txOne)
	// 	}
	// 	bhvos = append(bhvos, *bhvo)
	// }

	// fmt.Println(result.Code, result.Message, string(bs))
	return bhvos, nil
}

//帐号余额
type Account struct {
	Balance       uint64 `json:"Balance"`
	BalanceFrozen uint64 `json:"BalanceFrozen"`
}

/*
	自定义请求head，body，method，参数用body传递
	获取添加金币记录
*/
func Post(addr, rpcUser, rpcPwd string, params map[string]interface{}) *PostResult {
	url := "/rpc"
	method := "POST"

	header := http.Header{"user": []string{rpcUser}, "password": []string{rpcPwd}}
	client := &http.Client{}
	//req, err := http.NewRequest("GET", "http://www.baidu.com/", nil)
	bs, err := json.Marshal(params)
	req, err := http.NewRequest(method, "http://"+addr+url, strings.NewReader(string(bs)))
	if err != nil {
		fmt.Println("创建request错误")
		return nil
	}
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求服务器错误", err.Error())
		return nil
	}
	// fmt.Println("response:", resp.StatusCode)

	var resultBs []byte

	if resp.StatusCode == 200 {
		resultBs, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("读取body内容错误")
			return nil
		}
		// fmt.Println(string(resultBs))

		result := new(PostResult)

		buf := bytes.NewBuffer(resultBs)
		decoder := json.NewDecoder(buf)
		decoder.UseNumber()
		err = decoder.Decode(result)
		return result
	}
	fmt.Println("StatusCode:", resp.StatusCode)
	return nil
}

type PostResult struct {
	Jsonrpc string      `json:"jsonrpc"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

/*
	获得一个随机数(0 - n]，包含0，不包含n
*/
func GetRandNum(n int64) int64 {
	if n == 0 {
		return 0
	}
	result, _ := crand.Int(crand.Reader, big.NewInt(int64(n)))
	return result.Int64()
}

type PayPlan struct {
	SrcIndex int    //转出账户索引地址
	DstIndex int    //转入账户索引地址
	Value    uint64 //转账额度
}

/*
	打印区块内容
*/
func PrintBlock(bhVO mining.BlockHeadVO) {
	bh := bhVO.BH
	engine.Log.Info("===============================================")
	engine.Log.Info("第%d个块 %s 交易数量 %d", bh.Height, hex.EncodeToString(bh.Hash), len(bh.Tx))
	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
	}
	bhvo := rpc.BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		GroupHeightGrowth: bh.GroupHeightGrowth,                     //
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	}
	bs, _ := json.Marshal(bhvo)
	engine.Log.Info(string(bs))

	for _, one := range bhVO.Txs {
		// engine.Log.Info("%v", one)
		engine.Log.Info("-----------------------------------------------")
		engine.Log.Info("交易id " + string(hex.EncodeToString(*one.GetHash())))
		// one.GetVOJSON()
		txItr := one.GetVOJSON()
		bs, _ := json.Marshal(txItr)
		engine.Log.Info(string(bs))
	}

}
