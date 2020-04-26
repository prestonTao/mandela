package controllers

import (
	"mandela/core"
	"mandela/core/persistence"
	"time"

	"github.com/astaxie/beego"
)

type MsgController struct {
	beego.Controller
}

func (this *MsgController) MsgPage() {
	fs := persistence.Friends_getall()

	this.Data["Fs"] = fs

	this.TplName = "message.tpl"
	// fmt.Println("  ========== ========== ============")
}

/*
	2.5分钟的长轮询来获取消息
*/
func (this *MsgController) GetMsg() {
	out := make(map[string]interface{})

	overtime := time.NewTicker(time.Second * 20)
	select {
	case <-overtime.C:
		out["Code"] = 1
	case <-this.Ctx.ResponseWriter.CloseNotify():
		// fmt.Println(isClose)
		overtime.Stop()
		//		return
	case msg := <-core.MsgChannl:
		overtime.Stop()
		out["Code"] = 0
		out["Id"] = msg.Id
		out["Index"] = msg.Index
		out["Content"] = msg.Content
	}
	this.Data["json"] = out
	this.ServeJSON(true)

	//	n, err := io.WriteString(this.Ctx.ResponseWriter, `{"Code":1}`)
	//	fmt.Println(n, err)

	//	n, err = this.Ctx.ResponseWriter.Write([]byte(`{"Code":1}`))
	//	//	this.Ctx.ResponseWriter.
	//	fmt.Println(n, err)

}

/*
	添加一个朋友
*/
func (this *MsgController) AddFriend() {
	out := make(map[string]interface{})
	id := this.GetString("ID")
	if id == "" {
		out["Code"] = 1
		this.Data["json"] = out
		this.ServeJSON(true)
		return
	}
	err := persistence.Friends_add(id)
	if err != nil {
		out["Code"] = 1
	} else {
		out["Code"] = 0
	}

	this.Data["json"] = out
	this.ServeJSON(true)
}
