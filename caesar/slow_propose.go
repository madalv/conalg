package caesar

import (
	"log/slog"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/model"
)

func (c *Caesar) SlowPropose(reqID string) {
	req, ok := c.History.Get(reqID)
	if !ok {
		slog.Warn("Request not found in history", config.ID, reqID)
	}
	replies := map[string]model.Response{}
	maxTimestamp := req.Timestamp
	pred := gs.NewSet[string]()

	c.Transport.BroadcastSlowPropose(&req)

	for reply := range req.ResponseChan {
		req, _ := c.History.Get(reqID)
		slog.Debug("Received reply", config.ID, req.ID, config.FROM, reply.From,
			config.TYPE, reply.Type, config.RESULT, reply.Result, config.PRED, reply.Pred.Cardinality(),
			config.TIMESTAMP, reply.Timestamp)

		if reply.Type != model.FASTP_REPLY {
			slog.Warn("Received unexpected reply", config.REPLY, reply, config.ID, req.ID)
		}

		if _, ok := replies[reply.From]; ok {
			slog.Warn("Received duplicate reply", config.REPLY, reply, config.ID, req.ID)
		} else {
			replies[reply.From] = reply
		}

		if reply.Timestamp > maxTimestamp {
			maxTimestamp = reply.Timestamp
		}

		for p := range reply.Pred.Iter() {
			pred.Add(p)
		}

		if repliesHaveNack(replies) {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			slog.Debug("----- REPLIES have NACK RETRYING ", config.ID, req.ID)
			go c.RetryPropose(req.ID)
			return
		} else if len(replies) == c.Cfg.FastQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			slog.Debug("-----CQ REACHED ----- ", config.ID, req.ID)
			c.StablePropose(&req)
			return
		}
	}
}

func (c *Caesar) ReceiveSlowPropose(sp model.Request) (model.Response, bool) {
	slog.Debug("Received Slow Propose", config.ID, sp.ID, config.FROM, sp.Proposer, config.TIMESTAMP, sp.Timestamp)
	c.Ballots.Set(sp.ID, sp.Ballot)
	c.Clock.SetTimestamp(sp.Timestamp)
	var req model.Request

	if c.Decided.Contains(sp.ID) {
		return model.Response{}, false
	}

	if !c.History.Has(sp.ID) {
		req = sp
		c.History.Set(req.ID, req)
	} else {
		req, _ = c.History.Get(sp.ID)
	}

	req, _ = c.History.Get(sp.ID)
	if req.Status == model.STABLE || req.Status == model.ACC {
		return model.Response{}, false
	}

	if c.Decided.Contains(sp.ID) {
		return model.Response{}, false
	}

	outcome, aborted := c.wait(req.ID, req.Payload, req.Timestamp)
	if aborted {
		return model.Response{}, false
	}
	slog.Debug("Computed outcome", config.ID, req.ID, config.OUTCOME, outcome)

	req, _ = c.History.Get(sp.ID)
	if req.Status == model.STABLE || req.Status == model.ACC {
		return model.Response{}, false
	}

	if !outcome {
		req.Status = model.REJ
		req.Forced = false
		c.History.Set(req.ID, req)
		newTS := c.Clock.NewTimestamp()
		newPred := c.computePred(req.ID, req.Payload, newTS, nil)
		return model.NewResponse(req.ID, model.FASTP_REPLY, outcome, newPred, c.Cfg.ID, newTS, req.Ballot), true
	} else {
		req.Status = model.SLOW_PEND
		req.Forced = false
		c.History.Set(req.ID, req)
		return model.NewResponse(req.ID, model.FASTP_REPLY, outcome, req.Pred, c.Cfg.ID, req.Timestamp, req.Ballot), true
	}
}
