package rpc

import (
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/keystore/kstore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/rpc/model"
	"mandela/rpc/networks"
	"mandela/rpc/sharebox"
	"mandela/rpc/store"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
)

type serverHandler func(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) ([]byte, error)

//访问接口，统一header里传user,password
var rpcHandler = map[string]serverHandler{
	"getinfo":               handleGetInfo,         //获取基本信息{"method":"getinfo","params":{"id":10}}
	"getnewaddress":         handleGetNewAddress,   //创建新地址 {"method":"getnewaddress","params":{"password":"123456"}}
	"listaccounts":          handleListAccounts,    //帐号列表{"method":"listaccounts"}
	"getaccount":            handleGetAccount,      //获取某一帐号余额{"method":"getaccount","params":{"address":"1AX9mfCRZkdEg5Ci3G5SLcyGgecj6GTzLo"}}
	"validateaddress":       handleValidateAddress, //验证地址{"method":"validateaddress","params":{"address":"12EUY1EVnLJe4Ejb1VaL9NbuDQbBEV"}}
	"import":                Import,                //导入钱包
	"export":                Export,                //导出钱包
	"sendtoaddress":         sendToAddress,         //转账
	"sendtoaddressmore":     sendToAddressmore,     //给多人转账
	"depositin":             depositIn,             //缴纳押金，成为见证人
	"depositout":            depositOut,            //退还押金
	"votein":                voteIn,                //给见证人投票押金
	"voteout":               voteOut,               //退还给见证人投票押金
	"updatepwd":             UpdatePwd,             //修改支付密码
	"createkeystore":        CreateKeystore,        //创建密钥文件
	"namesin":               NameIn,                //域名注册，续费，修改
	"namesout":              NameOut,               //域名注销，退还押金
	"getnames":              GetNames,              //获取自己注册的域名列表
	"findname":              FindName,              //查询域名
	"gettransactionhistory": GetTransactionHistoty, //获得转账交易历史记录
	"getwitnessinfo":        GetWitnessInfo,        //查询见证人状态
	"getcandidatelist":      GetCandidateList,      //获得候选见证人列表
	"getcommunitylist":      GetCommunityList,      //获取社区节点列表
	"getvotelist":           GetVoteList,           //获得自己投过票的列表
	"findtx":                FindTx,                //检查一笔交易是否成功
	"findblock":             FindBlock,             //通过区块高度查询一个区块信息
	"getcommunityreward":    GetCommunityReward,    //获取一个社区累计奖励
	"sendcommunityreward":   SendCommunityReward,   //分发社区奖励

	//好友管理部分
	"getfriendlist": GetContactsList, //获取好友列表
	"addfriend":     AddContacts,     //添加好友
	"delfriend":     DelContacts,     //删除好友
	"updatefriend":  UpdateContacts,  //修改好友信息
	"getfriendinfo": GetFriendInfo,   //获取好友信息

	//聊天模块
	"agreetoadd":            AgreeToAdd,            //同意添加好友
	"getmsghistory":         GetMsgHistory,         //获取消息历史记录
	"sendmsg":               SendMsg,               //发送文本消息
	"resendmsg":             ResendMsg,             //重发文本消息
	"getnewmsg":             GetNewMsg,             //获取新消息
	"isreadmsg":             IsReadMsg,             //设置消息已读
	"isreadaddfirend":       IsReadAddFirend,       //修改添加好友消息已读状态
	"removemsg":             RemoveMsg,             //删除聊天记录
	"removemsgall":          RemoveMsgAll,          //删除指定好友的所有聊天记录
	"sendpicmsg":            SendPicMsg,            //发送信息
	"sendfilemsg":           SendFileMsg,           //发送信息
	"resendfilemsg":         ReSendFileAction,      //重发图文信息
	"getmsgstate":           GetMsgState,           //获取消息发送状态
	"updateproperty":        UpdateNickName,        //修改用户昵称
	"getmyinfo":             GetMyInfo,             //获取用户属性
	"getfriendbasecoinaddr": GetFriendBaseCoinAddr, //获取好友收款地址
	"sendpaymsg":            SendPayMsg,            //发送收款消息
	"updatepaystatus":       UpdatePayStatus,       //修改到帐状态

	"sendmsgothore": More, //发送其他消息，发送不同类型的消息。

	//网络部分
	"getnetworkinfo": networks.NetworkInfo, //获取本节点网络信息

	//sharebox共享存储部分
	"getsharefolderlist":       sharebox.ShareFolderList,          //查询共享文件夹列表
	"addlocalsharefolder":      sharebox.AddLocalShareFoler,       //添加本地共享文件夹列表
	"dellocalsharefolder":      sharebox.DelLocalShareFoler,       //删除本地共享文件夹列表
	"getremotesharefolderlist": sharebox.GetRemoteShareFolderList, //获取远端节点共享文件夹列表

	//store 云存储部分
	// "uploadfile": store.UploadFile, //上传文件
	"searchfile":       store.SearchFileInfo,   //搜索文件索引
	"delfile":          store.DelFileInfo,      //删除本地文件索引
	"addfile":          store.AddFileInfo,      //增加本地索引
	"downloadprocone":  store.DownloadProcOne,  //单个文件下载进度
	"downloadproc":     store.DownloadProc,     //文件下载进度
	"downloadcomplete": store.DownloadComplete, //获取已下载列表
	"downloadstop":     store.DownLoadStop,     //暂停下载
	"downloaddel":      store.DownLoadDel,      //删除下载
	"addfolder":        store.AddFolder,        //增加目录
	"delfolder":        store.DelFolder,        //删除目录
	"upfolder":         store.UpFolder,         //修改目录
	"listfolder":       store.ListFolder,       //目录列表
	"moveto":           store.Moveto,           //修改文件所属文件夹
	"getspacesize":     store.GetSpaceSize,     //查询空间大小
	"addspacesize":     store.AddSpaceSize,     //增加空间大小
	"delspacesize":     store.DelSpaceSize,     //删除空间大小
	"getfilelist":      store.GetFileList,      //获取文件列表

	"stopservice": StopService, //关闭服务器

}

