package controllers

import (
	"github.com/astaxie/beego"
)

// Operations about object
type AccountController struct {
	beego.Controller
}

//详情
type Getinfo struct {
	Balance float64 `json:"balance"`
	Testnet bool    `json:"testnet"`
	Blocks  uint64  `json:"blocks"`
}

//详情
type Result struct {
	Jsonrpc string  `json:"jsonrpc"`
	Code    int     `json:"code"`
	Result  Getinfo `json:"result"`
}

// @Title Getinfo
// @Description getinfo
// @Param	body		body 	models.Object	true		"The object content"
// @Success 200 {string} models.Object.Id
// @Failure 403 body is empty
// @router /getinfo [post]
func (o *AccountController) Getinfo() {
	var ob Getinfo
	var result Result
	result.Jsonrpc = "2.0"
	result.Code = 2000
	result.Result = ob
	o.Data["json"] = result
	o.ServeJSON()
}

// @Title GetNewAddress
// @Description GetNewAddress
// @Param	body		body 	models.Object	true		"The object content"
// @Success 200 {string} models.Object.Id
// @Failure 403 body is empty
// @router /getnewaddress [post]
func (o *AccountController) GetNewAddress() {
	var ob Getinfo
	var result Result
	result.Jsonrpc = "2.0"
	result.Code = 2000
	result.Result = ob
	o.Data["json"] = result
	o.ServeJSON()
}

// @Title ListAccounts
// @Description ListAccounts
// @Param	body		body 	models.Object	true		"The object content"
// @Success 200 {string} models.Object.Id
// @Failure 403 body is empty
// @router /listAccounts [post]
func (o *AccountController) ListAccounts() {
	var ob Getinfo
	var result Result
	result.Jsonrpc = "2.0"
	result.Code = 2000
	result.Result = ob
	o.Data["json"] = result
	o.ServeJSON()
}

// @Title GetAccount
// @Description GetAccount
// @Param	body		body 	models.Object	true		"The object content"
// @Success 200 {string} models.Object.Id
// @Failure 403 body is empty
// @router /getAccount [post]
func (o *AccountController) GetAccount() {
	var ob Getinfo
	var result Result
	result.Jsonrpc = "2.0"
	result.Code = 2000
	result.Result = ob
	o.Data["json"] = result
	o.ServeJSON()
}
