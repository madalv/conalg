package util

import (
	"conalg/model"
	"sync"
	"time"

	"github.com/gookit/slog"
)

type Analyzer struct {
	active                  bool
	reqChannel              chan model.Request
	stableChannel           chan model.Request
	nrStableRequests        int64
	nrRequests              int64
	timestamps              map[string]uint64
	avgTotalDuration        float64
	avgDeliveryDuration     float64
	avgProposalDuration     float64
	totalDeliveryDuration   int64
	totalProposalDuration   int64
	totalProcessingDuration int64
	mu                      sync.Mutex
}

func NewAnalyzer(active bool) *Analyzer {

	a := &Analyzer{
		reqChannel:    make(chan model.Request),
		stableChannel: make(chan model.Request),
		timestamps:    make(map[string]uint64),
		active:        active,
		mu:            sync.Mutex{},
	}

	go func() {
		for req := range a.reqChannel {
			a.processRequest(req)
		}
	}()

	go func() {
		for req := range a.stableChannel {
			a.processStable(req)
		}
	}()

	ticker := time.NewTicker(10 * time.Second)

	if a.active {
		go func() {
			for range ticker.C {
				a.mu.Lock()
				slog.Infof("> Average total duration: %f ms", a.avgTotalDuration)
				slog.Infof("> Average delivery duration: %f ms", a.avgDeliveryDuration)
				slog.Infof("> Average proposal duration: %f ms", a.avgProposalDuration)
				slog.Infof("> Number of requests processed: %d", a.nrRequests)
				slog.Infof("> Number of stable requests processed: %d", a.nrStableRequests)
				a.mu.Unlock()
			}
		}()
	}

	return a
}

func (a *Analyzer) SendReq(req model.Request) {
	a.reqChannel <- req
}

func (a *Analyzer) SendStable(req model.Request) {
	a.stableChannel <- req
}

func (a *Analyzer) processStable(model.Request) {
	a.nrStableRequests += 1
}

func (a *Analyzer) processRequest(req model.Request) {
	now := time.Now()
	totalDuration := now.Sub(req.ProposeTime)
	deliveryDuration := now.Sub(req.StableTime)
	proposalDuration := req.StableTime.Sub(req.ProposeTime)

	a.mu.Lock()
	a.totalDeliveryDuration += deliveryDuration.Milliseconds()
	a.totalProposalDuration += proposalDuration.Milliseconds()
	a.totalProcessingDuration += totalDuration.Milliseconds()
	a.mu.Unlock()

	a.nrRequests += 1
	lastTs, ok := a.timestamps[string(req.Payload)]
	if !ok {
		a.timestamps[string(req.Payload)] = req.Timestamp
	}

	if req.Timestamp < lastTs {
		slog.Warnf("Request %s has a timestamp %d smaller than the last conflicting one %d", req.ID, req.Timestamp, lastTs)
	} else {
		a.timestamps[string(req.Payload)] = req.Timestamp
	}

	a.mu.Lock()
	a.avgTotalDuration = float64(a.totalProcessingDuration / a.nrRequests)
	a.avgDeliveryDuration = float64(a.totalDeliveryDuration / a.nrRequests)
	a.avgProposalDuration = float64(a.totalProposalDuration / a.nrRequests)
	a.mu.Unlock()

	if a.active {
		slog.Infof("Request %s took %d ms stable->delivered, %d ms proposed->stable and %d ms in total. Timestamps: %d, last for payload %s: %d",
			req.ID, deliveryDuration.Milliseconds(), proposalDuration.Milliseconds(), totalDuration.Milliseconds(), req.Timestamp, string(req.Payload), lastTs)
	}
}
