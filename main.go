package main

import (
	"proxy-manager/pkg/proxymanager"
)

const AppName = "Proxy Helper"

func main() {
	manager := proxymanager.NewApp(AppName)
	defer manager.Close()
	manager.Run()
}
