package transport

import (
	"conalg/pb"
	"io"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.ConalgServer
}

func NewGRPCServer() *grpc.Server {
	slog.Infof("Initializing gRPC Server")
	s := grpc.NewServer()

	pb.RegisterConalgServer(s, &grpcServer{})
	return s
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

		slog.Infof("Received Fast Propose: %v", msg)
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
