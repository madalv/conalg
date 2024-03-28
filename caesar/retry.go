package caesar

import (
	"conalg/model"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

func (c *Caesar) RetryPropose(req model.Request) {
	slog.Debugf("Retrying %s %s", req.Payload, req.ID)
	c.Transport.BroadcastRetryPropose(&req)

	replies := map[string]model.Response{}
	pred := gs.NewThreadUnsafeSet[string]()

	for reply := range req.ResponseChan {

		if reply.Type != model.RETRY_REPLY {
			continue
		}

		slog.Debugf("Received %+v for %s  ", reply, req.ID)

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

func (c *Caesar) ReceiveRetryPropose(rp model.Request) model.Response {
	slog.Debugf("Received Retry Propose %v", rp)
	c.Ballots.Set(rp.ID, rp.Ballot)
	req, ok := c.History.Get(rp.ID)
	if !ok {
		slog.Fatalf("Request %s not found in history", rp.ID)
	}

	req.Timestamp = rp.Timestamp
	req.Pred = rp.Pred
	req.Status = model.ACC
	req.Ballot = rp.Ballot
	req.Forced = false
	c.History.Set(req.ID, req)

	update := model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.ACC,
		Pred:      gs.NewThreadUnsafeSet[string](req.Pred.ToSlice()...),
	}
	c.Publisher.Publish(update)

	slog.Debugf("-----------------> Published status update for %s %s %s", update.RequestID, update.Status, update.Pred.String())

	newPred := c.computePred(req.ID, req.Payload, req.Timestamp, nil)
	for p := range req.Pred.Iter() {
		newPred.Add(p)
	}
	return model.NewResponse(req.ID, model.RETRY_REPLY, true, newPred, c.Cfg.ID, req.Timestamp, req.Ballot)
}
