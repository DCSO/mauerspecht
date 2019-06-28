package main

import (
	"github.com/DCSO/mauerspecht"
	"github.com/DCSO/mauerspecht/server"

	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
)

func main() {
	var cfgfile string
	flag.StringVar(&cfgfile, "config", "mauerspecht.json", "Config file")
	flag.Parse()
	buf, err := ioutil.ReadFile(cfgfile)
	if err != nil {
		log.Fatalf("open config file %s: %v", cfgfile, err)
	}
	var serverConfig mauerspecht.Config
	if err := json.Unmarshal(buf, &serverConfig); err != nil {
		log.Fatalf("read config file %s: %v", cfgfile, err)
	}
	s, err := server.New(serverConfig)
	if err != nil {
		log.Fatalf("Error while initializing server: %v", err)
	}
	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)
	<-stop
	s.Close()
}
