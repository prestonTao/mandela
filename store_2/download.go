package store

import (
	"mandela/config"
	sql "mandela/sqlite3_db"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var (
	DP       = make(map[string]*DownProc) //下载进度
	StopDown = make(map[string]bool)      //暂停
)

func init() {
	go func() {
		for range time.NewTicker(5 * time.Second).C {
			syncData()
		}
	}()
}

//同步数据到数据库
func syncData() {
	for key, val := range DP {
		d := sql.Downprogress{
			Hash:  key,
			Rate:  val.Rate,
			Speed: val.Speed,
		}
		d.Update()
		if val.Rate == 100 {
			delete(DP, key)
		}
	}
}

//下载进度
type DownProc struct {
	Hash     string
	Fi       *FileInfo
	Rate     float32
	Speed    uint64
	Size     uint64
	SizeList map[string]uint64
	Time     int64
	Lock     *sync.Mutex
}

func NewDownProc(fi *FileInfo) *DownProc {
	dp := DownProc{
		Hash:     fi.Hash.B58String(),
		Fi:       fi,
		Time:     time.Now().Unix(),
		Size:     0,
		SizeList: make(map[string]uint64),
		Lock:     &sync.Mutex{},
	}
	dp.AddDownProc()
	DP[dp.Hash] = &dp         //下载百分比
	StopDown[dp.Hash] = false //是否暂停
	return &dp
}

//增加下载文件
func (dp *DownProc) AddDownProc() error {
	defer dp.Lock.Unlock()
	dp.Lock.Lock()
	d := sql.Downprogress{
		Hash:     dp.Hash,
		FileInfo: dp.Fi.JSON(),
		Rate:     1,
		State:    1,
	}
	return d.Add()
}

//更新下载进度与速度
func (dp *DownProc) UpdateDownProc(chunk string, size uint64) error {
	defer dp.Lock.Unlock()
	dp.Lock.Lock()
	dp.SizeList[chunk] = size
	dp.Size = 0
	for _, val := range dp.SizeList {
		dp.Size = dp.Size + val
	}
	if dp.Size > dp.Fi.Size {
		dp.Size = dp.Fi.Size
	}
	rate := (float32(dp.Size) / float32(dp.Fi.Size)) * float32(100)
	tcz := (time.Now().Unix() - dp.Time) //所用时间
	if tcz == 0 {
		tcz = 1
	}
	speed := int64(dp.Size) / tcz
	dp.Rate = rate
	dp.Speed = uint64(speed)
	DP[dp.Hash] = dp
	return nil
}

//获取文件大小
func (dp *DownProc) GetSize(path string) uint64 {
	defer dp.Lock.Unlock()
	dp.Lock.Lock()
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return uint64(info.Size())
}

//下载进度(RPC调用)
type DProgress struct {
	Hash   string
	Name   string  //文件名
	Rate   float32 //进度比例
	Num    uint64  //已下载字节
	AllNum uint64  //总共字节
	Speed  uint64  //下载速度
	State  int     //状态 1 下载中 2 暂停
}

//获取单个文件下载进度

func DownloadProgressOne(fi *FileInfo) (dp DProgress) {
	if fi == nil {
		return
	}
	db := sql.Downprogress{}
	row := db.Get(fi.Hash.B58String())
	if val, ok := DP[fi.Hash.B58String()]; ok {
		dp.Rate = val.Rate
		dp.Speed = val.Speed
	} else {
		dp.Rate = row.Rate
		dp.Speed = row.Speed
	}
	dp.Hash = fi.Hash.B58String()
	dp.Name = fi.Name
	dp.State = row.State
	dp.AllNum = fi.Size
	dp.Num = fi.Size * uint64(dp.Rate) / 100
	return
}

//获取所有文件下载进度
func DownloadProgress() (dps []DProgress) {
	db := sql.Downprogress{}
	rows := db.List()
	if len(rows) == 0 {
		return
	}
	for _, row := range rows {
		dp := DProgress{}
		if val, ok := DP[row.Hash]; ok {
			dp.Rate = val.Rate
			dp.Speed = val.Speed
		} else {
			dp.Rate = row.Rate
			dp.Speed = row.Speed
		}
		// if dp.Rate == 100 {
		// 	continue
		// }
		fi, err := ParseFileinfo(row.FileInfo)
		if err != nil {
			fmt.Println(err)
		}
		dp.Hash = fi.Hash.B58String()
		dp.Name = fi.Name
		dp.State = row.State
		dp.AllNum = fi.Size
		dp.Num = fi.Size * uint64(dp.Rate) / 100
		dps = append(dps, dp)
	}
	return
}

type DComplete struct {
	Name string
	Size int64
	Path string
}

//已下载的文件
func DownLoadComplete() (dcs []DComplete) {
	fileInfos, _ := ioutil.ReadDir(config.Store_temp)
	for _, fileinfo := range fileInfos {
		//fmt.Printf("%+v", fileinfo)
		dc := DComplete{
			Name: fileinfo.Name(),
			Size: fileinfo.Size(),
			Path: config.Store_temp,
		}
		dcs = append(dcs, dc)
	}
	fileInfosf, _ := ioutil.ReadDir(config.Store_files)
	for _, fileinfo := range fileInfosf {
		//fmt.Printf("%+v", fileinfo)
		dc := DComplete{
			Name: fileinfo.Name(),
			Size: fileinfo.Size(),
			Path: config.Store_files,
		}
		dcs = append(dcs, dc)
	}
	return
}

//暂停下载
func DownLoadStop(hash string) error {
	StopDown[hash] = true
	d := sql.Downprogress{
		Hash:  hash,
		State: 2,
	}
	d.Update()
	return nil
}

//删除下载列表
func DwonLoadDel(hash string) error {
	d := sql.Downprogress{
		Hash: hash,
	}
	d.Delete()
	delete(DP, hash)
	delete(StopDown, hash)
	return nil
}

//获取已用空间大小
func getSpaceSize() (size uint64) {
	fileInfos, _ := ioutil.ReadDir(config.Store_fileinfo_self)
	for _, fileinfo := range fileInfos {
		fi := parseFile(filepath.Join(config.Store_fileinfo_self, fileinfo.Name()))
		if fi != nil {
			size += fi.Size
		}
	}
	return
}

//引用空间索引列表，用于验证空间使用情况
type SpaceList struct {
	List []*FileInfo
}

//获取空间文件列表
func (sl *SpaceList) GetSpaceList() {
	fileInfos, _ := ioutil.ReadDir(config.Store_fileinfo_self)
	for _, fileinfo := range fileInfos {
		fi := parseFile(filepath.Join(config.Store_fileinfo_self, fileinfo.Name()))
		if fi != nil {
			sl.List = append(sl.List, fi)
		}
	}
}
func (sl *SpaceList) Json() []byte {
	rs, err := json.Marshal(sl)
	if err != nil {
		fmt.Println(err)
	}
	return rs
}
func ParseSpaceList(rs []byte) *SpaceList {
	sl := new(SpaceList)
	// err := json.Unmarshal(rs, sl)
	decoder := json.NewDecoder(bytes.NewBuffer(rs))
	decoder.UseNumber()
	err := decoder.Decode(sl)
	if err != nil {
		fmt.Println(err)
	}
	return sl
}
