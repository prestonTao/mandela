package main

import (
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/token/payment"
	"mandela/config"
	"mandela/core/utils/crypto"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	start()
}

func start() {
	txJsonItrs := make([]interface{}, 0)
	err := json.Unmarshal([]byte(txs), &txJsonItrs)
	if err != nil {
		fmt.Println("解析交易字符串错误")
		panic(err.Error())
	}

	txItrs := make([]mining.TxItr, 0)
	for i, one := range txJsonItrs {
		bs, err := json.Marshal(one)
		if err != nil {
			fmt.Println("序列化单个交易错误", i)
			panic(err.Error())
		}
		txItr := parseTx(bs)
		if txItr.Class() == config.Wallet_tx_type_pay {
			// createSign(txItr)
			printVO(txItr)
			txItrs = append(txItrs, txItr)
		} else if txItr.Class() == config.Wallet_tx_type_token_payment {
			txItr := parseTxTokenPay(bs)
			// reSign(txItr, seed, pwd)
			printVO(txItr)
			txItrs = append(txItrs, txItr)
		} else {
			panic("未识别的交易类型")
		}
	}
	fmt.Println("共有交易", len(txItrs))

	//开始上传交易
	peer1 := Peer{
		// Addr: "47.114.76.113:2080", //节点地址及端口，杭州
		// Addr: "47.106.221.172:2080", //节点地址及端口，深圳
		// Addr: "39.108.82.129:2080", //节点地址及端口，区块链1
		// Addr: "47.108.14.67:2080", //节点地址及端口，成都
		// Addr:       "47.108.210.229:2080",   //节点地址及端口，本地
		Addr:       "47.108.50.1:2080", //节点地址及端口，本地
		AddressMax: 50000,              //收款地址总数
		RPCUser:    "test",             //rpc用户名
		RPCPwd:     "testp",            //rpc密码
		WalletPwd:  "123456789",        //
		// PayChan:    make(chan *Address, 10), //
	}
	fmt.Println(peer1)

	// for _, one := range txItrs {
	// 	txBs, _ := one.Json()
	// 	// fmt.Println(strconv.Quote(string(*txBs)))
	// 	peer1.PushTx(strconv.Quote(string(*txBs)))
	// }

}

type Peer struct {
	Addr       string //节点地址及端口
	AddressMax uint64 //收款地址总数
	RPCUser    string //rpc用户名
	RPCPwd     string //rpc密码
	WalletPwd  string //钱包支付密码
	// PayChan    chan *Address //
}

/*
	上传交易
*/
func (this *Peer) PushTx(txString string) bool {
	fmt.Println("上传交易")
	fmt.Println(txString)
	//{"method":"getnewaddress","params":{"password":"123456789"}}
	paramsChild := map[string]interface{}{
		"tx": txString,
	}
	params := map[string]interface{}{
		"method": "pushtx",
		"params": paramsChild,
	}
	result := Post(this.Addr, this.RPCUser, this.RPCPwd, params)
	_, err := json.Marshal(result.Result)
	if err != nil {
		fmt.Println("序列化错误", err.Error())
		return false
	}
	return true
}

func parseTx(bs []byte) *mining.Tx_Pay {
	txbaseVO := new(mining.TxBaseVO)

	err := json.Unmarshal(bs, txbaseVO)
	panicError(err)

	txhash, err := hex.DecodeString(txbaseVO.Hash)
	panicError(err)

	vins := make([]*mining.Vin, 0)
	for _, one := range txbaseVO.Vin {
		txid, err := hex.DecodeString(one.Txid)
		panicError(err)
		puk, err := hex.DecodeString(one.Puk)
		panicError(err)
		vin := mining.Vin{
			Txid: txid,     //UTXO 前一个交易的id
			Vout: one.Vout, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
			Puk:  puk,      //公钥
		}
		vins = append(vins, &vin)
	}

	vouts := make([]*mining.Vout, 0)
	for _, one := range txbaseVO.Vout {

		vout := mining.Vout{
			Value:        one.Value,                                //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressFromB58String(one.Address), //钱包地址
			FrozenHeight: one.FrozenHeight,                         //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
		}
		vouts = append(vouts, &vout)
	}

	blockHash, err := hex.DecodeString(txbaseVO.BlockHash)
	panicError(err)

	txbase := mining.TxBase{
		Hash:       txhash,              //本交易hash，不参与区块hash，只用来保存
		Type:       txbaseVO.Type,       //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  txbaseVO.Vin_total,  //输入交易数量
		Vin:        vins,                //交易输入
		Vout_total: txbaseVO.Vout_total, //输出交易数量
		Vout:       vouts,               //交易输出
		Gas:        txbaseVO.Gas,        //交易手续费，此字段不参与交易hash
		LockHeight: txbaseVO.LockHeight, //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		// LockHeight: lockHeight,               //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		Payload:   []byte(txbaseVO.Payload), //备注信息
		BlockHash: blockHash,                //本交易属于的区块hash，不参与区块hash，只用来保存
	}
	txPay := mining.Tx_Pay{
		TxBase: txbase,
	}
	return &txPay
}

