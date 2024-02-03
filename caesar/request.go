package caesar

import (
	"conalg/pb"
	"time"

	gs "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
)

const (
	FAST_PEND = "FAST_PENDING"
	SLOW_PEND = "SLOW_PENDING"
	ACC       = "ACCEPTED"
	REJ       = "REJECTED"
	STABLE    = "STABLE"
	WAITING   = "WAITING"
)

type Request struct {
	Proposer     string
	ID           string
	Payload      []byte
	Timestamp    uint64
	Pred         gs.Set[string]
	Status       string
	Ballot       uint
	Forced       bool
	ProposeTime  time.Time
	StableTime   time.Time
	ResponseChan chan Response
}

func NewRequest(payload []byte, ts uint64, fq int, proposer string) Request {
	return Request{
		ID:           uuid.NewString(),
		Payload:      payload,
		Timestamp:    ts,
		Pred:         gs.NewSet[string](),
		Status:       WAITING,
		Ballot:       0,
		Forced:       false,
		ResponseChan: make(chan Response, fq),
		ProposeTime:  time.Now(),
		Proposer:     proposer,
	}
}

func (r *Request) ToFastProposePb() *pb.FastPropose {
	return &pb.FastPropose{
		RequestId: r.ID,
		Payload:   r.Payload,
		Time:      r.Timestamp,
		Whitelist: nil,
		Ballot:    uint64(r.Ballot),
	}
}
