package rpc

import (
	"mandela/chain_witness_vote/db"
	"mandela/chain_witness_vote/mining"
	"mandela/chain_witness_vote/mining/name"
	"mandela/chain_witness_vote/mining/token/payment"
	"mandela/chain_witness_vote/mining/token/publish"
	"mandela/cloud_reward/server"
	"mandela/config"
	"mandela/core/engine"
	"mandela/core/keystore"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"mandela/rpc/model"
	"mandela/sqlite3_db"
	"bytes"
	"encoding/hex"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-xorm/xorm"
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
	Payload    string   //
}

/*
	历史记录
*/

func GetTransactionHistoty(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	if mining.GetLongChain() == nil {
		res, err = model.Tojson(model.Success)
		return
	}
	//如果是见证人，需要时间间隔控制
	if ok, _, _, _, _ := mining.GetWitnessStatus(); ok {
		utils.SetTimeToken(config.TIMETOKEN_GetTransactionHistoty, time.Second*5)
		if allow := utils.GetTimeToken(config.TIMETOKEN_GetTransactionHistoty, false); !allow {
			res, err = model.Tojson(model.Success)
			return
		}
	}

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
	hivos := make([]HistoryItemVO, 0)

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Tojson(hivos)
		return
	}
	his := chain.GetHistoryBalance(startId, total)

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
			Payload:    string(one.Payload),          //
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
	查询数据库中key对应的value
*/
func MergeTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	switchItr, ok := rj.Get("switch")
	if !ok {
		res, err = model.Errcode(model.NoField, "switch")
		return
	}
	isOpen := switchItr.(bool)

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

	if pwd != config.Wallet_keystore_default_pwd {
		res, err = model.Tojson(model.FailPwd)
		return
	}

	var unifieaddr *crypto.AddressCoin
	unifieaddrItr, ok := rj.Get("unifieaddr") //归集地址
	if ok {
		addrStr := unifieaddrItr.(string)
		if addrStr != "" {
			addrMul := crypto.AddressFromB58String(addrStr)
			if !crypto.ValidAddr(config.AddrPre, addrMul) {
				res, err = model.Errcode(model.ContentIncorrectFormat, "unifieaddr")
				// engine.Log.Info("11111111111111111111")
				return
			}
			unifieaddr = &addrMul
		}
	}

	totalMax := uint64(0)
	totalMaxItr, ok := rj.Get("totalmax")
	if ok {
		totalMax = uint64(totalMaxItr.(float64))
	}

	if isOpen {
		mining.SwitchOpenMergeTx(pwd, gas, unifieaddr, totalMax)
	} else {
		mining.SwitchCloseMergeTx(pwd)
	}

	res, err = model.Tojson(model.Success)
	return
}

/*
	查询节点总数
*/
func GetNodeTotal(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	outMap := make(map[string]interface{})
	//获取存储超级节点地址
	nameinfo := name.FindName(config.Name_store)
	if nameinfo == nil {
		//域名不存在
		engine.Log.Debug("Domain name does not exist")
		outMap["total_addr"] = 0  //全网节点总数
		outMap["total_space"] = 0 //全网存储空间总数
		res, err = model.Tojson(outMap)
		return
	}
	// nets := client.GetCloudPeer()
	// if nets == nil {
	// 	return
	// }
	//判断自己是否在超级节点地址里
	have := false
	for _, one := range nameinfo.NetIds {
		if bytes.Equal(nodeStore.NodeSelf.IdInfo.Id, one) {
			have = true
			break
		}
	}
	//没有在列表里
	if !have {
		engine.Log.Debug("You are not in the super node address")
		outMap["total_addr"] = config.GetSpaceTotalAddr() //全网节点总数
		outMap["total_space"] = config.GetSpaceTotal()    //全网存储空间总数
		res, err = model.Tojson(outMap)
		return
	}
	// peers := server.CountStorePeers()
	// total := len(*peers)
	outMap["total_addr"], outMap["total_space"] = server.CountStoreTotal()
	// = config.GetSpaceTotalAddr() //全网节点总数
	// outMap["total_space"] = config.GetSpaceTotal()    //全网存储空间总数
	res, err = model.Tojson(outMap)
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

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Tojson(winfo)
		return
	}
	var witnessAddr crypto.AddressCoin
	winfo.IsCandidate, winfo.IsBackup, winfo.IsKickOut, witnessAddr, winfo.Value = mining.GetWitnessStatus()
	winfo.Addr = witnessAddr.B58String()

	addr := keystore.GetCoinbase()
	winfo.Payload = mining.FindWitnessName(addr.Addr)

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
	@voteType    int    投票类型，1=给见证人投票；2=给社区节点投票；3=轻节点押金；
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
	// sort.Sort(&vinfos)
	sort.Stable(&vinfos)
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

	outMap := make(map[string]interface{})

	txItr, code := mining.FindTxJsonVo(txid)
	outMap["txinfo"] = txItr
	outMap["upchaincode"] = code
	res, err = model.Tojson(outMap)
	return
}

