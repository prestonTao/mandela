package store

import (
	"mandela/core/nodeStore"
	"mandela/core/virtual_node"
	"mandela/rpc/model"
	"mandela/sharebox"
	ystore "mandela/store"
	"mandela/store/fs"
	"fmt"
	"net/http"
	"sort"
)

/*
	上传文件
*/
func UploadFile() {
	fmt.Println("123456789")
}

type fileinfoObj struct {
	*ystore.FileIndex
	HasCode string
}
type fileinfoObjList []*fileinfoObj

func (fil fileinfoObjList) Len() int {
	return len(fil)
}
func (fil fileinfoObjList) Less(i, j int) bool {
	return fil[i].FileIndex.Time > fil[j].FileIndex.Time
}
func (fil fileinfoObjList) Swap(i, j int) {
	fil[i], fil[j] = fil[j], fil[i]
}

// func (this fileinfoObjList) JSON()*[]byte{
// 	for
// }

/*
	获取文件列表
*/
func GetFileList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	_, fileInfos := ystore.GetFileindexToSelfAll()
	// kind := this.GetString("kind")

	var fileinfoList fileinfoObjList
	// findNextFile:
	for _, fileinfo := range fileInfos {
		var fiobj fileinfoObj
		fiobj.HasCode = fileinfo.Hash.B58String()
		fiobj.FileIndex = fileinfo

		fileinfoList = append(fileinfoList, &fiobj)

	}
	sort.Sort(fileinfoList)
	res, err = model.Tojson(fileinfoList)
	return
}

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
		return
	}
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
//搜索文件索引
func SearchFileInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hash, ok := rj.Get("hash") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "hash")
		return
	}
	fn := hash.(string)

	fid := virtual_node.AddressFromB58String(fn)

	//查询文件信息
	fileindex := ystore.FindFileindex(fid)
	if fileindex == nil {
		fileindex, err = ystore.FindFileindexOpt(fn)
		if fileindex == nil || err != nil {
			fmt.Println("网络中查找文件信息失败")
			res, err = model.Errcode(5003, "网络中查找文件信息失败")
			return
		}
	}
	if fileindex.CryptUser != nil && fileindex.CryptUser.B58String() != nodeStore.NodeSelf.IdInfo.Id.B58String() {
		fmt.Println("加密文件只有所有者有权下载")
		res, err = model.Errcode(5003, "加密文件只有所有者有权下载")
		return
	}
	res, err = model.Tojson(fileindex)
	return
}

//删除文件索引
func DelFileInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hash, ok := rj.Get("hash") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "hash")
		return
	}
	fn := hash.(string)
	fid := virtual_node.AddressFromB58String(fn)
	//查询文件信息
	err = ystore.DelFileInfoFromSelf(fid)
	if err != nil {
		fmt.Println("删除文件失败", err.Error())
		res, err = model.Errcode(model.Nomarl, err.Error())
	} else {
		res, err = model.Tojson("删除成功")
	}
	return
}

//增加本地索引
func AddFileInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hash, ok := rj.Get("hash") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "hash")
		return
	}
	fn := hash.(string)
	fid := virtual_node.AddressFromB58String(fn)
	//查询文件信息
	finfo := ystore.FindFileindex(fid)
	if finfo == nil {
		finfo, err = ystore.FindFileindexOpt(fn)
		if finfo == nil || err != nil {
			fmt.Println("网络中查找文件信息失败")
			res, err = model.Errcode(5002, "file not fund")
			return
		}
	}
	fmt.Printf("文件索引:%+v", finfo)
	if finfo.CryptUser != nil && finfo.CryptUser.B58String() != nodeStore.NodeSelf.IdInfo.Id.B58String() {
		fmt.Println("加密文件只有所有者有权下载")
		res, err = model.Errcode(5002, "加密文件只有所有者有权下载")
		return
	}
	//增加文件所有者
	vnodeinfo := fs.GetNotUseSpace(finfo.Size)

	finfo.AddFileOwner(*vnodeinfo)

	//	finfo.AddFileUser(nodeStore.NodeSelf.IdInfo.Id.B58String())
	err = ystore.AddFileindexToSelf(finfo, vnodeinfo.Vid)
	res, err = model.Tojson("添加成功")
	return
}

//获取单个文件下载进度
func DownloadProcOne(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hash, ok := rj.Get("hash") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "hash")
		return
	}
	fn := hash.(string)

	vid := virtual_node.AddressFromB58String(fn)

	//查询文件信息
	finfo := ystore.FindFileindex(vid)
	dp := ystore.DownloadProgressOne(finfo)
	res, err = model.Tojson(dp)
	return
}

//获取文件下载进度
func DownloadProc(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	dp := ystore.DownloadProgress()
	res, err = model.Tojson(dp)
	return
}

//获取已下载列表
func DownloadComplete(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	dp := ystore.DownLoadComplete()
	res, err = model.Tojson(dp)
	return
}

//暂停下载
func DownLoadStop(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hash, ok := rj.Get("hash") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "hash")
		return
	}
	fn := hash.(string)
	ystore.DownLoadStop(fn)
	res, err = model.Tojson("暂停成功")
	return
}

//删除下载列表
func DownLoadDel(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	hash, ok := rj.Get("hash") //文件hash
	if !ok {
		res, err = model.Errcode(5002, "hash")
		return
	}
	fn := hash.(string)
	ystore.DwonLoadDel(fn)
	res, err = model.Tojson("删除成功")
	return
}

/*
	添加扩展云存储空间
*/
func SetSpaceSize(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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
func GetSpaceSize(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	total := fs.GetSpaceSize()
	useSize := fs.GetUseSpaceSize()
	result := make(map[string]uint64, 0)
	result["TotalSize"] = total
	result["UseSize"] = useSize

	//	vnodeinfo := virtual_node.GetVnodeNumber()
	//	for _, one := range vnodeinfo {
	//		voOne := VnodeinfoVO{
	//			Nid:   one.Nid.B58String(), //节点真实网络地址
	//			Index: one.Index,           //节点第几个空间，从1开始,下标为0的节点为实际节点。
	//			Vid:   one.Vid.B58String(), //vid，虚拟节点网络地址
	//		}
	//		result = append(result, voOne)
	//	}
	res, err = model.Tojson(result)
	return
}

/*
	添加扩展云存储空间
*/
func AddSpaceSize(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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

	fs.AddSpace(absPath, n)
	//	virtual_node.SetupVnodeNumber(n)
	res, err = model.Tojson("success")
	return
}

/*
	删除扩展云存储空间
*/
func DelSpaceSize(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
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
