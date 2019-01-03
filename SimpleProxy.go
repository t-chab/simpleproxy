package main

import (
	"flag"
	"fmt"
	"github.com/elazarl/goproxy"
	"log"
	"net/http"
	"strconv"
)

import b64 "encoding/base64"

func main() {
	// Define command line args
	proxyHost := flag.String("host", "localhost",
		"Hostname or ip address used to listen for incoming connections.")
	proxyPort := flag.Int("port", 8080,
		"TCP port used to listen for incoming connections.")
	proxyLogin := flag.String("proxyLogin", "",
		"Login to use for proxy auth.")
	proxyPassword := flag.String("proxyPassword", "",
		"Login to use for proxy auth.")
	isVerbose := flag.Bool("verbose", false, "Verbose logging. Default to false.")
	// Fetch command line args
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *isVerbose

	if proxyLogin != nil && *proxyLogin != "" {
		credentials := b64.StdEncoding.EncodeToString([]byte(*proxyLogin + ":" + *proxyPassword))
		proxy.OnRequest().DoFunc(
			func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
				r.Header.Set("Proxy-Authorization", credentials)
				return r, nil
			})
	}
	addr := *proxyHost + ":" + strconv.Itoa(*proxyPort)
	fmt.Println("Starting proxy server on", addr)
	log.Fatal(http.ListenAndServe(addr, proxy))
}
