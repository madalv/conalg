package caesar

import (
	"conalg/model"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

func (c *Caesar) RetryPropose(req model.Request) {
	slog.Infof("Retrying %s %s", req.Payload, req.ID)
	c.Transport.BroadcastRetryPropose(&req)

	replies := map[string]model.Response{}
	pred := gs.NewThreadUnsafeSet[string]()

	for reply := range req.ResponseChan {
		slog.Debugf("Received %+v for req %s /%s ", reply, req.Payload, req.ID)

		if reply.Type != model.RETRY_REPLY {
			slog.Warnf("Received unexpected reply %+v for req %s /%s ", reply, req.Payload, req.ID)
			continue
		}

		if _, ok := replies[reply.From]; ok {
			slog.Warnf("Received duplicate reply %+v for req %s /%s ", reply, req.Payload, req.ID)
		} else {
			replies[reply.From] = reply
		}

		for p := range reply.Pred.Iter() {
			pred.Add(p)
		}

		if len(replies) == c.Cfg.ClassicQuorum {
			slog.Infof("CQ REACHED (retry successful) ----- %s /%s", req.Payload, req.ID)
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
		Pred:      req.Pred,
	}
	c.Publisher.Publish(update)

	slog.Warnf("Published status update for %v", update)

	newPred := c.computePred(req.ID, req.Payload, req.Timestamp, nil)
	for p := range req.Pred.Iter() {
		newPred.Add(p)
	}
	return model.NewResponse(req.ID, model.RETRY_REPLY, true, newPred, c.Cfg.ID, req.Timestamp, req.Ballot)
}
