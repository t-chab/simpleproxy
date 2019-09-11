package main

import (
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	SuccessExitCode int = 0
)

func main() {
	systray.Run(onReady, onExit)
	loadConfiguration()
}

func startHttpServer() *http.Server {
	addr, proxy := getProxyHandler()
	srv := &http.Server{
		Addr:    addr,
		Handler: proxy,
	}

	// returns ErrServerClosed on graceful close
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Error: %s", err)
	}

	return srv
}

func configure() {
	err := open.Run(getConfigFilePath())
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
	systray.SetIcon(getIcon("assets/simple-proxy.ico"))
	systray.SetTooltip("simple-proxy")
	go startHttpServer()

	mConfigure := systray.AddMenuItem("Configure", "Configure simple-proxy")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit application")
	for {
		select {
		case <-mConfigure.ClickedCh:
			log.Printf("Configure option clicked.")
			configure()
		case <-mQuit.ClickedCh:
			log.Printf("Quit option clicked.")
			exitApp()
			return
		}
	}
}

func onExit() {
	// Cleaning stuff here.
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		log.Fatal(err)
	}
	return b
}
