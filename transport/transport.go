package transport

import (
	"conalg/caesar"
	"net"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
)

type grpcTransport struct {
	clients         []*grpcClient
	server          *grpc.Server
	fastProposeChan chan *caesar.Request
}

func NewGRPCTransport(nodes []string, port string) (*grpcTransport, error) {
	slog.Info("Initializing gRPC Transport Module")

	module := &grpcTransport{
		clients:         []*grpcClient{},
		server:          NewGRPCServer(),
		fastProposeChan: make(chan *caesar.Request, 100),
	}

	go module.ListenToChannels()

	return module, nil
}

func (t *grpcTransport) ListenToChannels() {
	for req := range t.fastProposeChan {
		slog.Info("Broadcasting Fast Propose")
		for _, client := range t.clients {
			client.sendFastProposeStream(req)
		}
	}
}

func (t *grpcTransport) RunServer(port string) error {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	if err := t.server.Serve(listener); err != nil {
		return err
	}
	return nil
}

func (t *grpcTransport) BroadcastFastPropose(req *caesar.Request) {
	t.fastProposeChan <- req
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
