package main

import (
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/elazarl/goproxy"
)

const (
	ProxyAuthHeader  = "Proxy-Authorization"
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36"
)

func setBasicAuth(username, password string, req *http.Request) {
	if username != "" {
		req.Header.Set(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
	}
}

func basicAuth(username, password string) string {
	return b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func getTargetProxyURL(host string, port int) string {
	return "http://" + host + ":" + strconv.Itoa(port)
}

func setUpTargetProxy(config ProxyConfig, proxy *goproxy.ProxyHttpServer) {
	login := config.proxyLogin
	password := config.proxyPassword

	targetProxyURL := getTargetProxyURL(config.targetProxyHost, config.targetProxyPort)
	log.Println("Forwarding queries to proxy at ", targetProxyURL)
	if login != "" {
		fmt.Println("Using user account", login)
	}
	proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		setBasicAuth(login, password, req)

		return url.Parse(getTargetProxyURL(config.targetProxyHost, config.targetProxyPort))
	}
	connectReqHandler := func(req *http.Request) {
		setBasicAuth(login, password, req)
		if req.UserAgent() == "" {
			fmt.Println("No User-Agent found, using default ", DefaultUserAgent)
			req.Header.Set("User-Agent", DefaultUserAgent)
		}
	}
	proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(targetProxyURL, connectReqHandler)
	proxy.Verbose = config.logVerbose
}

func getProxyHandler(enableForward bool) (string, *goproxy.ProxyHttpServer) {
	proxyConfig := getProxyConfig()

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = proxyConfig.logVerbose

	if proxyConfig.targetProxyHost == "" || !enableForward {
		log.Println("No target proxy host defined or forward disabled. Will act as a simple proxy ...")
	} else {
		setUpTargetProxy(proxyConfig, proxy)
	}

	addr := proxyConfig.listenAddress + ":" + strconv.Itoa(proxyConfig.listenPort)
	log.Println("Will start proxy server on", addr)

	return addr, proxy
}
