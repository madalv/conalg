package caesar

import (
	"conalg/model"
	"errors"
	"time"

	"github.com/gookit/slog"
)

func (c *Caesar) StablePropose(req model.Request) {
	slog.Debugf("Stable Proposing %s", req.Payload)
	c.Transport.BroadcastStablePropose(&req)
}

func (c *Caesar) ReceiveStablePropose(sp model.Request) error {
	time.Sleep(2 * time.Second)
	slog.Debugf("Received stable propose for %s: %v", sp.Payload, sp)
	c.Ballots.Set(sp.ID, sp.Ballot)
	var req model.Request

	if !c.History.Has(sp.ID) {
		slog.Errorf("Received stable propose for unknown request %+v", sp.ID)
		req = sp
	} else {
		req, _ = c.History.Get(sp.ID)
	}

	req.Status = model.STABLE
	c.History.Set(req.ID, req)

	// break loop
	err := c.breakLoop(req.ID)
	for err != nil {
		time.Sleep(10 * time.Second)
		err = c.breakLoop(req.ID)
	}

	if c.deliverable(req.ID) {
		c.deliver(req)
		return nil
	}

	ch := c.Publisher.Subscribe()
	defer c.Publisher.CancelSubscription(ch)

	for update := range ch {
		if update.Status == model.ACC && req.Pred.Contains(update.RequestID) {
			if c.deliverable(req.ID) {
				c.deliver(req)
				return nil
			}
		}
	}

	return nil
}

func (c *Caesar) deliver(req model.Request) {
	c.Decided.Add(req.ID)
	c.Publisher.Publish(model.StatusUpdate{
		RequestID: req.ID,
		Status:    model.ACC,
		Pred:      req.Pred,
	})
	c.Executer.Execute(req.Payload)
	c.History.Remove(req.ID)
	slog.Debugf("Request %s delivered", req.Payload)
}

func (c *Caesar) breakLoop(id string) error {
	req, ok := c.History.Get(id)
	if !ok {
		slog.Errorf("Couldn't retrieve key %s", id)
		return errors.New("request not found")
	}

	if req.Status != model.STABLE {
		slog.Errorf("Request %s is not stable", id)
		return errors.New("request not stable")
	}

	iterator := req.Pred.Iter()
	for predID := range iterator {
		pred, ok := c.History.Get(predID)
		if !ok && c.Decided.Contains(predID) {
			continue
		} else if !ok {
			slog.Errorf("Couldn't retrieve key %s", id)
			return errors.New("request not found")
		}

		if pred.Status == model.STABLE {
			if pred.Timestamp < req.Timestamp {
				pred.Pred.Remove(id)
				c.History.Set(pred.ID, pred)
			} else if pred.Timestamp > req.Timestamp {
				req.Pred.Remove(predID)
			}
		}
	}
	c.History.Set(req.ID, req)
	return nil
}

func (c *Caesar) deliverable(id string) bool {
	req, ok := c.History.Get(id)
	if !ok {
		slog.Errorf("Couldn't retrieve key %s", id)
		return false
	}
	res := req.Pred.IsSubset(c.Decided)
	slog.Debugf("Request %s is deliverable: %t", id, res)
	return res
}