/*
	查询数据库中key对应的value
*/
func GetValueForKey(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	keyItr, ok := rj.Get("key")
	if !ok {
		res, err = model.Errcode(model.NoField, "key")
		return
	}
	keyBs, e := hex.DecodeString(keyItr.(string))
	if e != nil {
		res, err = model.Errcode(model.TypeWrong, "key")
		return
	}
	value, e := db.LevelDB.Find(keyBs)
	if e != nil {
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	res, err = model.Tojson(value)
	return
}

type BlockHeadVO struct {
	Hash              string   //区块头hash
	Height            uint64   //区块高度(每秒产生一个块高度，uint64容量也足够使用上千亿年)
	GroupHeight       uint64   //矿工组高度
	GroupHeightGrowth uint64   //组高度增长量。默认0为自动计算增长量（兼容之前的区块）,最少增量为1
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
	bh := mining.LoadBlockHeadByHeight(height)
	// bh := mining.FindBlockHead(height)
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

//
var findCommunityStartHeightByAddrOnceLock = new(sync.Mutex)
var findCommunityStartHeightByAddrOnce = make(map[string]bool) // new(sync.Once)

/*
	查找一个地址成为社区节点的开始高度
*/
func findCommunityStartHeightByAddr(addr crypto.AddressCoin) {
	// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
	ok := false
	findCommunityStartHeightByAddrOnceLock.Lock()
	_, ok = findCommunityStartHeightByAddrOnce[utils.Bytes2string(addr)]
	findCommunityStartHeightByAddrOnce[utils.Bytes2string(addr)] = false
	findCommunityStartHeightByAddrOnceLock.Unlock()
	if ok {
		// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
		return
	}
	// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
	utils.Go(func() {
		bhHash, err := db.LevelTempDB.Find(mining.BuildCommunityAddrStartHeight(addr))
		if err != nil {
			engine.Log.Error("this addr not community:%s error:%s", addr.B58String(), err.Error())
			return
		}
		var sn *sqlite3_db.SnapshotReward
		//判断数据库是否有快照记录
		sn, _, err = mining.FindNotSendReward(&addr)
		if err != nil && err.Error() != xorm.ErrNotExist.Error() {
			engine.Log.Error("querying database Error %s", err.Error())
			return
		}
		//有记录，就不再恢复历史记录了
		if sn != nil {
			// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
			return
		}
		//建立多个快照，后面一次性保存
		snapshots := make([]sqlite3_db.SnapshotReward, 0)

		var bhvo *mining.BlockHeadVO
		var txItr mining.TxItr
		// var err error
		var addrTx crypto.AddressCoin
		var ok bool
		var cs *mining.CommunitySign
		var have bool
		for {
			if bhHash == nil || len(*bhHash) <= 0 {
				// engine.Log.Info("findCommunityStartHeightByAddr 22222222222222")
				break
			}
			bhvo, err = mining.LoadBlockHeadVOByHash(bhHash)
			if err != nil {
				engine.Log.Error("findCommunityStartHeightByAddr load blockhead error:%s", err.Error())
				return
			}
			// engine.Log.Info("findCommunityStartHeightByAddr Community start count block height:%d", bhvo.BH.Height)
			bhHash = &bhvo.BH.Nextblockhash
			if len(snapshots) <= 0 {
				//创建一个空奖励快照
				snapshotsOne := sqlite3_db.SnapshotReward{
					Addr:        addr,           //社区节点地址
					StartHeight: bhvo.BH.Height, //快照开始高度
					EndHeight:   bhvo.BH.Height, //快照结束高度
					Reward:      0,              //此快照的总共奖励
					LightNum:    0,              //奖励的轻节点数量
				}
				snapshots = append(snapshots, snapshotsOne)
			}
			for _, txItr = range bhvo.Txs {
				//判断交易类型
				if txItr.Class() != config.Wallet_tx_type_pay {
					continue
				}
				//检查签名
				addrTx, ok, cs = mining.CheckPayload(txItr)
				if !ok {
					//签名不正确
					continue
				}
				//判断地址是否属于自己
				_, ok = keystore.FindAddress(addrTx)
				if !ok {
					//签名者地址不属于自己
					continue
				}
				//判断有没有这个快照
				have = false
				for _, one := range snapshots {
					if bytes.Equal(addr, one.Addr) && one.StartHeight == cs.StartHeight && one.EndHeight == cs.EndHeight {
						have = true
						break
					}
				}
				if have {
					continue
				}
				//没有这个快照就创建
				snapshotsOne := sqlite3_db.SnapshotReward{
					Addr:        addr,           //社区节点地址
					StartHeight: cs.StartHeight, //快照开始高度
					EndHeight:   cs.EndHeight,   //快照结束高度
					Reward:      0,              //此快照的总共奖励
					LightNum:    0,              //奖励的轻节点数量
				}

				snapshots = append(snapshots, snapshotsOne)
			}
		}
		// engine.Log.Info("findCommunityStartHeightByAddr Community end count block!")
		//开始保存快照
		for i, _ := range snapshots {
			one := snapshots[i]
			err = one.Add(&one)
			if err != nil {
				engine.Log.Info(err.Error())
			}
		}
		// engine.Log.Info("findCommunityStartHeightByAddr Community finish count block!")
	})

}

/*
	查询社区奖励
*/
func GetCommunityReward(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// engine.Log.Info("11111111111111111111")
	var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //押金冻结的地址
	if !ok {
		// engine.Log.Info("11111111111111111111")
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
			// engine.Log.Info("11111111111111111111")
			return
		}
	}
	// engine.Log.Info("11111111111111111111")
	// engine.Log.Info("地址 "+addrStr+" d%", mining.GetAddrState(*addr))

	if mining.GetAddrState(*addr) != 2 {
		//本地址不是社区节点
		res, err = model.Errcode(model.RuleField, "address")
		// engine.Log.Info("11111111111111111111")
		return
	}

	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Errcode(model.Nomarl, "The chain end is not synchronized")
		// engine.Log.Info("11111111111111111111")
		return
	}
	currentHeight := chain.GetCurrentBlock()

	// engine.Log.Info("11111111111111111111")
	// engine.Log.Info("11111111111111111111")
	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(addr)
	if err != nil {
		// engine.Log.Info("222222222222222222222")
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("11111111111111111111")
		return
	}
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//查询代上链的奖励是否已经上链
		*notSend, _ = checkTxUpchain(*notSend, currentHeight)
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
		// engine.Log.Info("11111111111111111111")
		res, err = model.Tojson(rewardTotal)
		return
	}
	if ns == nil {
		//需要加载以前的奖励快照
		findCommunityStartHeightByAddr(*addr)
		res, err = model.Errcode(model.Nomarl, "load reward history")
		// engine.Log.Info("11111111111111111111")
		return
	}
	// engine.Log.Info("11111111111111111111")
	startHeight := ns.EndHeight + 1

	// engine.Log.Info("333333333333333333")
	//奖励都分配完了，查询新的奖励
	rt, _, err := mining.GetRewardCount(addr, startHeight, 0)
	if err != nil {
		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
			res, err = model.Errcode(model.RewardCountSync, err.Error())
			return
		}
		// engine.Log.Info("444444444444444")
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson(rt)
	// engine.Log.Info("11111111111111111111")
	return
}

