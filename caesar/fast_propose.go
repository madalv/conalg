package caesar

import (
	"conalg/models"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

// TODO what the fuck is up with the damn timeout??
func (c *Caesar) FastPropose(req models.Request) {
	// slog.Debugf("Fast Proposing %s", req)

	replies := map[string]models.Response{}
	maxTimestamp := req.Timestamp
	pred := gs.NewSet[string]()

	c.Transport.BroadcastFastPropose(&req)

	for {
		select {
		case reply := <-req.ResponseChan:
			slog.Debugf("Received %s for req %s", reply, req.ID)

			// slog.Debug(c.History.Items())

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

			// case <-time.After(30 * time.Second):
			// 	slog.Warnf("Fast Propose timed out for req %s, DO SMTH??", req.ID)
		}
	}

}

func (c *Caesar) ReceiveFastPropose(fp models.Request) models.Response {
	slog.Debugf("Received Fast Propose %v", fp)
	c.Ballots.Set(fp.ID, fp.Ballot)
	c.Clock.SetTimestamp(fp.Timestamp)
	var req models.Request

	if !c.History.Has(fp.ID) {
		slog.Debugf("Adding new request %s to history", fp.ID)
		req = fp
		req.Status = models.PRE_FAST_PEND
		c.History.Set(req.ID, req)
	} else {
		req, _ = c.History.Get(fp.ID)
	}

	req.Status = models.FAST_PEND
	req.Forced = !req.Whitelist.IsEmpty()
	req.Pred = c.computePred(req.ID, req.Payload, req.Timestamp, req.Whitelist)
	slog.Debug("computed pred: ", req.Pred)

	c.History.Set(req.ID, req)

	outcome := c.wait(req.ID, req.Payload, req.Timestamp)
	slog.Debug("computed outcome: ", outcome)
	if !outcome {
		req.Status = models.REJ
		req.Forced = false
		c.History.Set(req.ID, req)
		newTS := c.Clock.NewTimestamp()
		newPred := c.computePred(req.ID, req.Payload, newTS, nil)
		return models.NewResponse(req.ID, models.FASTP_REPLY, outcome, newPred, c.Cfg.ID, newTS, req.Ballot)
	} else {
		return models.NewResponse(req.ID, models.FASTP_REPLY, outcome, req.Pred, c.Cfg.ID, req.Timestamp, req.Ballot)
	}
}

func (c *Caesar) ReceiveFastProposeResponse(r models.Response) {
	req, ok := c.History.Get(r.RequestID)
	slog.Warn(req)
	if !ok {
		slog.Warnf("Received response for unknown request %s", r.RequestID)
		return
	}
	req.ResponseChan <- r
}
