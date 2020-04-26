package role

import (
	// "bytes"
	acc "common/message/m_acc"
	// "database/sql"
	"fmt"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/thrsafe"
	// "strconv"
)

type DataModule struct {
}

func (this *DataModule) AccountloginToo(param *acc.LoginPlayer_MA, itbid int32, conn mysql.Conn) *AccParams {
	str := fmt.Sprintf(`set @iisfcm = %d`, 1)
	res, err := conn.Start(str)
	if err != nil {
		fmt.Println("存储过程调用失败", err.Error())
	}

	str = fmt.Sprintf(`set @iserverid = %d`, 0)
	res, err = conn.Start(str)
	if err != nil {
		fmt.Println("存储过程调用失败", err.Error())
	}

	str = fmt.Sprintf(`call mfxy_accountlogin(%d,'%s','%s','%s','%s',@iserverid,%d,%d,%d,@iisfcm,@ifcmtime,@iaccid,@igmlevel,@igamepoints,@szprotectques,@szprotectansw,@iretval)`,
		itbid, *param.SzAccount, *param.SzPassword, *param.SzIp, "haha", *param.ProductId, *param.PlatformId, *param.LoginType)
	fmt.Println(str)
	res, err = conn.Start(str)
	if err != nil {
		fmt.Println("存储过程调用失败", err.Error())
	}

	str = fmt.Sprintf(`select @iserverid,@iisfcm,@ifcmtime,@iaccid,@igmlevel,@igamepoints,@szprotectques,@szprotectansw,@iretval`)
	fmt.Println(str)
	res, err = conn.Start(str)
	if err != nil {
		fmt.Println("存储过程调用失败", err.Error())
	}

	row, _ := res.GetRow()

	repParams := new(AccParams)
	repParams.Serverid = uint64(row.Int(0))
	repParams.Isfcm = uint32(row.Int(1))
	repParams.Ifcmtime = uint64(row.Int64(2))
	repParams.Accid = uint64(row.Int64(3))
	repParams.Gmlevel = uint32(row.Int(4))
	repParams.Points = uint64(row.Int(5))
	repParams.Productid = uint64(row.Int64(6))
	repParams.Protectansw = row.Str(7)
	repParams.Iretval = int32(row.Int(8))
	fmt.Println("datamodul end")
	return repParams
}

func (this *DataModule) AccountloginOut(itbid int32, param *acc.LoginPlayer_MA) {
	// idbid, itbid := GetDatabaseId(param.SzAccount)

	// str := fmt.Sprintf(`call mfxy_accountlogout(%d,"%s","%s","%s",%d,%d,%d,%d)`,
	// 	itbid, param.SzAccount, param.SzPassword, param.Ip, param.SzIp, param.ProductId, param.PlatformId, param.LoginType)
	// fmt.Println(str)
	// conn := this.connStore["conn"+strconv.Atoi(idbid)]
	// res, err := conn.Start(str)
	// if err != nil {
	// 	fmt.Println("存储过程调用失败")
	// }
	// fmt.Println("调用成功")
	// row, _ := res.GetRow()
	// fmt.Println(row.Str(0))
}

type AccParams struct {
	Account     string //帐号(in).
	Pwd         string //密码(in).
	IP          string //登入IP(in).
	Cardid      string //卡号(in).
	Serverid    uint64 //游戏服务器ID(input).
	Productid   uint64 //产品ID(in).
	Platformid  uint32 //平台ID(in).
	Type        uint32 //验证类型(登入,挂机重上)(GM,非GM等?)(登出类型(登出,挂机))(in).
	Level       uint32 //角色等级(logout,offline中,用于统计)(in).
	Isfcm       uint32 //防沉迷标识(inout).
	Ifcmtime    uint64 //防沉迷时间(在线时长,分钟)(inout).
	Accid       uint64 //帐号ID(out,in).
	Gmlevel     uint32 //GM权限(out).
	Points      uint64 //游戏币(out).
	Protectques string //二级密码问题(out).
	Protectansw string //二级密码答案(out).
	Iretval     int32  //存储过程执行结果
}