/*
	给轻节点分发奖励
*/
func SendCommunityReward(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	// engine.Log.Info("22222222222222")
	// var addr *crypto.AddressCoin
	addrItr, ok := rj.Get("address") //社区节点地址
	if !ok {
		res, err = model.Errcode(model.NoField, "address")
		// engine.Log.Info("22222222222222")
		return
	}
	addrStr := addrItr.(string)
	if addrStr == "" {
		res, err = model.Errcode(model.ContentIncorrectFormat, "address")
		// engine.Log.Info("22222222222222")
		return
	}
	addr := crypto.AddressFromB58String(addrStr)
	if !crypto.ValidAddr(config.AddrPre, addr) {
		res, err = model.Errcode(model.ContentIncorrectFormat, "address")
		// engine.Log.Info("22222222222222")
		return
	}

	if mining.GetAddrState(addr) != 2 {
		//本地址不是社区节点
		res, err = model.Errcode(model.RuleField, "address")
		// engine.Log.Info("22222222222222")
		return
	}
	// engine.Log.Info("22222222222222")
	//查询社区节点公钥
	puk, ok := keystore.GetPukByAddr(addr)
	if !ok {
		res, err = model.Errcode(model.Nomarl, config.ERROR_public_key_not_exist.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	gasItr, ok := rj.Get("gas")
	if !ok {
		res, err = model.Errcode(model.NoField, "gas")
		// engine.Log.Info("22222222222222")
		return
	}
	gas := uint64(gasItr.(float64))
	// engine.Log.Info("22222222222222")
	pwdItr, ok := rj.Get("pwd")
	if !ok {
		res, err = model.Errcode(model.NoField, "pwd")
		// engine.Log.Info("22222222222222")
		return
	}
	pwd := pwdItr.(string)

	startHeightItr, ok := rj.Get("startheight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startheight")
		// engine.Log.Info("22222222222222")
		return
	}
	startHeight := uint64(startHeightItr.(float64))

	endheightItr, ok := rj.Get("endheight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endheight")
		// engine.Log.Info("22222222222222")
		return
	}
	endheight := uint64(endheightItr.(float64))
	chain := mining.GetLongChain()
	if chain == nil {
		res, err = model.Errcode(model.Nomarl, "The chain end is not synchronized")
		// engine.Log.Info("22222222222222")
		return
	}
	currentHeight := chain.GetCurrentBlock()
	//查询未分配的奖励
	ns, notSend, err := mining.FindNotSendReward(&addr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	// engine.Log.Info("22222222222222")
	if ns != nil && notSend != nil && len(*notSend) > 0 {
		//查询待上链的奖励是否已经上链
		*notSend, ok = checkTxUpchain(*notSend, currentHeight)
		if ok {
			//有未上链的奖励
			res, err = model.Errcode(model.Nomarl, "There are rewards that are not linked")
			// engine.Log.Info("22222222222222")
			return
		}
		cs := mining.NewCommunitySign(puk, ns.StartHeight, ns.EndHeight)
		//有未分配完的奖励，继续分配
		err = mining.DistributionReward(notSend, gas, pwd, cs, currentHeight)
		if err != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			// engine.Log.Info("22222222222222")
			return
		}
		res, err = model.Tojson(model.Success)
		// engine.Log.Info("22222222222222")
		return
	}
	//检查奖励开始高度，避免重复奖励
	if ns != nil && startHeight <= ns.EndHeight {
		res, err = model.Errcode(model.Nomarl, "Repeat reward")
		// engine.Log.Info("22222222222222")
		return
	}
	// engine.Log.Info("22222222222222")
	//奖励都分配完了，查询新的奖励
	rt, notSend, err := mining.GetRewardCount(&addr, startHeight, endheight)
	if err != nil {
		if err.Error() == config.ERROR_get_reward_count_sync.Error() {
			res, err = model.Errcode(model.RewardCountSync, err.Error())
			return
		}
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	if !rt.IsGrant {
		now := time.Now().Unix()
		blockNum := (config.Mining_community_reward_time / config.Mining_block_time) - (rt.Height - rt.StartHeight)
		wait := blockNum * config.Mining_block_time
		futuer := time.Unix(now+int64(wait), 0)
		// engine.Log.Info("22222222222222")
		res, err = model.Errcode(model.Nomarl, "Please distribute the reward after "+futuer.Format("2006-01-02 15:04:05"))
		return
	}
	//创建快照
	err = mining.CreateRewardCount(addr, rt, *notSend)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		// engine.Log.Info("22222222222222")
		return
	}
	//创建快照成功后，删除缓存
	mining.CleanRewardCountProcessMap(&addr)

	// engine.Log.Info("22222222222222")
	// cs := mining.NewCommunitySign(puk, rt.StartHeight, rt.Height)
	// err = mining.DistributionReward(notSend, gas, pwd, cs, currentHeight)
	// if err != nil {
	// 	res, err = model.Errcode(model.Nomarl, err.Error())
	// 	return
	// }
	res, err = model.Tojson(model.Success)
	// engine.Log.Info("22222222222222")
	return
}

