package rpc

import (
	"mandela/chain_witness_vote/mining"
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils/crypto"
	"mandela/rpc/model"
	"encoding/hex"
	"math/big"
	"net/http"
	"sort"
	"time"
)

type HistoryItemVO struct {
	GenerateId string   //
	IsIn       bool     //资金转入转出方向，true=转入;false=转出;
	Type       uint64   //交易类型
	InAddr     []string //输入地址
	OutAddr    []string //输出地址
	Value      uint64   //交易金额
	Txid       string   //交易id
	Height     uint64   //区块高度
}

/*
	历史记录
*/
func GetTransactionHistoty(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	id := ""
	idItr, ok := rj.Get("id")
	if ok {
		if !rj.VerifyType("id", "string") {
			res, err = model.Errcode(model.TypeWrong, "id")
			return
		}
		id = idItr.(string)
	}
	var startId *big.Int
	if id != "" {
		var ok bool
		startId, ok = new(big.Int).SetString(id, 10)
		if !ok {
			res, err = model.Errcode(model.TypeWrong, "id")
			return
		}
	}

	total := 0
	totalItr, ok := rj.Get("total")
	if ok {
		total = int(totalItr.(float64))
		// fmt.Println("total", total)
	}

	his := mining.GetLongChain().GetHistoryBalance(startId, total)

	hivos := make([]HistoryItemVO, 0)
	for _, one := range his {
		hivo := HistoryItemVO{
			GenerateId: one.GenerateId.String(),      //
			IsIn:       one.IsIn,                     //资金转入转出方向，true=转入;false=转出;
			Type:       one.Type,                     //交易类型
			InAddr:     make([]string, 0),            //输入地址
			OutAddr:    make([]string, 0),            //输出地址
			Value:      one.Value,                    //交易金额
			Txid:       hex.EncodeToString(one.Txid), //交易id
			Height:     one.Height,                   //区块高度
		}

		for _, two := range one.InAddr {
			hivo.InAddr = append(hivo.InAddr, two.B58String())
		}

		for _, two := range one.OutAddr {
			hivo.OutAddr = append(hivo.OutAddr, two.B58String())
		}

		hivos = append(hivos, hivo)
	}

	res, err = model.Tojson(hivos)
	return
}

/*
	见证人信息
*/
type WitnessInfo struct {
	IsCandidate bool   //是否是候选见证人
	IsBackup    bool   //是否是备用见证人
	IsKickOut   bool   //没有按时出块，已经被踢出局，只有退还押金，重新缴纳押金成为候选见证人
	Addr        string //见证人地址
	Payload     string //
	Value       uint64 //见证人押金
}

/*
	查询自己是否是见证人
*/
func GetWitnessInfo(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	winfo := WitnessInfo{}
	winfo.IsCandidate, winfo.IsBackup, winfo.IsKickOut, winfo.Addr, winfo.Value = mining.GetWitnessStatus()

	addr := keystore.GetCoinbase()
	winfo.Payload = mining.FindWitnessName(addr)

	res, err = model.Tojson(winfo)
	return
}

/*
	见证人
*/
type WitnessVO struct {
	Addr            string //见证人地址
	Payload         string //
	Score           uint64 //押金
	Vote            uint64 //投票者押金
	CreateBlockTime int64  //预计出块时间
}

/*
	获取候选见证人列表
*/
func GetCandidateList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	wbg := mining.GetWitnessListSort()

	wvos := make([]WitnessVO, 0)
	for _, one := range append(wbg.Witnesses, wbg.WitnessBackup...) {

		name := mining.FindWitnessName(*one.Addr)

		wvo := WitnessVO{
			Addr:            one.Addr.B58String(), //见证人地址
			Payload:         name,                 //
			Score:           one.Score,            //押金
			Vote:            one.VoteNum,          //      voteValue,            //投票票数
			CreateBlockTime: one.CreateBlockTime,  //预计出块时间
		}
		wvos = append(wvos, wvo)
	}

	res, err = model.Tojson(wvos)
	return
}

/*
	获取社区节点列表
*/
func GetCommunityList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vss := mining.GetCommunityListSort()
	res, err = model.Tojson(vss)
	return
}

/*
	投票详情
*/
type VoteInfoVO struct {
	Txid        string //交易id
	WitnessAddr string //见证人地址
	Value       uint64 //投票数量
	Height      uint64 //区块高度
	AddrSelf    string //自己投票的地址
	Payload     string //
}

type Vinfos struct {
	infos []VoteInfoVO
}

func (this *Vinfos) Len() int {
	return len(this.infos)
}

func (this *Vinfos) Less(i, j int) bool {
	if this.infos[i].Height < this.infos[j].Height {
		return false
	} else {
		return true
	}
}

