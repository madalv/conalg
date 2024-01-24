package main

import (
	"conalg/config"
	"conalg/transport"
	"time"

	"github.com/gookit/slog"
)

func main() {
	// TODO init module with config & receptor

	cfg := config.NewConfig()

	transport, err := transport.NewGRPCTransport(cfg.Nodes, cfg.Port)
	if err != nil {
		slog.Fatal(err)
	}

	slog.Debug(cfg)

	go func() {
		time.Sleep(1 * time.Second)
		transport.ConnectToNodes(cfg.Nodes)

		transport.BroadcastFastPropose("test", 0, 1)
	}()

	transport.RunServer(cfg.Port)
}