/*
	检查交易是否上链，未上链，超过上链高度，则取消上链。
	已上链的则修改数据库
*/
func checkTxUpchain(notSend []sqlite3_db.RewardLight, currentHeight uint64) ([]sqlite3_db.RewardLight, bool) {

	txidUpchain := make(map[string]int)                     //保存已经上链的交易
	txidNotUpchain := make(map[string]int)                  //保存未上链的交易
	resultUpchain := make([]sqlite3_db.RewardLight, 0)      //保存需要修改为上链的奖励记录
	resultUnLockHeight := make([]sqlite3_db.RewardLight, 0) //保存需要回滚的奖励记录
	resultReward := make([]sqlite3_db.RewardLight, 0)       //返回的未上链的结果
	haveNotUpchain := false                                 //保存是否存在未上链的奖励
	for i, _ := range notSend {
		one := notSend[i]
		if one.Txid == nil {
			resultReward = append(resultReward, one)
			continue
		}
		//查询交易是否上链
		//先查询缓存
		_, ok := txidUpchain[utils.Bytes2string(one.Txid)]
		if ok {
			//上链了
			resultUpchain = append(resultUpchain, one)
			continue
		}
		_, ok = txidNotUpchain[utils.Bytes2string(one.Txid)]
		if ok {
			//未上链，判断是否超过冻结高度
			if one.LockHeight < currentHeight {
				//回滚
				resultUnLockHeight = append(resultUnLockHeight, one)
			} else {
				//等待
				resultReward = append(resultReward, one)
				haveNotUpchain = true
			}
			continue
		}

		//缓存没有，只有去数据库查询
		txItr, err := mining.LoadTxBase(one.Txid)
		blockhash, berr := db.GetTxToBlockHash(&one.Txid)
		// if berr != nil || blockhash == nil {
		// 	return config.ERROR_tx_format_fail
		// }
		// txItr, err := mining.FindTxBase(one.Txid)
		if err != nil || txItr == nil || berr != nil || blockhash == nil {
			// if err != nil || txItr == nil || txItr.GetBlockHash() == nil {
			txidNotUpchain[utils.Bytes2string(one.Txid)] = 0
			//未上链，判断是否超过冻结高度
			if one.LockHeight < currentHeight {
				//回滚
				resultUnLockHeight = append(resultUnLockHeight, one)
			} else {
				//等待
				resultReward = append(resultReward, one)
				haveNotUpchain = true
			}
		} else {
			//上链了，修改状态
			txidUpchain[utils.Bytes2string(one.Txid)] = 0
			resultUpchain = append(resultUpchain, one)
		}
	}
	//奖励回滚
	if len(resultUnLockHeight) > 0 {
		ids := make([]uint64, 0)
		for _, one := range resultUnLockHeight {
			ids = append(ids, one.Id)
		}
		err := new(sqlite3_db.RewardLight).RemoveTxid(ids)
		if err != nil {
			engine.Log.Error(err.Error())
		}
	}
	//奖励修改为已经上链
	if len(resultUpchain) > 0 {
		var err error
		for _, one := range resultUpchain {
			err = one.UpdateDistribution(one.Id, one.Reward)
			if err != nil {
				engine.Log.Error(err.Error())
			}
		}
	}
	return resultReward, haveNotUpchain
}

