package caesar

import (
	"time"

	"log/slog"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/model"
)

// TODO separate model for propose from requests!!!
// TODO timeout as var
func (c *Caesar) FastPropose(reqID string) {
	req, ok := c.History.Get(reqID)
	if !ok {
		slog.Warn("Request not found in history", config.ID, reqID)
	}
	replies := map[string]model.Response{}
	maxTimestamp := req.Timestamp
	pred := gs.NewSet[string]()

	c.Transport.BroadcastFastPropose(&req)
	broadcastTime := time.Now()

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
			go c.RetryPropose(req)
			return
		} else if len(replies) == c.Cfg.FastQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			slog.Debug("-----FQ REACHED ----- ", config.ID, req.ID)
			c.StablePropose(req)
			return
		} else if time.Since(broadcastTime) >= 200*time.Millisecond && len(replies) == c.Cfg.ClassicQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			// TODO send slow propose
			slog.Debug("----- CQ REACHED (timeout exceeded, slow proposing) -----", config.ID, req.ID)
			return
		}
	}
}

func (c *Caesar) ReceiveFastPropose(fp model.Request) (model.Response, bool) {
	slog.Debug("Received Fast Propose", config.ID, fp.ID, config.FROM, fp.Proposer, config.TIMESTAMP, fp.Timestamp)
	c.Ballots.Set(fp.ID, fp.Ballot)
	c.Clock.SetTimestamp(fp.Timestamp)
	var req model.Request

	if c.Decided.Contains(fp.ID) {
		return model.Response{}, false
	}

	if !c.History.Has(fp.ID) {
		req = fp
		c.History.Set(req.ID, req)
	} else {
		req, _ = c.History.Get(fp.ID)
	}

	pred := c.computePred(req.ID, req.Payload, req.Timestamp, req.Whitelist)
	slog.Debug("Computed pred", config.ID, req.ID, config.PRED, pred.Cardinality(), config.STATUS, req.Status)

	req, _ = c.History.Get(fp.ID)
	if req.Status == model.STABLE || req.Status == model.ACC {
		return model.Response{}, false
	}

	if c.Decided.Contains(fp.ID) {
		return model.Response{}, false
	}

	req.Status = model.FAST_PEND
	req.Forced = !req.Whitelist.IsEmpty()
	req.Pred = pred
	c.History.Set(req.ID, req)

	outcome, aborted := c.wait(req.ID, req.Payload, req.Timestamp)
	if aborted {
		return model.Response{}, false
	}
	slog.Debug("Computed outcome", config.ID, req.ID, config.OUTCOME, outcome)

	req, _ = c.History.Get(fp.ID)
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
		return model.NewResponse(req.ID, model.FASTP_REPLY, outcome, req.Pred, c.Cfg.ID, req.Timestamp, req.Ballot), true
	}
}
