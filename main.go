package main

import (
	"context"
	"log"

	// _ "net/http/pprof"
	"os"

	"github.com/mingoooo/tail-based-sampling/agent"
	"github.com/mingoooo/tail-based-sampling/collector"
)

var (
	AgentCfg = agent.Cfg{
		CacheLen:  3.5 * 1000 * 1000,
		ReadLimit: 512 * 1024 * 1024,
	}
	CollectorCfg = collector.Cfg{
		RpcPort: "8003",
		Agents:  []string{"8000", "8001"},
	}
)

func main() {
	httpPort := os.Getenv("SERVER_PORT")
	// f, err := os.Create(fmt.Sprintf("trace_%s.out", httpPort))
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	switch httpPort {
	case "8000":
		cfg := AgentCfg
		cfg.HttpPort = httpPort
		cfg.DataSuffix = "1"
		a, err := agent.New(&cfg)
		if err != nil {
			log.Fatal(err)
		}
		if err := a.Run(ctx, cancel); err != nil {
			log.Fatal(err)
		}
	case "8001":
		cfg := AgentCfg
		cfg.HttpPort = httpPort
		cfg.DataSuffix = "2"
		a, err := agent.New(&cfg)
		if err != nil {
			log.Fatal(err)
		}
		if err := a.Run(ctx, cancel); err != nil {
			log.Fatal(err)
		}
	case "8002":
		cfg := CollectorCfg
		cfg.HttpPort = httpPort
		c := collector.New(&cfg)
		if err := c.Run(ctx, cancel); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("Port number not match")
	}
}
