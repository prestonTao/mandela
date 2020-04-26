package main

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/core/keystore"
	"mandela/core/utils/crypto"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"

	"golang.org/x/crypto/ed25519"
)

type TData struct {
	Status int    `json:"status"`
	Data   []Item `json:"data"`
}
type Item struct {
	Addr     string
	Txid     string
	Value    uint64
	OutIndex uint64
	VoteType uint16
	Height   uint64
}

//解析txitem
func ParseTxItems(jsonstr string) ([]*mining.TxItem, error) {
	tdata := TData{}
	err := json.Unmarshal([]byte(jsonstr), &tdata)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Printf("%+v", tdata)
	var txitem []*mining.TxItem
	for _, val := range tdata.Data {
		addr := crypto.AddressFromB58String(val.Addr)
		txid, _ := hex.DecodeString(val.Txid)
		it := &mining.TxItem{Height: val.Height, Addr: &addr, Txid: txid, Value: val.Value, OutIndex: val.OutIndex, VoteType: val.VoteType}
		txitem = append(txitem, it)
	}
	fmt.Printf("%+v", txitem[0])
	return txitem, nil
}

func main() {
	//pwd := "123456789"
	//初始化数据库
	path := filepath.Join("./wallet", "data")
	fmt.Println("path:", path)
	db.InitDB(path)
	//初始化keystore
	keystore.Load("conf/keystore.key")
	als := keystore.GetAddrAll()
	for _, val := range als {
		fmt.Println(val.B58String())
	}
	//item地址
	address := "1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW"
	addr := crypto.AddressFromB58String(address)
	//item
	itemstr := `{"data":[{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":49467275494,"Txid":"0000000000000000c0a525e4f772236a80abc775421302307aa0f1195b3bca3da6be061d5989238b","OutIndex":0,"Height":371,"VoteType":0,"State":0},{"Type":8,"Payload":"","Addr":"1PK8U46W2tJPegLtcBKQ3UYPXMF6xXptUW","Value":10000000000000,"Txid":"010000000000000073f4cca8525730c0f106180008f2d6a9790bfd7f127b415ddb5500e7a37b079e","OutIndex":0,"Height":1,"VoteType":0,"State":0}],"status":200}`
	items, err := ParseTxItems(itemstr)
	if err != nil {
		fmt.Println(err)
		return
	}
	//item地址公钥
	pubs := make(map[string]ed25519.PublicKey, 0)
	pub, ok := keystore.GetPukByAddr(addr)
	if !ok {
		fmt.Println("no addr:", address)
		return
	}
	pubs[address] = pub
	//收款地址
	addressto := "1LzRZ6j5WLrL2atL4Km8NXPfhWagHkcjCB"
	addrto := crypto.AddressFromB58String(addressto)
	// //构建转帐
	// rs, err := mining.CreateTxPayM(1000, items, pubs, &addrto, 100000000, 100, "")
	// fmt.Println(rs, err)
	// bs, err := rs.Json()
	// fmt.Println(string(*bs), err)

	//构建投票
	rs, err := mining.CreateTxVoteInM(1000, items, pubs, 2, addrto, "", 1000000000, 100, "")
	fmt.Println(rs, err)
	bs, err := rs.Json()
	fmt.Println(string(*bs), err)
}
