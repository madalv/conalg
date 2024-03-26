package transport

import (
	"conalg/model"
	"conalg/pb"
	"context"
	"io"
	"time"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	pb.ConalgClient
	fastProposeStream   pb.Conalg_FastProposeStreamClient
	stableProposeStream pb.Conalg_StableStreamClient
	retryProposeStream  pb.Conalg_RetryStreamClient
	address             string
	receiver            Receiver
	self                bool
}

func newGRPCClient(addr string, rec Receiver, self bool) (*grpcClient, error) {
	slog.Infof("Starting Client on %s", addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewConalgClient(conn)

	slog.Infof("Initializing Fast Propose Stream on %s", addr)
	fpStream, err := client.FastProposeStream(context.Background())
	if err != nil {
		slog.Fatal(err)
	}

	stableStream, err := client.StableStream(context.Background())
	if err != nil {
		slog.Fatal(err)
	}

	retryStream, err := client.RetryStream(context.Background())
	if err != nil {
		slog.Fatal(err)
	}

	c := &grpcClient{
		client,
		fpStream,
		stableStream,
		retryStream,
		addr,
		rec,
		self,
	}

	go c.receiveFastProposeResponse()
	go c.receiveRetryResponse()
	return c, nil
}

func (c *grpcClient) sendFastPropose(req *model.Request) {
	if !c.self {
		time.Sleep(50 * time.Millisecond)
	}
	err := c.fastProposeStream.Send(req.ToProposePb(model.FASTP_PROP))
	if err != nil {
		slog.Error(err)
	}
}

func (c *grpcClient) sendStablePropose(req *model.Request) {
	if !c.self {
		time.Sleep(50 * time.Millisecond)
	}
	err := c.stableProposeStream.Send(req.ToProposePb(model.STABLE_PROP))
	if err != nil {
		slog.Error(err)
	}
}

func (c *grpcClient) sendRetryPropose(req *model.Request) {
	if !c.self {
		time.Sleep(50 * time.Millisecond)
	}
	err := c.retryProposeStream.Send(req.ToProposePb(model.RETRY_PROP))
	if err != nil {
		slog.Error(err)
	}
}

func (c *grpcClient) receiveRetryResponse() {
	for {
		msg, err := c.retryProposeStream.Recv()

		if err == io.EOF {
			slog.Warn("EOF")
			break
		}
		if err != nil {
			slog.Error(err)
		}

		c.receiver.ReceiveResponse(model.FromResponsePb(msg))
	}
}

func (c *grpcClient) receiveFastProposeResponse() {
	for {
		msg, err := c.fastProposeStream.Recv()

		if err == io.EOF {
			slog.Warn("EOF")
			break
		}
		if err != nil {
			slog.Error(err)
		}

		c.receiver.ReceiveResponse(model.FromResponsePb(msg))
	}
}