/*
	关闭服务
*/
func StopService(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	utils.StopService()
	res, err = model.Tojson("success")
	return
}

//获取基本信息
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "balance": 0,
//        "testnet": false,
//        "blocks": 0
//    }
//}
func handleGetInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) ([]byte, error) {
	value := mining.GetBalance()

	info := model.Getinfo{
		Netid:          []byte(config.AddrPre),
		TotalAmount:    config.Mining_coin_total,
		Balance:        value,
		Testnet:        true,
		Blocks:         mining.GetLongChain().GetCurrentBlock(),
		Group:          0,
		StartingBlock:  mining.GetLongChain().GetStartingBlock(), //区块开始高度
		HighestBlock:   mining.GetHighestBlock(),                 //所链接的节点的最高高度
		CurrentBlock:   mining.GetLongChain().GetCurrentBlock(),  //已经同步到的区块高度
		PulledStates:   mining.GetLongChain().GetPulledStates(),  //正在同步的区块高度
		BlockTime:      config.Mining_block_time,                 //出块时间
		LightNode:      config.Mining_light_min,                  //轻节点押金数量
		CommunityNode:  config.Mining_vote,                       //社区节点押金数量
		WitnessNode:    config.Mining_deposit,                    //见证人押金数量
		NameDepositMin: config.Mining_name_deposit_min,           //
		AddrPre:        config.AddrPre,                           //
	}
	res, err := model.Tojson(info)
	return res, err
}

//创建新地址
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "address": "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q"
//    }
//}
func handleGetNewAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if !rj.VerifyType("password", "string") {
		res, err = model.Errcode(model.TypeWrong, "password")
		return
	}
	password, ok := rj.Get("password")
	if !ok {
		res, err = model.Errcode(model.NoField, "password")
		return
	}
	// keystore.GetNewAddr()

	addr, err := keystore.GetNewAddr(password.(string))
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, _ = model.Errcode(model.Nomarl)
		return
	}
	getnewadress := model.GetNewAddress{Address: addr.B58String()}
	res, err = model.Tojson(getnewadress)
	return
}

type AccountVO struct {
	Index    int    //排序
	AddrCoin string //收款地址
	Value    uint64 //余额
	Type     int    //1=见证人;2=社区节点;3=轻节点;4=什么也不是
}

