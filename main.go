package main

//go:generate go get github.com/dim13/file2go
//go:generate file2go -in ./assets/simpleproxy.ico

import (
	"context"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	SuccessExitCode int = 0
)

func main() {
	cmdLineFlags(NewDefaultValues())
	systray.Run(onReady, onExit)
}

// Start httpServer, and set forwarding according to boolean status.
func startHttpServer(forwarding bool) *http.Server {
	addr, proxy := getProxyHandler(forwarding)
	var srv = &http.Server{
		Addr:         addr,
		Handler:      proxy,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 90 * time.Second,
	} // returns ErrServerClosed on graceful close
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Error: %s", err)
	}

	return srv
}

func stopHttpServer(srv *http.Server) {
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Panicf("Error: %s", err)
	}
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
	systray.SetIcon(getIcon())
	systray.SetTooltip("simple-proxy")
	mConfigure := systray.AddMenuItem("Configure", "Configure simple-proxy")
	systray.AddSeparator()
	mForward := systray.AddMenuItem("Forward", "Reload config and restart simple-proxy")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit application")
	var srv = &http.Server{}
	for {
		select {
		case <-mConfigure.ClickedCh:
			log.Printf("Configure option clicked.")
			configure()
		case <-mForward.ClickedCh:
			forwardingStatus := mForward.Checked()
			if forwardingStatus {
				log.Printf("Disabling forwarding ...")
				mForward.Uncheck()
			} else {
				log.Printf("Enabling forwarding.")
				mForward.Check()
			}
			stopHttpServer(srv)
			srv = startHttpServer(false)
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

func getIcon() []byte {
	return SimpleproxyIco
}
