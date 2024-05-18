package caesar

import (
	"errors"
	"time"

	"log/slog"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/madalv/conalg/config"
	"github.com/madalv/conalg/model"
)

func (c *Caesar) StablePropose(req *model.Request) {
	slog.Debug("Stable Proposing for", config.ID, req.ID)
	c.Transport.BroadcastStablePropose(req)
}

func (c *Caesar) ReceiveStablePropose(sp model.Request) error {
	slog.Debug("Received Stable Propose", config.ID, sp.ID, config.FROM, sp.Proposer, config.TIMESTAMP, sp.Timestamp)

	c.Ballots.Set(sp.ID, sp.Ballot)
	var req model.Request

	if !c.History.Has(sp.ID) {
		slog.Error("Received stable propose for unknown request", config.ID, sp.ID)
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

	slog.Debug("-----------------> Published status update", config.UPDATE_ID, update.RequestID,
		config.UPDATE_STATUS, update.Status, config.UPDATE_PRED, update.Pred.Cardinality())

	if c.AnalyzerEnabled {
		c.Analyzer.SendStable(req)
	}

	// break loop
	err := c.breakLoop(req.ID)
	for err != nil {
		slog.Error(err.Error())
	}

	if c.deliverable(req.ID) {
		c.deliver(req.ID)
		return nil
	}

	ch := c.Publisher.Subscribe()

	for update := range ch {
		if req.Pred.Contains(update.RequestID) ||
			update.Status == "PRED_REMOVAL" || update.Status == "DECIDED" {
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
	c.Decided.Add(id)
	req, _ := c.History.Get(id)

	update := model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.DECIDED,
		Pred:      gs.NewSet(req.Pred.ToSlice()...),
	}
	c.Publisher.Publish(update)

	slog.Debug("-----------------> Published status update", config.UPDATE_ID, update.RequestID,
		config.UPDATE_STATUS, update.Status, config.UPDATE_PRED, update.Pred.Cardinality())
	c.Executer.Execute(req.Payload)
	c.History.Remove(req.ID)
	slog.Debug("Request delivered", config.ID, req.ID)

	if c.AnalyzerEnabled {
		c.Analyzer.SendReq(req)
	}
}

func (c *Caesar) breakLoop(id string) error {
	req, ok := c.History.Get(id)
	if !ok {
		slog.Error("Couldn't retrieve key", config.ID, id)
		return errors.New("request not found")
	}

	iterator := req.Pred.Iter()
	newPredSet := gs.NewSet(req.Pred.ToSlice()...)
	for predID := range iterator {
		pred, ok := c.History.Get(predID)

		if !ok && c.Decided.Contains(predID) {
			continue
		} else if !ok {
			slog.Warn("Couldn't retrieve pred key", config.ID, predID)
			continue
		}

		if pred.Status == model.STABLE {
			if pred.Timestamp < req.Timestamp && pred.Pred.Contains(id) {
				pred.Pred.Remove(id)

				c.History.Set(pred.ID, pred)
				update := model.StatusUpdate{
					RequestID: pred.ID,
					Status:    "PRED_REMOVAL",
				}
				c.Publisher.Publish(update)

			} else if pred.Timestamp >= req.Timestamp {
				newPredSet.Remove(predID)
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
		slog.Error("Couldn't retrieve key", config.ID, id)
		return false
	}
	pred := gs.NewSet(req.Pred.ToSlice()...)
	res := pred.IsSubset(c.Decided)
	if !res {
		slog.Debug("---> Request not deliverable", config.ID, id, config.DIFF, pred.Difference(c.Decided).Cardinality())
	} else {
		slog.Debug("---> Request deliverable", config.ID, id)
	}

	return res
}
