package sqlite3_db

import (
	"mandela/config"
	"database/sql"
	"fmt"
	"sync"

	"github.com/go-xorm/xorm"
	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/logoove/sqlite"
)

var once sync.Once

var db *sql.DB

var engineDB *xorm.Engine
var dblock = new(sync.Mutex) //读写数据库的全局锁

var (
	table_friends          *xorm.Session //联系人表格
	table_sharefolder      *xorm.Session //共享目录表格
	table_msglog           *xorm.Session //聊天消息记录
	table_downloadprogress *xorm.Session // 下载列表
	table_storefolder      *xorm.Session // 存储文件夹
	table_storefile        *xorm.Session //文件列表
	table_property         *xorm.Session //个人信息，如昵称等
	table_peerinfo         *xorm.Session //节点信息，如节点公钥等
	table_snapshot         *xorm.Session //社区节点给轻节点奖励快照
	table_reward           *xorm.Session //社区节点给轻节点奖励信息
	table_msgcache         *xorm.Session //消息缓存

	table_wallet_txitem *xorm.Session //未花费的余额
)

func Init() {
	once.Do(connect)

}

func connect() {
	var err error
	engineDB, err = xorm.NewEngine("sqlite3", "file:"+config.SQLITE3DB_path+"?cache=shared")
	if err != nil {
		fmt.Println(err)
	}
	engineDB.ShowSQL(config.SQL_SHOW)

	ok, err := engineDB.IsTableExist(Friends{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(Friends{}).CreateTable(Friends{}) //创建表格
		table_friends = engineDB.Table(Friends{})        //切换表格
		table_friends.CreateIndexes(Friends{})           //创建索引
		table_friends.CreateUniques(Friends{})           //创建唯一性约束

		engineDB.Table(ShareFolder{}).CreateTable(ShareFolder{}) //
		table_sharefolder = engineDB.Table(ShareFolder{})        //
		table_sharefolder.CreateIndexes(ShareFolder{})           //创建索引
		table_sharefolder.CreateUniques(ShareFolder{})           //创建唯一性约束
	} else {
		table_friends = engineDB.Table(Friends{})         //切换表格
		table_sharefolder = engineDB.Table(ShareFolder{}) //
	}

	ok, err = engineDB.IsTableExist(MsgLog{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(MsgLog{}).CreateTable(MsgLog{}) //
		table_msglog = engineDB.Table(MsgLog{})        //
		table_msglog.CreateIndexes(MsgLog{})           //创建索引
		table_msglog.CreateUniques(MsgLog{})           //创建唯一性约束
	} else {
		table_msglog = engineDB.Table(MsgLog{}) //
	}

	LoadMsgLogGenerateID()

	//存储相关
	ok, err = engineDB.IsTableExist(Downprogress{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(Downprogress{}).CreateTable(Downprogress{}) //创建表格
	}
	table_downloadprogress = engineDB.Table(Downprogress{}) //切换表格

	//存储文件夹分类相关
	ok, err = engineDB.IsTableExist(StoreFolder{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(StoreFolder{}).CreateTable(StoreFolder{}) //创建表格
	}
	table_storefolder = engineDB.Table(StoreFolder{}) //切换表格
	//存储文件管理
	ok, err = engineDB.IsTableExist(StoreFolderFile{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(StoreFolderFile{}).CreateTable(StoreFolderFile{}) //创建表格
	}
	table_storefile = engineDB.Table(StoreFolderFile{}) //切换表格
	//个人属性管理
	ok, err = engineDB.IsTableExist(Property{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(Property{}).CreateTable(Property{}) //创建表格
	}
	table_property = engineDB.Table(Property{}) //切换表格

	//节点信息
	ok, err = engineDB.IsTableExist(PeerInfo{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(PeerInfo{}).CreateTable(PeerInfo{}) //创建表格
		table_peerinfo = engineDB.Table(PeerInfo{})        //切换表格
		table_peerinfo.CreateIndexes(PeerInfo{})           //创建索引
		table_peerinfo.CreateUniques(PeerInfo{})           //创建唯一性约束
	} else {
		//如果表存在，先删除表再创建表，目的是删除表中的所有记录
		engineDB.Table(PeerInfo{}).CreateTable(PeerInfo{}) //创建表格
		table_peerinfo = engineDB.Table(PeerInfo{})        //切换表格
		table_peerinfo.CreateIndexes(PeerInfo{})           //创建索引
		table_peerinfo.CreateUniques(PeerInfo{})           //创建唯一性约束
	}

	//社区节点奖励快照
	ok, err = engineDB.IsTableExist(SnapshotReward{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(SnapshotReward{}).CreateTable(SnapshotReward{}) //创建表格
		table_snapshot = engineDB.Table(SnapshotReward{})              //切换表格
		table_snapshot.CreateIndexes(SnapshotReward{})                 //创建索引
		table_snapshot.CreateUniques(SnapshotReward{})                 //创建唯一性约束
	} else {
		//如果表存在，先删除表再创建表，目的是删除表中的所有记录
		engineDB.Table(SnapshotReward{}).CreateTable(SnapshotReward{}) //创建表格
		table_snapshot = engineDB.Table(SnapshotReward{})              //切换表格
		table_snapshot.CreateIndexes(SnapshotReward{})                 //创建索引
		table_snapshot.CreateUniques(SnapshotReward{})                 //创建唯一性约束
	}
	//社区节点奖励
	ok, err = engineDB.IsTableExist(RewardLight{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(RewardLight{}).CreateTable(RewardLight{}) //创建表格
		table_reward = engineDB.Table(RewardLight{})             //切换表格
		table_reward.CreateIndexes(RewardLight{})                //创建索引
		table_reward.CreateUniques(RewardLight{})                //创建唯一性约束
	} else {
		//如果表存在，先删除表再创建表，目的是删除表中的所有记录
		engineDB.Table(RewardLight{}).CreateTable(RewardLight{}) //创建表格
		table_reward = engineDB.Table(RewardLight{})             //切换表格
		table_reward.CreateIndexes(RewardLight{})                //创建索引
		table_reward.CreateUniques(RewardLight{})                //创建唯一性约束
	}

	//消息缓存
	ok, err = engineDB.IsTableExist(MessageCache{})
	if err != nil {
		fmt.Println(err)
	}
	if !ok {
		engineDB.Table(MessageCache{}).CreateTable(MessageCache{}) //创建表格
		table_msgcache = engineDB.Table(MessageCache{})            //切换表格
		table_msgcache.CreateIndexes(MessageCache{})               //创建索引
		table_msgcache.CreateUniques(MessageCache{})               //创建唯一性约束
	} else {
		//如果表存在，先删除表再创建表，目的是删除表中的所有记录
		// engineDB.Table(MessageCache{}).CreateTable(MessageCache{}) //创建表格
		table_msgcache = engineDB.Table(MessageCache{}) //切换表格
		table_msgcache.CreateIndexes(MessageCache{})    //创建索引
		table_msgcache.CreateUniques(MessageCache{})    //创建唯一性约束
	}
	initWalletTable()
}

/*
	初始化钱包相关数据库
*/
func initWalletTable() {

	//如果表存在，先删除表再创建表，目的是删除表中的所有记录
	engineDB.DropTables(TxItem{})
	engineDB.Table(TxItem{}).CreateTable(TxItem{}) //创建表格
	table_wallet_txitem = engineDB.Table(TxItem{}) //切换表格
	table_wallet_txitem.CreateIndexes(TxItem{})    //创建索引
	table_wallet_txitem.CreateUniques(TxItem{})    //创建唯一性约束

}

// func connect2() {
// 	var err error
// 	db, err = sql.Open("sqlite3", config.SQLITE3DB_path)
// 	if err != nil {
// 		// fmt.Println(err)
// 	}
// 	//	defer db.Close()

// 	sqlStmt := `
// 	CREATE TABLE friends (
//   id varchar(255) NOT NULL,
//   name varchar(255) DEFAULT NULL,
//   PRIMARY KEY (id)
// );
// 	`

// 	//	sqlStmt := `
// 	//	create table friends (id integer not null primary key, name text);
// 	//	delete from friends;
// 	//	`
// 	_, err = db.Exec(sqlStmt)
// 	if err != nil {
// 		log.Printf("%q: %s\n", err, sqlStmt)
// 		//		return
// 	}

// 	sqlStmt = `
// 	create table user (id integer not null primary key, name text);
// 	delete from user;
// 	`
// 	_, err = db.Exec(sqlStmt)
// 	if err != nil {
// 		log.Printf("%q: %s\n", err, sqlStmt)
// 		//		return
// 	}

// 	//	Friends_add("123456")

// 	_, err = Friends_getall()
// 	if err != nil {
// 		// fmt.Println(err)
// 	}
// 	// fmt.Println(fs)

// 	//将所有用户id载入内存，方便查询
// 	if loadFriends() != nil {
// 		panic("载入用户到内存失败 查询数据库错误")
// 	}
// }
