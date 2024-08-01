package proxymanager

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"proxymanager/pkg/networksetup"
	"proxymanager/pkg/v2ray"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type Proxy struct {
	NetworkServiceName string
	On                 bool
}

type ProxyManager struct {
	AppName       string
	AppDir        string
	Configuration *Configuration
	Logger        *slog.Logger
	LoggerFile    *os.File
	V2rayCmd      *exec.Cmd
	Proxies       []Proxy
}

type Configuration struct {
	V2rayDir    string `json:"v2ray_dir"`
	V2rayStdOut string `sjon:"v2ray_stdout"`
	V2rayStdErr string `sjon:"v2ray_stderr"`
}

func (conf *Configuration) V2rayConfFilePath() string {
	return filepath.Join(conf.V2rayDir, "config.json")
}

func NewApp(appName string) *ProxyManager {
	return &ProxyManager{AppName: appName}
}

func (manager *ProxyManager) Close() {
	defer func() {
		if manager.LoggerFile != nil {
			_ = manager.LoggerFile.Sync()
			_ = manager.LoggerFile.Close()
		}
	}()

	manager.LogInfo("closing proxy manager")
	if manager.V2rayCmd != nil {
		if err := manager.StopV2ray(); err != nil {
			manager.LogErr("closed proxy manager", err)
		}
	}

	for idx := range manager.Proxies {
		proxy := &manager.Proxies[idx]
		if err := ToggleOffSocksProxy(proxy.NetworkServiceName); err != nil {
			manager.LogErr("toggle off socks proxy", err)
		} else {
			proxy.On = false
		}
	}
	manager.LogInfo("closed proxy manager")

}

func (manager *ProxyManager) Init() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".proxymanager")

	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return err
	}
	manager.AppDir = configDir

	logFilePath := filepath.Join(configDir, "log.log")

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return err
	}
	manager.LoggerFile = logFile

	logger := slog.New(slog.NewTextHandler(logFile, nil))
	manager.Logger = logger

	if err := manager.LoadConfiguration(); err != nil {
		return err
	}

	if err := manager.LoadNetworkServices(); err != nil {
		return err
	}

	return nil
}

