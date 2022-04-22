package rpc

import (
	"mandela/chain_witness_vote/mining/tx_name_in"
	"mandela/chain_witness_vote/mining/tx_name_out"
	"mandela/cloud_reward/client"
	"mandela/cloud_space/fs"
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/core/virtual_node"
	"mandela/rpc/model"
	"encoding/hex"
	"net/http"
	"strings"
)

/*
	存储挖矿空间质押
*/
func SpacesMiningIn(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
	if ok {
		addrStr := addrItr.(string)
		if addrStr != "" {
			addrMul := crypto.AddressFromB58String(addrStr)
			addr = &addrMul
		}

		if addrStr != "" {
			dst := crypto.AddressFromB58String(addrStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(model.ContentIncorrectFormat, "address")
				return
			}
		}
	}

	amountItr, ok := rj.Get("amount") //转账金额
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := uint64(amountItr.(float64))
	if amount < config.Mining_name_deposit_min {
		res, err = model.Errcode(model.Nomarl, config.ERROR_name_deposit.Error())
		return
	}

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	//域名解析的节点地址参数
	ids := make([]nodeStore.AddressNet, 0)
	netIdsItr, ok := rj.Get("netids") //名称解析的网络id
	if ok {
		netIds := netIdsItr.([]interface{})
		for _, one := range netIds {
			netidOne := one.(string)
			idOne := nodeStore.AddressFromB58String(netidOne)
			ids = append(ids, idOne)
		}
	}

	//收款地址参数
	coins := make([]crypto.AddressCoin, 0)
	addrcoinsItr, ok := rj.Get("addrcoins") //名称解析的收款地址
	if ok {
		addrcoins := addrcoinsItr.([]interface{})
		for _, one := range addrcoins {
			addrcoinOne := one.(string)
			idOne := crypto.AddressFromB58String(addrcoinOne)
			coins = append(coins, idOne)
		}
	}

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	txpay, err := tx_name_in.NameIn(nil, addr, amount, gas, frozenHeight, pwd, comment, name, ids, coins)
	if err == nil {
		// res, err = model.Tojson("success")

		result, e := utils.ChangeMap(txpay)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(result)

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(model.NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())

	return
}

/*
	存储空间挖矿取消质押
*/
func SpacesMiningOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// isReg := false //转账类型，true=注册名称；false=注销名称；

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金退还地址
	if ok {
		addrStr := addrItr.(string)
		if addrStr != "" {
			addrMul := crypto.AddressFromB58String(addrStr)
			addr = &addrMul
		}

		if addrStr != "" {
			dst := crypto.AddressFromB58String(addrStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(model.ContentIncorrectFormat, "address")
				return
			}
		}
	}

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	txpay, err := tx_name_out.NameOut(nil, addr, 0, gas, frozenHeight, pwd, comment, name)
	if err == nil {
		// res, err = model.Tojson("success")

		result, e := utils.ChangeMap(txpay)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(result)

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(model.NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_not_exist.Error() {
		res, err = model.Errcode(model.NotExist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())
	return

}

/*
	添加扩展云存储空间
*/
func SetCloudSpaceSize(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	nItr, ok := rj.Get("n") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "n")
		return
	}
	n := uint64(nItr.(float64))
	virtual_node.SetupVnodeNumber(n)
	res, err = model.Tojson("success")
	return
}

type VnodeinfoVO struct {
	Nid   string `json:"nid"`   //节点真实网络地址
	Index uint64 `json:"index"` //节点第几个空间，从1开始,下标为0的节点为实际节点。
	Vid   string `json:"vid"`   //vid，虚拟节点网络地址
}

/*
	查看扩展云存储空间
*/
func GetCloudSpaceList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	list := fs.GetSpaceList()
	total := fs.GetSpaceSize()
	useSize := fs.GetUseSpaceSize()
	result := make(map[string]interface{}, 0)
	result["Code"] = 0
	result["SpaceList"] = list
	result["TotalSize"] = total
	result["UseSize"] = useSize
	res, err = model.Tojson(result)
	return
}

/*
	添加扩展云存储空间
*/
func AddCloudSpaceSize(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//增加几个单位的云存储空间，一个单位空间为10G
	nItr, ok := rj.Get("n") //文件hash
	if !ok {
		res, err = model.Errcode(model.NoField, "n")
		return
	}
	n := uint64(nItr.(float64))

	absPath := ""
	//空间存放路径
	absPathItr, ok := rj.Get("absPath") //文件hash
	if ok {
		absPath = absPathItr.(string)
	}

	utils.Go(func() {
		fs.AddSpace(absPath, n)
		client.SendOnlineHeartBeat()
	})
	//	virtual_node.SetupVnodeNumber(n)
	res, err = model.Tojson("success")
	return
}

