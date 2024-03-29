package caesar

import (
	"conalg/model"
	"errors"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/gookit/slog"
)

func (c *Caesar) StablePropose(req model.Request) {
	slog.Debugf("Stable Proposing %s", req.ID)
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
		Pred:      gs.NewSet[string](req.Pred.ToSlice()...),
	}

	c.Publisher.Publish(update)

	slog.Debugf("-----------------> Published status update for %s %s %s", update.RequestID, update.Status, update.Pred.String())

	if c.AnalyzerEnabled {
		c.Analyzer.SendStable(req)
	}

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
			update.Status == "PRED_REMOVAL" || update.Status == "DECIDED" {
			slog.Debugf("----> Received update for %s: %v", req.ID, update)
			c.breakLoop(req.ID)
			if c.deliverable(req.ID) {
				c.Publisher.CancelSubscription(ch)
				c.deliver(req.ID)
				return nil
			}
		}
	}

	return nil
}

func (c *Caesar) deliver(id string) {

	req, _ := c.History.Get(id)
	c.Decided.Add(req.ID)

	update := model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.DECIDED,
		Pred:      gs.NewSet[string](req.Pred.ToSlice()...),
	}
	c.Publisher.Publish(update)

	slog.Debugf("-----------------> Published status update for %s %s %s", update.RequestID, update.Status, update.Pred.String())
	c.Executer.Execute(req.Payload)
	c.History.Remove(req.ID)
	slog.Debugf("Request %s %s delivered", req.Payload, req.ID)

	if c.AnalyzerEnabled {
		c.Analyzer.SendReq(req)
	}
}

func (c *Caesar) breakLoop(id string) error {
	// slog.Infof("Breaking loop for %s", id)
	req, ok := c.History.Get(id)
	if !ok {
		slog.Errorf("Couldn't retrieve key %s", id)
		return errors.New("request not found")
	}

	iterator := req.Pred.Iter()
	newPredSet := gs.NewSet[string](req.Pred.ToSlice()...)
	for predID := range iterator {
		pred, ok := c.History.Get(predID)

		if !ok && c.Decided.Contains(predID) {
			continue
		} else if !ok {
			slog.Errorf("Couldn't retrieve key %s", id)
			continue
		}

		if pred.Status == model.STABLE {

			if pred.Timestamp < req.Timestamp && pred.Pred.Contains(id) {
				// slog.Infof("-> Pred %s (TS %d) contains current request %s (TS %d)", predID, pred.Timestamp, req.ID, req.Timestamp)
				pred.Pred.Remove(id)

				c.History.Set(pred.ID, pred)

				// slog.Infof("-> Removed %s from %s's pred", id, pred.ID)
				update := model.StatusUpdate{
					RequestID: pred.ID,
					Status:    "PRED_REMOVAL",
				}

				c.Publisher.Publish(update)
				// slog.Infof("-----------------> Published status update for %s %s", update.RequestID, update.Status)

			} else if pred.Timestamp >= req.Timestamp {
				// slog.Infof("-> Current request %s has a pred %s with higher timestamp %d", id, predID, pred.Timestamp)
				newPredSet.Remove(predID)
				// slog.Infof("-> Removed %s from %s's pred", predID, id)
			}
		}
	}

	req.Pred = newPredSet
	c.History.Set(req.ID, req)
	return nil
}

func (c *Caesar) deliverable(id string) bool {
	req, ok := c.History.Get(id)
	if !ok {
		slog.Errorf("Couldn't retrieve key %s", id)
		return false
	}

	pred := gs.NewSet[string](req.Pred.ToSlice()...)

	res := pred.IsSubset(c.Decided)
	if !res {
		slog.Warnf("---> Request for %s not deliverable: %s", id, pred.Difference(c.Decided).String())
	} else {
		slog.Debugf("---> Request for %s deliverable", id)
	}

	return res
}
