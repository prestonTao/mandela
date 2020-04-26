package utils

import (
	"mandela/core/engine"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

/*
	检查目录是否存在，不存在则创建
*/
func CheckCreateDir(dir_path string) {
	if ok, err := PathExists(dir_path); err == nil && !ok {
		Mkdir(dir_path)
	}
}

/*
	判断一个路径的文件是否存在
*/
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

/*
	递归创建目录
*/
func Mkdir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	//	err := os.Mkdir(path, os.ModeDir)
	if err != nil {
		//		fmt.Println("创建文件夹失败", path, err)
		return err
	}
	return nil
}

/*
	保存对象为json格式
*/
func SaveJsonFile(name string, o interface{}) error {
	bs, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return SaveFile(name, &bs)
}

/*
	保存文件
	保存文件步骤：
	1.创建临时文件
	2.
*/
func SaveFile(name string, bs *[]byte) error {
	//创建临时文件
	now := strconv.Itoa(int(time.Now().Unix()))
	tempname := name + "." + now
	file, err := os.OpenFile(tempname, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		file.Close()
		return err
	}
	_, err = file.Write(*bs)
	if err != nil {
		file.Close()
		return err
	}
	file.Close()
	//删除旧文件
	ok, err := PathExists(name)
	if err != nil {
		engine.Log.Info("删除旧文件失败", err)
		return err
	}
	if ok {
		err = os.Remove(name)
		if err != nil {
			return err
		}
	}

	//重命名文件
	err = os.Rename(tempname, name)
	if err != nil {
		return err
	}
	return nil
}

/*
	把文件全路径切分成子路径
*/
func FilePathSplit(path string) []string {
	names := make([]string, 0)
	// var beforName = path
	var dirName = path
	var fileName string
	for i := 0; i < 10; i++ {
		_, fileName = filepath.Split(dirName)
		if fileName == "" {
			names = append(names, dirName)
			break
		}
		dirName = filepath.Dir(dirName)
		names = append(names, fileName)
	}

	myReverse(names)
	return names
}

func myReverse(l []string) {
	for i := 0; i < int(len(l)/2); i++ {
		li := len(l) - i - 1
		l[i], l[li] = l[li], l[i]
	}
}
