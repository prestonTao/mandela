package proxyhttp

import (
	"mandela/config"
	gconfig "mandela/config"
	"mandela/core/engine"
	mc "mandela/core/message_center"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func Register() {
	mc.Register_p2p(gconfig.MSGID_http_request, AgentHttpRequest)
	mc.Register_p2p(gconfig.MSGID_http_response, AgentHttpRespones)

	// mc.Register_p2p(gconfig.MSGID_http_getwebinfo, GetWebinfo)
	// mc.Register_p2p(gconfig.MSGID_http_getwebinfo_recv, GetWebinfo_recv)

}

type HttpRequest struct {
	Port    uint16         `json:"port"`
	Method  string         `json:"method"`
	Header  http.Header    `json:"header"`
	Body    []byte         `json:"body"`
	Url     string         `json:"url"`
	Cookies []*http.Cookie `json:"cookies"`
}

func (this *HttpRequest) JSON() *[]byte {
	bs, _ := json.Marshal(this)
	return &bs
}

type HttpResponse struct {
	StatusCode int            `json:"statuscode"`
	Header     http.Header    `json:"header"`
	Body       []byte         `json:"body"`
	Cookies    []*http.Cookie `json:"cookies"`
}

func (this *HttpResponse) JSON() *[]byte {
	bs, _ := json.Marshal(this)
	return &bs
}

/*
	发送http请求消息
*/
func SendHttpRequest(id *nodeStore.TempId, request HttpRequest) *[]byte {

	fmt.Println("请求信息", id, request.Method, request.Url, request.Port)
	message, ok, _ := mc.SendP2pMsg(gconfig.MSGID_http_request, id.PeerId, request.JSON())
	if ok {
		// return flood.WaitRequest(mc.MSG_WAIT_http_request, hex.EncodeToString(message.Body.Hash), 0)
		bs, _ := flood.WaitRequest(mc.MSG_WAIT_http_request, utils.Bytes2string(message.Body.Hash), 0)
		return bs
	}
	return nil

}

/*
	http请求
*/
func AgentHttpRequest(c engine.Controller, msg engine.Packet, message *mc.Message) {
	//	fmt.Println("保存一个公钥对应的域名")

	fmt.Println("接收到http请求")
	fmt.Println("请求内容", string(*message.Body.Content))
	//发送给自己的，自己处理
	request := new(HttpRequest)
	// err := json.Unmarshal(*message.Body.Content, &request)
	decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	decoder.UseNumber()
	err := decoder.Decode(&request)
	if err != nil {
		fmt.Println("json 错误", err)
		return
	}
	if request.Port == 0 {
		request.Port = config.WebPort
	}

	//开始请求
	client := &http.Client{}
	//req, err := http.NewRequest("GET", "http://localhost:" + strconv.Itoa(request.Port) + request.Url, nil)
	//	bs, err := json.Marshal(params)

	urls := strings.SplitN(request.Url, "/", 3)
	//	fmt.Println("--", len(urls), urls)

	url := ""
	if len(urls) <= 2 {
		url = "http://localhost:" + strconv.Itoa(int(request.Port))
	} else {
		url = "http://localhost:" + strconv.Itoa(int(request.Port)) + "/" + urls[2]
	}

	fmt.Println("请求地址", url)
	body := bytes.NewBuffer(request.Body)
	req, err := http.NewRequest(request.Method, url, body)
	if err != nil {
		fmt.Println("创建request错误")
		return
	}
	req.Header = request.Header
	for _, one := range request.Cookies {
		req.AddCookie(one)
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求服务器错误", err)
		return
	}

	//	//	url := "http://localhost:" + strconv.Itoa(request.Port) + request.Url

	//	resp, err := http.Get("http://localhost:" + strconv.Itoa(request.Port) + request.Url)
	//	if err != nil {
	//		fmt.Println("请求服务器错误")
	//		return
	//	}
	out := bytes.NewBuffer([]byte{})
	//	fmt.Println(resp.StatusCode)
	if resp.StatusCode == 200 || resp.StatusCode == 206 {
		io.Copy(out, resp.Body)
	}
	//	fmt.Println("")
	//	fmt.Println("header", resp.Header)
	response := HttpResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Body:       out.Bytes(),
		Cookies:    resp.Cookies(),
	}

	//给发送者回复
	mc.SendP2pReplyMsg(message, gconfig.MSGID_http_response, response.JSON())

}

/*
	http返回消息
*/
func AgentHttpRespones(c engine.Controller, msg engine.Packet, message *mc.Message) {
	//	fmt.Println("保存一个公钥对应的域名")

	// flood.ResponseWait(mc.MSG_WAIT_http_request, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(mc.MSG_WAIT_http_request, utils.Bytes2string(message.Body.Hash), message.Body.Content)

}

// /*
// 	获取节点web信息
// */
// func GetWebinfo(c engine.Controller, msg engine.Packet, message *mc.Message) {
// 	//	fmt.Println("保存一个公钥对应的域名")
// 	nwVO := new(NodeWebinfoVO)
// 	nwVO.WebPort = config.WebPort

// 	bs, _ := json.Marshal(nwVO)

// 	//给发送者回复
// 	mc.SendP2pReplyMsg(message, gconfig.MSGID_http_getwebinfo, &bs)

// }

// type NodeWebinfoVO struct {
// 	WebPort uint16
// }

// /*
// 	获取节点web信息
// */
// func GetWebinfo_recv(c engine.Controller, msg engine.Packet, message *mc.Message) {
// 	//	fmt.Println("保存一个公钥对应的域名")
// 	if message.Body.Content == nil {
// 		message.Body.Content = []byte("ok")
// 	}

// 	flood.ResponseWait(mc.CLASS_http_getwebinfo, message.Body.Hash.B58String(), message.Body.Content)

// }
