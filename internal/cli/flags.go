package cli

import (
	"flag"
	"os"
	"fmt"
	"net/url"
)

func Flags() (int, *url.URL, bool) {
	port := flag.Int("port", 8080, "Port to run the proxy server on")
	origin := flag.String("origin", "", "Origin server to proxy to")
	clear := flag.Bool("clear", false, "Clear the cache and exit")

	flag.Parse()

	if *clear {
		return 0, nil, true
	}

	if *origin == "" {
		fmt.Println("Origin server must be specified")
		os.Exit(1)
	}

	parsedOrigin, err := url.Parse(*origin)

	if err != nil {
		fmt.Println("Failed to parse origin URL:", err)
		os.Exit(1)
	}

	if *port <= 0 || *port > 65535 {
		fmt.Println("Invalid port number")
		os.Exit(1)
	}
	return *port, parsedOrigin, *clear
}

