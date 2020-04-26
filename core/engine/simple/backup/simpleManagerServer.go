package main

import (
	// _ "./cxv"
	server "gameServerEngine"
)

func main() {
	server.ConfigPath = "config/managerServer.ini"
	server.StartUP()
}
