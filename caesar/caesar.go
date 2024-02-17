package caesar

import (
	"conalg/config"
	"conalg/models"

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
	Ballots   cmap.ConcurrentMap[string, uint]
	Clock     *Clock
	Cfg       config.Config
	Transport Transport
	Executer  Application
	// used in the wait function, to notify about the status of a request
	StatusUpdateChan chan models.Request
}

func NewCaesar(Cfg config.Config, transport Transport, app Application) *Caesar {
	slog.Info("Initializing Caesar Module")

	return &Caesar{
		History:   cmap.New[models.Request](),
		Ballots:   cmap.New[uint](),
		Clock:     NewClock(uint64(len(Cfg.Nodes))),
		Cfg:       Cfg,
		Transport: transport,
		Executer:  app,
		StatusUpdateChan: make(chan models.Request),
	}
}

func (c *Caesar) Propose(payload []byte) {
	req := models.NewRequest(payload, c.Clock.NewTimestamp(), c.Cfg.FastQuorum, c.Cfg.ID)
	slog.Debugf("Proposing %s", req)
	c.History.Set(req.ID, req)
	slog.Debug(c.History.Items())
	go c.FastPropose(req)
}

// computePred computes the predecessor set for a request
// TODO reformat this, jesus christ
func (c *Caesar) computePred(reqID string, payload []byte, timestamp uint64, whitelist gs.Set[string]) (pred gs.Set[string]) {
	slog.Debug("Computing PRED: ", reqID, payload, timestamp, whitelist)
	pred = gs.NewSet[string]()
	iterator := c.History.IterBuffered()
	if whitelist == nil {
		whitelist = gs.NewSet[string]()
	}

	for kv := range iterator {
		_, req := kv.Key, kv.Val
		slog.Debug(req.ID, string(req.Payload), req.Timestamp, req.Status)
		if reqID != req.ID && c.Executer.DetermineConflict(payload, req.Payload) {
			if whitelist.IsEmpty() && req.Timestamp < timestamp {
				pred.Add(req.ID)
			} else if !whitelist.IsEmpty() {
				if whitelist.Contains(req.ID) ||
					(req.Timestamp < timestamp && (req.Status == models.SLOW_PEND ||
						req.Status == models.ACC ||
						req.Status == models.STABLE)) {
					pred.Add(req.ID)
				}
			}
		}
	}
	return pred
}

// repliesHaveNack checks if any of the replies have a nack status
func repliesHaveNack(replies map[string]models.Response) bool {
	for _, reply := range replies {
		if !reply.Status {
			return true
		}
	}
	return false
}

func wait