func (this *Vinfos) Swap(i, j int) {
	this.infos[i], this.infos[j] = this.infos[j], this.infos[i]
}

/*
	获得自己给哪些见证人投过票的列表
*/
func GetVoteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	vtItr, ok := rj.Get("votetype")
	if !ok {
		res, err = model.Errcode(model.NoField, "votetype")
		return
	}
	voteType := uint16(vtItr.(float64))

	balances := mining.GetVoteList()
	vinfos := Vinfos{
		infos: make([]VoteInfoVO, 0),
	}
	for _, one := range balances {
		// fmt.Println("查看", one)
		one.Txs.Range(func(k, v interface{}) bool {
			ti := v.(*mining.TxItem)
			if ti.VoteType != voteType {
				return true
			}
			viVO := VoteInfoVO{
				Txid:        hex.EncodeToString(ti.Txid), //
				WitnessAddr: one.Addr.B58String(),        //见证人地址
				Value:       ti.Value,                    //投票数量
				Height:      ti.Height,                   //区块高度
				AddrSelf:    ti.Addr.B58String(),         //自己投票的地址
			}
			viVO.Payload = mining.FindWitnessName(*ti.Addr)

			vinfos.infos = append(vinfos.infos, viVO)
			return true
		})
	}

	// fmt.Println("排序前查看投票", vinfos)
	sort.Sort(&vinfos)
	// fmt.Println("排序后查看投票", vinfos)
	res, err = model.Tojson(vinfos.infos)
	return
}

// /*
// 	获得自己给哪些社区节点投过票的列表
// */
// func GetCommunityVoteList(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
// 	balances := mining.GetVoteList()
// 	vinfos := Vinfos{
// 		infos: make([]VoteInfoVO, 0),
// 	}
// 	for _, one := range balances {
// 		// fmt.Println("查看", one)
// 		one.Txs.Range(func(k, v interface{}) bool {
// 			ti := v.(*mining.TxItem)
// 			viVO := VoteInfoVO{
// 				Txid:        hex.EncodeToString(ti.Txid), //
// 				WitnessAddr: one.Addr.B58String(),        //见证人地址
// 				Value:       ti.Value,                    //投票数量
// 				Height:      ti.Height,                   //区块高度
// 				AddrSelf:    "",                          //自己投票的地址
// 			}
// 			vinfos.infos = append(vinfos.infos, viVO)
// 			return true
// 		})
// 	}

// 	// fmt.Println("排序前查看投票", vinfos)
// 	sort.Sort(&vinfos)
// 	// fmt.Println("排序后查看投票", vinfos)
// 	res, err = model.Tojson(vinfos.infos)
// 	return
// }

/*
	查询一笔交易是否成功
	@return    int    1=未确认；2=成功；3=失败；
*/
func FindTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	txItr, ok := rj.Get("txid")
	if !ok {
		res, err = model.Errcode(model.NoField, "txid")
		return
	}
	txidStr := txItr.(string)

	txid, err := hex.DecodeString(txidStr)
	if err != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	lockheightItr, ok := rj.Get("lockheight")
	if !ok {
		res, err = model.Errcode(model.NoField, "lockheight")
		return
	}
	lockheight := uint64(lockheightItr.(float64))

	result := mining.FindTx(txid, lockheight)

	res, err = model.Tojson(result)
	return
}

type BlockHeadVO struct {
	Hash              string   //区块头hash
	Height            uint64   //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
	GroupHeight       uint64   //矿工组高度
	Previousblockhash string   //上一个区块头hash
	Nextblockhash     string   //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
	NTx               uint64   //交易数量
	MerkleRoot        string   //交易默克尔树根hash
	Tx                []string //本区块包含的交易id
	Time              int64    //出块时间，unix时间戳
	Witness           string   //此块见证人地址
	Sign              string   //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
}

/*
	通过区块高度查询一个区块详细信息
*/
func FindBlock(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	heightItr, ok := rj.Get("height")
	if !ok {
		res, err = model.Errcode(model.NoField, "height")
		return
	}
	height := uint64(heightItr.(float64))

	bh := mining.FindBlockHead(height)
	if bh == nil {
		res, err = model.Errcode(model.NotExist)
		return
	}

	txs := make([]string, 0)
	for _, one := range bh.Tx {
		txs = append(txs, hex.EncodeToString(one))
	}

	bhvo := BlockHeadVO{
		Hash:              hex.EncodeToString(bh.Hash),              //区块头hash
		Height:            bh.Height,                                //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
		GroupHeight:       bh.GroupHeight,                           //矿工组高度
		Previousblockhash: hex.EncodeToString(bh.Previousblockhash), //上一个区块头hash
		Nextblockhash:     hex.EncodeToString(bh.Nextblockhash),     //下一个区块头hash,可能有多个分叉，但是要保证排在第一的链是最长链
		NTx:               bh.NTx,                                   //交易数量
		MerkleRoot:        hex.EncodeToString(bh.MerkleRoot),        //交易默克尔树根hash
		Tx:                txs,                                      //本区块包含的交易id
		Time:              bh.Time,                                  //出块时间，unix时间戳
		Witness:           bh.Witness.B58String(),                   //此块见证人地址
		Sign:              hex.EncodeToString(bh.Sign),              //见证人出块时，见证人对块签名，以证明本块是指定见证人出块。
	}

	res, err = model.Tojson(bhvo)
	return

}

