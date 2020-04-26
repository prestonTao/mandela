package wallet

import (
	"mandela/chain_witness_vote"

	"github.com/astaxie/beego"
)

type Account struct {
	beego.Controller
}

func (this *Account) GetInfo() {
	//	names, _ := store.GetFileinfoToSelfAll()

	//	fmt.Println("网络文件个数为", len(names))
	//	this.Data["Names"] = names

	this.Data["CheckKey"] = chain_witness_vote.CheckKey()

	this.TplName = "wallet/index.tpl"
}
