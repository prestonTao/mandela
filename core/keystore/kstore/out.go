package kstore

import (
	"fmt"
)

//返回数据格式
type DataInfo struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func (d *DataInfo) Json() string {
	rs, _ := json.Marshal(d)
	return string(rs)
}
func ParseDataInfo(bs string) DataInfo {
	di := DataInfo{}
	err := json.Unmarshal([]byte(bs), &di)
	// decoder := json.NewDecoder(bytes.NewBuffer([]byte(bs)))
	// decoder.UseNumber()
	// err := decoder.Decode(&di)
	if err != nil {
		fmt.Println(err)
	}
	return di
}
func Out(code int, data interface{}) string {
	d := DataInfo{Code: code, Data: data}
	return d.Json()
}
