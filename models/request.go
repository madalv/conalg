package models

import (
	"conalg/pb"
	"fmt"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
)

type Status string

const (
	FAST_PEND     Status = "FAST_PENDING"
	PRE_FAST_PEND Status = "PRE_FAST_PENDING"
	SLOW_PEND     Status = "SLOW_PENDING"
	ACC           Status = "ACCEPTED"
	REJ           Status = "REJECTED"
	STABLE        Status = "STABLE"
	DEFAULT       Status = "DEFAULT"
)

type Request struct {
	Proposer     string
	ID           string
	Payload      []byte
	Timestamp    uint64
	Pred         gs.Set[string]
	Status       Status
	Ballot       uint
	Forced       bool
	ProposeTime  time.Time
	StableTime   time.Time
	ResponseChan chan Response
	Whitelist    gs.Set[string]

}

func NewRequest(payload []byte, ts uint64, fq int, proposer uint64) Request {
	return Request{
		ID:           uuid.NewString(),
		Payload:      payload,
		Timestamp:    ts,
		Pred:         gs.NewSet[string](),
		Status:       DEFAULT,
		Ballot:       0,
		Forced:       false,
		ResponseChan: make(chan Response, fq),
		ProposeTime:  time.Now(),
		Proposer:     fmt.Sprintf("NODE_%d", proposer),
		Whitelist:    gs.NewSet[string](),
		
	}
}

func FromFastProposePb(fp *pb.FastPropose) Request {
	return Request{
		ID:        fp.RequestId,
		Payload:   fp.Payload,
		Timestamp: fp.Time,
		Whitelist: gs.NewSet(fp.Whitelist...),
		Ballot:    uint(fp.Ballot),
	}
}

func (r *Request) ToFastProposePb() *pb.FastPropose {
	if r.Whitelist == nil {
		r.Whitelist = gs.NewSet[string]()
	}
	return &pb.FastPropose{
		RequestId: r.ID,
		Payload:   r.Payload,
		Time:      r.Timestamp,
		Whitelist: r.Whitelist.ToSlice(),
		Ballot:    uint64(r.Ballot),
	}
}
