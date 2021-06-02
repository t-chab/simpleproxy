package proxy

import (
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

func (p *Proxy) setUpTargetProxy() {
	login := p.config.ProxyLogin
	password := p.config.ProxyPassword

	proxy := p.currentProxy

	targetProxyURL := getTargetProxyURL(p.config.TargetProxyHost, p.config.TargetProxyPort)
	log.Println("Forwarding queries to proxy at ", targetProxyURL)
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
}

// StartHTTPServer Start a new http server, and set forwarding according to boolean status.
func (p *Proxy) StartHTTPServer(forwarding bool) {
	const WriteTimeout = 90
	const ReadTimeout = 60
	srv := &http.Server{
		Addr:         p.config.ListenAddress + ":" + strconv.Itoa(p.config.ListenPort),
		ReadTimeout:  ReadTimeout * time.Second,
		WriteTimeout: WriteTimeout * time.Second,
	} // returns ErrServerClosed on graceful close
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("Error: %s", err)
	}
}

// StopHTTPServer Stop current http server instance
func (p *Proxy) StopHTTPServer() {
	if p.currentProxy == nil {
		log.Println("No http proxy instance currently running.")
		return
	}
	/*
		if err := p.currentProxy.Shutdown(context.Background()); err != nil {
			log.Panicf("Error: %s", err)
		}
	*/
}

func (p *Proxy) GetProxyHandler(enableForward bool) {
	proxyConfig := config.GetProxyConfig()
	p.currentProxy = goproxy.NewProxyHttpServer()
	p.currentProxy.Verbose = proxyConfig.LogVerbose

	if proxyConfig.TargetProxyHost == "" || !enableForward {
		log.Println("No target proxy host defined or forward disabled. Will act as a simple proxy ...")
	} else {
		p.setUpTargetProxy()
	}

	addr := p.config.ListenAddress + ":" + strconv.Itoa(p.config.ListenPort)
	log.Println("Will start proxy server on", addr)
}
