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
	pred := gs.NewThreadUnsafeSet[string]()

	c.Transport.BroadcastFastPropose(&req)
	broadcastTime := time.Now()

	for reply := range req.ResponseChan {
		req, _ := c.History.Get(reqID)
		slog.Debugf("Received %+v for req %s /%s ", reply, req.Payload, req.ID)

		if reply.Type != model.FASTP_REPLY {
			slog.Warnf("Received unexpected reply %+v for req %s /%s ", reply, req.Payload, req.ID)
		}

		if _, ok := replies[reply.From]; ok {
			slog.Warnf("Received duplicate reply %+v for req %s /%s ", reply, req.Payload, req.ID)
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
			slog.Infof("REPLIES have NACK. must RETRY.  req %s %s", req.Payload, req.ID)
			go c.RetryPropose(req)
			return
		} else if len(replies) == c.Cfg.FastQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			slog.Infof("FQ REACHED ----- %s %s", req.Payload, req.ID)
			c.StablePropose(req)
			return
		} else if time.Since(broadcastTime) >= 200*time.Millisecond && len(replies) == c.Cfg.ClassicQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			// TODO send slow propose
			slog.Infof("CQ REACHED (timeout exceeded, slow proposing) ----- %s %s", req.Payload, req.ID)
			return
		}
	}
}

func (c *Caesar) ReceiveFastPropose(fp model.Request) model.Response {
	slog.Debugf("Received Fast Propose %v", fp)
	c.Ballots.Set(fp.ID, fp.Ballot)
	c.Clock.SetTimestamp(fp.Timestamp)
	var req model.Request

	if !c.History.Has(fp.ID) {
		slog.Debugf("Adding new request %s /%s to history", fp.Payload, fp.ID)
		req = fp
		c.History.Set(req.ID, req)
	} else {
		req, _ = c.History.Get(fp.ID)
	}

	req.Status = model.FAST_PEND
	req.Forced = !req.Whitelist.IsEmpty()
	req.Pred = c.computePred(req.ID, req.Payload, req.Timestamp, req.Whitelist)
	slog.Debugf("computed pred /%s: %v %s", req.ID, req.Pred, req.Status)

	c.History.Set(req.ID, req)

	outcome := c.wait(req.ID, req.Payload, req.Timestamp)
	slog.Debugf("computed outcome /%s: %b", req.ID, outcome)
	if !outcome {
		req.Status = model.REJ
		req.Forced = false
		c.History.Set(req.ID, req)
		newTS := c.Clock.NewTimestamp()
		newPred := c.computePred(req.ID, req.Payload, newTS, nil)
		return model.NewResponse(req.ID, model.FASTP_REPLY, outcome, newPred, c.Cfg.ID, newTS, req.Ballot)
	} else {
		return model.NewResponse(req.ID, model.FASTP_REPLY, outcome, req.Pred, c.Cfg.ID, req.Timestamp, req.Ballot)
	}
}
