package transport

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/madalv/conalg/config"

	"log/slog"

	"github.com/madalv/conalg/model"
)

/*
Receiver is an interface that defines the methods through
which the transport module serves the reponses back to
the consensus module (aka Caesar module)
*/
type Receiver interface {
	ReceiveResponse(model.Response)
	ReceiveFastPropose(model.Request) (model.Response, bool)
	ReceiveRetryPropose(model.Request) (model.Response, bool)
	ReceiveSlowPropose(model.Request) (model.Response, bool)
	ReceiveStablePropose(model.Request) error
}

type grpcTransport struct {
	clients  []*grpcClient
	cfg      config.Config
	receiver Receiver
	mu       sync.RWMutex
}

func NewGRPCTransport(cfg config.Config) (*grpcTransport, error) {
	slog.Info("Initializing gRPC Transport Module")

	module := &grpcTransport{
		clients: []*grpcClient{},
		cfg:     cfg,
		mu:      sync.RWMutex{},
	}

	return module, nil
}

func (t *grpcTransport) SetReceiver(r Receiver) {
	t.receiver = r
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
	t.mu.RLock()
	for _, client := range t.clients {
		go client.sendFastPropose(req)
	}
	t.mu.RUnlock()
}

func (t *grpcTransport) BroadcastSlowPropose(req *model.Request) {
	t.mu.RLock()
	for _, client := range t.clients {
		client.sendSlowPropose(req)
	}
	t.mu.RUnlock()
}

func (t *grpcTransport) BroadcastStablePropose(req *model.Request) {
	t.mu.RLock()
	for _, client := range t.clients {
		go client.sendStablePropose(req)
	}
	t.mu.RUnlock()
}

func (t *grpcTransport) BroadcastRetryPropose(req *model.Request) {
	t.mu.RLock()
	for _, client := range t.clients {
		go client.sendRetryPropose(req)
	}
	t.mu.RUnlock()
}

func (t *grpcTransport) ConnectToNodes() error {
	slog.Info("Connecting to nodes. . .")

	envVar := os.Getenv("TIMES")

	myMap := parseEnvVarToMap(envVar)
	fmt.Println("Parsed map:", myMap)

	clients := make([]*grpcClient, 0, len(t.cfg.Nodes)+1)
	for _, node := range t.cfg.Nodes {
		time, err := strconv.Atoi(myMap[node])
		if err != nil {
			slog.Error(err.Error(), "val", myMap[node], "node", node)
		}
		c, err := newGRPCClient(node, t.receiver, time)

		if err != nil {
			return err
		}
		clients = append(clients, c)
	}

	c, err := newGRPCClient(t.cfg.Port, t.receiver, 0)
	if err != nil {
		return err
	}

	clients = append(clients, c)

	t.mu.Lock()
	t.clients = clients
	t.mu.Unlock()
	return nil
}

func parseEnvVarToMap(envVar string) map[string]string {
	result := make(map[string]string)
	pairs := strings.Split(envVar, ",")

	for _, pair := range pairs {
		kv := strings.Split(pair, ">")
		if len(kv) == 2 {
			key := kv[0]
			value := kv[1]
			result[key] = value
		} else {
			fmt.Printf("Invalid pair: %s\n", pair)
		}
	}
	return result
}
