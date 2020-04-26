package mining

import (
	"mandela/config"
	"mandela/core/keystore"
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"crypto/ed25519"
	"encoding/json"
)

const (
	SIGN_TYPE_community_reward = 1 //社区节点发放奖励签名
)

type CommunitySign struct {
	Type        uint64 //类型
	StartHeight uint64 //快照开始高度
	EndHeight   uint64 //快照结束高度
	Rand        uint64 //随机数
	Puk         []byte //公钥
	Sign        []byte //签名
}

func (this *CommunitySign) Json() []byte {
	bs, _ := json.Marshal(this)
	return bs
}

func ParseCommunitySign(bs []byte) (*CommunitySign, error) {
	cs := new(CommunitySign)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(cs)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func NewCommunitySign(puk []byte, startHeight, endHeight uint64) *CommunitySign {
	max := utils.BytesToUint64([]byte{255 - 128, 255, 255, 255, 255, 255, 255, 255})
	r := utils.GetRandNum(int64(max))
	return &CommunitySign{
		Type:        SIGN_TYPE_community_reward, //
		StartHeight: startHeight,                //快照开始高度
		EndHeight:   endHeight,                  //快照结束高度
		Rand:        uint64(r),                  //随机数
		Puk:         puk,                        //
		Sign:        nil,                        //签名
	}
}

/*
	签名
*/
func SignPayload(txItr TxItr, puk []byte, prk ed25519.PrivateKey, startHeight, endHeight uint64) TxItr {
	cs := NewCommunitySign(puk, startHeight, endHeight)
	txItr.SetPayload(cs.Json())
	//所有签名字段设置为空
	for i, _ := range *txItr.GetVin() {
		txItr.SetSign(uint64(i), nil)
	}
	signDst := txItr.Serialize()
	sign := keystore.Sign(prk, *signDst)
	cs.Sign = sign
	txItr.SetPayload(cs.Json())
	return txItr
}

/*
	验证签名
	@return    crypto.AddressCoin    签名者地址
	@return    bool                  签名是否正确
*/
func CheckPayload(txItr TxItr) (crypto.AddressCoin, bool, *CommunitySign) {

	bs := txItr.GetPayload()
	if bs == nil || len(bs) <= 0 {
		return nil, false, nil
	}
	cs, err := ParseCommunitySign(bs)
	if err != nil {
		return nil, false, nil
	}
	if cs.Puk == nil || len(cs.Puk) <= 0 || cs.Sign == nil || len(cs.Sign) <= 0 {
		return nil, false, nil
	}
	addr := crypto.BuildAddr(config.AddrPre, cs.Puk)
	if cs.Type != SIGN_TYPE_community_reward {
		return nil, false, nil
	}
	signtmp := cs.Sign
	cs.Sign = nil
	txItr.SetPayload(cs.Json())

	signs := make([][]byte, 0)
	//所有签名字段设置为空
	for i, _ := range *txItr.GetVin() {
		//txItr.GetSign()
		signs = append(signs, (*txItr.GetVin())[i].Sign)
		txItr.SetSign(uint64(i), nil)
	}
	signDst := txItr.Serialize()
	cs.Sign = signtmp
	//传进来的参数被改变了值，现在改回去
	txItr.SetPayload(cs.Json())
	for i, _ := range signs {
		txItr.SetSign(uint64(i), signs[i])
	}
	if !ed25519.Verify(cs.Puk, *signDst, cs.Sign) {
		return addr, false, nil
	}

	return addr, true, cs
}
