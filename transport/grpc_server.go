package transport

import (
	"conalg/config"
	"conalg/models"
	"conalg/pb"
	"io"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.ConalgServer
	cfg      config.Config
	receiver Receiver
}

func NewGRPCServer(cfg config.Config, r Receiver) *grpc.Server {
	slog.Infof("Initializing gRPC Server")
	s := grpc.NewServer()

	pb.RegisterConalgServer(s, &grpcServer{cfg: cfg, receiver: r})
	return s
}

func (t *grpcServer) SetReceiver(r Receiver) {
	t.receiver = r
}

func (srv *grpcServer) FastProposeStream(stream pb.Conalg_FastProposeStreamServer) error {

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		go func(stream pb.Conalg_FastProposeStreamServer, msg *pb.FastPropose) {
			outcome := srv.receiver.ReceiveFastPropose(models.FromFastProposePb(msg))
			err = stream.Send(models.ToFastProposeResponse(outcome))
			if err != nil {
				slog.Error(err)
			}
		}(stream, msg)
	}
}