/*
	检查是否有交易已经生成，但是未上链的奖励
*/
// func checkHaveNotUpchain(notSend []sqlite3_db.Reward)bool{

// }

/*
	发布一个token
*/
func TokenPublish(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			_, ok := keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(model.ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok = rj.Get("address") //押金冻结的地址
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

	amount := uint64(0)

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	name := ""
	nameItr, ok := rj.Get("name") //token的名称
	if ok {
		name = nameItr.(string)
	}
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if strings.Contains(name, ".") || strings.Contains(name, " ") {
		res, err = model.Errcode(5002, "name")
		return
	}

	//symbol token单位
	symbol := ""
	symbolItr, ok := rj.Get("symbol")
	if ok {
		symbol = symbolItr.(string)
	}
	//对名称做限制，不能和万维网域名重复，名称不能带"."字符。
	if strings.Contains(symbol, ".") || strings.Contains(symbol, " ") {
		res, err = model.Errcode(5002, "symbol")
		return
	}

	//token发行总量
	supplyItr, ok := rj.Get("supply") //token发行总量
	if !ok {
		res, err = model.Errcode(5002, "supply")
		return
	}
	supply := uint64(supplyItr.(float64))
	if supply < config.Witness_token_supply_min {
		res, err = model.Errcode(model.Nomarl, config.ERROR_token_min_fail.Error())
		return
	}

	var owner crypto.AddressCoin
	ownerItr, ok := rj.Get("owner") //押金冻结的地址
	if ok {
		ownerStr := ownerItr.(string)
		if ownerStr != "" {
			ownerMul := crypto.AddressFromB58String(ownerStr)
			owner = ownerMul
		}

		if ownerStr != "" {
			dst := crypto.AddressFromB58String(ownerStr)
			if !crypto.ValidAddr(config.AddrPre, dst) {
				res, err = model.Errcode(model.ContentIncorrectFormat, "owner")
				return
			}
		}
	}

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	//收款地址参数

	// @addr    *crypto.AddressCoin    收款地址
	// @amount    uint64    转账金额
	// @gas    uint64    手续费
	// @pwd    string    支付密码
	// @name    string    Token名称全称
	// @symbol    string    Token单位，符号
	// @supply    uint64    发行总量
	// @owner    crypto.AddressCoin    所有者
	txItr, e := publish.PublishToken(&src, addr, amount, gas, frozenHeight, pwd, comment, name, symbol, supply, owner)
	if e == nil {
		// res, err = model.Tojson("success")
		result, e := utils.ChangeMap(txItr)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txItr.GetHash())

		res, err = model.Tojson(result)
		return
	}
	if e.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if e.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(model.NotEnough)
		return
	}
	if e.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, e.Error())

	return
}

