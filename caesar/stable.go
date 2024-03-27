package caesar

import (
	"conalg/model"
	"errors"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

func (c *Caesar) StablePropose(req model.Request) {
	slog.Infof("Stable Proposing %s %s", req.Payload, req.ID)
	c.Transport.BroadcastStablePropose(&req)
}

func (c *Caesar) ReceiveStablePropose(sp model.Request) error {
	// time.Sleep(5 * time.Second)
	slog.Debugf("Received stable propose for %s %s: %+v", sp.ID, sp.Payload, sp)
	c.Ballots.Set(sp.ID, sp.Ballot)
	var req model.Request

	if !c.History.Has(sp.ID) {
		slog.Errorf("Received stable propose for unknown request %+v", sp.ID)
		req = sp
	} else {
		req, _ = c.History.Get(sp.ID)
	}

	req.Status = model.STABLE
	req.StableTime = time.Now()
	c.History.Set(req.ID, req)

	update := model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.STABLE,
		Pred:      req.Pred,
	}

	c.Publisher.Publish(update)

	slog.Warnf("Published status update for %v", update)

	// break loop
	err := c.breakLoop(req.ID)
	for err != nil {
		slog.Error(err)
	}

	if c.deliverable(req.ID) {
		c.deliver(req.ID)
		return nil
	}

	ch := c.Publisher.Subscribe()

	for update := range ch {
		if req.Pred.Contains(update.RequestID) ||
			update.Status == "PRED_REMOVAL" || update.Status == "STABLE" || update.Status == "DECIDED" {
			slog.Infof("----> Received update %v for %s", update, req.ID)
			c.breakLoop(req.ID)
			if c.deliverable(req.ID) {
				c.deliver(req.ID)
				// c.Publisher.CancelSubscription(ch)
				return nil
			}
		}
	}

	return nil
}

func (c *Caesar) deliver(id string) {

	req, _ := c.History.Get(id)
	// c.decidedMu.Lock()
	c.Decided.Add(req.ID)
	// c.decidedMu.Unlock()

	update := model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.DECIDED,
		Pred:      req.Pred,
	}
	c.Publisher.Publish(update)

	slog.Warnf("Published status update for %v", update)
	c.Executer.Execute(req.Payload)
	c.History.Remove(req.ID)
	slog.Debugf("Request %s %s delivered", req.Payload, req.ID)

	c.Analyzer.SendReq(req)
}

func (c *Caesar) breakLoop(id string) error {
	slog.Debugf("Breaking loop for %s", id)
	req, ok := c.History.Get(id)
	if !ok {
		slog.Errorf("Couldn't retrieve key %s", id)
		return errors.New("request not found")
	}

	iterator := req.Pred.Iter()
	for predID := range iterator {
		slog.Infof("Checking pred %s (%s)", predID, req.ID)
		pred, ok := c.History.Get(predID)
		// c.decidedMu.RLock()
		if !ok && c.Decided.Contains(predID) {
			slog.Infof("Request %s has a decided pred %s", id, predID)
			continue
		} else if !ok {
			slog.Errorf("Couldn't retrieve key %s", id)
			continue
		}
		// c.decidedMu.RUnlock()

		if pred.Status == model.STABLE {

			if pred.Timestamp < req.Timestamp && pred.Pred.Contains(id) {
				slog.Infof("-> Current request %s has a pred %s with lower timestamp %d", id, predID, pred.Timestamp)
				pred.Pred.Remove(id)

				c.History.Set(pred.ID, pred)

				slog.Infof("Removed %s from %s's pred", id, pred.ID)
				c.Publisher.Publish(model.StatusUpdate{
					RequestID: pred.ID,
					Status:    "PRED_REMOVAL",
				})
				slog.Warn("Published pred removal for", predID)

			} else if pred.Timestamp >= req.Timestamp {
				slog.Infof("-> Current request %s has a pred %s with higher timestamp %d", id, predID, pred.Timestamp)
				req.Pred.Remove(predID)
				slog.Infof("Removed %s from %s's pred", predID, id)
			}
		}
	}
	slog.Infof("Finished breaking loop for %s", id)

	c.History.Set(req.ID, req)

	slog.Infof("History set for", id)
	return nil
}

func (c *Caesar) deliverable(id string) bool {
	slog.Infof("Checking if %s is deliverable", id)
	req, ok := c.History.Get(id)
	if !ok {
		slog.Errorf("Couldn't retrieve key %s", id)
		return false
	}

	pred := gs.NewSet[string](req.Pred.ToSlice()...)

	// c.decidedMu.RLock()
	res := pred.IsSubset(c.Decided)
	if !res {
		slog.Errorf(pred.Difference(c.Decided).String())
	}

	// c.decidedMu.RUnlock()
	slog.Debugf("Request %s %s is deliverable: %t", id, req.Payload, res)
	return res
}