/*
	删除扩展云存储空间
*/
func DelCloudSpaceSize(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	//增加几个单位的云存储空间，一个单位空间为10G
	nItr, ok := rj.Get("n") //文件hash
	if !ok {
		res, err = model.Errcode(model.NoField, "n")
		return
	}
	n := uint64(nItr.(float64))

	fs.DelSpace(n)
	//	virtual_node.SetupVnodeNumber(n)
	res, err = model.Tojson("success")
	return
}

/*
	删除一个云存储空间
*/
func DelCloudSpaceOne(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	dbpathItr, ok := rj.Get("dbpath") //分片文件存储路径
	if !ok {
		res, err = model.Errcode(model.NoField, "dbpath")
		return
	}
	// n := uint64(nItr.(float64))

	fs.DelSpaceForDbPath(dbpathItr.(string))
	//	virtual_node.SetupVnodeNumber(n)
	res, err = model.Tojson("success")
	return
}

/*
	上传文件
*/
// func UploadFile() {
// 	fmt.Println("123456789")
// }

// type fileinfoObj struct {
// 	*ystore.FileIndex
// 	HasCode string
// }
// type fileinfoObjList []*fileinfoObj

// func (fil fileinfoObjList) Len() int {
// 	return len(fil)
// }
// func (fil fileinfoObjList) Less(i, j int) bool {
// 	return fil[i].FileIndex.Time > fil[j].FileIndex.Time
// }
// func (fil fileinfoObjList) Swap(i, j int) {
// 	fil[i], fil[j] = fil[j], fil[i]
// }

// func (this fileinfoObjList) JSON()*[]byte{
// 	for
// }

/*
	获取文件列表
*/
// func GetFileList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	_, fileInfos := ystore.GetFileindexToSelfAll()
// 	// kind := this.GetString("kind")

// 	var fileinfoList fileinfoObjList
// 	// findNextFile:
// 	for _, fileinfo := range fileInfos {
// 		var fiobj fileinfoObj
// 		fiobj.HasCode = fileinfo.Hash.B58String()
// 		fiobj.FileIndex = fileinfo

// 		fileinfoList = append(fileinfoList, &fiobj)

// 	}
// 	sort.Sort(fileinfoList)
// 	res, err = model.Tojson(fileinfoList)
// 	return
// }

/*
	查询共享文件夹列表
*/
// func ShareFolderList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	rootDir := sharebox.GetShareFolderRootsDetail()
// 	res, err = model.Tojson(rootDir)
// 	return
// }

/*
	添加本地共享文件夹
*/
// func AddLocalShareFoler(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	pathItr, ok := rj.Get("path") //文件夹绝对路径
// 	if !ok {
// 		res, err = model.Errcode(5002, "path")
// 		return
// 	}
// 	absPath := pathItr.(string)

// 	err = sharebox.AddLocalShareFolders(absPath)
// 	if err == nil {
// 		res, err = model.Tojson("success")
// 	}
// 	return
// }

/*
	添加本地共享文件夹
*/
// func DelLocalShareFoler(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	pathItr, ok := rj.Get("path") //文件夹绝对路径
// 	if !ok {
// 		res, err = model.Errcode(5002, "path")
// 		return
// 	}
// 	absPath := pathItr.(string)

// 	err = sharebox.DelLocalShareFolders(absPath)
// 	if err == nil {
// 		res, err = model.Tojson("success")
// 	}
// 	return
// }

/*
	查询远端节点共享文件夹列表
*/
// func GetRemoteShareFolderList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	idItr, ok := rj.Get("id") //远端节点id
// 	if !ok {
// 		res, err = model.Errcode(5002, "id")
// 		return
// 	}
// 	id := idItr.(string)

// 	var rootDir *sharebox.DirVO
// 	rootDir, err = sharebox.GetRemoteShareFolderDetail(id)
// 	if err != nil {
// 		return
// 	}
// 	// bs, _ := json.Marshal(rootDir)
// 	// fmt.Println(string(bs))
// 	res, err = model.Tojson(rootDir)
// 	return
// }

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
//搜索文件索引
// func SearchFileInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	hash, ok := rj.Get("hash") //文件hash
// 	if !ok {
// 		res, err = model.Errcode(5002, "hash")
// 		return
// 	}
// 	fn := hash.(string)

