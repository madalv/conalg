package transport

import (
	"conalg/config"
	"conalg/pb"
	"io"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
)

type grpcServer struct {
	pb.ConalgServer
	cfg config.Config
}

func NewGRPCServer(cfg config.Config) *grpc.Server {
	slog.Infof("Initializing gRPC Server")
	s := grpc.NewServer()

	pb.RegisterConalgServer(s, &grpcServer{cfg: cfg})
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

		err = stream.Send(&pb.FastProposeResponse{
			ResquestId: msg.RequestId,
			Ballot:     msg.Ballot,
			Time:       msg.Time,
			Pred:       nil,
			Result:     true,
			From:       srv.cfg.ID,
		})

		if err != nil {
			return err
		}
	}
}
