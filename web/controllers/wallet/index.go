package wallet

import (
	"mandela/chain_witness_vote"
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/rpc"
	"mandela/rpc/model"

	"github.com/astaxie/beego"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Index struct {
	beego.Controller
}

func (this *Index) Index() {
	//	names, _ := store.GetFileinfoToSelfAll()

	//	fmt.Println("网络文件个数为", len(names))
	//	this.Data["Names"] = names

	this.Data["CheckKey"] = chain_witness_vote.CheckKey()

	this.TplName = "wallet/index.tpl"
}

func (this *Index) Getinfo() {
	info := model.Getinfo{
		Netid:          nil,                                       //
		TotalAmount:    config.Mining_coin_total,                  //
		Balance:        0,                                         //
		BalanceFrozen:  0,                                         //
		BalanceLockup:  0,                                         //
		Testnet:        false,                                     //
		Blocks:         mining.GetLongChain().GetCurrentBlock(),   //
		Group:          0,                                         //
		StartingBlock:  mining.GetLongChain().GetStartingBlock(),  //区块开始高度
		StartBlockTime: mining.GetLongChain().GetStartBlockTime(), //
		HighestBlock:   mining.GetHighestBlock(),                  //所链接的节点的最高高度
		CurrentBlock:   mining.GetLongChain().GetCurrentBlock(),   //已经同步到的区块高度
		PulledStates:   mining.GetLongChain().GetPulledStates(),   //正在同步的区块高度
		BlockTime:      config.Mining_block_time,                  //出块时间
		LightNode:      config.Mining_light_min,                   //轻节点押金数量
		CommunityNode:  config.Mining_vote,                        //社区节点押金数量
		WitnessNode:    config.Mining_deposit,                     //见证人押金数量
		NameDepositMin: config.Mining_name_deposit_min,            //
		AddrPre:        config.AddrPre,                            //
		TokenBalance:   nil,                                       //
	}
	this.Data["json"] = info
	this.ServeJSON()
	return
}

func (this *Index) Block() {

	// engine.Log.Info("requstBody %s", string(this.Ctx.Input.RequestBody))

	out := make(map[string]interface{})
	paramsMap := make(map[string]interface{})

	err := json.Unmarshal(this.Ctx.Input.RequestBody, &paramsMap)
	if err != nil {
		out["Msg"] = "not find param"
		out["Code"] = 1
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	value, ok := paramsMap["height"]
	if !ok {
		out["Msg"] = "not find height param"
		out["Code"] = 1
		this.Data["json"] = out
		this.ServeJSON()
		return
	}

	height, ok := value.(float64)
	if !ok {
		out["Msg"] = "height param fail"
		out["Code"] = 1
		this.Data["json"] = out
		this.ServeJSON()
		return
	}

	bhvo := mining.BlockHeadVO{
		Txs: make([]mining.TxItr, 0), //交易明细
	}
	// engine.Log.Info("查询索引 %s", config.BlockHeight+strconv.Itoa(int(height)))
	bh := mining.LoadBlockHeadByHeight(uint64(height))
	// bh := mining.FindBlockHead(uint64(height))
	if bh == nil {
		out["Code"] = 1
		out["Msg"] = "not find blockhead"
		this.Data["json"] = out
		this.ServeJSON()
		return
	}
	bhvo.BH = bh

	for _, one := range bh.Tx {
		txItr, e := mining.LoadTxBase(one)
		// txItr, e := mining.FindTxBase(one)
		if e != nil {
			out["Msg"] = "not find tx"
			out["Code"] = 1
			this.Data["json"] = out
			this.ServeJSON()
			return
		}
		bhvo.Txs = append(bhvo.Txs, txItr)
	}

	out["Code"] = 0
	out["Data"] = bhvo
	this.Data["json"] = out
	this.ServeJSON()
	return
}

func (this *Index) GetWitnessList() {

	// engine.Log.Info("requstBody %s", string(this.Ctx.Input.RequestBody))

	out := make(map[string]interface{})

	wbg := mining.GetWitnessListSort()

	wvos := make([]rpc.WitnessVO, 0)
	for _, one := range append(wbg.Witnesses, wbg.WitnessBackup...) {

		name := mining.FindWitnessName(*one.Addr)

		wvo := rpc.WitnessVO{
			Addr:            one.Addr.B58String(), //见证人地址
			Payload:         name,                 //
			Score:           one.Score,            //押金
			Vote:            one.VoteNum,          //      voteValue,            //投票票数
			CreateBlockTime: one.CreateBlockTime,  //预计出块时间
		}
		wvos = append(wvos, wvo)
	}

	out["Code"] = 0
	out["Data"] = wvos
	this.Data["json"] = out
	this.ServeJSON()
	return
}

///*
//	添加文件到云盘
//*/
//func (this *Index) AddFile() {
//	out := make(map[string]interface{})

//	//	this.GetFile()
//	hs, err := this.GetFiles("files[]")
//	if err != nil {
//		fmt.Println("获取文件头失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}

//	f, err := hs[0].Open()
//	if err != nil {
//		fmt.Println("打开文件失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	filename := hs[0].Filename

//	//	fi = &FileInfo{
//	//		Name: p.FileName(),
//	//		Type: p.Header.Get("Content-Type"),
//	//	}

//	//	f, h, err := this.Ctx.Request.FormFile("file")
//	//	fmt.Println(f, h, err)

//	//保存文件到本地
//	newfile, err := os.OpenFile(filepath.Join(config.Store_temp, filename), os.O_RDWR|os.O_CREATE, os.ModePerm)
//	if err != nil {
//		fmt.Println("保存文件到本地失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	//	newfile, err := os.Create(filepath.Join(config.Store_temp, filename))
//	buf, err := ioutil.ReadAll(f)
//	newfile.Write(buf)
//	newfile.Close()
//	f.Close()

//	//文件切片
//	fi, err := store.Diced(filename)
//	if err != nil {
//		fmt.Println("文件切片失败", err)
//		//文件切片失败
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	fmt.Println("22222")

//	//	store.AddFileinfoToLocal(fi, true)
//	err = store.AddFileinfoToSelf(fi, true)
//	if err != nil {
//		fmt.Println("保存文件索引到本地失败", err)
//		out["Code"] = 1
//		this.Data["json"] = out
//		this.ServeJSON()
//		return
//	}
//	fmt.Println("33333")

//	//上传文件信息到网络中
//	err = store.UpNetFileinfo(fi)
//	if err != nil {
//		fmt.Println("上传文件索引到网络中失败", err)
//		out["Code"] = 1
//	} else {
//		out["Code"] = 0
//		out["HashName"] = fi.Hash.B58String()
//	}

//	this.Data["json"] = out
//	this.ServeJSON()
//	return

//	//	this.TplName = "store/index.tpl"
//}

///*
//	下载文件
//	1.查询文件信息
//	2.把文件同步到本地
//	3.传输文件给请求者
//*/
//func (this *Index) GetFile() {
//	fn := this.Ctx.Input.Param(":hash")
//	haveLocal := true

//	//查询文件信息
//	fileinfo := store.FindFileinfo(fn)
//	if fileinfo == nil {
//		haveLocal = false
//		var err error
//		fileinfo, err = store.FindFileinfoOpt(fn)
//		if fileinfo == nil || err != nil {
//			fmt.Println("网络中查找文件信息失败")
//			return
//		}
//	}
//	fmt.Println("获取到的文件信息", fileinfo)

//	//把文件下载到本地
//	err := store.DownloadFileOpt(fileinfo)
//	if err != nil {
//		fmt.Println("下载文件失败", err)
//		return
//	}
//	fmt.Println("下载文件成功")

//	if !haveLocal {
//		store.AddFileinfoToLocal(fileinfo, true)
//	}

//	//
//	file, err := os.Open(filepath.Join(config.Store_temp, fileinfo.Name))
//	if err != nil {
//		fmt.Println(err)
//		return
//	}

//	io.Copy(this.Ctx.ResponseWriter, file)
//	file.Close()

//	//上传文件信息到网络中
//	//	store.UpNetFileinfo(fileinfo)

//	//	this.TplName = "store/index.tpl"

//	//	io.Copy(this.Ctx.ResponseWriter, buf)
//}
