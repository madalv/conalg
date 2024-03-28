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

func InitConalgModule(executer Application, envpath string, lvl slog.Level, analyzerOn bool) Conalg {

	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
	})

	slog.SetLogLevel(lvl)

	cfg := config.NewConfig(envpath)
	slog.Debug(cfg)

	transport, err := transport.NewGRPCTransport(cfg)
	if err != nil {
		slog.Fatal(err)
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