/*
	使用token支付
*/
func TokenPay(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var src crypto.AddressCoin
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcaddr := addrItr.(string)
		if srcaddr != "" {
			src = crypto.AddressFromB58String(srcaddr)
			_, ok := keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(model.ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	var addr *crypto.AddressCoin
	addrItr, ok = rj.Get("address") //押金冻结的地址
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

	gasItr, ok := rj.Get("gas") //手续费
	if !ok {
		res, err = model.Errcode(5002, "gas")
		return
	}
	gas := uint64(gasItr.(float64))

	pwdItr, ok := rj.Get("pwd") //支付密码
	if !ok {
		res, err = model.Errcode(5002, "pwd")
		return
	}
	pwd := pwdItr.(string)

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	// var owner crypto.AddressCoin
	txidItr, ok := rj.Get("txid") //发布token的交易id
	if !ok {
		res, err = model.Errcode(5002, "txid")
		return
	}
	txid := txidItr.(string)

	frozenHeight := uint64(0)
	frozenHeightItr, ok := rj.Get("frozen_height")
	if ok {
		frozenHeight = uint64(frozenHeightItr.(float64))
	}

	//收款地址参数

	// @addr    *crypto.AddressCoin    收款地址
	// @amount    uint64    转账金额
	// @gas    uint64    手续费
	// @pwd    string    支付密码
	// @txid    string    发布token的交易id
	txItr, e := payment.TokenPay(&src, addr, amount, gas, frozenHeight, pwd, comment, txid)
	if e == nil {
		result, e := utils.ChangeMap(txItr)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txItr.GetHash())

		res, err = model.Tojson(result)
		return

		// res, err = model.Tojson("success")
		// return
	}
	if e.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if e.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(model.NotEnough)
		return
	}
	if e.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, e.Error())

	return
}

