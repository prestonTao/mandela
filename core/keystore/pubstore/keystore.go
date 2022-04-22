package pubstore

import (
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/mr-tron/base58"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/hkdf"
)

type Keystore struct {
	filepath string        //keystore文件存放路径
	Wallets  []*Wallet     `json:"wallets"`  //keystore中的所有钱包
	Coinbase uint64        `json:"coinbase"` //当前默认使用的收付款地址
	DHIndex  uint64        `json:"dhindex"`  //DH密钥，指向钱包位置
	Lock     *sync.RWMutex //
}

func NewKeystore() *Keystore {
	keys := Keystore{
		Lock: new(sync.RWMutex), //
	}
	return &keys
}

/*
	从磁盘文件加载keystore
*/
func (this *Keystore) Load(seed string) error {
	this.Lock.Lock()
	// var keystore Keystore
	// bs, err := ioutil.ReadFile(this.filepath)
	// if err != nil {
	// 	return err
	// }
	//fmt.Println(string(bs))

	// err = json.Unmarshal(bs, &this.Wallets)
	bs := []byte(seed)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err := decoder.Decode(&this.Wallets)
	if err != nil {
		return err
	}
	if len(this.Wallets) <= 0 {
		//钱包文件损坏:钱包个数为0
		return errors.New("Damaged wallet file: the number of wallets is 0")
	}
	for i, _ := range this.Wallets {
		//this.Wallets[i].lock = new(sync.RWMutex)
		if !this.Wallets[i].CheckIntact() {
			//钱包文件损坏:第" + strconv.Itoa(i+1) + "个钱包不完整
			return errors.New("Damaged wallet file: No" + strconv.Itoa(i+1) + "Wallet incomplete")
		}
	}
	this.Lock.Unlock()
	return nil
}

/*
	从磁盘文件加载keystore
*/
func (this *Keystore) Save() error {
	// engine.Log.Info("v%", this.Wallets)

	bs, err := json.Marshal(this.Wallets)
	if err != nil {
		return err
	}
	// engine.Log.Info(string(bs))
	return utils.SaveFile(this.filepath, &bs)
}

/*
	创建一个新的种子文件
*/
func (this *Keystore) CreateNewWallet(password [32]byte) error {
	key, err := crypto.Rand32Byte()
	if err != nil {
		return err
	}
	chainCode, err := crypto.Rand32Byte()
	if err != nil {
		return err
	}
	iv, err := crypto.Rand16Byte()
	if err != nil {
		return err
	}

	// fmt.Println("创建的随机数长度", len(key), len(chainCode), len(iv))

	wallet, err := NewWallet(key, chainCode, iv, password)
	if err != nil {
		return err
	}
	this.Lock.Lock()
	this.Wallets = append(this.Wallets, wallet)
	this.Lock.Unlock()
	return nil
}

/*
	使用随机数创建一个新的种子文件
*/
func (this *Keystore) CreateNewWalletRand(rand1, rand2 []byte, password [32]byte) error {

	var key, chainCode [32]byte
	var iv [16]byte

	r := hkdf.New(sha256.New, rand1, rand2, []byte("rsZUpEuXUqqwXBvSy3EcievAh4cMj6QL"))
	buf := make([]byte, 96)
	_, _ = io.ReadFull(r, buf)
	copy(key[:], buf[:32])
	copy(chainCode[:], buf[32:64])
	copy(iv[:], buf[64:80])

	// pwd := sha256.Sum256(rand)

	// fmt.Println("创建的随机数长度", len(key), len(chainCode), len(iv))

	wallet, err := NewWallet(key, chainCode, iv, password)
	if err != nil {
		return err
	}
	this.Lock.Lock()
	this.Wallets = append(this.Wallets, wallet)
	this.Lock.Unlock()
	return nil
}

