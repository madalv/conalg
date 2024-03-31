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
			outcome, ok := srv.receiver.ReceiveFastPropose(model.FromProposePb(msg))
			if !ok {
				slog.Warnf("Fast Propose interrupted for %s", model.FromProposePb(msg).ID)
				return
			}
			err = stream.Send(model.ToResponsePb(outcome))
			if err != nil {
				slog.Error(err)
			}
			slog.Warnf("~~~ Sent Fast Propose Response for %s to %s", outcome.RequestID, outcome.From)
		}(stream, msg)
	}
}

func (srv *grpcServer) RetryStream(stream pb.Conalg_RetryStreamServer) error {
	for {
		msg, err := stream.Recv()
		slog.Warnf("~~~ Received Retry Propose for %s from %s", msg.RequestId, msg.From)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		go func(stream pb.Conalg_RetryStreamServer, msg *pb.Propose) {
			outcome, ok := srv.receiver.ReceiveRetryPropose(model.FromProposePb(msg))

			if !ok {
				slog.Warnf("Retry interrupted for %s", model.FromProposePb(msg).ID)
				return
			}

			err = stream.Send(model.ToResponsePb(outcome))
			if err != nil {
				slog.Error(err)
			}
			slog.Warnf("~~~ Sent Retry Response for %s to %s", outcome.RequestID, outcome.From)
		}(stream, msg)
	}
}

func (srv *grpcServer) StableStream(stream pb.Conalg_StableStreamServer) error {
	for {
		msg, err := stream.Recv()
		slog.Warnf("~~~ Received Stable Propose for %s from %s", msg.RequestId, msg.From)
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
