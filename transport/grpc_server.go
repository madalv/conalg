package transport

import (
	"conalg/config"
	"conalg/model"
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
	slog.Infof("Initializing gRPC Server .  .  .")
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

		go func(stream pb.Conalg_FastProposeStreamServer, msg *pb.Propose) {
			outcome := srv.receiver.ReceiveFastPropose(model.FromProposePb(msg))
			err = stream.Send(model.ToResponsePb(outcome))
			if err != nil {
				slog.Error(err)
			}
		}(stream, msg)
	}
}

func (srv *grpcServer) StableStream(stream pb.Conalg_StableStreamServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		go func(stream pb.Conalg_StableStreamServer, msg *pb.Propose) {
			err := srv.receiver.ReceiveStablePropose(model.FromProposePb(msg))
			if err != nil {
				slog.Error(err)
			}
		}(stream, msg)
	}
}
