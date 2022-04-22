package store

import (
	"mandela/chain_witness_vote/mining/name"
	"mandela/cloud_space"
	"mandela/cloud_space/fs"
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/virtual_node"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/astaxie/beego"
)

type Index struct {
	beego.Controller
}

func (this *Index) Index() {
	// this.Ctx.Redirect(http.StatusMovedPermanently, "/static")
	this.TplName = "store/index.tpl"

}

type fileinfoObj struct {
	*cloud_space.FileIndex
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

//GetList 根据类型获取本地节点数据 @Param	kind		query 	string	true		"类型"
func (this *Index) GetList() {
	_, fileInfos := cloud_space.GetFileindexToSelfAll()
	// kind := this.GetString("kind")

	var result map[string]interface{} = make(map[string]interface{})
	var fileinfoList fileinfoObjList
	// findNextFile:
	for _, fileinfo := range fileInfos {
		var fiobj fileinfoObj
		fiobj.HasCode = fileinfo.Hash.B58String()
		fiobj.FileIndex = fileinfo

		fileinfoList = append(fileinfoList, &fiobj)

	}
	sort.Sort(fileinfoList)
	result["data"] = fileinfoList
	result["status"] = 200
	this.Data["json"] = &result
	this.ServeJSON()
}

/*
	添加文件到云盘
*/
func (this *Index) AddFile() {
	fmt.Println("开始上传文件-----")
	out := make(map[string]interface{})

	filePath := this.Ctx.Request.FormValue("file")
	fmt.Println("filepath:", filePath)

	//判断是否注册域名
	//	nameinfo := name.FindNameToNet(nodeStore.NodeSelf.IdInfo.Id.B58String())
	//	//fmt.Printf("xxxxx%+v", nameinfo)

	//	if nameinfo == nil {
	//		fmt.Println("未注册域名")
	//		out["Code"] = 1
	//		out["Message"] = "未注册域名"
	//		this.Data["json"] = out
	//		this.ServeJSON()
	//		return
	//	}
	//	if nameinfo.Deposit < store.DepositMin {
	//		fmt.Println("域名冻结押金不足")
	//		out["Code"] = 1
	//		out["Message"] = "域名冻结押金不足"
	//		this.Data["json"] = out
	//		this.ServeJSON()
	//		return
	//	}
	// this.GetFile()
	// hs, err := this.GetFiles("file")
	// if err != nil {
	// 	fmt.Println("获取文件头失败", err)
	// 	out["Code"] = 1
	// 	out["Message"] = "获取文件头失败"
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// fmt.Println("获取到了文件", hs)

	// f, err := hs[0].Open()
	// defer f.Close()
	// if err != nil {
	// 	fmt.Println("打开文件失败", err)
	// 	out["Code"] = 1
	// 	out["Message"] = "打开文件失败"
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// filename := hs[0].Filename

	// fmt.Println("filename:", filename)

	//	fi = &FileInfo{
	//		Name: p.FileName(),
	//		Type: p.Header.Get("Content-Type"),
	//	}

	//	f, h, err := this.Ctx.Request.FormFile("file")
	//	fmt.Println(f, h, err)

	// f, err := os.Open(filePath)
	// // file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0600)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// //保存文件到本地
	// newfile, err := os.OpenFile(filepath.Join(config.Store_temp, filename), os.O_RDWR|os.O_CREATE, os.ModePerm)
	// defer newfile.Close()
	// //验证空间是否已用完
	// st, _ := newfile.Stat()
	// vnodeinfo := fs.GetNotUseSpace(uint64(st.Size()))
	// if vnodeinfo == nil {

	// 	//	}

	// 	//	if !store.CheckSpace(uint64(st.Size())) {
	// 	fmt.Println("空间不足", err)
	// 	out["Code"] = 1
	// 	out["Message"] = "空间不足"
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// if err != nil {
	// 	fmt.Println("保存文件到本地失败", err)
	// 	out["Code"] = 1
	// 	out["Message"] = "保存文件到本地失败"
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// //	newfile, err := os.Create(filepath.Join(config.Store_temp, filename))
	// // buf, err := ioutil.ReadAll(f)
	// // newfile.Write(buf)
	// if _, err := io.Copy(newfile, f); err != nil {
	// 	fmt.Println("文件上传失败", err)
	// 	//文件切片失败
	// 	out["Code"] = 1
	// 	out["Message"] = "文件上传失败"
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }

	_, filename := filepath.Split(filePath)

	fmt.Println("开始上传文件----- 1111111111")
	//文件切片
	fi, err := cloud_space.Diced(filePath)
	fi.Name = filename
	if err != nil {
		fmt.Println("文件切片失败", err)
		//文件切片失败
		out["Code"] = 1
		out["Message"] = "文件切片失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	fmt.Println("开始上传文件----- 2222222222222")

	//本节点网络地址
	vnodeinfo := nodeStore.NodeSelf.IdInfo.Id

	//添加文件所有者
	fi.AddFileOwner(vnodeinfo)

	//获取网络地址

	//保存文件索引到本地
	err = cloud_space.AddFileindexToSelf(fi)
	if err != nil {
		fmt.Println("保存文件索引到本地失败", err)
		out["Code"] = 1
		out["Message"] = "保存文件索引到本地失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	fmt.Println("开始上传文件----- 333333333333333")
	//这个文件索引归自己管理
	// cloud_space.AddFileindexToNet(fi, vnodeinfo.Vid)

	fmt.Println("开始上传文件----- 4444444444444444")
	//上传文件索引到网络中
	// err = store.UpNetFileindex(fi, &vnodeinfo.Vid)
	// if err != nil {
	// 	fmt.Println("上传文件索引到网络中失败", err)
	// 	out["Message"] = "上传文件索引到网络中失败"
	// 	out["Code"] = 1
	// } else {
	// 	fmt.Println("上传文件成功")
	// 	out["Code"] = 0
	// 	out["Size"] = fi.Size
	// 	out["HashName"] = fi.Hash.B58String()
	// 	//把文件块分散到网络中
	// 	//		go store.SyncFileChunkToPeer(fi)
	// }

	out["Code"] = 0
	out["Size"] = fi.Size
	out["HashName"] = fi.Hash.B58String()
	this.Data["json"] = out
	this.ServeJSON()
	return

	//	this.TplName = "store/index.tpl"
}

/*
	添加加密文件到云盘
*/
func (this *Index) AddCryptFile() {
	fmt.Println("开始上传文件-----")
	out := make(map[string]interface{})
	//判断是否注册域名
	nameinfo := name.FindNameToNet(nodeStore.NodeSelf.IdInfo.Id.B58String())
	//fmt.Printf("xxxxx%+v", nameinfo)
	if nameinfo == nil {
		fmt.Println("未注册域名")
		out["Code"] = 1
		out["Message"] = "未注册域名"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	if nameinfo.Deposit < cloud_space.DepositMin {
		fmt.Println("域名冻结押金不足")
		out["Code"] = 1
		out["Message"] = "域名冻结押金不足"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	// this.GetFile()
	hs, err := this.GetFiles("files[]")
	if err != nil {
		fmt.Println("获取文件头失败", err)
		out["Code"] = 1
		out["Message"] = "获取文件头失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}

	f, err := hs[0].Open()
	defer f.Close()
	if err != nil {
		fmt.Println("打开文件失败", err)
		out["Code"] = 1
		out["Message"] = "打开文件失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	filename := hs[0].Filename

	//	fi = &FileInfo{
	//		Name: p.FileName(),
	//		Type: p.Header.Get("Content-Type"),
	//	}

	//	f, h, err := this.Ctx.Request.FormFile("file")
	//	fmt.Println(f, h, err)

	//保存文件到本地
	newfile, err := os.OpenFile(filepath.Join(config.Store_temp, filename), os.O_RDWR|os.O_CREATE, os.ModePerm)
	defer newfile.Close()
	//验证空间是否已用完
	st, _ := newfile.Stat()
	vnodeinfo := fs.GetNotUseSpace(uint64(st.Size()))
	if vnodeinfo == nil {
		//	if !store.CheckSpace(uint64(st.Size())) {
		fmt.Println("空间不足", err)
		out["Code"] = 1
		out["Message"] = "空间不足"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	if err != nil {
		fmt.Println("保存文件到本地失败", err)
		out["Code"] = 1
		out["Message"] = "保存文件到本地失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	//	newfile, err := os.Create(filepath.Join(config.Store_temp, filename))
	// buf, err := ioutil.ReadAll(f)
	// newfile.Write(buf)
	if _, err := io.Copy(newfile, f); err != nil {
		fmt.Println("文件上传失败", err)
		//文件切片失败
		out["Code"] = 1
		out["Message"] = "文件上传失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	data, err := ioutil.ReadFile(filepath.Join(config.Store_temp, filename))
	if err != nil {
		fmt.Println("读取文件失败")
		out["Code"] = 1
		out["Message"] = "读取文件失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	dc, err := cloud_space.Encrypt(data)
	if err != nil {
		fmt.Println("加密文件失败")
		out["Code"] = 1
		out["Message"] = "加密文件失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	//加密后的文件名
	cfilename := filename + cloud_space.Suffix
	cryptfile, err := os.Create(filepath.Join(config.Store_fileinfo_cache, cfilename))
	if err != nil {
		fmt.Println(err)
		return
	}
	cryptfile.Write(dc)
	cryptfile.Close()
	//文件切片
	fi, err := cloud_space.Diced(filepath.Join(config.Store_fileinfo_cache, cfilename))
	fi.CryptUser = &nodeStore.NodeSelf.IdInfo.Id
	fi.Name = cfilename
	if err != nil {
		fmt.Println("文件切片失败", err)
		//文件切片失败
		out["Code"] = 1
		out["Message"] = "文件切片失败"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	// err = cloud_space.AddFileindexToSelf(fi, vnodeinfo.Vid)
	// if err != nil {
	// 	fmt.Println("保存文件索引到本地失败", err)
	// 	out["Code"] = 1
	// 	out["Message"] = "保存文件索引到本地失败"
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// //上传文件信息到网络中
	// err = store.UpNetFileindex(fi, &vnodeinfo.Vid)
	// if err != nil {
	// 	fmt.Println("上传文件索引到网络中失败", err)
	// 	out["Code"] = 1
	// 	out["Message"] = "上传文件索引到网络中失败"
	// } else {
	// 	out["Code"] = 0
	// 	out["Size"] = fi.Size
	// 	out["HashName"] = fi.Hash.B58String()
	// 	//把文件块分散到网络中
	// 	//		go store.SyncFileChunkToPeer(fi)
	// }

	out["Code"] = 0
	out["Size"] = fi.Size
	out["HashName"] = fi.Hash.B58String()
	this.Data["json"] = out
	this.ServeJSON()
	return

	//	this.TplName = "store/index.tpl"
}

/*
	下载文件
	1.查询文件信息
	2.把文件同步到本地
	3.传输文件给请求者
*/
func (this *Index) GetFile() {
	fn := this.Ctx.Input.Param(":hash")

	fid := virtual_node.AddressFromB58String(fn)

	//判断是直接打开还是下载
	dw := this.Ctx.Input.Param(":down")
	var isDown bool
	if dw == "down" {
		isDown = true
	} else {
		isDown = false
	}
	haveLocal := true
	//查询文件信息
	fileinfo := cloud_space.FindFileindex(fid)
	if fileinfo == nil {
		haveLocal = false
		var err error
		fileinfo, err = cloud_space.FindFileindexOpt(fn)
		if fileinfo == nil || err != nil {
			fmt.Println("网络中查找文件信息失败")
			return
		}
	}
	fmt.Printf("获取到的文件信息 %+v", fileinfo)
	if fileinfo.CryptUser != nil && fileinfo.CryptUser.B58String() != nodeStore.NodeSelf.IdInfo.Id.B58String() {
		fmt.Println("加密文件只有所有者有权下载")
		return
	}
	//把文件下载到本地
	err := cloud_space.DownloadFileOpt(fileinfo, isDown)
	if err != nil {
		// fmt.Println("下载文件失败", err)
		return
	}
	// fmt.Println("下载文件成功")

	vnodes := virtual_node.GetVnodeSelf()
	if len(vnodes) <= 0 {
		return
	}
	if !haveLocal {

		cloud_space.AddFileindexToLocal(fileinfo, vnodes[0].Vid)
	}
	filename := fileinfo.Name
	//解密
	if fileinfo.CryptUser != nil && fileinfo.CryptUser.B58String() == nodeStore.NodeSelf.IdInfo.Id.B58String() {
		data, err := ioutil.ReadFile(filepath.Join(config.Store_temp, fileinfo.Name))
		if err != nil {
			fmt.Println("读取文件失败")
			return
		}
		dc, err := cloud_space.Decrypt(data)
		if err != nil {
			fmt.Println("解密文件失败")
			return
		}
		//解密后的文件名
		filename = strings.TrimRight(filename, cloud_space.Suffix)
		cryptfile, err := os.Create(filepath.Join(config.Store_temp, filename))
		if err != nil {
			fmt.Println(err)
			return
		}
		cryptfile.Write(dc)
		cryptfile.Close()
		os.Remove(filepath.Join(config.Store_temp, fileinfo.Name))
	}
	//
	file, err := os.Open(filepath.Join(config.Store_temp, filename))
	if err != nil {
		// fmt.Println(err)
		return
	}
	io.Copy(this.Ctx.ResponseWriter, file)
	file.Close()

	//上传文件信息到网络中
	// go store.UpNetFileindex(fileinfo, &vnodes[0].Vid)
	return
	//	this.TplName = "store/index.tpl"

	//	io.Copy(this.Ctx.ResponseWriter, buf)
}
