package main

import (
	"fmt"
	"log"

	"github.com/MohOdejimi/ProxyCache/internal/cli"
	"github.com/MohOdejimi/ProxyCache/internal/proxy"
)

func main() {
	port, origin, shouldClear := cli.Flags()

	if shouldClear {
		proxy.ClearCache()
		log.Println("Cache cleared successfully")
		return
	}
	err := proxy.StartProxyServer(port, origin)
	if err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
	fmt.Println(port, origin) 
}