/*
	获取地址列表
*/
func (this *Keystore) GetAddr() (addrs []crypto.AddressCoin) {
	return this.Wallets[this.Coinbase].GetAddr()
}
func (this *Keystore) Export(pass string) string {
	ks := this
	var seeds []Seed
	for _, val := range ks.Wallets {
		//fmt.Printf("%+v", val)
		//验证密码是否正确
		ok, err := CheckPass(pass, val.Key, val.ChainCode, val.IV, val.CheckHash, false)
		if !ok {
			return err.Error()
		}
		sd := Seed{Key: val.Key, ChainCode: val.ChainCode, IV: val.IV, CheckHash: Ripemd160(val.CheckHash), Index: len(val.Addrs)}
		seeds = append(seeds, sd)
	}
	s := Encode(seeds)
	//fmt.Println(s)
	return s
}

//编码种子
func Encode(seed interface{}) string {
	bs, err := json.Marshal(seed)
	if err != nil {
		fmt.Println(err)
	}
	s := base58.Encode(bs)
	return s
}

//解码种子
func Decode(s string) ([]Seed, error) {
	bs, err := base58.Decode(s)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var seeds []Seed
	err = json.Unmarshal(bs, &seeds)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return seeds, nil
}

//根据种子生成keystore
func (this *Keystore) SeedtoFile(password string, seed []Seed) error {
	this.Lock.Lock()
	this.Wallets = make([]*Wallet, 0)
	for _, val := range seed {
		//验证密码是否正确
		ok, err := CheckPass(password, val.Key, val.ChainCode, val.IV, val.CheckHash, true)
		if !ok {
			return err
		}
		var key, chaincode [32]byte
		var iv [16]byte
		copy(iv[:], val.IV)
		pwd := sha256.Sum256([]byte(password))
		//解密key
		k, s := crypto.DecryptCBC(val.Key, pwd[:], iv[:])
		if s != nil {
			return s
		}
		copy(key[:], k)
		//解密chaincode
		cc, ss := crypto.DecryptCBC(val.ChainCode, pwd[:], iv[:])
		if ss != nil {
			return ss
		}
		copy(chaincode[:], cc)
		wallet, err := NewWallet(key, chaincode, iv, pwd)
		if err != nil {
			fmt.Println(err)
		}
		for i := 0; i < val.Index; i++ {
			wallet.GetNewAddr(pwd)
		}
		this.Wallets = append(this.Wallets, wallet)
	}
	this.Lock.Unlock()
	return nil
	//return kst.Save()
}

//api
func (this *Keystore) GetAddrAll() []crypto.AddressCoin {
	addrs := make([]crypto.AddressCoin, 0)
	//this.lock.RLock()
	for _, one := range this.Wallets {
		addrs = append(addrs, one.GetAddr()...)
	}
	//this.lock.RUnlock()
	return addrs
}

/*
	通过地址获取密钥对
*/
func (this *Keystore) GetKeyByAddr(addr crypto.AddressCoin, password string) (rand []byte, prk ed25519.PrivateKey, puk ed25519.PublicKey, err error) {
	this.Lock.Lock()
	pwd := sha256.Sum256([]byte(password))
	//keystore.lock.RLock()
	for _, one := range this.Wallets {
		rand, prk, puk, err = one.GetKeyByAddr(addr, pwd)
		if err != nil {
			break
		}
	}
	this.Lock.Unlock()
	//keystore.lock.RUnlock()
	return
}

/*
	通过公钥获取密钥
*/
func (this *Keystore) GetKeyByPuk(puk []byte, password string) (rand []byte, prk ed25519.PrivateKey, err error) {
	this.Lock.Lock()
	pwd := sha256.Sum256([]byte(password))
	//keystore.lock.RLock()
	for _, one := range this.Wallets {
		rand, prk, err = one.GetKeyByPuk(puk, pwd)
		if err != nil {
			break
		}
	}
	this.Lock.Unlock()
	//keystore.lock.RUnlock()
	return
}

/*
	通过地址获取公钥
*/
func (this *Keystore) GetPukByAddr(addr crypto.AddressCoin) (puk ed25519.PublicKey, ok bool) {
	ok = false
	this.Lock.Lock()
	for _, one := range this.Wallets {
		if puk, ok = one.GetPukByAddr(addr); ok {
			break
		}
	}
	this.Lock.Unlock()
	return
}

/*
	签名
*/
func (this *Keystore) Sign(prk ed25519.PrivateKey, content []byte) []byte {
	return ed25519.Sign(prk, content)
}
