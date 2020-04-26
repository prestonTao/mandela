package doubleratchet

// TODO: During each DH ratchet step a new ratchet key pair and sending chain are generated.
// As the sending chain is not needed right away, these steps could be deferred until the party
// is about to send a new message.
//TODO:在每个DH棘轮步骤中，都会生成一个新的棘轮密钥对和发送链。由于不需要立即发送链，因此这些步骤可以推迟到该方将要发送新消息时再执行。

import (
	"mandela/core/utils/crypto/dh"
	"fmt"
)

// 双棘轮状态
type State struct {
	Crypto Crypto

	// DH 棘轮公钥 (远端的 key).
	DHr dh.Key

	// DH 棘轮公司约对(自己的棘轮对 key).
	DHs DHPair

	// 对称棘轮根链。
	RootCh kdfRootChain

	// 对称棘轮发送和接收链。
	SendCh, RecvCh kdfChain

	// 发送链中上一个消息的编号
	PN uint32

	//跳过消息密钥的字典，由棘轮公钥或头键和消息编号编制索引。
	MkSkipped KeysStorage

	//单个链中可以跳过的最大消息键数。应该设置得足够高，以允许例程丢失或延迟消息，但是设置得足够低，以至于恶意发送者不能触发过多的收件人计算。
	MaxSkip uint

	//接收头部key和下一个头部key。仅用于头部加密
	HKr, NHKr dh.Key

	//发送头部key和下一个头部key，仅用于头部加密。
	HKs, NHKs dh.Key

	//保留消息键的时间，以接收的消息数计，例如，如果maxkeep为5，我们只保留最后5个消息键，删除所有n-5。
	MaxKeep uint

	// 每个会话的最大消息密钥数，旧密钥将以FIFO方式删除
	MaxMessageKeysPerSession int

	// 当前棘轮步进的编号。
	Step uint

	// KeysCount 已经生成的key数量
	KeysCount uint
}

func DefaultState(sharedKey dh.Key) State {
	c := DefaultCrypto{}

	return State{
		DHs:    dh.DHPair{},
		Crypto: c,
		RootCh: kdfRootChain{CK: sharedKey, Crypto: c},
		//用共享密钥sharedKey填充CKs和CKr，让双方都可以开始发送和接收消息。
		SendCh:                   kdfChain{CK: sharedKey, Crypto: c},
		RecvCh:                   kdfChain{CK: sharedKey, Crypto: c},
		MkSkipped:                &KeysStorageInMemory{},
		MaxSkip:                  1000,
		MaxMessageKeysPerSession: 2000,
		MaxKeep:                  2000,
		KeysCount:                0,
	}
}

func (s *State) applyOptions(opts []option) error {
	for i := range opts {
		if err := opts[i](s); err != nil {
			return fmt.Errorf("failed to apply option: %s", err)
		}
	}
	return nil
}

func newState(sharedKey dh.Key, opts ...option) (State, error) {
	if sharedKey == [32]byte{} {
		return State{}, fmt.Errorf("sharedKey mustn't be empty")
	}

	s := DefaultState(sharedKey)
	if err := s.applyOptions(opts); err != nil {
		return State{}, err
	}

	return s, nil
}

// dhRatchet 执行单个棘轮步骤。
func (s *State) dhRatchet(m MessageHeader) error {
	s.PN = s.SendCh.N
	s.DHr = m.DH
	s.HKs = s.NHKs
	s.HKr = s.NHKr
	s.RecvCh, s.NHKr = s.RootCh.step(s.Crypto.DH(s.DHs, s.DHr))
	var err error
	s.DHs, err = s.Crypto.GenerateDH()
	if err != nil {
		return fmt.Errorf("failed to generate dh pair: %s", err)
	}
	s.SendCh, s.NHKs = s.RootCh.step(s.Crypto.DH(s.DHs, s.DHr))
	return nil
}

type skippedKey struct {
	key dh.Key
	nr  uint
	mk  dh.Key
	seq uint
}

// skipMessageKeys 在当前接收链跳过消息key
func (s *State) skipMessageKeys(key dh.Key, until uint) ([]skippedKey, error) {
	if until < uint(s.RecvCh.N) {
		return nil, fmt.Errorf("bad until: probably an out-of-order message that was deleted")
	}

	if uint(s.RecvCh.N)+s.MaxSkip < until {
		return nil, fmt.Errorf("too many messages")
	}

	skipped := []skippedKey{}
	for uint(s.RecvCh.N) < until {
		mk := s.RecvCh.step()
		skipped = append(skipped, skippedKey{
			key: key,
			nr:  uint(s.RecvCh.N - 1),
			mk:  mk,
			seq: s.KeysCount,
		})
		// Increment key count
		s.KeysCount++

	}
	return skipped, nil
}

func (s *State) applyChanges(sc State, sessionID []byte, skipped []skippedKey) error {
	*s = sc
	for _, skipped := range skipped {
		if err := s.MkSkipped.Put(sessionID, skipped.key, skipped.nr, skipped.mk, skipped.seq); err != nil {
			return err
		}
	}

	if err := s.MkSkipped.TruncateMks(sessionID, s.MaxMessageKeysPerSession); err != nil {
		return err
	}
	if s.KeysCount >= s.MaxKeep {
		if err := s.MkSkipped.DeleteOldMks(sessionID, s.KeysCount-s.MaxKeep); err != nil {
			return err
		}
	}

	return nil
}
