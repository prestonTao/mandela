package sharebox

import (
	"mandela/config"
	"mandela/sharebox"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	*sharebox.FileIndex
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
	// _, fileInfos := sharebox.GetFileinfoToSelfAll()
	// // kind := this.GetString("kind")

	// var result map[string]interface{} = make(map[string]interface{})
	// var fileinfoList fileinfoObjList
	// // findNextFile:
	// for _, fileinfo := range fileInfos {
	// 	var fiobj fileinfoObj
	// 	fiobj.HasCode = fileinfo.Hash.B58String()
	// 	fiobj.FileInfo = fileinfo

	// 	fileinfoList = append(fileinfoList, &fiobj)

	// }
	// sort.Sort(fileinfoList)
	// result["data"] = fileinfoList
	// result["status"] = 200
	// this.Data["json"] = &result
	// this.ServeJSON()
}

/*
	添加文件到云盘
*/
func (this *Index) AddFile() {
	out := make(map[string]interface{})
	// //	this.GetFile()
	// hs, err := this.GetFiles("files[]")
	// if err != nil {
	// 	fmt.Println("获取文件头失败", err)
	// 	out["Code"] = 1
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }

	// f, err := hs[0].Open()
	// if err != nil {
	// 	fmt.Println("打开文件失败", err)
	// 	out["Code"] = 1
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// filename := hs[0].Filename

	// //	fi = &FileInfo{
	// //		Name: p.FileName(),
	// //		Type: p.Header.Get("Content-Type"),
	// //	}

	// //	f, h, err := this.Ctx.Request.FormFile("file")
	// //	fmt.Println(f, h, err)

	// //保存文件到本地
	// newfile, err := os.OpenFile(filepath.Join(config.Store_temp, filename), os.O_RDWR|os.O_CREATE, os.ModePerm)
	// if err != nil {
	// 	fmt.Println("保存文件到本地失败", err)
	// 	out["Code"] = 1
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// //	newfile, err := os.Create(filepath.Join(config.Store_temp, filename))
	// buf, err := ioutil.ReadAll(f)
	// newfile.Write(buf)
	// newfile.Close()
	// f.Close()

	// //文件切片
	// fi, err := sharebox.Diced(filename)
	// if err != nil {
	// 	fmt.Println("文件切片失败", err)
	// 	//文件切片失败
	// 	out["Code"] = 1
	// 	this.Data["json"] = out
	// 	this.ServeJSON()
	// 	return
	// }
	// fmt.Println("22222")

	// //	sharebox.AddFileinfoToLocal(fi, true)
	// //如果是app上传，则本节点自己不能管理
	// if this.GetString("appos") == "" {
	// 	err = sharebox.AddFileinfoToSelf(fi, true)
	// 	if err != nil {
	// 		fmt.Println("保存文件索引到本地失败", err)
	// 		out["Code"] = 1
	// 		this.Data["json"] = out
	// 		this.ServeJSON()
	// 		return
	// 	}
	// 	fmt.Println("33333")
	// }
	// //上传文件信息到网络中
	// err = sharebox.UpNetFileinfo(fi)
	// if err != nil {
	// 	fmt.Println("上传文件索引到网络中失败", err)
	// 	out["Code"] = 1
	// } else {
	// 	out["Code"] = 0
	// 	out["Size"] = fi.Size
	// 	out["HashName"] = fi.Hash.B58String()
	// 	//把文件块分散到网络中
	// 	go sharebox.SyncFileChunkToPeer(fi)
	// }

	this.Data["json"] = out
	this.ServeJSON()
	return

	//	this.TplName = "sharebox/index.tpl"
}

/*
	下载文件
	1.查询文件信息
	2.把文件同步到本地
	3.传输文件给请求者
*/
func (this *Index) GetFile() {
	fn := this.Ctx.Input.Param(":hash")
	// haveLocal := true
	// fmt.Println("11111111111", fn)

	//查询文件信息
	fileinfo := sharebox.FindFileindex(fn)
	if fileinfo == nil {
		// haveLocal = false
		var err error
		fileinfo, err = sharebox.DownloadFileindexOpt(fn)
		if fileinfo == nil || err != nil {
			fmt.Println("网络中查找文件信息失败")
			return
		}
	}
	fmt.Println("获取到的文件信息", fileinfo)

	fileAbsPath := ""

	shareFile := sharebox.FindFile(fileinfo.Hash.B58String())
	if shareFile == nil {
		//把文件下载到本地
		err := sharebox.DownloadFileOpt(fileinfo)
		if err != nil {
			fmt.Println("下载文件失败", err)
			return
		}
		fmt.Println("下载文件成功")
		fileAbsPath = filepath.Join(config.Store_files, fileinfo.Hash.B58String())
		fileNowPath := filepath.Join(config.Store_files, fileinfo.Name)
		//TODO 下载文件成功后，给他加个后缀名
		os.Rename(fileAbsPath, fileNowPath)
		fileAbsPath = fileNowPath
		// newfile, err := os.OpenFile(filepath.Join(config.Store_files, fileinfo.Name), os.O_RDWR|os.O_CREATE, os.ModePerm)
		// file, err := os.Open(fileAbsPath)
		// if err != nil {
		// 	// fmt.Println(err)
		// 	return
		// }

		// io.Copy(newfile, file)
		// file.Close()
		// newfile.Close()
	} else {
		fileAbsPath = shareFile.Path
	}

	// if !haveLocal {
	// 	sharebox.AddFileinfoToLocal(fileinfo, true)
	// }

	//
	file, err := os.Open(fileAbsPath)
	if err != nil {
		// fmt.Println(err)
		return
	}

	io.Copy(this.Ctx.ResponseWriter, file)
	file.Close()

	//上传文件信息到网络中
	// go sharebox.UpNetFileinfo(fileinfo)
	return
}
