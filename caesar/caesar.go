package caesar

import (
	"conalg/config"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
	cmap "github.com/orcaman/concurrent-map/v2"
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
	RunServer() error
	ConnectToNodes() error
}

type Caesar struct {
	// map of commands and their status
	history cmap.ConcurrentMap[string, Request]
	// array mapping command c to its ballot nr
	ballots   map[string]uint
	clock     *Clock
	cfg       config.Config
	transport Transport
}

func NewCaesar(cfg config.Config, transport Transport) *Caesar {
	slog.Info("Initializing Caesar Module")

	return &Caesar{
		history:   cmap.New[Request](),
		ballots:   make(map[string]uint),
		clock:     NewClock(uint64(len(cfg.Nodes))),
		cfg:       cfg,
		transport: transport,
	}
}

func (c *Caesar) ReceiveFastProposeResponse(r Response) {
	req, ok := c.history.Get(r.RequestID)
	if !ok {
		slog.Warnf("Received response for unknown request %s", r.RequestID)
		return
	}

	req.ResponseChan <- r
}

func (c *Caesar) Propose(payload []byte) {
	req := NewRequest(payload, c.clock.NewTimestamp(), c.cfg.FastQuorum, c.cfg.ID)
	slog.Debugf("Proposing %s", req)
	c.history.Set(req.ID, req)
	go c.FastPropose(req)
}

// TODO what the fuck is up with the damn timeout??
func (c *Caesar) FastPropose(req Request) {
	slog.Debugf("Fast Proposing %s", req)

	replies := map[string]Response{}
	maxTimestamp := req.Timestamp
	pred := gs.NewSet[string]()

	c.transport.BroadcastFastPropose(&req)

	select {
	case reply := <-req.ResponseChan:
		slog.Debugf("Received %s for req %s", reply, req.ID)

		if _, ok := replies[reply.From]; ok {
			slog.Warnf("Received duplicate reply %s for req %s", reply, req.ID)
		} else {
			replies[reply.From] = reply
		}

		if reply.Timestamp > maxTimestamp {
			maxTimestamp = reply.Timestamp
		}

		pred.Union(reply.Pred)

		slog.Debug(pred)
		slog.Debug(maxTimestamp)

		if repliesHaveNack(replies) {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.history.Set(req.ID, req)
			slog.Debugf("REPLIES have NACK. must RETRY.  req %s\n Replies: \n %s", req, replies)

			// TODO retry
			return
		} else if len(replies) == c.cfg.FastQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.history.Set(req.ID, req)
			slog.Debugf("FQ REACHED ----- %s. \n Replies: \n %s", req, replies)
			// TODO stable
			return
		} else if len(replies) == c.cfg.ClassicQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.history.Set(req.ID, req)
			// TODO send slow propose
			slog.Debugf("CQ REACHED ----- %s. \n Replies: \n %s", req, replies)
			return
		}

	case <-time.After(30 * time.Second):
		slog.Warnf("Fast Propose timed out for req %s, DO SMTH??", req.ID)
	}
}

func repliesHaveNack(replies map[string]Response) bool {
	for _, reply := range replies {
		if !reply.Status {
			return true
		}
	}
	return false
}
