package doubleratchet

import (
	"mandela/core/utils/crypto/dh"
	"bytes"
	"sort"
	"sync"
)

//密钥存储是抽象内存或持久密钥存储的接口。
type KeysStorage interface {
	// get按给定的键和消息编号返回消息键。
	Get(k dh.Key, msgNum uint) (mk dh.Key, ok bool, err error)

	// 将给定的mk保存在指定的键和msgnum下。
	Put(sessionID []byte, k dh.Key, msgNum uint, mk dh.Key, keySeqNum uint) error

	// DeleteMk ensures there's no message key under the specified key and msgNum.
	//deletemk确保在指定的密钥和msgnum下没有消息密钥。
	DeleteMk(k dh.Key, msgNum uint) error

	// DeleteOldMKeys deletes old message keys for a session.
	//deleteoldmkeys删除会话的旧消息键。
	DeleteOldMks(sessionID []byte, deleteUntilSeqKey uint) error

	// TruncateMks truncates the number of keys to maxKeys.
	//truncatemks将键数截断为maxkeys。
	TruncateMks(sessionID []byte, maxKeys int) error

	// Count returns number of message keys stored under the specified key.
	//count返回存储在指定密钥下的消息密钥数。
	Count(k dh.Key) (uint, error)

	// 全部返回所有键
	All() (map[dh.Key]map[uint]dh.Key, error)
}

// keystorageinmemory是内存中的消息密钥存储。
type KeysStorageInMemory struct {
	keys *sync.Map
	// keys map[dh.Key]map[uint]InMemoryKey
}

// get按给定的键和消息编号返回消息键。
func (this *KeysStorageInMemory) Get(pubKey dh.Key, msgNum uint) (dh.Key, bool, error) {
	if this.keys == nil {
		return dh.Key{}, false, nil
	}
	msgs, ok := this.keys.Load(pubKey)
	// msgs, ok := this.keys[pubKey]
	if !ok {
		return dh.Key{}, false, nil
	}
	m := msgs.(*sync.Map)
	mkItr, ok := m.Load(msgNum)
	// mk, ok := msgs[msgNum]
	if !ok {
		return dh.Key{}, false, nil
	}
	mk := mkItr.(InMemoryKey)
	return mk.messageKey, true, nil
}

type InMemoryKey struct {
	messageKey dh.Key
	seqNum     uint
	sessionID  []byte
}

// 将给定的mk保存在指定的键和msgnum下。
func (s *KeysStorageInMemory) Put(sessionID []byte, pubKey dh.Key, msgNum uint, mk dh.Key, seqNum uint) error {
	if s.keys == nil {
		s.keys = new(sync.Map)
		// s.keys = make(map[dh.Key]map[uint]InMemoryKey)
	}
	// if _, ok := s.keys[pubKey]; !ok {
	if _, ok := s.keys.Load(pubKey); !ok {
		s.keys.Store(pubKey, new(sync.Map))
		// s.keys[pubKey] = make(map[uint]InMemoryKey)
	}
	imk := InMemoryKey{
		sessionID:  sessionID,
		messageKey: mk,
		seqNum:     seqNum,
	}
	// s.keys[pubKey][msgNum] = imk
	msgsItr, ok := s.keys.Load(pubKey)
	if !ok {
		return nil
	}
	msgs := msgsItr.(*sync.Map)
	msgs.Store(msgNum, imk)

	return nil
}

// deletemk确保在指定的密钥和msgnum下没有消息密钥。
func (s *KeysStorageInMemory) DeleteMk(pubKey dh.Key, msgNum uint) error {
	if s.keys == nil {
		return nil
	}
	// if _, ok := s.keys[pubKey]; !ok {
	msgsItr, ok := s.keys.Load(pubKey)
	if !ok {
		return nil
	}
	msgs := msgsItr.(*sync.Map)

	// if _, ok := s.keys[pubKey][msgNum]; !ok {
	_, ok = msgs.Load(msgNum)
	if !ok {
		return nil
	}
	// inMemoryKey := mkItr.(InMemoryKey)
	msgs.Delete(msgNum)

	total := 0
	msgs.Range(func(k, v interface{}) bool {
		total++
		//只需要大于1就可以退出了
		return false
	})
	if total <= 0 {
		s.keys.Delete(pubKey)
	}

	// delete(s.keys[pubKey], msgNum)
	// if len(s.keys[pubKey]) == 0 {
	// 	delete(s.keys, pubKey)
	// }
	return nil
}

