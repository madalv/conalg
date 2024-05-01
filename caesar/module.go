package caesar

import (
	"time"

	"log/slog"

	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/transport"
)

type Conalg interface {
	Propose(payload []byte)
}

type Application interface {
	DetermineConflict(c1, c2 []byte) bool
	Execute(c []byte)
}

func InitConalgModule(executer Application, envpath string, lvl slog.Level, analyzerOn bool) Conalg {
	config.ConfigLogger(lvl)

	cfg := config.NewConfig(envpath)
	slog.Debug("CRead config", config.CFG, cfg)

	transport, err := transport.NewGRPCTransport(cfg)
	if err != nil {
		slog.Error("Failed to create transport", config.ERR, err)
	}

	caesarModule := NewCaesar(cfg, transport, executer, analyzerOn)

	transport.SetReceiver(caesarModule)

	go func() {
		time.Sleep(1 * time.Second)
		transport.ConnectToNodes()
	}()

	go transport.RunServer()

	return caesarModule
}
