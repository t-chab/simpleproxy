package main

//go:generate go get github.com/dim13/file2go
//go:generate file2go -in ./assets/simpleproxy.ico

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

const (
	SuccessExitCode int = 0
)

func main() {
	cmdLineFlags(NewDefaultValues())
	// to create file if it doesn't exists
	if !fileExists(getConfigFilePath()) {
		getProxyConfig()
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

// Start httpServer, and set forwarding according to boolean status.
func startHTTPServer(forwarding bool) *http.Server {
	addr, proxy := getProxyHandler(forwarding)
	const WriteTimeout = 90
	const ReadTimeout = 60
	srv := &http.Server{
		Addr:         addr,
		Handler:      proxy,
		ReadTimeout:  ReadTimeout * time.Second,
		WriteTimeout: WriteTimeout * time.Second,
	} // returns ErrServerClosed on graceful close
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Error: %s", err)
	}

	return srv
}

func stopHTTPServer(srv *http.Server) {
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
	srv := &http.Server{}
	for {
		select {

		case <-mConfigure.ClickedCh:

			log.Printf("Configure option clicked.")
			stopHTTPServer(srv)
			configure()
			newConfiguration := getProxyConfig()
			srv = startHTTPServer(newConfiguration.forwardingStatus)
			break

		case <-mForward.ClickedCh:

			forwardingStatus := mForward.Checked()
			if forwardingStatus {
				log.Printf("Disabling forwarding ...")
				mForward.Uncheck()
			} else {
				log.Printf("Enabling forwarding.")
				mForward.Check()
			}

			stopHTTPServer(srv)
			srv = startHTTPServer(forwardingStatus)
			break

		case <-mQuit.ClickedCh:

			log.Printf("Quit option clicked.")
			stopHTTPServer(srv)
			exitApp()
			break
		}
	}
}

func onExit() {
	// Cleaning stuff here.
}

func getIcon() []byte {
	return SimpleproxyIco
}
