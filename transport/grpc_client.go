package transport

import (
	"conalg/pb"
	"context"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	pb.ConalgClient
	fastProposeChan chan *pb.FastPropose
	address         string
}

func newGRPCClient(addr string) (*grpcClient, error) {
	slog.Infof("Starting Client on %s", addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewConalgClient(conn)
	client.FastProposeStream(context.Background())

	c := &grpcClient{
		client,
		make(chan *pb.FastPropose),
		addr,
	}

	go func() {
		c.initFastProposeStream()
	}()

	return c, nil
}

func (c *grpcClient) initFastProposeStream() {
	slog.Infof("Initializing Fast Propose Stream on %s", c.address)

	stream, err := c.FastProposeStream(context.Background())
	if err != nil {
		slog.Fatal(err)
	}

	for msg := range c.fastProposeChan {
		err = stream.Send(msg)
		if err != nil {
			slog.Error(err)
		}
	}
}
