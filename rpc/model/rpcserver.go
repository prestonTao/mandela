package model

import (
	"reflect"
)

type RpcHandler interface {
	SetBody(data []byte)
	GetBody() []byte
	Out(data []byte)
	Err(code, data string)
	Validate() (msg string, ok bool)
}
type RpcJson struct {
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

func (rj *RpcJson) Get(key string) (interface{}, bool) {
	v, b := rj.Params[key]
	return v, b
}
func (rj *RpcJson) Type(key string) string {
	v, b := rj.Get(key)
	if !b {
		return ""
	}
	return reflect.TypeOf(v).String()
}
func (rj *RpcJson) VerifyType(key, types string) bool {
	if rj.Type(key) == types {
		return true
	} else {
		return false
	}

}
