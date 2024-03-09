package caesar

import (
	"conalg/model"
	"errors"

	"github.com/gookit/slog"
)

func (c *Caesar) ReceiveStablePropose(sp model.Request) error {
	c.Ballots.Set(sp.ID, sp.Ballot)
	var req model.Request

	if !c.History.Has(sp.ID) {
		slog.Errorf("Received stable propose for unknown request %s", sp.ID)
		req = sp
	} else {
		req, _ = c.History.Get(sp.ID)
	}

	req.Status = model.STABLE
	c.History.Set(req.ID, req)

	// break loop
	err := c.breakLoop(req.ID)
	if err != nil {
		return err
	}

	// deliverable?
	return nil
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
		if !ok {
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
