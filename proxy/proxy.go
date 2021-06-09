package proxy

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"github.com/elazarl/goproxy"
	"github.com/inconshreveable/go-vhost"

	"github.com/tchabaud/simpleproxy/config"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

const (
	AuthHeader       = "Proxy-Authorization"
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/78.0.3904.97 Safari/537.36"
)

type dumbResponseWriter struct {
	net.Conn
}

func (dumb dumbResponseWriter) Header() http.Header {
	panic("Header() should not be called on this ResponseWriter")
}

func (dumb dumbResponseWriter) Write(buf []byte) (int, error) {
	if bytes.Equal(buf, []byte("HTTP/1.0 200 OK\r\n\r\n")) {
		return len(buf), nil // throw away the HTTP OK response from the faux CONNECT request
	}
	return dumb.Conn.Write(buf)
}

func (dumb dumbResponseWriter) WriteHeader(code int) {
	panic("WriteHeader() should not be called on this ResponseWriter")
}

type Proxy struct {
	listener net.Listener
	config   config.ProxyConfig
}

// ResetListener Start a new http server, and set forwarding according to boolean status.
func (p *Proxy) ResetListener(forwarding bool) {
	if p.listener != nil {
		log.Fatal("A proxy instance is already running. Can't start a new one !")
	}
	if forwarding {
		p.listener = p.StartForwardingProxy()
	} else {
		p.listener = p.StartStandaloneProxy()
	}
}

func (p *Proxy) StartStandaloneProxy() net.Listener {
	var proxy *goproxy.ProxyHttpServer
	ln, err := net.Listen("tcp", net.JoinHostPort(p.config.ListenAddress, string(rune(p.config.ListenPort))))
	if err != nil {
		log.Fatalf("Error listening for http connections - %v", err)
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			log.Printf("Error accepting new connection - %v", err)
			continue
		}
		go func(c net.Conn) {
			tlsConn, err := vhost.TLS(c)
			if err != nil {
				log.Printf("Error accepting new connection - %v", err)
			}
			if tlsConn.Host() == "" {
				log.Printf("Cannot support non-SNI enabled clients")
				return
			}
			connectReq := &http.Request{
				Method: "CONNECT",
				URL: &url.URL{
					Opaque: tlsConn.Host(),
					Host:   net.JoinHostPort(tlsConn.Host(), "443"),
				},
				Host:       tlsConn.Host(),
				Header:     make(http.Header),
				RemoteAddr: c.RemoteAddr().String(),
			}
			resp := dumbResponseWriter{tlsConn}
			proxy.ServeHTTP(resp, connectReq)
		}(c)
	}
	return ln
}

func (p *Proxy) StartForwardingProxy() net.Listener {
	proxy := goproxy.NewProxyHttpServer()
	targetProxyURL := getTargetProxyURL(p.config.TargetProxyHost, p.config.TargetProxyPort)
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
	l, err := net.Listen("tcp", net.JoinHostPort(p.config.ListenAddress, string(rune(p.config.ListenPort))))
	if err != nil {
		log.Fatalf("Error listening for http connections - %v", err)
	}
	err = http.Serve(l, proxy)
	if err != nil {
		log.Fatalf("Can't start new server - %v", err)
	}
	return l
}

// StopProxy Stop current http server instance
func (p *Proxy) StopProxy() {
	if p.listener == nil {
		log.Println("No http proxy instance currently running.")
		return
	}
	err := p.listener.Close()
	if err != nil {
		log.Fatalf("Can't close properly the server - %v", err)
	}
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
