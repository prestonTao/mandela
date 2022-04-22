package rpc

import (
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/name"
	"mandela/chain_witness_vote/mining/tx_name_in"
	"mandela/chain_witness_vote/mining/tx_name_out"
	"mandela/config"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/rpc/model"
	"bytes"
	"encoding/hex"
	"net/http"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

/*
	域名注册，修改，续期
*/
func NameIn(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
	if ok {
		addrStr := addrItr.(string)
		if addrStr != "" {
			addrMul := crypto.AddressFromB58String(addrStr)
			addr = &addrMul
		}

		if addrStr != "" {
			dst := crypto.AddressFromB58String(addrStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(model.ContentIncorrectFormat, "address")
				return
			}
		}
	}

	amountItr, ok := rj.Get("amount") //转账金额
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := uint64(amountItr.(float64))
	if amount < config.Mining_name_deposit_min {
		res, err = model.Errcode(model.Nomarl, config.ERROR_name_deposit.Error())
		return
	}

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	//域名解析的节点地址参数
	ids := make([]nodeStore.AddressNet, 0)
	netIdsItr, ok := rj.Get("netids") //名称解析的网络id
	if ok {
		netIds := netIdsItr.([]interface{})
		for _, one := range netIds {
			netidOne := one.(string)
			idOne := nodeStore.AddressFromB58String(netidOne)
			ids = append(ids, idOne)
		}
	}

	//收款地址参数
	coins := make([]crypto.AddressCoin, 0)
	addrcoinsItr, ok := rj.Get("addrcoins") //名称解析的收款地址
	if ok {
		addrcoins := addrcoinsItr.([]interface{})
		for _, one := range addrcoins {
			addrcoinOne := one.(string)
			idOne := crypto.AddressFromB58String(addrcoinOne)
			coins = append(coins, idOne)
		}
	}

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	txpay, err := tx_name_in.NameIn(nil, addr, amount, gas, frozenHeight, pwd, comment, name, ids, coins)
	if err == nil {
		// res, err = model.Tojson("success")

		result, e := utils.ChangeMap(txpay)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(result)

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(model.NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())

	return
}

/*
	域名注销
*/
func NameOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// isReg := false //转账类型，true=注册名称；false=注销名称；

	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金退还地址
	if ok {
		addrStr := addrItr.(string)
		if addrStr != "" {
			addrMul := crypto.AddressFromB58String(addrStr)
			addr = &addrMul
		}

		if addrStr != "" {
			dst := crypto.AddressFromB58String(addrStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(model.ContentIncorrectFormat, "address")
				return
			}
		}
	}

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	name := nameItr.(string)
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if name == "" {
		res, err = model.Errcode(5002, "name")
		return
	}
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	txpay, err := tx_name_out.NameOut(nil, addr, 0, gas, frozenHeight, pwd, comment, name)
	if err == nil {
		// res, err = model.Tojson("success")

		result, e := utils.ChangeMap(txpay)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txpay.GetHash())

		res, err = model.Tojson(result)

		return
	}
	if err.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if err.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(model.NotEnough)
		return
	}
	if err.Error() == config.ERROR_name_not_exist.Error() {
		res, err = model.Errcode(model.NotExist)
		return
	}
	res, err = model.Errcode(model.Nomarl, err.Error())
	return

}

/*
	获取自己注册的域名列表
*/
func GetNames(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	nameinfoVOs := make([]NameinfoVO, 0)

	names := name.GetNameList()
	for _, one := range names {
		nets := make([]string, 0)
		for _, two := range one.NetIds {
			nets = append(nets, two.B58String())
		}
		addrs := make([]string, 0)
		for _, two := range one.AddrCoins {
			addrs = append(addrs, two.B58String())
		}
		voOne := NameinfoVO{
			Name:           one.Name,           //域名
			NetIds:         nets,               //节点地址
			AddrCoins:      addrs,              //钱包收款地址
			Height:         one.Height,         //注册区块高度，通过现有高度计算出有效时间
			NameOfValidity: one.NameOfValidity, //有效块数量
			Deposit:        one.Deposit,
		}
		nameinfoVOs = append(nameinfoVOs, voOne)
	}

	res, err = model.Tojson(nameinfoVOs)
	return
}

type NameinfoVO struct {
	Name           string   //域名
	NetIds         []string //节点地址
	AddrCoins      []string //钱包收款地址
	Height         uint64   //注册区块高度，通过现有高度计算出有效时间
	NameOfValidity uint64   //有效块数量
	Deposit        uint64   //冻结金额
}

/*
	查询域名
*/
func FindName(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	nameItr, ok := rj.Get("name") //注册的名称
	if !ok {
		res, err = model.Errcode(5002, "name")
		return
	}
	nameStr := nameItr.(string)
	nameinfo := name.FindNameToNet(nameStr)
	if nameinfo == nil || nameinfo.CheckIsOvertime(mining.GetHighestBlock()) {
		res, err = model.Errcode(model.NotExist, nameStr)
		return
	}
	bs, err := json.Marshal(nameinfo)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "find name formate failt 1")
		return
	}
	result := make(map[string]interface{})
	// err = json.Unmarshal(bs, &result)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&result)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, "find name formate failt 2")
		return
	}
	// result["DepositMin"] = store.DepositMin

	res, err = model.Tojson(result)
	return
}
