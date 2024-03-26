package caesar

import (
	"conalg/config"
	"conalg/model"
	"conalg/util"
	"context"
	"errors"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
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
	Publisher util.Publisher[model.StatusUpdate]
	Decided   gs.Set[string]
	Analyzer  *util.Analyzer
}

func NewCaesar(Cfg config.Config, transport Transport, app Application) *Caesar {
	slog.Info("Initializing Caesar Module")

	return &Caesar{
		History:   cmap.New[model.Request](),
		Ballots:   cmap.New[uint](),
		Clock:     NewClock(uint64(len(Cfg.Nodes))),
		Cfg:       Cfg,
		Transport: transport,
		Executer:  app,
		Decided:   gs.NewSet[string](),
		Publisher: util.NewBroadcastServer[model.StatusUpdate](
			context.Background(),
			make(chan model.StatusUpdate)),
		Analyzer: util.NewAnalyzer(),
	}
}

func (c *Caesar) Propose(payload []byte) {
	req := model.NewRequest(payload, c.Clock.NewTimestamp(), c.Cfg.FastQuorum, c.Cfg.ID)
	slog.Debugf("Proposing %v", req)
	c.History.Set(req.ID, req)
	go c.FastPropose(req.ID)
}

// computePred computes the predecessor set for a request
func (c *Caesar) computePred(reqID string, payload []byte, timestamp uint64, whitelist gs.Set[string]) (pred gs.Set[string]) {
	// slog.Debug("Computing PRED: ", reqID, payload, timestamp, whitelist)
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
		slog.Warnf("Received response for unknown request %s", r.RequestID)
		return
	}
	req.ResponseChan <- r
}

func (c *Caesar) computeWaitlist(reqID string, payload []byte, timestamp uint64) (gs.Set[string], error) {
	slog.Debugf("Computing waitlist for request %s %s", reqID, payload)
	waitgroup := gs.NewSet[string]()
	var err error

	c.History.IterCb(func(k string, req model.Request) {
		// slog.Debug(req.ID, string(req.Payload), req.Timestamp, req.Status)

		if reqID != req.ID &&
			c.Executer.DetermineConflict(payload, req.Payload) &&
			req.Timestamp > timestamp &&
			!req.Pred.Contains(reqID) {
			if req.Status == model.STABLE {
				err = errors.New("auto NACK: stable request doesn't contain the request in its predecessor set")
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
func (c *Caesar) wait(id string, payload []byte, timestamp uint64) bool {
	// slog.Debugf("Request %s is waiting", id)
	waitlist, err := c.computeWaitlist(id, payload, timestamp)
	if err != nil {
		slog.Warnf("Auto NACK: %s", err)
		return false
	}

	slog.Debugf("Request %s %s is waiting with waitlist %v", id, payload, waitlist)

	if waitlist.IsEmpty() {
		return true
	}

	ch := c.Publisher.Subscribe()
	defer c.Publisher.CancelSubscription(ch)

	for update := range ch {
		if !waitlist.Contains(update.RequestID) {
			continue
		}

		if update.Status == model.DECIDED || update.Status == model.STABLE {

			if !update.Pred.Contains(id) {
				slog.Warnf("Request %s %s is DONE WAITING - false outcome", id, payload)
				return false
			} else {
				waitlist.Remove(update.RequestID)
			}
		}

		if waitlist.IsEmpty() {
			slog.Warnf("Request %s %s is DONE WAITING - true outcome", id, payload)
			return true
		}
	}
	return false
}
