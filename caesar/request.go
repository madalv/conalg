package caesar

import (
	"conalg/pb"
	"time"

	gs "github.com/deckarep/golang-set/v2"
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

func (r *Request) ToFastProposePb() *pb.FastPropose {
	return &pb.FastPropose{
		Id:        r.ID,
		Payload:   r.Payload,
		Time:      r.Timestamp,
		Whitelist: nil,
		Ballot:    uint64(r.Ballot),
	}
}
