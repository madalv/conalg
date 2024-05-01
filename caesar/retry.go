package caesar

import (
	"log/slog"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/model"
)

func (c *Caesar) RetryPropose(req model.Request) {
	req, _ = c.History.Get(req.ID)
	if req.Status == model.STABLE {
		slog.Error("Request has another status than STABLE", config.ID, req.ID, config.STATUS, req.Status)
		return
	}
	slog.Debug("Retrying request", config.ID, req.ID)
	c.Transport.BroadcastRetryPropose(&req)

	replies := map[string]model.Response{}
	pred := gs.NewSet[string]()

	for reply := range req.ResponseChan {

		if reply.Type != model.RETRY_REPLY {
			continue
		}

		slog.Debug("Received reply", config.ID, req.ID, config.FROM, reply.From,
			config.TYPE, reply.Type, config.RESULT, reply.Result, config.PRED, reply.Pred.Cardinality(),
			config.TIMESTAMP, reply.Timestamp)

		if _, ok := replies[reply.From]; ok {
			slog.Warn("Received duplicate reply", config.REPLY, reply, config.ID, req.ID)
		} else {
			replies[reply.From] = reply
		}

		for p := range reply.Pred.Iter() {
			pred.Add(p)
		}

		if len(replies) == c.Cfg.ClassicQuorum {
			slog.Debug("----- CQ REACHED (retry successful) -----", config.ID, req.ID)
			c.StablePropose(req)
			return
		}
	}
}

func (c *Caesar) ReceiveRetryPropose(rp model.Request) (model.Response, bool) {
	slog.Debug("Received Retry Propose", config.ID, rp.ID, config.FROM, rp.Proposer, config.TIMESTAMP, rp.Timestamp)

	c.Ballots.Set(rp.ID, rp.Ballot)
	req, ok := c.History.Get(rp.ID)
	if !ok {
		slog.Error("Request not found in history", config.ID, rp.ID)
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

	update := model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.ACC,
		Pred:      gs.NewSet[string](req.Pred.ToSlice()...),
	}

	c.Publisher.Publish(update)

	slog.Debug("-----------------> Published status update", config.UPDATE_ID, update.RequestID,
		config.UPDATE_STATUS, update.Status, config.UPDATE_PRED, update.Pred.Cardinality())

	newPred := c.computePred(req.ID, req.Payload, req.Timestamp, nil)
	for p := range req.Pred.Iter() {
		newPred.Add(p)
	}
	return model.NewResponse(req.ID, model.RETRY_REPLY, true, newPred, c.Cfg.ID, req.Timestamp, req.Ballot), true
}
