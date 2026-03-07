package proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func StartProxyServer(port int, origin *url.URL) error {  
	router := httprouter.New()

	addr := ":" + strconv.Itoa(port)
	fmt.Printf("Starting proxy server on %s, forwarding to %s\n", addr, origin)

	router.GET("/*path", ProxyHandler(origin))
	log.Fatal(http.ListenAndServe(addr, router))

	return nil
}