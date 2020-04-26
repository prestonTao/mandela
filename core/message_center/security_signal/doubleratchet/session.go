package doubleratchet

import (
	"mandela/core/utils/crypto/dh"
	"fmt"
)

// 参与双棘轮算法的缔约方会议。
type Session interface {
	// RatchetEncrypt performs a symmetric-key ratchet step, then AEAD-encrypts the message with
	// the resulting message key.
	//Ratchetencrypt执行对称的密钥棘轮步骤，然后用AEAD-encrypts加密消息
	//返回消息键。
	RatchetEncrypt(plaintext, associatedData []byte) (Message, error)

	// 调用ratchetdecrypt来对消息进行AEAD解密。
	RatchetDecrypt(m Message, associatedData []byte) ([]byte, error)

	//DeleteMk 从数据库中删除消息密钥
	DeleteMk(dh.Key, uint32) error
}

type sessionState struct {
	id []byte
	State
	storage SessionStorage
}

// 新建使用共享密钥创建会话。
func New(id []byte, sharedKey dh.Key, keyPair DHPair, storage SessionStorage, opts ...option) (Session, error) {
	state, err := newState(sharedKey, opts...)
	if err != nil {
		return nil, err
	}
	state.DHs = keyPair

	session := &sessionState{id: id, State: state, storage: storage}

	return session, session.store()
}

// newWithRemoteKey创建与另一方的共享密钥和公钥的会话。
func NewWithRemoteKey(id []byte, sharedKey, remoteKey dh.Key, storage SessionStorage, opts ...option) (Session, error) {
	state, err := newState(sharedKey, opts...)
	if err != nil {
		return nil, err
	}
	state.DHs, err = state.Crypto.GenerateDH()
	if err != nil {
		return nil, fmt.Errorf("can't generate key pair: %s", err)
	}
	// fmt.Println("NewWithRemoteKey keyPair", hex.EncodeToString(state.DHs.PublicKey().([32]byte)), hex.EncodeToString(state.DHs.PrivateKey()))
	state.DHr = remoteKey
	state.SendCh, _ = state.RootCh.step(state.Crypto.DH(state.DHs, state.DHr))

	session := &sessionState{id: id, State: state, storage: storage}

	return session, session.store()
}

// 从sessionstorage实现加载会话并应用选项。
func Load(id []byte, store SessionStorage, opts ...option) (Session, error) {
	state, err := store.Load(id)
	if err != nil {
		return nil, err
	}

	if state == nil {
		return nil, nil
	}

	if err = state.applyOptions(opts); err != nil {
		return nil, err
	}

	s := &sessionState{id: id, State: *state}
	s.storage = store

	return s, nil
}

func (s *sessionState) store() error {
	if s.storage != nil {
		err := s.storage.Save(s.id, &s.State)
		if err != nil {
			return err
		}
	}
	return nil
}

// RatchetEncrypt Ratchetencrypt执行对称的密钥棘轮步骤，然后加密消息
// the resulting message key.
func (this *sessionState) RatchetEncrypt(plaintext, ad []byte) (Message, error) {
	var (
		h = MessageHeader{
			DH: this.DHs.GetPublicKey(),
			N:  this.SendCh.N,
			PN: this.PN,
		}
		mk = this.SendCh.step()
	)
	ct := this.Crypto.Encrypt(mk, plaintext, append(ad, h.Encode()...))

	// Store state
	if err := this.store(); err != nil {
		return Message{}, err
	}

	return Message{h, ct}, nil
}

// DeleteMk 删除一个消息 key
func (s *sessionState) DeleteMk(dh dh.Key, n uint32) error {
	return s.MkSkipped.DeleteMk(dh, uint(n))
}

// RatchetDecrypt 解密消息
func (this *sessionState) RatchetDecrypt(m Message, ad []byte) ([]byte, error) {
	// 这个消息是否为跳过的消息。
	mk, ok, err := this.MkSkipped.Get(m.Header.DH, uint(m.Header.N))
	if err != nil {
		return nil, err
	}

	if ok {
		plaintext, err := this.Crypto.Decrypt(mk, m.Ciphertext, append(ad, m.Header.Encode()...))
		if err != nil {
			return nil, fmt.Errorf("不能解密跳过的消息: %s", err)
		}
		if err := this.store(); err != nil {
			return nil, err
		}
		return plaintext, nil
	}

	var (
		// 所有的修改必须应用在不同的会话对象，以便此会话不会被修改，也不会保留在脏会话中。
		sc = this.State

		skippedKeys1 []skippedKey
		skippedKeys2 []skippedKey
	)

	// 有新的棘轮钥匙吗？
	if m.Header.DH != sc.DHr {
		if skippedKeys1, err = sc.skipMessageKeys(sc.DHr, uint(m.Header.PN)); err != nil {
			return nil, fmt.Errorf("can't skip previous chain message keys: %s", err)
		}
		if err = sc.dhRatchet(m.Header); err != nil {
			return nil, fmt.Errorf("can't perform ratchet step: %s", err)
		}
	}

	// 别忘了，更新当前链。
	if skippedKeys2, err = sc.skipMessageKeys(sc.DHr, uint(m.Header.N)); err != nil {
		return nil, fmt.Errorf("can't skip current chain message keys: %s", err)
	}
	mk = sc.RecvCh.step()
	plaintext, err := this.Crypto.Decrypt(mk, m.Ciphertext, append(ad, m.Header.Encode()...))
	if err != nil {
		return nil, fmt.Errorf("can't decrypt: %s", err)
	}

	// 追加当前密钥，等待确认
	skippedKeys := append(skippedKeys1, skippedKeys2...)
	skippedKeys = append(skippedKeys, skippedKey{
		key: sc.DHr,
		nr:  uint(m.Header.N),
		mk:  mk,
		seq: sc.KeysCount,
	})

	// 增加key的数量
	sc.KeysCount++

	// Apply changes.
	if err := this.applyChanges(sc, this.id, skippedKeys); err != nil {
		return nil, err
	}

	// Store state
	if err := this.store(); err != nil {
		return nil, err
	}

	return plaintext, nil
}