/*
	查询社区奖励
*/
func GetCommunityReward(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
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

	// engine.Log.Info("地址 "+addrStr+" d%", mining.GetAddrState(*addr))

	if mining.GetAddrState(*addr) != 2 {
		//本地址不是社区节点
		res, err = model.Errcode(model.RuleField, "address")
		return
	}

	// engine.Log.Info("11111111111111111111")

	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(addr)
	if err != nil {
		// engine.Log.Info("222222222222222222222")
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//有未分配完的奖励
		community := ns.Reward / 10
		light := ns.Reward - community
		rewardTotal := mining.RewardTotal{
			CommunityReward: community,                           //社区节点奖励
			LightReward:     light,                               //轻节点奖励
			StartHeight:     ns.StartHeight,                      //
			Height:          ns.EndHeight,                        //最新区块高度
			IsGrant:         false,                               //是否可以分发奖励，24小时候才可以分发奖励
			AllLight:        ns.LightNum,                         //所有轻节点数量
			RewardLight:     ns.LightNum - uint64(len(*notSend)), //已经奖励的轻节点数量
		}

		res, err = model.Tojson(rewardTotal)
		return
	}

	startHeight := uint64(config.Mining_block_start_height)
	if ns != nil {
		startHeight = ns.EndHeight + 1
	}
	// engine.Log.Info("333333333333333333")
	//奖励都分配完了，查询新的奖励
	rt, _, err := mining.GetRewardCount(addr, startHeight, 0)
	if err != nil {
		// engine.Log.Info("444444444444444")
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson(rt)
	return
}

/*
	给轻节点分发奖励
*/
func SendCommunityReward(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		return
	}
	addrStr := addrItr.(string)
	if addrStr != "" {
		addrMul := crypto.AddressFromB58String(addrStr)
		addr = &addrMul
	}
	var dst crypto.AddressCoin
	if addrStr != "" {
		dst = crypto.AddressFromB58String(addrStr)
		if !crypto.ValidAddr(config.AddrPre, dst) {
			res, err = model.Errcode(model.ContentIncorrectFormat, "address")
			return
		}
	}

	// engine.Log.Info("地址 "+addrStr+" d%", mining.GetAddrState(*addr))

	if mining.GetAddrState(*addr) != 2 {
		//本地址不是社区节点
		res, err = model.Errcode(model.RuleField, "address")
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

	// //查询余额是否足够
	// if amount > mining.GetBalance() {
	// 	res, err = model.Errcode(model.Nomarl, "余额不足")
	// 	return
	// }

	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(addr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//有未分配完的奖励
		//查询社区节点公钥
		puk, ok := keystore.GetPukByAddr(dst)
		if ok {
			res, err = model.Errcode(model.Nomarl, config.ERROR_public_key_not_exist.Error())
			return
		}

		cs := mining.NewCommunitySign(puk, ns.StartHeight, ns.EndHeight)

		err = mining.DistributionReward(notSend, gas, pwd, *cs)
		if err != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}

		res, err = model.Tojson(model.Success)
		return
	}

	startHeight := uint64(1)
	if ns != nil {
		startHeight = ns.EndHeight + 1
	}

	//奖励都分配完了，查询新的奖励
	rt, notSend, err := mining.GetRewardCount(addr, startHeight, 0)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	if !rt.IsGrant {
		now := time.Now().Unix()
		blockNum := (config.Mining_community_reward_time / config.Mining_block_time) - (rt.Height - rt.StartHeight)
		wait := blockNum * config.Mining_block_time
		futuer := time.Unix(now+int64(wait), 0)

		res, err = model.Errcode(model.Nomarl, "请在"+futuer.Format("2006-01-02 15:04:05")+"之后再分配！")
		return
	}

	//创建快照
	err = mining.CreateRewardCount(addrStr, rt, *notSend)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}

	//查询社区节点公钥
	puk, ok := keystore.GetPukByAddr(dst)
	if ok {
		res, err = model.Errcode(model.Nomarl, config.ERROR_public_key_not_exist.Error())
		return
	}

	cs := mining.NewCommunitySign(puk, rt.StartHeight, rt.Height)

	err = mining.DistributionReward(notSend, gas, pwd, *cs)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson(model.Success)
	return
}
