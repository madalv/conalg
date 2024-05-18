package util

import (
	"sync"
	"time"

	"log/slog"

	"github.com/madalv/conalg/model"
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
	nrRequestsInOrder       int64
	mu                      sync.Mutex
	promMetrics             *Metrics
	nodeID                  string
}

func NewAnalyzer(active bool, nodeID string) *Analyzer {
	slog.Info("Created new Analyzer. . . ")

	a := &Analyzer{
		reqChannel:    make(chan model.Request),
		stableChannel: make(chan model.Request),
		timestamps:    make(map[string]uint64),
		active:        active,
		mu:            sync.Mutex{},
		nodeID:        nodeID,
		promMetrics:   NewMetrics(),
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
				if a.nrRequests > 0 {
					slog.Info("> Average total duration (ms)", "avg_total", a.avgTotalDuration)
					slog.Info("> Number of requests processed", "n_delivered", a.nrRequests)
					slog.Info("> Number of stable requests processed", "n_stable", a.nrStableRequests)
					slog.Info("> Percentage of requests in order", "percentage", float64(a.nrRequestsInOrder)/float64(a.nrRequests)*100)
					a.mu.Unlock()
				}

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
	a.mu.Lock()
	a.nrStableRequests += 1
	a.mu.Unlock()
}

func (a *Analyzer) processRequest(req model.Request) {
	now := time.Now()
	totalDuration := now.Sub(req.ProposeTime)

	// observe total duration of request!
	a.promMetrics.duration.Observe(float64(totalDuration.Milliseconds()))

	deliveryDuration := now.Sub(req.StableTime)
	proposalDuration := req.StableTime.Sub(req.ProposeTime)

	a.totalDeliveryDuration += deliveryDuration.Milliseconds()
	a.totalProposalDuration += proposalDuration.Milliseconds()
	a.totalProcessingDuration += totalDuration.Milliseconds()

	a.nrRequests += 1
	a.promMetrics.reqCounter.Inc()


	inOrder := float64(a.nrRequestsInOrder)/float64(a.nrRequests)*100
	a.promMetrics.inOrderPercentage.Set(inOrder)

	lastTs, ok := a.timestamps[string(req.Payload)]
	if !ok {
		a.timestamps[string(req.Payload)] = req.Timestamp
	}

	if req.Timestamp < lastTs {
		// slog.Warn("Request has a timestamp smaller than the last conflicting", "id", req.ID, "ts", req.Timestamp, "last_ts", lastTs)
	} else {
		a.nrRequestsInOrder += 1
		a.timestamps[string(req.Payload)] = req.Timestamp
	}

	a.mu.Lock()
	a.avgTotalDuration = float64(a.totalProcessingDuration / a.nrRequests)
	a.avgDeliveryDuration = float64(a.totalDeliveryDuration / a.nrRequests)
	a.avgProposalDuration = float64(a.totalProposalDuration / a.nrRequests)
	a.mu.Unlock()

	if a.active {
		slog.Info("Request delivered with timestamps:",
			"id", req.ID, "stable->delivered", deliveryDuration.Milliseconds(), "proposed->stable", proposalDuration.Milliseconds(),
			"total", totalDuration.Milliseconds(), "ts", req.Timestamp, "payload", string(req.Payload), "last_ts", lastTs)
	}
}
