package doubleratchet

import "fmt"

// option是一个构造函数选项。
type option func(*State) error

// withmaxskip指定单个链中跳过消息的最大数目。
// nolint: golint
func WithMaxSkip(n int) option {
	return func(s *State) error {
		if n < 0 {
			return fmt.Errorf("n must be non-negative")
		}
		s.MaxSkip = uint(n)
		return nil
	}
}

// withmaxkeep指定保留消息密钥的时间，以接收的消息数计
// nolint: golint
func WithMaxKeep(n int) option {
	return func(s *State) error {
		if n < 0 {
			return fmt.Errorf("n must be non-negative")
		}
		s.MaxKeep = uint(n)
		return nil
	}
}

// withmaxmessagekeyspersession指定每个会话的最大消息密钥数
// nolint: golint
func WithMaxMessageKeysPerSession(n int) option {
	return func(s *State) error {
		if n < 0 {
			return fmt.Errorf("n must be non-negative")
		}
		s.MaxMessageKeysPerSession = n
		return nil
	}
}

// withkeystorage用指定的替换默认密钥存储。
// nolint: golint
func WithKeysStorage(ks KeysStorage) option {
	return func(s *State) error {
		if ks == nil {
			return fmt.Errorf("KeysStorage mustn't be nil")
		}
		s.MkSkipped = ks
		return nil
	}
}

// WithCrypto将默认加密替换为指定的加密方法。
// nolint: golint
func WithCrypto(c Crypto) option {
	return func(s *State) error {
		if c == nil {
			return fmt.Errorf("Crypto mustn't be nil")
		}
		s.Crypto = c
		s.RootCh.Crypto = c
		s.SendCh.Crypto = c
		s.RecvCh.Crypto = c
		return nil
	}
}
