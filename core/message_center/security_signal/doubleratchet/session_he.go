package doubleratchet

import (
	"mandela/core/utils/crypto/dh"
	"fmt"
)

// 会话这是一个涉及加密头修改的双棘轮算法的会话。
type SessionHE interface {
	// RatchetEncrypt performs a symmetric-key ratchet step, then AEAD-encrypts
	// the header-encrypted message with the resulting message key.
	//Ratchetencrypt执行对称的密钥棘轮步骤，然后AEAD加密
	//头用结果消息密钥加密的消息。
	RatchetEncrypt(plaintext, associatedData []byte) MessageHE

	// RatchetDecrypt is called to AEAD-decrypt header-encrypted messages.
	RatchetDecrypt(m MessageHE, associatedData []byte) ([]byte, error)
}

type sessionHE struct {
	State
}

// NewHE creates session with the shared keys.
func NewHE(sharedKey, sharedHka, sharedNhkb dh.Key, keyPair DHPair, opts ...option) (SessionHE, error) {
	state, err := newState(sharedKey, opts...)
	if err != nil {
		return nil, err
	}
	state.DHs = keyPair
	state.NHKs = sharedNhkb
	state.HKs = sharedHka
	state.NHKr = sharedHka
	return &sessionHE{state}, nil
}

// NewHEWithRemoteKey creates session with the shared keys and public key of the other party.
func NewHEWithRemoteKey(sharedKey, sharedHka, sharedNhkb, remoteKey dh.Key, opts ...option) (SessionHE, error) {
	state, err := newState(sharedKey, opts...)
	if err != nil {
		return nil, err
	}
	state.DHs, err = state.Crypto.GenerateDH()
	if err != nil {
		return nil, fmt.Errorf("can't generate key pair: %s", err)
	}
	state.DHr = remoteKey
	state.SendCh, state.NHKs = state.RootCh.step(state.Crypto.DH(state.DHs, state.DHr))
	state.HKs = sharedHka
	state.NHKr = sharedNhkb
	state.HKr = sharedHka
	return &sessionHE{state}, nil
}

// RatchetEncrypt performs a symmetric-key ratchet step, then encrypts the header with
// the corresponding header key and the message with resulting message key.
func (s *sessionHE) RatchetEncrypt(plaintext, ad []byte) MessageHE {
	var (
		h = MessageHeader{
			DH: s.DHs.GetPublicKey(),
			N:  s.SendCh.N,
			PN: s.PN,
		}
		mk   = s.SendCh.step()
		hEnc = s.Crypto.Encrypt(s.HKs, h.Encode(), nil)
	)
	return MessageHE{
		Header:     hEnc,
		Ciphertext: s.Crypto.Encrypt(mk, plaintext, append(ad, hEnc...)),
	}
}

// RatchetDecrypt is called to AEAD-decrypt header-encrypted messages.
func (s *sessionHE) RatchetDecrypt(m MessageHE, ad []byte) ([]byte, error) {
	// Is the message one of the skipped?
	if plaintext, err := s.trySkippedMessages(m, ad); err != nil || plaintext != nil {
		return plaintext, err
	}

	h, step, err := s.decryptHeader(m.Header)
	if err != nil {
		return nil, fmt.Errorf("can't decrypt header: %s", err)
	}

	var (
		// All changes must be applied on a different session object, so that this session won't be modified nor left in a dirty session.
		sc = s.State

		skippedKeys1 []skippedKey
		skippedKeys2 []skippedKey
	)
	if step {
		if skippedKeys1, err = sc.skipMessageKeys(sc.HKr, uint(h.PN)); err != nil {
			return nil, fmt.Errorf("can't skip previous chain message keys: %s", err)
		}
		if err = sc.dhRatchet(h); err != nil {
			return nil, fmt.Errorf("can't perform ratchet step: %s", err)
		}
	}

	// After all, update the current chain.
	if skippedKeys2, err = sc.skipMessageKeys(sc.HKr, uint(h.N)); err != nil {
		return nil, fmt.Errorf("can't skip current chain message keys: %s", err)
	}
	mk := sc.RecvCh.step()
	plaintext, err := s.Crypto.Decrypt(mk, m.Ciphertext, append(ad, m.Header...))
	if err != nil {
		return nil, fmt.Errorf("can't decrypt: %s", err)
	}

	if err = s.applyChanges(sc, []byte("FIXME"), append(skippedKeys1, skippedKeys2...)); err != nil {
		return nil, fmt.Errorf("failed to apply changes: %s", err)
	}

	return plaintext, nil
}

func (s *sessionHE) decryptHeader(encHeader []byte) (MessageHeader, bool, error) {
	if encoded, err := s.Crypto.Decrypt(s.HKr, encHeader, nil); err == nil {
		h, err := MessageEncHeader(encoded).Decode()
		return h, false, err
	}
	if encoded, err := s.Crypto.Decrypt(s.NHKr, encHeader, nil); err == nil {
		h, err := MessageEncHeader(encoded).Decode()
		return h, true, err
	}
	return MessageHeader{}, false, fmt.Errorf("invalid header")
}

func (s *sessionHE) trySkippedMessages(m MessageHE, ad []byte) ([]byte, error) {
	allMessages, err := s.MkSkipped.All()
	if err != nil {
		return nil, err
	}

	for hk, keys := range allMessages {
		for n, mk := range keys {
			hEnc, err := s.Crypto.Decrypt(hk, m.Header, nil)
			if err != nil {
				continue
			}
			h, err := MessageEncHeader(hEnc).Decode()
			if err != nil {
				return nil, fmt.Errorf("can't decode header %s for skipped message key under (%s, %d)", hEnc, hk, n)
			}
			if uint(h.N) != n {
				continue
			}
			plaintext, err := s.Crypto.Decrypt(mk, m.Ciphertext, append(ad, m.Header...))
			if err != nil {
				return nil, fmt.Errorf("can't decrypt skipped message: %s", err)
			}
			_ = s.MkSkipped.DeleteMk(hk, n)
			return plaintext, nil
		}
	}
	return nil, nil
}
