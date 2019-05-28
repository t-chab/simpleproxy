package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/jdxcode/netrc"
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

// See https://www.gnu.org/software/inetutils/manual/html_node/The-_002enetrc-file.html
// for details about .netrc file format
func getNetRcCredentials(machine *string) (*string, *string) {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Error getting current user :", err)
	}
	netrcFile := filepath.Join(usr.HomeDir, ".netrc")
	if _, err := os.Stat(netrcFile); os.IsNotExist(err) {
		fmt.Println("Skipping because .netrc file does not exists in", usr.HomeDir)
		empty := ""
		return &empty, &empty
	} else {
		n, err := netrc.Parse(netrcFile)
		if err != nil {
			fmt.Println("Error parsing .netrc file :", err)
		}
		login := n.Machine(*machine).Get("login")
		password := n.Machine(*machine).Get("password")
		fmt.Printf("Credentials loaded successfully for host %q\n", *machine)
		return &login, &password
	}
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

func setUpTargetProxy(login *string, password *string, targetHost *string, targetPort *int, proxy *goproxy.ProxyHttpServer) {
	if login == nil || *login == "" {
		fmt.Println("Empty credentials, trying to automatically find them from ~/.netrc file ...")
		login, password = getNetRcCredentials(targetHost)
	}
	targetProxyUrl := getTargetProxyUrl(*targetHost, *targetPort)
	fmt.Println("Forwarding queries to proxy at " + targetProxyUrl)
	if login != nil && *login != "" {
		fmt.Println("Using user account", *login)
	}
	proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		setBasicAuth(*login, *password, req)
		return url.Parse(getTargetProxyUrl(*targetHost, *targetPort))
	}
	connectReqHandler := func(req *http.Request) {
		setBasicAuth(*login, *password, req)
	}
	proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(targetProxyUrl, connectReqHandler)
}

func main() {
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

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *isVerbose

	if targetProxyHost == nil || *targetProxyHost == "" {
		fmt.Println("No target proxy host defined. Will act as a simple proxy ...")
	} else {
		setUpTargetProxy(proxyLogin, proxyPassword, targetProxyHost, targetProxyPort, proxy)
	}

	addr := *proxyHost + ":" + strconv.Itoa(*proxyPort)
	fmt.Println("Starting proxy server on", addr)
	log.Fatal(http.ListenAndServe(addr, proxy))
}
