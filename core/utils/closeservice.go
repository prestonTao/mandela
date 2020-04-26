package utils

var stopservice = make(chan bool, 1)

/*
	创建一个等待关机命令服务
*/
func GetStopService() chan bool {
	return stopservice
}

/*
	发送停止命令
*/
func StopService() {
	select {
	case stopservice <- false:
	default:
	}
}
