package message_center

import (
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/message_center/flood"
	"mandela/core/nodeStore"
	"mandela/core/utils/crypto/dh"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	fmt.Println("获取节点地址和身份公钥")

	//回复消息
	data := nodeStore.NodeSelf.IdInfo.JSON()
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

	idinfo := nodeStore.Parse(*message.Body.Content)
	sni := SearchNodeInfo{
		Id:      message.Head.Sender,
		SuperId: message.Head.SenderSuperId,
		CPuk:    idinfo.CPuk,
	}
	bs, _ := json.Marshal(sni)

	flood.ResponseWait(config.CLASS_security_searchAddr, hex.EncodeToString(message.Body.Hash), &bs)

}

type SearchNodeInfo struct {
	Id      *nodeStore.AddressNet //
	SuperId *nodeStore.AddressNet //
	CPuk    dh.Key                //
}

func ParserSearchNodeInfo(bs *[]byte) (*SearchNodeInfo, error) {
	sni := new(SearchNodeInfo)
	// err := json.Unmarshal(*bs, sni)
	decoder := json.NewDecoder(bytes.NewBuffer(*bs))
	decoder.UseNumber()
	err := decoder.Decode(sni)
	return sni, err
}

type ShareKey struct {
	// IV_DH_PUK dh.Key//向量
	Idinfo   nodeStore.IdInfo //身份密钥公钥
	A_DH_PUK dh.Key           //A公钥
	B_DH_PUK dh.Key           //B公钥
}

/*
	获取节点地址和身份公钥
*/
func CreatePipe(c engine.Controller, msg engine.Packet, message *Message) {

	if !message.CheckSendhash() {
		return
	}

	// fmt.Println("收到Hello消息", string(*message.Body.Content))

	shareKey := new(ShareKey)
	// err := json.Unmarshal(*message.Body.Content, shareKey)
	decoder := json.NewDecoder(bytes.NewBuffer(*message.Body.Content))
	decoder.UseNumber()
	err := decoder.Decode(shareKey)
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
	data := nodeStore.NodeSelf.IdInfo.JSON()
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

	flood.ResponseWait(config.CLASS_im_security_create_pipe, hex.EncodeToString(message.Body.Hash), message.Body.Content)
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
