package main

import (
	"mandela/core/utils/crypto"
	"mandela/rpc/model"
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func main() {
	payMore()
}

func payMore() {
	bs, err := ioutil.ReadFile("config.json")
	if err != nil {
		// fmt.Println("2222222222222222222222222222")
		panic("读取配置文件错误：" + err.Error())
		return
	}
	cfi := new(Config)
	// err = json.Unmarshal(bs, cfi)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(cfi)
	if err != nil {
		// fmt.Println("44444444444444444444444444")
		panic("解析配置文件错误：" + err.Error())
		return
	}

	// conf := Config{
	// 	RpcHost: "192.168.28.35:5081", //节点地址及端口号
	// 	RpcUser: "test",               //远程调用用户名
	// 	RpcPwd:  "testp",              //远程调用密码
	// }

	payNumbers := readTextFile("addrs.txt")

	input := bufio.NewScanner(os.Stdin)
	fmt.Print("请输入支付密码:")
	input.Scan()
	pwd := input.Text()
	// fmt.Println(line)
	// return

	pageNumber := 10

	for i := 0; i <= len(payNumbers)/pageNumber; i++ {
		start := i * pageNumber
		end := (i + 1) * pageNumber
		params := map[string]interface{}{
			"addresses": payNumbers[start:end],
			"gas":       0,
			"pwd":       pwd,
		}
		payMores := Info{
			Method: "sendtoaddressmore",
			Params: params,
		}
		resultBs := HttpPost(*cfi, payMores)
		resultVO := new(RpcResult)
		json.Unmarshal(resultBs, resultVO)
		// bs, _ := json.Marshal(resultVO.Result)
		if resultVO.Code != model.Success {
			log.Println("start:" + strconv.Itoa(start) + " end:" + strconv.Itoa(end))
			break
		} else {
		}
		log.Println(string(resultBs))
		time.Sleep(time.Second * 10)
	}
	log.Println("剩下 " + strconv.Itoa(len(payNumbers)%pageNumber) + " 个未发放")
}

/*
	多人转账
*/
type PayNumber struct {
	Address string `json:"address"` //转账地址
	Amount  uint64 `json:"amount"`  //转账金额
	// FrozenHeight uint64 `json:"frozen_height"` //冻结高度
}

type Info struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func HttpPost(conf Config, info Info) []byte {
	jsons, _ := json.Marshal(info)
	result := string(jsons)
	jsonInfo := strings.NewReader(result)
	req, _ := http.NewRequest("POST", "http://"+conf.WebAddr+":"+strconv.Itoa(int(conf.WebPort))+"/rpc", jsonInfo)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("user", conf.RpcUser)
	req.Header.Add("password", conf.RpcPassword)
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

// type Config struct {
// 	RpcHost string //节点地址及端口号
// 	RpcUser string //远程调用用户名
// 	RpcPwd  string //远程调用密码
// }

//一行一行的读文本文件，并打印出来
func readTextFile(path string) []PayNumber {
	index := 1
	pns := make([]PayNumber, 0)
	file, _ := os.Open(path)
	buf := bufio.NewReader(file)
	for {
		// if index > 500 {
		// 	panic("超过了500个地址")
		// }
		line, _, err := buf.ReadLine()
		if err != nil {
			break
		}
		fmt.Println(string(line))

		strs := strings.Split(string(line), " ")
		if len(strs) != 2 {
			panic("第" + strconv.Itoa(index) + "行格式错误")
		}
		addr := crypto.AddressFromB58String(strs[0])
		if addr == nil {
			panic("第" + strconv.Itoa(index) + "行，地址格式错误")
		}
		ok := crypto.ValidAddr("TEST", addr)
		if !ok {
			panic("第" + strconv.Itoa(index) + "行，地址格式错误")
		}

		amout, err := strconv.Atoi(strs[1])
		if err != nil {
			panic("第" + strconv.Itoa(index) + "行,金额格式错误")
		}
		pnOne := PayNumber{
			Address: strs[0],       //转账地址
			Amount:  uint64(amout), //转账金额
		}
		pns = append(pns, pnOne)
		index++
	}
	file.Close()
	return pns
}

type Config struct {
	Netid       uint32 `json:"netid"`       //
	IP          string `json:"ip"`          //ip地址
	Port        uint16 `json:"port"`        //监听端口
	WebAddr     string `json:"WebAddr"`     //
	WebPort     uint16 `json:"WebPort"`     //
	WebStatic   string `json:"WebStatic"`   //
	WebViews    string `json:"WebViews"`    //
	RpcServer   bool   `json:"RpcServer"`   //
	RpcUser     string `json:"RpcUser"`     //
	RpcPassword string `json:"RpcPassword"` //
	Miner       bool   `json:"miner"`       //本节点是否是矿工
	NetType     string `json:"NetType"`     //正式网络release/测试网络test
	AddrPre     string `json:"AddrPre"`     //收款地址前缀
}
