package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/getlantern/systray"
	"github.com/jdxcode/netrc"
	"github.com/skratchdot/open-golang/open"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

import b64 "encoding/base64"

const (
	ProxyAuthHeader = "Proxy-Authorization"
)

type ProxyConfig struct {
	proxyLogin      string
	proxyPassword   string
	proxyHost       string
	proxyPort       int
	targetProxyHost string
	targetProxyPort int
	logVerbose      bool
}

func getConfigFileName() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user :", err)
	}
	netrcFile := filepath.Join(usr.HomeDir, ".netrc")
	if _, err := os.Stat(netrcFile); os.IsNotExist(err) {
		fmt.Println("File not found in", usr.HomeDir, "a new file will be created.")
		template := []byte("machine [WRITE_HERE_TARGET_PROXY_HOST_NAME]\n\tlogin [WRITE_LOGIN_HERE]\n\tpassword [WRITE_PASSWORD_HERE]\n")
		err := ioutil.WriteFile(netrcFile, template, 0644)
		if err != nil {
			log.Fatal("Can't write credentials configuration file template.")
		}
	}

	return netrcFile
}

func getProxyConfig() ProxyConfig {
	// Define command line args
	proxyHost := flag.String("host", "localhost",
		"Hostname or ip address used to listen for incoming connections.")
	proxyPort := flag.Int("port", 8118,
		"TCP port used to listen for incoming connections.")

	targetProxyHost := flag.String("targetProxyHost", "",
		"Hostname or ip address of the target proxy where the queries will be forwarded.")
	targetProxyPort := flag.Int("targetProxyPort", 8080,
		"Port number of the target proxy where the queries will be forwarded.")

	proxyLogin := flag.String("proxyLogin", "",
		"Login to use for proxy auth.")
	proxyPassword := flag.String("proxyPassword", "",
		"Login to use for proxy auth.")

	isVerbose := flag.Bool("verbose", false, "Verbose logging. Default to false.")

	// Fetch command line args
	flag.Parse()

	return ProxyConfig{
		proxyLogin:      *proxyLogin,
		proxyPassword:   *proxyPassword,
		proxyHost:       *proxyHost,
		proxyPort:       *proxyPort,
		targetProxyHost: *targetProxyHost,
		targetProxyPort: *targetProxyPort,
		logVerbose:      *isVerbose,
	}
}

// See https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html
// for details about .netrc file format
func getNetRcCredentials(machine string) (string, string) {
	n, err := netrc.Parse(getConfigFileName())
	if err != nil {
		fmt.Println("Error parsing .netrc file :", err)
		return "", ""
	}
	login := n.Machine(machine).Get("login")
	password := n.Machine(machine).Get("password")
	fmt.Printf("Credentials loaded successfully for host %q\n", machine)
	return login, password
}

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
		login, password = getNetRcCredentials(config.targetProxyHost)
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

	addr := proxyConfig.proxyHost + ":" + strconv.Itoa(proxyConfig.proxyPort)
	fmt.Println("Will start proxy server on", addr)

	return addr, proxy
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("assets/simple-proxy.ico"))
	systray.SetTooltip("simple-proxy")

	addr, proxy := getProxyHandler()

	var srv http.Server

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
				log.Fatal(http.ListenAndServe(addr, proxy))
			case <-mStop.ClickedCh:
				if err := srv.Shutdown(context.Background()); err != nil {
					// Error from closing listeners, or context timeout:
					log.Printf("simple-proxy shutdown: %v", err)
				}
			case <-mConfigure.ClickedCh:
				err := open.Run(getConfigFileName())
				if err != nil {
					log.Fatal("Can't open configuration !")
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				fmt.Println("Exiting now...")
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
		fmt.Print(err)
	}
	return b
}
