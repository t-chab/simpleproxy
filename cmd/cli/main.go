package main

import (
	_ "embed"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"github.com/tchabaud/simpleproxy/config"
	"github.com/tchabaud/simpleproxy/proxy"
	"log"
	"os"
	"strconv"
)

//go:embed assets/simpleproxy.ico
var SimpleProxyIco []byte

//go:embed assets/simpleproxy.png
var SimpleProxyPng []byte

var proxyInstance proxy.Proxy

const (
	SuccessExitCode int = 0
)

func main() {
	config.CmdLineFlags(config.NewDefaultValues())
	// to create file if it doesn't exists
	if !fileExists(config.GetConfigFilePath()) {
		config.GetProxyConfig()
	}
	onExit := func() {
		log.Println("Thanks for using simple proxy !")
	}
	isForwardingEnabled, err := strconv.ParseBool(config.ForwardingStatus)
	if err != nil {
		isForwardingEnabled = false
	}
	if isForwardingEnabled {
		proxyInstance.StartForwardingProxy()
	} else {
		proxyInstance.StartStandaloneProxy()
	}
	systray.Run(onReady, onExit)
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func configure() {
	err := open.Run(config.GetConfigFilePath())
	if err != nil {
		log.Fatal("Can't open configuration !", err)
	}
}

func exitApp() {
	systray.Quit()
	log.Println("Exiting now...")
	os.Exit(SuccessExitCode)
}

func onReady() {
	systray.SetIcon(SimpleProxyIco)
	systray.SetTemplateIcon(SimpleProxyPng, SimpleProxyIco)
	systray.SetTooltip("Simple Proxy")

	mForward := systray.AddMenuItem("Forward", "Reload config and restart simpleproxy")

	systray.AddSeparator()
	mConfigure := systray.AddMenuItem("Configure", "Configure simpleproxy")

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit application")

	go func() {
		for {
			select {
			case <-mConfigure.ClickedCh:
				log.Printf("Configure option clicked.")
				proxyInstance.StopProxy()
				configure()
				newConfiguration := config.GetProxyConfig()
				proxyInstance.ResetListener(newConfiguration.ForwardingStatus)

			case <-mForward.ClickedCh:
				if mForward.Checked() {
					log.Printf("Disabling forwarding ...")
					mForward.Uncheck()
					mForward.SetTitle("Forward")
				} else {
					log.Printf("Enabling forwarding.")
					mForward.Check()
					mForward.SetTitle("✅ Forward")
				}

			case <-mQuit.ClickedCh:
				log.Printf("Quit option clicked.")
				proxyInstance.StopProxy()
				exitApp()
			}
		}
	}()
}