/*
	解析token交易
*/
func parseTxTokenPay(bs []byte) *payment.TxToken {
	txbaseVO := new(payment.TxToken_VO)

	err := json.Unmarshal(bs, txbaseVO)
	panicError(err)

	txhash, err := hex.DecodeString(txbaseVO.Hash)
	panicError(err)

	vins := make([]*mining.Vin, 0)
	for _, one := range txbaseVO.Vin {
		txid, err := hex.DecodeString(one.Txid)
		panicError(err)
		puk, err := hex.DecodeString(one.Puk)
		panicError(err)
		vin := mining.Vin{
			Txid: txid,     //UTXO 前一个交易的id
			Vout: one.Vout, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
			Puk:  puk,      //公钥
		}
		vins = append(vins, &vin)
	}

	vouts := make([]*mining.Vout, 0)
	for _, one := range txbaseVO.Vout {

		vout := mining.Vout{
			Value:        one.Value,                                //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressFromB58String(one.Address), //钱包地址
			FrozenHeight: one.FrozenHeight,                         //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
		}
		vouts = append(vouts, &vout)
	}

	blockHash, err := hex.DecodeString(txbaseVO.BlockHash)
	panicError(err)

	txbase := mining.TxBase{
		Hash:       txhash,              //本交易hash，不参与区块hash，只用来保存
		Type:       txbaseVO.Type,       //交易类型，默认0=挖矿所得，没有输入;1=普通转账到地址交易
		Vin_total:  txbaseVO.Vin_total,  //输入交易数量
		Vin:        vins,                //交易输入
		Vout_total: txbaseVO.Vout_total, //输出交易数量
		Vout:       vouts,               //交易输出
		Gas:        txbaseVO.Gas,        //交易手续费，此字段不参与交易hash
		LockHeight: txbaseVO.LockHeight, //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		// LockHeight: lockHeight,               //本交易锁定在小于等于这个高度的块中，超过这个高度，块将不被打包到区块中。
		Payload:   []byte(txbaseVO.Payload), //备注信息
		BlockHash: blockHash,                //本交易属于的区块hash，不参与区块hash，只用来保存
	}

	tokenVin := make([]*mining.Vin, 0)
	for _, one := range txbaseVO.Token_Vin {
		txid, err := hex.DecodeString(one.Txid)
		panicError(err)
		puk, err := hex.DecodeString(one.Puk)
		panicError(err)
		vin := mining.Vin{
			Txid: txid,     //UTXO 前一个交易的id
			Vout: one.Vout, //一个输出索引（vout），用于标识来自该交易的哪个UTXO被引用（从零开始）
			Puk:  puk,      //公钥
		}
		tokenVin = append(tokenVin, &vin)
	}

	tokenVouts := make([]*mining.Vout, 0)
	for _, one := range txbaseVO.Token_Vout {

		vout := mining.Vout{
			Value:        one.Value,                                //输出金额 = 实际金额 * 100000000
			Address:      crypto.AddressFromB58String(one.Address), //钱包地址
			FrozenHeight: one.FrozenHeight,                         //冻结高度。小于等于这个冻结高度，未花费的交易余额不能使用
		}
		tokenVouts = append(tokenVouts, &vout)
	}

	txPay := payment.TxToken{
		TxBase:           txbase,
		Token_Vin_total:  txbaseVO.Token_Vin_total,  //输入交易数量
		Token_Vin:        tokenVin,                  //交易输入
		Token_Vout_total: txbaseVO.Token_Vout_total, //输出交易数量
		Token_Vout:       tokenVouts,                //交易输出
	}
	return &txPay
}

func panicError(err error) {
	if err != nil {
		panic(err)
	}
}

func printVO(txItr mining.TxItr) {
	txPayBs, err := json.Marshal(txItr.GetVOJSON())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(string(txPayBs))

	fmt.Println("打印可以上链的交易:")

	txBs, _ := txItr.Json()
	fmt.Println(strconv.Quote(string(*txBs)))

	fmt.Println("end!!!!!!")
}

/*
	自定义请求head，body，method，参数用body传递
	获取添加金币记录
*/
func Post(addr, rpcUser, rpcPwd string, params map[string]interface{}) *PostResult {
	url := "/rpc"
	method := "POST"

	header := http.Header{"user": []string{rpcUser}, "password": []string{rpcPwd}}
	client := &http.Client{}
	//req, err := http.NewRequest("GET", "http://www.baidu.com/", nil)
	bs, err := json.Marshal(params)
	req, err := http.NewRequest(method, "http://"+addr+url, strings.NewReader(string(bs)))
	if err != nil {
		fmt.Println("创建request错误")
		return nil
	}
	req.Header = header
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("请求服务器错误", err.Error())
		return nil
	}
	// fmt.Println("response:", resp.StatusCode)

	var resultBs []byte

	if resp.StatusCode == 200 {
		resultBs, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("读取body内容错误")
			return nil
		}
		// fmt.Println(string(resultBs))

		result := new(PostResult)

		buf := bytes.NewBuffer(resultBs)
		decoder := json.NewDecoder(buf)
		decoder.UseNumber()
		err = decoder.Decode(result)
		return result
	}
	fmt.Println("StatusCode:", resp.StatusCode)
	return nil
}

