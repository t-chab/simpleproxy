package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"

	b64 "encoding/base64"
)

const (
	ProxyAuthHeader = "Proxy-Authorization"
)

func setBasicAuth(username, password string, req *http.Request) {
	if username != "" {
		req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
	}
}

func basicAuth(username, password string) string {
	return b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func getTargetProxyUrl(host string, port int) string {
	return "http://" + host + ":" + strconv.Itoa(port)
}

func setUpTargetProxy(config ProxyConfig, proxy *goproxy.ProxyHttpServer) {
	var login string
	var password string
	if config.proxyLogin == "" {
		fmt.Println("Empty credentials, trying to automatically find them from ~/.netrc file ...")
		login, password = getProxyCredentials()
	}
	targetProxyUrl := getTargetProxyUrl(config.targetProxyHost, config.targetProxyPort)
	fmt.Println("Forwarding queries to proxy at " + targetProxyUrl)
	if login != "" {
		fmt.Println("Using user account", login)
	}
	proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		setBasicAuth(login, password, req)
		return url.Parse(getTargetProxyUrl(config.targetProxyHost, config.targetProxyPort))
	}
	connectReqHandler := func(req *http.Request) {
		setBasicAuth(login, password, req)
	}
	proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(targetProxyUrl, connectReqHandler)
}

func getProxyHandler() (string, *goproxy.ProxyHttpServer) {
	proxyConfig := getProxyConfig()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = proxyConfig.logVerbose

	if proxyConfig.targetProxyHost == "" {
		fmt.Println("No target proxy host defined. Will act as a simple proxy ...")
	} else {
		setUpTargetProxy(proxyConfig, proxy)
	}

	addr := proxyConfig.listenAddress + ":" + strconv.Itoa(proxyConfig.listenPort)
	fmt.Println("Will start proxy server on", addr)

	return addr, proxy
}

func main() {
	loadConfiguration()
	systray.Run(onReady, onExit)
}

func startHttpServer() *http.Server {
	addr, proxy := getProxyHandler()
	go func() {
		// returns ErrServerClosed on graceful close
		if err := http.ListenAndServe(addr, proxy); err != http.ErrServerClosed {
			log.Printf("Error: %s", err)
		}
	}()

	// returning reference so caller can call Shutdown()
	return &http.Server{}
}

func stopHttpServer(srv *http.Server) {
	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (pkill -2)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error during shutdown; %s", err)
	}
}

func onReady() {
	systray.SetIcon(getIcon("assets/simple-proxy.ico"))
	systray.SetTooltip("simple-proxy")

	srv := startHttpServer()

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
				srv = startHttpServer()
			case <-mStop.ClickedCh:
				stopHttpServer(srv)
			case <-mConfigure.ClickedCh:
				err := open.Run(getConfigFilePath())
				if err != nil {
					log.Fatal("Can't open configuration !", err)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				log.Println("Exiting now...")
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