/*
	使用token多人转账
*/
func TokenPayMore(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	var src crypto.AddressCoin
	srcAddrStr := ""
	addrItr, ok := rj.Get("srcaddress")
	if ok {
		srcAddrStr = addrItr.(string)
		if srcAddrStr != "" {
			src = crypto.AddressFromB58String(srcAddrStr)
			//判断地址前缀是否正确
			// if !crypto.ValidAddr(config.AddrPre, src) {
			// 	res, err = model.Errcode(model.ContentIncorrectFormat, "srcaddress")
			// 	return
			// }
			//判断地址是否包含在keystone里面
			// if !keystore.FindAddress(src) {
			// 	res, err = model.Errcode(model.ContentIncorrectFormat, "srcaddress")
			// 	return
			// }
			_, ok := keystore.FindAddress(src)
			if !ok {
				res, err = model.Errcode(model.ContentIncorrectFormat, "srcaddress")
				return
			}
		}
	}

	addrItr, ok = rj.Get("addresses")
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
			Address:      dst,              //转账地址
			Amount:       one.Amount,       //转账金额
			FrozenHeight: one.FrozenHeight, //
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

	comment := ""
	commentItr, ok := rj.Get("comment")
	if ok && rj.VerifyType("comment", "string") {
		comment = commentItr.(string)
	}

	// var owner crypto.AddressCoin
	txidItr, ok := rj.Get("txid") //发布token的交易id
	if !ok {
		res, err = model.Errcode(5002, "txid")
		return
	}
	txid, e := hex.DecodeString(txidItr.(string))
	if e != nil {
		res, err = model.Errcode(model.TypeWrong, "txid")
		return
	}

	//收款地址参数

	// @addr    *crypto.AddressCoin    收款地址
	// @amount    uint64    转账金额
	// @gas    uint64    手续费
	// @pwd    string    支付密码
	// @txid    string    发布token的交易id
	txItr, e := payment.TokenPayMore(nil, src, addr, gas, pwd, comment, txid)
	if e == nil {
		result, e := utils.ChangeMap(txItr)
		if e != nil {
			res, err = model.Errcode(model.Nomarl, err.Error())
			return
		}
		result["hash"] = hex.EncodeToString(*txItr.GetHash())

		res, err = model.Tojson(result)
		return

		// res, err = model.Tojson("success")
		// return
	}
	if e.Error() == config.ERROR_password_fail.Error() {
		res, err = model.Errcode(model.FailPwd)
		return
	}
	if e.Error() == config.ERROR_not_enough.Error() {
		res, err = model.Errcode(model.NotEnough)
		return
	}
	if e.Error() == config.ERROR_name_exist.Error() {
		res, err = model.Errcode(model.Exist)
		return
	}
	res, err = model.Errcode(model.Nomarl, e.Error())

	return
}

