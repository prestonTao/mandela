package wallet

import (
	"mandela/chain_witness_vote"

	"github.com/astaxie/beego"
)

type Index struct {
	beego.Controller
}

func (this *Index) Index() {
	//	names, _ := store.GetFileinfoToSelfAll()

	//	fmt.Println("网络文件个数为", len(names))
	//	this.Data["Names"] = names

	this.Data["CheckKey"] = chain_witness_vote.CheckKey()

	this.TplName = "wallet/index.tpl"
}

///*
//	添加文件到云盘
//*/
//func (this *Index) AddFile() {
//	out := make(map[string]interface{})

//	//	this.GetFile()
//	hs, err := this.GetFiles("files[]")
//	if err != nil {
//		fmt.Println("获取文件头失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}

//	f, err := hs[0].Open()
//	if err != nil {
//		fmt.Println("打开文件失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	filename := hs[0].Filename

//	//	fi = &FileInfo{
//	//		Name: p.FileName(),
//	//		Type: p.Header.Get("Content-Type"),
//	//	}

//	//	f, h, err := this.Ctx.Request.FormFile("file")
//	//	fmt.Println(f, h, err)

//	//保存文件到本地
//	newfile, err := os.OpenFile(filepath.Join(config.Store_temp, filename), os.O_RDWR|os.O_CREATE, os.ModePerm)
//	if err != nil {
//		fmt.Println("保存文件到本地失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	//	newfile, err := os.Create(filepath.Join(config.Store_temp, filename))
//	buf, err := ioutil.ReadAll(f)
//	newfile.Write(buf)
//	newfile.Close()
//	f.Close()

//	//文件切片
//	fi, err := store.Diced(filename)
//	if err != nil {
//		fmt.Println("文件切片失败", err)
//		//文件切片失败
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	fmt.Println("22222")

//	//	store.AddFileinfoToLocal(fi, true)
//	err = store.AddFileinfoToSelf(fi, true)
//	if err != nil {
//		fmt.Println("保存文件索引到本地失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	fmt.Println("33333")

//	//上传文件信息到网络中
//	err = store.UpNetFileinfo(fi)
//	if err != nil {
//		fmt.Println("上传文件索引到网络中失败", err)
//		out["Code"] = 1
//	} else {
//		out["Code"] = 0
//		out["HashName"] = fi.Hash.B58String()
//	}

//	this.Data["json"] = out
//	this.ServeJSON()
//	return

//	//	this.TplName = "store/index.tpl"
//}

///*
//	下载文件
//	1.查询文件信息
//	2.把文件同步到本地
//	3.传输文件给请求者
//*/
//func (this *Index) GetFile() {
//	fn := this.Ctx.Input.Param(":hash")
//	haveLocal := true

//	//查询文件信息
//	fileinfo := store.FindFileinfo(fn)
//	if fileinfo == nil {
//		haveLocal = false
//		var err error
//		fileinfo, err = store.FindFileinfoOpt(fn)
//		if fileinfo == nil || err != nil {
//			fmt.Println("网络中查找文件信息失败")
//			return
//		}
//	}
//	fmt.Println("获取到的文件信息", fileinfo)

//	//把文件下载到本地
//	err := store.DownloadFileOpt(fileinfo)
//	if err != nil {
//		fmt.Println("下载文件失败", err)
//		return
//	}
//	fmt.Println("下载文件成功")

//	if !haveLocal {
//		store.AddFileinfoToLocal(fileinfo, true)
//	}

//	//
//	file, err := os.Open(filepath.Join(config.Store_temp, fileinfo.Name))
//	if err != nil {
//		fmt.Println(err)
//		return
//	}

//	io.Copy(this.Ctx.ResponseWriter, file)
//	file.Close()

//	//上传文件信息到网络中
//	//	store.UpNetFileinfo(fileinfo)

//	//	this.TplName = "store/index.tpl"

//	//	io.Copy(this.Ctx.ResponseWriter, buf)
//}
