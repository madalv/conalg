package transport

import (
	"conalg/pb"
	"google.golang.org/grpc"
	"io"
	"log/slog"
)

type grpcServer struct {
	pb.ConalgServer
}

func NewGRPCServer() *grpc.Server {
	slog.Info("Initializing gRPC Server")
	s := grpc.NewServer()

	pb.RegisterConalgServer(s, &grpcServer{})
	return s
}

func (srv *grpcServer) FastProposeStream(stream pb.Conalg_FastProposeStreamServer) error {
	slog.Info("Starting Fast Propose Stream Server")

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		slog.Debug(msg.C)
		err = stream.Send(&pb.FastProposeResponse{
			C:      msg.C,
			Ballot: msg.Ballot,
			Time:   msg.Time,
			Pred:   nil,
			Result: false,
		})
		
		if err != nil {
			return err
		}
	}
}
