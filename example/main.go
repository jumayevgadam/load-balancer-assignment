package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/jumayevgadam/golb"
)

func main() {
	var level string
	flag.StringVar(&level, "level", "basic", "there are 3 levels for lb: basic, intermediate, advanced")
	flag.Parse()

	urls := []*url.URL{
		{Host: "amazon.com"},
		{Host: "google.com"},
		{Host: "localhost:8083"},
	}

	balancer := golb.NewBalancer(level, urls)
	if balancer == nil {
		log.Println("unknown balancer level")
		return
	}

	ctx := context.Background()
	var wg sync.WaitGroup

	// Send 50 requests to test load balancing.
	for i := 0; i < 50; i++ {
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			resp, err := balancer.Invoke(ctx, fmt.Sprintf("request-%d", id))
			if err != nil {
				log.Printf("Request %d failed: %v", id, err)
			} else {
				fmt.Printf("Request %d response: %v\n", id, resp)
			}
		}(i)
	}

	wg.Wait()

	switch b := balancer.(type) {
	case *golb.BasicBalancer:
		for i := 0; i < 6; i++ {
			fmt.Println(b.GetNextServer().Host())
		}
	case *golb.IntermediateBalancer:
		for i := 0; i < 6; i++ {
			fmt.Println(b.GetNextServer().Host())
		}
	case *golb.AdvancedBalancer:
		for i := 0; i < 6; i++ {
			if backend := b.GetNextServer(); backend != nil {
				fmt.Println(backend.Host())
			} else {
				fmt.Println("No healthy backend available")
			}
		}

		b.StopHealthChecker() // stop the health checker before exiting.
	}
}
