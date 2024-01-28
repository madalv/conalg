package transport

import (
	"conalg/pb"
	"net"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
)

type Transport interface {
	BroadcastFastPropose(c string, ballot uint, time uint) error
	RunServer() error
	ConnectToNodes(nodes []string) error
}

type grpcTransport struct {
	clients []*grpcClient
	server  *grpc.Server
}

func NewGRPCTransport(nodes []string, port string) (*grpcTransport, error) {
	slog.Info("Initializing gRPC Transport Module")

	return &grpcTransport{
		clients: []*grpcClient{},
		server:  NewGRPCServer(),
	}, nil
}

func (t *grpcTransport) RunServer(port string) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		slog.Fatal(err)
	}

	if err := t.server.Serve(listener); err != nil {
		slog.Fatal(err)
	}
}

func (t *grpcTransport) BroadcastFastPropose(c string, ballot uint64, time uint64) error {
	slog.Info("Broadcasting Fast Propose")
	for _, client := range t.clients {
		client.fastProposeChan <- &pb.FastPropose{
			C:      c,
			Ballot: ballot,
			Time:   time,
		}
	}
	return nil
}

func (t *grpcTransport) ConnectToNodes(nodes []string) error {
	slog.Info("Connecting to nodes")
	clients := make([]*grpcClient, 0, len(nodes))
	for _, node := range nodes {
		c, err := newGRPCClient(node)

		if err != nil {
			return err
		}

		clients = append(clients, c)
	}

	t.clients = clients
	return nil
}
