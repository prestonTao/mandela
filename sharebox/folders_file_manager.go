/*
	管理根目录下文件结构
*/

package sharebox

import (
	sqldb "mandela/sqlite3_db"
	"bytes"
	"errors"
	"os"

	"io/ioutil"
	"path/filepath"
	"sync"
	// jsoniter "github.com/json-iterator/go"
)

// var json = jsoniter.ConfigCompatibleWithStandardLibrary

var rootDir = new(Dir)        //保存根目录
var fileHashs = new(sync.Map) //key:string=文件hash;value:File=文件信息;

func init() {
	rootDir.lock = new(sync.RWMutex)
}

/*
	加载共享文件夹目录
*/
func loadShareFolder() error {
	sf, err := new(sqldb.ShareFolder).GetAll()
	if err != nil {
		return err
	}
	for _, one := range sf {
		AddLocalShareFolders(one.Path)
	}
	return nil
}

/*
	定时同步文件索引
*/
func LoopSyncFileInfo() {
	// for range time.NewTimer(time.Hour).C {
	// 	fileHashs.Range(func(k, v interface{}) bool {
	// 		file := v.(*File)
	// 		FileIndex{}
	// 		return true
	// 	})
	// }

}

//保存目录结构
// var rootDirs = new(sync.Map) //key:string=文件hash;

/*
	扫描文件夹下的所有文件，并生成hash值
*/
func listFile(myfolder string, dir *Dir) {
	fileInfos, _ := ioutil.ReadDir(myfolder)
	for _, fileinfo := range fileInfos {
		filePath := filepath.Join(myfolder, fileinfo.Name())
		if fileinfo.IsDir() {
			newDir := NewDir(filePath)
			dir.AddDir(newDir)
			listFile(filePath, newDir)
		} else {
			hashName := BuildFileAddr(filePath)
			if hashName == nil {
				continue
			}
			newFile := NewFile(filePath, hashName)
			dir.AddFile(newFile)

			AddShareFile(filePath)
		}
	}
}

//根据文件路经查找对应的DIR
func getDir(file string, dirs []*Dir) (*Dir, error) {
	path := filepath.Dir(file)
	for _, dir := range dirs {
		if path == dir.Path {
			//fmt.Printf("%+v", dir)
			return dir, nil
		}
		if len(dir.Dirs) > 0 {
			return getDir(file, dir.Dirs)
		}
	}
	return nil, errors.New("no file")
}

//增加文件夹
func AddFold(path string) error {
	fileinfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fileinfo.IsDir() {
		return errors.New("该路径不是一个文件夹")
	}
	//判断是否重复
	for _, two := range rootDir.GetDirs() {
		if two.Path == path {
			//路径相同
			return errors.New("有相同文件夹")
		}
	}

	//添加文件监听
	err = WatcherFolder(path)
	if err != nil {
		return err
	}
	//遍历所有文件，将文件添加到共享列表
	dir := NewDir(path)
	dirs, err := getDir(path, rootDir.GetDirs())
	if err != nil {
		rootDir.Dirs[0].AddDir(dir)
	} else {
		dirs.AddDir(dir)
	}
	new(sqldb.ShareFolder).Add(path)

	go listFile(path, dir)
	return err
}

//删除文件夹
func DelFold(path string) error {
	paths := filepath.FromSlash(path)
	new(sqldb.ShareFolder).Del(paths)
	dirs, err := getDir(paths, rootDir.GetDirs())
	if err == nil {
		dirs.RemoveFile(paths)
	}
	return nil
}

//文件列表增加文件
func AddFile(filePath string) error {
	hashName := BuildFileAddr(filePath)
	if hashName == nil {
		return errors.New("build file hash error")
	}
	newFile := NewFile(filePath, hashName)
	dir, err := getDir(filePath, rootDir.GetDirs())
	if err == nil {
		dir.AddFile(newFile)
		AddShareFile(filePath)
	}
	return err
}

//文件列表删除文件
func DelFile(filePath string) error {
	dir, err := getDir(filePath, rootDir.GetDirs())
	if err != nil {
		return err
	}
	var temp []*File
	for _, v := range dir.Files {
		if filePath != v.Path {
			temp = append(temp, v)
		}
	}
	dir.Files = temp
	return nil
}

// func AddFile(path string, fileHash *utils.Multihash) error {
// 	fileinfo, err := os.Stat(path)
// 	if err != nil {
// 		return err
// 	}
// 	if fileinfo.IsDir() {
// 		newDir := NewDir(path)
// 		dirs := rootDir.GetDirs()
// 		if dirs == nil {
// 			rootDir.AddDir(newDir)
// 			return nil
// 		}
// 		return nil
// 	} else {
// 		file := NewFile(path, fileHash)
// 		for _, one := range rootDir.GetDirs() {
// 			//找到这个文件的根目录
// 			if filepath.HasPrefix(path, one.Path) {
// 				//根目录中找到所属文件夹
// 				paths := utils.FilePathSplit(path)
// 				rootPaths := utils.FilePathSplit(one)
// 				dstDir
// 				index := len(rootPaths)
// 				for i := 0; i < index; i++ {
// 					for _, two := range one.GetDirs() {
// 						if two.Name == paths[index] {
// 							two.
// 						}
// 					}

// 				}

// 				for i = 1; i < len(paths)-1; i++ {
// 					filepath.Join(paths[:i])
// 				}
// 				one.AddFile(file)
// 			}
// 		}

