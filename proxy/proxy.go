package proxy

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/tchabaud/simpleproxy/config"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	AuthHeader       = "Proxy-Authorization"
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36"
)

type Proxy struct {
	currentProxy *goproxy.ProxyHttpServer
	config       config.ProxyConfig
}

func setBasicAuth(username, password string, req *http.Request) {
	if username != "" {
		req.Header.Set(AuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
	}
}

func basicAuth(username, password string) string {
	return b64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

func getTargetProxyURL(host string, port int) string {
	return "http://" + host + ":" + strconv.Itoa(port)
}

// StartHTTPServer Start a new http server, and set forwarding according to boolean status.
func (p *Proxy) StartHTTPServer(forwarding bool) {
	const WriteTimeout = 90
	const ReadTimeout = 60
	if p.currentProxy != nil {
		log.Fatal("A proxy instance is already running. Can't start a new one !")
	}
	proxy := p.currentProxy
	if forwarding {
		targetProxyURL := getTargetProxyURL(p.config.TargetProxyHost, p	.config.TargetProxyPort)
		log.Println("Forwarding queries to proxy at ", targetProxyURL)
		login := p.config.ProxyLogin
		password := p.config.ProxyPassword
		if login != "" {
			fmt.Println("Using user account", login)
		}
		proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
			setBasicAuth(login, password, req)
			return url.Parse(getTargetProxyURL(p.config.TargetProxyHost, p.config.TargetProxyPort))
		}
		connectReqHandler := func(req *http.Request) {
			setBasicAuth(login, password, req)
			if req.UserAgent() == "" {
				fmt.Println("No User-Agent found, using default ", DefaultUserAgent)
				req.Header.Set("User-Agent", DefaultUserAgent)
			}
		}
		proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(targetProxyURL, connectReqHandler)
		proxy.Verbose = p.config.LogVerbose
	} else {
		connectReqHandler := func(req *http.Request) {
			if req.UserAgent() == "" {
				fmt.Println("No User-Agent found, overriding with default ", DefaultUserAgent)
				req.Header.Set("User-Agent", DefaultUserAgent)
			}
		}
		proxy.ConnectDial = proxy.OnRequest(goproxy.ReqCondition())
		proxy.Verbose = p.config.LogVerbose

	}
}

// StopHTTPServer Stop current http server instance
func (p *Proxy) StopHTTPServer() {
	if p.currentProxy == nil {
		log.Println("No http proxy instance currently running.")
		return
	}
	if err := p.currentProxy.Shutdown(context.Background()); err != nil {
		log.Panicf("Error: %s", err)
	}
}
func (p *Proxy) getProxyHandler() (string, *goproxy.ProxyHttpServer) {
	proxyConfig := getProxyConfig()

	goproxy.AlwaysMitm
	p.currentProxy.o := goproxy.NewProxyHttpServer()
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
