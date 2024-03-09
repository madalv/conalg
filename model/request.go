package model

import (
	"conalg/pb"
	"fmt"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
)

type Status string

const (
	DEFAULT       Status = "DEFAULT"
	PRE_FAST_PEND Status = "PRE_FAST_PENDING"
	FAST_PEND     Status = "FAST_PENDING"
	SLOW_PEND     Status = "SLOW_PENDING"
	ACC           Status = "ACCEPTED"
	REJ           Status = "REJECTED"
	STABLE        Status = "STABLE"
	DECIDED       Status = "DECIDED"
)

const (
	FASTP_PROP  = "FASTP_PROP"
	SLOWP_PROP  = "SLOWP_PROP"
	RETRY_PROP  = "RETRY_PROP"
	STABLE_PROP = "STABLE_PROP"
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
		Pred:         gs.NewThreadUnsafeSet[string](),
		Status:       DEFAULT,
		Ballot:       0,
		Forced:       false,
		ResponseChan: make(chan Response, fq),
		ProposeTime:  time.Now(),
		Proposer:     fmt.Sprintf("NODE_%d", proposer),
		Whitelist:    gs.NewThreadUnsafeSet[string](),
	}
}

func FromProposePb(fp *pb.Propose) Request {
	return Request{
		ID:        fp.RequestId,
		Payload:   fp.Payload,
		Timestamp: fp.Timestamp,
		Whitelist: gs.NewThreadUnsafeSet(fp.Whitelist...),
		Ballot:    uint(fp.Ballot),
		Pred:      gs.NewThreadUnsafeSet(fp.Pred...),
	}
}

func (r *Request) ToProposePb(propType string) *pb.Propose {
	if r.Whitelist == nil {
		r.Whitelist = gs.NewThreadUnsafeSet[string]()
	}
	return &pb.Propose{
		RequestId: r.ID,
		Payload:   r.Payload,
		Timestamp: r.Timestamp,
		Whitelist: r.Whitelist.ToSlice(),
		Ballot:    uint64(r.Ballot),
		Type:      propType,
		From:      r.Proposer,
		Pred:      r.Pred.ToSlice(),
	}
}