// 		paths := filepath.SplitList(path)
// 		fmt.Println("分割路径后111：", paths)
// 		paths, err := filepath.Glob(path)
// 		fmt.Println("分割路径后222：", paths, err)
// 		dirName := filepath.Dir(path)
// 		fmt.Println("目录名称", dirName)
// 		parentDirName := filepath.Dir(dirName)
// 		fmt.Println("父目录名称", parentDirName)
// 		dirName, fileName := filepath.Split(path)
// 		fmt.Println("目录名称和文件名称", dirName, fileName)

// 		ok := filepath.HasPrefix(path, dirName)
// 		fmt.Println("是否有前缀", ok)

// 		paths = utils.FilePathSplit(path)
// 		fmt.Println("拆分后的路径", paths)
// 	}
// 	return nil
// }

/*
	创建一个文件夹
*/
func NewDir(path string) *Dir {
	_, fileName := filepath.Split(path)
	dir := Dir{
		Name:  fileName,          //文件夹名称
		Path:  path,              //绝对路径
		Dirs:  make([]*Dir, 0),   //文件夹中的文件夹
		Files: make([]*File, 0),  //文件夹中的文件
		lock:  new(sync.RWMutex), //锁
	}
	return &dir
}

/*
	文件夹
*/
type Dir struct {
	Public bool          //是否公开
	Name   string        //文件夹名称
	Path   string        //绝对路径
	Dirs   []*Dir        //文件夹中的文件夹
	Files  []*File       //文件夹中的文件
	lock   *sync.RWMutex //锁
}

/*
	文件夹中添加文件夹
*/
func (this *Dir) AddDir(dir *Dir) {
	this.lock.Lock()
	if this.Dirs == nil {
		this.Dirs = make([]*Dir, 0)
	}
	this.Dirs = append(this.Dirs, dir)
	this.lock.Unlock()
}

/*
	文件夹中删除文件夹
*/
func (this *Dir) RemoveFile(path string) {
	this.lock.Lock()
	temp := make([]*Dir, 0)
	for i, one := range this.Dirs {
		if one.Path == path {
			temp = this.Dirs[:i]
			temp = append(temp, this.Dirs[i+1:]...)
			break
		}
	}
	this.Dirs = temp
	this.lock.Unlock()
}

/*
	获取文件夹列表
*/
func (this *Dir) GetDirs() (dirs []*Dir) {
	this.lock.RLock()
	dirs = this.Dirs
	this.lock.RUnlock()
	return
}

/*
	文件夹中添加文件
*/
func (this *Dir) AddFile(file *File) {
	this.lock.Lock()
	if this.Files == nil {
		this.Files = make([]*File, 0)
	}
	this.Files = append(this.Files, file)
	this.lock.Unlock()
	fileHashs.Store(file.FileInfo.B58String(), file)
}

/*
	获取文件夹列表
*/
func (this *Dir) GetFiles() (files []*File) {
	this.lock.RLock()
	files = this.Files
	this.lock.RUnlock()
	return
}

/*
	格式化
*/
func (this *Dir) JSON() ([]byte, error) {
	dirvo := this.conversionVO()
	return json.Marshal(dirvo)
}

/*
	转化为VO
*/
func (this *Dir) conversionVO() *DirVO {
	vo := DirVO{
		Public: this.Public,        //是否公开
		Name:   this.Name,          //文件夹名称
		Path:   this.Path,          //绝对路径
		Dirs:   make([]*DirVO, 0),  //文件夹中的文件夹
		Files:  make([]*FileVO, 0), //文件夹中的文件
	}
	for _, one := range this.Dirs {
		dirvoOne := one.conversionVO()
		vo.Dirs = append(vo.Dirs, dirvoOne)
	}
	for _, one := range this.Files {
		filevoOne := one.conversionVO()
		vo.Files = append(vo.Files, filevoOne)
	}
	return &vo
}

/*
	文件
*/
type File struct {
	// Public   bool             //是否公开
	Name     string   //文件名称
	Path     string   //绝对路径
	FileInfo FileAddr //文件hash值
}

/*
	创建文件
*/
func NewFile(path string, fileHash FileAddr) *File {
	_, fileName := filepath.Split(path)
	file := File{
		Name:     fileName, //文件名称
		Path:     path,     //绝对路径
		FileInfo: fileHash, //文件hash值
	}
	return &file
}

/*
	转化为VO
*/
func (this *File) conversionVO() *FileVO {
	vo := FileVO{
		Name:     this.Name,                 //文件名称
		Path:     this.Path,                 //绝对路径
		FileInfo: this.FileInfo.B58String(), //文件hash值
	}
	return &vo
}

type DirVO struct {
	Public bool      //是否公开
	Name   string    //文件夹名称
	Path   string    //绝对路径
	Dirs   []*DirVO  //文件夹中的文件夹
	Files  []*FileVO //文件夹中的文件
}

func ParseDirVO(bs *[]byte) *DirVO {
	dirvo := new(DirVO)
	// err := json.Unmarshal(*bs, dirvo)
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(dirvo)
	if err != nil {
		return nil
	}
	return dirvo
}

type FileVO struct {
	Name     string //文件名称
	Path     string //绝对路径
	FileInfo string //文件hash值
}

/*
	通过文件hash查找文件信息
*/
func FindFile(hash string) *File {
	fileItr, ok := fileHashs.Load(hash)
	if ok {
		file := fileItr.(*File)
		return file
	}
	return nil
}
