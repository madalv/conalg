package caesar

import (
	"conalg/models"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

// TODO what the fuck is up with the damn timeout??
func (c *Caesar) FastPropose(req models.Request) {
	slog.Debugf("Fast Proposing %s", req)

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

// TODO remove whitelist from actual request?
func (c *Caesar) ReceiveFastPropose(req models.Request) {
	go func(req models.Request) {
		slog.Debugf("Received Fast Propose %s", req)
		c.Ballots.Set(req.ID, req.Ballot)
		c.Clock.SetTimestamp(req.Timestamp)

		if !c.History.Has(req.ID) {
			slog.Debugf("Adding new request %s to history", req.ID)
			req.Status = models.PRE_FAST_PEND
			c.History.Set(req.ID, req)
		}

		req.Status = models.FAST_PEND
		req.Forced = !req.Whitelist.IsEmpty()
		req.Pred = c.computePred(req.ID, req.Payload, req.Timestamp, req.Whitelist)
		slog.Debug(req.Pred)

		c.History.Set(req.ID, req)

	}(req)
}

func (c *Caesar) ReceiveFastProposeResponse(r models.Response) {
	req, ok := c.History.Get(r.RequestID)
	if !ok {
		slog.Warnf("Received response for unknown request %s", r.RequestID)
		return
	}
	req.ResponseChan <- r
}
