package caesar

import (
	"conalg/config"
	"conalg/transport"
	"time"

	"github.com/gookit/slog"
)

type Conalg interface {
	Propose(payload []byte)
}

type Application interface {
	DetermineConflict(c1, c2 []byte) bool
	Execute(c []byte)
}

// TODO add env path as param
func InitConalgModule(executer Application) Conalg {
	cfg := config.NewConfig()
	slog.Debug(cfg)

	transport, err := transport.NewGRPCTransport(cfg)
	if err != nil {
		slog.Fatal(err)
	}

	caesarModule := NewCaesar(cfg, transport, executer)

	transport.SetReceiver(caesarModule)

	go func() {
		time.Sleep(1 * time.Second)
		transport.ConnectToNodes()
	}()

	go transport.RunServer()

	return caesarModule
}
