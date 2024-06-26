package model

import (
	"fmt"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/madalv/conalg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Status string

const (
	DEFAULT   Status = "DEFAULT" // initial state, pre-fast
	FAST_PEND Status = "FAST_PENDING"
	SLOW_PEND Status = "SLOW_PENDING"
	ACC       Status = "ACCEPTED" // request accepted on retry
	REJ       Status = "REJECTED"
	DECIDED   Status = "DECIDED" // request delivered - final state
	STABLE    Status = "STABLE"  // request stable but not delivered
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
		Pred:         gs.NewSet[string](),
		Status:       DEFAULT,
		Ballot:       0,
		Forced:       false,
		ResponseChan: make(chan Response, 1000),
		ProposeTime:  time.Now(),
		Proposer:     fmt.Sprintf("NODE_%d", proposer),
		Whitelist:    gs.NewSet[string](),
	}
}

func FromProposePb(fp *pb.Propose) Request {
	return Request{
		ID:          fp.RequestId,
		Payload:     fp.Payload,
		Timestamp:   fp.Timestamp,
		Whitelist:   gs.NewSet(fp.Whitelist...),
		Ballot:      uint(fp.Ballot),
		Pred:        gs.NewSet(fp.Pred...),
		ProposeTime: fp.ProposerTime.AsTime(),
	}
}

func (r *Request) ToProposePb(propType string) *pb.Propose {
	if r.Whitelist == nil {
		r.Whitelist = gs.NewSet[string]()
	}
	return &pb.Propose{
		RequestId:    r.ID,
		Payload:      r.Payload,
		Timestamp:    r.Timestamp,
		Whitelist:    r.Whitelist.ToSlice(),
		Ballot:       uint64(r.Ballot),
		Type:         propType,
		From:         r.Proposer,
		Pred:         r.Pred.ToSlice(),
		ProposerTime: timestamppb.New(r.ProposeTime),
	}
}
