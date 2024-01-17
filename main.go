package main

import (
	"conalg/pb"
	"conalg/transport"
	"net"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// TODO init module with config & receptor

	listener, err := net.Listen("tcp", "[::]:50000")
	if err != nil {
		slog.Fatal(err)
	}

	// TODO read out of config
	conn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Fatal(err)
	}

	defer conn.Close()

	client := pb.NewConalgClient(conn)
	// client.FastProposeStream()

	srv := transport.NewGRPCServer()
	if err := srv.Serve(listener); err != nil {
		slog.Fatal(err)
	}
}
