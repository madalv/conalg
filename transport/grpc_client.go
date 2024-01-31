package transport

import (
	"conalg/caesar"
	"conalg/pb"
	"context"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	pb.ConalgClient
	fastProposeStream pb.Conalg_FastProposeStreamClient
	address           string
}

func newGRPCClient(addr string) (*grpcClient, error) {
	slog.Infof("Starting Client on %s", addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewConalgClient(conn)
	client.FastProposeStream(context.Background())

	slog.Infof("Initializing Fast Propose Stream on %s", addr)
	fpStream, err := client.FastProposeStream(context.Background())
	if err != nil {
		slog.Fatal(err)
	}

	c := &grpcClient{
		client,
		fpStream,
		addr,
	}

	return c, nil
}

func (c *grpcClient) sendFastProposeStream(req *caesar.Request) {
	err := c.fastProposeStream.Send(req.ToFastProposePb())
	if err != nil {
		slog.Error(err)
	}
}