/*
	将组装好并签名的交易上链
*/
func PushTx(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {

	txJsonItr, ok := rj.Get("tx")
	if !ok {
		res, err = model.Errcode(model.NoField, "tx")
		return
	}

	txjson := txJsonItr.(string)
	txjsonBs := []byte(txjson)

	txItr, err := mining.ParseTxBaseProto(0, &txjsonBs)
	// txItr, err := mining.ParseTxBase(0, &txjsonBs)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	engine.Log.Info("rpc transaction received %s", hex.EncodeToString(*txItr.GetHash()))
	if e := txItr.Check(); e != nil {
		engine.Log.Info("transaction check fail:%s", err.Error())
		res, err = model.Errcode(model.Nomarl, e.Error())
		return
	}
	err = mining.AddTx(txItr)
	if err != nil {
		res, err = model.Errcode(model.Nomarl, err.Error())
		return
	}
	res, err = model.Tojson(model.Success)
	return
}

/*
	通过一定范围的区块高度查询多个区块详细信息
*/
func FindBlockRange(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := uint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	endHeight := uint64(endHeightItr.(float64))

	if endHeight < startHeight {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}

	//待返回的区块
	bhvos := make([]mining.BlockHeadVO, 0, endHeight-startHeight+1)

	for i := startHeight; i <= endHeight; i++ {

		bhvo := mining.BlockHeadVO{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}
		bhvo.BH = bh
		bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))

		for _, one := range bh.Tx {
			txItr, e := mining.LoadTxBase(one)
			// txItr, e := mining.FindTxBase(one)
			if e != nil {
				res, err = model.Errcode(model.Nomarl, e.Error())
				return
			}
			bhvo.Txs = append(bhvo.Txs, txItr)
		}

		bhvos = append(bhvos, bhvo)
	}

	res, err = model.Tojson(bhvos)
	return

}

/*
	通过一定范围的区块高度查询多个区块详细信息
*/
func FindBlockRangeProto(rj *model.RpcJson, w http.ResponseWriter, r *http.Request) (res []byte, err error) {
	startHeightItr, ok := rj.Get("startHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "startHeight")
		return
	}
	startHeight := uint64(startHeightItr.(float64))

	endHeightItr, ok := rj.Get("endHeight")
	if !ok {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}
	endHeight := uint64(endHeightItr.(float64))

	if endHeight < startHeight {
		res, err = model.Errcode(model.NoField, "endHeight")
		return
	}

	//待返回的区块
	bhvos := make([]*[]byte, 0, endHeight-startHeight+1)

	for i := startHeight; i <= endHeight; i++ {

		bhvo := mining.BlockHeadVO{
			// Txs: make([]mining.TxItr, 0), //交易明细
		}
		bh := mining.LoadBlockHeadByHeight(i)
		// bh := mining.FindBlockHead(i)
		if bh == nil {
			break
		}
		bhvo.BH = bh
		bhvo.Txs = make([]mining.TxItr, 0, len(bh.Tx))

		for _, one := range bh.Tx {
			txItr, e := mining.LoadTxBase(one)
			// txItr, e := mining.FindTxBase(one)
			if e != nil {
				res, err = model.Errcode(model.Nomarl, e.Error())
				return
			}
			bhvo.Txs = append(bhvo.Txs, txItr)
		}
		bs, e := bhvo.Proto()
		if e != nil {
			res, err = model.Errcode(model.Nomarl, e.Error())
			return
		}
		bhvos = append(bhvos, bs)

		// bhvos = append(bhvos, bhvo)
	}

	res, err = model.Tojson(bhvos)
	return

}
