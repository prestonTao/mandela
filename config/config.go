package config

import (
	"mandela/core/config"
	"mandela/core/engine"
	"mandela/core/utils"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	Path_configDir     = "conf"            //配置文件存放目录
	Path_config        = "config.json"     //配置文件名称
	Core_keystore      = "keystore.key"    //密钥文件
	Core_addr_prk      = "addr_ec_prk.pem" //地址私钥文件名称
	Core_addr_puk      = "addr_ec_puk.pem" //地址公钥文件名称
	Core_addr_prk_type = "EC PRIVATE KEY"  //地址私钥文件抬头
	Core_addr_puk_type = "EC PUBLIC KEY"   //地址公钥文件抬头
)

const (
	Name_prk = "name_ec_prk.pem" //地址私钥文件名称
	Name_puk = "name_ec_puk.pem" //地址公钥文件名称
)

const (
	Store_path_dir            = "store" //本地共享文件存储目录名称
	Store_path_fileinfo_self  = "self"  //自己上传的文件索引存储目录名称
	Store_path_fileinfo_local = "local" //本地下载过的文件索引存储目录名称
	Store_path_fileinfo_net   = "net"   //网络需要保存的文件索引存储目录名称
	Store_path_fileinfo_cache = "cache" //缓存中保存的文件索引存储目录名称
	Store_path_temp           = "temp"  //临时文件夹，本地上传存放目录，存放未切片的完整文件
	Store_path_files          = "files" //带扩展名的完整文件
	IsRemoveStore             = false   //启动时删除本地所有文件分片及分片索引
	//	IsCreateId                = true    //启动时是否要创建新的id

	HashCode    = utils.SHA3_256 //
	NodeIDLevel = 256            //节点id比特位数
)

var (
	WebAddr                = "" //
	WebPort         uint16 = 0  //本地监听端口
	Web_path_static        = "" //网页静态文件路径
	Web_path_views         = "" //网页模板文件路径
	AddrPre                = "" //收款地址前缀
	//	NetIds                 = []byte{0} //网络号
	NetType_release = "release" //正式网络
	NetType         = "test"    //网络类型:正式网络release/测试网络test
	RPCUser         = ""        //
	RPCPassword     = ""        //
)

var (
	Store_dir            string = filepath.Join(Store_path_dir)                            //本地共享文件存储目录路径
	Store_fileinfo_self  string = filepath.Join(Store_path_dir, Store_path_fileinfo_self)  //自己上传的文件索引存储目录路径
	Store_fileinfo_local string = filepath.Join(Store_path_dir, Store_path_fileinfo_local) //本地下载过的文件索引存储目录路径
	Store_fileinfo_net   string = filepath.Join(Store_path_dir, Store_path_fileinfo_net)   //网络需要保存的文件索引存储目录路径
	Store_fileinfo_cache string = filepath.Join(Store_path_dir, Store_path_fileinfo_cache) //缓存中保存的文件索引存储目录路径
	Store_temp           string = filepath.Join(Store_path_dir, Store_path_temp)           //临时文件夹，本地上传存放目录，存放未切片的完整文件
	Store_files          string = filepath.Join(Store_path_dir, Store_path_files)          //存放带扩展名的完整文件
)

func init() {

	// fmt.Println("1111111111111")
	ok, err := utils.PathExists(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		// fmt.Println("22222222222222")
		panic("检查配置文件错误：" + err.Error())
		return
	}
	// fmt.Println("3333333333333")
	if !ok {
		// fmt.Println("4444444444444")
		cfi := new(Config)
		cfi.Port = 9981
		bs, _ := json.Marshal(cfi)
		// fmt.Println("5555555555555555")
		f, err := os.OpenFile(filepath.Join(Path_configDir, Path_config), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			// fmt.Println("666666666666666666")
			panic("创建配置文件错误：" + err.Error())
			return
		}
		// fmt.Println("77777777777777777")
		_, err = f.Write(bs)
		if err != nil {
			// fmt.Println("88888888888888")
			panic("写入配置文件错误：" + err.Error())
			return
		}
		// fmt.Println("999999999999999")
		f.Close()
	}
	// fmt.Println("11111111111111111111111111")
	bs, err := ioutil.ReadFile(filepath.Join(Path_configDir, Path_config))
	if err != nil {
		// fmt.Println("2222222222222222222222222222")
		panic("读取配置文件错误：" + err.Error())
		return
	}
	// fmt.Println("3333333333333333333")
	cfi := new(Config)
	// err = json.Unmarshal(bs, cfi)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(cfi)
	if err != nil {
		// fmt.Println("44444444444444444444444444")
		panic("解析配置文件错误：" + err.Error())
		return
	}
	// fmt.Printf("55555555555555555555555 %+v", cfi)
	config.Init_LocalIP = cfi.IP
	config.Init_LocalPort = cfi.Port
	config.Init_GatewayPort = cfi.Port
	Web_path_static = cfi.WebStatic
	Web_path_views = cfi.WebViews
	engine.Netid = cfi.Netid
	WebAddr = cfi.WebAddr
	WebPort = cfi.WebPort
	Miner = cfi.Miner
	NetType = cfi.NetType
	//	NetType = NetType_release //正式版所有节点都为release
	AddrPre = cfi.AddrPre
	RPCUser = cfi.RpcUser
	RPCPassword = cfi.RpcPassword

}

type Config struct {
	Netid       uint32 `json:"netid"`       //
	IP          string `json:"ip"`          //ip地址
	Port        uint16 `json:"port"`        //监听端口
	WebAddr     string `json:"WebAddr"`     //
	WebPort     uint16 `json:"WebPort"`     //
	WebStatic   string `json:"WebStatic"`   //
	WebViews    string `json:"WebViews"`    //
	RpcServer   bool   `json:"RpcServer"`   //
	RpcUser     string `json:"RpcUser"`     //
	RpcPassword string `json:"RpcPassword"` //
	Miner       bool   `json:"miner"`       //本节点是否是矿工
	NetType     string `json:"NetType"`     //正式网络release/测试网络test
	AddrPre     string `json:"AddrPre"`     //收款地址前缀
}
