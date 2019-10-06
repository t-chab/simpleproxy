package main

import (
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"github.com/therecipe/qt/widgets"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	SuccessExitCode int = 0
)

func main() {
	if true {
		// needs to be called once before you can start using the QWidgets
		app := widgets.NewQApplication(len(os.Args), os.Args)

		window := buildSettingsWindow()

		// make the window visible
		window.Show()

		// start the main Qt event loop
		// and block until app.Exit() is called
		// or the window is closed by the user
		app.Exec()
	}
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
