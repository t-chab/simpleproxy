package main

import (
	"context"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
)

const (
	SUCCESS_EXIT_CODE int = 0
)

func main() {
	systray.Run(onReady, onExit)
	loadConfiguration()
}

func startHttpServer(stopSignal chan os.Signal) {
	addr, proxy := getProxyHandler()
	srv := &http.Server{
		Addr:    addr,
		Handler: proxy,
	}

	// returns ErrServerClosed on graceful close
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Error: %s", err)
	}

	<-stopSignal
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Printf("Error during shutdown; %v", err)
	}
	log.Printf("Http server stopped.")
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
	os.Exit(SUCCESS_EXIT_CODE)
}

func onReady() {
	systray.SetIcon(getIcon("assets/simple-proxy.ico"))
	systray.SetTooltip("simple-proxy")

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt)

	go func() {
		mStart := systray.AddMenuItem("Start", "Launch simple-proxy")
		mStop := systray.AddMenuItem("Stop", "Stop simple-proxy")
		systray.AddSeparator()
		mConfigure := systray.AddMenuItem("Configure", "Configure simple-proxy")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("Quit", "Exit application")
		for {
			select {
			case <-mStart.ClickedCh:
				startHttpServer(stopSignal)
			case <-mStop.ClickedCh:
				stopMsg := <-stopSignal
				log.Printf("Stop signal received: ", stopMsg)
			case <-mConfigure.ClickedCh:
				configure()
			case <-mQuit.ClickedCh:
				exitApp()
				return
			}
		}
	}()
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
