package nodeStore

import (
	"mandela/core/utils/crypto/dh"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"time"

	"golang.org/x/crypto/ed25519"
)

/*
	保存节点的id
	ip地址
	不同协议的端口
*/
type Node struct {
	IdInfo               IdInfo    `json:"idinfo"`  //节点id信息，id字符串以16进制显示
	IsSuper              bool      `json:"issuper"` //是不是超级节点，超级节点有外网ip地址，可以为其他节点提供代理服务
	Addr                 string    `json:"addr"`    //外网ip地址
	TcpPort              uint16    `json:"tcpport"` //TCP端口
	IsApp                bool      `json:"isapp"`   //是不是手机端节点
	MachineID            int64     `json:"mid"`     //每个节点启动的时候生成一个随机数，用作判断多个节点使用同一个key连入网络的情况
	lastContactTimestamp time.Time //最后检查的时间戳
}

func (this *Node) FlashOnlineTime() {
	this.lastContactTimestamp = time.Now()

}

func (this *Node) Marshal() []byte {
	nodeBs, err := json.Marshal(this)
	if err != nil {
		return nil
	}
	return nodeBs
}

func ParseNode(bs []byte) (*Node, error) {
	node := new(Node)
	// err := json.Unmarshal(bs, node)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(node)
	//	fmt.Printf("dddd%+v %v", node, err)
	return node, err
}

//Id信息
type IdInfo struct {
	Id   AddressNet        `json:"id"`   //id，节点网络地址
	EPuk ed25519.PublicKey `json:"epuk"` //ed25519公钥，身份密钥的公钥
	CPuk dh.Key            `json:"cpuk"` //curve25519公钥,DH公钥
	V    uint32            `json:"v"`    //DH公钥版本，低版本将被弃用，用于自动升级更换DH公钥协议
	Sign []byte            `json:"sign"` //ed25519私钥签名,Sign(V + CPuk)
	// Ctype string           `json:"ctype"` //签名方法 如ecdsa256 ecdsa512
}

/*
	给idInfo签名
*/
func (this *IdInfo) SignDHPuk(prk ed25519.PrivateKey) {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, this.V)
	buf.Write(this.CPuk[:])
	this.Sign = ed25519.Sign(prk, buf.Bytes())
}

/*
	验证签名
*/
func (this *IdInfo) CheckSignDHPuk() bool {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, this.V)
	buf.Write(this.CPuk[:])

	return ed25519.Verify(this.EPuk, buf.Bytes(), this.Sign)
	// this.Sign = ed25519.Sign(prk, buf.Bytes())
}

/*
	解析一个idInfo
*/
func (this *IdInfo) Parse(code []byte) (err error) {
	// err = json.Unmarshal(code, this)
	decoder := json.NewDecoder(bytes.NewBuffer(code))
	decoder.UseNumber()
	err = decoder.Decode(this)
	return
}

//将此节点id详细信息构建为标准code
func (this *IdInfo) JSON() []byte {
	str, _ := json.Marshal(this)
	return str
}

/*
	检查idInfo是否合法
	1.地址生成合法
	2.签名正确
	@return   true:合法;false:不合法;
*/
func CheckIdInfo(idInfo IdInfo) bool {

	//检查地址是否是公钥生成
	// ok, _ := utils.Verify(idInfo.Puk, *idInfo.Id, idInfo.Sign)

	//验证签名
	ok := idInfo.CheckSignDHPuk()
	if !ok {
		return false
	}

	//验证地址
	return CheckPukAddr(idInfo.EPuk, idInfo.Id)
}

func Parse(idInfoByte []byte) IdInfo {
	idInfo := IdInfo{}
	idInfo.Parse(idInfoByte)
	return idInfo
}

/*
	临时id
*/
type TempId struct {
	SuperPeerId *AddressNet `json:"superpeerid"` //更新在线时间
	PeerId      *AddressNet `json:"peerid"`      //更新在线时间
	UpdateTime  int64       `json:"updatetime"`  //更新在线时间
}

/*
	创建一个临时id
*/
func NewTempId(superId, peerId *AddressNet) *TempId {
	return &TempId{
		SuperPeerId: superId,
		PeerId:      peerId,
		UpdateTime:  time.Now().Unix(),
	}
}
