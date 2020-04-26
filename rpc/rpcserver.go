package rpc

import (
	"mandela/rpc/model"
	"encoding/json"
	"fmt"
	"net/http"
)

func parseJson(jsonb []byte) (*model.RpcJson, error) {
	var rpcjson model.RpcJson
	err := json.Unmarshal(jsonb, &rpcjson)
	// decoder := json.NewDecoder(bytes.NewBuffer(jsonb))
	// decoder.UseNumber()
	// err := decoder.Decode(&rpcjson)
	//fmt.Printf("%+v\n", rpcjson)
	return &rpcjson, err
}
func Route(rh model.RpcHandler, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	data := rh.GetBody()
	rj, err := parseJson(data)
	if err != nil {
		// fmt.Println(err)
	}
	hd, ok := rpcHandler[rj.Method]
	if ok {
		res, err = hd(rj, w, r)
	} else {
		res, err = model.Errcode(model.NoMethod, rj.Method)
	}
	return
}

/*
	上传文件
*/
func UploadFile(rh model.RpcHandler) (res []byte, err error) {
	fmt.Println("开始上传文件")
	// rh.
	return nil, nil
}
