package token

import (
	"mandela/chain_witness_vote/db"
	"mandela/config"
	"bytes"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
	保存token详细信息
*/
func SaveTokenInfo(tokenid []byte, name, symbol string, supply uint64) error {
	tokeninfo := TokenInfo{
		Txid:   tokenid, //发布交易地址
		Name:   name,    //名称（全称）
		Symbol: symbol,  //单位
		Supply: supply,  //发行总量
	}
	bs, err := json.Marshal(tokeninfo)
	if err != nil {
		return err
	}
	return db.LevelTempDB.Save(BuildKeyForPublishToken(tokenid), &bs)
}

/*
	查询token信息信息
*/
func FindTokenInfo(tokenid []byte) (*TokenInfo, error) {
	bs, err := db.LevelTempDB.Find(BuildKeyForPublishToken(tokenid))
	if err != nil {
		return nil, err
	}
	tokeninfo := new(TokenInfo)
	buf := bytes.NewBuffer(*bs)
	decoder := json.NewDecoder(buf)
	decoder.UseNumber()
	err = decoder.Decode(tokeninfo)
	return tokeninfo, err
}

/*
	构建发布token的信息
*/
func BuildKeyForPublishToken(txid []byte) []byte {
	return append([]byte(config.TokenInfo), txid...) //[]byte(config.TokenInfo + txidStr)
}

/*
	token信息
*/
type TokenInfo struct {
	Txid   []byte //发布交易地址
	Name   string //名称（全称）
	Symbol string //单位
	Supply uint64 //发行总量
}