//地址列表
func handleListAccounts(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vos := make([]AccountVO, 0)

	// list := make(map[string]uint64)
	addr := keystore.GetAddr()
	for i, val := range addr {
		// list[val.B58String()] = mining.GetBalanceForAddr(&val)
		vo := AccountVO{
			Index:    i,
			AddrCoin: val.B58String(),
			Value:    mining.GetBalanceForAddr(&val),
			Type:     mining.GetAddrState(val),
		}
		vos = append(vos, vo)
	}
	res, err = model.Tojson(vos)
	return res, err
}

//获取某一帐号余额
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "Balance": 0
//    }
//}
func handleGetAccount(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}
	_, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	// fmt.Println(addr)
	getaccount := model.GetAccount{}
	res, err = model.Tojson(getaccount)
	return
}

//验证地址
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "IsVerify": true,
//        "IsMine": false,
//        "IsType": 1,
//        "Version": 0,
//        "ExpVersion": 0,
//        }
//    }
//}
func handleValidateAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if !rj.VerifyType("address", "string") {
		res, err = model.Errcode(model.TypeWrong, "address")
		return
	}
	addr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addrCoin := crypto.AddressFromB58String(addr.(string))
	ok = crypto.ValidAddr(config.AddrPre, addrCoin)
	res, err = model.Tojson(ok)
	return
}

