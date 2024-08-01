package main

import (
	"proxymanager/pkg/proxymanager"
)

const AppName = "Proxy Manager"

func main() {
	manager := proxymanager.NewApp(AppName)
	defer manager.Close()
	manager.Run()
}
