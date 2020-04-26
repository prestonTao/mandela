package keystore

import (
	"mandela/core/utils"
	"mandela/core/utils/crypto"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"strconv"
	"sync"

	"golang.org/x/crypto/hkdf"
)

type Keystore struct {
	filepath string        //keystore文件存放路径
	Wallets  []*Wallet     `json:"wallets"`  //keystore中的所有钱包
	Coinbase uint64        `json:"coinbase"` //当前默认使用的收付款地址
	DHIndex  uint64        `json:"dhindex"`  //DH密钥，指向钱包位置
	lock     *sync.RWMutex //
}

func NewKeystore(filepath string) *Keystore {
	keys := Keystore{
		filepath: filepath,          //keystore文件存放路径
		lock:     new(sync.RWMutex), //
	}
	return &keys
}

/*
	从磁盘文件加载keystore
*/
func (this *Keystore) Load() error {
	// var keystore Keystore
	bs, err := ioutil.ReadFile(this.filepath)
	if err != nil {
		return err
	}
	//fmt.Println(string(bs))

	// err = json.Unmarshal(bs, &this.Wallets)
	decoder := json.NewDecoder(bytes.NewBuffer(bs))
	decoder.UseNumber()
	err = decoder.Decode(&this.Wallets)
	if err != nil {
		return err
	}
	if len(this.Wallets) <= 0 {
		//钱包文件损坏:钱包个数为0
		return errors.New("Damaged wallet file: the number of wallets is 0")
	}
	for i, _ := range this.Wallets {
		this.Wallets[i].lock = new(sync.RWMutex)
		if !this.Wallets[i].CheckIntact() {
			//钱包文件损坏:第" + strconv.Itoa(i+1) + "个钱包不完整
			return errors.New("Damaged wallet file: No" + strconv.Itoa(i+1) + "Wallet incomplete")
		}
	}
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
	this.lock.Lock()
	this.Wallets = append(this.Wallets, wallet)
	this.lock.Unlock()
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
	this.lock.Lock()
	this.Wallets = append(this.Wallets, wallet)
	this.lock.Unlock()
	return nil
}

/*
	获取地址列表
*/
func (this *Keystore) GetAddr() (addrs []crypto.AddressCoin) {
	return this.Wallets[this.Coinbase].GetAddr()
}
