package boot

/*
import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"mandela/core"
)

func StartCommandWindow() {
	//命令控制中心发送程序停止命令
	stopChan := make(chan bool, 1)
	//命令行输入的命令和参数
	lineChan := make(chan string, 1)

	reader := bufio.NewReader(os.Stdin)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	//启动anonymousnet程序
	core.StartService()

	// go web.StartWeb()

	running := true
	for running {
		go ReadLine(reader, lineChan)
		select {
		case dataStr := <-lineChan:
			//执行命令
			CtlCenter(strings.Split(dataStr, " "), stopChan)
		case <-c:
			//Ctrl + c 退出程序
			fmt.Println("Ctrl + c 退出程序")
			running = false
			core.ShutdownCallback()
		case <-stopChan:
			//stop 命令退出程序
			fmt.Println("stop 命令退出程序")
			running = false
			core.ShutdownCallback()
		}
	}
}

func ReadLine(reader *bufio.Reader, c chan string) {
	data, _, _ := reader.ReadLine()
	c <- string(data)
}

//命令控制中心
func CtlCenter(commands []string, stopChan chan bool) {
	switch commands[0] {
	case "help":

	case "quit":
		stopChan <- false
	case "exit":
		stopChan <- false
	case "start":
		Launcher()
	case "send":
		SendMsgAll_script(commands)
	case "see":
		SelectAllPeer(commands[1])
	case "createdomain":
		CreateDomain_script(commands)
	}
}

//查询自己保存的逻辑节点
func SelectAllPeer(domain string) {
	switch domain {
	case "all":
		core.See()
	case "left":
		//		core.SeeLeftNode()
	case "right":
		//		core.SeeRightNode()
	case "super":
		core.SeeSuperNode()
	}
}

//给节点发送消息

func SendMsgAll_script(commands []string) {
	if len(commands) == 1 {
		core.SendMsgForAll("hello")
	}
	if len(commands) == 2 {
		core.SendMsgForAll(commands[1])
	}
	if len(commands) == 3 {
		core.SendMsgForOne_opt(commands[1], commands[2])
	}
}

//创建一个域名
//createdomain www.mandela.io []

func CreateDomain_script(commands []string) {
	args := make([]string, 3)
	args = append(args, commands[1:]...)
	core.CreateAccount(args[0], args[1], args[2])
}

//连接到anonymousnet网络
func Launcher() {
	core.StartService()
}
*/
