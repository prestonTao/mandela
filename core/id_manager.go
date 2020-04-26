/*
	加载本地配置文件中的idinfo
		1.读取并解析本地idinfo配置文件。
*/
package core

//import (
//	"encoding/json"
//	"io/ioutil"
//	"os"
//	"path/filepath"
//	gconfig "mandela/config"
//	"mandela/core/engine"
//	"mandela/core/nodeStore"
//)

//var Path_Id = filepath.Join(gconfig.Path_configDir, "idinfo.json") //超级节点地址列表文件路径
/*
	初始化变量，并且加载idinfo
*/
//func Initialization() {

//	//加载本地idinfo
//	LoadIdInfo()

//	//没有idinfo的新节点
//	if len(Init_IdInfo.Id) == 0 {
//		//连接网络并得到一个idinfo
//		GetId()
//	}
//}

/*
	加载本地的idInfo
*/
//func LoadIdInfo() (*nodeStore.IdInfo, error) {
//	data, err := ioutil.ReadFile(Path_Id)
//	//本地没有idinfo文件
//	if err != nil {
//		// fmt.Println("读取idinfo.json文件出错")
//		engine.Log.Warn("读取idinfo.json文件出错")
//		return nil, err
//	}

//	idInfo := new(nodeStore.IdInfo)
//	err = json.Unmarshal(data, &idInfo)
//	if err != nil {
//		// fmt.Println("解析idinfo.json文件错误")
//		engine.Log.Warn("解析idinfo.json文件错误")
//		return nil, err
//	}
//	return idInfo, nil
//}

/*
	保存idinfo到本地文件
*/
//func SaveIdInfo() {
//	fileBytes, _ := json.Marshal(nodeStore.NodeSelf.IdInfo)
//	file, _ := os.Create(Path_Id)
//	file.Write(fileBytes)
//	file.Close()
//}

/*
	连接超级节点，得到一个id
	@ addr   超级节点ip地址
*/
//func GetId() (newIdInfo *nodeStore.IdInfo, err error) {
//	zaro, _ := hex.DecodeString(utils.Str_zaro)
//	idInfo := nodeStore.IdInfo{
//		Id:          zaro,
//		CreateTime:  time.Now().Format("2006-01-02 15:04:05.999999999"),
//		Domain:      utils.GetRandomDomain(),
//		Name:        "",
//		Email:       "",
//		SuperNodeId: zaro,
//	}

//	// idInfo = nodeStore.IdInfo{
//	// 	Id:     Str_zaro,
//	// 	Name:   "nimei",
//	// 	Email:  "qqqqq@qq.com",
//	// 	Domain: "djfkafjkls",
//	// }
//	ip, port, err := addrm.GetSuperAddrOne(false)
//	if err != nil {
//		return nil, err
//	}
//	conn, err := net.Dial("tcp", ip+":"+strconv.Itoa(int(port)))
//	if err != nil {
//		err = errors.New("连接超级节点失败")
//		return
//	}

//	/*
//		向对方发送自己的名称
//	*/
//	lenght := int32(len(idInfo.Build()))
//	buf := bytes.NewBuffer([]byte{})
//	binary.Write(buf, binary.BigEndian, lenght)
//	buf.Write(idInfo.Build())
//	conn.Write(buf.Bytes())

//	/*
//		对方服务器创建好id后，发送给自己
//	*/
//	lenghtByte := make([]byte, 4)
//	io.ReadFull(conn, lenghtByte)
//	nameLenght := binary.BigEndian.Uint32(lenghtByte)
//	nameByte := make([]byte, nameLenght)
//	n, e := conn.Read(nameByte)
//	if e != nil {
//		err = e
//		return
//	}
//	//得到对方生成的名称
//	newIdInfo = new(nodeStore.IdInfo)
//	json.Unmarshal(nameByte[:n], newIdInfo)
//	conn.Close()

//	nodeStore.NodeSelf.IdInfo = *newIdInfo

//	return
//}
