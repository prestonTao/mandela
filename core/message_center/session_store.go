package message_center

import (
	"mandela/core/keystore"
	"mandela/core/message_center/security_signal/doubleratchet"
	"mandela/core/nodeStore"
	"mandela/core/utils"
	"sync"
)

var sessionManager = NewSessionManager()

type SessionManager struct {
	nodeStore *sync.Map //key:string=节点地址;value:*Session=通道状态信息;
}

/*
	添加一个发送管道
*/
func (this *SessionManager) AddSendPipe(id nodeStore.AddressNet, sk, sharedHka, sharedNhkb [32]byte) error {
	//fmt.Println("添加一个发送管道", hex.EncodeToString(sk[:]), hex.EncodeToString(sharedHka[:]), hex.EncodeToString(sharedNhkb[:]))
	session, err := doubleratchet.NewHE(sk, sharedHka, sharedNhkb, keystore.GetDHKeyPair().KeyPair)
	if err != nil {
		return err
	}

	sessionKey := NewSessionKey(sk, sharedHka, sharedNhkb)
	sessionKey.sessionHE = session

	str := utils.Bytes2string(id) //id.B58String()
	value, ok := this.nodeStore.Load(str)
	if ok {
		session := value.(*Session)
		session.sendPipe = &sessionKey
	} else {
		session := NewSession(id)
		session.sendPipe = &sessionKey
		this.nodeStore.Store(str, &session)
	}
	return nil
}

/*
	删除发送管道
*/
func (this *SessionManager) RemoveSendPipe(id nodeStore.AddressNet) {
	value, ok := this.nodeStore.Load(utils.Bytes2string(id))
	if ok {
		session := value.(*Session)
		session.sendPipe = nil
	}
}

/*
	添加一个接收管道
*/
func (this *SessionManager) AddRecvPipe(id nodeStore.AddressNet, sk, sharedHka, sharedNhkb, puk [32]byte) error {

	session, err := doubleratchet.NewHEWithRemoteKey(sk, sharedHka, sharedNhkb, puk)
	if err != nil {
		return err
	}

	sessionKey := NewSessionKey(sk, sharedHka, sharedNhkb)
	sessionKey.sessionHE = session

	str := utils.Bytes2string(id)
	value, ok := this.nodeStore.Load(str)
	if ok {
		session := value.(*Session)
		session.recvPipe = &sessionKey
	} else {
		session := NewSession(id)
		session.recvPipe = &sessionKey
		this.nodeStore.Store(str, &session)
	}
	return nil
}

/*
	获取一个发送棘轮
*/
func (this *SessionManager) GetSendRatchet(id nodeStore.AddressNet) doubleratchet.SessionHE {
	str := utils.Bytes2string(id) //id.B58String()
	value, ok := this.nodeStore.Load(str)
	if !ok {
		return nil
	}
	session := value.(*Session)
	if session.sendPipe == nil {
		return nil
	}
	return session.sendPipe.sessionHE
}

/*
	获取一个接收棘轮
*/
func (this *SessionManager) GetRecvRatchet(id nodeStore.AddressNet) doubleratchet.SessionHE {
	str := utils.Bytes2string(id) //id.B58String()
	value, ok := this.nodeStore.Load(str)
	if !ok {
		return nil
	}
	session := value.(*Session)
	if session.recvPipe == nil {
		return nil
	}
	return session.recvPipe.sessionHE
}

/*
	创建一个新的节点管理器
*/
func NewSessionManager() SessionManager {
	return SessionManager{nodeStore: new(sync.Map)}
}

type Session struct {
	Id       nodeStore.AddressNet //节点地址
	recvPipe *SessionKey          //接收管道
	sendPipe *SessionKey          //发送管道
}

func NewSession(id nodeStore.AddressNet) Session {
	return Session{
		Id: id, //节点地址
		// recvPipe: make([]SessionKey, 0), //接收管道
		// sendPipe: make([]SessionKey, 0), //发送管道
	}
}

/*
	加密通道
*/
type SessionKey struct {
	sk         [32]byte                //协商密钥
	sharedHka  [32]byte                //随机密钥
	sharedNhkb [32]byte                //随机密钥
	sessionHE  doubleratchet.SessionHE //双棘轮算法状态
}

func NewSessionKey(sk, sharedHka, sharedNhkb [32]byte) SessionKey {
	return SessionKey{
		sk:         sk,         //协商密钥
		sharedHka:  sharedHka,  //随机密钥
		sharedNhkb: sharedNhkb, //随机密钥
	}
}
