package caesar

import (
	"conalg/pb"

	gs "github.com/deckarep/golang-set/v2"
)

const (
	FASTP_REPLY = "FASTP_REPLY"
	SLOWP_REPLY = "SLOWP_REPLY"
	RETRY_REPLY = "RETRY_REPLY"
)

type Response struct {
	From      string
	RequestID string
	Type      string
	Status    bool
	Pred      gs.Set[string]
	Timestamp uint64
}

func FromFastProposeResponse(pb *pb.FastProposeResponse) Response {
	pred := gs.NewSet[string](pb.Pred...)
	return Response{
		From:      pb.From,
		RequestID: pb.ResquestId,
		Type:      FASTP_REPLY,
		Status:    pb.Result,
		Pred:      pred,
		Timestamp: pb.Time,
	}
}
