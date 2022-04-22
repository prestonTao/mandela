package message_center

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto/dh"
	"mandela/protos/go_protos"

	"github.com/gogo/protobuf/proto"
)

/*
	获取节点地址和身份公钥
*/
func SearchAddress(c engine.Controller, msg engine.Packet, message *Message) {

	if !message.CheckSendhash() {
		return
	}

	// prk := keystore.GetDHKeyPair().KeyPair.GetPrivateKey()
	// puk := keystore.GetDHKeyPair().KeyPair.GetPublicKey()
	// fmt.Println("密钥对公钥", hex.EncodeToString(nodeStore.NodeSelf.IdInfo.CPuk[:]))
	// fmt.Println("密钥对", hex.EncodeToString(prk[:]), "\n",
	// 	hex.EncodeToString(puk[:]))
	// fmt.Println("获取节点地址和身份公钥")

	//回复消息
	// data := nodeStore.NodeSelf.IdInfo.JSON()
	data, _ := nodeStore.NodeSelf.IdInfo.Proto()
	SendP2pReplyMsg(message, config.MSGID_SearchAddr_recv, &data)

}

/*
	获取节点地址和身份公钥_返回
*/
func SearchAddress_recv(c engine.Controller, msg engine.Packet, message *Message) {

	if !message.CheckSendhash() {
		return
	}

	// fmt.Println("收到Hello消息", string(*message.Body.Content))

	// idinfo := nodeStore.Parse(*message.Body.Content)
	idinfo, _ := nodeStore.ParseIdInfo(*message.Body.Content)
	sni := SearchNodeInfo{
		Id:      message.Head.Sender,
		SuperId: message.Head.SenderSuperId,
		CPuk:    idinfo.CPuk,
	}
	// bs, _ := json.Marshal(sni)
	bs, _ := sni.Proto() //json.Marshal(sni)

	// flood.ResponseWait(config.CLASS_security_searchAddr, hex.EncodeToString(message.Body.Hash), &bs)
	flood.ResponseWait(config.CLASS_security_searchAddr, utils.Bytes2string(message.Body.Hash), &bs)

}

type SearchNodeInfo struct {
	Id      *nodeStore.AddressNet //
	SuperId *nodeStore.AddressNet //
	CPuk    dh.Key                //
}

func (this *SearchNodeInfo) Proto() ([]byte, error) {
	snip := go_protos.SearchNodeInfo{
		Id:      *this.Id,
		SuperId: *this.SuperId,
		CPuk:    this.CPuk[:],
	}
	return snip.Marshal()
}

// func ParserSearchNodeInfo(bs *[]byte) (*SearchNodeInfo, error) {
// 	sni := new(SearchNodeInfo)
// 	// err := json.Unmarshal(*bs, sni)
// 	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
// 	decoder.UseNumber()
// 	err := decoder.Decode(sni)
// 	return sni, err
// }

func ParserSearchNodeInfo(bs []byte) (*SearchNodeInfo, error) {
	snip := new(go_protos.SearchNodeInfo)
	err := proto.Unmarshal(bs, snip)
	if err != nil {
		return nil, err
	}
	var cpuk dh.Key = [32]byte{}
	copy(cpuk[:], snip.CPuk)

	// id := nodeStore.AddressNet(snip.Id)
	// superId := nodeStore.AddressNet(snip.SuperId)
	sni := SearchNodeInfo{
		// Id:      &id,
		// SuperId: &superId,
		CPuk: cpuk,
	}
	if snip.Id != nil && len(snip.Id) > 0 {
		id := nodeStore.AddressNet(snip.Id)
		sni.Id = &id
	}

	if snip.SuperId != nil && len(snip.SuperId) > 0 {
		superId := nodeStore.AddressNet(snip.SuperId)
		sni.SuperId = &superId
	}

	return &sni, nil
	// err := json.Unmarshal(*bs, sni)
	// decoder := json.NewDecoder(bytes.NewBuffer(bs))
	// decoder.UseNumber()
	// err := decoder.Decode(sni)
	// return sni, err
}

type ShareKey struct {
	// IV_DH_PUK dh.Key//向量
	Idinfo   nodeStore.IdInfo //身份密钥公钥
	A_DH_PUK dh.Key           //A公钥
	B_DH_PUK dh.Key           //B公钥
}

