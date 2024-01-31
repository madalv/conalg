package caesar

import (
	"conalg/config"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/gookit/slog"
)

type Conalg interface {
	Propose(payload []byte)
}

type Application interface {
	DetermineConflict(c1, c2 []byte) bool
	Execute(c []byte)
}

type Transport interface {
	BroadcastFastPropose(req *Request)
	RunServer(port string) error
	ConnectToNodes(nodes []string) error
}

type Caesar struct {
	// map of commands and their status
	// TODO private kv store
	history map[string]Request
	// array mapping command c to its ballot nr
	ballots   map[string]uint
	clock     *Clock
	cfg       config.Config
	transport Transport
}

func NewCaesar(cfg config.Config, transport Transport) *Caesar {
	slog.Info("Initializing Caesar Module")
	return &Caesar{
		history:   make(map[string]Request),
		ballots:   make(map[string]uint),
		clock:     NewClock(uint64(len(cfg.Nodes))),
		cfg:       cfg,
		transport: transport,
	}
}

func (c *Caesar) Propose(payload []byte) {
	req := Request{
		ID:           uuid.NewString(),
		Payload:      payload,
		Timestamp:    c.clock.NewTimestamp(),
		Pred:         gs.NewSet[string](),
		Status:       WAITING,
		Ballot:       0,
		Forced:       false,
		ResponseChan: make(chan Response),
		ProposeTime:  time.Now(),
	}

	slog.Infof("Proposing %s", req)

	c.history[req.ID] = req
}

// TODO timeout :3
func (c *Caesar) FastPropose(req *Request) {
	slog.Infof("Fast Proposing %s", req)

	c.ballots[req.ID] = 0
	c.transport.BroadcastFastPropose(req)
}
