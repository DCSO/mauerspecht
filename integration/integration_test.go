package integration

import (
	"github.com/DCSO/mauerspecht"
	"github.com/DCSO/mauerspecht/client"
	"github.com/DCSO/mauerspecht/server"

	"fmt"
	"testing"
)

func TestIntegration(t *testing.T) {
	var (
		c   *client.Client
		s   *server.Server
		err error
	)
	config := mauerspecht.Config{Hostname: "localhost"}
	for i := 10000; i < 40000; i++ {
		config.HTTPPorts = []int{i}
		s, err = server.New(config)
		if err == nil {
			break
		}
		s.Close()
	}
	if s == nil {
		t.Fatal("could not start server")
	}
	c, err = client.New(fmt.Sprintf("http://%s:%d/", config.Hostname, config.HTTPPorts[0]), "")
	if err != nil {
		t.Fatalf("could not create client: %v", err)
	}
	c.Run()
}
