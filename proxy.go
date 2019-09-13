package main

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"net/url"
	"strconv"
)

const ProxyAuthHeader = "Proxy-Authorization"

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
	login := config.proxyLogin
	password := config.proxyPassword

	targetProxyUrl := getTargetProxyUrl(config.targetProxyHost, config.targetProxyPort)
	log.Println("Forwarding queries to proxy at ", targetProxyUrl)
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
		log.Println("No target proxy host defined. Will act as a simple proxy ...")
	} else {
		setUpTargetProxy(proxyConfig, proxy)
	}

	addr := proxyConfig.listenAddress + ":" + strconv.Itoa(proxyConfig.listenPort)
	log.Println("Will start proxy server on", addr)

	return addr, proxy
}