func LoadConfiguration(configFile string) (*Configuration, error) {
	configuration := Configuration{}

	if _, err := os.Stat(configFile); err == nil {
		if bytes, err := os.ReadFile(configFile); err != nil {
			return nil, err
		} else {
			if err := json.Unmarshal(bytes, &configuration); err != nil {
				return nil, err
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return &configuration, nil
}

func (manager *ProxyManager) LoadConfiguration() error {
	configuration, err := LoadConfiguration(manager.ConfigurationFilePath())
	if err != nil {
		return err
	}

	if configuration.V2rayStdErr == "" {
		configuration.V2rayStdErr = filepath.Join(manager.AppDir, "stderr.log")
	}

	if configuration.V2rayStdOut == "" {
		configuration.V2rayStdOut = filepath.Join(manager.AppDir, "stdout.log")
	}

	manager.Configuration = configuration

	return nil
}

func (manager *ProxyManager) LoadNetworkServices() error {
	networkServices, err := networksetup.ListallNetworkServices()
	if err != nil {
		return err
	}
	for _, ns := range networkServices {
		manager.Proxies = append(manager.Proxies, Proxy{
			NetworkServiceName: ns,
			On:                 false,
		})
	}

	return nil
}

func (manager *ProxyManager) LogErr(msg string, err error) {
	manager.Logger.Error(msg, slog.Any("err", err))
}

func (manager *ProxyManager) LogInfo(msg string, args ...any) {
	manager.Logger.Error(msg, args...)
}

func (manager *ProxyManager) ConfigurationFilePath() string {
	return filepath.Join(manager.AppDir, "proxymanager.json")
}

func (manager *ProxyManager) SaveConfiguration() error {
	filePath := manager.ConfigurationFilePath()
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := json.Marshal(manager.Configuration)
	if err != nil {
		return err
	}
	if _, err := file.Write(bytes); err != nil {
		return err
	}
	return nil
}

func (manager *ProxyManager) Run() {
	a := app.New()
	mainWindow := a.NewWindow(manager.AppName)
	mainWindow.Resize(fyne.NewSize(600, 600))

	if err := manager.Init(); err != nil {
		errDialog := dialog.NewError(err, mainWindow)
		errDialog.Resize(fyne.NewSize(200, 200))
		errDialog.Show()
		errDialog.SetOnClosed(func() {
			os.Exit(1)
		})
		mainWindow.ShowAndRun()
	}

	v2rayDirectoryEntry := widget.NewEntry()
	v2rayDirectoryEntry.SetText(manager.Configuration.V2rayDir)

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "V2ray Directory", Widget: v2rayDirectoryEntry}},
		OnSubmit: func() {
			manager.Configuration.V2rayDir = v2rayDirectoryEntry.Text
			if err := manager.SaveConfiguration(); err != nil {
				manager.LogErr("save configuration", err)
			}
			mainWindow.Hide()
		},
		SubmitText: "Save",
	}

	layout := container.NewVBox(form)
	mainWindow.SetContent(layout)
	mainWindow.Resize(fyne.NewSize(600, 0))

	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu(manager.AppName)

		var refreshMenuItems func()

		refreshMenuItems = func() {
			items := []*fyne.MenuItem{
				fyne.NewMenuItem("Configure", func() {
					mainWindow.Show()
				}),
				fyne.NewMenuItemSeparator(),
			}
			if manager.V2rayCmd == nil {
				items = append(items, fyne.NewMenuItem("V2ray", func() {
					if err := manager.ExecV2ray(); err != nil {
						manager.Logger.Error("start v2ray err: " + err.Error())
					}
					refreshMenuItems()
				}))
			} else {
				items = append(items, fyne.NewMenuItem("✓ V2ray", func() {
					if err := manager.StopV2ray(); err != nil {
						manager.Logger.Error("stop v2ray err: " + err.Error())
					}
					refreshMenuItems()
				}))
			}

			if len(manager.Proxies) > 0 {
				items = append(items, fyne.NewMenuItemSeparator())
			}
			for idx := range manager.Proxies {
				proxy := &manager.Proxies[idx]
				if proxy.On {
					itemName := fmt.Sprintf("✓ %s", proxy.NetworkServiceName)
					items = append(items, fyne.NewMenuItem(itemName, func() {
						if err := ToggleOffSocksProxy(proxy.NetworkServiceName); err != nil {
							manager.LogErr("toggle off proxy error: ", err)
						} else {
							proxy.On = false
						}
						refreshMenuItems()
					}))
				} else {
					itemName := proxy.NetworkServiceName

					items = append(items, fyne.NewMenuItem(itemName, func() {
						conf, err := LoadV2rayConf(manager.Configuration.V2rayConfFilePath())
						if err != nil {
							manager.LogErr("load v2ray conf", err)
						}

						sockeProxy := conf.FindSocksProxy()
						if sockeProxy == nil {
							manager.Logger.Error("No socks proxy configured")
						} else {
							if err := ToggleOnSocksProxy(proxy.NetworkServiceName, sockeProxy.Listen, sockeProxy.Port); err != nil {
								manager.LogErr("enable proxy error:", err)
							} else {
								proxy.On = true
							}
						}
						refreshMenuItems()
					}))
				}
			}

			items = append(items, fyne.NewMenuItemSeparator())

			menu.Items = items

			menu.Refresh()
			desk.SetSystemTrayMenu(menu)
		}

		refreshMenuItems()
	}

	mainWindow.SetCloseIntercept(func() {
		mainWindow.Hide()
	})
	mainWindow.ShowAndRun()
}

func (manager *ProxyManager) ExecV2ray() error {
	dir := manager.Configuration.V2rayDir
	programName := "v2ray"
	cmdPath := filepath.Join(dir, programName)
	cmd := exec.Command(cmdPath)
	cmd.Dir = dir

	stdoutFilePath := manager.Configuration.V2rayStdOut
	stdoutFile, err := os.OpenFile(stdoutFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return err
	}
	cmd.Stdout = stdoutFile

	stderrFilePath := manager.Configuration.V2rayStdErr
	stderrFile, err := os.OpenFile(stderrFilePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return err
	}
	cmd.Stderr = stderrFile

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		defer stdoutFile.Close()
		defer stderrFile.Close()
		if err := cmd.Wait(); err != nil {
			manager.LogErr("cmd Wait", err)
		}
	}()

	manager.V2rayCmd = cmd
	return nil
}

func (manager *ProxyManager) StopV2ray() error {
	if err := manager.V2rayCmd.Process.Kill(); err != nil {
		return err
	}
	manager.V2rayCmd = nil
	return nil
}

func LoadV2rayConf(path string) (*v2ray.Configuration, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	conf := v2ray.Configuration{}

	if err := json.Unmarshal(bytes, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func ToggleOnSocksProxy(networkService string, domain string, port int) error {
	if err := networksetup.SetSocksFirewallProxy(networkService, domain, port); err != nil {
		return err
	}

	if err := networksetup.SetSocksFirewallProxyState(networkService, true); err != nil {
		return err
	}

	return nil
}

func ToggleOffSocksProxy(networkService string) error {
	if err := networksetup.SetSocksFirewallProxyState(networkService, false); err != nil {
		return err
	}

	return nil
}