// truncatemks将键数截断为maxkeys。
func (s *KeysStorageInMemory) TruncateMks(sessionID []byte, maxKeys int) error {
	var seqNos []uint
	// Collect all seq numbers

	// for _, keys := range s.keys {
	// 	for _, inMemoryKey := range keys {
	// 		if bytes.Equal(inMemoryKey.sessionID, sessionID) {
	// 			seqNos = append(seqNos, inMemoryKey.seqNum)
	// 		}
	// 	}
	// }

	if s.keys == nil {
		return nil
	}

	// done := true
	s.keys.Range(func(k, v interface{}) bool {
		msgs := v.(*sync.Map)
		msgs.Range(func(kj, vj interface{}) bool {
			inMemoryKey := vj.(InMemoryKey)
			if bytes.Equal(inMemoryKey.sessionID, sessionID) {
				seqNos = append(seqNos, inMemoryKey.seqNum)
				// done = false
				// return done
			}
			return true
		})
		return true
	})

	// Nothing to do if we haven't reached the limit
	if len(seqNos) <= maxKeys {
		return nil
	}

	// Take the sequence numbers we care about
	sort.Slice(seqNos, func(i, j int) bool { return seqNos[i] < seqNos[j] })
	toDeleteSlice := seqNos[:len(seqNos)-maxKeys]

	// Put in map for easier lookup
	toDelete := make(map[uint]bool)

	for _, seqNo := range toDeleteSlice {
		toDelete[seqNo] = true
	}

	// for pubKey, keys := range s.keys {
	// 	for i, inMemoryKey := range keys {
	// 		if toDelete[inMemoryKey.seqNum] && bytes.Equal(inMemoryKey.sessionID, sessionID) {
	// 			delete(s.keys[pubKey], i)
	// 		}
	// 	}
	// }
	// done = true
	s.keys.Range(func(k, v interface{}) bool {
		// pubKey := k.(dh.Key)
		keys := v.(*sync.Map)
		keys.Range(func(kj, vj interface{}) bool {
			inMemoryKey := vj.(InMemoryKey)
			if toDelete[inMemoryKey.seqNum] && bytes.Equal(inMemoryKey.sessionID, sessionID) {
				keys.Delete(kj)
				// done = false
				// return done
			}
			return true
		})
		return true
	})

	return nil
}

// deleteoldmkeys删除会话的旧消息键。
func (s *KeysStorageInMemory) DeleteOldMks(sessionID []byte, deleteUntilSeqKey uint) error {
	// for pubKey, keys := range s.keys {
	// 	for i, inMemoryKey := range keys {
	// 		if inMemoryKey.seqNum <= deleteUntilSeqKey && bytes.Equal(inMemoryKey.sessionID, sessionID) {
	// 			delete(s.keys[pubKey], i)
	// 		}
	// 	}
	// }
	if s.keys == nil {
		return nil
	}
	// done := true
	s.keys.Range(func(k, v interface{}) bool {
		keys := v.(*sync.Map)
		keys.Range(func(kj, vj interface{}) bool {
			inMemoryKey := vj.(InMemoryKey)
			if inMemoryKey.seqNum <= deleteUntilSeqKey && bytes.Equal(inMemoryKey.sessionID, sessionID) {
				keys.Delete(kj)
			}
			return true
		})
		return true
	})
	return nil
}

// count返回存储在指定密钥下的消息密钥数。
func (s *KeysStorageInMemory) Count(pubKey dh.Key) (uint, error) {
	if s.keys == nil {
		return 0, nil
	}
	keysItr, ok := s.keys.Load(pubKey)
	if !ok {
		return 0, nil
	}
	keys := keysItr.(*sync.Map)
	total := 0
	keys.Range(func(k, v interface{}) bool {
		total++
		return true
	})
	return uint(total), nil

	// return uint(len(s.keys[pubKey])), nil
}

// 全部返回所有键
func (s *KeysStorageInMemory) All() (map[dh.Key]map[uint]dh.Key, error) {
	response := make(map[dh.Key]map[uint]dh.Key)

	// for pubKey, keys := range s.keys {
	// 	response[pubKey] = make(map[uint]dh.Key)
	// 	for n, key := range keys {
	// 		response[pubKey][n] = key.messageKey
	// 	}
	// }
	if s.keys == nil {
		return response, nil
	}

	s.keys.Range(func(k, v interface{}) bool {
		pubKey := k.(dh.Key)
		response[pubKey] = make(map[uint]dh.Key)
		keys := v.(*sync.Map)
		keys.Range(func(kj, vj interface{}) bool {
			n := kj.(uint)
			key := vj.(InMemoryKey)
			response[pubKey][n] = key.messageKey
			return true
		})
		return true
	})

	return response, nil
}
