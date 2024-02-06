package caesar

import (
	"conalg/config"
	"conalg/models"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type Transport interface {
	BroadcastFastPropose(req *models.Request)
	RunServer() error
	ConnectToNodes() error
}

type Caesar struct {
	// map of commands and their status
	History cmap.ConcurrentMap[string, models.Request]
	// array mapping command c to its ballot nr
	Ballots   map[string]uint
	Clock     *Clock
	Cfg       config.Config
	Transport Transport
	Executer  Application
}

func NewCaesar(Cfg config.Config, transport Transport, app Application) *Caesar {
	slog.Info("Initializing Caesar Module")

	return &Caesar{
		History:   cmap.New[models.Request](),
		Ballots:   make(map[string]uint),
		Clock:     NewClock(uint64(len(Cfg.Nodes))),
		Cfg:       Cfg,
		Transport: transport,
		Executer:  app,
	}
}

func (c *Caesar) ReceiveFastProposeResponse(r models.Response) {
	req, ok := c.History.Get(r.RequestID)
	if !ok {
		slog.Warnf("Received response for unknown request %s", r.RequestID)
		// slog.Debug(c.History.Items())
		return
	}

	req.ResponseChan <- r
	// slog.Debug("Sent response...")
}

func (c *Caesar) Propose(payload []byte) {
	req := models.NewRequest(payload, c.Clock.NewTimestamp(), c.Cfg.FastQuorum, c.Cfg.ID)
	slog.Debugf("Proposing %s", req)
	c.History.Set(req.ID, req)
	slog.Debug(c.History.Items())
	go c.FastPropose(req)
}

// TODO what the fuck is up with the damn timeout??
func (c *Caesar) FastPropose(req models.Request) {
	slog.Debugf("Fast Proposing %s", req)

	replies := map[string]models.Response{}
	maxTimestamp := req.Timestamp
	pred := gs.NewSet[string]()

	c.Transport.BroadcastFastPropose(&req)

	for {
		select {
		case reply := <-req.ResponseChan:
			slog.Debugf("Received %s for req %s", reply, req.ID)

			// slog.Debug(c.History.Items())

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
				c.History.Set(req.ID, req)
				slog.Debugf("REPLIES have NACK. must RETRY.  req %s\n Replies: \n %s", req, replies)

				// TODO retry
				return
			} else if len(replies) == c.Cfg.FastQuorum {
				req.Timestamp = maxTimestamp
				req.Pred = pred
				c.History.Set(req.ID, req)
				slog.Debugf("FQ REACHED ----- %s. \n Replies: \n %s", req, replies)
				// TODO stable
				return
			} else if len(replies) == c.Cfg.ClassicQuorum {
				req.Timestamp = maxTimestamp
				req.Pred = pred
				c.History.Set(req.ID, req)
				// TODO send slow propose
				slog.Debugf("CQ REACHED ----- %s. \n Replies: \n %s", req, replies)
				return
			}

		case <-time.After(30 * time.Second):
			slog.Warnf("Fast Propose timed out for req %s, DO SMTH??", req.ID)
		}
	}

}

func repliesHaveNack(replies map[string]models.Response) bool {
	for _, reply := range replies {
		if !reply.Status {
			return true
		}
	}
	return false
}
