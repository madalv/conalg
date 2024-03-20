package util

import (
	"conalg/model"
	"time"

	"github.com/gookit/slog"
)

type Analyzer struct {
	reqChannel          chan model.Request
	nrRequests          int``
	avgTotalDuration    time.Duration
	avgDeliveryDuration time.Duration
	avgProposalDuration time.Duration
}

func NewAnalyzer() *Analyzer {

	a := &Analyzer{
		reqChannel: make(chan model.Request, 100),
	}

	go func() {
		for req := range a.reqChannel {
			a.processRequest(req)
		}
	}()
	return a
}

func (a *Analyzer) SendReq(req model.Request) {
	a.reqChannel <- req
}

func (a *Analyzer) processRequest(req model.Request) {
	now := time.Now()
	totalDuration := now.Sub(req.ProposeTime)
	deliveryDuration := now.Sub(req.StableTime)
	proposalDuration := req.StableTime.Sub(req.ProposeTime)

	a.nrRequests += 1

	a.avgTotalDuration = (a.avgTotalDuration*time.Duration(a.nrRequests) + totalDuration) / time.Duration(a.nrRequests)
	a.avgDeliveryDuration = (a.avgDeliveryDuration*time.Duration(a.nrRequests) + deliveryDuration) / time.Duration(a.nrRequests)
	a.avgProposalDuration = (a.avgProposalDuration*time.Duration(a.nrRequests) + proposalDuration) / time.Duration(a.nrRequests)

	slog.Infof("Request %s took %d ms to be delivered, %d ms to be proposed and %d ms in total", req.ID, deliveryDuration.Milliseconds(), proposalDuration.Milliseconds(), totalDuration.Milliseconds())
}
