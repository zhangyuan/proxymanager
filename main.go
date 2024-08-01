package main

import (
	"proxymanager/pkg/proxymanager"
)

const AppName = "Proxy Helper"

func main() {
	manager := proxymanager.NewApp(AppName)
	defer manager.Close()
	manager.Run()
}