type PostResult struct {
	Jsonrpc string      `json:"jsonrpc"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Result  interface{} `json:"result"`
}

var txs = `[
			{
				"hash": "040000000000000049031a8302d565335a1ad5097c46a9d630b61e763f7177c27082ae493486dc46",
				"type": 4,
				"vin_total": 1,
				"vin": [
					{
						"txid": "04000000000000004ea4400dbcd9cee48275e275b7e9cc11af002ca002e9f0334f20081219536c2d",
						"vout": 1,
						"puk": "66363e8d9e661c91fe2532fa084d6cd2f4f3036e313879dc9572bb36875c69c8",
						"sign": "35ed331853a47af193922fd7729d005dd1fa78f37b9838e9a149b0e83c97908898f44f1a2539508f5e1761771d105d94cd7246086fbb49104ca79c4a6016e80a"
					}
				],
				"vout_total": 2,
				"vout": [
					{
						"value": 4363333,
						"address": "ZHC8vBoP67qjgy4ZCUzE3AQukTiERQtPJqVa4",
						"frozen_height": 1615140000
					},
					{
						"value": 99987653334,
						"address": "ZHCF1Pzr32uRs8T6gCEdEF1WYfxL11zgpRJB4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1027040,
				"payload": "{\"Type\":50001,\"Payloads\":\"eyJpZCI6MTM1NjE0NDE1NjA2NjI4NzYxNywidHlwZSI6MjAwMDMsImZyb20iOiJaSENEcjRlZ1pTNXB2THJKSktRZlhNYmtYQ1ltQUwxeEplMTY0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50IjoxMTkwMDAwMDAwMCwicGF5bG9hZCI6WyIxMjgzOTU0NTI0NzQ3MDEwMDQ5IiwiMTMwOTAwMDAwMCIsIiIsIiJdLCJjcmVhdGVkIjoxNjEyNTg2MDAxLCJ2b3V0cyI6WyIzMDAxMSIsIjMwMDAzIiwiMzAwMDUiLCIzMDAwNiIsIjMwMDAwIiwiMjAwMDAiLCIyMDAwNCIsIjMwMDEwIiwiMzAwMDgiXX0=\",\"Rand\":0,\"Puk\":\"2ZVnSpLmH6EOkE8M0I5cpA2z65/hxWX03ceR4AbGMqc=\",\"Sign\":\"ybSHR1bT446A1cwM+VK+u87vUSxUaNx3FrYRk/GqZEaLfFCa9VhdBWRWpFuWYhPeXv8Aby10VfCvER6L5tEhDw==\"}",
				"blockhash": ""
			},
			{
				"hash": "0b000000000000007e375aaeaf0e3c98fc9097098fdee951712cdcb08235ab4f1b322d1ff318f631",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b00000000000000e952ef8b46891530bdf8a81fcc7b98462cc4973abdb531267b1e1d54785578c4",
						"vout": 0,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": "13f1f4489026e15d4507602dbe0605cfc19c9fe459da38d655179c54d0fb502910b5473be13d97ccbf6fee055e056be6fc2a41247dd781b1c30318d42eb6f700"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9999990,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1042146,
				"payload": "{\"Type\":50010,\"Payloads\":\"eyJpZCI6MTM2MTA4MDQwMzUxNTc5MzQwOSwidHlwZSI6NTAwMTAsImZyb20iOiJaSEM2aUhGb0VoclFlenNCTlVRcW1adnNtNnBOWVJIQkRVc0Q0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50IjoyOTkwMDAwMDAwMCwicGF5bG9hZCI6WyIxMjgzOTU0NTI0NzQ3MDEwMDQ5IiwiMTAxNjYwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxMzM0MTg1Nn0=\",\"Rand\":0,\"Puk\":\"INRWx2953g8tBBCSwqTeU3W1avyUis4MIFC/c/Oc4+k=\",\"Sign\":\"8YqsQwhiXcH2jgobsvueuqHeXRxpQgfIO0YajtdVJJyn2T72pKc7OEHvro9MPUQzLv0yg/S/zYfw+6yQfDiyAg==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b00000000000000899bf1513b758c19bef1bdd4f5a6bff89124f0d5394c5dacbb7f2ea5ca88ddb2",
						"vout": 38999,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 29900000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 299970100000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b00000000000000d7c756c11815e9eb79d71e6caaf9b1515c36254f51a1bb3bb60c2a9f3413f969",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b00000000000000c5c1eba20a9ede78b45dd9210e87e36e8e10411fc41b2834be4c9f6f6d4e4660",
						"vout": 0,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": "966ce6977fc4293aed6ed857968f94cc5761c552e1676d5ed121553c30b91d2e68b444776daa227d8afaec624f856f89550fc9928a6feff9907435fbf5d34b06"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9999990,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1047778,
				"payload": "{\"Type\":50010,\"Payloads\":\"eyJpZCI6MTM2MTMxNjY2MzU1MTQ1MTEzNywidHlwZSI6NTAwMTAsImZyb20iOiJaSENHa3hyOEZSc0w4RTNDdEFTRTJKblJGRGJIWFRRd2hhRVY0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50Ijo3ODAwMDAwMDAsInBheWxvYWQiOlsiMTI4Mzk1NDUyNDc0NzAxMDA0OSIsIjc4MDAwMDAwIiwiIiwiIl0sImNyZWF0ZWQiOjE2MTMzOTgxNzR9\",\"Rand\":0,\"Puk\":\"INRWx2953g8tBBCSwqTeU3W1avyUis4MIFC/c/Oc4+k=\",\"Sign\":\"WSt1TZM5ye4IFqeGDQRSUVoXAq4KfxKdB8LrNPxCFuODnfcPRyZKNVewjEb1pcbmE54uAe/osKj6oZQ8oWCoDg==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b0000000000000013e94233a8366967f81a582f8dea4d496c4637b7f21e364e63af79a191f27f1e",
						"vout": 2,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 780000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 299999120000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b00000000000000eeea182c111245e04675971871ec766c3061afe33c0becb5caadc5eee6dfaf75",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b00000000000000d3ad76f707e66be1e8ee2388b066a2a7333a5f9870775cf8b789d46c07fbe368",
						"vout": 0,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": "f336b16071f90084be3a765da141ba125a899671d8de75202ef8cf2ec6c1b2598e172d913cea451d560f4af65ea7abf773bfc771eb4bcd49fad3ad424ee5c60e"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9999990,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 883300,
				"payload": "{\"Type\":50010,\"Payloads\":\"eyJpZCI6MTM1ODI2MTQwODU0Njc0MjI3NCwidHlwZSI6NTAwMTAsImZyb20iOiJaSENNS3BoUEpQcHNTUGtONFlNQThpd1BZVDdSbXg3Nm91YWU0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50Ijo5OTgwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCI2OTkwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxMjY2OTgyNn0=\",\"Rand\":0,\"Puk\":\"INRWx2953g8tBBCSwqTeU3W1avyUis4MIFC/c/Oc4+k=\",\"Sign\":\"BFHNuNRIm6T5PoUOfZ+Q7mrX++qxW+ZYjCiQFXA7MrWHvAufiwco7heVRgsPKZwf+KinkmBEiy8F7FRAXsE3DA==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b00000000000000899bf1513b758c19bef1bdd4f5a6bff89124f0d5394c5dacbb7f2ea5ca88ddb2",
						"vout": 30440,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 9980000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 299990020000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "040000000000000021d76a287f35af2bde993de6a98fc4a67762c04595efb949088f085a53a2040c",
				"type": 4,
				"vin_total": 1,
				"vin": [
					{
						"txid": "04000000000000002b48d353a97552ecc943d7391ad1decb366101c00c5abb88cf318507d9a5beac",
						"vout": 1,
						"puk": "66363e8d9e661c91fe2532fa084d6cd2f4f3036e313879dc9572bb36875c69c8",
						"sign": "3baeb6fb5b70c58e15229853adf2dce18a943bbc488e22f4fb801802409b0779f6b1569db6628695f4a05567e0c87a5043a3a67bd5cdccb9ca04a1789f470107"
					}
				],
				"vout_total": 2,
				"vout": [
					{
						"value": 513333,
						"address": "ZHC8vBoP67qjgy4ZCUzE3AQukTiERQtPJqVa4",
						"frozen_height": 1617818400
					},
					{
						"value": 99979886667,
						"address": "ZHCF1Pzr32uRs8T6gCEdEF1WYfxL11zgpRJB4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1242933,
				"payload": "{\"Type\":50001,\"Payloads\":\"eyJpZCI6MTM1MjEyMjg5ODk0MTA1NDk3OCwidHlwZSI6MjAwMDMsImZyb20iOiJaSEMzUkZHRnlSWXN5YUpMekY4Sk1DTkJpbnM0cEtNd2l1MVM0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50IjozMDgwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCIxNTQwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxNTM0OTc2MSwidm91dHMiOlsiMjAwMDAiLCIyMDAwNCIsIjMwMDExIiwiMzAwMTAiLCIzMDAwOCIsIjMwMDA1IiwiMzAwMDAiLCIzMDAwMyIsIjMwMDA2Il19\",\"Rand\":0,\"Puk\":\"2ZVnSpLmH6EOkE8M0I5cpA2z65/hxWX03ceR4AbGMqc=\",\"Sign\":\"tH1LK1rVKylixmnPMQ15UvRoPR4WB9moqRugCIj6szn0/Vzd4TT5w7lfpW9/SploXm9NnnsrFQf55a7Aw+7LBA==\"}",
				"blockhash": ""
			},
			{
				"hash": "04000000000000002896316a740412a19832667c17279273458e5ca31c62ecbff9eaef34a03ac608",
				"type": 4,
				"vin_total": 1,
				"vin": [
					{
						"txid": "04000000000000004b643dcf7cd16f66d38bccb8a9b4935fd3dd9597af39cf2e001031c7a42af1bf",
						"vout": 1,
						"puk": "66363e8d9e661c91fe2532fa084d6cd2f4f3036e313879dc9572bb36875c69c8",
						"sign": "914500188ff1cd1784b0bfd502aeea24718c2c1ebc53ca435b72357b5350765ebe80153c5342fd8c34efeb0f28f557415a423aa83d911e56a433543d19638c09"
					}
				],
				"vout_total": 2,
				"vout": [
					{
						"value": 3993333,
						"address": "ZHC8vBoP67qjgy4ZCUzE3AQukTiERQtPJqVa4",
						"frozen_height": 1617818400
					},
					{
						"value": 99988206667,
						"address": "ZHCF1Pzr32uRs8T6gCEdEF1WYfxL11zgpRJB4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1246455,
				"payload": "{\"Type\":50001,\"Payloads\":\"eyJpZCI6MTM2ODE3NjIwNDk0MzU2ODg5NywidHlwZSI6MjAwMDMsImZyb20iOiJaSEM2ZHM3MWFCMlhUMm1MQ0p2UFkzUTNQWFF5d2tGcHRrRkE0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50Ijo0OTkwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCIxMTk4MDAwMDAwIiwiIiwiIl0sImNyZWF0ZWQiOjE2MTUzODQ5ODMsInZvdXRzIjpbIjMwMDAwIiwiMzAwMTAiLCIzMDAwMyIsIjMwMDA4IiwiMzAwMDUiLCIzMDAwNiIsIjIwMDAwIiwiMjAwMDQiLCIzMDAxMSJdfQ==\",\"Rand\":0,\"Puk\":\"2ZVnSpLmH6EOkE8M0I5cpA2z65/hxWX03ceR4AbGMqc=\",\"Sign\":\"IODmIuT2jEsDB8I8oUeQ4SnGvN5mMNdiztYzn/TJKapd+Mevf4WHIZ/suvnKKQ707TJy3AtDKVnwHd6uP6KJAg==\"}",
				"blockhash": ""
			},
			{
				"hash": "0400000000000000a0017fb62dffd38054e404079f93e391d976d1eff0b75e5ae0763cf587f21341",
				"type": 4,
				"vin_total": 1,
				"vin": [
					{
						"txid": "04000000000000002ac52de0b8ba92252b4a0b63d6a0e75ec024cb93ecd45b9cee346feb731b0934",
						"vout": 1,
						"puk": "66363e8d9e661c91fe2532fa084d6cd2f4f3036e313879dc9572bb36875c69c8",
						"sign": "347b21ac8ebbd497a9b918804e7e02d7b5310d917aec5863809d3348d0f7df550e3257a19a45943a611b0ac1b95527dc3152a49c956e8b099b98b87bfced0a0c"
					}
				],
				"vout_total": 2,
				"vout": [
					{
						"value": 743333,
						"address": "ZHC8vBoP67qjgy4ZCUzE3AQukTiERQtPJqVa4",
						"frozen_height": 1617818400
					},
					{
						"value": 99994710001,
						"address": "ZHCF1Pzr32uRs8T6gCEdEF1WYfxL11zgpRJB4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1311540,
				"payload": "{\"Type\":50001,\"Payloads\":\"eyJpZCI6MTM2OTI5NzE5MTMxNTU1MDIxMCwidHlwZSI6MjAwMDMsImZyb20iOiJaSENBZUFqY3EzUFo0ckc1akwzYjZNNlF4Zmp1SnR3TFVWSjY0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50Ijo0NDUwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCIyMjMwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxNjAzNTgzNCwidm91dHMiOlsiMzAwMDgiLCIzMDAwMyIsIjMwMDA1IiwiMzAwMDAiLCIyMDAwNCIsIjMwMDExIiwiMzAwMTAiLCIzMDAwNiIsIjIwMDAwIl19\",\"Rand\":0,\"Puk\":\"2ZVnSpLmH6EOkE8M0I5cpA2z65/hxWX03ceR4AbGMqc=\",\"Sign\":\"N9ds6sisf8xA5rUWPNIMyICCV+SrlUuelWEsI5jTyQzOdtvnAggvX327FeBrCivoZT92smfvXE5yF+6wRDgCBQ==\"}",
				"blockhash": ""
			},
			{
				"hash": "0b0000000000000070cfaa958f371568f8d7d0a939bfe568ff0e7af691b64f90518843a0ca3dd55f",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "04000000000000008feeee68e489ff9f9df39eb9b7132c6b8d1b2e2ca579a9766ba561966ec4f1ea",
						"vout": 0,
						"puk": "116d59fe87cb6e1f2f45857af06390f0d4102f7fb8d6ca9c7f54822c1fb3e373",
						"sign": "d85958760b5b419033141dfc34bec89582638c25346e40cd1aca4f712f7d107efa81c4dd962074bf596dafe96ced76767afbaa19d94535fdced634c77b2eba0a"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 1,
						"address": "ZHCPRe83DJ4MaTyBAeFH4UWESg5V8j2U77M74",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1311834,
				"payload": "{\"Type\":10002,\"Payloads\":\"eyJpZCI6MTM3MjM5MjIwMzk0NzQ0MjE3NywidHlwZSI6MTAwMDIsImZyb20iOiJaSENQUmU4M0RKNE1hVHlCQWVGSDRVV0VTZzVWOGoyVTc3TTc0IiwidG8iOiJaSENNbXVWVWtqQXdwTXo0ZmQxZ0VWNTVwRVhYakdLMWlQQzQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50IjoxNTAwMDAwMDAwMCwiY3JlYXRlZCI6MTYxNjAzODc3NH0=\",\"Rand\":0,\"Puk\":\"EW1Z/ofLbh8vRYV68GOQ8NQQL3+41sqcf1SCLB+z43M=\",\"Sign\":\"Ei4T3WKi9A249d97Mye+XBiZyXjPVuYhVJIprGPKGOo5ivDLo1xKCpRSoZXiACTFTQM9Yn5sPcWeU2tezKDUAA==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b00000000000000ee783a19e0d273598eaf1f06c4c8919077c1f3df603f40e5d38720cc79769f43",
						"vout": 0,
						"puk": "116d59fe87cb6e1f2f45857af06390f0d4102f7fb8d6ca9c7f54822c1fb3e373",
						"sign": ""
					}
				],
				"token_vout_total": 1,
				"token_vout": [
					{
						"value": 15000000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b00000000000000378a9f629f8c37fdde49dd93f034cd98bb80fb2467f0891c4a4e7f82dea6144f",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b000000000000008d81c9d36e2a069075f47101c92039260fb75b262719692286c9f566968ac2fc",
						"vout": 0,
						"puk": "d995674a92e61fa10e904f0cd08e5ca40db3eb9fe1c565f4ddc791e006c632a7",
						"sign": "ce1f6a3d0157a90b8061d39d69c3d4f34f7e5206e885abbb7b62e00ba58cf3360ddce6ec4fb2cbeeb52e1069e1b6940bb0068116fdb77b53f175dc595f691409"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 2000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1312762,
				"payload": "{\"Type\":50001,\"Payloads\":\"eyJpZCI6MTM3MjQzMDE1MTg3OTMxMTM2MiwidHlwZSI6NTAwMDEsImZyb20iOiJaSENNQ0xLVUZyc21WRDEzOUg0UFliRWkzWGp4VWlHWUZScVo0IiwidG8iOiJaSEM0dlpySmRDcEQyQjNzNTRZRkNEMTVkeTRIZXdhaXBEM0w0IiwiY29udHJhY3QiOiIwYTAwMDAwMDAwMDAwMDAwZmQ1ODk2NjJiYWI3MjRiYWE5NzE4NTQ1MjFkYzlkZmY0YmMyZTZiOTM2MmZhNzg3NTlmODhkZjM2ZGUwNjY1NyIsImFtb3VudCI6Mjk5MDAwMDAwMDAsInBheWxvYWQiOlsiMTM1NTM2MDU4ODQwNjc0MzA0MiIsIjEwMTY2MDAwMDAwIiwiIiwiIl0sImNyZWF0ZWQiOjE2MTYwNDgwNTUsInZvdXRzIjpbIjMwMDA2IiwiMzAwMDAiLCIyMDAwMCIsIjIwMDA0IiwiMzAwMDMiLCIzMDAxMSIsIjMwMDEwIiwiMzAwMDgiLCIzMDAwNSJdfQ==\",\"Rand\":0,\"Puk\":\"2ZVnSpLmH6EOkE8M0I5cpA2z65/hxWX03ceR4AbGMqc=\",\"Sign\":\"IPAeDtwrAH4lQ4m7VdhH19CujJQGYM4sz1+0TluZ48bB9JrNy37SqLzBaE/FUKWjGUCLm8b1HlRC1ygf2EciAA==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b000000000000003e14cd36461cc6c523330ffa28647132b33ca4264d67b81f4bc034683f59ff58",
						"vout": 0,
						"puk": "d995674a92e61fa10e904f0cd08e5ca40db3eb9fe1c565f4ddc791e006c632a7",
						"sign": ""
					}
				],
				"token_vout_total": 10,
				"token_vout": [
					{
						"value": 508300000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 1617991200
					},
					{
						"value": 19734000000,
						"address": "ZHC4vZrJdCpD2B3s54YFCD15dy4HewaipD3L4",
						"frozen_height": 0
					},
					{
						"value": 2033200000,
						"address": "ZHCMCLKUFrsmVD139H4PYbEi3XjxUiGYFRqZ4",
						"frozen_height": 0
					},
					{
						"value": 813280000,
						"address": "ZHCMCLKUFrsmVD139H4PYbEi3XjxUiGYFRqZ4",
						"frozen_height": 0
					},
					{
						"value": 609960000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 1617991200
					},
					{
						"value": 1219920000,
						"address": "ZHCPmPhKgkgCkqeUstavLcfZGKe9SafrcbzG4",
						"frozen_height": 0
					},
					{
						"value": 2033200000,
						"address": "ZHCCgUBcVDUjFafsdEUmhs4MtAtynbAfJewj4",
						"frozen_height": 0
					},
					{
						"value": 1016600000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 0
					},
					{
						"value": 813280000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 1617991200
					},
					{
						"value": 1118260000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b000000000000009950c7069aed11ef6e5c9387ca857c7ed7272211f9346c4664ef05fc1c7a95cf",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "040000000000000067097df77c95b1501d7785eb8f4e45489e1a22d5336f726c3c374af66eeea7b5",
						"vout": 0,
						"puk": "0d1f0d7787185bfb10ecbc030b4e8a8a9a0c1779256d3f0f69fb91898ca37fa7",
						"sign": "48930d99d55572ea4119bed24dae910a87a9b8290f8239e3f69d6000f4c8d4fcd56b3bbb579503b9cf4a0c764865b66bf01512be2d60feed488a02396dcf3c03"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9800000,
						"address": "ZHCAHQRCCK7ajAEAemKDvvgNAcKWQwGCMyRa4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1313184,
				"payload": "{\"Type\":50000,\"Payloads\":\"W3siaWQiOjEzNzI0NDg3NTk3Mjg1NzQ0NjUsInR5cGUiOjUwMDAwLCJmcm9tIjoiWkhDQUhRUkNDSzdhakFFQWVtS0R2dmdOQWNLV1F3R0NNeVJhNCIsInRvIjoiWkhDTDRENzlGWjFlb0tCMW52SGg5MTV1b0s4amhpbUhiSE40IiwiY29udHJhY3QiOiIwYTAwMDAwMDAwMDAwMDAwZmQ1ODk2NjJiYWI3MjRiYWE5NzE4NTQ1MjFkYzlkZmY0YmMyZTZiOTM2MmZhNzg3NTlmODhkZjM2ZGUwNjY1NyIsImFtb3VudCI6MTAwMDAwMDAwMDAsInBheWxvYWQiOlsiMTI4Mzk1NDUyNDc0NzAxMDA0OSIsIjUwMDAwMDAwMCIsIiIsIiJdLCJjcmVhdGVkIjoxNjE2MDUyMjU4fV0=\",\"Rand\":0,\"Puk\":\"DR8Nd4cYW/sQ7LwDC06KipoMF3klbT8PafuRiYyjf6c=\",\"Sign\":\"7VcvCqDSBQg26cX11ocwP98pQOp5RqzFkyy0p3A/oqiJAyvsyWP9T4kVLTwaJQMUFNXrW+XGbwmvtFF8T0paDg==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 5,
				"token_vin": [
					{
						"txid": "0b000000000000006f746640d2db4e898e0db476cb5b7a8ccbe001c526958c876a07339439216948",
						"vout": 3,
						"puk": "0d1f0d7787185bfb10ecbc030b4e8a8a9a0c1779256d3f0f69fb91898ca37fa7",
						"sign": ""
					},
					{
						"txid": "0b00000000000000fa3024ee220096640454ad161c773b2336c6c714cb3cc2266a51dc83750f5a99",
						"vout": 0,
						"puk": "0d1f0d7787185bfb10ecbc030b4e8a8a9a0c1779256d3f0f69fb91898ca37fa7",
						"sign": ""
					},
					{
						"txid": "0b000000000000001fe89e9902160149a6c8689301f4ce865e8f3d323e7755ff2bbf34510081a736",
						"vout": 1,
						"puk": "0d1f0d7787185bfb10ecbc030b4e8a8a9a0c1779256d3f0f69fb91898ca37fa7",
						"sign": ""
					},
					{
						"txid": "0b000000000000001fe89e9902160149a6c8689301f4ce865e8f3d323e7755ff2bbf34510081a736",
						"vout": 7,
						"puk": "0d1f0d7787185bfb10ecbc030b4e8a8a9a0c1779256d3f0f69fb91898ca37fa7",
						"sign": ""
					},
					{
						"txid": "0b000000000000004d0ba3eedbc84cb423f82428f3ba3de7aaaf954da79ad56c830c4238cc7d0679",
						"vout": 0,
						"puk": "0d1f0d7787185bfb10ecbc030b4e8a8a9a0c1779256d3f0f69fb91898ca37fa7",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 10000000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 1309240000,
						"address": "ZHCAHQRCCK7ajAEAemKDvvgNAcKWQwGCMyRa4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b00000000000000fccb908badec192b91986e264d009a130cc7a735a494532a574bd2455d6a8784",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b00000000000000a0524bd77345b6769f56b900b525f2cd8ed14f9f35d9e7613935b72f5dbdbbc4",
						"vout": 0,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": "866ae7b78881e808fac5cac84c0f79824131854d1a1d970785d80c90bf37249a78ee2c12ca7eb0c65a776501fb7233c18de2e3fed7aaab6d8af03177686dc20a"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9999990,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1314336,
				"payload": "{\"Type\":50010,\"Payloads\":\"eyJpZCI6MTM3MjQ5NzE2NjY2NDk5NDgxOCwidHlwZSI6NTAwMTAsImZyb20iOiJaSENZV2plWW42eGVCdW1BTHlVcXlLZDdlN1RvRFpRNTFQSzQiLCJ0byI6IlpIQ0w0RDc5RloxZW9LQjFudkhoOTE1dW9LOGpoaW1IYkhONCIsImNvbnRyYWN0IjoiMGEwMDAwMDAwMDAwMDAwMGZkNTg5NjYyYmFiNzI0YmFhOTcxODU0NTIxZGM5ZGZmNGJjMmU2YjkzNjJmYTc4NzU5Zjg4ZGYzNmRlMDY2NTciLCJhbW91bnQiOjM5OTAwMDAwMDAsInBheWxvYWQiOlsiMTI4Mzk1NDUyNDc0NzAxMDA0OSIsIjkxODAwMDAwMCIsIiIsIiJdLCJjcmVhdGVkIjoxNjE2MDYzNzk5fQ==\",\"Rand\":0,\"Puk\":\"INRWx2953g8tBBCSwqTeU3W1avyUis4MIFC/c/Oc4+k=\",\"Sign\":\"niSdaanB5eUoRAHDFePh782L5296sCjSxYsVSw6EP+0snR3rpauFPj1rrRmKy4ar8fxro4VCurur0I71VIVMBg==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b00000000000000899bf1513b758c19bef1bdd4f5a6bff89124f0d5394c5dacbb7f2ea5ca88ddb2",
						"vout": 18440,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 3990000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 299996010000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0400000000000000eef60f55da3e1f1abb19966808dc79cd0dc60ad3cdedb7fa838078ccf04edde8",
				"type": 4,
				"vin_total": 1,
				"vin": [
					{
						"txid": "04000000000000002ac9da3bc9419e122d04a779066a5d25f2759bdb266e251578720f3457f92875",
						"vout": 1,
						"puk": "269ddd00ae0a2d4a763bb2152df8736f422a797fdf84d7e6848db2de00628109",
						"sign": "9dfd04dee3c6cf066bdc842b7483261b973d8ff229bd6dcd6951ef250febdca35922ced812964872ed9d82d852fc53680be597a846878449830a5f7066fb4906"
					}
				],
				"vout_total": 2,
				"vout": [
					{
						"value": 22733333,
						"address": "ZHC4eeAnpLFDcKYQkFwt9pR8VmS5QDmzZSMw4",
						"frozen_height": 1617818400
					},
					{
						"value": 199967420636,
						"address": "ZHCETPdw9e2iBxUTACniTqdEwGthboGnUov84",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1315147,
				"payload": "{\"Type\":50001,\"Payloads\":\"eyJpZCI6MTM3MDAwNzE3ODIzMjcyOTYwMSwidHlwZSI6MjAwMDEsImZyb20iOiJaSEM0ZWVBbnBMRkRjS1lRa0Z3dDlwUjhWbVM1UURtelpTTXc0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50Ijo3NTgwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCI2ODIwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxNjA3MTkwMywidm91dHMiOlsiMzAwMDgiLCIzMDAwNiIsIjIwMDAwIiwiMjAwMDQiLCIzMDAxMSIsIjMwMDEwIiwiMzAwMDMiLCIzMDAwNSIsIjMwMDAwIl19\",\"Rand\":0,\"Puk\":\"2ZVnSpLmH6EOkE8M0I5cpA2z65/hxWX03ceR4AbGMqc=\",\"Sign\":\"PCIE6vuqDknWltvTxlCe5BptXLTQ+Bczg9BzDQwKkEWJHkTxccUxJGUSatSqk9DBSwiLuaG2ntsXoRxdw6XSCw==\"}",
				"blockhash": ""
			},
			{
				"hash": "0b00000000000000d66f641a886f1d64fb14f20193fa5163aab096deb5df566262b0d8db54509489",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b0000000000000017212af022b9c4d38ddcfc5ceb5829c10892bb32b68104e952c5d2d43b316ede",
						"vout": 0,
						"puk": "d995674a92e61fa10e904f0cd08e5ca40db3eb9fe1c565f4ddc791e006c632a7",
						"sign": "e5e33d8b7645c21c52b54b5dfca4e51892c915f4b05f9b9ad33f80af7f6239bac1b946ba61b19b3694f59f517f7cac8c726416d98ab77db8abb85b070ee10505"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 2000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1315656,
				"payload": "{\"Type\":50001,\"Payloads\":\"eyJpZCI6MTM3MDc1NzU4OTI4OTkzMDc1MywidHlwZSI6NTAwMDEsImZyb20iOiJaSENFc0RHU1ZnRXBWOGJ0cDJURFBDZFg1aUJoVlZ3b05mYUQ0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50IjozOTkwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCI5MTgwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxNjA3Njk5Mywidm91dHMiOlsiMzAwMDAiLCIyMDAwMCIsIjIwMDA0IiwiMzAwMTAiLCIzMDAwOCIsIjMwMDAzIiwiMzAwMTEiLCIzMDAwNSIsIjMwMDA2Il19\",\"Rand\":0,\"Puk\":\"2ZVnSpLmH6EOkE8M0I5cpA2z65/hxWX03ceR4AbGMqc=\",\"Sign\":\"m19iqzzamZIz3AYrV3qILvtogEibRbhR4KLp4RPHG3FCXXqaSrFwObWobXsoGFNTbtHO6ryEWpszvEoGf7XhCA==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b00000000000000eda656ccd5618ca1725187b904fe205380568a814bf24bbbd1aa02a90eb1b9e3",
						"vout": 0,
						"puk": "d995674a92e61fa10e904f0cd08e5ca40db3eb9fe1c565f4ddc791e006c632a7",
						"sign": ""
					}
				],
				"token_vout_total": 10,
				"token_vout": [
					{
						"value": 3072000000,
						"address": "ZHCL4D79FZ1eoKB1nvHh915uoK8jhimHbHN4",
						"frozen_height": 0
					},
					{
						"value": 183600000,
						"address": "ZHCEsDGSVgEpV8btp2TDPCdX5iBhVVwoNfaD4",
						"frozen_height": 0
					},
					{
						"value": 73440000,
						"address": "ZHCEsDGSVgEpV8btp2TDPCdX5iBhVVwoNfaD4",
						"frozen_height": 0
					},
					{
						"value": 183600000,
						"address": "ZHCCgUBcVDUjFafsdEUmhs4MtAtynbAfJewj4",
						"frozen_height": 0
					},
					{
						"value": 91800000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 0
					},
					{
						"value": 55080000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 1617991200
					},
					{
						"value": 110160000,
						"address": "ZHCPmPhKgkgCkqeUstavLcfZGKe9SafrcbzG4",
						"frozen_height": 0
					},
					{
						"value": 73440000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 1617991200
					},
					{
						"value": 45900000,
						"address": "ZHCPDjGL58NGWt3dDH8rv4b9MvZnMpqgXcZC4",
						"frozen_height": 1617991200
					},
					{
						"value": 100980000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b00000000000000a74da5c96f65fe2526cd0a3c39da53f6de88ee604aeff782799c5442d99a9afe",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b00000000000000374e84225bdc98804bdf120b9d4abc32a0c94385707192209dc49ff4d4522f20",
						"vout": 0,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": "623f1e2c84409c4a930f07ef75b4a40d3e2230bbb62d54f9bc5c3f50927afce2c7ba7a954519ff65738a3ca2c5c481ef962b57dbf1611827b9f3653312f70f04"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9999990,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1315753,
				"payload": "{\"Type\":50010,\"Payloads\":\"eyJpZCI6MTM3MjU1NjU4NDQ0MjEzODYyNiwidHlwZSI6NTAwMTAsImZyb20iOiJaSEM5R2FZRE02ak5NRFpYeFhWQVJvU2RpdVVDYkJDUE5tMjY0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50Ijo4NzUwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCI2MTMwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxNjA3Nzk2NX0=\",\"Rand\":0,\"Puk\":\"INRWx2953g8tBBCSwqTeU3W1avyUis4MIFC/c/Oc4+k=\",\"Sign\":\"OzBdgNmNOtIfjuXIfl8P/CPa68+BZLa9AUKkNRvrE37NPgy/2bNKOAU6QGS2lgxx3GKcg340fjxz4Gm0gHu5Bg==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b000000000000005d7509497caeb14253da5a5aa0acf03e7bcbf52fd753993dbd7e940544468c48",
						"vout": 2,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 8750000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 299991150000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b000000000000005e9261573fc232fd183421df6db4208075546a018223bd31fb42e5707004f0f6",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b0000000000000085aaa3bed8b7d91db487609b6d0e9a5e8f1af91e6bb8e547e15c6904bd5cda7b",
						"vout": 0,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": "3957e40127a0954447d8d1c4063d08788407d71ccf6e13e474f1102263d52d7ef5a4b5e2cea8e250a66a24c935e1e3d2eccbeae5d1282ca0079367feddf2fd0a"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9999990,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1315753,
				"payload": "{\"Type\":50010,\"Payloads\":\"eyJpZCI6MTM3MjU1NjU4NDU1OTU3OTEzNywidHlwZSI6NTAwMTAsImZyb20iOiJaSEM5R2FZRE02ak5NRFpYeFhWQVJvU2RpdVVDYkJDUE5tMjY0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50IjoyOTkwMDAwMDAwLCJwYXlsb2FkIjpbIjEyODM5NTQ1MjQ3NDcwMTAwNDkiLCIyOTkwMDAwMDAiLCIiLCIiXSwiY3JlYXRlZCI6MTYxNjA3Nzk2NX0=\",\"Rand\":0,\"Puk\":\"INRWx2953g8tBBCSwqTeU3W1avyUis4MIFC/c/Oc4+k=\",\"Sign\":\"nA3GEsp7qYFmC2UOfb8CEUB7XK8ZE9ps5dfkpUiHUVJjyuCrPfO10n6AEMxmum7hciNqkErIU9K98mJslNjdCA==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b00000000000000488e35f10e02213be0c7520585edcf1296ac66b2efeec29736676477e021eb2a",
						"vout": 2,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 2990000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 299996910000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			},
			{
				"hash": "0b0000000000000044ea8ab4d197af5892bd9febd78a8ae2c6b066073d8c56fcfac08c0834776789",
				"type": 11,
				"vin_total": 1,
				"vin": [
					{
						"txid": "0b00000000000000af8e8bbde8970e5961d7cea29928dce80ba9c64cdc05a232a70f22a47ace0f55",
						"vout": 0,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": "628d6e601ac8d48e86ba2e5d4926cd4472b87132a75367b8788c46a995eea1d4a44b8adfff6e400cb50442b56d9a28a3e48d5ed532c83c87fe8ee3d49ea9f908"
					}
				],
				"vout_total": 1,
				"vout": [
					{
						"value": 9999990,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"gas": 0,
				"lock_height": 1316922,
				"payload": "{\"Type\":50010,\"Payloads\":\"eyJpZCI6MTM3MjYwNTI3NDk0NzA2MzgwOSwidHlwZSI6NTAwMTAsImZyb20iOiJaSEM5aHl3bWFMZkFKVUJoTFo2czFjYWp2QzNGWXRKSjVFTUY0IiwidG8iOiJaSENMNEQ3OUZaMWVvS0IxbnZIaDkxNXVvSzhqaGltSGJITjQiLCJjb250cmFjdCI6IjBhMDAwMDAwMDAwMDAwMDBmZDU4OTY2MmJhYjcyNGJhYTk3MTg1NDUyMWRjOWRmZjRiYzJlNmI5MzYyZmE3ODc1OWY4OGRmMzZkZTA2NjU3IiwiYW1vdW50IjoxMTUwMDAwMDAwMCwicGF5bG9hZCI6WyIxMjgzOTU0NTI0NzQ3MDEwMDQ5IiwiMjMwMDAwMDAwMCIsIiIsIiJdLCJjcmVhdGVkIjoxNjE2MDg5NjYxfQ==\",\"Rand\":0,\"Puk\":\"INRWx2953g8tBBCSwqTeU3W1avyUis4MIFC/c/Oc4+k=\",\"Sign\":\"qws6H26O3of5cL28JA6LxYVtdq897Q1isRFjMDon1+mLH/TLof03QkwPAjQTYURI9lqAlIuf+1yP1GVr+jJUCw==\"}",
				"blockhash": "",
				"token_name": "",
				"token_symbol": "",
				"token_supply": 0,
				"token_vin_total": 1,
				"token_vin": [
					{
						"txid": "0b00000000000000742b6109150dcd5ace8c410a878c605049427cdc4434d3dc478782030cb728eb",
						"vout": 1,
						"puk": "20d456c76f79de0f2d041092c2a4de5375b56afc948ace0c2050bf73f39ce3e9",
						"sign": ""
					}
				],
				"token_vout_total": 2,
				"token_vout": [
					{
						"value": 11500000000,
						"address": "ZHC3sDuZPmeBeLfCi7HV3t2crHgGkAjejE9d4",
						"frozen_height": 0
					},
					{
						"value": 299983810000000,
						"address": "ZHCMmuVUkjAwpMz4fd1gEV55pEXXjGK1iPC4",
						"frozen_height": 0
					}
				],
				"token_publish_txid": ""
			}
		]`
