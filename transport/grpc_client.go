package transport

import (
	"conalg/model"
	"conalg/pb"
	"context"
	"io"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	pb.ConalgClient
	fastProposeStream pb.Conalg_FastProposeStreamClient
	address           string
	receiver          Receiver
}

func newGRPCClient(addr string, rec Receiver) (*grpcClient, error) {
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

	c := &grpcClient{
		client,
		fpStream,
		addr,
		rec,
	}

	go c.receiveFastProposeResponse()
	return c, nil
}

func (c *grpcClient) sendFastPropose(req *model.Request) {
	err := c.fastProposeStream.Send(req.ToProposePb(model.FASTP_PROP))
	if err != nil {
		slog.Error(err)
	}
}

func (c *grpcClient) receiveFastProposeResponse() {
	for {
		msg, err := c.fastProposeStream.Recv()
		// slog.Info(c.address)
		if err == io.EOF {
			slog.Warn("EOF")
			break
		}
		if err != nil {
			slog.Error(err)
		}
		// slog.Debugf("Received Fast Propose Response: %v", msg)
		c.receiver.ReceiveResponse(model.FromResponsePb(msg))
	}
}
