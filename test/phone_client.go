package main

import (
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/tx_name_in"
	"mandela/config"
	"mandela/rpc/model"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Config struct {
	RpcHost string //节点地址及端口号
	RpcUser string //远程调用用户名
	RpcPwd  string //远程调用密码
}

func main() {
	conf := Config{
		RpcHost: "127.0.0.1:5081", //节点地址及端口号
		RpcUser: "test",           //远程调用用户名
		RpcPwd:  "testp",          //远程调用密码
	}
	getinfoParams := Info{
		Method: "getinfo",
	}
	localHeight := uint64(0)
	remoteHeight := uint64(0)
	for {

		//获取远程节点最新高度
		result := HttpPost(conf, getinfoParams)
		fmt.Println(string(result))

		resultVO := new(RpcResult)
		json.Unmarshal(result, resultVO)
		bs, _ := json.Marshal(resultVO.Result)

		getinfo := new(model.Getinfo)
		json.Unmarshal(bs, getinfo)

		fmt.Println("最新高度", getinfo.CurrentBlock)
		remoteHeight = getinfo.CurrentBlock

		//循环拉取区块，直到本地高度和远程高度一样
		for localHeight < remoteHeight {
			params := map[string]interface{}{
				"startHeight": localHeight + 1,
				"endHeight":   localHeight + 1,
			}
			getBlockParams := Info{
				Method: "getblocksrange",
				Params: params,
			}
			resultBs := HttpPost(conf, getBlockParams)
			// fmt.Println(string(resultBs))

			resultVO := new(RpcResult)
			json.Unmarshal(resultBs, resultVO)
			bs, _ := json.Marshal(resultVO.Result)

			// fmt.Println("=================\n", string(bs))

			array := make([]interface{}, 0)
			json.Unmarshal(bs, &array)
			for _, one := range array {
				bsOne, _ := json.Marshal(one)

				bhvo, _ := mining.ParseBlockHeadVO(&bsOne)
				// fmt.Println(bhvo, err)
				bhvojsonbs, _ := bhvo.Json()
				fmt.Println(string(*bhvojsonbs))

				//开始处理区块中的交易
				count(bhvo)

				localHeight = bhvo.BH.Height

			}

		}

		//当本地高度和远程高度一致，则间隔一秒钟获取最新高度
		time.Sleep(time.Second)
	}

	fmt.Println("hello")
}

type Info struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func HttpPost(conf Config, info Info) []byte {
	jsons, _ := json.Marshal(info)
	result := string(jsons)
	jsonInfo := strings.NewReader(result)
	req, _ := http.NewRequest("POST", "http://"+conf.RpcHost+"/rpc", jsonInfo)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("user", conf.RpcUser)
	req.Header.Add("password", conf.RpcPwd)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("error create client:%v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("error getInfo:%v", err)
	}
	return body
}

type RpcResult struct {
	JsonRpc string      `json:"jsonrpc"`
	Code    int         `json:"code"`
	Result  interface{} `json:"result"`
}

//统计交易，保存成txItem
func count(bhvo *mining.BlockHeadVO) {
	for _, txOne := range bhvo.Txs {
		switch txOne.Class() {
		case config.Wallet_tx_type_pay:
			vouts := txOne.GetVout()
			//构建TxItem
			for voutIndex, voutOne := range *vouts {
				ok := voutOne.CheckIsSelf()
				// addrInfo, ok := keystore.FindAddress(vout.Address)
				if !ok {
					continue
				}
				//把这个txItem保存起来，后面交易要用
				txItem := mining.TxItem{
					Addr:         &voutOne.Address,     //
					Value:        voutOne.Value,        //余额
					Txid:         bhvo.BH.Hash,         //交易id
					OutIndex:     uint64(voutIndex),    //交易输出index，从0开始
					Height:       bhvo.BH.Height,       //
					LockupHeight: voutOne.FrozenHeight, //锁仓高度
				}
				fmt.Println("保存这个txitem", txItem)
			}
		case config.Wallet_tx_type_account:
			tx := txOne.(*tx_name_in.Tx_account)
			fmt.Println("注册了域名", string(tx.Account))
		}
	}
}
