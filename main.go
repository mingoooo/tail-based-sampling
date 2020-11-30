package main

import (
	"context"
	"log"
	"os"

	"github.com/mingoooo/tail-based-sampling/agent"
	"github.com/mingoooo/tail-based-sampling/collector"
)

var (
	httpPort string
)

func main() {
	httpPort := os.Getenv("SERVER_PORT")

	ctx, cancel := context.WithCancel(context.Background())
	switch httpPort {
	case "8000":
		a, err := agent.New(httpPort, "1")
		if err != nil {
			log.Fatal(err)
		}
		if err := a.Run(ctx, cancel); err != nil {
			log.Fatal(err)
		}
	case "8001":
		a, err := agent.New(httpPort, "2")
		if err != nil {
			log.Fatal(err)
		}
		if err := a.Run(ctx, cancel); err != nil {
			log.Fatal(err)
		}
	case "8002":
		c := collector.New(httpPort, "8003", []string{"8000", "8001"})
		if err := c.Run(ctx, cancel); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Port number not match")
	}
}
