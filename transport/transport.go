package transport

import (
	"conalg/config"
	"conalg/model"
	"net"

	"github.com/gookit/slog"
)

/*
Receiver is an interface that defines the methods through
which the transport module serves the reponses back to
the consensus module (aka Caesar module)
*/
type Receiver interface {
	ReceiveResponse(model.Response)
	ReceiveFastPropose(model.Request) model.Response
	ReceiveRetryPropose(model.Request) model.Response
	ReceiveStablePropose(model.Request) error
}

type grpcTransport struct {
	clients           []*grpcClient
	fastProposeChan   chan *model.Request
	stableProposeChan chan *model.Request
	retryProposeChan  chan *model.Request
	cfg               config.Config
	receiver          Receiver
}

func NewGRPCTransport(cfg config.Config) (*grpcTransport, error) {
	slog.Info("Initializing gRPC Transport Module")

	module := &grpcTransport{
		clients:           []*grpcClient{},
		fastProposeChan:   make(chan *model.Request, 50),
		stableProposeChan: make(chan *model.Request, 50),
		retryProposeChan:  make(chan *model.Request, 50),
		cfg:               cfg,
	}

	go module.ListenFastProposeChan()
	go module.ListenStableProposeChan()
	go module.ListenRetryProposeChan()

	return module, nil
}

func (t *grpcTransport) SetReceiver(r Receiver) {
	t.receiver = r
}

func (t *grpcTransport) ListenFastProposeChan() {
	for req := range t.fastProposeChan {
		for _, client := range t.clients {
			client.sendFastPropose(req)
		}
	}
}

func (t *grpcTransport) ListenRetryProposeChan() {
	for req := range t.retryProposeChan {
		for _, client := range t.clients {
			client.sendRetryPropose(req)
		}
	}
}

func (t *grpcTransport) ListenStableProposeChan() {
	for req := range t.stableProposeChan {
		for _, client := range t.clients {
			client.sendStablePropose(req)
		}
	}
}

func (t *grpcTransport) RunServer() error {
	server := NewGRPCServer(t.cfg, t.receiver)
	listener, err := net.Listen("tcp", t.cfg.Port)
	if err != nil {
		return err
	}

	if err := server.Serve(listener); err != nil {
		return err
	}
	return nil
}

func (t *grpcTransport) BroadcastFastPropose(req *model.Request) {
	t.fastProposeChan <- req
}

func (t *grpcTransport) BroadcastStablePropose(req *model.Request) {
	t.stableProposeChan <- req
}

func (t *grpcTransport) BroadcastRetryPropose(req *model.Request) {
	t.retryProposeChan <- req
}

func (t *grpcTransport) ConnectToNodes() error {
	slog.Info("Connecting to nodes")
	clients := make([]*grpcClient, 0, len(t.cfg.Nodes)+1)
	for _, node := range t.cfg.Nodes {
		c, err := newGRPCClient(node, t.receiver)

		if err != nil {
			return err
		}

		clients = append(clients, c)
	}

	c, err := newGRPCClient(t.cfg.Port, t.receiver)
	if err != nil {
		return err
	}

	clients = append(clients, c)

	t.clients = clients
	return nil
}
