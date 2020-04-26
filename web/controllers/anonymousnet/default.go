package anonymousnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/name"
	"mandela/core/nodeStore"
	"mandela/proxyhttp"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {

	this.TplName = "mandela/index.tpl"
}

/*
	代理
	xxxxxxxxxxxx:8080
*/
func (this *MainController) Agent() {
	url := this.Ctx.Input.Param(":splat")

	//	fmt.Println("uri", this.Ctx.Request.RequestURI)

	// fmt.Println("请求地址", url)

	temp := strings.SplitN(url, "/", 2)
	url = "/" + url
	//	if len(temp) > 1 {
	//		url = "/" + temp[1]
	//	} else {
	//		url = "/"
	//	}

	var id *nodeStore.TempId
	port := uint16(0)

	//判断能否解析为id，能解析为id则当作id访问，不能解析则当作域名访问
	temp = strings.SplitN(temp[0], ":", 2)

	height := mining.GetHighestBlock()

	// targetMhId := name.FindNameRandOne(temp[0])
	// if targetMhId == nil {
	// 	fmt.Println("在属于自己的域名中没有找到")

	// }
	targetMhId := name.FindNameToNetRandOne(temp[0], height)
	if targetMhId == nil {
		fmt.Println("在所有域名中没有找到")
		fmt.Println("开始解析节点id")
		id := nodeStore.AddressFromB58String(temp[0])
		targetMhId = &id
	}

	fmt.Println("找到的节点id是", targetMhId.B58String())

	id = nodeStore.NewTempId(targetMhId, targetMhId)

	// // targetMhId, err := utils.FromB58String(temp[0])
	// if err == nil {
	// 	//能解析为id，通过地址访问
	// } else {
	// 	//通过域名访问
	// 	id = cache_store.GetAddressInName(temp[0])
	// }
	if len(temp) > 1 {
		var err error
		p, err := strconv.Atoi(temp[1])
		if err != nil {
			// fmt.Println("格式错误：", err)
			//TODO 格式错误，返回错误信息
			return
		}
		port = uint16(p)
	}

	//	temp = strings.SplitN(temp[0], "id:", 2)
	//	if len(temp) > 1 {
	//		temp = strings.SplitN(temp[1], ":", 2)
	//		//		var err error
	//		//		targetId, err := hex.DecodeString(temp[0])
	//		//		if err != nil {
	//		//			fmt.Println(err)
	//		//			//TODO 格式错误，返回错误信息
	//		//			return
	//		//		}
	//		targetMhId, err := utils.FromB58String(temp[0])
	//		if err != nil {
	//			fmt.Println(err)
	//			//TODO 格式错误，返回错误信息
	//			return
	//		}
	//		id = nodeStore.NewTempId(&targetMhId, &targetMhId)

	//		if len(temp) > 1 {
	//			port, err = strconv.Atoi(temp[1])
	//			if err != nil {
	//				fmt.Println(err)
	//				//TODO 格式错误，返回错误信息
	//				return
	//			}
	//		}

	//	} else {
	//		//是通过域名方式访问的
	//		temp = strings.SplitN(temp[0], ":", 2)
	//		//		name := temp[0]
	//		id = cache_store.GetAddressInName(temp[0])
	//		if len(temp) > 1 {
	//			var err error
	//			port, err = strconv.Atoi(temp[1])
	//			if err != nil {
	//				fmt.Println(err)
	//				//TODO 格式错误，返回错误信息
	//				return
	//			}
	//		}
	//	}

	//	this.Ctx.Request.Header
	bodyData := make([]byte, 0)
	this.Ctx.Request.Body.Read(bodyData)

	//	data, err := json.Marshal(this.Ctx.Request.Header)
	//	fmt.Println("-----", string(data), "\n", bodyData, err)

	request := proxyhttp.HttpRequest{
		Port:    port,
		Method:  this.Ctx.Request.Method,
		Header:  this.Ctx.Request.Header,
		Body:    bodyData,
		Url:     url,
		Cookies: this.Ctx.Request.Cookies(),
	}

	bs := proxyhttp.SendHttpRequest(id, request)

	if bs == nil {
		fmt.Println("返回为空")
		return
	}

	response := new(proxyhttp.HttpResponse)
	// err := json.Unmarshal(*bs, &response)
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
		return
	}
	//	fmt.Println("recv header", response.Header)
	//	this.Ctx.Request.Response.Header = response.Header

	this.Ctx.ResponseWriter.WriteHeader(response.StatusCode)

	for key, value := range response.Header {
		this.Ctx.ResponseWriter.Header().Set(key, value[0])
		for i := 1; i < len(value); i++ {
			this.Ctx.ResponseWriter.Header().Add(key, value[i])
		}
	}

	//	this.Ctx.Output.Context.
	for _, one := range response.Cookies {
		http.SetCookie(this.Ctx.ResponseWriter, one)
	}

	//	this.Ctx.
	//	fmt.Println("body", response.Body, len(response.Body))
	//	if len(response.Body) > 0 {
	//	}
	this.Ctx.ResponseWriter.Write(response.Body)

	//	resp, err := http.Get("http://" + url)
	//	if err != nil {
	//		fmt.Println("请求服务器错误")
	//		return
	//	}
	//	if resp.StatusCode == 200 {
	//		io.Copy(this.Ctx.ResponseWriter, resp.Body)
	//	}

	return

	//	//	url := "/findcoin"
	//	method := this.Ctx.Request.Method
	//	//	params := map[string]interface{}{
	//	//		"Page": 0,
	//	//		"Size": 10,
	//	//	}

	//	//	header := http.Header{"RANGE": []string{"0-200"},
	//	//		"User-Agent": []string{"OperatingPlatform"},
	//	//		"Accept":     []string{"text/html, image/gif, image/jpeg, *; q=.2, */*; q=.2"}}
	//	client := &http.Client{}
	//	//req, err := http.NewRequest("GET", "http://www.baidu.com/", nil)
	//	//	bs, err := json.Marshal(params)
	//	req, err := http.NewRequest(method, url, nil)
	//	if err != nil {
	//		fmt.Println("创建request错误")
	//		return
	//	}
	//	//	req.Header = header
	//	resp, err = client.Do(req)
	//	if err != nil {
	//		fmt.Println("请求服务器错误")
	//		return
	//	}
	//	fmt.Println("response:", resp.StatusCode)
	//	if resp.StatusCode == 200 {
	//		robots, err := ioutil.ReadAll(resp.Body)
	//		if err != nil {
	//			fmt.Println("读取body内容错误")
	//			return
	//		}
	//		fmt.Println(string(robots))

	//		//		io.Copy(this.Ctx.ResponseWriter, resp.Body)
	//		io.WriteString(this.Ctx.ResponseWriter, string(robots))
	//	}
}

/*
	代理
*/
func (this *MainController) AgentToo() {
	// fmt.Println("没有命中")
}
