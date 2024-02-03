package transport

import (
	"conalg/caesar"
	"conalg/config"
	"net"

	"github.com/gookit/slog"
	"google.golang.org/grpc"
)

type grpcTransport struct {
	clients         []*grpcClient
	server          *grpc.Server
	fastProposeChan chan *caesar.Request
	cfg             config.Config
}

func NewGRPCTransport(cfg config.Config) (*grpcTransport, error) {
	slog.Info("Initializing gRPC Transport Module")

	module := &grpcTransport{
		clients:         []*grpcClient{},
		server:          NewGRPCServer(),
		fastProposeChan: make(chan *caesar.Request, 100),
		cfg:             cfg,
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

func (t *grpcTransport) RunServer() error {
	listener, err := net.Listen("tcp", t.cfg.Port)
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

func (t *grpcTransport) ConnectToNodes() error {
	slog.Info("Connecting to nodes")
	clients := make([]*grpcClient, 0, len(t.cfg.Nodes)+1)
	for _, node := range t.cfg.Nodes {
		c, err := newGRPCClient(node)

		if err != nil {
			return err
		}

		clients = append(clients, c)
	}

	c, err := newGRPCClient(t.cfg.Port)
	if err != nil {
		return err
	}

	clients = append(clients, c)
	slog.Debug(clients)

	t.clients = clients
	return nil
}
