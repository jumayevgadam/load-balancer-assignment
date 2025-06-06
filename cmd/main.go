package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/jumayevgadam/golb/loader"
	"github.com/jumayevgadam/golb/selector"
)

func main() {
	var level string
	flag.StringVar(&level, "level", "basic", "There are 3 level of load-balancer, basic, intermediate, advanced")
	flag.Parse()

	backendServers := []*url.URL{
		{Host: "localhost:8081"},
		{Host: "localhost:8082"},
		{Host: "localhost:8083"},
		{Host: "localhost:8084"},
		// {Host: "localhost:8085"},
		// {Host: "localhost:8086"},
	}

	balancer := loader.NewBalancer(level, backendServers)
	if balancer == nil {
		log.Println("unknown balancer level")
		return
	}

	switch b := balancer.(type) {
	case *selector.BasicBalancer:
		for i := 0; i < 6; i++ {
			fmt.Println(b.NextServer().Host())
		} // also will come here other levels.
	}
}