/*
	转账
*/
func sendToAddress(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// fmt.Println("++++++++++++++++++++\n时间开始")
	// start := time.Now()

	addrItr, ok := rj.Get("address")
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addr := addrItr.(string)

	dst := crypto.AddressFromB58String(addr)
	if !crypto.ValidAddr(config.AddrPre, dst) {
		res, err = model.Errcode(model.ContentIncorrectFormat, "address")
		return
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := uint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(model.AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	commentItr, ok := rj.Get("comment")
	if !ok {
		res, err = model.Errcode(model.NoField, "comment")
		return
	}
	comment := commentItr.(string)
	// fmt.Println("转账到地址", addr, amount, pwd, comment)

	// dst, e := utils.FromB58String(addr)
	// if err != nil {
	// 	err = e
	// 	res, _ = model.Errcode(5003, "error")
	// 	return
	// }

	//查询余额是否足够
	if amount > mining.GetBalance() {
		res, err = model.Errcode(model.BalanceNotEnough)
		return
	}
	// engine.Log.Info("创建转账交易错误 00000000000000")
	txpay, err := mining.SendToAddress(&dst, amount, gas, pwd, comment)
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(model.AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)
	return
}

/*
	多人转账
*/
type PayNumber struct {
	Address string `json:"address"` //转账地址
	Amount  uint64 `json:"amount"`  //转账金额
}

/*
	给多人转账
*/
func sendToAddressmore(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// fmt.Println("++++++++++++++++++++\n时间开始")
	// start := time.Now()

	addrItr, ok := rj.Get("addresses")
	if !ok {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	bs, err := json.Marshal(addrItr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}

	addrs := make([]PayNumber, 0)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&addrs)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "addresses")
		return
	}
	//给多人转账，可是没有地址
	if len(addrs) <= 0 {
		res, err = model.Errcode(model.NoField, "addresses")
		return
	}

	amount := uint64(0)

	addr := make([]mining.PayNumber, 0)
	for _, one := range addrs {
		dst := crypto.AddressFromB58String(one.Address)
		//验证地址前缀
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(model.ContentIncorrectFormat, "addresses")
			return
		}
		pnOne := mining.PayNumber{
			Address: dst,        //转账地址
			Amount:  one.Amount, //转账金额
		}
		addr = append(addr, pnOne)
		amount += one.Amount
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	commentItr, ok := rj.Get("comment")
	if !ok {
		res, err = model.Errcode(model.NoField, "comment")
		return
	}
	comment := commentItr.(string)

	//查询余额是否足够
	if amount+gas > mining.GetBalance() {
		res, err = model.Errcode(model.BalanceNotEnough)
		return
	}

	// engine.Log.Info("创建转账交易错误 00000000000000")
	txpay, err := mining.SendToMoreAddress(addr, gas, pwd, comment)
	if err != nil {
		// engine.Log.Info("创建转账交易错误 11111111")
		if err.Error() == config.ERROR_password_fail.Error() {
			// engine.Log.Info("创建转账交易错误 222222222222")
			res, err = model.Errcode(model.FailPwd)
			return
		}
		// engine.Log.Info("创建转账交易错误 333333333333")
		if err.Error() == config.ERROR_amount_zero.Error() {
			res, err = model.Errcode(model.AmountIsZero, "amount")
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result, err := utils.ChangeMap(txpay)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	result["hash"] = hex.EncodeToString(*txpay.GetHash())

	res, err = model.Tojson(result)
	return
}

//缴纳押金，成为见证人
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//        "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//        "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//    }
//}
func depositIn(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := uint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(model.AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	payload := ""
	payloadItr, ok := rj.Get("payload")
	if ok {
		payload = payloadItr.(string)
	}

	//查询余额是否足够
	if amount > mining.GetBalance() {
		res, err = model.Errcode(model.BalanceNotEnough)
		return
	}

	err = mining.DepositIn(amount, gas, pwd, payload)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	config.SubmitDepositin = true
	return
}

//退还押金
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//        "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//        "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//    }
//}
func depositOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	addr := ""
	addrItr, ok := rj.Get("address")
	if ok {
		addr = addrItr.(string)

	}

	if addr != "" {
		dst := crypto.AddressFromB58String(addr)
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(model.ContentIncorrectFormat, "address")
			return
		}
	}

	amount := uint64(0)
	amountItr, ok := rj.Get("amount")
	if ok {
		amount = uint64(amountItr.(float64))
		if amount < 0 {
			res, err = model.Errcode(model.AmountIsZero, "amount")
			return
		}
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	engine.Log.Info("address:%s amount:%d gas:%d", addr, amount, gas)

	err = mining.DepositOut(addr, amount, gas, pwd)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

//缴纳押金，成为见证人
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//        "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//        "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//    }
//}
func voteIn(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	addr := ""
	addrItr, ok := rj.Get("address")
	if ok {
		addr = addrItr.(string)
	}

	if addr != "" {
		dst := crypto.AddressFromB58String(addr)
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(model.ContentIncorrectFormat, "address")
			return
		}
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(5002, "amount")
		return
	}
	amount := uint64(amountItr.(float64))
	if amount <= 0 {
		res, err = model.Errcode(model.AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	payload := ""
	payloadItr, ok := rj.Get("payload")
	if ok {
		payload = payloadItr.(string)
	}

	var witnessAddr crypto.AddressCoin
	witnessAddrItr, ok := rj.Get("witness")
	if ok {
		// res, err = model.Errcode(5002, "witness")
		// return
		witnessStr := witnessAddrItr.(string)

		witnessAddr = crypto.AddressFromB58String(witnessStr)

		if witnessStr != "" {
			dst := crypto.AddressFromB58String(witnessStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(model.ContentIncorrectFormat, "witness")
				return
			}
		}
	}

	//查询余额是否足够
	if amount > mining.GetBalance() {
		res, err = model.Errcode(model.BalanceNotEnough)
		return
	}

	err = mining.VoteIn(voteType, witnessAddr, addr, amount, gas, pwd, payload)
	if err != nil {
		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

//退还押金
//{
//    "jsonrpc": "2.0",
//    "code": 2000,
//    "result": {
//        "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
//        "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
//        "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
//    }
//}
func voteOut(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	txid := ""
	txidItr, ok := rj.Get("txid")
	if ok {
		txid = txidItr.(string)
	}

	addr := ""
	addrItr, ok := rj.Get("address")
	if ok {
		addr = addrItr.(string)
	}
	if addr != "" {
		dst := crypto.AddressFromB58String(addr)
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(model.ContentIncorrectFormat, "address")
			return
		}
	}

	amountItr, ok := rj.Get("amount")
	if !ok {
		res, err = model.Errcode(model.NoField, "amount")
		return
	}
	amount := uint64(amountItr.(float64))
	if amount < 0 {
		res, err = model.Errcode(model.AmountIsZero, "amount")
		return
	}

	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	// witnessAddrItr, ok := rj.Get("witness")
	// if !ok {
	// 	res, err = model.Errcode(model.NoField, "witness")
	// 	return
	// }
	// witnessStr := witnessAddrItr.(string)
	// witnessAddr := crypto.AddressFromB58String(witnessStr)

	var witnessAddr crypto.AddressCoin
	witnessAddrItr, ok := rj.Get("witness")
	if ok {
		witnessStr := witnessAddrItr.(string)
		witnessAddr = crypto.AddressFromB58String(witnessStr)
	}

	err = mining.VoteOut(&witnessAddr, txid, addr, amount, gas, pwd)
	if err != nil {
		// engine.Log.Info("--------------- 取消投票错误" + err.Error())

		if err.Error() == config.ERROR_password_fail.Error() {
			res, err = model.Errcode(model.FailPwd)
			return
		}
		//余额不足
		if err.Error() == config.ERROR_not_enough.Error() {
			res, err = model.Errcode(model.BalanceNotEnough)
			return
		}
		//投票已经存在
		if err.Error() == config.ERROR_vote_exist.Error() {
			res, err = model.Errcode(model.VoteExist)
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson("success")
	return
}

// /*
// 	缴纳轻节点押金
// */
// func depositInLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(5002, "amount")
// 		return
// 	}
// 	amount := uint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(5002, "gas")
// 		return
// 	}
// 	gas := uint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(5002, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	var witnessAddr crypto.AddressCoin
// 	// witnessAddrItr, ok := rj.Get("witness")
// 	// if ok {
// 	// 	// res, err = model.Errcode(5002, "witness")
// 	// 	// return
// 	// 	witnessStr := witnessAddrItr.(string)

// 	// 	witnessAddr = crypto.AddressFromB58String(witnessStr)
// 	// }

// 	//查询余额是否足够
// 	if amount > mining.GetBalance() {
// 		res, err = model.Errcode(model.Nomarl, "余额不足")
// 		return
// 	}

// 	err = mining.VoteIn(witnessAddr, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	轻节点退还押金
// */
// func depositOutLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	txid := ""
// 	txidItr, ok := rj.Get("txid")
// 	if ok {
// 		txid = txidItr.(string)
// 	}

// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "amount")
// 		return
// 	}
// 	amount := uint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "gas")
// 		return
// 	}
// 	gas := uint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	witnessAddrItr, ok := rj.Get("witness")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "witness")
// 		return
// 	}
// 	witnessStr := witnessAddrItr.(string)
// 	witnessAddr := crypto.AddressFromB58String(witnessStr)

// 	err = mining.VoteOut(&witnessAddr, txid, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	轻节点给社区节点投票
// */
// func voteInLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(5002, "amount")
// 		return
// 	}
// 	amount := uint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(5002, "gas")
// 		return
// 	}
// 	gas := uint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(5002, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	var witnessAddr crypto.AddressCoin
// 	witnessAddrItr, ok := rj.Get("witness")
// 	if ok {
// 		// res, err = model.Errcode(5002, "witness")
// 		// return
// 		witnessStr := witnessAddrItr.(string)

// 		witnessAddr = crypto.AddressFromB58String(witnessStr)
// 	}

// 	//查询余额是否足够
// 	if amount > mining.GetBalance() {
// 		res, err = model.Errcode(model.Nomarl, "余额不足")
// 		return
// 	}

// 	err = mining.VoteIn(witnessAddr, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// /*
// 	轻节点给社区节点投票 退还押金
// */
// func voteOutLight(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	txid := ""
// 	txidItr, ok := rj.Get("txid")
// 	if ok {
// 		txid = txidItr.(string)
// 	}

// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "amount")
// 		return
// 	}
// 	amount := uint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "gas")
// 		return
// 	}
// 	gas := uint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	witnessAddrItr, ok := rj.Get("witness")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "witness")
// 		return
// 	}
// 	witnessStr := witnessAddrItr.(string)
// 	witnessAddr := crypto.AddressFromB58String(witnessStr)

// 	err = mining.VoteOut(&witnessAddr, txid, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

// //退还押金
// //{
// //    "jsonrpc": "2.0",
// //    "code": 2000,
// //    "result": {
// //        "12FRzz2xrVtEm9cwzgFArrLE7VA7ks": 0,
// //        "12GJJknncS2MmbXh26ZHAMbd3CjCHy": 0,
// //        "12Hixu5fzDrVoQt1fDL5vHw2Aahw1q": 0
// //    }
// //}
// func voteOutOne(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

// 	txidItr, ok := rj.Get("txid")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "txid")
// 		return
// 	}
// 	txid := txidItr.(string)

// 	addr := ""
// 	addrItr, ok := rj.Get("address")
// 	if ok {
// 		addr = addrItr.(string)
// 	}

// 	amountItr, ok := rj.Get("amount")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "amount")
// 		return
// 	}
// 	amount := uint64(amountItr.(float64))

// 	gasItr, ok := rj.Get("gas")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "gas")
// 		return
// 	}
// 	gas := uint64(gasItr.(float64))

// 	pwdItr, ok := rj.Get("pwd")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "pwd")
// 		return
// 	}
// 	pwd := pwdItr.(string)

// 	witnessAddrItr, ok := rj.Get("witness")
// 	if !ok {
// 		res, err = model.Errcode(model.NoField, "witness")
// 		return
// 	}
// 	witnessStr := witnessAddrItr.(string)
// 	witnessAddr := crypto.AddressFromB58String(witnessStr)

// 	err = mining.VoteOutOne(txid, addr, amount, gas, pwd)
// 	if err != nil {
// 		res, err = model.Errcode(model.Nomarl, err.Error())
// 		return
// 	}
// 	res, err = model.Tojson("success")
// 	return
// }

/*
	修改钱包支付密码
*/
func UpdatePwd(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	oldpwdItr, ok := rj.Get("oldpwd")
	if !ok {
		res, err = model.Errcode(5002, "oldpwd")
		return
	}
	oldpwd := oldpwdItr.(string)

	pwdItr, ok := rj.Get("newpwd")
	if !ok {
		res, err = model.Errcode(5002, "newpwd")
		return
	}
	pwd := pwdItr.(string)

	ok, err = keystore.UpdatePwd(oldpwd, pwd)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if !ok {
		//密码错误
		res, err = model.Errcode(model.FailPwd, errors.New("password fail").Error())
		return
	}
	config.Wallet_keystore_default_pwd = pwd
	res, err = model.Tojson("success")
	return
}

/*
	创建钱包
*/
func CreateKeystore(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	randomItr, ok := rj.Get("random")
	if !ok {
		res, err = model.Errcode(model.NoField, "random")
		return
	}

	randomItrs := randomItr.([]interface{})
	buf := bytes.NewBuffer(nil)
	for _, one := range randomItrs {
		onePoint := uint16(one.(float64))
		_, e := buf.Write(utils.Uint16ToBytes(onePoint))
		if e != nil {
			res, err = model.Errcode(model.Nomarl, e.Error())
			return
		}
	}
	if buf.Len() != 4000 {
		//随机数长度不等于2000
		res, err = model.Errcode(model.Nomarl, "Random number length not equal to 2000")
		return
	}

	rand1 := buf.Bytes()[:2000]
	rand2 := buf.Bytes()[2000:]

	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		return
	}
	pwd := pwdItr.(string)

	err = keystore.CreateKeystoreRand(filepath.Join(config.Path_configDir, config.Core_keystore), rand1, rand2, pwd)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	res, err = model.Tojson("success")
	return
}

/*
	导出钱包
*/
func Export(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	pwdItr, ok := rj.Get("password")
	if !ok {
		res, err = model.Errcode(5002, "password")
		return
	}
	pwd := pwdItr.(string)
	rs := kstore.Export(pwd)
	di := kstore.ParseDataInfo(rs)
	if di.Code == 500 {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	res, err = model.Tojson(di.Data)
	return
}

/*
	导入钱包
*/
func Import(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	pwdItr, ok := rj.Get("password")
	if !ok {
		res, err = model.Errcode(5002, "password")
		return
	}
	pwd := pwdItr.(string)
	seeds, ok := rj.Get("seed")
	if !ok {
		res, err = model.Errcode(5002, "seed")
		return
	}
	seed := seeds.(string)
	rs := kstore.Import(config.Path_configDir, pwd, seed)
	di := kstore.ParseDataInfo(rs)
	if di.Code == 500 {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	res, err = model.Tojson(di.Data)
	return
}
