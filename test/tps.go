package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	// for {
	tps()
	// }
}

func tps() {
	url := "/rpc"
	method := "POST"
	params := map[string]interface{}{
		"address": "13gyTcKozkP5b1bSTxFWn8sYttcoWT7Bmt",
		"amount":  100000000,
		"gas":     0,
		"pwd":     "123456789",
		"comment": "test",
	}

	rpcParans := map[string]interface{}{
		"method": "sendtoaddress",
		"params": params,
	}

	header := http.Header{
		"user":         []string{"test"},
		"password":     []string{"testp"},
		"Content-Type": []string{"application/json"}}
	client := &http.Client{}
	//req, err := http.NewRequest("GET", "http://www.baidu.com/", nil)
	bs, err := json.Marshal(rpcParans)
	req, err := http.NewRequest(method, "http://127.0.0.1:2080"+url, strings.NewReader(string(bs)))
	if err != nil {
		fmt.Println("创建request错误")
		return
	}
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求服务器错误")
		return
	}
	fmt.Println("response:", resp.StatusCode)
	if resp.StatusCode == 200 {
		robots, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("读取body内容错误")
			return
		}

		fmt.Println(len(robots), string(robots))
	}
}
