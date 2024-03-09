package caesar

import (
	"conalg/model"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

func (c *Caesar) FastPropose(req model.Request) {
	// slog.Debugf("Fast Proposing %s", req)

	replies := map[string]model.Response{}
	maxTimestamp := req.Timestamp
	pred := gs.NewSet[string]()

	c.Transport.BroadcastFastPropose(&req)

	for reply := range req.ResponseChan {
		slog.Debugf("Received %v for req %s", reply, req.ID)

		if reply.Type != model.FASTP_REPLY {
			slog.Warn("Received wrong type of reply ", reply.Type, " for req ", req.ID)
		}

		if _, ok := replies[reply.From]; ok {
			slog.Warnf("Received duplicate reply %s for req %s", reply, req.ID)
		} else {
			replies[reply.From] = reply
		}

		if reply.Timestamp > maxTimestamp {
			maxTimestamp = reply.Timestamp
		}

		pred.Union(reply.Pred)

		slog.Debug(pred, maxTimestamp)

		if repliesHaveNack(replies) {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			slog.Debugf("REPLIES have NACK. must RETRY.  req %s\n Replies: \n %s", req, replies)

			// TODO retry
			return
		} else if len(replies) == c.Cfg.FastQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			slog.Debugf("FQ REACHED ----- %s. \n Replies: \n %s", req, replies)
			// TODO stable
			return
		} else if len(replies) == c.Cfg.ClassicQuorum {
			req.Timestamp = maxTimestamp
			req.Pred = pred
			c.History.Set(req.ID, req)
			// TODO send slow propose
			slog.Debugf("CQ REACHED ----- %s. \n Replies: \n %s", req, replies)
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
		slog.Debugf("Adding new request %s to history", fp.ID)
		req = fp
		req.Status = model.PRE_FAST_PEND
		c.History.Set(req.ID, req)
	} else {
		req, _ = c.History.Get(fp.ID)
	}

	req.Status = model.FAST_PEND
	req.Forced = !req.Whitelist.IsEmpty()
	req.Pred = c.computePred(req.ID, req.Payload, req.Timestamp, req.Whitelist)
	slog.Debug("computed pred: ", req.Pred)

	c.History.Set(req.ID, req)

	outcome := c.wait(req.ID, req.Payload, req.Timestamp)
	slog.Debug("computed outcome: ", outcome)
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


