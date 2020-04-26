package rpc

import (
	// "fmt"
	"mandela/rpc/model"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/astaxie/beego"
)

var (
	Port   = 8080
	Server = false //是否开启RPC true 开启 false 关闭
	//	Allowip  = "127.0.0.1"  //添加rpc ip地址白名单，为空则所有连接可用。
	Allowip  = "" //添加rpc ip地址白名单，为空则所有连接可用。
	User     string
	Password string
)

type Handler struct {
	w    http.ResponseWriter
	r    *http.Request
	body []byte
}

func (h *Handler) init(w http.ResponseWriter, r *http.Request) *Handler {
	h.w = w
	h.r = r
	return h
}
func (h *Handler) Out(data []byte) {
	h.w.Header().Add("Content-Type", "application/json")
	datas := append(append([]byte(`{"jsonrpc":"2.0","code":`), append([]byte(strconv.Itoa(model.Success)), byte(','))...), data[1:]...)
	h.w.Write(datas)
	return
}
func (h *Handler) Err(code, data string) {
	//codes, _ := strconv.Atoi(code)
	//h.w.WriteHeader(codes)
	h.w.Header().Add("Content-Type", "application/json")
	h.w.Write([]byte(`{"jsonrpc":"2.0","code":` + code + `,"message":"` + data + `"}`))
	return
}
func (h *Handler) Validate() (msg string, ok bool) {

	if Allowip != "" && h.RemoteIp() != Allowip {
		msg = "deny ip"
		ok = true
	}

	if h.r.Header.Get("user") != User || h.r.Header.Get("password") != Password {
		msg = "user or password is wrong"
		ok = true
	}
	return
}
func (h *Handler) doHandler() {
	vali, ok := h.Validate()
	if ok {
		h.Err("301", vali)
		return
	}

	body, err := ioutil.ReadAll(h.r.Body)
	if err != nil {
		// fmt.Println(err)
		h.Err("401", "body empty")
		return
	}
	h.SetBody(body)
	//fmt.Printf("%+v\n %s\n", body, body)

	//上传文件
	if h.r.Header.Get("file") == "upload" {
		// fh, ok := h.r.MultipartForm.File["file"]
		// if ok {
		// 	// fh.
		// }
		// res, err := UploadFile(h)
		// if err != nil {
		// 	h.Err(string(res), err.Error())
		// 	return
		// }
		// h.Out(res)
	} else {
		//普通RPC调用
		res, err := Route(h, h.w, h.r)
		if err != nil {
			h.Err(string(res), err.Error())
			return
		}
		h.Out(res)
	}

}
func (h *Handler) SetBody(data []byte) {
	h.body = data
}
func (h *Handler) GetBody() []byte {
	return h.body
}
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.init(w, r).doHandler()
}
func (h *Handler) RemoteIp() string {
	remoteAddr := h.r.RemoteAddr
	if ip := h.r.Header.Get("XRealIP"); ip != "" {
		remoteAddr = ip
	} else if ip = h.r.Header.Get("XForwardedFor"); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}

//func RegisterRpcServer() {
//	if Server == 1 {
//		fmt.Println("rpcserver listen on :", Port)
//		fmt.Println(http.ListenAndServe(":"+strconv.Itoa(Port), &Handler{}))
//	}
//}

type Bind struct {
	beego.Controller
}

func (i *Bind) Index() {
	if Server {
		handler := &Handler{}
		handler.ServeHTTP(i.Ctx.ResponseWriter, i.Ctx.Request)
	}
	return
}
