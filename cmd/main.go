package main

import (
	"fmt"
	"log"

	"github.com/MohOdejimi/ProxyCache/internal/cli"
	"github.com/MohOdejimi/ProxyCache/internal/proxy"
)

func main() {
	port, origin := cli.Flags()
	err := proxy.StartProxyServer(port, origin)
	if err != nil {
		log.Fatalf("Failed to start proxy server: %v", err)
	}
	fmt.Println(port, origin) 
}