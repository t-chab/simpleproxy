package main

import (
	"github.com/getlantern/systray"
	"github.com/gobuffalo/packr/v2"
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
	systray.Run(onReady, onExit)
	loadConfiguration()
}

func startHttpServer() *http.Server {
	addr, proxy := getProxyHandler()
	srv := &http.Server{
		Addr:         addr,
		Handler:      proxy,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 90 * time.Second,
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
	systray.SetIcon(getIcon())
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

func getIcon() []byte {
	// set up a new box by giving it a (relative) path to a folder on disk:
	box := packr.New("assets", "./assets")

	// Get the []byte representation of a file, or an error if it doesn't exist:
	ico, err := box.Find("simple-proxy.ico")
	if err != nil {
		log.Fatal(err)
	}
	return ico
}
