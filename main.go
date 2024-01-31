package main

import (
	"conalg/caesar"
	"conalg/config"
	"conalg/transport"
	"time"

	"github.com/gookit/slog"
)

// TODO add start/end times for requests to track
// done? remove channels from grpc clients, use ch per stream

func main() {
	cfg := config.NewConfig()

	transport, err := transport.NewGRPCTransport(cfg.Nodes, cfg.Port)
	if err != nil {
		slog.Fatal(err)
	}

	_ = caesar.NewCaesar(cfg, transport)

	slog.Debug(cfg)

	go func() {
		time.Sleep(1 * time.Second)
		transport.ConnectToNodes(cfg.Nodes)

		// transport.BroadcastFastPropose("test", 0, 1)
	}()

	err = transport.RunServer(cfg.Port)
	if err != nil {
		slog.Fatal(err)
	}
}
