package sharebox

import (
	"mandela/rpc/model"
	"mandela/sharebox"
	"net/http"
)

/*
	查询共享文件夹列表
*/
func ShareFolderList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	rootDir := sharebox.GetShareFolderRootsDetail()
	res, err = model.Tojson(rootDir)
	return
}

/*
	添加本地共享文件夹
*/
func AddLocalShareFoler(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	pathItr, ok := rj.Get("path") //文件夹绝对路径
	if !ok {
		res, err = model.Errcode(5002, "path")
		return
	}
	absPath := pathItr.(string)

	err = sharebox.AddLocalShareFolders(absPath)
	if err == nil {
		res, err = model.Tojson("success")
	}
	return
}

/*
	添加本地共享文件夹
*/
func DelLocalShareFoler(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	pathItr, ok := rj.Get("path") //文件夹绝对路径
	if !ok {
		res, err = model.Errcode(5002, "path")
		return
	}
	absPath := pathItr.(string)

	err = sharebox.DelLocalShareFolders(absPath)
	if err == nil {
		res, err = model.Tojson("success")
	}
	return
}

/*
	查询远端节点共享文件夹列表
*/
func GetRemoteShareFolderList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	idItr, ok := rj.Get("id") //远端节点id
	if !ok {
		res, err = model.Errcode(5002, "id")
		return
	}
	id := idItr.(string)

	var rootDir *sharebox.DirVO
	rootDir, err = sharebox.GetRemoteShareFolderDetail(id)
	if err != nil {
		// fmt.Println("1有没有超时", err)
		res, err = model.Errcode(model.Nomarl, "fail")
		return
	}
	// fmt.Println("2有没有超时", err)
	// bs, _ := json.Marshal(rootDir)
	// fmt.Println(string(bs))
	res, err = model.Tojson(rootDir)
	return
}

// /*
// 	购买资源下载权限付费
// */
// func Resources(rj *model.RpcJson) (res []byte, err error) {

// 	// var addr *utils.Multihash
// 	// addrItr, ok := rj.Get("address") //转账地址
// 	// if ok {
// 	// 	addrStr := addrItr.(string)
// 	// 	fmt.Println("发送到的地址：", addrStr)
// 	// 	addrMul, e := utils.FromB58String(addrStr)
// 	// 	if e != nil {
// 	// 		res, err = model.Errcode(5002, "address")
// 	// 		return
// 	// 	}
// 	// 	addr = &addrMul
// 	// }

// 	// amountItr, ok := rj.Get("amount") //转账金额
// 	// if !ok {
// 	// 	res, err = model.Errcode(5002, "amount")
// 	// 	return
// 	// }
// 	// amount := uint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas") //手续费
// 	if !ok {
// 		res, err = model.Errcode(5002, "gas")
// 		return
// 	}
// 	gas := uint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd") //支付密码
// 	if !ok {
// 		res, err = model.Errcode(5002, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	nameItr, ok := rj.Get("name") //注册的名称
// 	if !ok {
// 		res, err = model.Errcode(5002, "name")
// 		return
// 	}
// 	name := nameItr.(string)

// 	netIdsItr, ok := rj.Get("resources") //名称解析的网络id
// 	if !ok {
// 		res, err = model.Errcode(5002, "resources")
// 		return
// 	}
// 	netIds := netIdsItr.([]interface{})
// 	ress := make([]utils.Multihash, 0)
// 	for _, one := range netIds {
// 		netidOne := one.(string)
// 		idOne, e := utils.FromB58String(netidOne)
// 		if e != nil {
// 			res, err = model.Errcode(5002, "resources")
// 			return
// 		}
// 		ress = append(ress, idOne)
// 	}

// 	err = tx_resources_download.Resources(gas, pwd, name, ress)
// 	if err == nil {
// 		res, err = model.Tojson("success")
// 	}
// 	return
// }

// /*
// 	上传资源付费
// */
// func ResourcesUpload(rj *model.RpcJson) (res []byte, err error) {

// 	// var addr *utils.Multihash
// 	// addrItr, ok := rj.Get("address") //转账地址
// 	// if ok {
// 	// 	addrStr := addrItr.(string)
// 	// 	fmt.Println("发送到的地址：", addrStr)
// 	// 	addrMul, e := utils.FromB58String(addrStr)
// 	// 	if e != nil {
// 	// 		res, err = model.Errcode(5002, "address")
// 	// 		return
// 	// 	}
// 	// 	addr = &addrMul
// 	// }

// 	amountItr, ok := rj.Get("amount") //付费金额
// 	if !ok {
// 		res, err = model.Errcode(5002, "amount")
// 		return
// 	}
// 	amount := uint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas") //手续费
// 	if !ok {
// 		res, err = model.Errcode(5002, "gas")
// 		return
// 	}
// 	gas := uint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd") //支付密码
// 	if !ok {
// 		res, err = model.Errcode(5002, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	nameItr, ok := rj.Get("name") //注册的名称
// 	if !ok {
// 		res, err = model.Errcode(5002, "name")
// 		return
// 	}
// 	name := nameItr.(string)

// 	netIdsItr, ok := rj.Get("resources") //名称解析的网络id
// 	if !ok {
// 		res, err = model.Errcode(5002, "resources")
// 		return
// 	}
// 	netIds := netIdsItr.([]interface{})
// 	ress := make([]utils.Multihash, 0)
// 	for _, one := range netIds {
// 		netidOne := one.(string)
// 		idOne, e := utils.FromB58String(netidOne)
// 		if e != nil {
// 			res, err = model.Errcode(5002, "resources")
// 			return
// 		}
// 		ress = append(ress, idOne)
// 	}

// 	err = tx_resources_upload.Resources(amount, gas, pwd, name, ress)
// 	if err == nil {
// 		res, err = model.Tojson("success")
// 	}
// 	return
// }
