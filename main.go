package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

import b64 "encoding/base64"

const (
	ProxyAuthHeader = "Proxy-Authorization"
)

func SetBasicAuth(username, password string, req *http.Request) {
	if username != "" {
		fmt.Println("Setting credentials on query for user", username)
		req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
		fmt.Println("Credentials set on query for user", username)
	}
}

func basicAuth(username, password string) string {
	return b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func GetTargetProxyUrl(host string, port int) string {
	return "http://" + host + ":" + strconv.Itoa(port)
}

func main() {
	// Define command line args
	proxyHost := flag.String("host", "localhost",
		"Hostname or ip address used to listen for incoming connections.")
	proxyPort := flag.Int("port", 8080,
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

	targetProxyUrl := GetTargetProxyUrl(*targetProxyHost, *targetProxyPort)

	fmt.Println("Forwarding queries to proxy at " + targetProxyUrl)
	if proxyLogin != nil && *proxyLogin != "" {
		fmt.Println("Using user account", *proxyLogin)
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *isVerbose

	if targetProxyHost == nil || *targetProxyHost == "" {
		fmt.Println("No target proxy host defined. Will use localhost ...")
		*targetProxyHost = "localhost"
	}

	proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		SetBasicAuth(*proxyLogin, *proxyPassword, req)
		return url.Parse(GetTargetProxyUrl(*targetProxyHost, *targetProxyPort))
	}

	connectReqHandler := func(req *http.Request) {
		SetBasicAuth(*proxyLogin, *proxyPassword, req)
	}

	proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(targetProxyUrl, connectReqHandler)

	addr := *proxyHost + ":" + strconv.Itoa(*proxyPort)
	fmt.Println("Starting proxy server on", addr)
	log.Fatal(http.ListenAndServe(addr, proxy))
}
