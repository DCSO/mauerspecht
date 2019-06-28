package main

import (
	"github.com/DCSO/mauerspecht/client"

	"flag"
	"log"
)

func main() {
	var serverURL string
	var proxyURL string
	flag.StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	flag.StringVar(&proxyURL, "proxy", "", "Proxy URL")
	flag.Parse()
	c, err := client.New(serverURL, proxyURL)
	if err != nil {
		log.Fatalf("Error while initializing client: %v", err)
	}
	c.Run()
}
