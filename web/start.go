package web

import (
	"mandela/config"
	routers "mandela/web/routers"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/astaxie/beego"
)

//Start Start
func Start() {

	beego.BConfig.WebConfig.Session.SessionOn = false
	beego.BConfig.Listen.HTTPPort = int(config.WebPort)
	//	搴旂敤鐩戝惉鍦板潃锛岄粯璁や负绌猴紝鐩戝惉鎵€鏈夌殑缃戝崱 IP銆
	beego.BConfig.Listen.HTTPSAddr = config.WebAddr
	//瀛樺湪瀹㈡埛绔¯鐨 cookie 鍚嶇О锛岄粯璁ゅ€兼槸 beegosessionID銆
	beego.BConfig.WebConfig.Session.SessionName = "mandela"
	//session 杩囨湡鏃堕棿锛岄粯璁ゅ€兼槸 3600 绉掋€
	beego.BConfig.WebConfig.Session.SessionGCMaxLifetime = 3600
	beego.BConfig.CopyRequestBody = true
	beego.BConfig.WebConfig.TemplateLeft = "<%"
	beego.BConfig.WebConfig.TemplateRight = "%>"
	// beego.PprofOn = true

	//home
	//	beego.SetStaticPath("/static", `D:\workspaces\go\src\mandela\web\static`)
	//	beego.BConfig.WebConfig.ViewsPath = `D:\workspaces\go\src\mandela\web\views`

	//inc
	//	beego.SetStaticPath("/static", `D:\workspace\src\mandela\web\static`)
	//	beego.BConfig.WebConfig.ViewsPath = `D:\workspace\src\mandela\web\views`
	// beego.SetStaticPath("/static", config.Web_path_static)
	beego.BConfig.WebConfig.ViewsPath = config.Web_path_views
	beego.SetStaticPath("/static", config.Web_path_static)
	routers.Start()
	//	go openLocalWeb()
	// beego.Run()
}

// Open calls the OS default program for uri
func openLocalWeb() {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start http://127.0.0.1:"+strconv.Itoa(int(config.WebPort))+"/")
	case "darwin":
		cmd = exec.Command("open", "http://127.0.0.1:"+strconv.Itoa(int(config.WebPort))+"/")
	case "linux":
		cmd = exec.Command("xdg-open", "http://127.0.0.1:"+strconv.Itoa(int(config.WebPort))+"/")

	}
	err := cmd.Start()
	if err != nil {
		// fmt.Printf("启动页面的时候发生错误:%s", err.Error())
	}
}
