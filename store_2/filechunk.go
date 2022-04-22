/*
	文件分块
*/
package store

import (
	gconfig "mandela/config"
	"mandela/core/nodeStore"
	"crypto/sha1"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type File struct {
	Name       string                //真实文件名称
	Hash       *nodeStore.AddressNet //完整文件hash值
	FileChunk  []*FileChunk          //文件块内容
	ChunkCount uint64                //文件块总数
	lock       *sync.RWMutex         //读写锁
}

/*
	添加一个块
*/
func (this *File) AddFileChunk(chunk *FileChunk) {
	//不合法的块编号
	if chunk.No >= this.ChunkCount {
		return
	}
	this.lock.Lock()
	//判断块重复
	have := false
	for _, one := range this.FileChunk {
		if one.No == chunk.No {
			have = true
			break
		}
	}
	if !have {
		//		fmt.Println("添加一个文件块")
		this.FileChunk = append(this.FileChunk, chunk)
	}
	this.lock.Unlock()
}

/*
	检查文件块是否完整
*/
func (this *File) Check() bool {
	this.lock.Lock()
	sort.Sort(this)
	this.lock.Unlock()
	//	fmt.Println("check", len(this.FileChunk), int(this.ChunkCount))
	if len(this.FileChunk) == int(this.ChunkCount) {
		return true
	}
	return false
}

//type FileChunkContent struct {
//	No      uint64 //文件块编号
//	Hash    string //fileinfo hash
//	Index   uint64 //下载块起始位置
//	Length  uint64 //下载块长度
//	Content bool   //是否把块数据块下载到了本地
//}

/*
	创建一个文件块内容
*/
//func NewFileChunkContent(fc *FileChunk)*FileChunkContent {
//return &FileChunkContent{
//		No    : fc.No,
//	Hash    :fc.Hash, //fileinfo hash
//	Index  :fc.//下载块起始位置
//	Length  uint64 //下载块长度
//	Content []byte //块数据
//}
//}

/*
	把所有切片组装成完整文件
*/
func (this *File) Assemble() error {
	this.lock.Lock()
	sort.Sort(this)
	this.lock.Unlock()
	f, err := os.OpenFile(filepath.Join(gconfig.Store_temp, this.Hash.B58String()), os.O_RDWR|os.O_CREATE, os.ModePerm)
	//	f, err := os.Create(filepath.Join(gconfig.Store_temp, this.Hash))
	if err != nil {
		f.Close()
		return err
	}
	for _, one := range this.FileChunk {
		// fmt.Println("读文件块", one.Hash)
		bs, err := ioutil.ReadFile(filepath.Join(gconfig.Store_dir, one.Hash.B58String()))
		if err != nil {
			f.Close()
			return err
		}
		_, err = f.Write(bs)
		if err != nil {
			f.Close()
			return err
		}
	}
	f.Close()
	return os.Rename(filepath.Join(gconfig.Store_temp, this.Hash.B58String()), filepath.Join(gconfig.Store_temp, this.Name))
}

func (this *File) Len() int {
	return len(this.FileChunk)
}

func (this *File) Less(i, j int) bool {
	return this.FileChunk[i].No < this.FileChunk[j].No // 按值排序
}

func (this *File) Swap(i, j int) {
	this.FileChunk[i], this.FileChunk[j] = this.FileChunk[j], this.FileChunk[i]
	//	ms[i], ms[j] = ms[j], ms[i]
}

/*
	创建一个新的文件
*/
func NewFile(fi *FileInfo) *File {
	return &File{
		Name:       fi.Name,
		Hash:       fi.Hash,
		FileChunk:  []*FileChunk{},
		ChunkCount: fi.ChunkCount,     //文件块总数
		lock:       new(sync.RWMutex), //读写锁
	}
}

/*
	把文件切片
	@name  string  真实文件名称(temp文件夹中的文件名称)
*/
func Diced(name string) (*FileInfo, error) {

	//	bs, err := ioutil.ReadFile(filepath.Join(gconfig.Store_temp, name))
	//	if err != nil {
	//		return nil, err
	//	}

	//f, err := os.Open(filepath.Join(gconfig.Store_temp, name))
	f, err := os.Open(name)
	defer f.Close()
	if err != nil {
		// fmt.Println("111", err)
		return nil, err
	}
	fileinfo, err := f.Stat()

	//先计算整个文件的hash值
	h := sha1.New()
	_, err = io.Copy(h, f)
	if err != nil {
		// fmt.Println("222", err)
		return nil, err
	}
	// bs, err := utils.Encode(h.Sum(nil), gconfig.HashCode)
	// if err != nil {
	// 	// fmt.Println("333", err)
	// 	return nil, err
	// }
	// hashName := utils.Multihash(bs)
	hashName := nodeStore.AddressNet(h.Sum(nil))
	//	hashName := hex.EncodeToString(h.Sum(nil))
	chunkCount := fileinfo.Size() / Chunk_size
	if (fileinfo.Size() % Chunk_size) > 0 {
		chunkCount = chunkCount + 1
	}

	fi := NewFileInfo(&hashName, name, uint64(chunkCount))
	fi.Size = uint64(fileinfo.Size())
	//上传时间
	fi.Time = time.Now().Unix()
	bs := make([]byte, Chunk_size)
	for i := 0; i < int(fileinfo.Size()/Chunk_size); i++ {
		f.Seek(int64(i*Chunk_size), 0)
		//		_, err = io.ReadAtLeast(f, bs, Chunk_size)
		_, err := f.Read(bs)
		if err != nil {
			// fmt.Println("333", n, err)
			return nil, err
		}
		//		buf := bufio.NewReader(f)
		//		bs, err := buf.Peek(Chunk_size)
		//		if err != nil {
		//			return nil, err
		//		}
		chunkHash, err := writeFileChunk(&bs)
		if err != nil {
			// fmt.Println("444", err)
			return nil, err
		}
		//		fmt.Println("读取了", n, chunkHash)
		chunk := NewFileChunk(uint64(i), chunkHash)
		chunk.Size = uint64(len(bs))
		fi.AddChunk(chunk)

	}
	//有余就再把余下的文件做成一个块
	if fileinfo.Size()%Chunk_size > 0 {
		f.Seek(fileinfo.Size()/Chunk_size*Chunk_size, 0)
		n, err := f.Read(bs)
		//		_, err = io.ReadAtLeast(f, bs, Chunk_size)
		if err != nil {
			// fmt.Println("555", err)
			return nil, err
		}
		bs = bs[:n]
		chunkHash, err := writeFileChunk(&bs)
		if err != nil {
			// fmt.Println("666", err)
			return nil, err
		}
		//		fmt.Println("读取了", n, chunkHash)
		chunk := NewFileChunk(uint64(fileinfo.Size()/Chunk_size), chunkHash)
		chunk.Size = uint64(len(bs))
		fi.AddChunk(chunk)

	}

	return fi, nil
}

/*
	写一个文件块到磁盘
	@return    string    文件名称（文件hash值）
	@return    error     返回错误
*/
func writeFileChunk(bs *[]byte) (*nodeStore.AddressNet, error) {
	sha1hash := sha1.New()
	_, err := sha1hash.Write(*bs)
	if err != nil {
		// fmt.Println(err)
		return nil, err
	}

	//计算文件的sha1 Hash值
	// hashName := utils.Multihash(sha1hash.Sum(nil))
	hashName := nodeStore.AddressNet(sha1hash.Sum(nil))

	//把文件块写到目标文件夹
	file, err := os.OpenFile(filepath.Join(gconfig.Store_dir, hashName.B58String()), os.O_RDWR|os.O_CREATE, os.ModePerm)
	//	file, err := os.Create(filepath.Join(gconfig.Store_dir, hashName))
	if err != nil {
		file.Close()
		return nil, err
	}

	_, err = file.Write(*bs)
	if err != nil {
		// fmt.Println(err)
		file.Close()
		return nil, err
	}
	file.Close()
	return &hashName, nil
}