func (this *ShareKey) Proto() ([]byte, error) {
	// var cpuk dh.Key = [32]byte{}
	// copy(cpuk[:], this.Idinfo.CPuk)
	idinfo := go_protos.IdInfo{
		Id:   this.Idinfo.Id,
		EPuk: this.Idinfo.EPuk,
		CPuk: this.Idinfo.CPuk[:],
		V:    this.Idinfo.V,
		Sign: this.Idinfo.Sign,
	}
	skp := go_protos.ShareKey{
		Idinfo:   &idinfo,
		A_DH_PUK: this.A_DH_PUK[:],
		B_DH_PUK: this.B_DH_PUK[:],
	}
	return skp.Marshal()
}

func ParseShareKey(bs []byte) (*ShareKey, error) {
	skp := new(go_protos.ShareKey)
	err := proto.Unmarshal(bs, skp)
	if err != nil {
		return nil, err
	}
	var cpuk dh.Key = [32]byte{}
	copy(cpuk[:], skp.Idinfo.CPuk)

	idinfo := nodeStore.IdInfo{
		Id:   skp.Idinfo.Id,
		EPuk: skp.Idinfo.EPuk,
		CPuk: cpuk,
		V:    skp.Idinfo.V,
		Sign: skp.Idinfo.Sign,
	}
	var apuk, bpuk dh.Key = [32]byte{}, [32]byte{}
	copy(apuk[:], skp.A_DH_PUK)
	copy(bpuk[:], skp.B_DH_PUK)

	shareKey := ShareKey{
		Idinfo:   idinfo,
		A_DH_PUK: apuk,
		B_DH_PUK: bpuk,
	}
	return &shareKey, nil
}

/*
	获取节点地址和身份公钥
*/
func CreatePipe(c engine.Controller, msg engine.Packet, message *Message) {

	if !message.CheckSendhash() {
		return
	}

	shareKey, err := ParseShareKey(*message.Body.Content)
	// fmt.Println("收到Hello消息", string(*message.Body.Content))

	// shareKey := new(ShareKey)
	// // err := json.Unmarshal(*message.Body.Content, shareKey)
	// decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	// decoder.UseNumber()
	// err := decoder.Decode(shareKey)
	if err != nil {
		return
	}

	sk := dh.KeyExchange(dh.NewDHPair(keystore.GetDHKeyPair().KeyPair.GetPrivateKey(), shareKey.Idinfo.CPuk))
	sharedHka := dh.KeyExchange(dh.NewDHPair(keystore.GetDHKeyPair().KeyPair.GetPrivateKey(), shareKey.A_DH_PUK))
	sharedNhkb := dh.KeyExchange(dh.NewDHPair(keystore.GetDHKeyPair().KeyPair.GetPrivateKey(), shareKey.B_DH_PUK))
	err = sessionManager.AddRecvPipe(*message.Head.Sender, sk, sharedHka, sharedNhkb, shareKey.Idinfo.CPuk)
	if err != nil {
		return
	}

	//回复消息
	// data := nodeStore.NodeSelf.IdInfo.JSON()
	data, _ := nodeStore.NodeSelf.IdInfo.Proto()
	SendP2pReplyMsg(message, config.MSGID_security_create_pipe_recv, &data)

}

/*
	获取节点地址和身份公钥_返回
*/
func CreatePipe_recv(c engine.Controller, msg engine.Packet, message *Message) {
	if !message.CheckSendhash() {
		return
	}

	//fmt.Println("收到Hello消息", string(*message.Body.Content))

	// idinfo := nodeStore.Parse(*message.Body.Content)

	// flood.ResponseWait(config.CLASS_im_security_create_pipe, hex.EncodeToString(message.Body.Hash), message.Body.Content)
	flood.ResponseWait(config.CLASS_im_security_create_pipe, utils.Bytes2string(message.Body.Hash), message.Body.Content)
}

/*
	解密错误
*/
func Pipe_error(c engine.Controller, msg engine.Packet, message *Message) {
	if !message.CheckSendhash() {
		return
	}
	sessionManager.RemoveSendPipe(*message.Head.Sender)
}
