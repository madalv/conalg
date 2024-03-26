package caesar

import (
	"conalg/model"
	"errors"
	"fmt"
	"time"

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
	err := c.breakLoop(req.ID) // TODO add count for tries so it doesn't loop forever
	for err != nil {
		time.Sleep(10 * time.Second)
		slog.Error(err)
		err = c.breakLoop(req.ID)
	}

	if c.deliverable(req.ID) {
		c.deliver(req)
		return nil
	}

	ch := c.Publisher.Subscribe()

	for update := range ch {
		if req.Pred.Contains(update.RequestID) {
			c.breakLoop(req.ID)
			if c.deliverable(req.ID) {
				c.deliver(req)
				c.Publisher.CancelSubscription(ch)
				return nil
			}
		}
	}

	return nil
}

func (c *Caesar) deliver(req model.Request) {
	c.Decided.Add(req.ID)
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
	// slog.Debug(req.Proposer)
	if req.Proposer == fmt.Sprintf("NODE_%d", c.Cfg.ID) {
		c.Analyzer.SendReq(req)
	}
}

func (c *Caesar) breakLoop(id string) error {
	slog.Debugf("Breaking loop for %s", id)
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
	if !res {
		slog.Errorf(req.Pred.Difference(c.Decided).String())
	}
	slog.Debugf("Request %s %s is deliverable: %t", id, req.Payload, res)
	return res
}
