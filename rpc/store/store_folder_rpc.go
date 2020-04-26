package store

import (
	"mandela/rpc/model"
	"mandela/store"
	"net/http"
)

/*
	增加本地文件夹
*/
func AddFolder(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	name, ok := rj.Get("name") //文件夹名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	fname := name.(string)
	var pid uint64
	parentid, ok := rj.Get("parentid")
	if ok {
		pid = uint64(parentid.(float64))
	}
	err = store.AddFolder(pid, fname)
	if err == nil {
		res, err = model.Tojson("success")
	}
	return
}

/*
	删除本地文件夹
*/
func DelFolder(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	idstr, ok := rj.Get("id") //文件夹名称
	if !ok {
		res, err = model.Errcode(5002, "id")
		return
	}
	id := idstr.(float64)
	err = store.DelFolder(uint64(id))
	if err == nil {
		res, err = model.Tojson("success")
	}
	return
}

/*
	修改本地文件夹
*/
func UpFolder(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	idstr, ok := rj.Get("id") //文件夹名称
	if !ok {
		res, err = model.Errcode(5002, "id")
		return
	}
	id := uint64(idstr.(float64))
	namestr, ok := rj.Get("name") //文件夹名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := namestr.(string)
	var pid uint64
	parentid, ok := rj.Get("parentid")
	if ok {
		pid = uint64(parentid.(float64))
	}
	err = store.UpFolder(id, pid, name)
	if err != nil {
		res, err = model.Tojson(err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

/**
	文件夹列表
**/
func ListFolder(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	pidstr, ok := rj.Get("parentid") //文件夹名称
	if !ok {
		res, err = model.Errcode(5002, "parentid")
		return
	}
	pid := uint64(pidstr.(float64))
	list := store.ListFolder(pid)
	res, err = model.Tojson(list)
	return
}

/**
修改文件所属目录
**/
func Moveto(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hashstr, ok := rj.Get("hash") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "hash")
		return
	}
	hash := hashstr.(string)
	pidstr, ok := rj.Get("pid") //文件夹名称
	if !ok {
		res, err = model.Errcode(5002, "parentid")
		return
	}
	pid := uint64(pidstr.(float64))
	b := store.Moveto(hash, pid)
	res, err = model.Tojson(b)
	return
}
