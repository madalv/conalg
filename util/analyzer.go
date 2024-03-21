package util

import (
	"conalg/model"
	"time"

	"github.com/gookit/slog"
)

type Analyzer struct {
	reqChannel              chan model.Request
	nrRequests              int64
	timestamps              map[string]uint64
	avgTotalDuration        float64
	avgDeliveryDuration     float64
	avgProposalDuration     float64
	totalDeliveryDuration   int64
	totalProposalDuration   int64
	totalProcessingDuration int64
}

func NewAnalyzer() *Analyzer {

	a := &Analyzer{
		reqChannel: make(chan model.Request, 100),
		timestamps: make(map[string]uint64),
	}

	go func() {
		for req := range a.reqChannel {
			a.processRequest(req)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for range ticker.C {
			slog.Infof("> Average total duration: %f ms", a.avgTotalDuration)
			slog.Infof("> Average delivery duration: %f ms", a.avgDeliveryDuration)
			slog.Infof("> Average proposal duration: %f ms", a.avgProposalDuration)
			slog.Infof("> Number of requests processed: %d", a.nrRequests)
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

	a.totalDeliveryDuration += deliveryDuration.Milliseconds()
	a.totalProposalDuration += proposalDuration.Milliseconds()
	a.totalProcessingDuration += totalDuration.Milliseconds()

	a.nrRequests += 1
	lastTs, ok := a.timestamps[string(req.Payload)]
	if !ok {
		a.timestamps[string(req.Payload)] = req.Timestamp
	}

	if req.Timestamp < lastTs {
		slog.Warnf("Request %s has a timestamp %d smaller than the last conflicting one %d", req.ID, req.Timestamp, lastTs)
	}

	a.avgTotalDuration = float64(a.totalProcessingDuration / a.nrRequests)
	a.avgDeliveryDuration = float64(a.totalDeliveryDuration / a.nrRequests)
	a.avgProposalDuration = float64(a.totalProposalDuration / a.nrRequests)

	slog.Infof("Request %s took %d ms stable->delivered, %d ms proposed->stable and %d ms in total. Timestamps: %d, last for payload %s: %d", 
	req.ID, deliveryDuration.Milliseconds(), proposalDuration.Milliseconds(), totalDuration.Milliseconds(), req.Timestamp, string(req.Payload), lastTs)
}
