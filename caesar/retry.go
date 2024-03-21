package caesar

import (
	"conalg/model"

	"github.com/gookit/slog"
)

func (c *Caesar) RetryPropose(req model.Request) {
	slog.Infof("Retrying %s %s", req.Payload, req.ID)
	c.Transport.BroadcastRetryPropose(&req)
}

func (c *Caesar) ReceiveRetryPropose(req model.Request) model.Response {
	// TODO retry
	return model.Response{}
}
