package caesar

import (
	"context"
	"errors"
	"log/slog"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/model"
	"github.com/madalv/conalg/util"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type Transport interface {
	BroadcastFastPropose(req *model.Request)
	BroadcastStablePropose(req *model.Request)
	BroadcastRetryPropose(req *model.Request)
	RunServer() error
	ConnectToNodes() error
}

type Caesar struct {
	// map of commands and their status
	History cmap.ConcurrentMap[string, model.Request]
	// array mapping command c to its ballot nr
	Ballots   cmap.ConcurrentMap[string, uint]
	Clock     *Clock
	Cfg       config.Config
	Transport Transport
	Executer  Application
	// used in the wait function, to notify about the status of a request
	Publisher       util.Publisher[model.StatusUpdate]
	Decided         gs.Set[string]
	Analyzer        *util.Analyzer
	AnalyzerEnabled bool
	ResponseChannel chan model.Response
}

func NewCaesar(Cfg config.Config, transport Transport, app Application, analyzerOn bool) *Caesar {
	slog.Info("Initializing Caesar Module!")

	c := &Caesar{
		History:   cmap.New[model.Request](),
		Ballots:   cmap.New[uint](),
		Clock:     NewClock(),
		Cfg:       Cfg,
		Transport: transport,
		Executer:  app,
		Decided:   gs.NewSet[string](),
		Publisher: util.NewBroadcastServer[model.StatusUpdate](
			context.Background(),
			make(chan model.StatusUpdate)),
		Analyzer:        util.NewAnalyzer(analyzerOn),
		AnalyzerEnabled: analyzerOn,
	}

	// TODO delete this
	ticker := time.NewTicker(10 * time.Second)
	go func(c *Caesar) {
		for range ticker.C {
			c.History.IterCb(func(k string, req model.Request) {
				if req.Status != model.STABLE {
					slog.Info("Request: ", config.ID, req.ID, config.STATUS, req.Status, config.TIMESTAMP, req.Timestamp)
				}
			})
		}
	}(c)
	return c
}

func (c *Caesar) Propose(payload []byte) {
	req := model.NewRequest(payload, c.Clock.NewTimestamp(), c.Cfg.FastQuorum, c.Cfg.ID)
	slog.Debug("Proposing", config.ID, req.ID)
	c.History.Set(req.ID, req)
	go c.FastPropose(req.ID)
}

// computePred computes the predecessor set for a request
func (c *Caesar) computePred(reqID string, payload []byte, timestamp uint64, whitelist gs.Set[string]) (pred gs.Set[string]) {
	pred = gs.NewSet[string]()

	if whitelist == nil {
		whitelist = gs.NewSet[string]()
	}

	c.History.IterCb(func(k string, req model.Request) {
		if reqID != req.ID && c.Executer.DetermineConflict(payload, req.Payload) {
			if whitelist.IsEmpty() && req.Timestamp < timestamp {
				pred.Add(req.ID)
			} else if !whitelist.IsEmpty() {
				if whitelist.Contains(req.ID) ||
					(req.Timestamp < timestamp && (req.Status == model.SLOW_PEND ||
						req.Status == model.ACC ||
						req.Status == model.STABLE)) {
					pred.Add(req.ID)
				}
			}
		}
	})
	return pred
}

// repliesHaveNack checks if any of the replies have a nack status
func repliesHaveNack(replies map[string]model.Response) bool {
	for _, reply := range replies {
		if !reply.Result {
			return true
		}
	}
	return false
}

func (c *Caesar) ReceiveResponse(r model.Response) {
	req, ok := c.History.Get(r.RequestID)
	if !ok {
		slog.Debug("Received response for unknown request", config.ID, r.RequestID)
		return
	}
	req.ResponseChan <- r
}

func (c *Caesar) computeWaitlist(reqID string, payload []byte, timestamp uint64) (gs.Set[string], error) {
	waitgroup := gs.NewSet[string]()
	var err error

	c.History.IterCb(func(k string, req model.Request) {
		if reqID != req.ID &&
			c.Executer.DetermineConflict(payload, req.Payload) &&
			req.Timestamp > timestamp &&
			!req.Pred.Contains(reqID) {
			if req.Status == model.STABLE {
				err = errors.New("stable request doesn't contain the request in its predecessor set")
			} else {
				waitgroup.Add(req.ID)
			}
		}
	})

	return waitgroup, err
}

/*
wait computes the waitlist for a request, then it waits
for all the requests in the waitgroup to be stable.
If all the requests in the wg are stable, but at least one STILL
does not contain the request in its predecessor set, then the request is NACKed.

waitgroup = all conflicting requests with a greater timestamp
and that does NOT contain the request in its predecessor set
*/
func (c *Caesar) wait(id string, payload []byte, timestamp uint64) (res bool, aborted bool) {

	waitlist, err := c.computeWaitlist(id, payload, timestamp)
	if err != nil {
		slog.Debug("Auto NACK", config.ERR, err)
		return false, false
	}

	slog.Debug("Request waiting", config.ID, id, config.WAITLIST, waitlist.Cardinality())

	if waitlist.IsEmpty() {
		return true, false
	}

	ch := c.Publisher.Subscribe()
	defer c.Publisher.CancelSubscription(ch)

	req, _ := c.History.Get(id)
	if req.Status == model.STABLE || req.Status == model.ACC {
		return false, true
	}

	for update := range ch {

		if update.RequestID == id {
			slog.Debug("Got update while waiting, aborting...", config.UPDATE_STATUS, update.Status, config.UPDATE_ID, update.RequestID)
			return false, true
		}

		if waitlist.Contains(update.RequestID) &&
			(update.Status == model.ACC || update.Status == model.STABLE) {
			slog.Debug("----> Received update", config.ID, id, config.UPDATE_ID, update.RequestID, config.UPDATE_STATUS, update.Status, config.UPDATE_PRED, update.Pred.Cardinality())

			if !update.Pred.Contains(id) {
				c.Publisher.CancelSubscription(ch)
				slog.Debug("Request DONE WAITING", config.ID, id, config.OUTCOME, "false")
				return false, false
			} else {
				waitlist.Remove(update.RequestID)
			}
		}

		if waitlist.IsEmpty() {
			c.Publisher.CancelSubscription(ch)
			slog.Debug("Request DONE WAITING", config.ID, id, config.OUTCOME, "true")
			return true, false
		}
	}
	return false, false
}
