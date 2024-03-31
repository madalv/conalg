package caesar

import (
	"conalg/model"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

func (c *Caesar) RetryPropose(req model.Request) {
	req, _ = c.History.Get(req.ID)
	if req.Status == model.STABLE {
		slog.Fatalf("Request %s is in %s", req.ID, req.Status)
		return
	}
	slog.Debugf("Retrying for %s", req.ID)
	c.Transport.BroadcastRetryPropose(&req)

	replies := map[string]model.Response{}
	pred := gs.NewSet[string]()

	for reply := range req.ResponseChan {

		if reply.Type != model.RETRY_REPLY {
			continue
		}

		slog.Debugf("Received for %s reply %s %s %t %d %d", req.ID, reply.From, reply.Type, reply.Result, reply.Pred.Cardinality(), reply.Timestamp)

		if _, ok := replies[reply.From]; ok {
			slog.Warnf("Received duplicate reply %+v for %s %s ", reply, req.Payload, req.ID)
		} else {
			replies[reply.From] = reply
		}

		for p := range reply.Pred.Iter() {
			pred.Add(p)
		}

		if len(replies) == c.Cfg.ClassicQuorum {
			slog.Debugf("----- CQ REACHED (retry successful) ----- %s for %s", req.Payload, req.ID)
			c.StablePropose(req)
			return
		}
	}
}

func (c *Caesar) ReceiveRetryPropose(rp model.Request) (model.Response, bool) {

	slog.Debugf("Received retry propose for %s %s: %d", rp.ID, rp.Payload, rp.Pred.Cardinality())

	update := model.StatusUpdate{
		RequestID: rp.ID,
		Status:    "PRE_ACCEPTED",
	}


	// slog.Debugf("-----------------> Trying to publish status update for %s %s", update.RequestID, update.Status)

	c.Publisher.Publish(update)

	slog.Debugf("-----------------> Published status update for %s %s", update.RequestID, update.Status)

	c.Ballots.Set(rp.ID, rp.Ballot)
	req, ok := c.History.Get(rp.ID)
	if !ok {
		slog.Fatalf("Request %s not found in history", rp.ID)
	}

	if req.Status == model.STABLE {
		return model.Response{}, false
	}

	req.Timestamp = rp.Timestamp
	req.Pred = rp.Pred
	req.Status = model.ACC
	req.Ballot = rp.Ballot
	req.Forced = false
	c.History.Set(req.ID, req)

	update = model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.ACC,
		Pred:      gs.NewSet[string](req.Pred.ToSlice()...),
	}

	// slog.Debugf("-----------------> Trying to publish status update for %s %s %d", update.RequestID, update.Status, update.Pred.Cardinality())

	c.Publisher.Publish(update)

	slog.Debugf("-----------------> Published status update for %s %s %d", update.RequestID, update.Status, update.Pred.Cardinality())

	newPred := c.computePred(req.ID, req.Payload, req.Timestamp, nil)
	for p := range req.Pred.Iter() {
		newPred.Add(p)
	}
	return model.NewResponse(req.ID, model.RETRY_REPLY, true, newPred, c.Cfg.ID, req.Timestamp, req.Ballot), true
}
