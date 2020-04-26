package controllers

import (
	"mandela/chain_witness_vote/mining"
	gconfig "mandela/config"
	"mandela/core/cache_store"
	"mandela/core/message_center"
	"mandela/core/nodeStore"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {
	// this.TplName = "store/index.tpl"
	// this.Ctx.Redirect(http.StatusMovedPermanently, "/static/index.html")
	this.Ctx.Redirect(http.StatusFound, "/self/test")
}

/*
	测试
*/
func (this *MainController) Test() {

	this.Data["Ip"] = nodeStore.NodeSelf.Addr

	this.Data["RootExist"] = cache_store.Root.Exist

	this.Data["IsSuper"] = nodeStore.NodeSelf.IsSuper
	if nodeStore.SuperPeerId != nil {
		this.Data["SuperId"] = nodeStore.SuperPeerId.B58String()
	} else {
		this.Data["SuperId"] = ""
	}

	//	fmt.Println("首页")
	this.Data["ID"] = nodeStore.NodeSelf.IdInfo.Id.B58String()

	ids := nodeStore.GetLogicNodes()
	idsStr := make([]string, 0)
	for _, one := range ids {
		idsStr = append(idsStr, one.B58String())
	}
	this.Data["ids"] = idsStr

	names := cache_store.Debug_GetAllName()
	this.Data["names"] = names

	this.TplName = "test.tpl"
}

/*
	发送消息
*/
func (this *MainController) SendMeg() {
	id := this.GetString("id")
	recvId := nodeStore.AddressFromB58String(id)

	content := []byte(this.GetString("content"))

	message_center.SendP2pMsgHE(gconfig.MSGID_TextMsg, &recvId, &content)

	out := make(map[string]interface{})
	out["Code"] = 0
	this.Data["json"] = out
	this.ServeJSON(true)
}

/*
	代理
*/
func (this *MainController) AgentToo() {
	fmt.Println("没有命中")
}

/*
	测试按钮
*/
func (this *MainController) BtTest() {
	// fmt.Println("测试按钮")
	mining.Seekvote()
	out := make(map[string]interface{})
	out["Code"] = 0
	this.Data["json"] = out
	this.ServeJSON(true)
}