// 	fid := virtual_node.AddressFromB58String(fn)

// 	//查询文件信息
// 	fileindex := ystore.FindFileindex(fid)
// 	if fileindex == nil {
// 		fileindex, err = ystore.FindFileindexOpt(fn)
// 		if fileindex == nil || err != nil {
// 			fmt.Println("网络中查找文件信息失败")
// 			res, err = model.Errcode(5003, "网络中查找文件信息失败")
// 			return
// 		}
// 	}
// 	if fileindex.CryptUser != nil && fileindex.CryptUser.B58String() != nodeStore.NodeSelf.IdInfo.Id.B58String() {
// 		fmt.Println("加密文件只有所有者有权下载")
// 		res, err = model.Errcode(5003, "加密文件只有所有者有权下载")
// 		return
// 	}
// 	res, err = model.Tojson(fileindex)
// 	return
// }

//删除文件索引
// func DelFileInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	hash, ok := rj.Get("hash") //文件hash
// 	if !ok {
// 		res, err = model.Errcode(5002, "hash")
// 		return
// 	}
// 	fn := hash.(string)
// 	fid := virtual_node.AddressFromB58String(fn)
// 	//查询文件信息
// 	err = ystore.DelFileInfoFromSelf(fid)
// 	if err != nil {
// 		fmt.Println("删除文件失败", err.Error())
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 	} else {
// 		res, err = model.Tojson("删除成功")
// 	}
// 	return
// }

//增加本地索引
// func AddFileInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	hash, ok := rj.Get("hash") //文件hash
// 	if !ok {
// 		res, err = model.Errcode(5002, "hash")
// 		return
// 	}
// 	fn := hash.(string)
// 	fid := virtual_node.AddressFromB58String(fn)
// 	//查询文件信息
// 	finfo := ystore.FindFileindex(fid)
// 	if finfo == nil {
// 		finfo, err = ystore.FindFileindexOpt(fn)
// 		if finfo == nil || err != nil {
// 			fmt.Println("网络中查找文件信息失败")
// 			res, err = model.Errcode(5002, "file not fund")
// 			return
// 		}
// 	}
// 	fmt.Printf("文件索引:%+v", finfo)
// 	if finfo.CryptUser != nil && finfo.CryptUser.B58String() != nodeStore.NodeSelf.IdInfo.Id.B58String() {
// 		fmt.Println("加密文件只有所有者有权下载")
// 		res, err = model.Errcode(5002, "加密文件只有所有者有权下载")
// 		return
// 	}
// 	//增加文件所有者
// 	vnodeinfo := fs.GetNotUseSpace(finfo.Size)

// 	finfo.AddFileOwner(*vnodeinfo)

// 	//	finfo.AddFileUser(nodeStore.NodeSelf.IdInfo.Id.B58String())
// 	err = ystore.AddFileindexToSelf(finfo, vnodeinfo.Vid)
// 	res, err = model.Tojson("添加成功")
// 	return
// }

//获取单个文件下载进度
// func DownloadProcOne(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	hash, ok := rj.Get("hash") //文件hash
// 	if !ok {
// 		res, err = model.Errcode(5002, "hash")
// 		return
// 	}
// 	fn := hash.(string)

// 	vid := virtual_node.AddressFromB58String(fn)

// 	//查询文件信息
// 	finfo := ystore.FindFileindex(vid)
// 	dp := ystore.DownloadProgressOne(finfo)
// 	res, err = model.Tojson(dp)
// 	return
// }

//获取文件下载进度
// func DownloadProc(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	dp := ystore.DownloadProgress()
// 	res, err = model.Tojson(dp)
// 	return
// }

//获取已下载列表
// func DownloadComplete(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	dp := ystore.DownLoadComplete()
// 	res, err = model.Tojson(dp)
// 	return
// }

//暂停下载
// func DownLoadStop(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	hash, ok := rj.Get("hash") //文件hash
// 	if !ok {
// 		res, err = model.Errcode(5002, "hash")
// 		return
// 	}
// 	fn := hash.(string)
// 	ystore.DownLoadStop(fn)
// 	res, err = model.Tojson("暂停成功")
// 	return
// }

//删除下载列表
// func DownLoadDel(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	hash, ok := rj.Get("hash") //文件hash
// 	if !ok {
// 		res, err = model.Errcode(5002, "hash")
// 		return
// 	}
// 	fn := hash.(string)
// 	ystore.DwonLoadDel(fn)
// 	res, err = model.Tojson("删除成功")
// 	return
// }
