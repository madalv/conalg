package caesar

import (
	"conalg/model"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

// TODO separate model for propose from requests!!!
// TODO timeout as var
func (c *Caesar) FastPropose(reqID string) {
	req, ok := c.History.Get(reqID)
	if !ok {
		slog.Fatalf("Request %s not found in history", reqID)
	}
	replies := map[string]model.Response{}
	maxTimestamp := req.Timestamp
	pred := gs.NewSet[string]()

	c.Transport.BroadcastFastPropose(&req)
	broadcastTime := time.Now()

	for reply := range req.ResponseChan {
		req, _ := c.History.Get(reqID)
		slog.Debugf("Received for %s reply %s %s %t %d %d", req.ID, reply.From, reply.Type, reply.Result, reply.Pred.Cardinality(), reply.Timestamp)

		if reply.Type != model.FASTP_REPLY {
			slog.Warnf("Received unexpected reply %+v for  %s ", reply, req.ID)
		}

		if _, ok := replies[reply.From]; ok {
			slog.Warnf("Received duplicate reply %+v for %s ", reply, req.ID)
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
			slog.Debugf("----- REPLIES have NACK. must RETRY for %s ", req.ID)
			go c.RetryPropose(req)
			return
		} else if len(replies) == c.Cfg.FastQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			slog.Debugf("-----FQ REACHED ----- for %s", req.ID)
			c.StablePropose(req)
			return
		} else if time.Since(broadcastTime) >= 200*time.Millisecond && len(replies) == c.Cfg.ClassicQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			// TODO send slow propose
			slog.Debugf("----- CQ REACHED (timeout exceeded, slow proposing) ----- for %s", req.ID)
			return
		}
	}
}

func (c *Caesar) ReceiveFastPropose(fp model.Request) (model.Response, bool) {

	slog.Debugf("Received Fast Propose for %s %s %s %d", fp.ID, fp.Payload, fp.Proposer, fp.Timestamp)
	c.Ballots.Set(fp.ID, fp.Ballot)
	c.Clock.SetTimestamp(fp.Timestamp)
	var req model.Request

	if !c.History.Has(fp.ID) {
		// slog.Debugf("Adding new request %s to history", fp.ID)
		req = fp
		c.History.Set(req.ID, req)
	} else {
		req, _ = c.History.Get(fp.ID)
	}

	pred := c.computePred(req.ID, req.Payload, req.Timestamp, req.Whitelist)
	slog.Debugf("Computed pred for %s: %d %s", req.ID, req.Pred.Cardinality(), req.Status)

	req, _ = c.History.Get(fp.ID)
	if req.Status == model.STABLE || req.Status == model.ACC {
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
	slog.Debugf("Computed outcome for %s: %t", req.ID, outcome)

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
