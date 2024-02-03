package main

import (
	"conalg/caesar"
	"conalg/config"
	"conalg/transport"
	"time"

	"github.com/gookit/slog"
)

// TODO add start/end times for requests to track


/*
Qs:
- what happens if stable msg not received by all nodes?
- how much should the timeout for fast propose be? what exactly happens if it times out?
*/

func main() {
	cfg := config.NewConfig()

	transport, err := transport.NewGRPCTransport(cfg)
	if err != nil {
		slog.Fatal(err)
	}

	_ = caesar.NewCaesar(cfg, transport)

	slog.Debug(cfg)

	go func() {
		time.Sleep(1 * time.Second)
		transport.ConnectToNodes()
	}()

	err = transport.RunServer()
	if err != nil {
		slog.Fatal(err)
	}
}
