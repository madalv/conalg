package main

import (
	"conalg/config"
	"conalg/transport"
	"net"

	"github.com/gookit/slog"
)

func main() {
	// TODO init module with config & receptor

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Fatal(err)
	}

	listener, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		slog.Fatal(err)
	}

	// conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	slog.Fatal(err)
	// }

	// defer conn.Close()

	// client := pb.NewConalgClient(conn)
	// client.FastProposeStream()

	srv := transport.NewGRPCServer()
	if err := srv.Serve(listener); err != nil {
		slog.Fatal(err)
	}
}
