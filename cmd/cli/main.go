package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/tchabaud/simpleproxy/config"
	"github.com/tchabaud/simpleproxy/proxy"
)

//go:embed assets/simpleproxy.ico
var SimpleProxyIco []byte

//go:embed assets/simpleproxy.png
var SimpleProxyPng []byte

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
		now := time.Now()
		ioutil.WriteFile(fmt.Sprintf(`on_exit_%d.txt`, now.UnixNano()), []byte(now.String()), 0644)
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
	mQuit := systray.AddMenuItem("Quit", "Exit application")

	systray.AddSeparator()
	mConfigure := systray.AddMenuItem("Configure", "Configure simpleproxy")

	systray.AddSeparator()
	mForward := systray.AddMenuItem("Forward", "Reload config and restart simpleproxy")
	go func() {
		for {
			select {
			case <-mConfigure.ClickedCh:
				log.Printf("Configure option clicked.")
				stopHTTPServer(srv)
				configure()
				newConfiguration := config.GetProxyConfig()
				srv = startHTTPServer(newConfiguration.ForwardingStatus)

			case <-mForward.ClickedCh:
				if mForward.Checked() {
					log.Printf("Disabling forwarding ...")
					mForward.Uncheck()
				} else {
					log.Printf("Enabling forwarding.")
					mForward.Check()
				}

			case <-mQuit.ClickedCh:
				log.Printf("Quit option clicked.")
				stopHTTPServer(srv)
				exitApp()

				stopHTTPServer(srv)
				srv = startHTTPServer(mForward.Checked())
			}
		}
	}()
}

func onExit() {
	// Cleaning stuff here.
}
