package transport

import (
	"context"
	"io"

	"log/slog"

	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/model"
	"github.com/madalv/conalg/pb"
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
	slog.Info("Starting Client", config.ADDR, addr)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := pb.NewConalgClient(conn)

	slog.Info("Initializing Fast Propose Stream", config.ADDR, addr)
	fpStream, err := client.FastProposeStream(context.Background())
	if err != nil {
		slog.Error(err.Error())
	}

	stableStream, err := client.StableStream(context.Background())
	if err != nil {
		slog.Error(err.Error())
	}

	retryStream, err := client.RetryStream(context.Background())
	if err != nil {
		slog.Error(err.Error())
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
	err := c.fastProposeStream.Send(req.ToProposePb(model.FASTP_PROP))
	if err != nil {
		slog.Error(err.Error())
	}
}

func (c *grpcClient) sendStablePropose(req *model.Request) {
	m := req.ToProposePb(model.RETRY_PROP)
	err := c.stableProposeStream.Send(m)
	if err != nil {
		slog.Error(err.Error())
	}
}

func (c *grpcClient) sendRetryPropose(req *model.Request) {
	m := req.ToProposePb(model.RETRY_PROP)

	err := c.retryProposeStream.Send(m)
	if err != nil {
		slog.Error(err.Error())

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
			slog.Error(err.Error())
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
			slog.Error(err.Error())
		}

		c.receiver.ReceiveResponse(model.FromResponsePb(msg))
	}
}
