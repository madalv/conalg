package transport

import (
	"io"

	"log/slog"

	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/model"
	"github.com/madalv/conalg/pb"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.ConalgServer
	cfg      config.Config
	receiver Receiver
}

func NewGRPCServer(cfg config.Config, r Receiver) *grpc.Server {
	slog.Info("Initializing gRPC Server .  .  .")
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
			outcome, ok := srv.receiver.ReceiveFastPropose(model.FromProposePb(msg))
			if !ok {
				return
			}
			err = stream.Send(model.ToResponsePb(outcome))
			if err != nil {
				slog.Error(err.Error())
			}
		}(stream, msg)
	}
}

func (srv *grpcServer) SlowProposeStream(stream pb.Conalg_SlowProposeStreamServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		go func(stream pb.Conalg_SlowProposeStreamServer, msg *pb.Propose) {
			outcome, ok := srv.receiver.ReceiveSlowPropose(model.FromProposePb(msg))
			if !ok {
				return
			}
			err = stream.Send(model.ToResponsePb(outcome))
			if err != nil {
				slog.Error(err.Error())
			}
		}(stream, msg)
	}
}

func (srv *grpcServer) RetryStream(stream pb.Conalg_RetryStreamServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		go func(stream pb.Conalg_RetryStreamServer, msg *pb.Propose) {
			outcome, ok := srv.receiver.ReceiveRetryPropose(model.FromProposePb(msg))

			if !ok {
				slog.Warn("Retry interrupted for ", config.ID, model.FromProposePb(msg).ID)
				return
			}

			err = stream.Send(model.ToResponsePb(outcome))
			if err != nil {
				slog.Error(err.Error())
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
				slog.Error(err.Error())
			}
		}(stream, msg)
	}
}
